// Package http provides the internal HTTP client used by capi/v3.
//
// This file defines authRetryTransport, an http.RoundTripper that sits on
// top of the retryablehttp-managed base transport and transparently handles
// HTTP 401 Unauthorized responses by refreshing the auth token once and
// replaying the request. It is installed by NewClient so that every request
// executed via Client.Do (and the retryablehttp retry machinery) benefits
// from the retry behaviour without any per-request bookkeeping.
//
// Design notes:
//
//   - The retry is attempted at most once per original request. The replayed
//     request goes directly through the base transport (not back through
//     this transport), so a second 401 on the retry is returned to the
//     caller unchanged.
//   - If the request body is a streaming body (Body != nil and GetBody ==
//     nil, i.e. the body cannot be rewound), the original 401 response is
//     returned as-is because a retry would require re-reading a consumed
//     reader.
//   - Any failure during the refresh or token fetch causes the original 401
//     response to be returned unchanged so the caller observes the same
//     semantics they would see without a token manager.
//   - A nil token manager disables the retry entirely; this matches the
//     behaviour of the legacy inline retry in handleResponseError.
package http

import (
	"net/http"

	"github.com/fivetwenty-io/capi/v3/internal/auth"
)

// authRetryTransport is an http.RoundTripper that transparently refreshes
// the auth token and replays a request once when the base transport returns
// HTTP 401 Unauthorized. It is installed by NewClient on the retryablehttp
// client's underlying *http.Client so that all requests issued through the
// retryablehttp machinery pick up the retry automatically.
type authRetryTransport struct {
	base         http.RoundTripper
	tokenManager auth.TokenManager
}

// newAuthRetryTransport returns an authRetryTransport wrapping the provided
// base RoundTripper and TokenManager.
//
// If base is nil, http.DefaultTransport is used so that the returned
// transport is always usable. A nil tokenManager is permitted; in that case
// RoundTrip degrades to a pure pass-through (no refresh, no retry).
func newAuthRetryTransport(base http.RoundTripper, tokenManager auth.TokenManager) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}

	return &authRetryTransport{base: base, tokenManager: tokenManager}
}

// RoundTrip executes the request against the base transport and, on an HTTP
// 401 Unauthorized response, attempts exactly one token refresh + retry
// cycle before returning. See the package doc on authRetryTransport for the
// full set of conditions under which the retry is skipped.
//
// On any condition that prevents retry (nil token manager, non-401 status,
// refresh failure, token fetch failure, or a non-rewindable streaming body)
// the original response is returned unchanged so the caller observes the
// same semantics they would see without this wrapper.
func (t *authRetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	// Fast path: only 401 with a token manager is eligible for retry.
	if resp.StatusCode != http.StatusUnauthorized || t.tokenManager == nil {
		return resp, nil
	}

	// If the request has a non-rewindable body, we cannot retry. Return
	// the original 401 untouched (body still open for the caller to read).
	if req.Body != nil && req.GetBody == nil {
		return resp, nil
	}

	ctx := req.Context()

	if refreshErr := t.tokenManager.RefreshToken(ctx); refreshErr != nil {
		// Refresh failed — return the original 401 unchanged.
		return resp, nil
	}

	token, tokenErr := t.tokenManager.GetToken(ctx)
	if tokenErr != nil {
		return resp, nil
	}

	// Build a cloned request with the refreshed Authorization header. We
	// clone (rather than mutate) so the caller's original *http.Request is
	// left untouched, matching the standard RoundTripper contract.
	retryReq := req.Clone(ctx)
	retryReq.Header.Set("Authorization", "Bearer "+token)

	// Rewind the request body if present. GetBody is guaranteed non-nil
	// at this point by the streaming-body check above, so any error here
	// is a genuine IO/wrap failure and we fall back to returning the
	// original 401 response unchanged.
	if req.Body != nil {
		newBody, getBodyErr := req.GetBody()
		if getBodyErr != nil {
			return resp, nil
		}

		retryReq.Body = newBody
	}

	// We are committed to the retry. Drain and close the original 401
	// response body so we do not leak the underlying connection, then
	// issue the replay directly against the base transport (bypassing
	// this wrapper so a second 401 is returned unchanged).
	_ = resp.Body.Close()

	return t.base.RoundTrip(retryReq)
}
