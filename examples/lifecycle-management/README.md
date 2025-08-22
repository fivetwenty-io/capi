# Application Lifecycle Management Examples

This directory contains examples for advanced Cloud Foundry application lifecycle management using the CAPI client library.

## Features Demonstrated

- **Revisions**: Immutable snapshots of application configuration
- **Sidecars**: Additional processes running alongside applications
- **Resource Matches**: Optimize package uploads by reusing existing resources
- **Advanced Deployment Patterns**: Blue-green deployments and rollback strategies

## Running the Examples

### Prerequisites

1. Set up environment variables:
```bash
export CF_USERNAME="your-username"
export CF_PASSWORD="your-password"
```

2. Ensure you have applications deployed in your CF environment for demonstration

### Run the Example

```bash
go run main.go https://api.your-cf-domain.com
```

## Example Operations

### Revisions

Application revisions provide version control for your app configuration:

1. **Revision Details**: View immutable snapshots of app configuration
2. **Environment Variables**: Access revision-specific environment variables
3. **Process Configuration**: Review process settings for each revision
4. **Metadata Management**: Add labels and annotations for tracking

#### Revision Use Cases

- **Rollback Capability**: Quickly revert to previous working configurations
- **Change Tracking**: Audit trail of configuration changes
- **Deployment Verification**: Ensure deployment integrity
- **A/B Testing**: Compare different application configurations

### Sidecars

Sidecars are additional processes that run alongside your application:

1. **Process Discovery**: Find sidecars associated with application processes
2. **Sidecar Details**: View configuration and resource allocation
3. **Sidecar Management**: Update sidecar configuration
4. **Process Types**: Understand which process types use sidecars

#### Sidecar Use Cases

- **Logging Agents**: Run log collection processes alongside apps
- **Monitoring**: Health check and metrics collection
- **Security**: Run security scanning or compliance tools
- **Data Processing**: Background data processing tasks

### Resource Matches

Resource matching optimizes application deployments:

1. **File Matching**: Identify files already present in the platform
2. **Upload Optimization**: Only upload new or changed files
3. **Bandwidth Savings**: Reduce deployment time and bandwidth usage
4. **Checksum Verification**: Ensure file integrity

#### Resource Match Benefits

- **Faster Deployments**: Skip uploading existing files
- **Reduced Bandwidth**: Significant savings for large applications
- **Cache Utilization**: Leverage platform-level file caching
- **Deployment Consistency**: Ensure identical files across deployments

## Advanced Patterns

### Blue-Green Deployment with Revisions

```go
// Deploy new version while keeping old version running
func blueGreenDeploy(client capi.Client, appGUID string) error {
    ctx := context.Background()
    
    // Get current revision (blue)
    currentRevision, err := client.Apps().GetCurrentRevision(ctx, appGUID)
    if err != nil {
        return err
    }
    
    // Deploy new version (green)
    deployment, err := client.Deployments().Create(ctx, &capi.DeploymentCreateRequest{
        Strategy: "rolling",
        Relationships: capi.DeploymentRelationships{
            App: capi.Relationship{Data: &capi.RelationshipData{GUID: appGUID}},
        },
    })
    if err != nil {
        return err
    }
    
    // Monitor deployment
    err = client.Jobs().PollUntilComplete(ctx, deployment.GUID, 10*time.Minute)
    if err != nil {
        // Rollback to previous revision if deployment fails
        rollbackReq := &capi.DeploymentCreateRequest{
            Relationships: capi.DeploymentRelationships{
                App: capi.Relationship{Data: &capi.RelationshipData{GUID: appGUID}},
            },
            Revision: &capi.Relationship{
                Data: &capi.RelationshipData{GUID: currentRevision.GUID},
            },
        }
        client.Deployments().Create(ctx, rollbackReq)
        return err
    }
    
    return nil
}
```

### Sidecar-Based Monitoring

```go
// Add monitoring sidecar to application
func addMonitoringSidecar(client capi.Client, appGUID string) error {
    ctx := context.Background()
    
    createReq := &capi.SidecarCreateRequest{
        Name:         "metrics-collector",
        Command:      "./metrics-agent --port 8080",
        ProcessTypes: []string{"web"},
        MemoryInMB:   64,
        Relationships: capi.SidecarRelationships{
            App: capi.Relationship{Data: &capi.RelationshipData{GUID: appGUID}},
        },
    }
    
    sidecar, err := client.Sidecars().Create(ctx, createReq)
    if err != nil {
        return err
    }
    
    fmt.Printf("Added monitoring sidecar: %s\n", sidecar.GUID)
    return nil
}
```

### Optimized Package Upload

```go
// Optimize package upload using resource matches
func optimizedPackageUpload(client capi.Client, packageGUID string, files []FileInfo) error {
    ctx := context.Background()
    
    // Prepare resource list
    var resources []capi.ResourceMatch
    for _, file := range files {
        resources = append(resources, capi.ResourceMatch{
            Path: file.Path,
            SHA1: file.SHA1,
            Size: file.Size,
            Mode: file.Mode,
        })
    }
    
    // Check for existing resources
    matches, err := client.ResourceMatches().Create(ctx, &capi.ResourceMatchesRequest{
        Resources: resources,
    })
    if err != nil {
        return err
    }
    
    // Determine which files need to be uploaded
    var filesToUpload []FileInfo
    matchedFiles := make(map[string]bool)
    
    for _, match := range matches.Resources {
        matchedFiles[match.SHA1] = true
    }
    
    for _, file := range files {
        if !matchedFiles[file.SHA1] {
            filesToUpload = append(filesToUpload, file)
        }
    }
    
    fmt.Printf("Optimization: uploading %d/%d files (%.1f%% savings)\n",
        len(filesToUpload), len(files),
        float64(len(matches.Resources))/float64(len(files))*100)
    
    // Upload only non-matching files
    return uploadFiles(client, packageGUID, filesToUpload)
}
```

## Monitoring and Observability

Use these patterns for monitoring and observability:

1. **Metric Collection**: Use sidecars to collect application metrics
2. **Log Aggregation**: Deploy logging sidecars for centralized logging
3. **Health Checks**: Implement custom health check processes
4. **Performance Monitoring**: Track resource usage patterns
5. **Security Scanning**: Run security analysis as sidecars

## CLI Equivalent

You can perform the same operations using the CLI:

```bash
# Revisions
capi revisions get revision-guid
capi revisions get-env revision-guid
capi revisions update revision-guid --metadata version=1.2.0

# Sidecars
capi sidecars list-for-process process-guid
capi sidecars get sidecar-guid
capi sidecars update sidecar-guid --memory 128

# Resource matches
capi resource-matches create resource-list.json
```