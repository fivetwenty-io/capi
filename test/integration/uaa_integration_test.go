// +build integration

package integration

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// UAAIntegrationTestSuite provides integration tests for UAA functionality
type UAAIntegrationTestSuite struct {
	suite.Suite
	capiPath      string
	uaaEndpoint   string
	clientID      string
	clientSecret  string
	adminUser     string
	adminPassword string
	testUserName  string
	testGroupName string
	testClientID  string
}

// SetupSuite initializes the test environment
func (suite *UAAIntegrationTestSuite) SetupSuite() {
	// Check for required environment variables
	suite.uaaEndpoint = os.Getenv("UAA_ENDPOINT")
	suite.clientID = os.Getenv("UAA_CLIENT_ID")
	suite.clientSecret = os.Getenv("UAA_CLIENT_SECRET")
	suite.adminUser = os.Getenv("UAA_ADMIN_USER")
	suite.adminPassword = os.Getenv("UAA_ADMIN_PASSWORD")

	if suite.uaaEndpoint == "" {
		suite.T().Skip("UAA_ENDPOINT environment variable not set, skipping integration tests")
	}

	// Find the capi binary
	suite.capiPath = os.Getenv("CAPI_BINARY_PATH")
	if suite.capiPath == "" {
		// Try to find it relative to test directory
		suite.capiPath = "../../capi"
	}

	// Verify capi binary exists and is executable
	if _, err := os.Stat(suite.capiPath); os.IsNotExist(err) {
		suite.T().Skipf("capi binary not found at %s, skipping integration tests", suite.capiPath)
	}

	// Generate unique test names to avoid conflicts
	timestamp := time.Now().Unix()
	suite.testUserName = fmt.Sprintf("test-user-%d", timestamp)
	suite.testGroupName = fmt.Sprintf("test-group-%d", timestamp)
	suite.testClientID = fmt.Sprintf("test-client-%d", timestamp)

	// Set UAA target
	suite.runCapiCommand("users", "target", suite.uaaEndpoint)
}

// TearDownSuite cleans up the test environment
func (suite *UAAIntegrationTestSuite) TearDownSuite() {
	// Clean up test resources
	if suite.testUserName != "" {
		suite.runCapiCommand("users", "delete-user", suite.testUserName, "--force")
	}
	if suite.testGroupName != "" {
		suite.runCapiCommand("users", "delete-group", suite.testGroupName, "--force")
	}
	if suite.testClientID != "" {
		suite.runCapiCommand("users", "delete-client", suite.testClientID, "--force")
	}
}

// runCapiCommand executes a capi command and returns stdout, stderr, and error
func (suite *UAAIntegrationTestSuite) runCapiCommand(args ...string) (string, string, error) {
	cmd := exec.Command(suite.capiPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// runCapiCommandWithInput executes a capi command with stdin input
func (suite *UAAIntegrationTestSuite) runCapiCommandWithInput(input string, args ...string) (string, string, error) {
	cmd := exec.Command(suite.capiPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = strings.NewReader(input)
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// authenticateAsAdmin authenticates as admin user for setup operations
func (suite *UAAIntegrationTestSuite) authenticateAsAdmin() {
	if suite.adminUser != "" && suite.adminPassword != "" {
		stdout, stderr, err := suite.runCapiCommand("users", "get-password-token",
			"--username", suite.adminUser,
			"--password", suite.adminPassword,
			"--client-id", suite.clientID,
			"--client-secret", suite.clientSecret)
		if err != nil {
			suite.T().Logf("Failed to authenticate as admin: %s\nStderr: %s", stdout, stderr)
		}
	} else if suite.clientID != "" && suite.clientSecret != "" {
		stdout, stderr, err := suite.runCapiCommand("users", "get-client-credentials-token",
			"--client-id", suite.clientID,
			"--client-secret", suite.clientSecret)
		if err != nil {
			suite.T().Logf("Failed to authenticate with client credentials: %s\nStderr: %s", stdout, stderr)
		}
	}
}

// Test UAA Context and Connection Management
func (suite *UAAIntegrationTestSuite) TestUAAContextManagement() {
	// Test target command
	stdout, stderr, err := suite.runCapiCommand("users", "target", suite.uaaEndpoint)
	suite.NoError(err, "Failed to set UAA target: %s", stderr)
	suite.Contains(stdout, "UAA endpoint set to")

	// Test context command
	stdout, stderr, err = suite.runCapiCommand("users", "context")
	suite.NoError(err, "Failed to get UAA context: %s", stderr)
	suite.Contains(stdout, suite.uaaEndpoint)

	// Test info command
	stdout, stderr, err = suite.runCapiCommand("users", "info")
	suite.NoError(err, "Failed to get UAA info: %s", stderr)
	suite.Contains(strings.ToLower(stdout), "uaa")

	// Test version command
	stdout, stderr, err = suite.runCapiCommand("users", "version")
	suite.NoError(err, "Failed to get UAA version: %s", stderr)
	suite.Contains(stdout, suite.uaaEndpoint)
}

// Test OAuth2 Token Management Workflow
func (suite *UAAIntegrationTestSuite) TestTokenManagement() {
	if suite.clientID == "" || suite.clientSecret == "" {
		suite.T().Skip("No client credentials provided, skipping token tests")
	}

	// Test client credentials token
	stdout, stderr, err := suite.runCapiCommand("users", "get-client-credentials-token",
		"--client-id", suite.clientID,
		"--client-secret", suite.clientSecret)
	suite.NoError(err, "Failed to get client credentials token: %s", stderr)
	suite.Contains(stdout, "Access Token")

	// Test token keys
	stdout, stderr, err = suite.runCapiCommand("users", "get-token-keys")
	suite.NoError(err, "Failed to get token keys: %s", stderr)

	// Test single token key
	stdout, stderr, err = suite.runCapiCommand("users", "get-token-key")
	suite.NoError(err, "Failed to get token key: %s", stderr)
	suite.Contains(stdout, "Key Type")
}

// Test User Lifecycle Management
func (suite *UAAIntegrationTestSuite) TestUserLifecycle() {
	suite.authenticateAsAdmin()

	// Test create user
	stdout, stderr, err := suite.runCapiCommand("users", "create-user", suite.testUserName,
		"--email", fmt.Sprintf("%s@example.com", suite.testUserName),
		"--password", "TempPassword123!",
		"--given-name", "Test",
		"--family-name", "User")
	suite.NoError(err, "Failed to create user: %s", stderr)
	suite.Contains(stdout, suite.testUserName)

	// Test get user
	stdout, stderr, err = suite.runCapiCommand("users", "get-user", suite.testUserName)
	suite.NoError(err, "Failed to get user: %s", stderr)
	suite.Contains(stdout, suite.testUserName)

	// Test list users with filter
	stdout, stderr, err = suite.runCapiCommand("users", "list-users",
		"--filter", fmt.Sprintf("userName eq \"%s\"", suite.testUserName))
	suite.NoError(err, "Failed to list users: %s", stderr)
	suite.Contains(stdout, suite.testUserName)

	// Test update user
	stdout, stderr, err = suite.runCapiCommand("users", "update-user", suite.testUserName,
		"--phone-number", "+1-555-0123")
	suite.NoError(err, "Failed to update user: %s", stderr)

	// Test deactivate user
	stdout, stderr, err = suite.runCapiCommand("users", "deactivate-user", suite.testUserName)
	suite.NoError(err, "Failed to deactivate user: %s", stderr)
	suite.Contains(stdout, "has been deactivated")

	// Test activate user
	stdout, stderr, err = suite.runCapiCommand("users", "activate-user", suite.testUserName)
	suite.NoError(err, "Failed to activate user: %s", stderr)
	suite.Contains(stdout, "has been activated")

	// Test delete user (cleanup will be done in TearDownSuite)
}

// Test Group Management Workflow
func (suite *UAAIntegrationTestSuite) TestGroupManagement() {
	suite.authenticateAsAdmin()

	// Test create group
	stdout, stderr, err := suite.runCapiCommand("users", "create-group", suite.testGroupName,
		"--description", "Integration test group")
	suite.NoError(err, "Failed to create group: %s", stderr)
	suite.Contains(stdout, suite.testGroupName)

	// Test get group
	stdout, stderr, err = suite.runCapiCommand("users", "get-group", suite.testGroupName)
	suite.NoError(err, "Failed to get group: %s", stderr)
	suite.Contains(stdout, suite.testGroupName)

	// Test list groups
	stdout, stderr, err = suite.runCapiCommand("users", "list-groups",
		"--filter", fmt.Sprintf("displayName eq \"%s\"", suite.testGroupName))
	suite.NoError(err, "Failed to list groups: %s", stderr)
	suite.Contains(stdout, suite.testGroupName)

	// Test group membership (requires a user)
	if suite.testUserName != "" {
		// Add user to group
		stdout, stderr, err = suite.runCapiCommand("users", "add-member", suite.testGroupName, suite.testUserName)
		if err == nil {
			suite.Contains(stdout, "Successfully added member")

			// Remove user from group
			stdout, stderr, err = suite.runCapiCommand("users", "remove-member", suite.testGroupName, suite.testUserName)
			suite.NoError(err, "Failed to remove member from group: %s", stderr)
			suite.Contains(stdout, "Successfully removed member")
		}
	}
}

// Test OAuth Client Management Workflow
func (suite *UAAIntegrationTestSuite) TestClientManagement() {
	suite.authenticateAsAdmin()

	// Test create client
	stdout, stderr, err := suite.runCapiCommand("users", "create-client", suite.testClientID,
		"--secret", "client-secret-123",
		"--name", "Integration Test Client",
		"--authorized-grant-types", "client_credentials",
		"--scope", "uaa.resource")
	suite.NoError(err, "Failed to create client: %s", stderr)
	suite.Contains(stdout, suite.testClientID)

	// Test get client
	stdout, stderr, err = suite.runCapiCommand("users", "get-client", suite.testClientID)
	suite.NoError(err, "Failed to get client: %s", stderr)
	suite.Contains(stdout, suite.testClientID)
	suite.Contains(stdout, "***") // Secret should be masked

	// Test get client with secret
	stdout, stderr, err = suite.runCapiCommand("users", "get-client", suite.testClientID, "--show-secret")
	suite.NoError(err, "Failed to get client with secret: %s", stderr)
	suite.Contains(stdout, "client-secret-123")

	// Test list clients
	stdout, stderr, err = suite.runCapiCommand("users", "list-clients")
	suite.NoError(err, "Failed to list clients: %s", stderr)
	suite.Contains(stdout, suite.testClientID)

	// Test update client
	stdout, stderr, err = suite.runCapiCommand("users", "update-client", suite.testClientID,
		"--name", "Updated Integration Test Client")
	suite.NoError(err, "Failed to update client: %s", stderr)

	// Test set client secret
	stdout, stderr, err = suite.runCapiCommand("users", "set-client-secret", suite.testClientID,
		"--secret", "new-client-secret-456")
	suite.NoError(err, "Failed to set client secret: %s", stderr)
	suite.Contains(stdout, "Successfully updated secret")
}

// Test Utility Commands
func (suite *UAAIntegrationTestSuite) TestUtilityCommands() {
	suite.authenticateAsAdmin()

	// Test userinfo command
	stdout, stderr, err := suite.runCapiCommand("users", "userinfo")
	if err == nil {
		suite.Contains(stdout, "Username")
	}

	// Test curl command
	stdout, stderr, err = suite.runCapiCommand("users", "curl", "/info")
	suite.NoError(err, "Failed to curl UAA info: %s", stderr)
	suite.Contains(strings.ToLower(stdout), "app")
}

// Test End-to-End User Management Workflow
func (suite *UAAIntegrationTestSuite) TestEndToEndUserWorkflow() {
	suite.authenticateAsAdmin()

	userEmail := fmt.Sprintf("%s@example.com", suite.testUserName)
	
	// Complete user lifecycle workflow
	// 1. Create user
	stdout, _, err := suite.runCapiCommand("users", "create-user", suite.testUserName,
		"--email", userEmail,
		"--password", "TempPassword123!",
		"--given-name", "End2End",
		"--family-name", "Test")
	suite.NoError(err, "Failed in E2E workflow: create user")

	// 2. Verify user exists
	stdout, _, err = suite.runCapiCommand("users", "get-user", suite.testUserName)
	suite.NoError(err, "Failed in E2E workflow: get user")
	suite.Contains(stdout, userEmail)

	// 3. Create group for user
	stdout, _, err = suite.runCapiCommand("users", "create-group", suite.testGroupName,
		"--description", "E2E test group")
	suite.NoError(err, "Failed in E2E workflow: create group")

	// 4. Add user to group
	stdout, _, err = suite.runCapiCommand("users", "add-member", suite.testGroupName, suite.testUserName)
	if err == nil {
		suite.Contains(stdout, "Successfully added member")
	}

	// 5. Update user information
	stdout, _, err = suite.runCapiCommand("users", "update-user", suite.testUserName,
		"--phone-number", "+1-555-9999")
	suite.NoError(err, "Failed in E2E workflow: update user")

	// 6. Deactivate and reactivate user
	stdout, _, err = suite.runCapiCommand("users", "deactivate-user", suite.testUserName)
	suite.NoError(err, "Failed in E2E workflow: deactivate user")

	stdout, _, err = suite.runCapiCommand("users", "activate-user", suite.testUserName)
	suite.NoError(err, "Failed in E2E workflow: activate user")

	// Cleanup is handled in TearDownSuite
}

// Test Error Handling and Edge Cases
func (suite *UAAIntegrationTestSuite) TestErrorHandling() {
	// Test operations without authentication
	stdout, stderr, err := suite.runCapiCommand("users", "create-user", "should-fail")
	suite.Error(err, "Expected error for unauthenticated request")
	suite.Contains(stderr, "not authenticated")

	// Test non-existent resource operations
	suite.authenticateAsAdmin()
	
	stdout, stderr, err = suite.runCapiCommand("users", "get-user", "non-existent-user-12345")
	suite.Error(err, "Expected error for non-existent user")

	stdout, stderr, err = suite.runCapiCommand("users", "get-group", "non-existent-group-12345")
	suite.Error(err, "Expected error for non-existent group")

	stdout, stderr, err = suite.runCapiCommand("users", "get-client", "non-existent-client-12345")
	suite.Error(err, "Expected error for non-existent client")
}

// TestUAAIntegrationSuite runs the complete integration test suite
func TestUAAIntegrationSuite(t *testing.T) {
	suite.Run(t, new(UAAIntegrationTestSuite))
}

// Test individual command help and usage
func TestUAACommandHelp(t *testing.T) {
	capiPath := os.Getenv("CAPI_BINARY_PATH")
	if capiPath == "" {
		capiPath = "../../capi"
	}

	if _, err := os.Stat(capiPath); os.IsNotExist(err) {
		t.Skipf("capi binary not found at %s, skipping help tests", capiPath)
	}

	commands := [][]string{
		{"users", "--help"},
		{"users", "context", "--help"},
		{"users", "target", "--help"},
		{"users", "info", "--help"},
		{"users", "version", "--help"},
		{"users", "get-password-token", "--help"},
		{"users", "get-client-credentials-token", "--help"},
		{"users", "create-user", "--help"},
		{"users", "list-users", "--help"},
		{"users", "create-group", "--help"},
		{"users", "create-client", "--help"},
		{"users", "curl", "--help"},
		{"users", "userinfo", "--help"},
	}

	for _, cmdArgs := range commands {
		t.Run(strings.Join(cmdArgs, " "), func(t *testing.T) {
			cmd := exec.Command(capiPath, cmdArgs...)
			var stdout bytes.Buffer
			cmd.Stdout = &stdout
			err := cmd.Run()
			
			// Help commands should exit with code 0 and contain usage information
			assert.NoError(t, err, "Help command should not error")
			output := stdout.String()
			assert.Contains(t, output, "Usage:", "Help output should contain usage information")
		})
	}
}