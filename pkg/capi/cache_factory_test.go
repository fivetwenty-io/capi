package capi_test

import (
	"context"
	"testing"
	"time"

	"github.com/fivetwenty-io/capi/v3/pkg/capi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheFactory_MemoryCache(t *testing.T) {
	config := &capi.CacheConfig{
		Type: capi.CacheTypeMemory,
		Memory: &capi.MemoryCacheConfig{
			MaxSize:         100,
			CleanupInterval: "1m",
		},
	}

	cache, err := capi.NewCacheFromConfig(config)
	require.NoError(t, err)
	require.NotNil(t, cache)

	// Test basic operations
	ctx := context.Background()
	entry := &capi.CacheEntry{
		Data:      []byte("test data"),
		ExpiresAt: time.Now().Add(1 * time.Hour),
		ETag:      "test-etag",
	}

	// Set
	err = cache.Set(ctx, "test-key", entry)
	assert.NoError(t, err)

	// Get
	retrieved, err := cache.Get(ctx, "test-key")
	require.NoError(t, err)
	assert.Equal(t, entry.Data, retrieved.Data)
	assert.Equal(t, entry.ETag, retrieved.ETag)

	// Has
	assert.True(t, cache.Has(ctx, "test-key"))

	// Delete
	err = cache.Delete(ctx, "test-key")
	assert.NoError(t, err)
	assert.False(t, cache.Has(ctx, "test-key"))
}

func TestCacheFactory_NoOpCache(t *testing.T) {
	config := &capi.CacheConfig{
		Type: capi.CacheTypeNone,
	}

	cache, err := capi.NewCacheFromConfig(config)
	require.NoError(t, err)
	require.NotNil(t, cache)

	ctx := context.Background()
	entry := &capi.CacheEntry{
		Data:      []byte("test data"),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	// Set should succeed but do nothing
	err = cache.Set(ctx, "test-key", entry)
	assert.NoError(t, err)

	// Get should always fail
	_, err = cache.Get(ctx, "test-key")
	assert.Error(t, err)

	// Has should always return false
	assert.False(t, cache.Has(ctx, "test-key"))

	// Delete should succeed but do nothing
	err = cache.Delete(ctx, "test-key")
	assert.NoError(t, err)

	// Clear should succeed but do nothing
	err = cache.Clear(ctx)
	assert.NoError(t, err)
}

func TestCacheBuilder(t *testing.T) {
	builder := capi.NewCacheBuilder()
	cache, err := builder.
		WithType(capi.CacheTypeMemory).
		WithMemoryConfig(50, "30s").
		WithOptions(&capi.CacheOptions{
			TTL:         10 * time.Minute,
			MaxSize:     50,
			EnableETags: true,
		}).
		Build()

	require.NoError(t, err)
	require.NotNil(t, cache)

	// Test that the cache works
	ctx := context.Background()
	entry := &capi.CacheEntry{
		Data:      []byte("builder test"),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	err = cache.Set(ctx, "builder-key", entry)
	assert.NoError(t, err)

	retrieved, err := cache.Get(ctx, "builder-key")
	require.NoError(t, err)
	assert.Equal(t, entry.Data, retrieved.Data)
}

func TestCacheChain(t *testing.T) {
	// Create two memory caches
	l1Cache := capi.NewMemoryCache(10)
	l2Cache := capi.NewMemoryCache(100)

	// Create chain
	chain := capi.NewCacheChain(l1Cache, l2Cache)

	ctx := context.Background()
	entry := &capi.CacheEntry{
		Data:      []byte("chain test"),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	// Set should store in both caches
	err := chain.Set(ctx, "chain-key", entry)
	assert.NoError(t, err)

	// Verify both caches have the entry
	assert.True(t, l1Cache.Has(ctx, "chain-key"))
	assert.True(t, l2Cache.Has(ctx, "chain-key"))

	// Delete from L1 only
	err = l1Cache.Delete(ctx, "chain-key")
	assert.NoError(t, err)

	// Get should still work (from L2) and repopulate L1
	retrieved, err := chain.Get(ctx, "chain-key")
	require.NoError(t, err)
	assert.Equal(t, entry.Data, retrieved.Data)

	// L1 should have the entry again
	assert.True(t, l1Cache.Has(ctx, "chain-key"))

	// Delete from chain should delete from both
	err = chain.Delete(ctx, "chain-key")
	assert.NoError(t, err)
	assert.False(t, l1Cache.Has(ctx, "chain-key"))
	assert.False(t, l2Cache.Has(ctx, "chain-key"))
}

func TestDefaultCacheConfig(t *testing.T) {
	config := capi.DefaultCacheConfig()
	assert.Equal(t, capi.CacheTypeMemory, config.Type)
	assert.NotNil(t, config.Memory)
	assert.Equal(t, 1000, config.Memory.MaxSize)
	assert.Equal(t, "1m", config.Memory.CleanupInterval)
	assert.NotNil(t, config.Options)
}

func TestCacheFactory_InvalidType(t *testing.T) {
	config := &capi.CacheConfig{
		Type: capi.CacheType("invalid"),
	}

	cache, err := capi.NewCacheFromConfig(config)
	assert.Error(t, err)
	assert.Nil(t, cache)
	assert.Contains(t, err.Error(), "unsupported cache type")
}

func TestCacheFactory_NilConfig(t *testing.T) {
	cache, err := capi.NewCacheFromConfig(nil)
	require.NoError(t, err)
	require.NotNil(t, cache)

	// Should use default config (memory cache)
	ctx := context.Background()
	entry := &capi.CacheEntry{
		Data:      []byte("default test"),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	err = cache.Set(ctx, "default-key", entry)
	assert.NoError(t, err)

	retrieved, err := cache.Get(ctx, "default-key")
	require.NoError(t, err)
	assert.Equal(t, entry.Data, retrieved.Data)
}
