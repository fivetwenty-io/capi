# Usage Monitoring Examples

This directory contains examples for monitoring Cloud Foundry usage using the CAPI client library.

## Features Demonstrated

- **Application Usage Events**: Track application resource consumption for billing and monitoring
- **Service Usage Events**: Monitor service instance lifecycle and usage
- **Audit Events**: Security and compliance logging of all API operations
- **Environment Variable Groups**: Global environment variable management
- **Usage Analytics**: Pattern analysis and reporting capabilities

## Running the Examples

### Prerequisites

1. Set up environment variables:
```bash
export CF_USERNAME="your-username"
export CF_PASSWORD="your-password"
```

2. Ensure you have appropriate permissions to view usage events and audit logs

### Run the Example

```bash
go run main.go https://api.your-cf-domain.com
```

## Example Operations

### Application Usage Events

The example demonstrates:

1. **List Events**: View recent application usage events
2. **Filter Events**: Filter by app name, space, organization, or time range
3. **Event Details**: Get comprehensive event information including:
   - Application state transitions (STOPPED → STARTED)
   - Instance count changes
   - Memory allocation changes
   - Process type information
   - Buildpack and package details

### Service Usage Events

The example shows:

1. **Service Events**: Track service instance lifecycle events
2. **Event Information**: Access service details including:
   - Service instance type (managed vs user-provided)
   - Service offering and plan information
   - Service broker details
   - State transitions

### Audit Events

Security and compliance tracking:

1. **Security Logging**: View all API operations for security monitoring
2. **Event Filtering**: Filter by event type, actor, or target resource
3. **Compliance Reporting**: Access actor, target, and operation details
4. **Event Data**: Review request parameters and context

### Environment Variable Groups

Global environment management:

1. **Running Variables**: Variables available to all running applications
2. **Staging Variables**: Variables available during application staging
3. **Variable Updates**: Modify global environment settings
4. **Impact Assessment**: Understand which applications are affected

## Usage Analytics

This example can be extended for usage analytics:

```go
// Calculate total memory usage by organization
func calculateOrgMemoryUsage(events []capi.AppUsageEvent) map[string]int {
    orgUsage := make(map[string]int)
    
    for _, event := range events {
        if event.State == "STARTED" {
            totalMemory := event.InstanceCount * event.MemoryInMBPerInstance
            orgUsage[event.OrganizationName] += totalMemory
        }
    }
    
    return orgUsage
}

// Track application scaling patterns
func trackScalingPatterns(events []capi.AppUsageEvent) {
    for _, event := range events {
        if event.PreviousInstanceCount != nil {
            change := event.InstanceCount - *event.PreviousInstanceCount
            if change > 0 {
                fmt.Printf("App %s scaled UP by %d instances\n", event.AppName, change)
            } else if change < 0 {
                fmt.Printf("App %s scaled DOWN by %d instances\n", event.AppName, -change)
            }
        }
    }
}
```

## Security Considerations

When working with usage events and audit logs:

1. **Access Control**: Ensure proper RBAC permissions for usage event access
2. **Data Sensitivity**: Usage events may contain sensitive application information
3. **Retention Policies**: Consider data retention requirements for compliance
4. **Log Aggregation**: Integrate with external logging systems for long-term storage

## CLI Equivalent

You can perform the same operations using the CLI:

```bash
# Application usage events
capi app-usage-events list --per-page 50
capi app-usage-events list --app-name my-app --start-time 2023-01-01T00:00:00Z
capi app-usage-events get event-guid

# Service usage events
capi service-usage-events list --per-page 50
capi service-usage-events get event-guid

# Audit events
capi audit-events list --per-page 50
capi audit-events list --types audit.app.create,audit.app.update
capi audit-events get event-guid

# Environment variable groups
capi env-var-groups get running
capi env-var-groups get staging
capi env-var-groups update running LOG_LEVEL=info TIMEOUT=30
```

## Performance Considerations

For large deployments:

1. **Pagination**: Use appropriate page sizes to balance performance and memory usage
2. **Filtering**: Apply filters to reduce the dataset size
3. **Batch Processing**: Process events in batches for analytics
4. **Caching**: Enable caching for frequently accessed data
5. **Time Windows**: Use time-based filtering for recent events