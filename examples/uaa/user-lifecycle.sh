#!/bin/bash
# user-lifecycle.sh - Complete user management workflow demonstration
# 
# This script demonstrates a complete user lifecycle including:
# - User creation with full attributes
# - User retrieval and verification
# - User updates and modifications
# - User activation/deactivation
# - User deletion
#
# Prerequisites:
# - Authenticated UAA session (run auth-setup.sh first)
# - Admin privileges for user management operations

set -e

# Configuration
USERNAME="${1:-demo.user}"
EMAIL="${2:-${USERNAME}@example.com}"
TEMP_PASSWORD="TempPass123!"

echo "=== UAA User Lifecycle Demonstration ==="
echo "Username: $USERNAME"
echo "Email: $EMAIL"
echo

# Function to pause and wait for user input
pause() {
    read -p "Press Enter to continue..."
}

# Step 1: Create user with full attributes
echo "Step 1: Creating user with complete profile..."
echo "Command: capi uaa create-user $USERNAME \\"
echo "  --email $EMAIL \\"
echo "  --password [REDACTED] \\"
echo "  --given-name Demo \\"
echo "  --family-name User \\"
echo "  --phone-number '+1-555-0123' \\"
echo "  --verified \\"
echo "  --active"
echo

if capi uaa create-user "$USERNAME" \
    --email "$EMAIL" \
    --password "$TEMP_PASSWORD" \
    --given-name "Demo" \
    --family-name "User" \
    --phone-number "+1-555-0123" \
    --verified \
    --active; then
    echo "✓ User created successfully!"
else
    echo "✗ Failed to create user. User may already exist."
    echo "Continuing with existing user..."
fi

echo
pause

# Step 2: Retrieve and display user details
echo "Step 2: Retrieving user details..."
echo "Command: capi uaa get-user $USERNAME"
echo

capi uaa get-user "$USERNAME"

echo
echo "Retrieving user details in JSON format for analysis:"
echo "Command: capi uaa get-user $USERNAME --output json"
echo

USER_JSON=$(capi uaa get-user "$USERNAME" --output json)
echo "$USER_JSON" | jq '.'

echo
pause

# Step 3: Update user attributes
echo "Step 3: Updating user attributes..."
echo "Command: capi uaa update-user $USERNAME \\"
echo "  --phone-number '+1-555-9999' \\"
echo "  --family-name 'UpdatedUser'"
echo

if capi uaa update-user "$USERNAME" \
    --phone-number "+1-555-9999" \
    --family-name "UpdatedUser"; then
    echo "✓ User updated successfully!"
else
    echo "✗ Failed to update user"
fi

echo
echo "Verifying updates:"
capi uaa get-user "$USERNAME" --attributes userName,name,phoneNumbers

echo
pause

# Step 4: Deactivate user
echo "Step 4: Deactivating user account..."
echo "Command: capi uaa deactivate-user $USERNAME"
echo

if capi uaa deactivate-user "$USERNAME"; then
    echo "✓ User deactivated successfully!"
else
    echo "✗ Failed to deactivate user"
fi

echo
echo "Verifying deactivation:"
capi uaa get-user "$USERNAME" --attributes userName,active

echo
pause

# Step 5: Reactivate user
echo "Step 5: Reactivating user account..."
echo "Command: capi uaa activate-user $USERNAME"
echo

if capi uaa activate-user "$USERNAME"; then
    echo "✓ User reactivated successfully!"
else
    echo "✗ Failed to reactivate user"
fi

echo
echo "Verifying reactivation:"
capi uaa get-user "$USERNAME" --attributes userName,active

echo
pause

# Step 6: Advanced user queries
echo "Step 6: Advanced user queries and filtering..."
echo

echo "Finding users with similar email domain:"
echo "Command: capi uaa list-users --filter 'email co \"example.com\"'"
capi uaa list-users --filter 'email co "example.com"' --attributes userName,email

echo
echo "Finding recently modified users:"
RECENT_DATE=$(date -d '1 hour ago' -u '+%Y-%m-%dT%H:%M:%S.000Z')
echo "Command: capi uaa list-users --filter 'meta.lastModified gt \"$RECENT_DATE\"'"
capi uaa list-users --filter "meta.lastModified gt \"$RECENT_DATE\"" --attributes userName,meta.lastModified

echo
pause

# Step 7: User attribute management
echo "Step 7: Managing user attributes..."
echo

echo "Updating multiple attributes:"
echo "Command: capi uaa update-user $USERNAME \\"
echo "  --given-name 'UpdatedDemo' \\"
echo "  --family-name 'FinalUser' \\"
echo "  --phone-number '+1-555-1234'"

if capi uaa update-user "$USERNAME" \
    --given-name "UpdatedDemo" \
    --family-name "FinalUser" \
    --phone-number "+1-555-1234"; then
    echo "✓ Multiple attributes updated successfully!"
fi

echo
echo "Final user state:"
capi uaa get-user "$USERNAME"

echo
pause

# Step 8: Cleanup (optional)
echo "Step 8: Cleanup (optional)..."
echo

read -p "Do you want to delete the demo user? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "Deleting user: $USERNAME"
    echo "Command: capi uaa delete-user $USERNAME --force"
    
    if capi uaa delete-user "$USERNAME" --force; then
        echo "✓ User deleted successfully!"
    else
        echo "✗ Failed to delete user"
    fi
else
    echo "User $USERNAME preserved for further testing."
    echo "To delete manually later, run: capi uaa delete-user $USERNAME"
fi

echo
echo "=== User Lifecycle Demonstration Complete ==="
echo
echo "Summary of operations performed:"
echo "1. ✓ Created user with complete profile"
echo "2. ✓ Retrieved user details in multiple formats"
echo "3. ✓ Updated user attributes"
echo "4. ✓ Deactivated user account"
echo "5. ✓ Reactivated user account"
echo "6. ✓ Performed advanced queries and filtering"
echo "7. ✓ Managed multiple user attributes"
echo "8. ✓ Optional cleanup"
echo
echo "This demonstrates the complete user management capabilities"
echo "available through the capi uaa commands."