package capi

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIError_Error(t *testing.T) {
	err := &APIError{
		Code:   10010,
		Title:  "CF-ResourceNotFound",
		Detail: "App not found",
	}

	assert.Equal(t, "CF-ResourceNotFound: App not found (code: 10010)", err.Error())
}

func TestErrorResponse_Error(t *testing.T) {
	tests := []struct {
		name     string
		response *ErrorResponse
		expected string
	}{
		{
			name:     "empty errors",
			response: &ErrorResponse{},
			expected: "unknown error",
		},
		{
			name: "single error",
			response: &ErrorResponse{
				Errors: []APIError{
					{
						Code:   10010,
						Title:  "CF-ResourceNotFound",
						Detail: "App not found",
					},
				},
			},
			expected: "CF-ResourceNotFound: App not found (code: 10010)",
		},
		{
			name: "multiple errors",
			response: &ErrorResponse{
				Errors: []APIError{
					{
						Code:   10010,
						Title:  "CF-ResourceNotFound",
						Detail: "App not found",
					},
					{
						Code:   10008,
						Title:  "CF-UnprocessableEntity",
						Detail: "Invalid request",
					},
				},
			},
			expected: "multiple errors: [{10010 CF-ResourceNotFound App not found} {10008 CF-UnprocessableEntity Invalid request}]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.response.Error())
		})
	}
}

func TestErrorResponse_FirstError(t *testing.T) {
	t.Run("with errors", func(t *testing.T) {
		response := &ErrorResponse{
			Errors: []APIError{
				{Code: 10010, Title: "CF-ResourceNotFound", Detail: "Not found"},
				{Code: 10008, Title: "CF-UnprocessableEntity", Detail: "Invalid"},
			},
		}

		first := response.FirstError()
		require.NotNil(t, first)
		assert.Equal(t, 10010, first.Code)
		assert.Equal(t, "CF-ResourceNotFound", first.Title)
	})

	t.Run("without errors", func(t *testing.T) {
		response := &ErrorResponse{}
		assert.Nil(t, response.FirstError())
	})
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "APIError not found",
			err:      &APIError{Code: ErrorCodeNotFound},
			expected: true,
		},
		{
			name:     "APIError other error",
			err:      &APIError{Code: ErrorCodeNotAuthenticated},
			expected: false,
		},
		{
			name: "ErrorResponse with not found",
			err: &ErrorResponse{
				Errors: []APIError{
					{Code: ErrorCodeNotFound},
				},
			},
			expected: true,
		},
		{
			name: "ErrorResponse without not found",
			err: &ErrorResponse{
				Errors: []APIError{
					{Code: ErrorCodeNotAuthenticated},
				},
			},
			expected: false,
		},
		{
			name:     "ErrorResponse empty",
			err:      &ErrorResponse{},
			expected: false,
		},
		{
			name:     "other error type",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsNotFound(tt.err))
		})
	}
}

func TestIsUnauthorized(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "APIError unauthorized",
			err:      &APIError{Code: ErrorCodeNotAuthenticated},
			expected: true,
		},
		{
			name:     "APIError other error",
			err:      &APIError{Code: ErrorCodeNotFound},
			expected: false,
		},
		{
			name: "ErrorResponse with unauthorized",
			err: &ErrorResponse{
				Errors: []APIError{
					{Code: ErrorCodeNotAuthenticated},
				},
			},
			expected: true,
		},
		{
			name: "ErrorResponse without unauthorized",
			err: &ErrorResponse{
				Errors: []APIError{
					{Code: ErrorCodeNotFound},
				},
			},
			expected: false,
		},
		{
			name:     "other error type",
			err:      errors.New("some error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsUnauthorized(tt.err))
		})
	}
}

func TestIsForbidden(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "APIError forbidden",
			err:      &APIError{Code: ErrorCodeNotAuthorized},
			expected: true,
		},
		{
			name:     "APIError other error",
			err:      &APIError{Code: ErrorCodeNotFound},
			expected: false,
		},
		{
			name: "ErrorResponse with forbidden",
			err: &ErrorResponse{
				Errors: []APIError{
					{Code: ErrorCodeNotAuthorized},
				},
			},
			expected: true,
		},
		{
			name: "ErrorResponse without forbidden",
			err: &ErrorResponse{
				Errors: []APIError{
					{Code: ErrorCodeNotFound},
				},
			},
			expected: false,
		},
		{
			name:     "other error type",
			err:      errors.New("some error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsForbidden(tt.err))
		})
	}
}

func TestParseErrorResponse(t *testing.T) {
	t.Run("valid error response", func(t *testing.T) {
		jsonData := `{
			"errors": [
				{
					"code": 10010,
					"title": "CF-ResourceNotFound",
					"detail": "App not found"
				},
				{
					"code": 10008,
					"title": "CF-UnprocessableEntity",
					"detail": "Invalid request"
				}
			]
		}`

		errResp, err := ParseErrorResponse([]byte(jsonData))
		require.NoError(t, err)
		require.NotNil(t, errResp)
		assert.Len(t, errResp.Errors, 2)
		assert.Equal(t, 10010, errResp.Errors[0].Code)
		assert.Equal(t, "CF-ResourceNotFound", errResp.Errors[0].Title)
		assert.Equal(t, "App not found", errResp.Errors[0].Detail)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		jsonData := `{invalid json}`

		errResp, err := ParseErrorResponse([]byte(jsonData))
		assert.Error(t, err)
		assert.Nil(t, errResp)
	})

	t.Run("empty response", func(t *testing.T) {
		jsonData := `{"errors": []}`

		errResp, err := ParseErrorResponse([]byte(jsonData))
		require.NoError(t, err)
		require.NotNil(t, errResp)
		assert.Len(t, errResp.Errors, 0)
	})
}

func TestErrorConstants(t *testing.T) {
	assert.Equal(t, 10010, ErrNotFound.Code)
	assert.Equal(t, "CF-ResourceNotFound", ErrNotFound.Title)

	assert.Equal(t, 10002, ErrUnauthorized.Code)
	assert.Equal(t, "CF-NotAuthenticated", ErrUnauthorized.Title)

	assert.Equal(t, 10003, ErrForbidden.Code)
	assert.Equal(t, "CF-NotAuthorized", ErrForbidden.Title)

	assert.Equal(t, 10008, ErrUnprocessable.Code)
	assert.Equal(t, "CF-UnprocessableEntity", ErrUnprocessable.Title)

	assert.Equal(t, 10001, ErrServiceUnavailable.Code)
	assert.Equal(t, "CF-ServiceUnavailable", ErrServiceUnavailable.Title)

	assert.Equal(t, 10005, ErrBadRequest.Code)
	assert.Equal(t, "CF-BadRequest", ErrBadRequest.Title)

	assert.Equal(t, 10013, ErrTooManyRequests.Code)
	assert.Equal(t, "CF-TooManyRequests", ErrTooManyRequests.Title)
}

func TestErrorResponse_JSONMarshaling(t *testing.T) {
	errResp := &ErrorResponse{
		Errors: []APIError{
			{
				Code:   10010,
				Title:  "CF-ResourceNotFound",
				Detail: "The app could not be found: test-app",
			},
		},
	}

	data, err := json.Marshal(errResp)
	require.NoError(t, err)

	var decoded ErrorResponse
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Len(t, decoded.Errors, 1)
	assert.Equal(t, errResp.Errors[0].Code, decoded.Errors[0].Code)
	assert.Equal(t, errResp.Errors[0].Title, decoded.Errors[0].Title)
	assert.Equal(t, errResp.Errors[0].Detail, decoded.Errors[0].Detail)
}
