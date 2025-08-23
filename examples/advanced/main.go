package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/fivetwenty-io/capi/v3/pkg/capi"
	"github.com/fivetwenty-io/capi/v3/pkg/cfclient"
)

func main() {
	// Example 1: Advanced client configuration
	fmt.Println("=== Advanced Client Configuration ===")
	advancedConfigExample()

	// Example 2: Concurrent operations
	fmt.Println("\n=== Concurrent Operations ===")
	concurrentOperationsExample()

	// Example 3: Batch operations
	fmt.Println("\n=== Batch Operations ===")
	batchOperationsExample()

	// Example 4: Advanced error handling and retries
	fmt.Println("\n=== Advanced Error Handling ===")
	errorHandlingExample()

	// Example 5: Streaming large datasets
	fmt.Println("\n=== Streaming Large Datasets ===")
	streamingExample()

	// Example 6: Custom interceptors
	fmt.Println("\n=== Custom Interceptors ===")
	interceptorsExample()

	// Example 7: Performance monitoring
	fmt.Println("\n=== Performance Monitoring ===")
	performanceMonitoringExample()
}

func advancedConfigExample() {
	// Create a highly customized client configuration
	config := &capi.Config{
		APIEndpoint:   "https://api.your-cf-domain.com",
		Username:      "your-username",
		Password:      "your-password",
		SkipTLSVerify: false,
		HTTPTimeout:   45 * time.Second,
		UserAgent:     "advanced-example/1.0.0",
		RetryMax:      5,
		RetryWaitMin:  1 * time.Second,
		RetryWaitMax:  10 * time.Second,
		Debug:         true,
	}

	client, err := cfclient.New(config)
	if err != nil {
		log.Printf("Failed to create advanced client: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test the configuration
	info, err := client.GetInfo(ctx)
	if err != nil {
		log.Printf("Failed to get API info: %v", err)
		return
	}

	fmt.Printf("Connected to CF API version %d with advanced configuration\n", info.Version)
}

func concurrentOperationsExample() {
	client, err := cfclient.NewWithPassword(
		"https://api.your-cf-domain.com",
		"your-username",
		"your-password",
	)
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}

	ctx := context.Background()

	// Get all organizations concurrently with their spaces
	orgs, err := client.Organizations().List(ctx, nil)
	if err != nil {
		log.Printf("Failed to list organizations: %v", err)
		return
	}

	fmt.Printf("Processing %d organizations concurrently...\n", len(orgs.Resources))

	// Use worker pool pattern for concurrent operations
	const numWorkers = 5
	orgChan := make(chan *capi.Organization, len(orgs.Resources))
	resultsChan := make(chan OrgSpaceResult, len(orgs.Resources))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for org := range orgChan {
				result := processOrganizationSpaces(client, ctx, org, workerID)
				resultsChan <- result
			}
		}(i)
	}

	// Send work to workers
	go func() {
		for _, org := range orgs.Resources {
			orgChan <- &org
		}
		close(orgChan)
	}()

	// Wait for workers to complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	totalSpaces := 0
	for result := range resultsChan {
		if result.Error != nil {
			fmt.Printf("Error processing org %s: %v\n", result.OrgName, result.Error)
		} else {
			fmt.Printf("Org %s has %d spaces (processed by worker %d)\n",
				result.OrgName, result.SpaceCount, result.WorkerID)
			totalSpaces += result.SpaceCount
		}
	}

	fmt.Printf("Total spaces across all organizations: %d\n", totalSpaces)
}

type OrgSpaceResult struct {
	OrgName    string
	SpaceCount int
	WorkerID   int
	Error      error
}

func processOrganizationSpaces(client capi.Client, ctx context.Context, org *capi.Organization, workerID int) OrgSpaceResult {
	params := capi.NewQueryParams().WithFilter("organization_guids", org.GUID)
	spaces, err := client.Spaces().List(ctx, params)
	if err != nil {
		return OrgSpaceResult{
			OrgName:  org.Name,
			WorkerID: workerID,
			Error:    err,
		}
	}

	return OrgSpaceResult{
		OrgName:    org.Name,
		SpaceCount: len(spaces.Resources),
		WorkerID:   workerID,
	}
}

func batchOperationsExample() {
	client, err := cfclient.NewWithPassword(
		"https://api.your-cf-domain.com",
		"your-username",
		"your-password",
	)
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}

	ctx := context.Background()

	// Simulate batch operations (since the client doesn't have native batch support,
	// we'll implement a pattern for handling multiple operations efficiently)

	type Operation struct {
		Type string
		Name string
		Func func() error
	}

	operations := []Operation{
		{
			Type: "org",
			Name: "batch-org-1",
			Func: func() error {
				_, err := client.Organizations().Create(ctx, &capi.OrganizationCreateRequest{
					Name: "batch-org-1",
				})
				return err
			},
		},
		{
			Type: "org",
			Name: "batch-org-2",
			Func: func() error {
				_, err := client.Organizations().Create(ctx, &capi.OrganizationCreateRequest{
					Name: "batch-org-2",
				})
				return err
			},
		},
		{
			Type: "org",
			Name: "batch-org-3",
			Func: func() error {
				_, err := client.Organizations().Create(ctx, &capi.OrganizationCreateRequest{
					Name: "batch-org-3",
				})
				return err
			},
		},
	}

	// Execute batch operations with rate limiting
	semaphore := make(chan struct{}, 3) // Limit to 3 concurrent operations
	var wg sync.WaitGroup
	results := make(chan OperationResult, len(operations))

	for _, op := range operations {
		wg.Add(1)
		go func(operation Operation) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			start := time.Now()
			err := operation.Func()
			duration := time.Since(start)

			results <- OperationResult{
				Type:     operation.Type,
				Name:     operation.Name,
				Duration: duration,
				Error:    err,
			}
		}(op)
	}

	// Wait for completion and collect results
	go func() {
		wg.Wait()
		close(results)
	}()

	successful := 0
	failed := 0
	for result := range results {
		if result.Error != nil {
			fmt.Printf("❌ %s %s failed in %v: %v\n",
				result.Type, result.Name, result.Duration, result.Error)
			failed++
		} else {
			fmt.Printf("✅ %s %s succeeded in %v\n",
				result.Type, result.Name, result.Duration)
			successful++
		}
	}

	fmt.Printf("Batch operations completed: %d successful, %d failed\n", successful, failed)
}

type OperationResult struct {
	Type     string
	Name     string
	Duration time.Duration
	Error    error
}

func errorHandlingExample() {
	client, err := cfclient.NewWithPassword(
		"https://api.your-cf-domain.com",
		"wrong-username", // Intentionally wrong credentials
		"wrong-password",
	)
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}

	ctx := context.Background()

	// Demonstrate comprehensive error handling with retries
	err = withExponentialBackoff(func() error {
		_, err := client.Organizations().List(ctx, nil)
		return err
	}, 3, time.Second)

	if err != nil {
		fmt.Printf("Operation failed after retries: %v\n", err)

		// Analyze the error
		analyzeError(err)
	}

	// Demonstrate timeout handling
	shortCtx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	_, err = client.Organizations().List(shortCtx, nil)
	if err != nil {
		fmt.Printf("Timeout error (expected): %v\n", err)
	}
}

func withExponentialBackoff(operation func() error, maxRetries int, baseDelay time.Duration) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryableError(err) {
			fmt.Printf("Non-retryable error encountered: %v\n", err)
			return err
		}

		if attempt < maxRetries {
			delay := time.Duration(math.Pow(2, float64(attempt))) * baseDelay
			fmt.Printf("Attempt %d failed, retrying in %v: %v\n", attempt+1, delay, err)
			time.Sleep(delay)
		}
	}

	return fmt.Errorf("operation failed after %d attempts: %w", maxRetries+1, lastErr)
}

func isRetryableError(err error) bool {
	if capiErr, ok := err.(*capi.ErrorResponse); ok {
		// Check if any error is retryable (5xx status codes)
		for _, apiErr := range capiErr.Errors {
			if apiErr.Code >= 50000 && apiErr.Code < 60000 {
				return true
			}
		}
		return false
	}

	// Retry network errors
	return true
}

func analyzeError(err error) {
	switch e := err.(type) {
	case *capi.ErrorResponse:
		fmt.Printf("CF API Error Analysis:\n")
		fmt.Printf("  Error count: %d\n", len(e.Errors))

		if len(e.Errors) > 0 {
			fmt.Printf("  Individual Errors:\n")
			for _, detail := range e.Errors {
				fmt.Printf("    - Code: %d, Title: %s, Detail: %s\n", detail.Code, detail.Title, detail.Detail)
			}
		}

		// Provide specific guidance based on error type
		if len(e.Errors) > 0 {
			firstError := e.Errors[0]
			switch {
			case firstError.Code >= 10000 && firstError.Code < 11000:
				fmt.Println("  💡 Check your credentials or refresh your token")
			case firstError.Code >= 10003 && firstError.Code < 10004:
				fmt.Println("  💡 You may lack the required permissions for this operation")
			case firstError.Code == 10010:
				fmt.Println("  💡 The requested resource was not found")
			case firstError.Code >= 10008 && firstError.Code < 10009:
				fmt.Println("  💡 The request was invalid - check required fields and formats")
			case firstError.Code >= 50000:
				fmt.Println("  💡 Server error - try again later or contact support")
			}
		}

	default:
		fmt.Printf("Other Error: %v (type: %T)\n", err, err)
	}
}

func streamingExample() {
	client, err := cfclient.NewWithPassword(
		"https://api.your-cf-domain.com",
		"your-username",
		"your-password",
	)
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}

	ctx := context.Background()

	// Stream applications in batches to reduce memory usage
	fmt.Println("Streaming applications in batches...")

	processedCount := 0
	batchSize := 10
	params := capi.NewQueryParams().WithPerPage(batchSize)

	appList, err := client.Apps().List(ctx, params)
	if err == nil {
		// Process each batch
		fmt.Printf("Processing batch of %d applications...\n", len(appList.Resources))

		for range appList.Resources {
			// Simulate processing each application
			processedCount++

			if processedCount%50 == 0 {
				fmt.Printf("Processed %d applications so far...\n", processedCount)
			}

			// Add artificial delay to simulate processing time
			time.Sleep(10 * time.Millisecond)
		}
	}

	if err != nil {
		log.Printf("Streaming failed: %v", err)
		return
	}

	fmt.Printf("Streaming completed. Total applications processed: %d\n", processedCount)
}

func interceptorsExample() {
	client, err := cfclient.NewWithPassword(
		"https://api.your-cf-domain.com",
		"your-username",
		"your-password",
	)
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}

	// Note: This is a conceptual example. The actual client implementation
	// would need to support interceptors. Here we show what it might look like.

	fmt.Println("Demonstrating request/response interceptor concepts...")

	// Conceptual request interceptor
	requestInterceptor := func(req *http.Request) error {
		fmt.Printf("🚀 Outgoing request: %s %s\n", req.Method, req.URL.Path)
		req.Header.Set("X-Custom-Client", "advanced-example")
		return nil
	}

	// Conceptual response interceptor
	responseInterceptor := func(resp *http.Response) error {
		fmt.Printf("📥 Incoming response: %d %s (took %v)\n",
			resp.StatusCode, resp.Status,
			resp.Header.Get("X-Response-Time"))
		return nil
	}

	// In a real implementation, you would add these interceptors:
	// client.AddRequestInterceptor(requestInterceptor)
	// client.AddResponseInterceptor(responseInterceptor)

	ctx := context.Background()

	// Make a request to demonstrate interceptor concepts
	_, err = client.GetInfo(ctx)
	if err != nil {
		log.Printf("Request failed: %v", err)
	}

	fmt.Printf("Request/response interceptors would capture: %v, %v\n",
		requestInterceptor != nil, responseInterceptor != nil)
}

func performanceMonitoringExample() {
	client, err := cfclient.NewWithPassword(
		"https://api.your-cf-domain.com",
		"your-username",
		"your-password",
	)
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}

	ctx := context.Background()

	// Performance monitoring structure
	metrics := &PerformanceMetrics{
		RequestCount: 0,
		TotalTime:    0,
		Errors:       0,
	}

	// Monitor multiple operations
	operations := []struct {
		name string
		fn   func() error
	}{
		{
			name: "list organizations",
			fn: func() error {
				_, err := client.Organizations().List(ctx, nil)
				return err
			},
		},
		{
			name: "get API info",
			fn: func() error {
				_, err := client.GetInfo(ctx)
				return err
			},
		},
		{
			name: "list applications",
			fn: func() error {
				params := capi.NewQueryParams().WithPerPage(10)
				_, err := client.Apps().List(ctx, params)
				return err
			},
		},
	}

	fmt.Println("Running performance monitoring tests...")

	for _, op := range operations {
		start := time.Now()
		err := op.fn()
		duration := time.Since(start)

		metrics.RequestCount++
		metrics.TotalTime += duration

		if err != nil {
			metrics.Errors++
			fmt.Printf("❌ %s failed in %v: %v\n", op.name, duration, err)
		} else {
			fmt.Printf("✅ %s completed in %v\n", op.name, duration)
		}
	}

	// Report performance metrics
	fmt.Printf("\nPerformance Summary:\n")
	fmt.Printf("  Total requests: %d\n", metrics.RequestCount)
	fmt.Printf("  Total time: %v\n", metrics.TotalTime)
	fmt.Printf("  Average time per request: %v\n", metrics.TotalTime/time.Duration(metrics.RequestCount))
	fmt.Printf("  Error rate: %.2f%%\n", float64(metrics.Errors)/float64(metrics.RequestCount)*100)
	fmt.Printf("  Requests per second: %.2f\n", float64(metrics.RequestCount)/metrics.TotalTime.Seconds())
}

type PerformanceMetrics struct {
	RequestCount int
	TotalTime    time.Duration
	Errors       int
}
