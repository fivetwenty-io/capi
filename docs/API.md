# Cloud Foundry API v3 Client Library Documentation

This document provides comprehensive documentation for the Cloud Foundry API v3 client library.

## Table of Contents

- [Getting Started](#getting-started)
- [Authentication](#authentication)
- [Client Configuration](#client-configuration)
- [Resource Operations](#resource-operations)
- [Pagination](#pagination)
- [Error Handling](#error-handling)
- [Caching](#caching)
- [Advanced Usage](#advanced-usage)

## Getting Started

### Installation

```bash
go get github.com/fivetwenty-io/capi-client
```

### Basic Usage

```go
package main

import (
    "context"
    "log"

    "github.com/fivetwenty-io/capi-client/pkg/cfclient"
)

func main() {
    client, err := cfclient.NewWithPassword(
        "https://api.your-cf-domain.com",
        "username",
        "password",
    )
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    orgs, err := client.Organizations().List(ctx, nil)
    if err != nil {
        log.Fatal(err)
    }

    // Process organizations...
}
```

## Authentication

The client supports multiple authentication methods:

### Username/Password Authentication

```go
client, err := cfclient.NewWithPassword(endpoint, username, password)
```

This method uses the OAuth2 Resource Owner Password Credentials grant type. The client will automatically discover the UAA endpoint from the CF API and obtain access tokens.

### Client Credentials Authentication

```go
client, err := cfclient.NewWithClientCredentials(endpoint, clientID, clientSecret)
```

This method uses OAuth2 Client Credentials grant type, suitable for service-to-service authentication.

### Access Token Authentication

```go
client, err := cfclient.NewWithToken(endpoint, accessToken)
```

Use this method when you already have a valid access token from another authentication flow.

### Refresh Token Authentication

```go
config := &capi.Config{
    APIEndpoint:  endpoint,
    RefreshToken: refreshToken,
    ClientID:     clientID,
    ClientSecret: clientSecret,
}
client, err := cfclient.New(config)
```

The client will automatically refresh the access token when it expires.

## Client Configuration

### Basic Configuration

```go
config := &capi.Config{
    APIEndpoint:   "https://api.cf.com",
    Username:      "user",
    Password:      "pass",
    SkipTLSVerify: false,
    Timeout:       30 * time.Second,
    UserAgent:     "my-app/1.0.0",
}

client, err := cfclient.New(config)
```

### Advanced Configuration

```go
config := &capi.Config{
    APIEndpoint:     "https://api.cf.com",
    TokenURL:        "https://uaa.cf.com/oauth/token", // Auto-discovered if not provided
    ClientID:        "cf",                             // Default CF CLI client
    ClientSecret:    "",
    Username:        "user",
    Password:        "pass",
    AccessToken:     "",
    RefreshToken:    "",
    SkipTLSVerify:   false,
    Timeout:         30 * time.Second,
    UserAgent:       "my-app/1.0.0",
    MaxRetries:      3,
    RetryDelay:      time.Second,
    RateLimitRPS:    100,
    RateLimitBurst:  10,
    
    // Caching configuration
    Cache: &capi.CacheConfig{
        Type: "memory", // "memory", "redis", "nats"
        TTL:  5 * time.Minute,
        Config: map[string]interface{}{
            "size": 1000, // For memory cache
        },
    },
    
    // HTTP client customization
    HTTPClient: &http.Client{
        Timeout: 30 * time.Second,
        Transport: &http.Transport{
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 10,
            IdleConnTimeout:     90 * time.Second,
        },
    },
}
```

## Resource Operations

### Organizations

#### List Organizations

```go
// List all organizations
orgs, err := client.Organizations().List(ctx, nil)

// List with query parameters
params := capi.NewQueryParams()
params.WithFilter("names", "my-org,other-org")
params.WithPerPage(50)
params.WithPage(1)
orgs, err := client.Organizations().List(ctx, params)

// List all organizations with automatic pagination
var allOrgs []*capi.Organization
err := client.Organizations().ListAll(ctx, nil, func(orgList *capi.OrganizationList) error {
    allOrgs = append(allOrgs, orgList.Resources...)
    return nil
})
```

#### Get Organization

```go
org, err := client.Organizations().Get(ctx, "org-guid")

// Include relationships
params := capi.NewQueryParams().WithInclude("spaces", "spaces.organization")
org, err := client.Organizations().Get(ctx, "org-guid", params)
```

#### Create Organization

```go
createReq := &capi.OrganizationCreate{
    Name: "new-org",
    Suspended: capi.Bool(false),
    Metadata: &capi.Metadata{
        Labels: map[string]string{
            "team":        "platform",
            "environment": "production",
        },
        Annotations: map[string]string{
            "created-by":    "automation",
            "creation-date": time.Now().Format(time.RFC3339),
        },
    },
}

org, err := client.Organizations().Create(ctx, createReq)
```

#### Update Organization

```go
updateReq := &capi.OrganizationUpdate{
    Name:      capi.String("updated-name"),
    Suspended: capi.Bool(false),
    Metadata: &capi.MetadataUpdate{
        Labels: map[string]*string{
            "version":     capi.String("2.0"),
            "environment": nil, // Remove this label
        },
        Annotations: map[string]*string{
            "updated-by": capi.String("admin"),
        },
    },
}

org, err := client.Organizations().Update(ctx, "org-guid", updateReq)
```

#### Delete Organization

```go
job, err := client.Organizations().Delete(ctx, "org-guid")

// Poll job completion
for {
    job, err = client.Jobs().Get(ctx, job.GUID)
    if err != nil {
        break
    }
    if job.State == "COMPLETE" || job.State == "FAILED" {
        break
    }
    time.Sleep(time.Second)
}
```

### Applications

#### List Applications

```go
// List all applications
apps, err := client.Applications().List(ctx, nil)

// Filter applications
params := capi.NewQueryParams()
params.WithFilter("names", "my-app")
params.WithFilter("space_guids", "space-guid")
params.WithFilter("states", "STARTED,STOPPED")
apps, err := client.Applications().List(ctx, params)

// Include relationships
params.WithInclude("space", "space.organization")
apps, err := client.Applications().List(ctx, params)
```

#### Application Lifecycle

```go
// Start application
app, err := client.Applications().Start(ctx, "app-guid")

// Stop application  
app, err := client.Applications().Stop(ctx, "app-guid")

// Restart application
app, err := client.Applications().Restart(ctx, "app-guid")
```

#### Scale Applications

```go
scaleReq := &capi.ProcessScale{
    Instances: capi.Int(5),
    Memory:    capi.String("1G"),
    Disk:      capi.String("2G"),
}

process, err := client.Applications().ScaleProcess(ctx, "app-guid", "web", scaleReq)
```

#### Environment Variables

```go
// Get environment variables
envVars, err := client.Applications().GetEnvironmentVariables(ctx, "app-guid")

// Set environment variables
setEnvReq := &capi.ApplicationEnvironmentVariables{
    Var: map[string]string{
        "DATABASE_URL": "postgres://...",
        "DEBUG":        "true",
    },
}
envVars, err := client.Applications().UpdateEnvironmentVariables(ctx, "app-guid", setEnvReq)
```

#### Application Stats and Processes

```go
// Get application stats
stats, err := client.Applications().GetStats(ctx, "app-guid")

// Get processes
processes, err := client.Applications().ListProcesses(ctx, "app-guid", nil)

// Get specific process
process, err := client.Applications().GetProcess(ctx, "app-guid", "web")
```

### Spaces

#### List and Manage Spaces

```go
// List spaces in organization
params := capi.NewQueryParams().WithFilter("organization_guids", "org-guid")
spaces, err := client.Spaces().List(ctx, params)

// Create space
createReq := &capi.SpaceCreate{
    Name: "development",
    Relationships: &capi.SpaceRelationships{
        Organization: &capi.Relationship{
            Data: &capi.RelationshipData{GUID: "org-guid"},
        },
    },
}
space, err := client.Spaces().Create(ctx, createReq)

// Get space with relationships
params = capi.NewQueryParams().WithInclude("organization")
space, err := client.Spaces().Get(ctx, "space-guid", params)
```

### Services

#### Service Offerings and Plans

```go
// List service offerings
offerings, err := client.ServiceOfferings().List(ctx, nil)

// Get service offering with plans
params := capi.NewQueryParams().WithInclude("service_plans")
offering, err := client.ServiceOfferings().Get(ctx, "offering-guid", params)

// List service plans
plans, err := client.ServicePlans().List(ctx, nil)
```

#### Service Instances

```go
// List service instances
instances, err := client.ServiceInstances().List(ctx, nil)

// Create managed service instance
createReq := &capi.ServiceInstanceCreate{
    Type: "managed",
    Name: "my-database",
    Relationships: &capi.ServiceInstanceRelationships{
        Space: &capi.Relationship{
            Data: &capi.RelationshipData{GUID: "space-guid"},
        },
        ServicePlan: &capi.Relationship{
            Data: &capi.RelationshipData{GUID: "plan-guid"},
        },
    },
    Parameters: map[string]interface{}{
        "storage": "10GB",
        "version": "13",
    },
}
instance, err := client.ServiceInstances().Create(ctx, createReq)

// Create user-provided service instance
createUserProvided := &capi.ServiceInstanceCreate{
    Type: "user-provided",
    Name: "my-external-service",
    Relationships: &capi.ServiceInstanceRelationships{
        Space: &capi.Relationship{
            Data: &capi.RelationshipData{GUID: "space-guid"},
        },
    },
    Credentials: map[string]interface{}{
        "uri":      "https://external-api.com",
        "api_key":  "secret-key",
    },
}
instance, err := client.ServiceInstances().Create(ctx, createUserProvided)
```

#### Service Bindings

```go
// List service credential bindings
bindings, err := client.ServiceCredentialBindings().List(ctx, nil)

// Create service binding
createReq := &capi.ServiceCredentialBindingCreate{
    Type: "app",
    Name: "my-binding",
    Relationships: &capi.ServiceCredentialBindingRelationships{
        ServiceInstance: &capi.Relationship{
            Data: &capi.RelationshipData{GUID: "instance-guid"},
        },
        App: &capi.Relationship{
            Data: &capi.RelationshipData{GUID: "app-guid"},
        },
    },
    Parameters: map[string]interface{}{
        "permission": "read-write",
    },
}
binding, err := client.ServiceCredentialBindings().Create(ctx, createReq)
```

## Pagination

The client provides several methods for handling paginated responses:

### Manual Pagination

```go
params := capi.NewQueryParams().WithPerPage(50).WithPage(1)

for {
    apps, err := client.Applications().List(ctx, params)
    if err != nil {
        return err
    }
    
    // Process applications
    for _, app := range apps.Resources {
        // Handle each application
    }
    
    // Check if there are more pages
    if apps.Pagination.Next == nil {
        break
    }
    
    params.WithPage(params.Page + 1)
}
```

### Automatic Pagination

```go
var allApps []*capi.Application

err := client.Applications().ListAll(ctx, nil, func(appList *capi.ApplicationList) error {
    allApps = append(allApps, appList.Resources...)
    
    // Optional: Add progress logging
    fmt.Printf("Loaded %d applications...\n", len(allApps))
    
    // Optional: Add rate limiting
    time.Sleep(100 * time.Millisecond)
    
    return nil
})
```

### Pagination Information

```go
apps, err := client.Applications().List(ctx, nil)
if err != nil {
    return err
}

fmt.Printf("Total results: %d\n", apps.Pagination.TotalResults)
fmt.Printf("Total pages: %d\n", apps.Pagination.TotalPages)
fmt.Printf("Current page: %d\n", apps.Pagination.First.Href)

if apps.Pagination.Next != nil {
    fmt.Printf("Next page: %s\n", apps.Pagination.Next.Href)
}
```

## Error Handling

The client provides rich error information through the `capi.Error` type:

### Basic Error Handling

```go
app, err := client.Applications().Get(ctx, "app-guid")
if err != nil {
    if capiErr, ok := err.(*capi.Error); ok {
        fmt.Printf("CF API Error: %s\n", capiErr.Title)
        fmt.Printf("Status: %d\n", capiErr.Status)
        fmt.Printf("Detail: %s\n", capiErr.Detail)
        
        // Handle specific errors
        switch capiErr.Status {
        case 404:
            fmt.Println("Application not found")
        case 401:
            fmt.Println("Authentication failed")
        case 403:
            fmt.Println("Authorization failed")
        }
        
        // Check individual error details
        for _, detail := range capiErr.Errors {
            fmt.Printf("Error: %s - %s\n", detail.Code, detail.Detail)
        }
    } else {
        // Network or other error
        fmt.Printf("Request failed: %v\n", err)
    }
}
```

### Error Types

The client can return several types of errors:

- `*capi.Error`: CF API errors (4xx, 5xx responses)
- `*url.Error`: Network errors
- `*json.SyntaxError`: JSON parsing errors
- `context.DeadlineExceeded`: Timeout errors
- `context.Canceled`: Canceled requests

### Retry Logic

```go
func withRetry(operation func() error, maxRetries int) error {
    var lastErr error
    
    for i := 0; i <= maxRetries; i++ {
        err := operation()
        if err == nil {
            return nil
        }
        
        lastErr = err
        
        // Check if error is retryable
        if capiErr, ok := err.(*capi.Error); ok {
            // Don't retry client errors (4xx)
            if capiErr.Status >= 400 && capiErr.Status < 500 {
                return err
            }
        }
        
        // Exponential backoff
        if i < maxRetries {
            delay := time.Duration(math.Pow(2, float64(i))) * time.Second
            time.Sleep(delay)
        }
    }
    
    return lastErr
}
```

## Caching

The client supports pluggable caching backends to improve performance:

### Memory Cache

```go
config := &capi.Config{
    // ... other config
    Cache: &capi.CacheConfig{
        Type: "memory",
        TTL:  5 * time.Minute,
        Config: map[string]interface{}{
            "size": 1000, // Maximum number of cached items
        },
    },
}
```

### Redis Cache

```go
config := &capi.Config{
    // ... other config
    Cache: &capi.CacheConfig{
        Type: "redis",
        TTL:  10 * time.Minute,
        Config: map[string]interface{}{
            "addr":     "localhost:6379",
            "password": "",
            "db":       0,
        },
    },
}
```

### NATS Cache

```go
config := &capi.Config{
    // ... other config
    Cache: &capi.CacheConfig{
        Type: "nats",
        TTL:  5 * time.Minute,
        Config: map[string]interface{}{
            "url":     "nats://localhost:4222",
            "subject": "cf.cache",
        },
    },
}
```

### Custom Cache Backend

```go
type MyCache struct {
    // implementation
}

func (c *MyCache) Get(key string) ([]byte, bool) {
    // implementation
}

func (c *MyCache) Set(key string, value []byte, ttl time.Duration) {
    // implementation
}

func (c *MyCache) Delete(key string) {
    // implementation
}

// Register custom cache type
capi.RegisterCacheFactory("mycache", func(config map[string]interface{}) (capi.Cache, error) {
    return &MyCache{}, nil
})
```

## Advanced Usage

### Context and Cancellation

```go
// Create context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Create context that can be canceled
ctx, cancel := context.WithCancel(context.Background())

// Cancel operation in another goroutine
go func() {
    time.Sleep(10 * time.Second)
    cancel()
}()

apps, err := client.Applications().List(ctx, nil)
if err == context.Canceled {
    fmt.Println("Operation was canceled")
}
```

### Rate Limiting

```go
config := &capi.Config{
    // ... other config
    RateLimitRPS:   50,  // 50 requests per second
    RateLimitBurst: 10,  // Allow burst of 10 requests
}
```

### Custom HTTP Client

```go
httpClient := &http.Client{
    Timeout: 60 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: false,
        },
    },
}

config := &capi.Config{
    // ... other config
    HTTPClient: httpClient,
}
```

### Interceptors

```go
// Add request interceptor
client.AddRequestInterceptor(func(req *http.Request) error {
    // Add custom headers
    req.Header.Set("X-Custom-Header", "value")
    
    // Log request
    fmt.Printf("Request: %s %s\n", req.Method, req.URL)
    
    return nil
})

// Add response interceptor
client.AddResponseInterceptor(func(resp *http.Response) error {
    // Log response
    fmt.Printf("Response: %d %s\n", resp.StatusCode, resp.Status)
    
    return nil
})
```

### Batch Operations

```go
batch := client.NewBatch()

// Add operations to batch
batch.Organizations().Create(&capi.OrganizationCreate{Name: "org1"})
batch.Organizations().Create(&capi.OrganizationCreate{Name: "org2"})
batch.Organizations().Create(&capi.OrganizationCreate{Name: "org3"})

// Execute batch
results, err := batch.Execute(ctx)
if err != nil {
    return err
}

// Process results
for i, result := range results {
    if result.Error != nil {
        fmt.Printf("Operation %d failed: %v\n", i, result.Error)
    } else {
        org := result.Resource.(*capi.Organization)
        fmt.Printf("Created organization: %s\n", org.Name)
    }
}
```

### Streaming Responses

```go
// Stream large result sets
err := client.Applications().Stream(ctx, nil, func(app *capi.Application) error {
    fmt.Printf("Processing application: %s\n", app.Name)
    
    // Process application immediately to reduce memory usage
    return processApplication(app)
})
```

This comprehensive documentation covers the major features and patterns for using the Cloud Foundry API v3 client library. For more specific examples, see the examples directory in the repository.