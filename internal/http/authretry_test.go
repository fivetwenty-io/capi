package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	capihttp "github.com/fivetwenty-io/capi/v3/internal/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Sentinel errors for refresh and token-fetch failure paths in tests.
var (
	errRefreshExplode = errors.New("test: refresh exploded")
	errTokenExplode   = errors.New("test: get-token exploded")
)

// refreshingTokenManager is a token manager stub that can be configured to
// fail either the initial GetToken, the RefreshToken call, or the post-
// refresh GetToken call. It tracks how many times each method is called so
// tests can assert the refresh-once-on-401 contract.
type refreshingTokenManager struct {
	initialToken   string
	refreshedToken string

	refreshErr    error
	getTokenErr   error
	postRefreshEr error

	getCalls     int32
	refreshCalls int32
	refreshed    int32 // 1 once RefreshToken has succeeded
}

func (m *refreshingTokenManager) GetToken(_ context.Context) (string, error) {
	atomic.AddInt32(&m.getCalls, 1)

	if atomic.LoadInt32(&m.refreshed) == 1 {
		if m.postRefreshEr != nil {
			return "", m.postRefreshEr
		}

		return m.refreshedToken, nil
	}

	if m.getTokenErr != nil {
		return "", m.getTokenErr
	}

	return m.initialToken, nil
}

func (m *refreshingTokenManager) RefreshToken(_ context.Context) error {
	atomic.AddInt32(&m.refreshCalls, 1)

	if m.refreshErr != nil {
		return m.refreshErr
	}

	atomic.StoreInt32(&m.refreshed, 1)

	return nil
}

func (m *refreshingTokenManager) SetToken(_ string, _ time.Time) {}

// TestAuthRetryTransport_Refresh401 verifies that a 401 response triggers a
// single RefreshToken + replay cycle and that the replay returns the new
// response to the caller.
func TestAuthRetryTransport_Refresh401(t *testing.T) {
	t.Parallel()

	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)

		if n == 1 {
			assert.Equal(t, "Bearer initial", r.Header.Get("Authorization"))
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		assert.Equal(t, "Bearer refreshed", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	tm := &refreshingTokenManager{
		initialToken:   "initial",
		refreshedToken: "refreshed",
	}
	client := capihttp.NewClient(server.URL, tm,
		capihttp.WithRetryConfig(0, time.Millisecond, time.Millisecond))

	resp, err := client.Get(context.Background(), "/v3/apps", nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]bool
	require.NoError(t, json.Unmarshal(resp.Body, &body))
	assert.True(t, body["ok"])

	assert.Equal(t, int32(2), atomic.LoadInt32(&attempts), "server should see exactly two attempts")
	assert.Equal(t, int32(1), atomic.LoadInt32(&tm.refreshCalls), "RefreshToken should be called exactly once")
}

// TestAuthRetryTransport_NonUnauthorizedPassThrough verifies that a non-401
// response does NOT trigger a refresh.
func TestAuthRetryTransport_NonUnauthorizedPassThrough(t *testing.T) {
	t.Parallel()

	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	tm := &refreshingTokenManager{initialToken: "initial", refreshedToken: "refreshed"}
	client := capihttp.NewClient(server.URL, tm,
		capihttp.WithRetryConfig(0, time.Millisecond, time.Millisecond))

	resp, err := client.Get(context.Background(), "/v3/apps", nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assert.Equal(t, int32(1), atomic.LoadInt32(&attempts))
	assert.Equal(t, int32(0), atomic.LoadInt32(&tm.refreshCalls), "no refresh should occur on 200")
}

// TestAuthRetryTransport_RefreshFailureReturnsOriginal401 verifies that a
// RefreshToken failure causes the original 401 to be surfaced to the caller.
func TestAuthRetryTransport_RefreshFailureReturnsOriginal401(t *testing.T) {
	t.Parallel()

	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"errors":[{"code":10002,"title":"CF-NotAuthenticated","detail":"Not authenticated"}]}`))
	}))
	defer server.Close()

	tm := &refreshingTokenManager{
		initialToken: "initial",
		refreshErr:   errRefreshExplode,
	}
	client := capihttp.NewClient(server.URL, tm,
		capihttp.WithRetryConfig(0, time.Millisecond, time.Millisecond))

	resp, err := client.Get(context.Background(), "/v3/apps", nil)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Contains(t, string(resp.Body), "CF-NotAuthenticated")

	assert.Equal(t, int32(1), atomic.LoadInt32(&attempts), "server should see exactly one attempt when refresh fails")
	assert.Equal(t, int32(1), atomic.LoadInt32(&tm.refreshCalls))
}

// TestAuthRetryTransport_PostRefreshGetTokenFailure verifies that a failure
// to fetch the token AFTER a successful refresh also surfaces the original
// 401 unchanged (no retry replay).
func TestAuthRetryTransport_PostRefreshGetTokenFailure(t *testing.T) {
	t.Parallel()

	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	tm := &refreshingTokenManager{
		initialToken:  "initial",
		postRefreshEr: errTokenExplode,
	}
	client := capihttp.NewClient(server.URL, tm,
		capihttp.WithRetryConfig(0, time.Millisecond, time.Millisecond))

	resp, err := client.Get(context.Background(), "/v3/apps", nil)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// Exactly one server attempt: refresh succeeded but the follow-up
	// GetToken failed before a retry could be issued.
	assert.Equal(t, int32(1), atomic.LoadInt32(&attempts))
	assert.Equal(t, int32(1), atomic.LoadInt32(&tm.refreshCalls))
}

// streamingReadCloser is an io.ReadCloser that, when paired with a request
// via http.NewRequestWithContext, yields a request whose GetBody is nil —
// simulating a non-rewindable streaming body.
type streamingReadCloser struct {
	inner io.Reader
}

func (s *streamingReadCloser) Read(p []byte) (int, error) { return s.inner.Read(p) }
func (s *streamingReadCloser) Close() error               { return nil }

// TestAuthRetryTransport_StreamingBodyNoRetry verifies that when the
// request body cannot be rewound (GetBody == nil), a 401 is returned to the
// caller WITHOUT a refresh + retry attempt. This path cannot be exercised
// through the high-level capihttp.Client (which always produces rewindable
// *bytes.Reader bodies), so we use the unexported constructor directly via
// the test-only accessor in export_test.go.
func TestAuthRetryTransport_StreamingBodyNoRetry(t *testing.T) {
	t.Parallel()

	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)

		_, _ = io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	tm := &refreshingTokenManager{initialToken: "initial", refreshedToken: "refreshed"}

	transport := capihttp.NewAuthRetryTransportForTest(http.DefaultTransport, tm)

	streaming := &streamingReadCloser{inner: bytes.NewReader([]byte("streamed"))}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, server.URL+"/v3/apps", streaming)
	require.NoError(t, err)

	// http.NewRequestWithContext only populates GetBody for body types it
	// recognises (*bytes.Buffer, *bytes.Reader, *strings.Reader). A bare
	// io.ReadCloser yields GetBody == nil — exactly the case we want.
	require.Nil(t, req.GetBody, "test precondition: request must have a non-rewindable body")
	req.Header.Set("Authorization", "Bearer initial")

	resp, err := transport.RoundTrip(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	// Exactly one server attempt — the streaming body prevents retry.
	assert.Equal(t, int32(1), atomic.LoadInt32(&attempts))
	assert.Equal(t, int32(0), atomic.LoadInt32(&tm.refreshCalls),
		"streaming body must not trigger RefreshToken because replay is impossible")
}

// TestAuthRetryTransport_NilTokenManager verifies that a transport with a
// nil token manager degrades to a pure pass-through: a 401 is returned
// unchanged with no refresh attempt.
func TestAuthRetryTransport_NilTokenManager(t *testing.T) {
	t.Parallel()

	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	transport := capihttp.NewAuthRetryTransportForTest(http.DefaultTransport, nil)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/v3/apps", nil)
	require.NoError(t, err)

	resp, err := transport.RoundTrip(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Equal(t, int32(1), atomic.LoadInt32(&attempts))
}

// TestAuthRetryTransport_RewindableBodyRetries verifies that a POST with a
// rewindable body (as produced by capihttp.Client.Post for a JSON body)
// successfully replays after a 401.
func TestAuthRetryTransport_RewindableBodyRetries(t *testing.T) {
	t.Parallel()

	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)

		body, _ := io.ReadAll(r.Body)
		assert.Contains(t, string(body), `"name":"test-app"`,
			"body must be replayed verbatim on retry")

		if n == 1 {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		assert.Equal(t, "Bearer refreshed", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	tm := &refreshingTokenManager{
		initialToken:   "initial",
		refreshedToken: "refreshed",
	}
	client := capihttp.NewClient(server.URL, tm,
		capihttp.WithRetryConfig(0, time.Millisecond, time.Millisecond))

	resp, err := client.Post(context.Background(), "/v3/apps", map[string]string{"name": "test-app"})
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	assert.Equal(t, int32(2), atomic.LoadInt32(&attempts))
	assert.Equal(t, int32(1), atomic.LoadInt32(&tm.refreshCalls))
}
