package http

import (
	"net/http"
	"net/url"

	"github.com/fivetwenty-io/capi/v3/internal/auth"
)

// NewAuthRetryTransportForTest exposes newAuthRetryTransport to tests in the
// _test package. It is a thin pass-through so tests can construct the
// transport directly without reflecting into unexported state.
func NewAuthRetryTransportForTest(base http.RoundTripper, tm auth.TokenManager) http.RoundTripper {
	return newAuthRetryTransport(base, tm)
}

// SameOriginForTest exposes sameOrigin to tests in the _test package.
func SameOriginForTest(a, b *url.URL) bool {
	return sameOrigin(a, b)
}

// LastResponseErrorHandlerForTest exposes lastResponseErrorHandler so tests
// can check how it treats the final response body.
func LastResponseErrorHandlerForTest(resp *http.Response, err error, numTries int) (*http.Response, error) {
	return lastResponseErrorHandler(resp, err, numTries)
}
