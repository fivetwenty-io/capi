#!/bin/bash
# auth-setup.sh - UAA Authentication Setup
#
# This script demonstrates different UAA authentication methods
# and helps establish an authenticated session for other examples.
#
# Supported authentication methods:
# 1. Client credentials (machine-to-machine)
# 2. Username/password (user authentication)
# 3. Environment variable based authentication
#
# Prerequisites:
# - Access to UAA server
# - Valid client credentials or user credentials
# - Network connectivity to UAA endpoint

set -e

# Default configuration (can be overridden by environment variables)
DEFAULT_UAA_ENDPOINT="${UAA_ENDPOINT:-https://uaa.system.domain.com}"
DEFAULT_CLIENT_ID="${UAA_CLIENT_ID:-admin}"
DEFAULT_CLIENT_SECRET="${UAA_CLIENT_SECRET}"
DEFAULT_USERNAME="${UAA_USERNAME}"
DEFAULT_PASSWORD="${UAA_PASSWORD}"

echo "=== UAA Authentication Setup ==="
echo

# Function to check if UAA endpoint is accessible
check_uaa_connectivity() {
    local endpoint="$1"
    echo "Checking UAA endpoint connectivity: $endpoint"
    
    if capi uaa target "$endpoint" >/dev/null 2>&1; then
        echo "✓ UAA endpoint is accessible"
        return 0
    else
        echo "✗ UAA endpoint is not accessible"
        return 1
    fi
}

# Function to authenticate with client credentials
auth_client_credentials() {
    local client_id="$1"
    local client_secret="$2"
    
    echo "Authenticating with client credentials..."
    echo "Client ID: $client_id"
    echo "Client Secret: [REDACTED]"
    
    if capi uaa get-client-credentials-token \
        --client-id "$client_id" \
        --client-secret "$client_secret"; then
        echo "✓ Client credentials authentication successful"
        return 0
    else
        echo "✗ Client credentials authentication failed"
        return 1
    fi
}

# Function to authenticate with username/password
auth_password_grant() {
    local username="$1"
    local password="$2"
    local client_id="$3"
    local client_secret="$4"
    
    echo "Authenticating with username/password..."
    echo "Username: $username"
    echo "Password: [REDACTED]"
    echo "Client ID: $client_id"
    
    local auth_cmd="capi uaa get-password-token --username \"$username\" --password \"$password\" --client-id \"$client_id\""
    
    if [[ -n "$client_secret" ]]; then
        auth_cmd="$auth_cmd --client-secret \"$client_secret\""
        echo "Client Secret: [REDACTED]"
    fi
    
    if eval "$auth_cmd"; then
        echo "✓ Password grant authentication successful"
        return 0
    else
        echo "✗ Password grant authentication failed"
        return 1
    fi
}

# Function to display current authentication status
show_auth_status() {
    echo "Current UAA context:"
    if capi uaa context; then
        echo
        echo "Authentication test - getting server info:"
        capi uaa info
    else
        echo "✗ Not authenticated or no UAA target set"
    fi
}

# Main authentication flow
main() {
    echo "Step 1: Setting UAA target endpoint"
    
    # Prompt for UAA endpoint if not provided
    if [[ -z "$DEFAULT_UAA_ENDPOINT" ]] || [[ "$DEFAULT_UAA_ENDPOINT" == "https://uaa.system.domain.com" ]]; then
        read -p "Enter UAA endpoint URL: " UAA_ENDPOINT
        DEFAULT_UAA_ENDPOINT="$UAA_ENDPOINT"
    fi
    
    echo "UAA Endpoint: $DEFAULT_UAA_ENDPOINT"
    
    # Set UAA target
    if ! check_uaa_connectivity "$DEFAULT_UAA_ENDPOINT"; then
        echo
        read -p "UAA endpoint not accessible. Skip SSL validation? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            echo "Setting UAA target with SSL validation disabled..."
            if capi uaa target "$DEFAULT_UAA_ENDPOINT" --skip-ssl-validation; then
                echo "✓ UAA target set (SSL validation disabled)"
            else
                echo "✗ Failed to set UAA target"
                exit 1
            fi
        else
            echo "Please check your UAA endpoint and network connectivity."
            exit 1
        fi
    fi
    
    echo
    echo "Step 2: Authentication"
    echo
    echo "Available authentication methods:"
    echo "1. Client credentials (recommended for automation)"
    echo "2. Username/password (for user authentication)"
    echo "3. Use environment variables"
    echo
    
    # Check if environment variables are available
    if [[ -n "$DEFAULT_CLIENT_ID" && -n "$DEFAULT_CLIENT_SECRET" ]]; then
        echo "Environment variables detected:"
        echo "  UAA_CLIENT_ID=$DEFAULT_CLIENT_ID"
        echo "  UAA_CLIENT_SECRET=[SET]"
        echo
        read -p "Use environment variables for authentication? (Y/n): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Nn]$ ]]; then
            if auth_client_credentials "$DEFAULT_CLIENT_ID" "$DEFAULT_CLIENT_SECRET"; then
                echo
                show_auth_status
                return 0
            else
                echo "Environment variable authentication failed. Falling back to manual input."
                echo
            fi
        fi
    fi
    
    # Manual authentication method selection
    read -p "Choose authentication method (1-2): " -n 1 -r
    echo
    
    case $REPLY in
        1)
            echo "Selected: Client credentials authentication"
            echo
            
            # Get client credentials
            if [[ -n "$DEFAULT_CLIENT_ID" ]]; then
                read -p "Client ID [$DEFAULT_CLIENT_ID]: " CLIENT_ID
                CLIENT_ID="${CLIENT_ID:-$DEFAULT_CLIENT_ID}"
            else
                read -p "Client ID: " CLIENT_ID
            fi
            
            if [[ -n "$DEFAULT_CLIENT_SECRET" ]]; then
                echo "Using client secret from environment variable"
                CLIENT_SECRET="$DEFAULT_CLIENT_SECRET"
            else
                read -s -p "Client Secret: " CLIENT_SECRET
                echo
            fi
            
            echo
            auth_client_credentials "$CLIENT_ID" "$CLIENT_SECRET"
            ;;
            
        2)
            echo "Selected: Username/password authentication"
            echo
            
            # Get user credentials
            if [[ -n "$DEFAULT_USERNAME" ]]; then
                read -p "Username [$DEFAULT_USERNAME]: " USERNAME
                USERNAME="${USERNAME:-$DEFAULT_USERNAME}"
            else
                read -p "Username: " USERNAME
            fi
            
            if [[ -n "$DEFAULT_PASSWORD" ]]; then
                echo "Using password from environment variable"
                PASSWORD="$DEFAULT_PASSWORD"
            else
                read -s -p "Password: " PASSWORD
                echo
            fi
            
            # Get client credentials for password grant
            if [[ -n "$DEFAULT_CLIENT_ID" ]]; then
                read -p "Client ID [$DEFAULT_CLIENT_ID]: " CLIENT_ID
                CLIENT_ID="${CLIENT_ID:-$DEFAULT_CLIENT_ID}"
            else
                read -p "Client ID [cf]: " CLIENT_ID
                CLIENT_ID="${CLIENT_ID:-cf}"
            fi
            
            read -s -p "Client Secret (optional): " CLIENT_SECRET
            echo
            
            echo
            auth_password_grant "$USERNAME" "$PASSWORD" "$CLIENT_ID" "$CLIENT_SECRET"
            ;;
            
        *)
            echo "Invalid selection. Please run the script again."
            exit 1
            ;;
    esac
    
    echo
    show_auth_status
}

# Run main function
main

echo
echo "=== Authentication Setup Complete ==="
echo
echo "You can now run other UAA management examples:"
echo "  ./user-lifecycle.sh"
echo "  ./group-management.sh"
echo "  ./oauth-clients.sh"
echo
echo "To check your authentication status at any time:"
echo "  capi uaa context"
echo
echo "To get current user information:"
echo "  capi uaa userinfo"