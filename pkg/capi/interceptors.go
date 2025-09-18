package capi

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/fivetwenty-io/capi/v3/internal/constants"
)

// Request represents an HTTP request that can be intercepted.
type Request struct {
	Method   string
	Path     string
	Headers  http.Header
	Body     []byte
	Metadata map[string]interface{}
}

// Response represents an HTTP response that can be intercepted.
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	Error      error
}

// RequestInterceptor is called before a request is sent.
type RequestInterceptor func(ctx context.Context, req *Request) error

// ResponseInterceptor is called after a response is received.
type ResponseInterceptor func(ctx context.Context, req *Request, resp *Response) error

// InterceptorChain manages a chain of interceptors.
type InterceptorChain struct {
	requestInterceptors  []RequestInterceptor
	responseInterceptors []ResponseInterceptor
}

// NewInterceptorChain creates a new interceptor chain.
func NewInterceptorChain() *InterceptorChain {
	return &InterceptorChain{
		requestInterceptors:  make([]RequestInterceptor, 0),
		responseInterceptors: make([]ResponseInterceptor, 0),
	}
}

// AddRequestInterceptor adds a request interceptor to the chain.
func (c *InterceptorChain) AddRequestInterceptor(interceptor RequestInterceptor) {
	c.requestInterceptors = append(c.requestInterceptors, interceptor)
}

// AddResponseInterceptor adds a response interceptor to the chain.
func (c *InterceptorChain) AddResponseInterceptor(interceptor ResponseInterceptor) {
	c.responseInterceptors = append(c.responseInterceptors, interceptor)
}

// ExecuteRequestInterceptors runs all request interceptors.
func (c *InterceptorChain) ExecuteRequestInterceptors(ctx context.Context, req *Request) error {
	for _, interceptor := range c.requestInterceptors {
		err := interceptor(ctx, req)
		if err != nil {
			return fmt.Errorf("request interceptor failed: %w", err)
		}
	}

	return nil
}

// ExecuteResponseInterceptors runs all response interceptors.
func (c *InterceptorChain) ExecuteResponseInterceptors(ctx context.Context, req *Request, resp *Response) error {
	for _, interceptor := range c.responseInterceptors {
		err := interceptor(ctx, req, resp)
		if err != nil {
			return fmt.Errorf("response interceptor failed: %w", err)
		}
	}

	return nil
}

// Common Interceptors

// LoggingInterceptor logs requests and responses.
func LoggingInterceptor(logger Logger) RequestInterceptor {
	return func(ctx context.Context, req *Request) error {
		logger.Debug("API Request", map[string]interface{}{
			"method": req.Method,
			"path":   req.Path,
		})

		return nil
	}
}

// LoggingResponseInterceptor logs responses.
func LoggingResponseInterceptor(logger Logger) ResponseInterceptor {
	return func(ctx context.Context, req *Request, resp *Response) error {
		fields := map[string]interface{}{
			"method":      req.Method,
			"path":        req.Path,
			"status_code": resp.StatusCode,
		}

		if resp.Error != nil {
			logger.Error("API Response Error", fields)
		} else {
			logger.Debug("API Response", fields)
		}

		return nil
	}
}

// RateLimitInterceptor implements client-side rate limiting.
func RateLimitInterceptor(requestsPerSecond int) RequestInterceptor {
	// Simple token bucket implementation
	bucket := make(chan struct{}, requestsPerSecond)

	// Fill the bucket initially
	for range requestsPerSecond {
		bucket <- struct{}{}
	}

	// Refill the bucket periodically
	go func() {
		ticker := time.NewTicker(time.Second / time.Duration(requestsPerSecond))
		defer ticker.Stop()

		for range ticker.C {
			select {
			case bucket <- struct{}{}:
			default:
				// Bucket is full
			}
		}
	}()

	return func(ctx context.Context, req *Request) error {
		select {
		case <-bucket:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// RetryInterceptor adds retry logic for failed requests.
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts after the initial try.
	MaxRetries int
	// RetryDelay is the base delay between retries (may be jittered/backed off).
	RetryDelay time.Duration
	// MaxDelay caps the maximum backoff delay between retries.
	MaxDelay time.Duration
	// RetryOnCodes lists HTTP status codes that should trigger a retry
	// (e.g., 429, 500, 502, 503, 504).
	RetryOnCodes []int
}

// DefaultRetryConfig returns default retry configuration.
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:   constants.LowRetryMax,
		RetryDelay:   1 * time.Second,
		MaxDelay:     constants.ExtendedRetryWaitMax,
		RetryOnCodes: []int{429, 500, 502, 503, 504},
	}
}

// RetryResponseInterceptor implements retry logic.
func RetryResponseInterceptor(config *RetryConfig) ResponseInterceptor {
	if config == nil {
		config = DefaultRetryConfig()
	}

	return func(ctx context.Context, req *Request, resp *Response) error {
		// Check if we should retry based on status code
		shouldRetry := false

		for _, code := range config.RetryOnCodes {
			if resp.StatusCode == code {
				shouldRetry = true

				break
			}
		}

		if !shouldRetry {
			return nil
		}

		// Set a retry marker in the response
		// The actual retry logic would be implemented in the HTTP client
		if resp.Headers == nil {
			resp.Headers = make(http.Header)
		}

		resp.Headers.Set("X-Should-Retry", "true")

		return nil
	}
}

// AuthenticationInterceptor adds authentication headers.
func AuthenticationInterceptor(tokenProvider func(context.Context) (string, error)) RequestInterceptor {
	return func(ctx context.Context, req *Request) error {
		token, err := tokenProvider(ctx)
		if err != nil {
			return fmt.Errorf("failed to get authentication token: %w", err)
		}

		if req.Headers == nil {
			req.Headers = make(http.Header)
		}

		req.Headers.Set("Authorization", "Bearer "+token)

		return nil
	}
}

// HeaderInterceptor adds custom headers to requests.
func HeaderInterceptor(headers map[string]string) RequestInterceptor {
	return func(ctx context.Context, req *Request) error {
		if req.Headers == nil {
			req.Headers = make(http.Header)
		}

		for key, value := range headers {
			req.Headers.Set(key, value)
		}

		return nil
	}
}

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

// TimeoutInterceptor adds a timeout to requests.
func TimeoutInterceptor(timeout time.Duration) RequestInterceptor {
	return func(ctx context.Context, req *Request) error {
		// The timeout context should be handled by the caller
		// This interceptor can validate timeout or prepare for it
		// but the actual timeout context should be created by the HTTP client
		return nil
	}
}

// MetricsInterceptor collects metrics about API calls.
type Metrics struct {
	TotalRequests   int64
	TotalErrors     int64
	TotalLatency    time.Duration
	AverageLatency  time.Duration
	LastRequestTime time.Time
}

// MetricsCollector collects API metrics.
type MetricsCollector struct {
	metrics  map[string]*Metrics
	onChange func(endpoint string, metrics *Metrics)
}

// NewMetricsCollector creates a new metrics collector.
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics: make(map[string]*Metrics),
	}
}

// SetOnChange sets a callback for when metrics change.
func (m *MetricsCollector) SetOnChange(fn func(endpoint string, metrics *Metrics)) {
	m.onChange = fn
}

// GetMetrics returns metrics for an endpoint.
func (m *MetricsCollector) GetMetrics(endpoint string) *Metrics {
	if metrics, ok := m.metrics[endpoint]; ok {
		return metrics
	}

	return nil
}

// MetricsRequestInterceptor records request start time.
func MetricsRequestInterceptor(collector *MetricsCollector) RequestInterceptor {
	return func(ctx context.Context, req *Request) error {
		// Store the start time in the request metadata
		if req.Metadata == nil {
			req.Metadata = make(map[string]interface{})
		}

		req.Metadata["start_time"] = time.Now()

		return nil
	}
}

// MetricsResponseInterceptor records response metrics.
func MetricsResponseInterceptor(collector *MetricsCollector) ResponseInterceptor {
	return func(ctx context.Context, req *Request, resp *Response) error {
		endpoint := fmt.Sprintf("%s %s", req.Method, req.Path)

		// Get or create metrics for this endpoint
		metrics, ok := collector.metrics[endpoint]
		if !ok {
			metrics = &Metrics{}
			collector.metrics[endpoint] = metrics
		}

		// Update metrics
		metrics.TotalRequests++
		metrics.LastRequestTime = time.Now()

		// Calculate latency if start time is available in request metadata
		if req.Metadata != nil {
			if startTime, ok := req.Metadata["start_time"].(time.Time); ok {
				latency := time.Since(startTime)
				metrics.TotalLatency += latency
				metrics.AverageLatency = metrics.TotalLatency / time.Duration(metrics.TotalRequests)
			}
		}

		// Count errors
		if resp.Error != nil || resp.StatusCode >= 400 {
			metrics.TotalErrors++
		}

		// Notify listener if set
		if collector.onChange != nil {
			collector.onChange(endpoint, metrics)
		}

		return nil
	}
}

// CircuitBreakerInterceptor implements circuit breaker pattern.
type CircuitBreakerConfig struct {
	Threshold        int           // Number of failures before opening
	Timeout          time.Duration // Time before trying again
	SuccessThreshold int           // Number of successes to close
}

// CircuitBreaker tracks circuit state.
type CircuitBreaker struct {
	config      *CircuitBreakerConfig
	failures    int
	successes   int
	state       string // "closed", constants.StatusOpen, constants.StatusHalfOpen
	lastFailure time.Time
}

// NewCircuitBreaker creates a new circuit breaker.
func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
	if config == nil {
		config = &CircuitBreakerConfig{
			Threshold:        constants.CircuitBreakerThreshold,
			Timeout:          constants.CircuitBreakerTimeout,
			SuccessThreshold: constants.CircuitBreakerSuccessThreshold,
		}
	}

	return &CircuitBreaker{
		config: config,
		state:  "closed",
	}
}

// CircuitBreakerRequestInterceptor checks circuit state before requests.
func CircuitBreakerRequestInterceptor(breaker *CircuitBreaker) RequestInterceptor {
	return func(ctx context.Context, req *Request) error {
		if breaker.state == constants.StatusOpen {
			// Check if timeout has passed
			if time.Since(breaker.lastFailure) > breaker.config.Timeout {
				breaker.state = constants.StatusHalfOpen
				breaker.successes = 0
			} else {
				return ErrCircuitBreakerOpen
			}
		}

		return nil
	}
}

// CircuitBreakerResponseInterceptor updates circuit state based on responses.
func CircuitBreakerResponseInterceptor(breaker *CircuitBreaker) ResponseInterceptor {
	return func(ctx context.Context, req *Request, resp *Response) error {
		if resp.Error != nil || resp.StatusCode >= 500 {
			// Record failure
			breaker.failures++
			breaker.lastFailure = time.Now()

			if breaker.failures >= breaker.config.Threshold {
				breaker.state = constants.StatusOpen
			}

			if breaker.state == constants.StatusHalfOpen {
				breaker.state = constants.StatusOpen
			}
		} else {
			// Record success
			switch breaker.state {
			case constants.StatusHalfOpen:
				breaker.successes++
				if breaker.successes >= breaker.config.SuccessThreshold {
					breaker.state = "closed"
					breaker.failures = 0
				}
			case "closed":
				breaker.failures = 0
			}
		}

		return nil
	}
}
