// +build integration

package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUAAWorkflow_CompleteUserJourney tests a complete user management journey
func TestUAAWorkflow_CompleteUserJourney(t *testing.T) {
	config := LoadTestConfig()
	config.SkipIfMissingConfig(t)
	
	runner := NewCommandRunner(config, t)
	
	// Setup
	require.NoError(t, runner.SetupUAATarget())
	require.NoError(t, runner.AuthenticateAdmin())
	
	// Generate unique test names
	userName := GenerateTestName("workflow-user")
	groupName := GenerateTestName("workflow-group")
	clientName := GenerateTestName("workflow-client")
	
	defer func() {
		// Cleanup
		runner.CleanupResource("user", userName)
		runner.CleanupResource("group", groupName)
		runner.CleanupResource("client", clientName)
	}()

	// 1. Create OAuth client for user authentication
	stdout, stderr, err := runner.Run("users", "create-client", clientName,
		"--secret", "workflow-secret-123",
		"--name", "Workflow Test Client",
		"--authorized-grant-types", "password,refresh_token",
		"--scope", "openid,profile")
	require.NoError(t, err, "Failed to create OAuth client: %s", stderr)
	assert.Contains(t, stdout, clientName)

	// 2. Create user
	userEmail := fmt.Sprintf("%s@workflow.test", userName)
	stdout, stderr, err = runner.Run("users", "create-user", userName,
		"--email", userEmail,
		"--password", "WorkflowPass123!",
		"--given-name", "Workflow",
		"--family-name", "User",
		"--phone-number", "+1-555-0123")
	require.NoError(t, err, "Failed to create user: %s", stderr)
	assert.Contains(t, stdout, userName)

	// 3. Verify user with JSON output
	stdout, stderr, err = runner.Run("users", "get-user", userName, "--output", "json")
	require.NoError(t, err, "Failed to get user with JSON output: %s", stderr)
	AssertJSONOutput(t, stdout)
	assert.Contains(t, stdout, userEmail)

	// 4. Create group
	stdout, stderr, err = runner.Run("users", "create-group", groupName,
		"--description", "Workflow test group for integration testing")
	require.NoError(t, err, "Failed to create group: %s", stderr)
	assert.Contains(t, stdout, groupName)

	// 5. Add user to group
	stdout, stderr, err = runner.Run("users", "add-member", groupName, userName)
	require.NoError(t, err, "Failed to add user to group: %s", stderr)
	assert.Contains(t, stdout, "Successfully added member")

	// 6. Update user information
	stdout, stderr, err = runner.Run("users", "update-user", userName,
		"--family-name", "UpdatedUser",
		"--phone-number", "+1-555-9999")
	require.NoError(t, err, "Failed to update user: %s", stderr)

	// 7. Verify user update
	stdout, stderr, err = runner.Run("users", "get-user", userName)
	require.NoError(t, err, "Failed to get updated user: %s", stderr)
	assert.Contains(t, stdout, "UpdatedUser")
	assert.Contains(t, stdout, "+1-555-9999")

	// 8. Test user authentication with created client
	stdout, stderr, err = runner.Run("users", "get-password-token",
		"--username", userName,
		"--password", "WorkflowPass123!",
		"--client-id", clientName,
		"--client-secret", "workflow-secret-123")
	if err == nil {
		assert.Contains(t, stdout, "Access Token")
		
		// 9. Test userinfo with user token
		stdout, stderr, err = runner.Run("users", "userinfo")
		if err == nil {
			assert.Contains(t, stdout, userName)
		}
	}

	// 10. Remove user from group
	stdout, stderr, err = runner.Run("users", "remove-member", groupName, userName)
	require.NoError(t, err, "Failed to remove user from group: %s", stderr)
	assert.Contains(t, stdout, "Successfully removed member")

	// 11. Deactivate and reactivate user
	stdout, stderr, err = runner.Run("users", "deactivate-user", userName)
	require.NoError(t, err, "Failed to deactivate user: %s", stderr)
	assert.Contains(t, stdout, "has been deactivated")

	stdout, stderr, err = runner.Run("users", "activate-user", userName)
	require.NoError(t, err, "Failed to activate user: %s", stderr)
	assert.Contains(t, stdout, "has been activated")
}

// TestUAAWorkflow_OutputFormats tests all output formats work correctly
func TestUAAWorkflow_OutputFormats(t *testing.T) {
	config := LoadTestConfig()
	config.SkipIfMissingConfig(t)
	
	runner := NewCommandRunner(config, t)
	
	// Setup
	require.NoError(t, runner.SetupUAATarget())
	require.NoError(t, runner.AuthenticateAdmin())

	// Test output formats for info command
	formats := []string{"table", "json", "yaml"}
	
	for _, format := range formats {
		t.Run(fmt.Sprintf("info_%s_format", format), func(t *testing.T) {
			stdout, stderr, err := runner.Run("users", "info", "--output", format)
			require.NoError(t, err, "Failed to get info with %s format: %s", format, stderr)
			
			switch format {
			case "json":
				AssertJSONOutput(t, stdout)
			case "yaml":
				AssertYAMLOutput(t, stdout)
			case "table":
				assert.Contains(t, stdout, "Property")
				assert.Contains(t, stdout, "Value")
			}
		})
	}
}

// TestUAAWorkflow_ErrorScenarios tests error handling in real scenarios
func TestUAAWorkflow_ErrorScenarios(t *testing.T) {
	config := LoadTestConfig()
	config.SkipIfMissingConfig(t)
	
	runner := NewCommandRunner(config, t)
	
	// Setup target but don't authenticate
	require.NoError(t, runner.SetupUAATarget())

	// Test operations without authentication
	testCases := []struct {
		name        string
		args        []string
		expectError bool
		errorText   string
	}{
		{
			name:        "create user without auth",
			args:        []string{"users", "create-user", "should-fail"},
			expectError: true,
			errorText:   "not authenticated",
		},
		{
			name:        "list users without auth",
			args:        []string{"users", "list-users"},
			expectError: true,
			errorText:   "not authenticated",
		},
		{
			name:        "get non-existent user",
			args:        []string{"users", "get-user", "non-existent-user-12345"},
			expectError: true,
			errorText:   "", // Will fail during auth check first
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := runner.Run(tc.args...)
			if tc.expectError {
				assert.Error(t, err, "Expected error for: %s", tc.name)
				if tc.errorText != "" {
					assert.Contains(t, stderr, tc.errorText, "Expected specific error text")
				}
			} else {
				assert.NoError(t, err, "Unexpected error for: %s\nStderr: %s", tc.name, stderr)
			}
		})
	}
}

// TestUAAWorkflow_PaginationAndFiltering tests list commands with pagination
func TestUAAWorkflow_PaginationAndFiltering(t *testing.T) {
	config := LoadTestConfig()
	config.SkipIfMissingConfig(t)
	
	runner := NewCommandRunner(config, t)
	
	// Setup
	require.NoError(t, runner.SetupUAATarget())
	require.NoError(t, runner.AuthenticateAdmin())

	// Test user listing with pagination
	stdout, stderr, err := runner.Run("users", "list-users", "--count", "5")
	require.NoError(t, err, "Failed to list users with pagination: %s", stderr)
	assert.Contains(t, stdout, "Username")

	// Test user listing with SCIM filter (if users exist)
	stdout, stderr, err = runner.Run("users", "list-users", "--filter", "active eq true", "--count", "10")
	require.NoError(t, err, "Failed to list users with filter: %s", stderr)

	// Test group listing
	stdout, stderr, err = runner.Run("users", "list-groups", "--count", "5")
	require.NoError(t, err, "Failed to list groups with pagination: %s", stderr)
	assert.Contains(t, stdout, "Display Name")

	// Test client listing
	stdout, stderr, err = runner.Run("users", "list-clients", "--count", "5")
	require.NoError(t, err, "Failed to list clients with pagination: %s", stderr)
	assert.Contains(t, stdout, "Client ID")
}

// TestUAAWorkflow_TokenManagement tests token lifecycle management
func TestUAAWorkflow_TokenManagement(t *testing.T) {
	config := LoadTestConfig()
	if config.ClientID == "" || config.ClientSecret == "" {
		t.Skip("Client credentials not provided, skipping token management tests")
	}
	config.SkipIfMissingConfig(t)
	
	runner := NewCommandRunner(config, t)
	
	// Setup
	require.NoError(t, runner.SetupUAATarget())

	// Test client credentials token flow
	stdout, stderr, err := runner.Run("users", "get-client-credentials-token",
		"--client-id", config.ClientID,
		"--client-secret", config.ClientSecret,
		"--token-format", "opaque")
	require.NoError(t, err, "Failed to get client credentials token: %s", stderr)
	assert.Contains(t, stdout, "Access Token")

	// Test token key retrieval
	stdout, stderr, err = runner.Run("users", "get-token-keys")
	require.NoError(t, err, "Failed to get token keys: %s", stderr)
	assert.Contains(t, stdout, "Key ID")

	// Test single token key
	stdout, stderr, err = runner.Run("users", "get-token-key")
	require.NoError(t, err, "Failed to get token key: %s", stderr)
	assert.Contains(t, stdout, "Key Type")

	// Test context shows authentication
	stdout, stderr, err = runner.Run("users", "context")
	require.NoError(t, err, "Failed to get context: %s", stderr)
	assert.Contains(t, stdout, "Authenticated")
}

// TestUAAWorkflow_DirectAPIAccess tests the curl command functionality
func TestUAAWorkflow_DirectAPIAccess(t *testing.T) {
	config := LoadTestConfig()
	config.SkipIfMissingConfig(t)
	
	runner := NewCommandRunner(config, t)
	
	// Setup
	require.NoError(t, runner.SetupUAATarget())
	require.NoError(t, runner.AuthenticateAdmin())

	// Test GET request to info endpoint
	stdout, stderr, err := runner.Run("users", "curl", "/info")
	require.NoError(t, err, "Failed to curl /info endpoint: %s", stderr)
	assert.Contains(t, stdout, "Status: 200")
	assert.Contains(t, strings.ToLower(stdout), "app")

	// Test GET request with custom headers
	stdout, stderr, err = runner.Run("users", "curl", "/info",
		"--method", "GET",
		"--header", "Accept: application/json")
	require.NoError(t, err, "Failed to curl with custom headers: %s", stderr)
	assert.Contains(t, stdout, "Status: 200")
}