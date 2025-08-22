# Quota Management Examples

This directory contains examples for managing Cloud Foundry quotas using the CAPI client library.

## Features Demonstrated

- **Organization Quotas**: Create, read, update, delete, and apply organization-level resource quotas
- **Space Quotas**: Manage space-level quotas within organizations
- **Resource Limits**: Configure limits for memory, instances, services, routes, and domains
- **Quota Application**: Apply quotas to multiple organizations or spaces

## Running the Examples

### Prerequisites

1. Set up environment variables:
```bash
export CF_USERNAME="your-username"
export CF_PASSWORD="your-password"
```

2. Ensure you have admin privileges or appropriate permissions to manage quotas

### Run the Example

```bash
go run main.go https://api.your-cf-domain.com
```

## Example Operations

### Organization Quotas

The example demonstrates:

1. **List Quotas**: View all existing organization quotas
2. **Create Quota**: Create a new quota with specific resource limits
3. **Get Quota**: Retrieve detailed quota information
4. **Update Quota**: Modify quota settings
5. **Delete Quota**: Clean up demo quota

### Space Quotas

The example shows:

1. **List Space Quotas**: View quotas within an organization
2. **Create Space Quota**: Create quotas for specific spaces
3. **Apply Quotas**: Associate quotas with spaces
4. **Remove Quotas**: Disassociate quotas from spaces

## Resource Limits

Quotas can control:

- **Application Limits**:
  - Total memory across all apps
  - Memory per application instance
  - Total number of application instances
  - Number of application tasks
  - Log rate limiting

- **Service Limits**:
  - Whether paid services are allowed
  - Total service instances
  - Total service keys

- **Route Limits**:
  - Total number of routes
  - Total reserved ports

- **Domain Limits**:
  - Total number of private domains

## Best Practices

1. **Start with Conservative Limits**: Begin with lower resource limits and increase as needed
2. **Monitor Usage**: Use usage events to understand actual consumption patterns
3. **Organization vs Space Quotas**: Use organization quotas for overall governance, space quotas for granular control
4. **Regular Review**: Periodically review and adjust quotas based on usage patterns
5. **Automation**: Consider automating quota management based on team size or project requirements

## CLI Equivalent

You can perform the same operations using the CLI:

```bash
# Organization quotas
capi org-quotas list
capi org-quotas create --name production --total-memory 10240 --instances 50
capi org-quotas apply production my-org-1 my-org-2

# Space quotas  
capi space-quotas create --name dev-quota --org my-org --total-memory 2048
capi space-quotas apply dev-quota my-space-1 my-space-2
```