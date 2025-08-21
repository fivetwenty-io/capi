# UAA User Management Commands

The `capi uaa` commands provide comprehensive UAA (User Account and Authentication) user management functionality. These commands allow you to manage users, groups, OAuth clients, and authentication tokens in your Cloud Foundry UAA environment.

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
# Option 1: Client credentials (machine-to-machine)
capi uaa get-client-credentials-token --client-id admin --client-secret admin-secret

# Option 2: Username/password (user authentication)
capi uaa get-password-token --username admin --password admin-pass --client-id cf

# Option 3: Authorization code flow (interactive)
capi uaa get-authcode-token --client-id my-client --redirect-uri http://localhost:8080/callback
```

### 3. Basic Operations

```bash
# List users
capi uaa list-users

# Create a user
capi uaa create-user john.doe --email john.doe@example.com --password SecurePass123!

# Get user info
capi uaa get-user john.doe

# Create a group
capi uaa create-group developers --description "Development team"

# Add user to group
capi uaa add-member developers john.doe
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
# Basic client credentials
capi uaa get-client-credentials-token --client-id my-client --client-secret my-secret

# Specify token format
capi uaa get-client-credentials-token \
    --client-id my-client \
    --client-secret my-secret \
    --token-format jwt

# Request specific scopes
capi uaa get-client-credentials-token \
    --client-id my-client \
    --client-secret my-secret \
    --scope "uaa.admin,scim.read"
```

### Password Grant

For user authentication:

```bash
# Username/password authentication
capi uaa get-password-token \
    --username admin \
    --password admin-pass \
    --client-id cf

# Specify client credentials
capi uaa get-password-token \
    --username user@example.com \
    --password user-pass \
    --client-id my-client \
    --client-secret my-secret

# Request specific scopes
capi uaa get-password-token \
    --username admin \
    --password admin-pass \
    --client-id cf \
    --scope "scim.read,scim.write"
```

### Authorization Code Grant

For web applications:

```bash
# Start authorization code flow
capi uaa get-authcode-token \
    --client-id my-web-app \
    --client-secret app-secret \
    --redirect-uri https://myapp.com/callback \
    --code AUTHORIZATION_CODE_FROM_CALLBACK

# With PKCE (Proof Key for Code Exchange)
capi uaa get-authcode-token \
    --client-id my-spa \
    --redirect-uri https://spa.com/callback \
    --code AUTHORIZATION_CODE \
    --code-verifier CODE_VERIFIER
```

### Token Refresh

```bash
# Refresh current token
capi uaa refresh-token

# Refresh with specific refresh token
capi uaa refresh-token --refresh-token REFRESH_TOKEN_VALUE

# Refresh with client credentials
capi uaa refresh-token \
    --refresh-token REFRESH_TOKEN \
    --client-id my-client \
    --client-secret my-secret
```

### Token Keys

```bash
# Get current JWT signing key
capi uaa get-token-key

# Get all JWT signing keys (including rotated keys)
capi uaa get-token-keys

# Get keys in JSON format
capi uaa get-token-keys --output json
```

## User Management

### Create Users

```bash
# Basic user creation
capi uaa create-user john.doe \
    --email john.doe@example.com \
    --password SecurePass123!

# Full user creation with all attributes
capi uaa create-user jane.smith \
    --email jane.smith@example.com \
    --password AnotherPass456! \
    --given-name Jane \
    --family-name Smith \
    --phone-number "+1-555-0123" \
    --origin uaa \
    --active \
    --verified

# Create user with external identity provider
capi uaa create-user external.user \
    --email external@ldap.company.com \
    --origin ldap \
    --external-id "cn=external,ou=users,dc=company,dc=com"
```

### List Users

```bash
# List all users
capi uaa list-users

# List with pagination
capi uaa list-users --count 50 --start-index 100

# List all users (automatic pagination)
capi uaa list-users --all

# Filter users with SCIM expressions
capi uaa list-users --filter 'active eq true'
capi uaa list-users --filter 'email co "example.com"'
capi uaa list-users --filter 'userName sw "admin"'
capi uaa list-users --filter 'origin eq "ldap"'
capi uaa list-users --filter 'meta.created gt "2023-01-01T00:00:00.000Z"'

# Sort users
capi uaa list-users --sort-by userName --sort-order ascending
capi uaa list-users --sort-by meta.created --sort-order descending

# Select specific attributes
capi uaa list-users --attributes userName,email,active

# Combined filtering and sorting
capi uaa list-users \
    --filter 'active eq true and email co "company.com"' \
    --sort-by userName \
    --count 25 \
    --attributes userName,email,name.givenName,name.familyName
```

### Get User Details

```bash
# Get user by username
capi uaa get-user john.doe

# Get user by UUID
capi uaa get-user 12345678-1234-1234-1234-123456789abc

# Get user with specific attributes
capi uaa get-user jane.smith --attributes userName,email,groups

# Get user in JSON format
capi uaa get-user john.doe --output json
```

### Update Users

```bash
# Update user attributes
capi uaa update-user john.doe \
    --email new.email@example.com \
    --given-name John \
    --family-name Doe-Updated \
    --phone-number "+1-555-9999"

# Activate/deactivate user
capi uaa update-user john.doe --active
capi uaa update-user john.doe --no-active

# Verify/unverify user email
capi uaa update-user john.doe --verified
capi uaa update-user john.doe --no-verified

# Update password (admin operation)
capi uaa update-user john.doe --password NewSecurePass789!
```

### User Status Management

```bash
# Activate user account
capi uaa activate-user john.doe

# Deactivate user account
capi uaa deactivate-user john.doe

# Check activation status
capi uaa get-user john.doe --attributes active
```

### Delete Users

```bash
# Delete user with confirmation
capi uaa delete-user john.doe

# Force delete without confirmation
capi uaa delete-user john.doe --force

# Delete user by UUID
capi uaa delete-user 12345678-1234-1234-1234-123456789abc --force
```

## Group Management

### Create Groups

```bash
# Basic group creation
capi uaa create-group developers

# Group with description
capi uaa create-group admins --description "System administrators"

# Group with initial members
capi uaa create-group qa-team \
    --description "Quality assurance team" \
    --members john.doe,jane.smith
```

### List Groups

```bash
# List all groups
capi uaa list-groups

# Filter groups
capi uaa list-groups --filter 'displayName sw "admin"'
capi uaa list-groups --filter 'meta.created gt "2023-01-01T00:00:00.000Z"'

# List with pagination
capi uaa list-groups --count 20 --start-index 0

# Get all groups
capi uaa list-groups --all
```

### Get Group Details

```bash
# Get group by name
capi uaa get-group developers

# Get group by UUID
capi uaa get-group 87654321-4321-4321-4321-210987654321

# Get group in JSON format
capi uaa get-group developers --output json
```

### Group Membership

```bash
# Add user to group (by username)
capi uaa add-member developers john.doe

# Add user to group (by UUID)
capi uaa add-member developers 12345678-1234-1234-1234-123456789abc

# Add member with specific origin
capi uaa add-member developers external.user --origin ldap

# Add member with type
capi uaa add-member developers service.account --type client

# Remove user from group
capi uaa remove-member developers john.doe

# Remove member by UUID
capi uaa remove-member developers 12345678-1234-1234-1234-123456789abc
```

### External Group Mapping

```bash
# Map external group to UAA group
capi uaa map-group \
    --group developers \
    --external-group "CN=Developers,OU=Groups,DC=company,DC=com" \
    --origin ldap

# Map SAML group
capi uaa map-group \
    --group admins \
    --external-group "admin-users" \
    --origin saml

# List group mappings
capi uaa list-group-mappings

# List mappings for specific origin
capi uaa list-group-mappings --origin ldap

# Remove group mapping
capi uaa unmap-group \
    --group developers \
    --external-group "CN=Developers,OU=Groups,DC=company,DC=com" \
    --origin ldap
```

## OAuth Client Management

### Create OAuth Clients

```bash
# Basic client creation
capi uaa create-client my-app \
    --secret app-secret-123 \
    --authorized-grant-types client_credentials

# Web application client
capi uaa create-client web-app \
    --secret web-secret \
    --name "My Web Application" \
    --authorized-grant-types authorization_code,refresh_token \
    --scope openid,profile,email \
    --redirect-uris https://myapp.com/callback,https://myapp.com/auth

# Single-page application (SPA)
capi uaa create-client spa-app \
    --name "My SPA" \
    --authorized-grant-types authorization_code \
    --scope openid,profile \
    --redirect-uris https://spa.com/callback \
    --access-token-validity 3600 \
    --refresh-token-validity 7200

# Service client with authorities
capi uaa create-client service-client \
    --secret service-secret \
    --authorized-grant-types client_credentials \
    --authorities uaa.admin,scim.read,scim.write \
    --scope uaa.none

# Mobile application
capi uaa create-client mobile-app \
    --name "Mobile App" \
    --authorized-grant-types password,refresh_token \
    --scope openid,profile \
    --auto-approve openid,profile
```

### List OAuth Clients

```bash
# List all clients
capi uaa list-clients

# Filter clients
capi uaa list-clients --filter 'client_id sw "app"'

# List with pagination
capi uaa list-clients --count 25

# List all clients
capi uaa list-clients --all
```

### Get Client Details

```bash
# Get client (secrets masked)
capi uaa get-client my-app

# Get client with secret visible
capi uaa get-client my-app --show-secret

# Get client in JSON format
capi uaa get-client my-app --output json --show-secret
```

### Update OAuth Clients

```bash
# Update client attributes
capi uaa update-client my-app \
    --name "Updated App Name" \
    --scope "openid,profile,email,custom.scope" \
    --access-token-validity 7200

# Update grant types
capi uaa update-client web-app \
    --authorized-grant-types authorization_code,refresh_token,password

# Update redirect URIs
capi uaa update-client spa-app \
    --redirect-uris https://newdomain.com/callback,https://spa.com/auth

# Update authorities
capi uaa update-client service-client \
    --authorities uaa.admin,scim.read,scim.write,custom.authority

# Set auto-approve scopes
capi uaa update-client mobile-app \
    --auto-approve openid,profile
```

### Client Secret Management

```bash
# Set new client secret
capi uaa set-client-secret my-app --secret new-secret-456

# Set secret with interactive input (more secure)
capi uaa set-client-secret my-app
# Prompts: Enter new secret: [hidden input]
#          Confirm secret: [hidden input]
```

### Delete OAuth Clients

```bash
# Delete client with confirmation
capi uaa delete-client my-app

# Force delete without confirmation
capi uaa delete-client my-app --force
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
# Get help for any command
capi uaa --help
capi uaa create-user --help
capi uaa list-users --help

# Get command examples
capi uaa create-user --help | grep -A 10 "Examples:"
```

## Advanced Examples

### Bulk User Operations

```bash
#!/bin/bash
# Bulk create users from CSV file

while IFS=, read -r username email firstname lastname; do
    echo "Creating user: $username"
    capi uaa create-user "$username" \
        --email "$email" \
        --given-name "$firstname" \
        --family-name "$lastname" \
        --password "TempPass123!" \
        --force-password-change
done < users.csv
```

### Group Synchronization

```bash
#!/bin/bash
# Sync LDAP groups to UAA

LDAP_GROUPS=("developers" "admins" "qa-team")
LDAP_BASE="OU=Groups,DC=company,DC=com"

for group in "${LDAP_GROUPS[@]}"; do
    echo "Mapping group: $group"
    capi uaa map-group \
        --group "$group" \
        --external-group "CN=$group,$LDAP_BASE" \
        --origin ldap
done
```

### OAuth Client Audit

```bash
#!/bin/bash
# Audit OAuth clients

echo "OAuth Client Security Audit"
echo "==========================="

capi uaa list-clients --output json | jq -r '.resources[] | 
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