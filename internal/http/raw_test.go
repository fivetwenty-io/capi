package http_test

import (
	"context"
	"io"
	nethttp "net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/fivetwenty-io/capi/v3/internal/http"
	"github.com/fivetwenty-io/capi/v3/pkg/capi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	originAPI     = "https://api.example.com"
	originAPIPort = "https://api.example.com:8443"
)

func TestSameOrigin(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		left, right string
		same        bool
	}{
		"identical":                  {originAPI, "https://api.example.com/x", true},
		"different host":             {originAPI, "https://other.example.com", false},
		"different host same port":   {originAPIPort, "https://evil.example.com:8443", false},
		"different scheme same host": {originAPIPort, "http://api.example.com:8443", false},
		"different port":             {originAPIPort, "https://api.example.com:9443", false},
		"https default port":         {originAPI, "https://api.example.com:443", true},
		"https is not port 80":       {originAPI, "https://api.example.com:80", false},
		"http default port":          {"http://api.example.com", "http://api.example.com:80", true},
		"http is not port 443":       {"http://api.example.com", "http://api.example.com:443", false},
	}

	for name, tcase := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			left, err := url.Parse(tcase.left)
			require.NoError(t, err)

			right, err := url.Parse(tcase.right)
			require.NoError(t, err)

			assert.Equal(t, tcase.same, http.SameOriginForTest(left, right))
			assert.Equal(t, tcase.same, http.SameOriginForTest(right, left))
		})
	}
}

type closeTrackingBody struct {
	io.Reader

	closed *atomic.Bool
}

func (b closeTrackingBody) Close() error {
	b.closed.Store(true)

	return nil
}

type stubTransport struct{ closed *atomic.Bool }

func (s stubTransport) RoundTrip(req *nethttp.Request) (*nethttp.Response, error) {
	return &nethttp.Response{
		StatusCode: nethttp.StatusOK,
		Header:     nethttp.Header{},
		Body:       closeTrackingBody{Reader: strings.NewReader("ok"), closed: s.closed},
		Request:    req,
	}, nil
}

func TestDoRawClosesResponseBody(t *testing.T) {
	t.Parallel()

	var closed atomic.Bool

	client := http.NewClient("http://api.example.test", nil,
		http.WithHTTPClient(&nethttp.Client{Transport: stubTransport{closed: &closed}}))

	result, err := client.DoRaw(context.Background(), capi.RawRequest{Path: "/v3/apps"})
	require.NoError(t, err)
	assert.Equal(t, "ok", string(result.Body))
	assert.True(t, closed.Load(), "the response body must be closed")
}
