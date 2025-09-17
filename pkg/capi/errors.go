package capi

import (
	"encoding/json"
	"errors"
	"fmt"
)

// APIError represents an error from the CF API.
type APIError struct {
	Code   int    `json:"code"   yaml:"code"`
	Title  string `json:"title"  yaml:"title"`
	Detail string `json:"detail" yaml:"detail"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s (code: %d)", e.Title, e.Detail, e.Code)
}

// ResponseError represents the error response from the API.
type ResponseError struct {
	Errors []APIError `json:"errors"`
}

// Error implements the error interface for ResponseError.
func (e *ResponseError) Error() string {
	if len(e.Errors) == 0 {
		return "unknown error"
	}

	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}

	return fmt.Sprintf("multiple errors: %v", e.Errors)
}

// FirstError returns the first error or nil.
func (e *ResponseError) FirstError() *APIError {
	if len(e.Errors) > 0 {
		return &e.Errors[0]
	}

	return nil
}

// Common error codes.
const (
	ErrorCodeNotFound               = 10010
	ErrorCodeNotAuthenticated       = 10002
	ErrorCodeNotAuthorized          = 10003
	ErrorCodeUnprocessableEntity    = 10008
	ErrorCodeServiceUnavailable     = 10001
	ErrorCodeBadRequest             = 10005
	ErrorCodeUniquenessError        = 10016
	ErrorCodeResourceNotFound       = 10010
	ErrorCodeInvalidRelation        = 10020
	ErrorCodeTooManyRequests        = 10013
	ErrorCodeMaintenanceInfo        = 10012
	ErrorCodeServiceInstanceQuota   = 10003
	ErrorCodeAsyncServiceInProgress = 10001
)

// Common error types.
var (
	ErrNotFound           = &APIError{Code: ErrorCodeNotFound, Title: "CF-ResourceNotFound"}
	ErrUnauthorized       = &APIError{Code: ErrorCodeNotAuthenticated, Title: "CF-NotAuthenticated"}
	ErrForbidden          = &APIError{Code: ErrorCodeNotAuthorized, Title: "CF-NotAuthorized"}
	ErrUnprocessable      = &APIError{Code: ErrorCodeUnprocessableEntity, Title: "CF-UnprocessableEntity"}
	ErrServiceUnavailable = &APIError{Code: ErrorCodeServiceUnavailable, Title: "CF-ServiceUnavailable"}
	ErrBadRequest         = &APIError{Code: ErrorCodeBadRequest, Title: "CF-BadRequest"}
	ErrTooManyRequests    = &APIError{Code: ErrorCodeTooManyRequests, Title: "CF-TooManyRequests"}
)

// Common static errors that can be wrapped with context.
var (
	ErrAPIAlreadyExists            = errors.New("API already exists")
	ErrAPINotFound                 = errors.New("API not found")
	ErrCannotDeleteOnlyAPI         = errors.New("cannot delete the only configured API")
	ErrNoHostInURL                 = errors.New("no host specified in URL")
	ErrInvalidFilePath             = errors.New("invalid file path")
	ErrPathTraversalAttempt        = errors.New("potential path traversal attempt")
	ErrPathTraversalNotAllowed     = errors.New("path traversal not allowed")
	ErrSpaceNotFound               = errors.New("space not found")
	ErrApplicationNotFound         = errors.New("application not found")
	ErrNoProcessesFound            = errors.New("no processes found")
	ErrProcessTypeNotFound         = errors.New("process type not found")
	ErrEnvironmentVariableNotFound = errors.New("environment variable not found")
	ErrInstanceIndexOutOfRange     = errors.New("instance index out of range")
	ErrOrganizationNotFound        = errors.New("organization not found")
	ErrBuildpackNotFound           = errors.New("buildpack not found")
	ErrBuildpackNameRequired       = errors.New("buildpack name is required")
	ErrNoAPIsConfigured            = errors.New("no APIs configured")
	ErrCurrentAPINotFound          = errors.New("current API not found in configuration")
	ErrAPINameOrEndpointRequired   = errors.New("API name or endpoint is required")
	ErrNoAPIEndpointConfigured     = errors.New("no API endpoint configured")
	ErrCouldNotDetermineAPIDomain  = errors.New("could not determine API domain for configuration")
	ErrStaticTokenCannotRefresh    = errors.New("static token cannot be refreshed")
	ErrCircuitBreakerOpen          = errors.New("circuit breaker is open")
	ErrNoMoreItems                 = errors.New("no more items")
	ErrConfigRequired              = errors.New("config is required")
	ErrAPIEndpointRequired         = errors.New("API endpoint is required")
	ErrSkipTLSOnlyInDev            = errors.New("skipTLS is only allowed in development environments")
	ErrRootInfoRequestFailed       = errors.New("root info request failed")
	ErrNoUAAOrLoginURL             = errors.New("no UAA or login URL found in API root response")
	ErrInvalidHealthCheckType      = errors.New("invalid health check type")
	ErrNotImplemented              = errors.New("not implemented")
	ErrNotAuthenticated            = errors.New("not authenticated")
	ErrUnknownConfigKey            = errors.New("unknown configuration key")
	ErrTokenFieldsCannotUnset      = errors.New("token fields cannot be unset via config command")
	ErrDomainNotFound              = errors.New("domain not found")
	ErrDomainNameRequired          = errors.New("domain name is required")
	ErrInvalidClientType           = errors.New("invalid client type")
)

// IsNotFound checks if the error is a not found error.
func IsNotFound(err error) bool {
	apiErr := &APIError{}
	if errors.As(err, &apiErr) {
		return apiErr.Code == ErrorCodeNotFound
	}

	errResp := &ResponseError{}
	if errors.As(err, &errResp) {
		first := errResp.FirstError()
		if first != nil {
			return first.Code == ErrorCodeNotFound
		}
	}

	return false
}

// IsUnauthorized checks if the error is an unauthorized error.
func IsUnauthorized(err error) bool {
	apiErr := &APIError{}
	if errors.As(err, &apiErr) {
		return apiErr.Code == ErrorCodeNotAuthenticated
	}

	errResp := &ResponseError{}
	if errors.As(err, &errResp) {
		first := errResp.FirstError()
		if first != nil {
			return first.Code == ErrorCodeNotAuthenticated
		}
	}

	return false
}

// IsForbidden checks if the error is a forbidden error.
func IsForbidden(err error) bool {
	apiErr := &APIError{}
	if errors.As(err, &apiErr) {
		return apiErr.Code == ErrorCodeNotAuthorized
	}

	errResp := &ResponseError{}
	if errors.As(err, &errResp) {
		first := errResp.FirstError()
		if first != nil {
			return first.Code == ErrorCodeNotAuthorized
		}
	}

	return false
}

// ParseResponseError parses an error response from JSON.
func ParseResponseError(data []byte) (*ResponseError, error) {
	var errResp ResponseError

	err := json.Unmarshal(data, &errResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response error: %w", err)
	}

	return &errResp, nil
}

// Test error variables for test files to comply with err113.
var (
	ErrAppNotFound = errors.New("app not found")
	ErrSomeError   = errors.New("some error")
)
