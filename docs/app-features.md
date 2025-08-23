# App Features Management

App features provide fine-grained control over application-level functionality like SSH access, revisions, and other capabilities.

## Commands

```bash
# List all features for an application
capi apps features list APP_NAME_OR_GUID

# Get details for a specific feature
capi apps features get APP_NAME_OR_GUID FEATURE_NAME

# Enable a feature for an application
capi apps features enable APP_NAME_OR_GUID FEATURE_NAME

# Disable a feature for an application  
capi apps features disable APP_NAME_OR_GUID FEATURE_NAME
```

## Examples

```bash
# List features for an app
capi apps features list my-app

# Enable SSH access for an app
capi apps features enable my-app ssh

# Disable revisions for an app
capi apps features disable my-app revisions

# Check if SSH is enabled
capi apps features get my-app ssh
```

## Common App Features

- `ssh` - SSH access to application containers
- `revisions` - Application revision tracking
- `log-cache` - Enhanced logging capabilities

## Output Formats

All commands support three output formats:

```bash
# Human-readable table (default)
capi apps features list my-app

# JSON for automation and scripting
capi apps features list my-app --output json

# YAML for configuration management
capi apps features list my-app --output yaml
```

## Resource Resolution

Commands support both GUIDs and names for user convenience:

```bash
# Using GUID
capi apps features enable app-123-456 ssh

# Using name (searches in targeted space)
capi apps features enable my-app ssh
```

## Error Handling

All commands provide clear error messages and helpful suggestions:

```bash
$ capi apps features enable non-existent-app ssh
Error: application 'non-existent-app' not found
```

## Security Considerations

- **Permissions**: All operations respect Cloud Foundry RBAC permissions
- **Authentication**: All operations require proper CF authentication
- **Audit Trail**: Operations are logged in CF audit events

## Troubleshooting

### Common Issues

**Permission Denied:**
```bash
Error: insufficient permissions to perform operation
Solution: Ensure you have appropriate space/org roles
```

**Resource Not Found:**
```bash  
Error: application 'my-app' not found
Solution: Check app name and ensure you're targeting the correct space
```

### Getting Help

```bash
# Get help for any command
capi apps features --help  

# Get detailed help for subcommands
capi apps features enable --help
```

## API Compatibility

All implemented features are compatible with Cloud Foundry API v3.199.0 and follow the official CF API specification for:

- `/v3/apps/{guid}/features` - App feature management

For complete API documentation, see the [Cloud Foundry API v3 specification](https://v3-apidocs.cloudfoundry.org/).