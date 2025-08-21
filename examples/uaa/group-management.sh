#!/bin/bash
# group-management.sh - Comprehensive group management demonstration
#
# This script demonstrates:
# - Group creation and configuration
# - User-to-group membership management
# - External group mapping (LDAP/SAML)
# - Group queries and filtering
# - Group cleanup operations
#
# Prerequisites:
# - Authenticated UAA session (run auth-setup.sh first)
# - Admin privileges for group management operations

set -e

echo "=== UAA Group Management Demonstration ==="
echo

# Configuration
DEMO_GROUPS=("developers" "qa-team" "admins" "contractors")
DEMO_USERS=("alice.dev" "bob.qa" "carol.admin" "dave.contractor")
USER_EMAILS=("alice@example.com" "bob@example.com" "carol@example.com" "dave@example.com")

# Function to create demo users if they don't exist
create_demo_users() {
    echo "Step 1: Creating demo users for group management..."
    echo
    
    for i in "${!DEMO_USERS[@]}"; do
        local username="${DEMO_USERS[$i]}"
        local email="${USER_EMAILS[$i]}"
        
        echo "Creating user: $username ($email)"
        
        if capi uaa get-user "$username" >/dev/null 2>&1; then
            echo "  User $username already exists"
        else
            if capi uaa create-user "$username" \
                --email "$email" \
                --password "DemoPass123!" \
                --given-name "${username%%.*}" \
                --family-name "Demo" \
                --verified; then
                echo "  ✓ User $username created"
            else
                echo "  ✗ Failed to create user $username"
            fi
        fi
    done
    
    echo
}

# Function to create demo groups
create_demo_groups() {
    echo "Step 2: Creating demo groups..."
    echo
    
    # Group descriptions
    declare -A GROUP_DESCRIPTIONS=(
        ["developers"]="Software development team"
        ["qa-team"]="Quality assurance team"
        ["admins"]="System administrators"
        ["contractors"]="External contractors"
    )
    
    for group in "${DEMO_GROUPS[@]}"; do
        local description="${GROUP_DESCRIPTIONS[$group]}"
        
        echo "Creating group: $group"
        echo "Description: $description"
        
        if capi uaa get-group "$group" >/dev/null 2>&1; then
            echo "  Group $group already exists"
        else
            if capi uaa create-group "$group" --description "$description"; then
                echo "  ✓ Group $group created"
            else
                echo "  ✗ Failed to create group $group"
            fi
        fi
        echo
    done
}

# Function to demonstrate group membership management
manage_group_membership() {
    echo "Step 3: Managing group membership..."
    echo
    
    # Define group assignments
    declare -A GROUP_ASSIGNMENTS=(
        ["developers"]="alice.dev"
        ["qa-team"]="bob.qa alice.dev"
        ["admins"]="carol.admin"
        ["contractors"]="dave.contractor"
    )
    
    # Add users to groups
    for group in "${!GROUP_ASSIGNMENTS[@]}"; do
        echo "Managing membership for group: $group"
        
        # Convert space-separated users to array
        IFS=' ' read -ra users <<< "${GROUP_ASSIGNMENTS[$group]}"
        
        for user in "${users[@]}"; do
            echo "  Adding $user to $group"
            
            if capi uaa add-member "$group" "$user"; then
                echo "    ✓ Added $user to $group"
            else
                echo "    ✗ Failed to add $user to $group (may already be a member)"
            fi
        done
        echo
    done
    
    # Display current group memberships
    echo "Current group memberships:"
    for group in "${DEMO_GROUPS[@]}"; do
        echo "Group: $group"
        if capi uaa get-group "$group" --output json >/dev/null 2>&1; then
            capi uaa get-group "$group" --output json | jq -r '
                if .members then
                    .members[] | "  - \(.value) (\(.type))"
                else
                    "  (No members)"
                end'
        else
            echo "  (Group not found)"
        fi
        echo
    done
}

# Function to demonstrate advanced group queries
demonstrate_group_queries() {
    echo "Step 4: Advanced group queries and filtering..."
    echo
    
    echo "1. List all groups:"
    capi uaa list-groups --attributes displayName,description
    echo
    
    echo "2. Filter groups by name pattern:"
    echo "   Command: capi uaa list-groups --filter 'displayName sw \"dev\"'"
    capi uaa list-groups --filter 'displayName sw "dev"' --attributes displayName,description
    echo
    
    echo "3. Groups with members:"
    echo "   Showing groups that have at least one member..."
    for group in "${DEMO_GROUPS[@]}"; do
        local member_count
        member_count=$(capi uaa get-group "$group" --output json 2>/dev/null | jq '.members | length' 2>/dev/null || echo "0")
        if [[ "$member_count" -gt 0 ]]; then
            echo "   $group: $member_count member(s)"
        fi
    done
    echo
    
    echo "4. Get group details in JSON format:"
    echo "   Command: capi uaa get-group developers --output json"
    if capi uaa get-group developers --output json 2>/dev/null; then
        echo
    else
        echo "   Group 'developers' not found"
    fi
}

# Function to demonstrate external group mapping
demonstrate_external_mapping() {
    echo "Step 5: External group mapping (LDAP/SAML simulation)..."
    echo
    
    echo "Note: This demonstrates the mapping commands. In a real environment,"
    echo "you would have actual LDAP/SAML groups to map."
    echo
    
    # Simulate LDAP group mappings
    declare -A LDAP_MAPPINGS=(
        ["developers"]="CN=Developers,OU=Teams,DC=company,DC=com"
        ["qa-team"]="CN=QualityAssurance,OU=Teams,DC=company,DC=com"
        ["admins"]="CN=Administrators,OU=Groups,DC=company,DC=com"
    )
    
    echo "Creating LDAP group mappings:"
    for group in "${!LDAP_MAPPINGS[@]}"; do
        local ldap_dn="${LDAP_MAPPINGS[$group]}"
        echo "  Mapping $group -> $ldap_dn"
        echo "  Command: capi uaa map-group --group \"$group\" --external-group \"$ldap_dn\" --origin ldap"
        
        # Note: This might fail if the group doesn't exist or mapping already exists
        if capi uaa map-group \
            --group "$group" \
            --external-group "$ldap_dn" \
            --origin ldap 2>/dev/null; then
            echo "    ✓ Mapping created"
        else
            echo "    ⚠ Mapping may already exist or group not found"
        fi
    done
    echo
    
    # Simulate SAML group mappings
    declare -A SAML_MAPPINGS=(
        ["developers"]="dev-team"
        ["qa-team"]="qa-team"
        ["admins"]="system-admins"
    )
    
    echo "Creating SAML group mappings:"
    for group in "${!SAML_MAPPINGS[@]}"; do
        local saml_group="${SAML_MAPPINGS[$group]}"
        echo "  Mapping $group -> $saml_group"
        echo "  Command: capi uaa map-group --group \"$group\" --external-group \"$saml_group\" --origin saml"
        
        if capi uaa map-group \
            --group "$group" \
            --external-group "$saml_group" \
            --origin saml 2>/dev/null; then
            echo "    ✓ Mapping created"
        else
            echo "    ⚠ Mapping may already exist or group not found"
        fi
    done
    echo
    
    # List all group mappings
    echo "All external group mappings:"
    if capi uaa list-group-mappings 2>/dev/null; then
        echo
    else
        echo "  No mappings found or command not available"
        echo
    fi
    
    echo "LDAP mappings only:"
    echo "Command: capi uaa list-group-mappings --origin ldap"
    if capi uaa list-group-mappings --origin ldap 2>/dev/null; then
        echo
    else
        echo "  No LDAP mappings found"
        echo
    fi
}

# Function to demonstrate membership manipulation
demonstrate_membership_changes() {
    echo "Step 6: Group membership manipulation..."
    echo
    
    # Add alice.dev to multiple groups
    echo "Adding alice.dev to multiple groups:"
    additional_groups=("admins" "contractors")
    
    for group in "${additional_groups[@]}"; do
        echo "  Adding alice.dev to $group"
        if capi uaa add-member "$group" "alice.dev"; then
            echo "    ✓ Added to $group"
        else
            echo "    ✗ Failed to add to $group (may already be a member)"
        fi
    done
    echo
    
    # Show alice.dev's group memberships
    echo "alice.dev's current group memberships:"
    for group in "${DEMO_GROUPS[@]}"; do
        if capi uaa get-group "$group" --output json 2>/dev/null | jq -e '.members[]? | select(.value == "alice.dev")' >/dev/null 2>&1; then
            echo "  ✓ Member of $group"
        else
            echo "  ✗ Not a member of $group"
        fi
    done
    echo
    
    # Remove alice.dev from contractors group
    echo "Removing alice.dev from contractors group:"
    echo "Command: capi uaa remove-member contractors alice.dev"
    if capi uaa remove-member "contractors" "alice.dev"; then
        echo "  ✓ Removed from contractors"
    else
        echo "  ✗ Failed to remove from contractors (may not be a member)"
    fi
    echo
    
    # Verify removal
    echo "Verifying removal from contractors group:"
    if capi uaa get-group "contractors" --output json 2>/dev/null | jq -e '.members[]? | select(.value == "alice.dev")' >/dev/null 2>&1; then
        echo "  ✗ alice.dev is still a member of contractors"
    else
        echo "  ✓ alice.dev is no longer a member of contractors"
    fi
    echo
}

# Function to show group management summary
show_management_summary() {
    echo "Step 7: Group management summary..."
    echo
    
    echo "Groups created:"
    for group in "${DEMO_GROUPS[@]}"; do
        if capi uaa get-group "$group" >/dev/null 2>&1; then
            local member_count
            member_count=$(capi uaa get-group "$group" --output json 2>/dev/null | jq '.members | length' 2>/dev/null || echo "0")
            echo "  ✓ $group ($member_count members)"
        else
            echo "  ✗ $group (not found)"
        fi
    done
    echo
    
    echo "Users created:"
    for user in "${DEMO_USERS[@]}"; do
        if capi uaa get-user "$user" >/dev/null 2>&1; then
            echo "  ✓ $user"
        else
            echo "  ✗ $user (not found)"
        fi
    done
    echo
    
    # Group membership matrix
    echo "Group membership matrix:"
    printf "%-15s" "User/Group"
    for group in "${DEMO_GROUPS[@]}"; do
        printf "%-12s" "$group"
    done
    echo
    
    for user in "${DEMO_USERS[@]}"; do
        printf "%-15s" "$user"
        for group in "${DEMO_GROUPS[@]}"; do
            if capi uaa get-group "$group" --output json 2>/dev/null | jq -e '.members[]? | select(.value == "'$user'")' >/dev/null 2>&1; then
                printf "%-12s" "✓"
            else
                printf "%-12s" "✗"
            fi
        done
        echo
    done
    echo
}

# Function to cleanup demo resources
cleanup_demo_resources() {
    echo "Step 8: Cleanup (optional)..."
    echo
    
    read -p "Do you want to clean up the demo resources (groups and users)? (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "Cleaning up demo resources..."
        
        # Remove external group mappings first
        echo "Removing external group mappings..."
        declare -A CLEANUP_MAPPINGS=(
            ["developers:ldap"]="CN=Developers,OU=Teams,DC=company,DC=com"
            ["qa-team:ldap"]="CN=QualityAssurance,OU=Teams,DC=company,DC=com"
            ["admins:ldap"]="CN=Administrators,OU=Groups,DC=company,DC=com"
            ["developers:saml"]="dev-team"
            ["qa-team:saml"]="qa-team"
            ["admins:saml"]="system-admins"
        )
        
        for mapping in "${!CLEANUP_MAPPINGS[@]}"; do
            IFS=':' read -r group origin <<< "$mapping"
            local external_group="${CLEANUP_MAPPINGS[$mapping]}"
            
            echo "  Removing mapping: $group ($origin) -> $external_group"
            if capi uaa unmap-group \
                --group "$group" \
                --external-group "$external_group" \
                --origin "$origin" 2>/dev/null; then
                echo "    ✓ Mapping removed"
            else
                echo "    ⚠ Mapping may not exist"
            fi
        done
        
        # Delete groups
        echo "Deleting demo groups..."
        for group in "${DEMO_GROUPS[@]}"; do
            echo "  Deleting group: $group"
            if capi uaa delete-group "$group" --force 2>/dev/null; then
                echo "    ✓ Group deleted"
            else
                echo "    ⚠ Group may not exist or deletion failed"
            fi
        done
        
        # Delete users
        echo "Deleting demo users..."
        for user in "${DEMO_USERS[@]}"; do
            echo "  Deleting user: $user"
            if capi uaa delete-user "$user" --force 2>/dev/null; then
                echo "    ✓ User deleted"
            else
                echo "    ⚠ User may not exist or deletion failed"
            fi
        done
        
        echo "Cleanup completed!"
    else
        echo "Demo resources preserved."
        echo
        echo "To manually clean up later:"
        echo "Groups: capi uaa delete-group {${DEMO_GROUPS[*]}} --force"
        echo "Users:  capi uaa delete-user {${DEMO_USERS[*]}} --force"
    fi
}

# Main execution flow
main() {
    create_demo_users
    create_demo_groups
    manage_group_membership
    demonstrate_group_queries
    demonstrate_external_mapping
    demonstrate_membership_changes
    show_management_summary
    cleanup_demo_resources
}

# Execute main function
main

echo
echo "=== Group Management Demonstration Complete ==="
echo
echo "This demonstration covered:"
echo "1. ✓ User creation for group management"
echo "2. ✓ Group creation with descriptions"
echo "3. ✓ Group membership management"
echo "4. ✓ Advanced group queries and filtering"
echo "5. ✓ External group mapping (LDAP/SAML)"
echo "6. ✓ Membership manipulation operations"
echo "7. ✓ Group management summary and reporting"
echo "8. ✓ Resource cleanup procedures"
echo
echo "Key commands demonstrated:"
echo "- capi uaa create-group"
echo "- capi uaa list-groups"
echo "- capi uaa get-group"
echo "- capi uaa add-member"
echo "- capi uaa remove-member"
echo "- capi uaa map-group"
echo "- capi uaa unmap-group"
echo "- capi uaa list-group-mappings"
echo "- capi uaa delete-group"