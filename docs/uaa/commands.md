# UAA User Management Commands

The `capi uaa` commands provide comprehensive UAA (User Account and Authentication) user management functionality. These commands are organized into logical sub-command groups for better usability and discoverability.

## New Command Structure (v2.0)

Commands are now organized hierarchically:
- **`capi uaa user`** - User management operations
- **`capi uaa group`** - Group management operations  
- **`capi uaa client`** - OAuth client management
- **`capi uaa token`** - Token operations
- **`capi uaa batch`** - Batch operations and utilities
- **`capi uaa integration`** - Integration and compatibility utilities

**Legacy commands remain available but are deprecated.** Use the new hierarchical structure for new implementations.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Quick Start](#quick-start)
3. [Context Management](#context-management)
4. [Authentication & Tokens](#authentication--tokens)
5. [User Management](#user-management)
6. [Group Management](#group-management)
7. [OAuth Client Management](#oauth-client-management)
8. [Utility Commands](#utility-commands)
9. [Output Formats](#output-formats)
10. [Security Considerations](#security-considerations)
11. [Troubleshooting](#troubleshooting)

## Prerequisites

- Access to a UAA server (part of Cloud Foundry deployment)
- Admin credentials or appropriate client credentials
- Network connectivity to the UAA endpoint

## Quick Start

### 1. Set UAA Target

```bash
# Set UAA endpoint (can be inferred from CF API if not specified)
capi uaa target https://uaa.your-cf-domain.com

# Check current context
capi uaa context
```

### 2. Authenticate

Choose one authentication method:

```bash
# Option 1: Client credentials (machine-to-machine) - NEW STRUCTURE
capi uaa token get-client-credentials --client-id admin --client-secret admin-secret

# Option 2: Username/password (user authentication) - NEW STRUCTURE
capi uaa token get-password --username admin --password admin-pass --client-id cf

# Option 3: Authorization code flow (interactive) - NEW STRUCTURE
capi uaa token get-authcode --client-id my-client --redirect-uri http://localhost:8080/callback

# Legacy commands (still work but deprecated):
# capi uaa get-client-credentials-token --client-id admin --client-secret admin-secret
# capi uaa get-password-token --username admin --password admin-pass --client-id cf
# capi uaa get-authcode-token --client-id my-client --redirect-uri http://localhost:8080/callback
```

### 3. Basic Operations

```bash
# List users - NEW STRUCTURE
capi uaa user list

# Create a user - NEW STRUCTURE
capi uaa user create john.doe --email john.doe@example.com --password SecurePass123!

# Get user info - NEW STRUCTURE
capi uaa user get john.doe

# Create a group - NEW STRUCTURE
capi uaa group create developers --description "Development team"

# Add user to group - NEW STRUCTURE
capi uaa group add-member developers john.doe

# Legacy commands (still work but deprecated):
# capi uaa list-users
# capi uaa create-user john.doe --email john.doe@example.com
# capi uaa get-user john.doe
# capi uaa create-group developers --description "Development team"
# capi uaa add-member developers john.doe
```

## Context Management

### Set UAA Target

```bash
# Set UAA endpoint explicitly
capi uaa target https://uaa.example.com

# UAA endpoint can also be inferred from CF API target
capi target -a https://api.example.com
capi uaa context  # Will show inferred UAA endpoint
```

### View Current Context

```bash
# Show current UAA context
capi uaa context

# Show context in JSON format
capi uaa context --output json

# Show context in YAML format
capi uaa context --output yaml
```

### Get UAA Server Information

```bash
# Get UAA server info
capi uaa info

# Get server info in JSON
capi uaa info --output json

# Get UAA version
capi uaa version
```

## Authentication & Tokens

### Client Credentials Grant

Best for machine-to-machine authentication:

```bash
# Basic client credentials - NEW STRUCTURE
capi uaa token get-client-credentials --client-id my-client --client-secret my-secret

# Specify token format - NEW STRUCTURE
capi uaa token get-client-credentials \
    --client-id my-client \
    --client-secret my-secret \
    --token-format jwt

# Request specific scopes - NEW STRUCTURE
capi uaa token get-client-credentials \
    --client-id my-client \
    --client-secret my-secret \
    --scope "uaa.admin,scim.read"

# Legacy commands (deprecated):
# capi uaa get-client-credentials-token --client-id my-client --client-secret my-secret
```

### Password Grant

For user authentication:

```bash
# Username/password authentication - NEW STRUCTURE
capi uaa token get-password \
    --username admin \
    --password admin-pass \
    --client-id cf

# Specify client credentials - NEW STRUCTURE
capi uaa token get-password \
    --username user@example.com \
    --password user-pass \
    --client-id my-client \
    --client-secret my-secret

# Request specific scopes - NEW STRUCTURE
capi uaa token get-password \
    --username admin \
    --password admin-pass \
    --client-id cf \
    --scope "scim.read,scim.write"

# Legacy commands (deprecated):
# capi uaa get-password-token --username admin --password admin-pass --client-id cf
```

### Authorization Code Grant

For web applications:

```bash
# Start authorization code flow - NEW STRUCTURE
capi uaa token get-authcode \
    --client-id my-web-app \
    --client-secret app-secret \
    --redirect-uri https://myapp.com/callback \
    --code AUTHORIZATION_CODE_FROM_CALLBACK

# With PKCE (Proof Key for Code Exchange) - NEW STRUCTURE
capi uaa token get-authcode \
    --client-id my-spa \
    --redirect-uri https://spa.com/callback \
    --code AUTHORIZATION_CODE \
    --code-verifier CODE_VERIFIER

# Legacy commands (deprecated):
# capi uaa get-authcode-token --client-id my-web-app --client-secret app-secret
```

### Token Refresh

```bash
# Refresh current token - NEW STRUCTURE
capi uaa token refresh

# Refresh with specific refresh token - NEW STRUCTURE
capi uaa token refresh --refresh-token REFRESH_TOKEN_VALUE

# Refresh with client credentials - NEW STRUCTURE
capi uaa token refresh \
    --refresh-token REFRESH_TOKEN \
    --client-id my-client \
    --client-secret my-secret

# Legacy commands (deprecated):
# capi uaa refresh-token
```

### Token Keys

```bash
# Get current JWT signing key - NEW STRUCTURE
capi uaa token get-key

# Get all JWT signing keys (including rotated keys) - NEW STRUCTURE
capi uaa token get-keys

# Get keys in JSON format - NEW STRUCTURE
capi uaa token get-keys --output json

# Legacy commands (deprecated):
# capi uaa get-token-key
# capi uaa get-token-keys
```

## User Management

> **Note**: All user management commands now use the `capi uaa user` sub-command structure. Legacy commands (`create-user`, `list-users`, etc.) are deprecated but still functional.

### Create Users

```bash
# Basic user creation - NEW STRUCTURE
capi uaa user create john.doe \
    --email john.doe@example.com \
    --password SecurePass123!

# Full user creation with all attributes - NEW STRUCTURE
capi uaa user create jane.smith \
    --email jane.smith@example.com \
    --password AnotherPass456! \
    --given-name Jane \
    --family-name Smith \
    --phone-number "+1-555-0123" \
    --origin uaa \
    --active \
    --verified

# Create user with external identity provider - NEW STRUCTURE
capi uaa user create external.user \
    --email external@ldap.company.com \
    --origin ldap \
    --external-id "cn=external,ou=users,dc=company,dc=com"

# Legacy commands (deprecated):
# capi uaa create-user john.doe --email john.doe@example.com --password SecurePass123!
```

### List Users

```bash
# List all users - NEW STRUCTURE
capi uaa user list

# List with pagination - NEW STRUCTURE
capi uaa user list --count 50 --start-index 100

# List all users (automatic pagination) - NEW STRUCTURE
capi uaa user list --all

# Filter users with SCIM expressions - NEW STRUCTURE
capi uaa user list --filter 'active eq true'
capi uaa user list --filter 'email co "example.com"'
capi uaa user list --filter 'userName sw "admin"'
capi uaa user list --filter 'origin eq "ldap"'
capi uaa user list --filter 'meta.created gt "2023-01-01T00:00:00.000Z"'

# Sort users - NEW STRUCTURE
capi uaa user list --sort-by userName --sort-order ascending
capi uaa user list --sort-by meta.created --sort-order descending

# Select specific attributes - NEW STRUCTURE
capi uaa user list --attributes userName,email,active

# Combined filtering and sorting - NEW STRUCTURE
capi uaa user list \
    --filter 'active eq true and email co "company.com"' \
    --sort-by userName \
    --count 25 \
    --attributes userName,email,name.givenName,name.familyName

# Legacy commands (deprecated):
# capi uaa list-users --filter 'active eq true'
```

### Get User Details

```bash
# Get user by username - NEW STRUCTURE
capi uaa user get john.doe

# Get user by UUID - NEW STRUCTURE
capi uaa user get 12345678-1234-1234-1234-123456789abc

# Get user with specific attributes - NEW STRUCTURE
capi uaa user get jane.smith --attributes userName,email,groups

# Get user in JSON format - NEW STRUCTURE
capi uaa user get john.doe --output json

# Legacy commands (deprecated):
# capi uaa get-user john.doe
```

### Update Users

```bash
# Update user attributes - NEW STRUCTURE
capi uaa user update john.doe \
    --email new.email@example.com \
    --given-name John \
    --family-name Doe-Updated \
    --phone-number "+1-555-9999"

# Activate/deactivate user - NEW STRUCTURE
capi uaa user update john.doe --active
capi uaa user update john.doe --no-active

# Verify/unverify user email - NEW STRUCTURE
capi uaa user update john.doe --verified
capi uaa user update john.doe --no-verified

# Update password (admin operation) - NEW STRUCTURE
capi uaa user update john.doe --password NewSecurePass789!

# Legacy commands (deprecated):
# capi uaa update-user john.doe --email new.email@example.com
```

### User Status Management

```bash
# Activate user account - NEW STRUCTURE
capi uaa user activate john.doe

# Deactivate user account - NEW STRUCTURE
capi uaa user deactivate john.doe

# Check activation status - NEW STRUCTURE
capi uaa user get john.doe --attributes active

# Legacy commands (deprecated):
# capi uaa activate-user john.doe
# capi uaa deactivate-user john.doe
```

### Delete Users

```bash
# Delete user with confirmation - NEW STRUCTURE
capi uaa user delete john.doe

# Force delete without confirmation - NEW STRUCTURE
capi uaa user delete john.doe --force

# Delete user by UUID - NEW STRUCTURE
capi uaa user delete 12345678-1234-1234-1234-123456789abc --force

# Legacy commands (deprecated):
# capi uaa delete-user john.doe --force
```

## Group Management

> **Note**: All group management commands now use the `capi uaa group` sub-command structure. Legacy commands (`create-group`, `add-member`, etc.) are deprecated but still functional.

### Create Groups

```bash
# Basic group creation - NEW STRUCTURE
capi uaa group create developers

# Group with description - NEW STRUCTURE
capi uaa group create admins --description "System administrators"

# Group with initial members - NEW STRUCTURE
capi uaa group create qa-team \
    --description "Quality assurance team" \
    --members john.doe,jane.smith

# Legacy commands (deprecated):
# capi uaa create-group developers --description "Development team"
```

### List Groups

```bash
# List all groups - NEW STRUCTURE
capi uaa group list

# Filter groups - NEW STRUCTURE
capi uaa group list --filter 'displayName sw "admin"'
capi uaa group list --filter 'meta.created gt "2023-01-01T00:00:00.000Z"'

# List with pagination - NEW STRUCTURE
capi uaa group list --count 20 --start-index 0

# Get all groups - NEW STRUCTURE
capi uaa group list --all

# Legacy commands (deprecated):
# capi uaa list-groups --filter 'displayName sw "admin"'
```

### Get Group Details

```bash
# Get group by name - NEW STRUCTURE
capi uaa group get developers

# Get group by UUID - NEW STRUCTURE
capi uaa group get 87654321-4321-4321-4321-210987654321

# Get group in JSON format - NEW STRUCTURE
capi uaa group get developers --output json

# Legacy commands (deprecated):
# capi uaa get-group developers
```

### Group Membership

```bash
# Add user to group (by username) - NEW STRUCTURE
capi uaa group add-member developers john.doe

# Add user to group (by UUID) - NEW STRUCTURE
capi uaa group add-member developers 12345678-1234-1234-1234-123456789abc

# Add member with specific origin - NEW STRUCTURE
capi uaa group add-member developers external.user --origin ldap

# Add member with type - NEW STRUCTURE
capi uaa group add-member developers service.account --type client

# Remove user from group - NEW STRUCTURE
capi uaa group remove-member developers john.doe

# Remove member by UUID - NEW STRUCTURE
capi uaa group remove-member developers 12345678-1234-1234-1234-123456789abc

# Legacy commands (deprecated):
# capi uaa add-member developers john.doe
# capi uaa remove-member developers john.doe
```

### External Group Mapping

```bash
# Map external group to UAA group - NEW STRUCTURE
capi uaa group map \
    --group developers \
    --external-group "CN=Developers,OU=Groups,DC=company,DC=com" \
    --origin ldap

# Map SAML group - NEW STRUCTURE
capi uaa group map \
    --group admins \
    --external-group "admin-users" \
    --origin saml

# List group mappings - NEW STRUCTURE
capi uaa group list-mappings

# List mappings for specific origin - NEW STRUCTURE
capi uaa group list-mappings --origin ldap

# Remove group mapping - NEW STRUCTURE
capi uaa group unmap \
    --group developers \
    --external-group "CN=Developers,OU=Groups,DC=company,DC=com" \
    --origin ldap

# Legacy commands (deprecated):
# capi uaa map-group --group developers --external-group "CN=Developers,DC=company" --origin ldap
# capi uaa list-group-mappings
# capi uaa unmap-group --group developers --external-group "CN=Developers,DC=company" --origin ldap
```

## OAuth Client Management

> **Note**: All OAuth client management commands now use the `capi uaa client` sub-command structure. Legacy commands (`create-client`, `list-clients`, etc.) are deprecated but still functional.

### Create OAuth Clients

```bash
# Basic client creation - NEW STRUCTURE
capi uaa client create my-app \
    --secret app-secret-123 \
    --authorized-grant-types client_credentials

# Web application client - NEW STRUCTURE
capi uaa client create web-app \
    --secret web-secret \
    --name "My Web Application" \
    --authorized-grant-types authorization_code,refresh_token \
    --scope openid,profile,email \
    --redirect-uris https://myapp.com/callback,https://myapp.com/auth

# Single-page application (SPA) - NEW STRUCTURE
capi uaa client create spa-app \
    --name "My SPA" \
    --authorized-grant-types authorization_code \
    --scope openid,profile \
    --redirect-uris https://spa.com/callback \
    --access-token-validity 3600 \
    --refresh-token-validity 7200

# Service client with authorities - NEW STRUCTURE
capi uaa client create service-client \
    --secret service-secret \
    --authorized-grant-types client_credentials \
    --authorities uaa.admin,scim.read,scim.write \
    --scope uaa.none

# Mobile application - NEW STRUCTURE
capi uaa client create mobile-app \
    --name "Mobile App" \
    --authorized-grant-types password,refresh_token \
    --scope openid,profile \
    --auto-approve openid,profile

# Legacy commands (deprecated):
# capi uaa create-client my-app --secret app-secret-123 --authorized-grant-types client_credentials
```

### List OAuth Clients

```bash
# List all clients - NEW STRUCTURE
capi uaa client list

# Filter clients - NEW STRUCTURE
capi uaa client list --filter 'client_id sw "app"'

# List with pagination - NEW STRUCTURE
capi uaa client list --count 25

# List all clients - NEW STRUCTURE
capi uaa client list --all

# Legacy commands (deprecated):
# capi uaa list-clients
```

### Get Client Details

```bash
# Get client (secrets masked) - NEW STRUCTURE
capi uaa client get my-app

# Get client with secret visible - NEW STRUCTURE
capi uaa client get my-app --show-secret

# Get client in JSON format - NEW STRUCTURE
capi uaa client get my-app --output json --show-secret

# Legacy commands (deprecated):
# capi uaa get-client my-app --show-secret
```

### Update OAuth Clients

```bash
# Update client attributes - NEW STRUCTURE
capi uaa client update my-app \
    --name "Updated App Name" \
    --scope "openid,profile,email,custom.scope" \
    --access-token-validity 7200

# Update grant types - NEW STRUCTURE
capi uaa client update web-app \
    --authorized-grant-types authorization_code,refresh_token,password

# Update redirect URIs - NEW STRUCTURE
capi uaa client update spa-app \
    --redirect-uris https://newdomain.com/callback,https://spa.com/auth

# Update authorities - NEW STRUCTURE
capi uaa client update service-client \
    --authorities uaa.admin,scim.read,scim.write,custom.authority

# Set auto-approve scopes - NEW STRUCTURE
capi uaa client update mobile-app \
    --auto-approve openid,profile

# Legacy commands (deprecated):
# capi uaa update-client my-app --name "Updated App Name" --scope "openid,profile"
```

### Client Secret Management

```bash
# Set new client secret - NEW STRUCTURE
capi uaa client set-secret my-app --secret new-secret-456

# Set secret with interactive input (more secure) - NEW STRUCTURE
capi uaa client set-secret my-app
# Prompts: Enter new secret: [hidden input]
#          Confirm secret: [hidden input]

# Legacy commands (deprecated):
# capi uaa set-client-secret my-app --secret new-secret-456
```

### Delete OAuth Clients

```bash
# Delete client with confirmation - NEW STRUCTURE
capi uaa client delete my-app

# Force delete without confirmation - NEW STRUCTURE
capi uaa client delete my-app --force

# Legacy commands (deprecated):
# capi uaa delete-client my-app --force
```

## Batch Operations

> **Note**: Batch operations are now organized under the `capi uaa batch` sub-command structure for better organization.

### Batch Import

```bash
# Import users from file - NEW STRUCTURE
capi uaa batch import --file users.json

# Import with validation - NEW STRUCTURE
capi uaa batch import --file users.csv --validate --dry-run

# Legacy commands (deprecated):
# capi uaa batch-import --file users.json
```

### Performance Testing

```bash
# Run performance tests - NEW STRUCTURE
capi uaa batch performance --users 1000 --concurrent 10

# Performance with specific operations - NEW STRUCTURE
capi uaa batch performance --operation create-users --count 500

# Legacy commands (deprecated):
# capi uaa performance --users 1000 --concurrent 10
```

### Cache Management

```bash
# Clear UAA cache - NEW STRUCTURE
capi uaa batch cache --clear

# Cache statistics - NEW STRUCTURE
capi uaa batch cache --stats

# Legacy commands (deprecated):
# capi uaa cache --clear
```

## Integration Commands

> **Note**: Integration utilities are now organized under the `capi uaa integration` sub-command structure.

### Compatibility Testing

```bash
# Check UAA compatibility - NEW STRUCTURE
capi uaa integration compatibility --version 4.30.0

# Test compatibility features - NEW STRUCTURE
capi uaa integration compatibility --test-features

# Legacy commands (deprecated):
# capi uaa compatibility --version 4.30.0
```

### Cloud Foundry Integration

```bash
# Test CF integration - NEW STRUCTURE
capi uaa integration cf --check-endpoints

# Validate CF permissions - NEW STRUCTURE
capi uaa integration cf --validate-permissions

# Legacy commands (deprecated):
# capi uaa cf-integration --check-endpoints
```

## Utility Commands

### Current User Information

```bash
# Get current user claims/info
capi uaa userinfo

# Get userinfo in JSON format
capi uaa userinfo --output json

# Get userinfo in YAML format
capi uaa userinfo --output yaml
```

### Direct UAA API Access

The `curl` command provides direct access to any UAA API endpoint:

```bash
# GET request to users endpoint
capi uaa curl /Users

# GET with custom headers
capi uaa curl /info --header "Accept: application/json"

# POST request with data
capi uaa curl /Users \
    --method POST \
    --header "Content-Type: application/json" \
    --data '{"userName":"test.user","emails":[{"value":"test@example.com"}]}'

# PUT request
capi uaa curl /Users/12345678-1234-1234-1234-123456789abc \
    --method PUT \
    --header "Content-Type: application/json" \
    --data '{"active":false}'

# DELETE request
capi uaa curl /Users/12345678-1234-1234-1234-123456789abc \
    --method DELETE

# Save response to file
capi uaa curl /Users --output users.json

# Multiple headers
capi uaa curl /oauth/token \
    --method POST \
    --header "Content-Type: application/x-www-form-urlencoded" \
    --header "Accept: application/json" \
    --data "grant_type=client_credentials&client_id=admin&client_secret=secret"
```

## Output Formats

All commands support multiple output formats:

### Table Format (Default)

```bash
capi uaa list-users
# Outputs human-readable table format
```

### JSON Format

```bash
capi uaa list-users --output json
# Outputs structured JSON for parsing

# Use with jq for filtering
capi uaa list-users --output json | jq '.resources[] | select(.active == true) | .userName'
```

### YAML Format

```bash
capi uaa list-users --output yaml
# Outputs YAML format for configuration files
```

### Examples

```bash
# Get all active users as JSON
capi uaa list-users --filter 'active eq true' --output json

# Export group information as YAML
capi uaa get-group developers --output yaml > developers-group.yml

# Get user emails for scripting
capi uaa list-users --output json | jq -r '.resources[].emails[0].value'

# Count users by origin
capi uaa list-users --output json | jq '.resources | group_by(.origin) | .[] | {origin: .[0].origin, count: length}'
```

## Security Considerations

### Credential Management

1. **Never log secrets**: The CLI automatically masks secrets in output and logs
2. **Use environment variables**: Set credentials in environment variables instead of command line
3. **Secure storage**: Store configuration in `~/.capi/config.yml` with restricted permissions

```bash
# Set environment variables
export UAA_CLIENT_ID=admin
export UAA_CLIENT_SECRET=admin-secret

# Use in commands
capi uaa get-client-credentials-token
```

### Token Security

```bash
# Tokens are stored securely in config file
# Check token storage location
capi config show | grep token

# Clear stored tokens
capi config unset uaa_token
capi config unset uaa_refresh_token
```

### HTTPS and SSL

```bash
# Always use HTTPS for production
capi uaa target https://uaa.production.com

# For development only (not recommended for production)
capi uaa target https://uaa.dev.com --skip-ssl-validation
```

### Audit and Logging

```bash
# Enable verbose logging for audit trails
capi --verbose users create-user audit.user

# Log operations to file
capi uaa list-users --output json > audit-users-$(date +%Y%m%d).json
```

## Troubleshooting

### Common Issues

#### Authentication Errors

```bash
# Check current authentication status
capi uaa context

# Verify UAA endpoint connectivity
capi uaa info

# Test with basic client credentials
capi uaa get-client-credentials-token --client-id admin --client-secret admin-secret
```

#### Connection Issues

```bash
# Test UAA endpoint connectivity
curl -k https://uaa.your-domain.com/info

# Check for SSL certificate issues
capi uaa target https://uaa.your-domain.com --skip-ssl-validation

# Verify network access and firewall settings
```

#### Permission Errors

```bash
# Check current user authorities
capi uaa userinfo

# Verify client authorities
capi uaa get-client admin --show-secret

# Test with admin client
capi uaa get-client-credentials-token --client-id admin --client-secret admin-secret
```

### Error Messages

#### "Not authenticated"
- Run authentication command first
- Check if tokens have expired
- Verify client credentials

#### "Insufficient scope"
- Client needs additional authorities/scopes
- Use admin client for administrative operations
- Check user permissions

#### "Resource not found"
- Verify resource exists
- Check spelling of usernames/IDs
- Ensure you have read permissions

### Debug Mode

```bash
# Enable verbose output for debugging
capi --verbose users list-users

# Check configuration
capi config show

# Test basic connectivity
capi uaa info --verbose
```

### Getting Help

```bash
# Get help for any command or sub-command group
capi uaa --help
capi uaa user --help
capi uaa user create --help
capi uaa group --help
capi uaa client --help
capi uaa token --help

# Get command examples - NEW STRUCTURE
capi uaa user create --help | grep -A 10 "Examples:"

# Legacy command help (deprecated)
# capi uaa create-user --help
# capi uaa list-users --help
```

## Advanced Examples

### Bulk User Operations

```bash
#!/bin/bash
# Bulk create users from CSV file - UPDATED FOR NEW STRUCTURE

while IFS=, read -r username email firstname lastname; do
    echo "Creating user: $username"
    capi uaa user create "$username" \
        --email "$email" \
        --given-name "$firstname" \
        --family-name "$lastname" \
        --password "TempPass123!" \
        --force-password-change
done < users.csv

# Alternative: Use batch import for large datasets
# capi uaa batch import --file users.csv
```

### Group Synchronization

```bash
#!/bin/bash
# Sync LDAP groups to UAA - UPDATED FOR NEW STRUCTURE

LDAP_GROUPS=("developers" "admins" "qa-team")
LDAP_BASE="OU=Groups,DC=company,DC=com"

for group in "${LDAP_GROUPS[@]}"; do
    echo "Mapping group: $group"
    capi uaa group map \
        --group "$group" \
        --external-group "CN=$group,$LDAP_BASE" \
        --origin ldap
done
```

### OAuth Client Audit

```bash
#!/bin/bash
# Audit OAuth clients - UPDATED FOR NEW STRUCTURE

echo "OAuth Client Security Audit"
echo "==========================="

capi uaa client list --output json | jq -r '.resources[] | 
{
    client_id: .client_id,
    grant_types: .authorized_grant_types,
    scopes: .scope,
    authorities: .authorities,
    auto_approve: .autoapprove
} | 
"Client: \(.client_id)
  Grant Types: \(.grant_types | join(", "))
  Scopes: \(.scopes | join(", "))
  Authorities: \(.authorities | join(", "))
  Auto Approve: \(.auto_approve | join(", "))
"'
```

This comprehensive guide covers all aspects of UAA user management with the `capi uaa` commands. For additional help, use the `--help` flag with any command or refer to the UAA documentation for more advanced SCIM filtering and OAuth2 concepts.