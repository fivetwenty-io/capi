# UAA Commands Quick Reference

This is a quick reference guide for the `capi uaa` UAA commands. For comprehensive documentation, see [commands.md](./commands.md).

## New Command Structure (v2.0)

**Commands are now organized hierarchically for better usability:**

- **`capi uaa user`** - User management (create, get, list, update, activate, deactivate, delete)
- **`capi uaa group`** - Group management (create, get, list, add-member, remove-member, map, unmap)
- **`capi uaa client`** - OAuth client management (create, get, list, update, set-secret, delete) 
- **`capi uaa token`** - Token operations (get-authcode, get-client-credentials, get-password, refresh, get-keys)
- **`capi uaa batch`** - Batch operations (import, performance, cache)
- **`capi uaa integration`** - Integration utilities (compatibility, cf)

**Legacy hyphenated commands are deprecated but still functional.**

## Quick Start

```bash
# Set UAA target and authenticate - NEW STRUCTURE
capi uaa target https://uaa.your-domain.com
capi uaa token get-client-credentials --client-id admin --client-secret admin-secret

# Check authentication status
capi uaa context

# Legacy commands (deprecated):
# capi uaa get-client-credentials-token --client-id admin --client-secret admin-secret
```

## Context Management

| Command | Description | Example |
|---------|-------------|---------|
| `capi uaa target <url>` | Set UAA endpoint | `capi uaa target https://uaa.cf.com` |
| `capi uaa context` | Show current context | `capi uaa context --output json` |
| `capi uaa info` | Get UAA server info | `capi uaa info` |
| `capi uaa version` | Get UAA version | `capi uaa version` |

## Authentication & Tokens (New Structure)

| New Command | Description | Example |
|-------------|-------------|---------|
| `capi uaa token get-client-credentials` | Client credentials grant | `--client-id admin --client-secret secret` |
| `capi uaa token get-password` | Password grant | `--username user --password pass --client-id cf` |
| `capi uaa token get-authcode` | Authorization code grant | `--client-id web-app --code AUTH_CODE` |
| `capi uaa token refresh` | Refresh access token | `capi uaa token refresh` |
| `capi uaa token get-key` | Get JWT signing key | `capi uaa token get-key` |
| `capi uaa token get-keys` | Get all JWT keys | `capi uaa token get-keys` |
| `capi uaa token get-implicit` | Implicit grant | `--client-id spa-app` |

### Legacy Token Commands (Deprecated)

| Legacy Command | New Command |
|----------------|-------------|
| `capi uaa get-client-credentials-token` | `capi uaa token get-client-credentials` |
| `capi uaa get-password-token` | `capi uaa token get-password` |
| `capi uaa get-authcode-token` | `capi uaa token get-authcode` |
| `capi uaa refresh-token` | `capi uaa token refresh` |
| `capi uaa get-token-key` | `capi uaa token get-key` |
| `capi uaa get-token-keys` | `capi uaa token get-keys` |

## User Management (New Structure)

### Basic Operations

| New Command | Description | Example |
|-------------|-------------|---------|
| `capi uaa user create <name>` | Create user | `--email user@example.com --password SecurePass123!` |
| `capi uaa user get <name>` | Get user details | `capi uaa user get john.doe` |
| `capi uaa user list` | List users | `--filter 'active eq true'` |
| `capi uaa user update <name>` | Update user | `--email new@example.com --phone +1-555-9999` |
| `capi uaa user delete <name>` | Delete user | `capi uaa user delete john.doe --force` |

### User Status

| New Command | Description | Example |
|-------------|-------------|---------|
| `capi uaa user activate <name>` | Activate user | `capi uaa user activate john.doe` |
| `capi uaa user deactivate <name>` | Deactivate user | `capi uaa user deactivate john.doe` |

### Legacy User Commands (Deprecated)

| Legacy Command | New Command |
|----------------|-------------|
| `capi uaa create-user <name>` | `capi uaa user create <name>` |
| `capi uaa get-user <name>` | `capi uaa user get <name>` |
| `capi uaa list-users` | `capi uaa user list` |
| `capi uaa update-user <name>` | `capi uaa user update <name>` |
| `capi uaa delete-user <name>` | `capi uaa user delete <name>` |
| `capi uaa activate-user <name>` | `capi uaa user activate <name>` |
| `capi uaa deactivate-user <name>` | `capi uaa user deactivate <name>` |

### Advanced User Queries

```bash
# Filter examples - NEW STRUCTURE
capi uaa user list --filter 'email co "example.com"'
capi uaa user list --filter 'active eq true and origin eq "uaa"'
capi uaa user list --filter 'meta.created gt "2023-01-01T00:00:00.000Z"'

# Sorting and pagination - NEW STRUCTURE
capi uaa user list --sort-by userName --count 50
capi uaa user list --all  # Get all users (auto-pagination)

# Attribute selection - NEW STRUCTURE
capi uaa user list --attributes userName,email,active

# Legacy examples (deprecated):
# capi uaa list-users --filter 'email co "example.com"'
# capi uaa list-users --sort-by userName --count 50
```

## Group Management (New Structure)

### Basic Operations

| New Command | Description | Example |
|-------------|-------------|---------|
| `capi uaa group create <name>` | Create group | `--description "Development team"` |
| `capi uaa group get <name>` | Get group details | `capi uaa group get developers` |
| `capi uaa group list` | List groups | `--filter 'displayName sw "admin"'` |
| `capi uaa group delete <name>` | Delete group | `capi uaa group delete old-team --force` |

### Membership Management

| New Command | Description | Example |
|-------------|-------------|---------|
| `capi uaa group add-member <group> <user>` | Add user to group | `capi uaa group add-member developers john.doe` |
| `capi uaa group remove-member <group> <user>` | Remove user from group | `capi uaa group remove-member developers john.doe` |

### External Group Mapping

| New Command | Description | Example |
|-------------|-------------|---------|
| `capi uaa group map` | Map external group | `--group devs --external-group "CN=Developers,DC=company" --origin ldap` |
| `capi uaa group unmap` | Remove mapping | `--group devs --external-group "CN=Developers,DC=company" --origin ldap` |
| `capi uaa group list-mappings` | List mappings | `--origin ldap` |

### Legacy Group Commands (Deprecated)

| Legacy Command | New Command |
|----------------|-------------|
| `capi uaa create-group <name>` | `capi uaa group create <name>` |
| `capi uaa get-group <name>` | `capi uaa group get <name>` |
| `capi uaa list-groups` | `capi uaa group list` |
| `capi uaa delete-group <name>` | `capi uaa group delete <name>` |
| `capi uaa add-member <group> <user>` | `capi uaa group add-member <group> <user>` |
| `capi uaa remove-member <group> <user>` | `capi uaa group remove-member <group> <user>` |
| `capi uaa map-group` | `capi uaa group map` |
| `capi uaa unmap-group` | `capi uaa group unmap` |
| `capi uaa list-group-mappings` | `capi uaa group list-mappings` |

## OAuth Client Management (New Structure)

### Basic Operations

| New Command | Description | Example |
|-------------|-------------|---------|
| `capi uaa client create <id>` | Create client | `--secret app-secret --authorized-grant-types client_credentials` |
| `capi uaa client get <id>` | Get client details | `--show-secret` to reveal secret |
| `capi uaa client list` | List clients | `capi uaa client list` |
| `capi uaa client update <id>` | Update client | `--name "New Name" --scope "new.scope"` |
| `capi uaa client set-secret <id>` | Update secret | `--secret new-secret` |
| `capi uaa client delete <id>` | Delete client | `capi uaa client delete old-app --force` |

### Legacy Client Commands (Deprecated)

| Legacy Command | New Command |
|----------------|-------------|
| `capi uaa create-client <id>` | `capi uaa client create <id>` |
| `capi uaa get-client <id>` | `capi uaa client get <id>` |
| `capi uaa list-clients` | `capi uaa client list` |
| `capi uaa update-client <id>` | `capi uaa client update <id>` |
| `capi uaa set-client-secret <id>` | `capi uaa client set-secret <id>` |
| `capi uaa delete-client <id>` | `capi uaa client delete <id>` |

### Client Types Examples

```bash
# Web application - NEW STRUCTURE
capi uaa client create web-app \
  --secret web-secret \
  --authorized-grant-types authorization_code,refresh_token \
  --scope openid,profile,email \
  --redirect-uris https://app.com/callback

# Single Page Application (SPA) - NEW STRUCTURE
capi uaa client create spa \
  --authorized-grant-types authorization_code \
  --scope openid,profile \
  --redirect-uris https://spa.com/callback

# Legacy examples (deprecated):
# capi uaa create-client web-app --secret web-secret --authorized-grant-types authorization_code,refresh_token

# Service client - NEW STRUCTURE
capi uaa client create service \
  --secret service-secret \
  --authorized-grant-types client_credentials \
  --authorities uaa.resource,custom.read

# Legacy example (deprecated):
# capi uaa create-client service --secret service-secret --authorized-grant-types client_credentials
```

## Batch Operations (New Structure)

| New Command | Description | Example |
|-------------|-------------|---------|  
| `capi uaa batch import` | Bulk import users | `--file users.json` |
| `capi uaa batch performance` | Performance testing | `--users 1000 --concurrent 10` |
| `capi uaa batch cache` | Cache management | `--clear` or `--stats` |

### Legacy Batch Commands (Deprecated)

| Legacy Command | New Command |
|----------------|-------------|
| `capi uaa batch-import` | `capi uaa batch import` |
| `capi uaa performance` | `capi uaa batch performance` |
| `capi uaa cache` | `capi uaa batch cache` |

## Integration Commands (New Structure)

| New Command | Description | Example |
|-------------|-------------|---------|  
| `capi uaa integration compatibility` | Check UAA compatibility | `--version 4.30.0` |
| `capi uaa integration cf` | CF integration tests | `--check-endpoints` |

### Legacy Integration Commands (Deprecated)

| Legacy Command | New Command |
|----------------|-------------|
| `capi uaa compatibility` | `capi uaa integration compatibility` |
| `capi uaa cf-integration` | `capi uaa integration cf` |

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
# Table format (default) - NEW STRUCTURE
capi uaa user list

# JSON format (for scripting) - NEW STRUCTURE
capi uaa user list --output json

# YAML format - NEW STRUCTURE
capi uaa user list --output yaml

# Legacy examples (deprecated):
# capi uaa list-users --output json
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
# From CSV file - NEW STRUCTURE
while IFS=, read -r username email firstname lastname; do
  capi uaa user create "$username" \
    --email "$email" \
    --given-name "$firstname" \
    --family-name "$lastname" \
    --password "TempPass123!"
done < users.csv

# Alternative: Use batch import - NEW STRUCTURE
capi uaa batch import --file users.csv

# Legacy approach (deprecated):
# while IFS=, read -r username email firstname lastname; do
#   capi uaa create-user "$username" --email "$email" --given-name "$firstname"
# done < users.csv
```

### User Audit

```bash
# Export all users - NEW STRUCTURE
capi uaa user list --all --output json > all-users.json

# Find inactive users - NEW STRUCTURE
capi uaa user list --filter 'active eq false' --output json

# Users by origin - NEW STRUCTURE
capi uaa user list --filter 'origin eq "ldap"' --attributes userName,email,origin

# Legacy examples (deprecated):
# capi uaa list-users --all --output json > all-users.json
# capi uaa list-users --filter 'active eq false' --output json
```

### Group Membership Report

```bash
# Get group members - NEW STRUCTURE
capi uaa group get developers --output json | jq '.members[] | .value'

# User's groups (requires iteration) - NEW STRUCTURE
for group in $(capi uaa group list --output json | jq -r '.resources[].displayName'); do
  if capi uaa group get "$group" --output json | jq -e '.members[]? | select(.value == "john.doe")' >/dev/null; then
    echo "$group"
  fi
done

# Legacy examples (deprecated):
# capi uaa get-group developers --output json | jq '.members[] | .value'
# for group in $(capi uaa list-groups --output json | jq -r '.resources[].displayName'); do
#   capi uaa get-group "$group" --output json
# done
```

### Client Security Audit

```bash
# Clients with admin authorities - NEW STRUCTURE
capi uaa client list --output json | jq '.resources[] | select(.authorities[]? == "uaa.admin")'

# Clients with password grant - NEW STRUCTURE
capi uaa client list --output json | jq '.resources[] | select(.authorized_grant_types[]? == "password")'

# Legacy examples (deprecated):
# capi uaa list-clients --output json | jq '.resources[] | select(.authorities[]? == "uaa.admin")'
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
# Enable verbose output - NEW STRUCTURE
capi --verbose uaa user list

# Check configuration
capi config show

# Legacy example (deprecated):
# capi --verbose uaa list-users
```

## Help

Get help for any command:

```bash
# Get help for main command or sub-command groups
capi uaa --help
capi uaa user --help
capi uaa group --help
capi uaa client --help
capi uaa token --help
capi uaa batch --help
capi uaa integration --help

# Get help for specific commands - NEW STRUCTURE
capi uaa user create --help
capi uaa user list --help
capi uaa token get-client-credentials --help

# Legacy command help (deprecated):
# capi uaa create-user --help
# capi uaa list-users --help
```

## Examples

See the [examples/uaa/](../examples/uaa/) directory for complete working examples:

- `auth-setup.sh` - Authentication setup
- `user-lifecycle.sh` - Complete user management workflow
- `group-management.sh` - Group management demonstration

## Migration Guide

**All legacy hyphenated commands are deprecated but remain functional for backward compatibility.** They are hidden from help output but will continue to work.

**Recommended Migration Path:**
1. Update scripts and automation to use new hierarchical commands
2. Train users on new command structure  
3. Legacy commands will be removed in a future major version

**Quick Command Mapping:**
- `capi uaa create-user` → `capi uaa user create`
- `capi uaa list-users` → `capi uaa user list`
- `capi uaa get-client-credentials-token` → `capi uaa token get-client-credentials`
- `capi uaa create-group` → `capi uaa group create`
- `capi uaa add-member` → `capi uaa group add-member`
- `capi uaa create-client` → `capi uaa client create`

This quick reference covers the most commonly used UAA commands and patterns. For detailed documentation with advanced examples, see the complete [UAA Commands Guide](./commands.md).