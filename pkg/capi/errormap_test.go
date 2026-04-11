package capi_test

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/fivetwenty-io/capi/v3/pkg/capi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const wellFormedNotFoundEnvelope = `{
  "errors": [
    {
      "code": 10010,
      "title": "CF-ResourceNotFound",
      "detail": "App not found"
    }
  ]
}`

const wellFormedMultiErrorEnvelope = `{
  "errors": [
    {
      "code": 10008,
      "title": "CF-UnprocessableEntity",
      "detail": "name must be unique"
    },
    {
      "code": 10008,
      "title": "CF-UnprocessableEntity",
      "detail": "stack must be present"
    }
  ]
}`

const malformedBody = `<html><body>502 Bad Gateway</body></html>`

// TestMapHTTPError_BelowThresholdReturnsNil verifies that success and
// informational / redirect statuses do NOT produce an error value. This is
// the contract the internal HTTP client relies on to short-circuit the
// error path for 2xx/3xx responses.
func TestMapHTTPError_BelowThresholdReturnsNil(t *testing.T) {
	t.Parallel()

	successStatuses := []int{
		0, // pseudo-value for "no response yet"; MapHTTPError treats it as <400
		http.StatusOK,
		http.StatusCreated,
		http.StatusAccepted,
		http.StatusNoContent,
		http.StatusMovedPermanently,
		http.StatusFound,
		http.StatusNotModified,
		399,
	}

	for _, status := range successStatuses {
		status := status
		t.Run(http.StatusText(status), func(t *testing.T) {
			t.Parallel()
			err := capi.MapHTTPError(status, []byte(wellFormedNotFoundEnvelope))
			assert.NoError(t, err, "status %d should produce nil error", status)
		})
	}
}

// sentinelCase describes a (status, sentinel) pair the mapper must honor.
type sentinelCase struct {
	name     string
	status   int
	sentinel error
}

func allSentinelCases() []sentinelCase {
	return []sentinelCase{
		{name: "404 -> ErrNotFound", status: http.StatusNotFound, sentinel: capi.ErrNotFound},
		{name: "401 -> ErrUnauthorized", status: http.StatusUnauthorized, sentinel: capi.ErrUnauthorized},
		{name: "403 -> ErrForbidden", status: http.StatusForbidden, sentinel: capi.ErrForbidden},
		{name: "409 -> ErrConflict", status: http.StatusConflict, sentinel: capi.ErrConflict},
		{name: "422 -> ErrUnprocessable", status: http.StatusUnprocessableEntity, sentinel: capi.ErrUnprocessable},
		{name: "429 -> ErrRateLimited", status: http.StatusTooManyRequests, sentinel: capi.ErrRateLimited},
		{name: "500 -> ErrServerError", status: http.StatusInternalServerError, sentinel: capi.ErrServerError},
		{name: "502 -> ErrServerError", status: http.StatusBadGateway, sentinel: capi.ErrServerError},
		{name: "503 -> ErrServerError", status: http.StatusServiceUnavailable, sentinel: capi.ErrServerError},
		{name: "504 -> ErrServerError", status: http.StatusGatewayTimeout, sentinel: capi.ErrServerError},
	}
}

// TestMapHTTPError_WellFormedBodyWrapsSentinel verifies that when the body
// is a valid CF v3 error envelope, the returned error unwraps to the
// correct sentinel via errors.Is AND unwraps to *ResponseError via
// errors.As, so existing callers that inspect APIError.Code / Title /
// Detail continue to work.
func TestMapHTTPError_WellFormedBodyWrapsSentinel(t *testing.T) {
	t.Parallel()

	for _, tc := range allSentinelCases() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := capi.MapHTTPError(tc.status, []byte(wellFormedNotFoundEnvelope))
			require.Error(t, err)

			assert.True(
				t,
				errors.Is(err, tc.sentinel),
				"errors.Is(err, %v) must be true for status %d",
				tc.sentinel, tc.status,
			)

			var envelope *capi.ResponseError
			require.True(t, errors.As(err, &envelope),
				"errors.As(err, *ResponseError) must be true when body is well-formed")
			require.NotNil(t, envelope)
			require.Len(t, envelope.Errors, 1)
			assert.Equal(t, 10010, envelope.Errors[0].Code)
			assert.Equal(t, "CF-ResourceNotFound", envelope.Errors[0].Title)
			assert.Equal(t, "App not found", envelope.Errors[0].Detail)
		})
	}
}

// TestMapHTTPError_MalformedBodyStillWrapsSentinel verifies that when the
// server returns a body that is NOT a valid CF error envelope (HTML,
// truncated JSON, empty object, etc.), MapHTTPError still wraps the correct
// sentinel so that `errors.Is` checks keep working, and includes the raw
// body in the error message for debuggability.
func TestMapHTTPError_MalformedBodyStillWrapsSentinel(t *testing.T) {
	t.Parallel()

	malformedInputs := []struct {
		name string
		body []byte
	}{
		{name: "html body", body: []byte(malformedBody)},
		{name: "truncated json", body: []byte(`{"errors": [`)},
		{name: "empty errors array", body: []byte(`{"errors": []}`)},
		{name: "wrong shape", body: []byte(`{"message": "nope"}`)},
		{name: "raw string", body: []byte(`"plain string"`)},
	}

	for _, tc := range allSentinelCases() {
		tc := tc
		for _, input := range malformedInputs {
			input := input
			t.Run(tc.name+"/"+input.name, func(t *testing.T) {
				t.Parallel()

				err := capi.MapHTTPError(tc.status, input.body)
				require.Error(t, err)

				assert.True(
					t,
					errors.Is(err, tc.sentinel),
					"errors.Is(err, %v) must be true for status %d with malformed body",
					tc.sentinel, tc.status,
				)

				// For malformed bodies, MapHTTPError should NOT produce a
				// *ResponseError in the error chain (the body did not
				// parse as one).
				var envelope *capi.ResponseError
				assert.False(
					t,
					errors.As(err, &envelope),
					"errors.As should not find a *ResponseError for malformed body %q", string(input.body),
				)

				// The raw body should be present in the error message for
				// human debugging.
				assert.Contains(t, err.Error(), string(input.body),
					"error message should include the raw body for debugging")
			})
		}
	}
}

// TestMapHTTPError_EmptyBody verifies MapHTTPError does not panic and still
// wraps the sentinel when given a nil or empty body. The error message
// includes the status code in place of a body.
func TestMapHTTPError_EmptyBody(t *testing.T) {
	t.Parallel()

	for _, tc := range allSentinelCases() {
		tc := tc
		t.Run(tc.name+"/nil body", func(t *testing.T) {
			t.Parallel()
			err := capi.MapHTTPError(tc.status, nil)
			require.Error(t, err)
			assert.True(t, errors.Is(err, tc.sentinel))
			assert.Contains(t, err.Error(), "status")
		})
		t.Run(tc.name+"/empty body", func(t *testing.T) {
			t.Parallel()
			err := capi.MapHTTPError(tc.status, []byte{})
			require.Error(t, err)
			assert.True(t, errors.Is(err, tc.sentinel))
			assert.Contains(t, err.Error(), "status")
		})
	}
}

// TestMapHTTPError_MultiErrorEnvelope verifies that a well-formed envelope
// containing multiple APIError entries is fully preserved in the error
// chain and that the returned error still unwraps to the correct sentinel.
func TestMapHTTPError_MultiErrorEnvelope(t *testing.T) {
	t.Parallel()

	err := capi.MapHTTPError(http.StatusUnprocessableEntity, []byte(wellFormedMultiErrorEnvelope))
	require.Error(t, err)

	assert.True(t, errors.Is(err, capi.ErrUnprocessable))

	var envelope *capi.ResponseError
	require.True(t, errors.As(err, &envelope))
	require.Len(t, envelope.Errors, 2)
	assert.Equal(t, "CF-UnprocessableEntity", envelope.Errors[0].Title)
	assert.Equal(t, "name must be unique", envelope.Errors[0].Detail)
	assert.Equal(t, "stack must be present", envelope.Errors[1].Detail)
}

// TestMapHTTPError_UnknownStatusCode verifies that a 4xx status code the
// mapper does not explicitly enumerate (e.g., 418 I'm a teapot) still
// produces a non-nil error. It must NOT match any of the named sentinels.
func TestMapHTTPError_UnknownClient4xx(t *testing.T) {
	t.Parallel()

	err := capi.MapHTTPError(http.StatusTeapot, []byte(`{"errors":[]}`))
	require.Error(t, err)

	namedSentinels := []error{
		capi.ErrNotFound,
		capi.ErrUnauthorized,
		capi.ErrForbidden,
		capi.ErrConflict,
		capi.ErrUnprocessable,
		capi.ErrRateLimited,
		capi.ErrServerError,
	}
	for _, s := range namedSentinels {
		assert.False(t, errors.Is(err, s),
			"unknown 4xx status must not match sentinel %v", s)
	}
	// The generic error message should still mention the status code.
	assert.Contains(t, err.Error(), "418")
}

// TestMapHTTPError_ServerErrorSentinelCatchesAll5xx verifies that any 5xx
// status — even one not explicitly listed — maps to ErrServerError.
func TestMapHTTPError_ServerErrorSentinelCatchesAll5xx(t *testing.T) {
	t.Parallel()

	for status := 500; status <= 599; status++ {
		err := capi.MapHTTPError(status, nil)
		require.Error(t, err, "status %d must produce a non-nil error", status)
		assert.True(
			t,
			errors.Is(err, capi.ErrServerError),
			"status %d must unwrap to ErrServerError", status,
		)
	}
}

// TestMapHTTPError_SentinelIdentity verifies that each sentinel is a
// distinct value and no sentinel unwraps to another sentinel — this is
// critical for errors.Is disambiguation by callers.
func TestMapHTTPError_SentinelIdentity(t *testing.T) {
	t.Parallel()

	sentinels := map[string]error{
		"ErrNotFound":      capi.ErrNotFound,
		"ErrUnauthorized":  capi.ErrUnauthorized,
		"ErrForbidden":     capi.ErrForbidden,
		"ErrConflict":      capi.ErrConflict,
		"ErrUnprocessable": capi.ErrUnprocessable,
		"ErrRateLimited":   capi.ErrRateLimited,
		"ErrServerError":   capi.ErrServerError,
	}

	for name, s := range sentinels {
		require.NotNil(t, s, "%s must not be nil", name)
		require.NotEmpty(t, s.Error(), "%s must have a non-empty message", name)
		assert.True(t, strings.HasPrefix(s.Error(), "capi: "),
			"%s message %q must start with 'capi: '", name, s.Error())
	}

	for nameA, a := range sentinels {
		for nameB, b := range sentinels {
			if nameA == nameB {
				continue
			}
			assert.False(t, errors.Is(a, b),
				"sentinels must be distinct: %s must not unwrap to %s", nameA, nameB)
		}
	}
}

// TestMapHTTPError_IsNotFoundCompatibility verifies that the helper
// IsNotFound (in errors.go) accepts errors produced by MapHTTPError — this
// ensures the sentinel path and the legacy APIError-code path are both
// recognized by the same helper, preserving backward compatibility for
// callers using the helper.
func TestMapHTTPError_IsNotFoundCompatibility(t *testing.T) {
	t.Parallel()

	err := capi.MapHTTPError(http.StatusNotFound, []byte(wellFormedNotFoundEnvelope))
	require.Error(t, err)
	assert.True(t, capi.IsNotFound(err),
		"capi.IsNotFound must return true for errors produced by MapHTTPError")

	err2 := capi.MapHTTPError(http.StatusNotFound, nil)
	require.Error(t, err2)
	assert.True(t, capi.IsNotFound(err2),
		"capi.IsNotFound must work even when body was empty")
}
