# UAA User Management Examples

This directory contains practical examples for using the `capi uaa` commands to manage UAA (User Account and Authentication) resources.

## New Command Structure (v2.0)

**Important: These examples use the new hierarchical command structure introduced in v2.0.**

- **`capi uaa user`** - User management operations
- **`capi uaa group`** - Group management operations  
- **`capi uaa client`** - OAuth client management
- **`capi uaa token`** - Token operations
- **`capi uaa batch`** - Batch operations and utilities
- **`capi uaa integration`** - Integration and compatibility utilities

**Legacy hyphenated commands are deprecated but still functional.** Examples show both new and legacy formats where applicable.

## Prerequisites

1. Access to a UAA server (typically part of a Cloud Foundry deployment)
2. Admin credentials or appropriate client credentials
3. The `capi` CLI tool installed and configured

## Examples Overview

- [Authentication](#authentication-examples)
- [User Management](#user-management-examples)
- [Group Management](#group-management-examples)
- [OAuth Client Management](#oauth-client-management-examples)
- [Bulk Operations](#bulk-operations-examples)
- [Integration Scripts](#integration-scripts)

## Authentication Examples

### Basic Authentication Setup

```bash
#!/bin/bash
# auth-setup.sh - Basic UAA authentication setup

# Set UAA endpoint
capi uaa target https://uaa.your-cf-domain.com

# Option 1: Admin client credentials - NEW STRUCTURE
capi uaa token get-client-credentials \
    --client-id admin \
    --client-secret admin-secret

# Option 2: User password authentication - NEW STRUCTURE
capi uaa token get-password \
    --username admin \
    --password admin-password \
    --client-id cf

# Legacy commands (deprecated):
# capi uaa get-client-credentials-token --client-id admin --client-secret admin-secret
# capi uaa get-password-token --username admin --password admin-password --client-id cf

# Verify authentication
capi uaa context
echo "Authentication complete!"
```

### Environment-Based Authentication

```bash
#!/bin/bash
# env-auth.sh - Use environment variables for authentication

# Set required environment variables
export UAA_ENDPOINT="https://uaa.your-cf-domain.com"
export UAA_CLIENT_ID="admin"
export UAA_CLIENT_SECRET="admin-secret"

# Authenticate using environment variables - NEW STRUCTURE
capi uaa target "$UAA_ENDPOINT"
capi uaa token get-client-credentials \
    --client-id "$UAA_CLIENT_ID" \
    --client-secret "$UAA_CLIENT_SECRET"

# Legacy command (deprecated):
# capi uaa get-client-credentials-token --client-id "$UAA_CLIENT_ID" --client-secret "$UAA_CLIENT_SECRET"

echo "Authenticated with client: $UAA_CLIENT_ID"
```

## User Management Examples

### Basic User Lifecycle

```bash
#!/bin/bash
# user-lifecycle.sh - Complete user management workflow

set -e

USERNAME="john.doe"
EMAIL="john.doe@example.com"
TEMP_PASSWORD="TempPass123!"

echo "Creating user: $USERNAME"
# NEW STRUCTURE
capi uaa user create "$USERNAME" \
    --email "$EMAIL" \
    --password "$TEMP_PASSWORD" \
    --given-name "John" \
    --family-name "Doe" \
    --phone-number "+1-555-0123"

echo "User created successfully!"

# Get user details - NEW STRUCTURE
echo "User details:"
capi uaa user get "$USERNAME"

# Update user - NEW STRUCTURE
echo "Updating user phone number..."
capi uaa user update "$USERNAME" --phone-number "+1-555-9999"

# Deactivate user - NEW STRUCTURE
echo "Deactivating user..."
capi uaa user deactivate "$USERNAME"

# Reactivate user - NEW STRUCTURE
echo "Reactivating user..."
capi uaa user activate "$USERNAME"

# Legacy commands (deprecated):
# capi uaa create-user "$USERNAME" --email "$EMAIL" --password "$TEMP_PASSWORD"
# capi uaa get-user "$USERNAME"
# capi uaa update-user "$USERNAME" --phone-number "+1-555-9999"
# capi uaa deactivate-user "$USERNAME"
# capi uaa activate-user "$USERNAME"

echo "User lifecycle demo complete!"
```

### Bulk User Creation

```bash
#!/bin/bash
# bulk-create-users.sh - Create multiple users from CSV

CSV_FILE="users.csv"

if [[ ! -f "$CSV_FILE" ]]; then
    echo "Creating sample CSV file: $CSV_FILE"
    cat > "$CSV_FILE" << EOF
username,email,firstname,lastname,phone
alice.smith,alice.smith@example.com,Alice,Smith,+1-555-0001
bob.jones,bob.jones@example.com,Bob,Jones,+1-555-0002
carol.white,carol.white@example.com,Carol,White,+1-555-0003
EOF
fi

echo "Creating users from $CSV_FILE..."

while IFS=, read -r username email firstname lastname phone; do
    # Skip header row
    if [[ "$username" == "username" ]]; then
        continue
    fi
    
    echo "Creating user: $username ($email)"
    
    # NEW STRUCTURE
    if capi uaa user create "$username" \
        --email "$email" \
        --given-name "$firstname" \
        --family-name "$lastname" \
        --phone-number "$phone" \
        --password "TempPass123!" \
        --verified; then
        echo "  ✓ User $username created successfully"
    else
        echo "  ✗ Failed to create user $username"
    fi
done < "$CSV_FILE"

echo "Bulk user creation complete!"
echo "Users created:"
# NEW STRUCTURE
capi uaa user list --filter 'email co "example.com"'

# Alternative: Use batch import - NEW STRUCTURE
# capi uaa batch import --file "$CSV_FILE"

# Legacy commands (deprecated):
# capi uaa create-user "$username" --email "$email" --given-name "$firstname"
# capi uaa list-users --filter 'email co "example.com"'
```

### User Search and Filtering

```bash
#!/bin/bash
# user-search.sh - Advanced user search examples

echo "=== User Search Examples ==="

# Find active users - NEW STRUCTURE
echo "1. Active users:"
capi uaa user list --filter 'active eq true' --attributes userName,email,active

# Find users by email domain - NEW STRUCTURE
echo -e "\n2. Users with example.com email:"
capi uaa user list --filter 'email co "example.com"' --attributes userName,email

# Legacy commands (deprecated):
# capi uaa list-users --filter 'active eq true' --attributes userName,email,active
# capi uaa list-users --filter 'email co "example.com"' --attributes userName,email

# Find recently created users (last 30 days)
THIRTY_DAYS_AGO=$(date -d '30 days ago' -u '+%Y-%m-%dT%H:%M:%S.000Z')
echo -e "\n3. Recently created users:"
capi uaa list-users --filter "meta.created gt \"$THIRTY_DAYS_AGO\"" --attributes userName,email,meta.created

# Find users by name pattern
echo -e "\n4. Users with 'admin' in username:"
capi uaa list-users --filter 'userName co "admin"' --attributes userName,email

# Find external users (not UAA origin)
echo -e "\n5. External identity provider users:"
capi uaa list-users --filter 'origin ne "uaa"' --attributes userName,email,origin

# Complex filter: Active users from specific domain, created this year
YEAR_START=$(date '+%Y-01-01T00:00:00.000Z')
echo -e "\n6. Active example.com users created this year:"
capi uaa list-users \
    --filter "active eq true and email co \"example.com\" and meta.created gt \"$YEAR_START\"" \
    --sort-by meta.created \
    --sort-order descending \
    --attributes userName,email,active,meta.created
```

## Group Management Examples

### Group and Membership Management

```bash
#!/bin/bash
# group-management.sh - Complete group management workflow

set -e

# Define groups and users
GROUPS=("developers" "qa-team" "admins")
USERS=("alice.smith" "bob.jones" "carol.white")

echo "=== Group Management Demo ==="

# Create groups
for group in "${GROUPS[@]}"; do
    echo "Creating group: $group"
    capi uaa create-group "$group" --description "Auto-created group: $group"
done

# Add users to groups
echo -e "\nAssigning users to groups..."

# Add Alice to developers and admins
capi uaa add-member developers alice.smith
capi uaa add-member admins alice.smith
echo "  ✓ Alice added to developers and admins"

# Add Bob to developers and qa-team
capi uaa add-member developers bob.jones
capi uaa add-member qa-team bob.jones
echo "  ✓ Bob added to developers and qa-team"

# Add Carol to qa-team
capi uaa add-member qa-team carol.white
echo "  ✓ Carol added to qa-team"

# Display group memberships
echo -e "\nGroup memberships:"
for group in "${GROUPS[@]}"; do
    echo "Group: $group"
    capi uaa get-group "$group" --output json | jq -r '.members[]? | "  - \(.value) (\(.type))"'
done

echo -e "\nGroup management demo complete!"
```

### External Group Mapping

```bash
#!/bin/bash
# external-group-mapping.sh - Map external LDAP/SAML groups

echo "=== External Group Mapping Demo ==="

# LDAP group mappings
LDAP_MAPPINGS=(
    "developers:CN=Developers,OU=Teams,DC=company,DC=com"
    "qa-team:CN=QA,OU=Teams,DC=company,DC=com"
    "admins:CN=Administrators,OU=Groups,DC=company,DC=com"
)

echo "Creating LDAP group mappings..."
for mapping in "${LDAP_MAPPINGS[@]}"; do
    IFS=':' read -r uaa_group ldap_group <<< "$mapping"
    
    echo "Mapping $uaa_group -> $ldap_group"
    capi uaa map-group \
        --group "$uaa_group" \
        --external-group "$ldap_group" \
        --origin ldap
done

# SAML group mappings
SAML_MAPPINGS=(
    "developers:dev-team"
    "qa-team:quality-assurance"
    "admins:system-admins"
)

echo -e "\nCreating SAML group mappings..."
for mapping in "${SAML_MAPPINGS[@]}"; do
    IFS=':' read -r uaa_group saml_group <<< "$mapping"
    
    echo "Mapping $uaa_group -> $saml_group"
    capi uaa map-group \
        --group "$uaa_group" \
        --external-group "$saml_group" \
        --origin saml
done

# List all mappings
echo -e "\nAll external group mappings:"
capi uaa list-group-mappings

echo -e "\nExternal group mapping demo complete!"
```

## OAuth Client Management Examples

### Client Types Demo

```bash
#!/bin/bash
# oauth-clients.sh - Create different types of OAuth clients

echo "=== OAuth Client Types Demo ==="

# 1. Web Application Client
echo "1. Creating web application client..."
capi uaa create-client web-app \
    --secret "web-app-secret-123" \
    --name "My Web Application" \
    --authorized-grant-types "authorization_code,refresh_token" \
    --scope "openid,profile,email,custom.read" \
    --redirect-uris "https://myapp.com/callback,https://myapp.com/auth" \
    --access-token-validity 3600 \
    --refresh-token-validity 86400

# 2. Single Page Application (SPA)
echo "2. Creating SPA client..."
capi uaa create-client spa-app \
    --name "My Single Page App" \
    --authorized-grant-types "authorization_code" \
    --scope "openid,profile" \
    --redirect-uris "https://spa.example.com/callback" \
    --access-token-validity 1800 \
    --auto-approve "openid,profile"

# 3. Mobile Application
echo "3. Creating mobile app client..."
capi uaa create-client mobile-app \
    --secret "mobile-secret-456" \
    --name "My Mobile App" \
    --authorized-grant-types "password,refresh_token" \
    --scope "openid,profile,offline_access" \
    --access-token-validity 3600 \
    --refresh-token-validity 604800

# 4. Service/Machine Client
echo "4. Creating service client..."
capi uaa create-client service-client \
    --secret "service-secret-789" \
    --name "Backend Service" \
    --authorized-grant-types "client_credentials" \
    --authorities "uaa.resource,custom.service.read,custom.service.write" \
    --scope "uaa.none"

# 5. Admin Client
echo "5. Creating admin client..."
capi uaa create-client admin-client \
    --secret "admin-secret-000" \
    --name "Administrative Client" \
    --authorized-grant-types "client_credentials" \
    --authorities "uaa.admin,scim.read,scim.write,groups.update" \
    --scope "uaa.none"

echo -e "\nCreated OAuth clients:"
capi uaa list-clients

echo -e "\nOAuth client demo complete!"
```

### Client Security Audit

```bash
#!/bin/bash
# client-audit.sh - Audit OAuth client security

echo "=== OAuth Client Security Audit ==="

# Get all clients and analyze security
echo "Analyzing OAuth client security..."

# Create temporary file for analysis
TEMP_FILE=$(mktemp)
capi uaa list-clients --output json > "$TEMP_FILE"

echo -e "\n1. Clients with password grant (potential security risk):"
jq -r '.resources[] | select(.authorized_grant_types[]? == "password") | "  - \(.client_id): \(.name // "No name")"' "$TEMP_FILE"

echo -e "\n2. Clients with implicit grant (deprecated):"
jq -r '.resources[] | select(.authorized_grant_types[]? == "implicit") | "  - \(.client_id): \(.name // "No name")"' "$TEMP_FILE"

echo -e "\n3. Clients with overly broad authorities:"
jq -r '.resources[] | select(.authorities[]? == "uaa.admin") | "  - \(.client_id): Has uaa.admin authority"' "$TEMP_FILE"

echo -e "\n4. Clients with long token validity (>1 hour):"
jq -r '.resources[] | select(.access_token_validity > 3600) | "  - \(.client_id): \(.access_token_validity)s token validity"' "$TEMP_FILE"

echo -e "\n5. Clients with auto-approve scopes:"
jq -r '.resources[] | select(.autoapprove | length > 0) | "  - \(.client_id): Auto-approves \(.autoapprove | join(", "))"' "$TEMP_FILE"

echo -e "\n6. Public clients (no secret):"
jq -r '.resources[] | select(.client_secret == null or .client_secret == "") | "  - \(.client_id): Public client (no secret)"' "$TEMP_FILE"

# Cleanup
rm "$TEMP_FILE"

echo -e "\nSecurity audit complete!"
echo "Review the findings above and update clients as needed."
```

## Bulk Operations Examples

### CSV Import/Export

```bash
#!/bin/bash
# csv-operations.sh - Import/export users via CSV

CSV_EXPORT="exported-users.csv"
CSV_IMPORT="import-users.csv"

echo "=== CSV Import/Export Demo ==="

# Export users to CSV
echo "Exporting users to CSV..."
echo "username,email,given_name,family_name,active,origin,created" > "$CSV_EXPORT"

capi uaa list-users --output json | jq -r '
.resources[] | 
[
    .userName,
    (.emails[0].value // ""),
    (.name.givenName // ""),
    (.name.familyName // ""),
    .active,
    .origin,
    .meta.created
] | @csv' >> "$CSV_EXPORT"

echo "Users exported to: $CSV_EXPORT"

# Create sample import file
echo -e "\nCreating sample import CSV..."
cat > "$CSV_IMPORT" << EOF
username,email,given_name,family_name,password
david.wilson,david.wilson@example.com,David,Wilson,TempPass123!
emma.davis,emma.davis@example.com,Emma,Davis,TempPass123!
frank.miller,frank.miller@example.com,Frank,Miller,TempPass123!
EOF

# Import users from CSV
echo "Importing users from CSV..."
while IFS=, read -r username email given_name family_name password; do
    # Skip header
    if [[ "$username" == "username" ]]; then
        continue
    fi
    
    echo "Importing: $username"
    if capi uaa create-user "$username" \
        --email "$email" \
        --given-name "$given_name" \
        --family-name "$family_name" \
        --password "$password"; then
        echo "  ✓ Imported $username"
    else
        echo "  ✗ Failed to import $username"
    fi
done < "$CSV_IMPORT"

echo -e "\nCSV operations complete!"
```

### Bulk Group Assignment

```bash
#!/bin/bash
# bulk-group-assignment.sh - Assign users to groups in bulk

echo "=== Bulk Group Assignment Demo ==="

# Define group assignments
declare -A GROUP_ASSIGNMENTS=(
    ["developers"]="alice.smith bob.jones david.wilson"
    ["qa-team"]="bob.jones carol.white emma.davis"
    ["admins"]="alice.smith frank.miller"
    ["all-users"]="alice.smith bob.jones carol.white david.wilson emma.davis frank.miller"
)

# Process group assignments
for group in "${!GROUP_ASSIGNMENTS[@]}"; do
    echo "Processing group: $group"
    
    # Create group if it doesn't exist
    if ! capi uaa get-group "$group" >/dev/null 2>&1; then
        echo "  Creating group: $group"
        capi uaa create-group "$group" --description "Auto-created group"
    fi
    
    # Add users to group
    users=(${GROUP_ASSIGNMENTS[$group]})
    for user in "${users[@]}"; do
        echo "  Adding $user to $group"
        if capi uaa add-member "$group" "$user"; then
            echo "    ✓ Added $user to $group"
        else
            echo "    ✗ Failed to add $user to $group"
        fi
    done
done

# Display final group memberships
echo -e "\nFinal group memberships:"
for group in "${!GROUP_ASSIGNMENTS[@]}"; do
    echo "Group: $group"
    capi uaa get-group "$group" --output json | jq -r '.members[]? | "  - \(.value)"' 2>/dev/null || echo "  (No members)"
done

echo -e "\nBulk group assignment complete!"
```

## Integration Scripts

### CF Integration

```bash
#!/bin/bash
# cf-integration.sh - Integrate UAA user management with CF operations

echo "=== CF + UAA Integration Demo ==="

# Function to create user and assign CF roles
create_cf_user() {
    local username=$1
    local email=$2
    local org=$3
    local space=$4
    local role=$5
    
    echo "Creating CF user: $username"
    
    # Create UAA user
    if capi uaa create-user "$username" \
        --email "$email" \
        --password "TempPass123!" \
        --verified; then
        echo "  ✓ UAA user created"
    else
        echo "  ✗ Failed to create UAA user"
        return 1
    fi
    
    # Assign CF role (using CF CLI or API)
    echo "  Assigning CF role: $role in $org/$space"
    # This would typically use CF CLI:
    # cf set-space-role "$username" "$org" "$space" "$role"
    echo "  ✓ CF role assigned (simulation)"
}

# Create development team
echo "Creating development team users..."
create_cf_user "dev1@example.com" "dev1@example.com" "my-org" "development" "SpaceDeveloper"
create_cf_user "dev2@example.com" "dev2@example.com" "my-org" "development" "SpaceDeveloper"

# Create QA team
echo -e "\nCreating QA team users..."
create_cf_user "qa1@example.com" "qa1@example.com" "my-org" "testing" "SpaceDeveloper"
create_cf_user "qa2@example.com" "qa2@example.com" "my-org" "testing" "SpaceAuditor"

echo -e "\nCF + UAA integration demo complete!"
```

### Monitoring and Alerting

```bash
#!/bin/bash
# monitoring.sh - Monitor UAA user activities

echo "=== UAA Monitoring Demo ==="

# Function to check user activity
check_user_activity() {
    local days_inactive=$1
    local cutoff_date
    cutoff_date=$(date -d "$days_inactive days ago" -u '+%Y-%m-%dT%H:%M:%S.000Z')
    
    echo "Checking for users inactive for $days_inactive+ days..."
    
    # Find users not modified recently (proxy for activity)
    capi uaa list-users \
        --filter "meta.lastModified lt \"$cutoff_date\"" \
        --output json | jq -r '
        .resources[] | 
        "User: \(.userName) | Last Modified: \(.meta.lastModified) | Active: \(.active)"'
}

# Function to check for security issues
security_check() {
    echo "Performing security checks..."
    
    # Check for users with weak/default passwords (this is a simulation)
    echo "1. Checking for potentially weak passwords..."
    echo "   (This would typically integrate with password policy checks)"
    
    # Check for inactive admin users
    echo "2. Checking admin group membership..."
    if capi uaa get-group admins >/dev/null 2>&1; then
        echo "   Admin group members:"
        capi uaa get-group admins --output json | jq -r '.members[]? | "   - \(.value)"'
    fi
    
    # Check for external users
    echo "3. External identity provider users:"
    capi uaa list-users \
        --filter 'origin ne "uaa"' \
        --attributes userName,origin,active | head -10
}

# Function to generate activity report
generate_report() {
    local report_file="uaa-activity-report-$(date +%Y%m%d).json"
    
    echo "Generating activity report: $report_file"
    
    # Create comprehensive report
    cat > "$report_file" << EOF
{
    "report_date": "$(date -u '+%Y-%m-%dT%H:%M:%S.000Z')",
    "total_users": $(capi uaa list-users --output json | jq '.totalResults'),
    "active_users": $(capi uaa list-users --filter 'active eq true' --output json | jq '.totalResults'),
    "inactive_users": $(capi uaa list-users --filter 'active eq false' --output json | jq '.totalResults'),
    "external_users": $(capi uaa list-users --filter 'origin ne "uaa"' --output json | jq '.totalResults'),
    "total_groups": $(capi uaa list-groups --output json | jq '.totalResults'),
    "total_clients": $(capi uaa list-clients --output json | jq '.totalResults')
}
EOF

    echo "Report saved to: $report_file"
    cat "$report_file" | jq '.'
}

# Run monitoring checks
check_user_activity 30
echo ""
security_check
echo ""
generate_report

echo -e "\nMonitoring demo complete!"
```

## Environment Setup

### Development Environment

```bash
#!/bin/bash
# dev-setup.sh - Set up development environment

echo "=== Development Environment Setup ==="

# Configuration
DEV_UAA_ENDPOINT="https://uaa.dev.example.com"
DEV_CLIENT_ID="dev-admin"
DEV_CLIENT_SECRET="dev-secret"

# Set up UAA target
echo "Setting up development UAA endpoint..."
capi uaa target "$DEV_UAA_ENDPOINT" --skip-ssl-validation

# Authenticate
echo "Authenticating with development credentials..."
capi uaa get-client-credentials-token \
    --client-id "$DEV_CLIENT_ID" \
    --client-secret "$DEV_CLIENT_SECRET"

# Create development users
echo "Creating development users..."
DEV_USERS=(
    "dev-user1:dev1@example.com:Developer:One"
    "dev-user2:dev2@example.com:Developer:Two"
    "test-user1:test1@example.com:Test:User"
)

for user_info in "${DEV_USERS[@]}"; do
    IFS=':' read -r username email first last <<< "$user_info"
    
    echo "Creating: $username"
    capi uaa create-user "$username" \
        --email "$email" \
        --given-name "$first" \
        --family-name "$last" \
        --password "DevPass123!" \
        --verified
done

# Create development groups
echo "Creating development groups..."
DEV_GROUPS=("dev-team" "test-team" "dev-admins")

for group in "${DEV_GROUPS[@]}"; do
    echo "Creating group: $group"
    capi uaa create-group "$group" --description "Development group: $group"
done

# Assign users to groups
echo "Assigning users to groups..."
capi uaa add-member dev-team dev-user1
capi uaa add-member dev-team dev-user2
capi uaa add-member test-team test-user1
capi uaa add-member dev-admins dev-user1

# Create development OAuth clients
echo "Creating development OAuth clients..."
capi uaa create-client dev-web-app \
    --secret "dev-web-secret" \
    --name "Development Web App" \
    --authorized-grant-types "authorization_code,refresh_token" \
    --scope "openid,profile,email" \
    --redirect-uris "http://localhost:3000/callback"

capi uaa create-client dev-api-client \
    --secret "dev-api-secret" \
    --name "Development API Client" \
    --authorized-grant-types "client_credentials" \
    --authorities "dev.api.read,dev.api.write"

echo "Development environment setup complete!"
echo ""
echo "Summary:"
echo "- UAA Endpoint: $DEV_UAA_ENDPOINT"
echo "- Users created: ${#DEV_USERS[@]}"
echo "- Groups created: ${#DEV_GROUPS[@]}"
echo "- OAuth clients created: 2"
echo ""
echo "You can now use these resources for development and testing."
```

## Usage Instructions

1. **Make scripts executable:**
   ```bash
   chmod +x *.sh
   ```

2. **Set up environment variables:**
   ```bash
   export UAA_ENDPOINT="https://uaa.your-domain.com"
   export UAA_CLIENT_ID="your-client-id"
   export UAA_CLIENT_SECRET="your-client-secret"
   ```

3. **Run examples:**
   ```bash
   ./auth-setup.sh
   ./user-lifecycle.sh
   ./group-management.sh
   ```

4. **Customize for your environment:**
   - Update endpoints and credentials
   - Modify user/group names
   - Adjust client configurations

These examples demonstrate real-world UAA management scenarios and can be adapted for your specific use cases.