package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/cloudfoundry/go-uaa"
)

// CacheEntry represents a cached item with expiry
type CacheEntry struct {
	Data      interface{}
	ExpiresAt time.Time
}

// IsExpired checks if the cache entry has expired
func (c *CacheEntry) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// UAACache provides caching functionality for UAA operations
type UAACache struct {
	cache map[string]*CacheEntry
	mutex sync.RWMutex
	ttl   time.Duration
}

// NewUAACache creates a new cache with specified TTL
func NewUAACache(ttl time.Duration) *UAACache {
	cache := &UAACache{
		cache: make(map[string]*CacheEntry),
		mutex: sync.RWMutex{},
		ttl:   ttl,
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Get retrieves an item from cache
func (c *UAACache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.cache[key]
	if !exists || entry.IsExpired() {
		return nil, false
	}

	return entry.Data, true
}

// Set stores an item in cache
func (c *UAACache) Set(key string, data interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache[key] = &CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

// Delete removes an item from cache
func (c *UAACache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.cache, key)
}

// Clear removes all items from cache
func (c *UAACache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache = make(map[string]*CacheEntry)
}

// cleanup periodically removes expired entries
func (c *UAACache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mutex.Lock()
		for key, entry := range c.cache {
			if entry.IsExpired() {
				delete(c.cache, key)
			}
		}
		c.mutex.Unlock()
	}
}

// Global cache instance
var globalCache = NewUAACache(10 * time.Minute)

// BatchOperation represents a batch operation request
type BatchOperation struct {
	Type     string      // "create", "update", "delete"
	Resource string      // "user", "group", "client"
	Data     interface{} // Operation data
	ID       string      // Optional ID for updates/deletes
}

// BatchResult represents the result of a batch operation
type BatchResult struct {
	Operation BatchOperation
	Result    interface{}
	Error     error
}

// BatchProcessor handles batch operations efficiently
type BatchProcessor struct {
	client     *UAAClientWrapper
	maxWorkers int
	batchSize  int
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(client *UAAClientWrapper) *BatchProcessor {
	return &BatchProcessor{
		client:     client,
		maxWorkers: 10,
		batchSize:  50,
	}
}

// ProcessBatch executes multiple operations in parallel
func (bp *BatchProcessor) ProcessBatch(operations []BatchOperation) []BatchResult {
	results := make([]BatchResult, len(operations))
	jobs := make(chan int, len(operations))

	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < bp.maxWorkers && i < len(operations); i++ {
		wg.Add(1)
		go bp.worker(jobs, operations, results, &wg)
	}

	// Send jobs
	for i := range operations {
		jobs <- i
	}
	close(jobs)

	// Wait for completion
	wg.Wait()

	return results
}

// worker processes batch operations
func (bp *BatchProcessor) worker(jobs <-chan int, operations []BatchOperation, results []BatchResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for i := range jobs {
		op := operations[i]
		result, err := bp.executeOperation(op)
		results[i] = BatchResult{
			Operation: op,
			Result:    result,
			Error:     err,
		}
	}
}

// executeOperation executes a single batch operation
func (bp *BatchProcessor) executeOperation(op BatchOperation) (interface{}, error) {
	switch op.Resource {
	case "user":
		return bp.executeUserOperation(op)
	case "group":
		return bp.executeGroupOperation(op)
	case "client":
		return bp.executeClientOperation(op)
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", op.Resource)
	}
}

// executeUserOperation executes user-related batch operations
func (bp *BatchProcessor) executeUserOperation(op BatchOperation) (interface{}, error) {
	switch op.Type {
	case "create":
		if user, ok := op.Data.(uaa.User); ok {
			return bp.client.Client().CreateUser(user)
		}
		return nil, fmt.Errorf("invalid user data for create operation")

	case "update":
		if user, ok := op.Data.(uaa.User); ok {
			return bp.client.Client().UpdateUser(user)
		}
		return nil, fmt.Errorf("invalid user data for update operation")

	case "delete":
		if op.ID != "" {
			return bp.client.Client().DeleteUser(op.ID)
		}
		return nil, fmt.Errorf("user ID required for delete operation")

	default:
		return nil, fmt.Errorf("unsupported user operation: %s", op.Type)
	}
}

// executeGroupOperation executes group-related batch operations
func (bp *BatchProcessor) executeGroupOperation(op BatchOperation) (interface{}, error) {
	switch op.Type {
	case "create":
		if group, ok := op.Data.(uaa.Group); ok {
			return bp.client.Client().CreateGroup(group)
		}
		return nil, fmt.Errorf("invalid group data for create operation")

	case "update":
		if group, ok := op.Data.(uaa.Group); ok {
			return bp.client.Client().UpdateGroup(group)
		}
		return nil, fmt.Errorf("invalid group data for update operation")

	case "delete":
		if op.ID != "" {
			return bp.client.Client().DeleteGroup(op.ID)
		}
		return nil, fmt.Errorf("group ID required for delete operation")

	default:
		return nil, fmt.Errorf("unsupported group operation: %s", op.Type)
	}
}

// executeClientOperation executes client-related batch operations
func (bp *BatchProcessor) executeClientOperation(op BatchOperation) (interface{}, error) {
	switch op.Type {
	case "create":
		if client, ok := op.Data.(uaa.Client); ok {
			return bp.client.Client().CreateClient(client)
		}
		return nil, fmt.Errorf("invalid client data for create operation")

	case "update":
		if client, ok := op.Data.(uaa.Client); ok {
			return bp.client.Client().UpdateClient(client)
		}
		return nil, fmt.Errorf("invalid client data for update operation")

	case "delete":
		if op.ID != "" {
			return bp.client.Client().DeleteClient(op.ID)
		}
		return nil, fmt.Errorf("client ID required for delete operation")

	default:
		return nil, fmt.Errorf("unsupported client operation: %s", op.Type)
	}
}

// CachedUserLookup performs cached user lookups
func CachedUserLookup(client *UAAClientWrapper, username string) (*uaa.User, error) {
	cacheKey := fmt.Sprintf("user:%s", username)

	// Check cache first
	if cached, found := globalCache.Get(cacheKey); found {
		if user, ok := cached.(*uaa.User); ok {
			return user, nil
		}
	}

	// Fetch from API
	user, err := client.Client().GetUserByUsername(username, "", "")
	if err != nil {
		return nil, err
	}

	// Cache the result
	globalCache.Set(cacheKey, user)
	return user, nil
}

// CachedGroupLookup performs cached group lookups
func CachedGroupLookup(client *UAAClientWrapper, groupName string) (*uaa.Group, error) {
	cacheKey := fmt.Sprintf("group:%s", groupName)

	// Check cache first
	if cached, found := globalCache.Get(cacheKey); found {
		if group, ok := cached.(*uaa.Group); ok {
			return group, nil
		}
	}

	// Fetch from API using list groups with filter
	groups, _, err := client.Client().ListGroups(fmt.Sprintf("displayName eq \"%s\"", groupName), "", "", uaa.SortAscending, 1, 1)
	if err != nil {
		return nil, err
	}

	if len(groups) == 0 {
		return nil, fmt.Errorf("group not found: %s", groupName)
	}

	group := &groups[0]

	// Cache the result
	globalCache.Set(cacheKey, group)
	return group, nil
}

// CachedClientLookup performs cached client lookups
func CachedClientLookup(client *UAAClientWrapper, clientID string) (*uaa.Client, error) {
	cacheKey := fmt.Sprintf("client:%s", clientID)

	// Check cache first
	if cached, found := globalCache.Get(cacheKey); found {
		if clientObj, ok := cached.(*uaa.Client); ok {
			return clientObj, nil
		}
	}

	// Fetch from API
	clientObj, err := client.Client().GetClient(clientID)
	if err != nil {
		return nil, err
	}

	// Cache the result
	globalCache.Set(cacheKey, clientObj)
	return clientObj, nil
}

// CachedServerInfo performs cached server info lookup
func CachedServerInfo(client *UAAClientWrapper) (map[string]interface{}, error) {
	cacheKey := "server_info"

	// Check cache first
	if cached, found := globalCache.Get(cacheKey); found {
		if serverInfo, ok := cached.(map[string]interface{}); ok {
			return serverInfo, nil
		}
	}

	// Fetch from API
	ctx := context.Background()
	serverInfo, err := client.GetServerInfo(ctx)
	if err != nil {
		return nil, err
	}

	// Cache the result
	globalCache.Set(cacheKey, serverInfo)
	return serverInfo, nil
}

// InvalidateUserCache removes user-related cache entries
func InvalidateUserCache(username string) {
	globalCache.Delete(fmt.Sprintf("user:%s", username))
}

// InvalidateGroupCache removes group-related cache entries
func InvalidateGroupCache(groupName string) {
	globalCache.Delete(fmt.Sprintf("group:%s", groupName))
}

// InvalidateClientCache removes client-related cache entries
func InvalidateClientCache(clientID string) {
	globalCache.Delete(fmt.Sprintf("client:%s", clientID))
}

// OptimizedPagination handles efficient pagination for large datasets
type OptimizedPagination struct {
	client   *UAAClientWrapper
	pageSize int
	maxPages int
	cache    bool
}

// NewOptimizedPagination creates a new optimized pagination handler
func NewOptimizedPagination(client *UAAClientWrapper) *OptimizedPagination {
	return &OptimizedPagination{
		client:   client,
		pageSize: 100, // Larger page size for efficiency
		maxPages: 50,  // Prevent infinite loops
		cache:    true,
	}
}

// GetAllUsers efficiently retrieves all users with caching
func (op *OptimizedPagination) GetAllUsers(filter, sortBy, attributes string, sortOrder uaa.SortOrder) ([]uaa.User, error) {
	cacheKey := fmt.Sprintf("all_users:%s:%s:%s:%s", filter, sortBy, attributes, sortOrder)

	// Check cache if enabled
	if op.cache {
		if cached, found := globalCache.Get(cacheKey); found {
			if users, ok := cached.([]uaa.User); ok {
				return users, nil
			}
		}
	}

	// Fetch all users with optimized pagination
	var allUsers []uaa.User
	startIndex := 1

	for page := 0; page < op.maxPages; page++ {
		users, pagination, err := op.client.Client().ListUsers(filter, sortBy, attributes, sortOrder, startIndex, op.pageSize)
		if err != nil {
			return nil, err
		}

		allUsers = append(allUsers, users...)

		// Check if we have more pages
		if pagination.TotalResults <= startIndex+len(users)-1 {
			break
		}

		startIndex += len(users)
	}

	// Cache the result if enabled
	if op.cache {
		globalCache.Set(cacheKey, allUsers)
	}

	return allUsers, nil
}

// GetAllGroups efficiently retrieves all groups with caching
func (op *OptimizedPagination) GetAllGroups(filter, sortBy, attributes string, sortOrder uaa.SortOrder) ([]uaa.Group, error) {
	cacheKey := fmt.Sprintf("all_groups:%s:%s:%s:%s", filter, sortBy, attributes, sortOrder)

	// Check cache if enabled
	if op.cache {
		if cached, found := globalCache.Get(cacheKey); found {
			if groups, ok := cached.([]uaa.Group); ok {
				return groups, nil
			}
		}
	}

	// Fetch all groups with optimized pagination
	var allGroups []uaa.Group
	startIndex := 1

	for page := 0; page < op.maxPages; page++ {
		groups, pagination, err := op.client.Client().ListGroups(filter, sortBy, attributes, sortOrder, startIndex, op.pageSize)
		if err != nil {
			return nil, err
		}

		allGroups = append(allGroups, groups...)

		// Check if we have more pages
		if pagination.TotalResults <= startIndex+len(groups)-1 {
			break
		}

		startIndex += len(groups)
	}

	// Cache the result if enabled
	if op.cache {
		globalCache.Set(cacheKey, allGroups)
	}

	return allGroups, nil
}

// GetAllClients efficiently retrieves all clients with caching
func (op *OptimizedPagination) GetAllClients(filter, sortBy, attributes string, sortOrder uaa.SortOrder) ([]uaa.Client, error) {
	cacheKey := fmt.Sprintf("all_clients:%s:%s:%s:%s", filter, sortBy, attributes, sortOrder)

	// Check cache if enabled
	if op.cache {
		if cached, found := globalCache.Get(cacheKey); found {
			if clients, ok := cached.([]uaa.Client); ok {
				return clients, nil
			}
		}
	}

	// Fetch all clients with optimized pagination
	var allClients []uaa.Client
	startIndex := 1

	for page := 0; page < op.maxPages; page++ {
		clients, pagination, err := op.client.Client().ListClients(filter, sortBy, sortOrder, startIndex, op.pageSize)
		if err != nil {
			return nil, err
		}

		allClients = append(allClients, clients...)

		// Check if we have more pages
		if pagination.TotalResults <= startIndex+len(clients)-1 {
			break
		}

		startIndex += len(clients)
	}

	// Cache the result if enabled
	if op.cache {
		globalCache.Set(cacheKey, allClients)
	}

	return allClients, nil
}

// BulkUserImport handles efficient bulk user import from JSON
func BulkUserImport(client *UAAClientWrapper, usersJSON []byte, parallel bool) ([]BatchResult, error) {
	var users []uaa.User
	if err := json.Unmarshal(usersJSON, &users); err != nil {
		return nil, fmt.Errorf("failed to parse users JSON: %w", err)
	}

	// Create batch operations
	operations := make([]BatchOperation, len(users))
	for i, user := range users {
		operations[i] = BatchOperation{
			Type:     "create",
			Resource: "user",
			Data:     user,
		}
	}

	if parallel {
		// Use batch processor for parallel execution
		processor := NewBatchProcessor(client)
		return processor.ProcessBatch(operations), nil
	} else {
		// Sequential processing
		results := make([]BatchResult, len(operations))
		for i, op := range operations {
			user := op.Data.(uaa.User)
			result, err := client.Client().CreateUser(user)
			results[i] = BatchResult{
				Operation: op,
				Result:    result,
				Error:     err,
			}
		}
		return results, nil
	}
}

// PerformanceMetrics tracks operation performance
type PerformanceMetrics struct {
	mutex           sync.RWMutex
	operations      map[string][]time.Duration
	cacheHits       int64
	cacheMisses     int64
	totalOperations int64
}

// Global performance metrics
var performanceMetrics = &PerformanceMetrics{
	operations: make(map[string][]time.Duration),
}

// TrackOperation records the duration of an operation
func (pm *PerformanceMetrics) TrackOperation(operation string, duration time.Duration) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.operations[operation] = append(pm.operations[operation], duration)
	pm.totalOperations++
}

// TrackCacheHit records a cache hit
func (pm *PerformanceMetrics) TrackCacheHit() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	pm.cacheHits++
}

// TrackCacheMiss records a cache miss
func (pm *PerformanceMetrics) TrackCacheMiss() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	pm.cacheMisses++
}

// GetMetrics returns current performance metrics
func (pm *PerformanceMetrics) GetMetrics() map[string]interface{} {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	var cacheHitRate float64
	totalCacheOps := pm.cacheHits + pm.cacheMisses
	if totalCacheOps > 0 {
		cacheHitRate = float64(pm.cacheHits) / float64(totalCacheOps) * 100
	}

	metrics := map[string]interface{}{
		"total_operations": pm.totalOperations,
		"cache_hits":       pm.cacheHits,
		"cache_misses":     pm.cacheMisses,
		"cache_hit_rate":   cacheHitRate,
		"operations":       make(map[string]interface{}),
	}

	// Calculate operation statistics
	for op, durations := range pm.operations {
		if len(durations) > 0 {
			var total time.Duration
			min := durations[0]
			max := durations[0]

			for _, d := range durations {
				total += d
				if d < min {
					min = d
				}
				if d > max {
					max = d
				}
			}

			avg := total / time.Duration(len(durations))

			metrics["operations"].(map[string]interface{})[op] = map[string]interface{}{
				"count":   len(durations),
				"average": avg.String(),
				"min":     min.String(),
				"max":     max.String(),
			}
		}
	}

	return metrics
}

// WithPerformanceTracking wraps a function with performance tracking
func WithPerformanceTracking(operation string, fn func() error) error {
	start := time.Now()
	err := fn()
	duration := time.Since(start)

	performanceMetrics.TrackOperation(operation, duration)
	return err
}
