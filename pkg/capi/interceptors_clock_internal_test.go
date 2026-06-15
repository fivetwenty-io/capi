package capi

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCircuitBreakerTimeout_Deterministic drives the circuit breaker's
// closed -> open -> half-open -> closed lifecycle using a controllable clock
// instead of time.Sleep, so the timeout transition is exercised exactly and
// without wall-clock flakiness. It complements the sleep-based TestCircuitBreaker
// in the external test package.
func TestCircuitBreakerTimeout_Deterministic(t *testing.T) {
	t.Parallel()

	// Controllable clock shared with the breaker via the unexported seam.
	current := time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)
	breaker := NewCircuitBreaker(&CircuitBreakerConfig{
		Threshold:        2,
		Timeout:          100 * time.Millisecond,
		SuccessThreshold: 1,
	})
	breaker.now = func() time.Time { return current }

	requestInterceptor := CircuitBreakerRequestInterceptor(breaker)
	responseInterceptor := CircuitBreakerResponseInterceptor(breaker)

	ctx := context.Background()
	req := &Request{Method: "GET", Path: "/test"}

	// Closed initially.
	require.NoError(t, requestInterceptor(ctx, req))

	// Two failures reach the threshold and open the circuit.
	for range 2 {
		require.NoError(t, responseInterceptor(ctx, req, &Response{StatusCode: 500}))
	}

	// Open: requests are rejected while still inside the timeout window.
	current = current.Add(50 * time.Millisecond)

	require.ErrorIs(t, requestInterceptor(ctx, req), ErrCircuitBreakerOpen)

	// Advancing past the timeout flips the circuit to half-open and admits one.
	current = current.Add(60 * time.Millisecond) // 110ms total > 100ms timeout

	require.NoError(t, requestInterceptor(ctx, req))

	// A success in half-open (SuccessThreshold == 1) closes the circuit again.
	require.NoError(t, responseInterceptor(ctx, req, &Response{StatusCode: 200}))
	assert.Equal(t, "closed", breaker.state)

	// Closed again: subsequent requests pass.
	require.NoError(t, requestInterceptor(ctx, req))
}
