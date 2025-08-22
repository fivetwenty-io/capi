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
- [Quota Management](#quota-management)
- [Usage Monitoring](#usage-monitoring)
- [Application Lifecycle](#application-lifecycle)
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

## Quota Management

The client provides comprehensive quota management for both organizations and spaces.

### Organization Quotas

Organization quotas define resource limits at the organization level.

#### List Organization Quotas

```go
// List all organization quotas
quotas, err := client.OrganizationQuotas().List(ctx, nil)

// List with filtering
params := capi.NewQueryParams()
params.WithFilter("names", "production,staging")
quotas, err := client.OrganizationQuotas().List(ctx, params)
```

#### Get Organization Quota

```go
quota, err := client.OrganizationQuotas().Get(ctx, "quota-guid")

// You can also get by name
params := capi.NewQueryParams().WithFilter("names", "production-quota")
quotas, err := client.OrganizationQuotas().List(ctx, params)
if len(quotas.Resources) > 0 {
    quota := &quotas.Resources[0]
}
```

#### Create Organization Quota

```go
totalMemory := 10240        // 10GB
instanceMemory := 1024      // 1GB
totalInstances := 50
totalAppTasks := 20
totalRoutes := 100
totalReservedPorts := 10
totalServices := 25
totalServiceKeys := 50
totalDomains := 10

createReq := &capi.OrganizationQuotaCreateRequest{
    Name: "production-quota",
    Apps: &capi.OrganizationQuotaApps{
        TotalMemoryInMB:         &totalMemory,
        TotalInstanceMemoryInMB: &instanceMemory,
        TotalInstances:          &totalInstances,
        TotalAppTasks:           &totalAppTasks,
    },
    Services: &capi.OrganizationQuotaServices{
        PaidServicesAllowed:     &[]bool{true}[0],
        TotalServiceInstances:   &totalServices,
        TotalServiceKeys:        &totalServiceKeys,
    },
    Routes: &capi.OrganizationQuotaRoutes{
        TotalRoutes:        &totalRoutes,
        TotalReservedPorts: &totalReservedPorts,
    },
    Domains: &capi.OrganizationQuotaDomains{
        TotalDomains: &totalDomains,
    },
}

quota, err := client.OrganizationQuotas().Create(ctx, createReq)
```

#### Update Organization Quota

```go
newMemory := 20480 // 20GB
newName := "updated-production-quota"
updateReq := &capi.OrganizationQuotaUpdateRequest{
    Name: &newName,
    Apps: &capi.OrganizationQuotaApps{
        TotalMemoryInMB: &newMemory,
    },
}

quota, err := client.OrganizationQuotas().Update(ctx, "quota-guid", updateReq)
```

#### Apply Organization Quota

```go
// Apply quota to multiple organizations
orgGUIDs := []string{"org-1-guid", "org-2-guid", "org-3-guid"}
relationship, err := client.OrganizationQuotas().ApplyToOrganizations(ctx, "quota-guid", orgGUIDs)
```

#### Delete Organization Quota

```go
err := client.OrganizationQuotas().Delete(ctx, "quota-guid")
```

### Space Quotas

Space quotas define resource limits at the space level within an organization.

#### List Space Quotas

```go
// List all space quotas
quotas, err := client.SpaceQuotas().List(ctx, nil)

// Filter by organization
params := capi.NewQueryParams()
params.WithFilter("organization_guids", "org-guid")
quotas, err := client.SpaceQuotas().List(ctx, params)
```

#### Create Space Quota

```go
totalMemory := 2048     // 2GB
totalInstances := 10
totalRoutes := 20
logRateLimit := 1000

createReq := &capi.SpaceQuotaV3CreateRequest{
    Name: "development-quota",
    Relationships: capi.SpaceQuotaRelationships{
        Organization: capi.Relationship{
            Data: &capi.RelationshipData{GUID: "org-guid"},
        },
    },
    Apps: &capi.SpaceQuotaApps{
        TotalMemoryInMB:         &totalMemory,
        TotalInstances:          &totalInstances,
        LogRateLimitInBytesPerSecond: &logRateLimit,
    },
    Routes: &capi.SpaceQuotaRoutes{
        TotalRoutes: &totalRoutes,
    },
}

quota, err := client.SpaceQuotas().Create(ctx, createReq)
```

#### Apply and Remove Space Quotas

```go
// Apply quota to multiple spaces
spaceGUIDs := []string{"space-1-guid", "space-2-guid"}
relationship, err := client.SpaceQuotas().ApplyToSpaces(ctx, "quota-guid", spaceGUIDs)

// Remove quota from a specific space
err := client.SpaceQuotas().RemoveFromSpace(ctx, "quota-guid", "space-guid")
```

## Usage Monitoring

The client provides comprehensive usage monitoring and auditing capabilities.

### Application Usage Events

Application usage events track resource consumption for billing and monitoring.

#### List Application Usage Events

```go
// List all app usage events
events, err := client.AppUsageEvents().List(ctx, nil)

// Filter by application, space, or organization
params := capi.NewQueryParams()
params.WithFilter("app_names", "my-app")
params.WithFilter("space_names", "production")
params.WithFilter("organization_names", "my-org")
events, err := client.AppUsageEvents().List(ctx, params)

// Filter by time range
params.WithFilter("created_ats[gte]", "2023-01-01T00:00:00Z")
params.WithFilter("created_ats[lte]", "2023-12-31T23:59:59Z")
events, err := client.AppUsageEvents().List(ctx, params)
```

#### Get Application Usage Event

```go
event, err := client.AppUsageEvents().Get(ctx, "event-guid")

// Access usage information
fmt.Printf("App: %s\n", event.AppName)
fmt.Printf("State Transition: %s -> %s\n", *event.PreviousState, event.State)
fmt.Printf("Instance Count: %d\n", event.InstanceCount)
fmt.Printf("Memory per Instance: %d MB\n", event.MemoryInMBPerInstance)
fmt.Printf("Space: %s\n", event.SpaceName)
fmt.Printf("Organization: %s\n", event.OrganizationName)
```

#### Purge and Reseed Usage Events

```go
// This operation removes all existing usage events and creates new ones
// based on the current state of applications
err := client.AppUsageEvents().PurgeAndReseed(ctx)
```

### Service Usage Events

Service usage events track service instance lifecycle and usage.

#### List Service Usage Events

```go
// List all service usage events
events, err := client.ServiceUsageEvents().List(ctx, nil)

// Filter by service information
params := capi.NewQueryParams()
params.WithFilter("service_instance_types", "managed_service_instance")
params.WithFilter("service_offering_names", "postgresql")
events, err := client.ServiceUsageEvents().List(ctx, params)
```

#### Get Service Usage Event

```go
event, err := client.ServiceUsageEvents().Get(ctx, "event-guid")

// Access service usage information
fmt.Printf("Service Instance: %s\n", event.ServiceInstanceName)
fmt.Printf("Service Type: %s\n", event.ServiceInstanceType)
fmt.Printf("Service Plan: %s\n", event.ServicePlanName)
fmt.Printf("Service Offering: %s\n", event.ServiceOfferingName)
fmt.Printf("Service Broker: %s\n", event.ServiceBrokerName)
```

### Audit Events

Audit events provide a comprehensive log of all API operations for security and compliance.

#### List Audit Events

```go
// List all audit events
events, err := client.AuditEvents().List(ctx, nil)

// Filter by event type
params := capi.NewQueryParams()
params.WithFilter("types", "audit.app.create,audit.app.update,audit.app.delete")
events, err := client.AuditEvents().List(ctx, params)

// Filter by actor (user)
params.WithFilter("actor_ids", "user-guid")
events, err := client.AuditEvents().List(ctx, params)

// Filter by target resource
params.WithFilter("target_ids", "app-guid")
events, err := client.AuditEvents().List(ctx, params)
```

#### Get Audit Event

```go
event, err := client.AuditEvents().Get(ctx, "event-guid")

// Access audit information
fmt.Printf("Event Type: %s\n", event.Type)
fmt.Printf("Actor: %s (%s)\n", event.Actor.Name, event.Actor.Type)
fmt.Printf("Target: %s (%s)\n", event.Target.Name, event.Target.Type)
fmt.Printf("Space: %s\n", event.Space.Name)
fmt.Printf("Organization: %s\n", event.Organization.Name)

// Access event data
if requestData, ok := event.Data["request"].(map[string]interface{}); ok {
    fmt.Printf("Request Data: %+v\n", requestData)
}
```

### Environment Variable Groups

Environment variable groups allow setting environment variables globally for running and staging applications.

#### Get Environment Variable Groups

```go
// Get running environment variables
runningEnvVars, err := client.EnvironmentVariableGroups().Get(ctx, "running")

// Get staging environment variables  
stagingEnvVars, err := client.EnvironmentVariableGroups().Get(ctx, "staging")

// Access variables
for key, value := range runningEnvVars.Var {
    fmt.Printf("%s=%v\n", key, value)
}
```

#### Update Environment Variable Groups

```go
// Update running environment variables
newVars := map[string]interface{}{
    "LOG_LEVEL":    "info",
    "TIMEOUT":      30,
    "FEATURE_FLAG": true,
}

runningEnvVars, err := client.EnvironmentVariableGroups().Update(ctx, "running", newVars)

// Update staging environment variables
stagingVars := map[string]interface{}{
    "BUILD_CACHE": true,
    "BUILD_ENV":   "production",
}

stagingEnvVars, err := client.EnvironmentVariableGroups().Update(ctx, "staging", stagingVars)
```

## Application Lifecycle

Advanced application lifecycle management features.

### Revisions

Revisions represent immutable snapshots of application configuration.

#### Get Revision

```go
revision, err := client.Revisions().Get(ctx, "revision-guid")

// Access revision information
fmt.Printf("Version: %d\n", revision.Version)
fmt.Printf("Deployable: %t\n", revision.Deployable)
fmt.Printf("Description: %s\n", *revision.Description)
fmt.Printf("Droplet GUID: %s\n", revision.Droplet.GUID)

// Access processes
for processType, process := range revision.Processes {
    fmt.Printf("Process %s: %d instances, %d MB memory\n", 
        processType, process.Instances, process.MemoryInMB)
}
```

#### Update Revision Metadata

```go
updateReq := &capi.RevisionUpdateRequest{
    Metadata: &capi.Metadata{
        Labels: map[string]string{
            "version":     "1.2.0",
            "environment": "production",
            "team":        "backend",
        },
    },
}

revision, err := client.Revisions().Update(ctx, "revision-guid", updateReq)
```

#### Get Revision Environment Variables

```go
envVars, err := client.Revisions().GetEnvironmentVariables(ctx, "revision-guid")

for key, value := range envVars {
    fmt.Printf("%s=%v\n", key, value)
}
```

### Sidecars

Sidecars are additional processes that run alongside application processes.

#### Get Sidecar

```go
sidecar, err := client.Sidecars().Get(ctx, "sidecar-guid")

fmt.Printf("Name: %s\n", sidecar.Name)
fmt.Printf("Command: %s\n", sidecar.Command)
fmt.Printf("Process Types: %v\n", sidecar.ProcessTypes)
if sidecar.MemoryInMB != nil {
    fmt.Printf("Memory: %d MB\n", *sidecar.MemoryInMB)
}
```

#### Update Sidecar

```go
newName := "updated-sidecar"
newCommand := "./updated-command"
newMemory := 256
updateReq := &capi.SidecarUpdateRequest{
    Name:         &newName,
    Command:      &newCommand,
    ProcessTypes: []string{"web", "worker"},
    MemoryInMB:   &newMemory,
}

sidecar, err := client.Sidecars().Update(ctx, "sidecar-guid", updateReq)
```

#### List Sidecars for Process

```go
params := capi.NewQueryParams().WithPerPage(50)
sidecars, err := client.Sidecars().ListForProcess(ctx, "process-guid", params)

for _, sidecar := range sidecars.Resources {
    fmt.Printf("Sidecar: %s (%s)\n", sidecar.Name, sidecar.Command)
}
```

### Resource Matches

Resource matches help optimize package uploads by identifying files already present in the platform.

#### Create Resource Matches

```go
// Prepare resource list for matching
resources := []capi.ResourceMatch{
    {
        Path: "app.js",
        SHA1: "da39a3ee5e6b4b0d3255bfef95601890afd80709",
        Size: 1024,
        Mode: "0644",
    },
    {
        Path: "package.json", 
        SHA1: "356a192b7913b04c54574d18c28d46e6395428ab",
        Size: 512,
        Mode: "0644",
    },
}

createReq := &capi.ResourceMatchesRequest{
    Resources: resources,
}

// Get list of resources that already exist in the platform
matches, err := client.ResourceMatches().Create(ctx, createReq)

// Upload only non-matching resources
var resourcesToUpload []capi.ResourceMatch
for _, resource := range resources {
    found := false
    for _, match := range matches.Resources {
        if resource.SHA1 == match.SHA1 {
            found = true
            break
        }
    }
    if !found {
        resourcesToUpload = append(resourcesToUpload, resource)
    }
}

fmt.Printf("Need to upload %d out of %d resources\n", 
    len(resourcesToUpload), len(resources))
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