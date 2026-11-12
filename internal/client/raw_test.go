package client_test

import (
	"bytes"
	"context"
	"encoding/pem"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/fivetwenty-io/capi/v3/internal/client"
	"github.com/fivetwenty-io/capi/v3/pkg/capi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	rawTestToken = "raw-secret-token"
	rawAppsPath  = "/v3/apps"
	rawXPath     = "/v3/x"
	rawBodyB     = `{"b":1}`
	rawPEMType   = "CERTIFICATE"
)

func newRawClient(t *testing.T, endpoint string, mutate ...func(*capi.Config)) *Client {
	t.Helper()

	cfg := &capi.Config{APIEndpoint: endpoint, AccessToken: rawTestToken}
	for _, m := range mutate {
		m(cfg)
	}

	c, err := New(context.Background(), cfg)
	require.NoError(t, err)

	return c
}

func TestRawSuccess(t *testing.T) {
	t.Parallel()

	var gotAuth, gotMethod, gotURI, gotBody string

	srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotMethod = r.Method
		gotURI = r.URL.RequestURI()
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)

		writer.Header().Set("X-Test", "yes")
		writer.WriteHeader(http.StatusNotFound)
		_, _ = writer.Write([]byte(`{"errors":[]}`))
	}))
	defer srv.Close()

	c := newRawClient(t, srv.URL)

	resp, err := c.Raw(context.Background(), capi.RawRequest{
		Method: "post",
		Path:   "/v3/things",
		Query:  map[string][]string{"a": {"b c"}},
		Body:   []byte(`{"x":1}`),
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.Status)
	assert.Equal(t, "yes", resp.Header.Get("X-Test"))
	assert.JSONEq(t, `{"errors":[]}`, string(resp.Body))
	assert.False(t, resp.Truncated)
	assert.Equal(t, "Bearer "+rawTestToken, gotAuth)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/v3/things?a=b+c", gotURI)
	assert.JSONEq(t, `{"x":1}`, gotBody)
}

func TestRawRejectsUnsafePaths(t *testing.T) {
	t.Parallel()

	var hits atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		hits.Add(1)
	}))
	t.Cleanup(srv.Close)

	client := newRawClient(t, srv.URL)

	cases := map[string]string{
		"empty":                         "",
		"no leading slash":              "v3/apps",
		"scheme":                        "https://evil.example/v3/apps",
		"scheme no slash":               "javascript:alert(1)",
		"host":                          "//evil.example/v3/apps",
		"user information":              "//user:pw@evil.example/v3/apps",
		"parent segment":                "/v3/../admin",
		"trailing parent segment":       "/v3/apps/..",
		"leading parent segment":        "/../etc",
		"encoded parent segment":        "/v3/%2e%2e/admin",
		"upper encoded parent segment":  "/v3/%2E%2E/admin",
		"mixed encoded parent segment":  "/v3/.%2e/admin",
		"double encoded parent segment": "/v3/%252e%252e/admin",
		"encoded slash":                 "/v3/apps%2f..%2fadmin",
		"upper encoded slash":           "/v3/apps%2Fadmin",
		"encoded backslash":             "/v3/apps%5cadmin",
		"backslash":                     "/v3\\apps",
		"control character":             "/v3/apps\x00",
		"newline":                       "/v3/apps\nHost: evil",
		"encoded control":               "/v3/apps%0d%0a",
		"delete character":              "/v3/apps\x7f",
		"query in path":                 "/v3/apps?x=1",
		"fragment in path":              "/v3/apps#x",
		"bad escape":                    "/v3/apps%zz",
		"encoded leading slash":         "/%2fevil.example/x",
	}

	for name, path := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp, err := client.Raw(context.Background(), capi.RawRequest{Method: "GET", Path: path})
			require.Error(t, err)
			assert.Nil(t, resp)
			require.ErrorIs(t, err, capi.ErrRawInvalidPath)
			assert.NotContains(t, err.Error(), rawTestToken)
		})
	}

	_, err := client.Raw(context.Background(), capi.RawRequest{Method: "TRACE", Path: rawAppsPath})
	require.ErrorIs(t, err, capi.ErrRawInvalidMethod)

	_, err = client.Raw(context.Background(), capi.RawRequest{Method: "GE\r\nT", Path: rawAppsPath})
	require.ErrorIs(t, err, capi.ErrRawInvalidMethod)

	assert.Zero(t, hits.Load(), "no request may be sent for a refused request")
}

func TestRawAcceptsEncodedPathData(t *testing.T) {
	t.Parallel()

	var gotPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
	}))
	defer srv.Close()

	c := newRawClient(t, srv.URL)

	_, err := c.Raw(context.Background(), capi.RawRequest{Path: "/v3/apps/a%20b"})
	require.NoError(t, err)
	assert.Equal(t, "/v3/apps/a%20b", gotPath)
}

func TestRawDropsTokenOnCrossHostRedirect(t *testing.T) {
	t.Parallel()

	var otherHits atomic.Int32

	var otherAuth atomic.Value

	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		otherHits.Add(1)
		otherAuth.Store(r.Header.Get("Authorization"))
	}))
	defer other.Close()

	first := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, other.URL+"/v3/landing", http.StatusTemporaryRedirect)
	}))
	defer first.Close()

	c := newRawClient(t, first.URL)

	resp, err := c.Raw(context.Background(), capi.RawRequest{Path: rawAppsPath})
	require.ErrorIs(t, err, capi.ErrRawRedirectRefused)
	assert.Nil(t, resp)
	assert.NotContains(t, err.Error(), rawTestToken)
	assert.NotContains(t, err.Error(), other.URL)

	if v := otherAuth.Load(); v != nil {
		assert.Empty(t, v, "the second host must not see the token")
	}

	assert.Zero(t, otherHits.Load(), "the second host must not be contacted")
}

func TestRawFollowsSameOriginRedirectWithToken(t *testing.T) {
	t.Parallel()

	var landedAuth string

	mux := http.NewServeMux()
	mux.HandleFunc("/v3/old", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/v3/new", http.StatusFound)
	})
	mux.HandleFunc("/v3/new", func(w http.ResponseWriter, r *http.Request) {
		landedAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte("ok"))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := newRawClient(t, srv.URL)

	resp, err := c.Raw(context.Background(), capi.RawRequest{Path: "/v3/old"})
	require.NoError(t, err)
	assert.Equal(t, "ok", string(resp.Body))
	assert.Equal(t, "Bearer "+rawTestToken, landedAuth)
}

func TestRawRefusesHTTPSDowngradeRedirect(t *testing.T) {
	t.Parallel()

	var plainHits atomic.Int32

	plain := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		plainHits.Add(1)
	}))
	defer plain.Close()

	secure := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, plain.URL+rawAppsPath, http.StatusFound)
	}))
	defer secure.Close()

	caPEM := string(pem.EncodeToMemory(&pem.Block{Type: rawPEMType, Bytes: secure.Certificate().Raw}))
	c := newRawClient(t, secure.URL, func(cfg *capi.Config) { cfg.CACertPEM = caPEM })

	_, err := c.Raw(context.Background(), capi.RawRequest{Path: rawAppsPath})
	require.ErrorIs(t, err, capi.ErrRawRedirectRefused)
	assert.Zero(t, plainHits.Load())
}

func TestRawStopsRedirectLoop(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/v3/loop", http.StatusFound)
	}))
	defer srv.Close()

	c := newRawClient(t, srv.URL)

	_, err := c.Raw(context.Background(), capi.RawRequest{Path: "/v3/loop"})
	require.ErrorIs(t, err, capi.ErrRawRedirectRefused)
}

func TestRawTruncatesAbove16MiB(t *testing.T) {
	t.Parallel()

	const limit = 16 << 20

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(bytes.Repeat([]byte("a"), limit+1024))
	}))
	defer srv.Close()

	c := newRawClient(t, srv.URL)

	resp, err := c.Raw(context.Background(), capi.RawRequest{Path: "/v3/big"})
	require.NoError(t, err)
	assert.True(t, resp.Truncated)
	assert.Len(t, resp.Body, limit)
}

func TestRawExactLimitIsNotTruncated(t *testing.T) {
	t.Parallel()

	const limit = 16 << 20

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(bytes.Repeat([]byte("a"), limit))
	}))
	defer srv.Close()

	c := newRawClient(t, srv.URL)

	resp, err := c.Raw(context.Background(), capi.RawRequest{Path: "/v3/big"})
	require.NoError(t, err)
	assert.False(t, resp.Truncated)
	assert.Len(t, resp.Body, limit)
}

func TestRawTransportErrorHidesURLAndToken(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	srv.Close()

	c := newRawClient(t, url, func(cfg *capi.Config) { cfg.AccessToken = rawTestToken })

	_, err := c.Raw(canceledContext(), capi.RawRequest{Path: rawAppsPath, Query: map[string][]string{"q": {"sensitive"}}})
	require.Error(t, err)
	require.ErrorIs(t, err, context.Canceled)
	assert.NotContains(t, err.Error(), "sensitive")
	assert.NotContains(t, err.Error(), rawTestToken)
}

func canceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	return ctx
}

type rawTokenManager struct {
	mu        sync.Mutex
	token     string
	refreshes int
}

func (m *rawTokenManager) GetToken(context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.token, nil
}

func (m *rawTokenManager) RefreshToken(context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.refreshes++
	m.token = "fresh-token"

	return nil
}

func (m *rawTokenManager) SetToken(string, time.Time) {}

func TestRawRefreshesTokenAndReplaysBody(t *testing.T) {
	t.Parallel()

	manager := &rawTokenManager{token: "stale-token"}

	var (
		lock   sync.Mutex
		auths  []string
		bodies []string
	)

	srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		data, _ := io.ReadAll(req.Body)

		lock.Lock()

		auths = append(auths, req.Header.Get("Authorization"))
		bodies = append(bodies, string(data))
		lock.Unlock()

		if req.Header.Get("Authorization") != "Bearer fresh-token" {
			writer.WriteHeader(http.StatusUnauthorized)

			return
		}

		writer.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	client, err := NewWithTokenManager(&capi.Config{APIEndpoint: srv.URL}, manager)
	require.NoError(t, err)

	resp, err := client.Raw(context.Background(), capi.RawRequest{Method: http.MethodPost, Path: rawXPath, Body: []byte(`{"replay":true}`)})
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.Status)

	lock.Lock()
	defer lock.Unlock()

	assert.Equal(t, 1, manager.refreshes)
	assert.Equal(t, []string{"Bearer stale-token", "Bearer fresh-token"}, auths)
	assert.Equal(t, []string{`{"replay":true}`, `{"replay":true}`}, bodies)
}

func TestRawReplaysBodyOnRetry(t *testing.T) {
	t.Parallel()

	var (
		lock   sync.Mutex
		bodies []string
	)

	srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		data, _ := io.ReadAll(req.Body)

		lock.Lock()

		bodies = append(bodies, string(data))
		attempt := len(bodies)
		lock.Unlock()

		if attempt < 3 {
			writer.WriteHeader(http.StatusServiceUnavailable)

			return
		}

		writer.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := newRawClient(t, srv.URL, func(cfg *capi.Config) {
		cfg.RetryMax = 3
		cfg.RetryWaitMin = time.Millisecond
		cfg.RetryWaitMax = 2 * time.Millisecond
	})

	resp, err := client.Raw(context.Background(), capi.RawRequest{Method: "PUT", Path: rawXPath, Body: []byte(rawBodyB)})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.Status)

	lock.Lock()
	defer lock.Unlock()

	assert.Equal(t, []string{rawBodyB, rawBodyB, rawBodyB}, bodies)
}

func TestRawReturnsLastResponseWhenRetriesRunOut(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusServiceUnavailable)
		_, _ = writer.Write([]byte("down"))
	}))
	defer srv.Close()

	client := newRawClient(t, srv.URL, func(cfg *capi.Config) {
		cfg.RetryMax = 2
		cfg.RetryWaitMin = time.Millisecond
		cfg.RetryWaitMax = 2 * time.Millisecond
	})

	resp, err := client.Raw(context.Background(), capi.RawRequest{Path: rawXPath})
	require.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, resp.Status)
	assert.Equal(t, "down", string(resp.Body))
}

func TestRawSendsStandardHeaders(t *testing.T) {
	t.Parallel()

	var withBody, withoutBody http.Header

	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			withBody = r.Header.Clone()
		} else {
			withoutBody = r.Header.Clone()
		}
	}))
	defer srv.Close()

	client := newRawClient(t, srv.URL)

	_, err := client.Raw(context.Background(), capi.RawRequest{Method: "POST", Path: rawXPath, Body: []byte(`{}`)})
	require.NoError(t, err)

	_, err = client.Raw(context.Background(), capi.RawRequest{Path: rawXPath})
	require.NoError(t, err)

	assert.Equal(t, "application/json", withBody.Get("Content-Type"))
	assert.Equal(t, "application/json", withBody.Get("Accept"))
	assert.NotEmpty(t, withBody.Get("User-Agent"))
	assert.Empty(t, withoutBody.Get("Content-Type"))
	assert.Equal(t, "application/json", withoutBody.Get("Accept"))
}

func TestRawRefusesRedirectToHostnameAlias(t *testing.T) {
	t.Parallel()

	var (
		aliasHits atomic.Int32
		baseURL   string
	)

	srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/start" {
			_, port, _ := net.SplitHostPort(strings.TrimPrefix(baseURL, "http://"))
			http.Redirect(writer, r, "http://localhost:"+port+"/landed", http.StatusFound)

			return
		}

		aliasHits.Add(1)
	}))
	defer srv.Close()

	baseURL = srv.URL
	client := newRawClient(t, srv.URL)

	_, err := client.Raw(context.Background(), capi.RawRequest{Path: "/start"})
	require.ErrorIs(t, err, capi.ErrRawRedirectRefused)
	assert.Zero(t, aliasHits.Load(), "the same port on another host name must not be contacted")
}

func TestRawDoesNotRetryRefusedRedirect(t *testing.T) {
	t.Parallel()

	var apiHits, otherHits atomic.Int32

	other := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		otherHits.Add(1)
	}))
	defer other.Close()

	api := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		apiHits.Add(1)
		http.Redirect(writer, r, other.URL+"/x", http.StatusTemporaryRedirect)
	}))
	defer api.Close()

	client := newRawClient(t, api.URL, func(cfg *capi.Config) {
		cfg.RetryMax = 3
		cfg.RetryWaitMin = time.Millisecond
		cfg.RetryWaitMax = 2 * time.Millisecond
	})

	_, err := client.Raw(context.Background(), capi.RawRequest{Method: "POST", Path: rawXPath, Body: []byte(`{}`)})
	require.ErrorIs(t, err, capi.ErrRawRedirectRefused)
	assert.Equal(t, int32(1), apiHits.Load(), "a refused redirect must not replay the request")
	assert.Zero(t, otherHits.Load())
}

func TestRawKeepsCancelCauseDuringBodyRead(t *testing.T) {
	t.Parallel()

	release := make(chan struct{})

	srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Length", "1000")
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte("partial"))

		if flusher, ok := writer.(http.Flusher); ok {
			flusher.Flush()
		}

		<-release
	}))
	t.Cleanup(srv.Close)
	t.Cleanup(func() { close(release) })

	client := newRawClient(t, srv.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_, err := client.Raw(ctx, capi.RawRequest{Path: rawXPath})
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestRawStopsWhenContextIsCanceled(t *testing.T) {
	t.Parallel()

	release := make(chan struct{})

	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		<-release
	}))
	t.Cleanup(srv.Close)
	t.Cleanup(func() { close(release) })

	client := newRawClient(t, srv.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := client.Raw(ctx, capi.RawRequest{Path: rawXPath})
	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, time.Since(start), 2*time.Second)
}
