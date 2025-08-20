# CLI Examples

This directory contains examples of using the `capi` CLI tool for various Cloud Foundry operations.

## Prerequisites

1. Install the `capi` CLI tool:
   ```bash
   go install github.com/fivetwenty-io/capi-client/cmd/capi@latest
   ```

2. Have access to a Cloud Foundry environment with:
   - API endpoint URL
   - Valid credentials (username/password or client credentials)

## Basic Usage Examples

### Login and Authentication

```bash
# Interactive login (prompts for credentials)
capi login

# Login with API endpoint
capi login -a https://api.your-cf-domain.com

# Login with username and password
capi login -a https://api.your-cf-domain.com -u username -p password

# Login with client credentials
capi login -a https://api.your-cf-domain.com --client-id my-client --client-secret my-secret

# Login with SSO
capi login -a https://api.your-cf-domain.com --sso

# Skip SSL validation (for development environments)
capi login -a https://api.your-cf-domain.com --skip-ssl-validation
```

### Targeting Organizations and Spaces

```bash
# Show current target
capi target

# Target an organization
capi target -o my-org

# Target an organization and space
capi target -o my-org -s my-space

# List available organizations after login
capi orgs list

# List spaces in current organization
capi spaces list
```

## Resource Management Examples

### Organizations

```bash
# List all organizations
capi orgs list

# List organizations in JSON format
capi orgs list --output json

# Get detailed information about an organization
capi orgs get my-org

# Create a new organization
capi orgs create new-org

# Update organization metadata
capi orgs update my-org --name updated-name

# Delete an organization
capi orgs delete my-org
```

### Spaces

```bash
# List spaces in current organization
capi spaces list

# List spaces in specific organization
capi spaces list -o my-org

# Get space details
capi spaces get my-space

# Create a new space
capi spaces create dev-space -o my-org

# Update space
capi spaces update my-space --name updated-space

# Delete a space
capi spaces delete my-space
```

### Applications

```bash
# List all applications
capi apps list

# List applications in specific space
capi apps list -s my-space

# Get application details
capi apps get my-app

# Get application details in YAML format
capi apps get my-app --output yaml

# Create an application
capi apps create my-app -s my-space

# Start an application
capi apps start my-app

# Stop an application
capi apps stop my-app

# Restart an application
capi apps restart my-app

# Scale an application
capi apps scale my-app --instances 3
capi apps scale my-app --memory 512M
capi apps scale my-app --disk 1G
capi apps scale my-app --instances 2 --memory 256M --disk 512M

# Get application environment variables
capi apps env my-app

# Set environment variables
capi apps set-env my-app KEY value
capi apps set-env my-app DEBUG true

# Unset environment variables
capi apps unset-env my-app KEY

# Get application stats
capi apps stats my-app

# Delete an application
capi apps delete my-app
```

### Services

```bash
# List service offerings
capi services list-offerings

# List service plans for an offering
capi services list-plans -o mysql

# List service instances
capi services list

# Get service instance details
capi services get my-service

# Create a service instance
capi services create mysql small my-service

# Update a service instance
capi services update my-service --plan medium

# Delete a service instance
capi services delete my-service

# Bind service to application
capi services bind my-service my-app

# Unbind service from application
capi services unbind my-service my-app
```

### Users and Roles

```bash
# List users in organization
capi users list -o my-org

# List users in space
capi users list -s my-space

# Get user details
capi users get user@example.com

# Assign organization role
capi roles assign user@example.com my-org OrgManager

# Remove organization role
capi roles remove user@example.com my-org OrgManager

# Assign space role
capi roles assign user@example.com my-org my-space SpaceDeveloper

# List user roles
capi roles list user@example.com
```

## Advanced Usage Examples

### Using Different Output Formats

```bash
# Default table format
capi orgs list

# JSON output for scripting
capi orgs list --output json | jq '.resources[].name'

# YAML output
capi orgs list --output yaml
```

### Filtering and Querying

```bash
# Filter applications by name
capi apps list --filter name=my-app

# Filter by multiple criteria (if supported)
capi apps list --filter state=STARTED,name=my-*
```

### Configuration Management

```bash
# Show current configuration
capi config show

# Show configuration in JSON
capi config show --output json

# Set default output format
capi config set output json

# Set default organization
capi config set organization my-org

# Unset a configuration value
capi config unset organization

# Clear all configuration
capi config clear
```

### Debugging and Verbose Output

```bash
# Enable verbose output
capi --verbose orgs list

# Disable colored output
capi --no-color orgs list

# Use custom config file
capi --config /path/to/config.yml orgs list
```

## Scripting Examples

### Bash Script for Deployment

```bash
#!/bin/bash

set -e

# Configuration
API_ENDPOINT="https://api.your-cf-domain.com"
ORG_NAME="my-org"
SPACE_NAME="production"
APP_NAME="my-app"

# Login
capi login -a "$API_ENDPOINT" -u "$CF_USERNAME" -p "$CF_PASSWORD"

# Target org and space
capi target -o "$ORG_NAME" -s "$SPACE_NAME"

# Check if app exists
if capi apps get "$APP_NAME" >/dev/null 2>&1; then
    echo "Application $APP_NAME exists, updating..."
    capi apps update "$APP_NAME"
else
    echo "Creating new application $APP_NAME..."
    capi apps create "$APP_NAME"
fi

# Start the application
capi apps start "$APP_NAME"

# Wait for application to start
echo "Waiting for application to start..."
sleep 30

# Check application stats
capi apps stats "$APP_NAME"

echo "Deployment completed successfully!"
```

### PowerShell Script for Windows

```powershell
# Configuration
$ApiEndpoint = "https://api.your-cf-domain.com"
$OrgName = "my-org"
$SpaceName = "development"

# Login
capi login -a $ApiEndpoint -u $env:CF_USERNAME -p $env:CF_PASSWORD

# Target org and space
capi target -o $OrgName -s $SpaceName

# Get list of applications in JSON format
$apps = capi apps list --output json | ConvertFrom-Json

# Process each application
foreach ($app in $apps.resources) {
    Write-Host "Processing application: $($app.name)"
    
    # Get application stats
    $stats = capi apps stats $app.name --output json | ConvertFrom-Json
    
    # Check if application is running
    if ($stats.state -eq "RUNNING") {
        Write-Host "  Application is running with $($stats.instances) instances"
    } else {
        Write-Host "  Application is not running (state: $($stats.state))"
    }
}
```

### JSON Processing with jq

```bash
# Get all organization names
capi orgs list --output json | jq -r '.resources[].name'

# Get applications with their states
capi apps list --output json | jq -r '.resources[] | "\(.name): \(.state)"'

# Count applications by state
capi apps list --output json | jq '.resources | group_by(.state) | .[] | {state: .[0].state, count: length}'

# Get memory usage across all applications
capi apps list --output json | jq '[.resources[] | .memory] | add'
```

## Error Handling

### Common Error Scenarios

```bash
# Handle authentication errors
if ! capi login -a "$API_ENDPOINT" -u "$USERNAME" -p "$PASSWORD"; then
    echo "Login failed. Please check your credentials."
    exit 1
fi

# Check if organization exists before targeting
if ! capi orgs get "$ORG_NAME" >/dev/null 2>&1; then
    echo "Organization $ORG_NAME does not exist or you don't have access."
    exit 1
fi

# Verify application exists before operations
if ! capi apps get "$APP_NAME" >/dev/null 2>&1; then
    echo "Application $APP_NAME not found."
    exit 1
fi
```

### Exit Code Handling

```bash
# Check exit codes for scripting
capi apps start my-app
if [ $? -eq 0 ]; then
    echo "Application started successfully"
else
    echo "Failed to start application"
    exit 1
fi
```

## Tips and Best Practices

1. **Use JSON output for scripting**: The `--output json` flag provides structured data that's easy to parse with tools like `jq`.

2. **Set default configuration**: Use `capi config set` to avoid repeating common flags.

3. **Use environment variables**: Set `CAPI_API`, `CAPI_USERNAME`, `CAPI_PASSWORD` environment variables to avoid passing credentials on command line.

4. **Check exit codes**: Always check exit codes in scripts to handle errors appropriately.

5. **Use verbose mode for debugging**: Add `--verbose` flag to see detailed HTTP requests and responses.

6. **Configuration file**: Store common settings in `~/.capi/config.yml` for consistency across sessions.