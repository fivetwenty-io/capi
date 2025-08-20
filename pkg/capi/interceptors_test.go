package capi_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/fivetwenty-io/capi-client/pkg/capi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInterceptorChain_RequestInterceptors(t *testing.T) {
	chain := capi.NewInterceptorChain()
	ctx := context.Background()

	var executionOrder []string

	// Add multiple interceptors
	chain.AddRequestInterceptor(func(ctx context.Context, req *capi.Request) error {
		executionOrder = append(executionOrder, "first")
		return nil
	})

	chain.AddRequestInterceptor(func(ctx context.Context, req *capi.Request) error {
		executionOrder = append(executionOrder, "second")
		return nil
	})

	req := &capi.Request{
		Method: "GET",
		Path:   "/test",
	}

	err := chain.ExecuteRequestInterceptors(ctx, req)
	require.NoError(t, err)

	assert.Equal(t, []string{"first", "second"}, executionOrder)
}

func TestInterceptorChain_ResponseInterceptors(t *testing.T) {
	chain := capi.NewInterceptorChain()
	ctx := context.Background()

	var executionOrder []string

	// Add multiple interceptors
	chain.AddResponseInterceptor(func(ctx context.Context, req *capi.Request, resp *capi.Response) error {
		executionOrder = append(executionOrder, "first")
		return nil
	})

	chain.AddResponseInterceptor(func(ctx context.Context, req *capi.Request, resp *capi.Response) error {
		executionOrder = append(executionOrder, "second")
		return nil
	})

	req := &capi.Request{
		Method: "GET",
		Path:   "/test",
	}
	resp := &capi.Response{
		StatusCode: 200,
	}

	err := chain.ExecuteResponseInterceptors(ctx, req, resp)
	require.NoError(t, err)

	assert.Equal(t, []string{"first", "second"}, executionOrder)
}

func TestHeaderInterceptor(t *testing.T) {
	headers := map[string]string{
		"X-Custom-Header": "custom-value",
		"X-Request-ID":    "123456",
	}

	interceptor := capi.HeaderInterceptor(headers)
	ctx := context.Background()
	req := &capi.Request{
		Method: "GET",
		Path:   "/test",
	}

	err := interceptor(ctx, req)
	require.NoError(t, err)

	assert.Equal(t, "custom-value", req.Headers.Get("X-Custom-Header"))
	assert.Equal(t, "123456", req.Headers.Get("X-Request-ID"))
}

func TestAuthenticationInterceptor(t *testing.T) {
	tokenProvider := func(ctx context.Context) (string, error) {
		return "test-token", nil
	}

	interceptor := capi.AuthenticationInterceptor(tokenProvider)
	ctx := context.Background()
	req := &capi.Request{
		Method: "GET",
		Path:   "/test",
	}

	err := interceptor(ctx, req)
	require.NoError(t, err)

	assert.Equal(t, "Bearer test-token", req.Headers.Get("Authorization"))
}

func TestTimeoutInterceptor(t *testing.T) {
	interceptor := capi.TimeoutInterceptor(5 * time.Second)
	ctx := context.Background()
	req := &capi.Request{
		Method:  "GET",
		Path:    "/test",
		Context: ctx,
	}

	err := interceptor(ctx, req)
	require.NoError(t, err)

	// Check that the context has a deadline
	deadline, ok := req.Context.Deadline()
	assert.True(t, ok)
	assert.True(t, deadline.After(time.Now()))
}

func TestMetricsCollector(t *testing.T) {
	collector := capi.NewMetricsCollector()

	var notifiedEndpoint string
	var notifiedMetrics *capi.Metrics

	collector.SetOnChange(func(endpoint string, metrics *capi.Metrics) {
		notifiedEndpoint = endpoint
		notifiedMetrics = metrics
	})

	// Set up interceptors
	requestInterceptor := capi.MetricsRequestInterceptor(collector)
	responseInterceptor := capi.MetricsResponseInterceptor(collector)

	ctx := context.Background()
	req := &capi.Request{
		Method:  "GET",
		Path:    "/v3/apps",
		Context: ctx,
	}

	// Execute request interceptor
	err := requestInterceptor(ctx, req)
	require.NoError(t, err)

	// Simulate some delay
	time.Sleep(10 * time.Millisecond)

	// Execute response interceptor with success
	resp := &capi.Response{
		StatusCode: 200,
	}
	err = responseInterceptor(ctx, req, resp)
	require.NoError(t, err)

	// Check metrics
	assert.Equal(t, "GET /v3/apps", notifiedEndpoint)
	assert.NotNil(t, notifiedMetrics)
	assert.Equal(t, int64(1), notifiedMetrics.TotalRequests)
	assert.Equal(t, int64(0), notifiedMetrics.TotalErrors)
	assert.True(t, notifiedMetrics.AverageLatency > 0)

	// Execute another request with error
	// Note: For the second request we need to test with a pre-set start time
	// but we can't access the internal contextKeyStartTime, so we'll just
	// test without it for now
	req2 := &capi.Request{
		Method:  "GET",
		Path:    "/v3/apps",
		Context: ctx,
	}
	resp2 := &capi.Response{
		StatusCode: 500,
	}
	err = responseInterceptor(ctx, req2, resp2)
	require.NoError(t, err)

	// Check updated metrics
	metrics := collector.GetMetrics("GET /v3/apps")
	assert.Equal(t, int64(2), metrics.TotalRequests)
	assert.Equal(t, int64(1), metrics.TotalErrors)
}

func TestCircuitBreaker(t *testing.T) {
	config := &capi.CircuitBreakerConfig{
		Threshold:        2,
		Timeout:          100 * time.Millisecond,
		SuccessThreshold: 1,
	}
	breaker := capi.NewCircuitBreaker(config)

	requestInterceptor := capi.CircuitBreakerRequestInterceptor(breaker)
	responseInterceptor := capi.CircuitBreakerResponseInterceptor(breaker)

	ctx := context.Background()
	req := &capi.Request{
		Method: "GET",
		Path:   "/test",
	}

	// Circuit should be closed initially
	err := requestInterceptor(ctx, req)
	require.NoError(t, err)

	// Simulate failures
	for i := 0; i < 2; i++ {
		resp := &capi.Response{StatusCode: 500}
		err = responseInterceptor(ctx, req, resp)
		require.NoError(t, err)
	}

	// Circuit should be open now
	err = requestInterceptor(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circuit breaker is open")

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Circuit should be half-open now
	err = requestInterceptor(ctx, req)
	require.NoError(t, err)

	// Simulate success
	resp := &capi.Response{StatusCode: 200}
	err = responseInterceptor(ctx, req, resp)
	require.NoError(t, err)

	// Circuit should be closed again
	err = requestInterceptor(ctx, req)
	require.NoError(t, err)
}

func TestRetryResponseInterceptor(t *testing.T) {
	config := &capi.RetryConfig{
		MaxRetries:   3,
		RetryDelay:   100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		RetryOnCodes: []int{429, 500, 502, 503, 504},
	}

	interceptor := capi.RetryResponseInterceptor(config)
	ctx := context.Background()
	req := &capi.Request{
		Method: "GET",
		Path:   "/test",
	}

	// Test retryable status code
	resp := &capi.Response{
		StatusCode: 500,
		Headers:    make(http.Header),
	}

	err := interceptor(ctx, req, resp)
	require.NoError(t, err)
	assert.Equal(t, "true", resp.Headers.Get("X-Should-Retry"))

	// Test non-retryable status code
	resp2 := &capi.Response{
		StatusCode: 404,
		Headers:    make(http.Header),
	}

	err = interceptor(ctx, req, resp2)
	require.NoError(t, err)
	assert.Equal(t, "", resp2.Headers.Get("X-Should-Retry"))
}
