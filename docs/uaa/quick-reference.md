# UAA Commands Quick Reference

This is a quick reference guide for the `capi uaa` UAA commands. For comprehensive documentation, see [uaa-commands.md](./uaa-commands.md).

## Quick Start

```bash
# Set UAA target and authenticate
capi uaa target https://uaa.your-domain.com
capi uaa get-client-credentials-token --client-id admin --client-secret admin-secret

# Check authentication status
capi uaa context
```

## Context Management

| Command | Description | Example |
|---------|-------------|---------|
| `capi uaa target <url>` | Set UAA endpoint | `capi uaa target https://uaa.cf.com` |
| `capi uaa context` | Show current context | `capi uaa context --output json` |
| `capi uaa info` | Get UAA server info | `capi uaa info` |
| `capi uaa version` | Get UAA version | `capi uaa version` |

## Authentication & Tokens

| Command | Description | Example |
|---------|-------------|---------|
| `capi uaa get-client-credentials-token` | Client credentials grant | `--client-id admin --client-secret secret` |
| `capi uaa get-password-token` | Password grant | `--username user --password pass --client-id cf` |
| `capi uaa get-authcode-token` | Authorization code grant | `--client-id web-app --code AUTH_CODE` |
| `capi uaa refresh-token` | Refresh access token | `capi uaa refresh-token` |
| `capi uaa get-token-key` | Get JWT signing key | `capi uaa get-token-key` |
| `capi uaa get-token-keys` | Get all JWT keys | `capi uaa get-token-keys` |

## User Management

### Basic Operations

| Command | Description | Example |
|---------|-------------|---------|
| `capi uaa create-user <name>` | Create user | `--email user@example.com --password SecurePass123!` |
| `capi uaa get-user <name>` | Get user details | `capi uaa get-user john.doe` |
| `capi uaa list-users` | List users | `--filter 'active eq true'` |
| `capi uaa update-user <name>` | Update user | `--email new@example.com --phone +1-555-9999` |
| `capi uaa delete-user <name>` | Delete user | `capi uaa delete-user john.doe --force` |

### User Status

| Command | Description | Example |
|---------|-------------|---------|
| `capi uaa activate-user <name>` | Activate user | `capi uaa activate-user john.doe` |
| `capi uaa deactivate-user <name>` | Deactivate user | `capi uaa deactivate-user john.doe` |

### Advanced User Queries

```bash
# Filter examples
capi uaa list-users --filter 'email co "example.com"'
capi uaa list-users --filter 'active eq true and origin eq "uaa"'
capi uaa list-users --filter 'meta.created gt "2023-01-01T00:00:00.000Z"'

# Sorting and pagination
capi uaa list-users --sort-by userName --count 50
capi uaa list-users --all  # Get all users (auto-pagination)

# Attribute selection
capi uaa list-users --attributes userName,email,active
```

## Group Management

### Basic Operations

| Command | Description | Example |
|---------|-------------|---------|
| `capi uaa create-group <name>` | Create group | `--description "Development team"` |
| `capi uaa get-group <name>` | Get group details | `capi uaa get-group developers` |
| `capi uaa list-groups` | List groups | `--filter 'displayName sw "admin"'` |
| `capi uaa delete-group <name>` | Delete group | `capi uaa delete-group old-team --force` |

### Membership Management

| Command | Description | Example |
|---------|-------------|---------|
| `capi uaa add-member <group> <user>` | Add user to group | `capi uaa add-member developers john.doe` |
| `capi uaa remove-member <group> <user>` | Remove user from group | `capi uaa remove-member developers john.doe` |

### External Group Mapping

| Command | Description | Example |
|---------|-------------|---------|
| `capi uaa map-group` | Map external group | `--group devs --external-group "CN=Developers,DC=company" --origin ldap` |
| `capi uaa unmap-group` | Remove mapping | `--group devs --external-group "CN=Developers,DC=company" --origin ldap` |
| `capi uaa list-group-mappings` | List mappings | `--origin ldap` |

## OAuth Client Management

### Basic Operations

| Command | Description | Example |
|---------|-------------|---------|
| `capi uaa create-client <id>` | Create client | `--secret app-secret --authorized-grant-types client_credentials` |
| `capi uaa get-client <id>` | Get client details | `--show-secret` to reveal secret |
| `capi uaa list-clients` | List clients | `capi uaa list-clients` |
| `capi uaa update-client <id>` | Update client | `--name "New Name" --scope "new.scope"` |
| `capi uaa set-client-secret <id>` | Update secret | `--secret new-secret` |
| `capi uaa delete-client <id>` | Delete client | `capi uaa delete-client old-app --force` |

### Client Types Examples

```bash
# Web application
capi uaa create-client web-app \
  --secret web-secret \
  --authorized-grant-types authorization_code,refresh_token \
  --scope openid,profile,email \
  --redirect-uris https://app.com/callback

# Single Page Application (SPA)
capi uaa create-client spa \
  --authorized-grant-types authorization_code \
  --scope openid,profile \
  --redirect-uris https://spa.com/callback

# Service client
capi uaa create-client service \
  --secret service-secret \
  --authorized-grant-types client_credentials \
  --authorities uaa.resource,custom.read
```

## Utility Commands

| Command | Description | Example |
|---------|-------------|---------|
| `capi uaa userinfo` | Get current user info | `capi uaa userinfo --output json` |
| `capi uaa curl <path>` | Direct API access | `capi uaa curl /Users --method GET` |

### curl Command Examples

```bash
# GET request
capi uaa curl /info

# POST with data
capi uaa curl /Users \
  --method POST \
  --header "Content-Type: application/json" \
  --data '{"userName":"test","emails":[{"value":"test@example.com"}]}'

# Save to file
capi uaa curl /Users --output users.json
```

## Output Formats

All commands support multiple output formats:

```bash
# Table format (default)
capi uaa list-users

# JSON format (for scripting)
capi uaa list-users --output json

# YAML format
capi uaa list-users --output yaml
```

## Common SCIM Filters

### User Filters

```bash
# Active users only
--filter 'active eq true'

# Email domain
--filter 'email co "example.com"'

# Username pattern
--filter 'userName sw "admin"'

# External users
--filter 'origin ne "uaa"'

# Recently created
--filter 'meta.created gt "2023-01-01T00:00:00.000Z"'

# Complex filter
--filter 'active eq true and email co "company.com" and origin eq "uaa"'
```

### Group Filters

```bash
# Group name pattern
--filter 'displayName sw "admin"'

# Groups with description
--filter 'description pr'  # pr = present (not null)
```

## Common Use Cases

### Bulk User Creation

```bash
# From CSV file
while IFS=, read -r username email firstname lastname; do
  capi uaa create-user "$username" \
    --email "$email" \
    --given-name "$firstname" \
    --family-name "$lastname" \
    --password "TempPass123!"
done < users.csv
```

### User Audit

```bash
# Export all users
capi uaa list-users --all --output json > all-users.json

# Find inactive users
capi uaa list-users --filter 'active eq false' --output json

# Users by origin
capi uaa list-users --filter 'origin eq "ldap"' --attributes userName,email,origin
```

### Group Membership Report

```bash
# Get group members
capi uaa get-group developers --output json | jq '.members[] | .value'

# User's groups (requires iteration)
for group in $(capi uaa list-groups --output json | jq -r '.resources[].displayName'); do
  if capi uaa get-group "$group" --output json | jq -e '.members[]? | select(.value == "john.doe")' >/dev/null; then
    echo "$group"
  fi
done
```

### Client Security Audit

```bash
# Clients with admin authorities
capi uaa list-clients --output json | jq '.resources[] | select(.authorities[]? == "uaa.admin")'

# Clients with password grant
capi uaa list-clients --output json | jq '.resources[] | select(.authorized_grant_types[]? == "password")'
```

## Environment Variables

Set these for easier authentication:

```bash
export UAA_ENDPOINT="https://uaa.your-domain.com"
export UAA_CLIENT_ID="admin"
export UAA_CLIENT_SECRET="admin-secret"
export UAA_USERNAME="admin"
export UAA_PASSWORD="admin-password"
```

## Error Handling

### Common Issues

- **"Not authenticated"**: Run authentication command first
- **"Insufficient scope"**: Client needs additional authorities
- **"Resource not found"**: Check spelling and existence
- **"SSL certificate error"**: Use `--skip-ssl-validation` for dev environments

### Debug Mode

```bash
# Enable verbose output
capi --verbose users list-users

# Check configuration
capi config show
```

## Help

Get help for any command:

```bash
capi uaa --help
capi uaa create-user --help
capi uaa list-users --help
```

## Examples

See the [examples/uaa/](../examples/uaa/) directory for complete working examples:

- `auth-setup.sh` - Authentication setup
- `user-lifecycle.sh` - Complete user management workflow
- `group-management.sh` - Group management demonstration

This quick reference covers the most commonly used UAA commands and patterns. For detailed documentation with advanced examples, see the complete [UAA Commands Guide](./uaa-commands.md).