package capi_test

import (
	"context"
	"testing"
	"time"

	"github.com/fivetwenty-io/capi/v3/pkg/capi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryCache_SetAndGet(t *testing.T) {
	t.Parallel()

	cache := capi.NewMemoryCache(10)
	ctx := context.Background()

	entry := &capi.CacheEntry{
		Data:      []byte("test data"),
		ExpiresAt: time.Now().Add(1 * time.Hour),
		ETag:      "abc123",
	}

	// Set entry
	err := cache.Set(ctx, "key1", entry)
	require.NoError(t, err)

	// Get entry
	retrieved, err := cache.Get(ctx, "key1")
	require.NoError(t, err)
	assert.Equal(t, entry.Data, retrieved.Data)
	assert.Equal(t, entry.ETag, retrieved.ETag)
}

func TestMemoryCache_GetNonExistent(t *testing.T) {
	t.Parallel()

	cache := capi.NewMemoryCache(10)
	ctx := context.Background()

	_, err := cache.Get(ctx, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "key not found")
}

func TestMemoryCache_GetExpired(t *testing.T) {
	t.Parallel()

	cache := capi.NewMemoryCache(10)
	ctx := context.Background()

	entry := &capi.CacheEntry{
		Data:      []byte("test data"),
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Already expired
		ETag:      "abc123",
	}

	err := cache.Set(ctx, "key1", entry)
	require.NoError(t, err)

	_, err = cache.Get(ctx, "key1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "entry expired")
}

func TestMemoryCache_Delete(t *testing.T) {
	t.Parallel()

	cache := capi.NewMemoryCache(10)
	ctx := context.Background()

	entry := &capi.CacheEntry{
		Data:      []byte("test data"),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	// Set and verify
	err := cache.Set(ctx, "key1", entry)
	require.NoError(t, err)
	assert.True(t, cache.Has(ctx, "key1"))

	// Delete
	err = cache.Delete(ctx, "key1")
	require.NoError(t, err)

	// Verify deleted
	assert.False(t, cache.Has(ctx, "key1"))
}

func TestMemoryCache_Clear(t *testing.T) {
	t.Parallel()

	cache := capi.NewMemoryCache(10)
	ctx := context.Background()

	// Add multiple entries
	for i := range 3 {
		entry := &capi.CacheEntry{
			Data:      []byte("test data"),
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}
		_ = cache.Set(ctx, string(rune('a'+i)), entry)
	}

	// Verify entries exist
	assert.True(t, cache.Has(ctx, "a"))
	assert.True(t, cache.Has(ctx, "b"))
	assert.True(t, cache.Has(ctx, "c"))

	// Clear cache
	err := cache.Clear(ctx)
	require.NoError(t, err)

	// Verify all cleared
	assert.False(t, cache.Has(ctx, "a"))
	assert.False(t, cache.Has(ctx, "b"))
	assert.False(t, cache.Has(ctx, "c"))
}

func TestMemoryCache_MaxSize(t *testing.T) {
	t.Parallel()

	cache := capi.NewMemoryCache(2)
	ctx := context.Background()

	// Add entries up to max size
	for i := range 3 {
		entry := &capi.CacheEntry{
			Data:      []byte("test data"),
			ExpiresAt: time.Now().Add(time.Duration(i+1) * time.Hour),
		}
		_ = cache.Set(ctx, string(rune('a'+i)), entry)
	}

	// The cache should have evicted the oldest entry
	// Since we can't easily check internal state, we verify behavior
	has := 0

	for i := range 3 {
		if cache.Has(ctx, string(rune('a'+i))) {
			has++
		}
	}

	assert.LessOrEqual(t, has, 2)
}

func TestMemoryCache_Cleanup(t *testing.T) {
	t.Parallel()

	cache := capi.NewMemoryCache(10)
	ctx := context.Background()

	// Add expired and non-expired entries
	expiredEntry := &capi.CacheEntry{
		Data:      []byte("expired"),
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	validEntry := &capi.CacheEntry{
		Data:      []byte("valid"),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	_ = cache.Set(ctx, "expired", expiredEntry)
	_ = cache.Set(ctx, "valid", validEntry)

	// Run cleanup
	cache.Cleanup()

	// Valid entry should still exist
	assert.True(t, cache.Has(ctx, "valid"))
	// Expired entry should be gone
	assert.False(t, cache.Has(ctx, "expired"))
}

func TestCacheManager_GetCacheKey(t *testing.T) {
	t.Parallel()

	manager := capi.NewCacheManager(nil, nil)

	// Test with no params
	key1 := manager.GetCacheKey("GET", "/v3/apps", nil)
	assert.Equal(t, "GET:/v3/apps", key1)

	// Test with params
	params := map[string]string{"page": "1", "per_page": "50"}
	key2 := manager.GetCacheKey("GET", "/v3/apps", params)
	assert.Contains(t, key2, "GET:/v3/apps:")
	assert.Contains(t, key2, "page")
	assert.Contains(t, key2, "per_page")
}

func TestCacheManager_SetAndGet(t *testing.T) {
	t.Parallel()

	cache := capi.NewMemoryCache(10)
	manager := capi.NewCacheManager(cache, nil)
	ctx := context.Background()

	data := []byte("test data")
	key := "test-key"

	// Set data
	err := manager.Set(ctx, key, data, 1*time.Hour)
	require.NoError(t, err)

	// Get data
	retrieved, err := manager.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, data, retrieved)

	// Check stats
	stats := manager.GetStats()
	assert.Equal(t, int64(1), stats.Hits)
	assert.Equal(t, int64(0), stats.Misses)
	assert.Equal(t, int64(1), stats.Sets)
}

func TestCacheManager_SetWithETag(t *testing.T) {
	t.Parallel()

	cache := capi.NewMemoryCache(10)
	manager := capi.NewCacheManager(cache, nil)
	ctx := context.Background()

	data := []byte("test data")
	key := "test-key"
	etag := "abc123"

	// Set data with ETag
	err := manager.SetWithETag(ctx, key, data, etag, 1*time.Hour)
	require.NoError(t, err)

	// Get data
	retrieved, err := manager.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, data, retrieved)
}

func TestCacheManager_Miss(t *testing.T) {
	t.Parallel()

	cache := capi.NewMemoryCache(10)
	manager := capi.NewCacheManager(cache, nil)
	ctx := context.Background()

	// Try to get non-existent key
	_, err := manager.Get(ctx, "nonexistent")
	require.Error(t, err)

	// Check stats
	stats := manager.GetStats()
	assert.Equal(t, int64(0), stats.Hits)
	assert.Equal(t, int64(1), stats.Misses)
}

func TestCacheStats_GetHitRate(t *testing.T) {
	t.Parallel()

	stats := &capi.CacheStats{
		Hits:   75,
		Misses: 25,
	}

	hitRate := stats.GetHitRate()
	assert.InDelta(t, 0.75, hitRate, 0.0001)

	// Test with no requests
	emptyStats := &capi.CacheStats{}
	assert.InDelta(t, 0.0, emptyStats.GetHitRate(), 0.0001)
}

func TestCachingPolicy_ShouldCache(t *testing.T) {
	t.Parallel()

	policy := capi.DefaultCachingPolicy()

	// Test GET requests (should cache)
	assert.True(t, policy.ShouldCache("GET", "/v3/apps", 200))

	// Test POST requests (should not cache by default)
	assert.False(t, policy.ShouldCache("POST", "/v3/apps", 201))

	// Test error responses (should not cache by default)
	assert.False(t, policy.ShouldCache("GET", "/v3/apps", 404))

	// Test excluded paths
	assert.False(t, policy.ShouldCache("GET", "/v3/jobs", 200))

	// Test with custom policy
	customPolicy := &capi.CachingPolicy{
		CacheGET:     true,
		CachePOST:    true,
		CacheErrors:  true,
		IncludePaths: []string{"/v3/apps"},
	}

	// Only included paths should be cached
	assert.True(t, customPolicy.ShouldCache("GET", "/v3/apps", 200))
	assert.False(t, customPolicy.ShouldCache("GET", "/v3/spaces", 200))

	// POST should be cached with custom policy
	assert.True(t, customPolicy.ShouldCache("POST", "/v3/apps", 201))

	// Errors should be cached with custom policy
	assert.True(t, customPolicy.ShouldCache("GET", "/v3/apps", 404))
}
