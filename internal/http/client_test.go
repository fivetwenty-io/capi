package http_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	capihttp "github.com/fivetwenty-io/capi/v3/internal/http"
	"github.com/fivetwenty-io/capi/v3/pkg/capi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockTokenManager for testing.
type MockTokenManager struct {
	token string
	err   error
}

func (m *MockTokenManager) GetToken(ctx context.Context) (string, error) {
	return m.token, m.err
}

func (m *MockTokenManager) RefreshToken(ctx context.Context) error {
	return nil
}

func (m *MockTokenManager) SetToken(token string, expiresAt time.Time) {
	m.token = token
}

// MockLogger for testing.
type MockLogger struct {
	logs []map[string]interface{}
}

func (l *MockLogger) Debug(msg string, fields map[string]interface{}) {
	l.logs = append(l.logs, map[string]interface{}{"level": "debug", "msg": msg, "fields": fields})
}

func (l *MockLogger) Info(msg string, fields map[string]interface{}) {
	l.logs = append(l.logs, map[string]interface{}{"level": "info", "msg": msg, "fields": fields})
}

func (l *MockLogger) Warn(msg string, fields map[string]interface{}) {
	l.logs = append(l.logs, map[string]interface{}{"level": "warn", "msg": msg, "fields": fields})
}

func (l *MockLogger) Error(msg string, fields map[string]interface{}) {
	l.logs = append(l.logs, map[string]interface{}{"level": "error", "msg": msg, "fields": fields})
}

//nolint:funlen // Test functions can be longer for comprehensive testing
func TestClient_Do(t *testing.T) {
	t.Parallel()
	t.Run("successful request", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			assert.Equal(t, "/v3/apps", request.URL.Path)
			assert.Equal(t, "GET", request.Method)
			assert.Equal(t, "Bearer test-token", request.Header.Get("Authorization"))
			assert.Equal(t, "application/json", request.Header.Get("Accept"))

			response := map[string]string{"guid": "app-guid", "name": "test-app"}
			_ = json.NewEncoder(writer).Encode(response)
		}))
		defer server.Close()

		tokenManager := &MockTokenManager{token: "test-token"}
		client := capihttp.NewClient(server.URL, tokenManager)

		req := &capihttp.Request{
			Method: "GET",
			Path:   "/v3/apps",
		}

		resp, err := client.Do(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]string

		err = json.Unmarshal(resp.Body, &result)
		require.NoError(t, err)
		assert.Equal(t, "app-guid", result["guid"])
		assert.Equal(t, "test-app", result["name"])
	})

	t.Run("request with query parameters", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			assert.Equal(t, "/v3/apps", request.URL.Path)
			assert.Equal(t, "page=2", request.URL.RawQuery)
			writer.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := capihttp.NewClient(server.URL, nil)

		req := &capihttp.Request{
			Method: "GET",
			Path:   "/v3/apps",
			Query:  url.Values{"page": []string{"2"}},
		}

		resp, err := client.Do(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("request with body", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			assert.Equal(t, "POST", request.Method)
			assert.Equal(t, "application/json", request.Header.Get("Content-Type"))

			var body map[string]string

			_ = json.NewDecoder(request.Body).Decode(&body)
			assert.Equal(t, "test-app", body["name"])

			writer.WriteHeader(http.StatusCreated)
		}))
		defer server.Close()

		client := capihttp.NewClient(server.URL, nil)

		req := &capihttp.Request{
			Method: "POST",
			Path:   "/v3/apps",
			Body:   map[string]string{"name": "test-app"},
		}

		resp, err := client.Do(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode)
	})

	t.Run("error response", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusNotFound)

			response := capi.ResponseError{
				Errors: []capi.APIError{
					{
						Code:   10010,
						Title:  "CF-ResourceNotFound",
						Detail: "App not found",
					},
				},
			}
			_ = json.NewEncoder(writer).Encode(response)
		}))
		defer server.Close()

		client := capihttp.NewClient(server.URL, nil)

		req := &capihttp.Request{
			Method: "GET",
			Path:   "/v3/apps/invalid",
		}

		resp, err := client.Do(context.Background(), req)
		require.Error(t, err)
		assert.Equal(t, 404, resp.StatusCode)

		errResp := &capi.ResponseError{}
		ok := errors.As(err, &errResp)
		require.True(t, ok)
		assert.Len(t, errResp.Errors, 1)
		assert.Equal(t, 10010, errResp.Errors[0].Code)
	})

	t.Run("custom headers", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			assert.Equal(t, "custom-value", request.Header.Get("X-Custom-Header"))
			writer.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := capihttp.NewClient(server.URL, nil)

		req := &capihttp.Request{
			Method: "GET",
			Path:   "/v3/apps",
			Headers: map[string]string{
				"X-Custom-Header": "custom-value",
			},
		}

		resp, err := client.Do(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("with debug logging", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(writer).Encode(map[string]string{"result": "ok"})
		}))
		defer server.Close()

		logger := &MockLogger{}
		client := capihttp.NewClient(server.URL, nil, capihttp.WithLogger(logger), capihttp.WithDebug(true))

		req := &capihttp.Request{
			Method: "GET",
			Path:   "/v3/apps",
		}

		_, err := client.Do(context.Background(), req)
		require.NoError(t, err)

		// Should have logged request and response
		assert.Len(t, logger.logs, 2)
		assert.Equal(t, "HTTP Request", logger.logs[0]["msg"])
		assert.Equal(t, "HTTP Response", logger.logs[1]["msg"])
	})
}

//nolint:funlen // Test functions can be longer for comprehensive testing
func TestClient_Methods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		fn     func(*capihttp.Client, context.Context) (*capihttp.Response, error)
	}{
		{
			name:   "GET",
			method: "GET",
			fn: func(c *capihttp.Client, ctx context.Context) (*capihttp.Response, error) {
				return c.Get(ctx, "/test", nil)
			},
		},
		{
			name:   "POST",
			method: "POST",
			fn: func(c *capihttp.Client, ctx context.Context) (*capihttp.Response, error) {
				return c.Post(ctx, "/test", map[string]string{"key": "value"})
			},
		},
		{
			name:   "PUT",
			method: "PUT",
			fn: func(c *capihttp.Client, ctx context.Context) (*capihttp.Response, error) {
				return c.Put(ctx, "/test", map[string]string{"key": "value"})
			},
		},
		{
			name:   "PATCH",
			method: "PATCH",
			fn: func(c *capihttp.Client, ctx context.Context) (*capihttp.Response, error) {
				return c.Patch(ctx, "/test", map[string]string{"key": "value"})
			},
		},
		{
			name:   "DELETE",
			method: "DELETE",
			fn: func(c *capihttp.Client, ctx context.Context) (*capihttp.Response, error) {
				return c.Delete(ctx, "/test")
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				assert.Equal(t, testCase.method, request.Method)
				assert.Equal(t, "/test", request.URL.Path)
				writer.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			client := capihttp.NewClient(server.URL, nil)
			resp, err := testCase.fn(client, context.Background())
			require.NoError(t, err)
			assert.Equal(t, 200, resp.StatusCode)
		})
	}
}

//nolint:funlen // Test functions can be longer for comprehensive testing
func TestClient_RetryLogic(t *testing.T) {
	t.Parallel()
	t.Run("retries on 5xx errors", func(t *testing.T) {
		t.Parallel()

		attempts := 0

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			attempts++
			if attempts < 3 {
				writer.WriteHeader(http.StatusInternalServerError)
			} else {
				writer.WriteHeader(http.StatusOK)
			}
		}))
		defer server.Close()

		client := capihttp.NewClient(server.URL, nil, capihttp.WithRetryConfig(3, 10*time.Millisecond, 100*time.Millisecond))

		resp, err := client.Get(context.Background(), "/test", nil)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, 3, attempts)
	})

	t.Run("retries on rate limiting", func(t *testing.T) {
		t.Parallel()

		attempts := 0

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			attempts++
			if attempts < 2 {
				writer.WriteHeader(http.StatusTooManyRequests)
			} else {
				writer.WriteHeader(http.StatusOK)
			}
		}))
		defer server.Close()

		client := capihttp.NewClient(server.URL, nil, capihttp.WithRetryConfig(3, 10*time.Millisecond, 100*time.Millisecond))

		resp, err := client.Get(context.Background(), "/test", nil)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, 2, attempts)
	})

	t.Run("does not retry on client errors", func(t *testing.T) {
		t.Parallel()

		attempts := 0

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			attempts++

			writer.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()

		client := capihttp.NewClient(server.URL, nil, capihttp.WithRetryConfig(3, 10*time.Millisecond, 100*time.Millisecond))

		resp, err := client.Get(context.Background(), "/test", nil)
		require.Error(t, err)
		assert.Equal(t, 400, resp.StatusCode)
		assert.Equal(t, 1, attempts) // Should not retry
	})
}
