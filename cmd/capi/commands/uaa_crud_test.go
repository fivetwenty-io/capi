package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateUsersCreateUserCommand(t *testing.T) {
	cmd := createUsersCreateUserCommand()
	assert.Equal(t, "create-user <username>", cmd.Use)
	assert.Equal(t, "Create a new user", cmd.Short)
	assert.NotNil(t, cmd.RunE)
	assert.Contains(t, cmd.Long, "Create a new user")

	// Check flags
	assert.NotNil(t, cmd.Flags().Lookup("email"))
	assert.NotNil(t, cmd.Flags().Lookup("password"))
	assert.NotNil(t, cmd.Flags().Lookup("given-name"))
	assert.NotNil(t, cmd.Flags().Lookup("family-name"))
	assert.NotNil(t, cmd.Flags().Lookup("phone-number"))
	assert.NotNil(t, cmd.Flags().Lookup("origin"))
	assert.NotNil(t, cmd.Flags().Lookup("active"))
	assert.NotNil(t, cmd.Flags().Lookup("verified"))
}

func TestCreateUsersGetUserCommand(t *testing.T) {
	cmd := createUsersGetUserCommand()
	assert.Equal(t, "get-user <username>", cmd.Use)
	assert.Equal(t, "Get user details", cmd.Short)
	assert.NotNil(t, cmd.RunE)
	assert.Contains(t, cmd.Long, "Look up a user")

	// Check flags
	assert.NotNil(t, cmd.Flags().Lookup("attributes"))
}

func TestCreateUsersListUsersCommand(t *testing.T) {
	cmd := createUsersListUsersCommand()
	assert.Equal(t, "list-users", cmd.Use)
	assert.Equal(t, "List users", cmd.Short)
	assert.NotNil(t, cmd.RunE)
	assert.Contains(t, cmd.Long, "Search and list users")

	// Check flags
	assert.NotNil(t, cmd.Flags().Lookup("filter"))
	assert.NotNil(t, cmd.Flags().Lookup("sort-by"))
	assert.NotNil(t, cmd.Flags().Lookup("sort-order"))
	assert.NotNil(t, cmd.Flags().Lookup("attributes"))
	assert.NotNil(t, cmd.Flags().Lookup("count"))
	assert.NotNil(t, cmd.Flags().Lookup("start-index"))
	assert.NotNil(t, cmd.Flags().Lookup("all"))
}

func TestCreateUsersUpdateUserCommand(t *testing.T) {
	cmd := createUsersUpdateUserCommand()
	assert.Equal(t, "update-user <username>", cmd.Use)
	assert.Equal(t, "Update user attributes", cmd.Short)
	assert.NotNil(t, cmd.RunE)
	assert.Contains(t, cmd.Long, "Update attributes")

	// Check flags
	assert.NotNil(t, cmd.Flags().Lookup("email"))
	assert.NotNil(t, cmd.Flags().Lookup("given-name"))
	assert.NotNil(t, cmd.Flags().Lookup("family-name"))
	assert.NotNil(t, cmd.Flags().Lookup("phone-number"))
	assert.NotNil(t, cmd.Flags().Lookup("active"))
	assert.NotNil(t, cmd.Flags().Lookup("verified"))
}

func TestCreateUsersActivateUserCommand(t *testing.T) {
	cmd := createUsersActivateUserCommand()
	assert.Equal(t, "activate-user <username>", cmd.Use)
	assert.Equal(t, "Activate a user account", cmd.Short)
	assert.NotNil(t, cmd.RunE)
	assert.Contains(t, cmd.Long, "Activate a user account")
}

func TestCreateUsersDeactivateUserCommand(t *testing.T) {
	cmd := createUsersDeactivateUserCommand()
	assert.Equal(t, "deactivate-user <username>", cmd.Use)
	assert.Equal(t, "Deactivate a user account", cmd.Short)
	assert.NotNil(t, cmd.RunE)
	assert.Contains(t, cmd.Long, "Deactivate a user account")
}

func TestCreateUsersDeleteUserCommand(t *testing.T) {
	cmd := createUsersDeleteUserCommand()
	assert.Equal(t, "delete-user <username>", cmd.Use)
	assert.Equal(t, "Delete a user", cmd.Short)
	assert.NotNil(t, cmd.RunE)
	assert.Contains(t, cmd.Long, "Delete a user")

	// Check flags
	assert.NotNil(t, cmd.Flags().Lookup("force"))
}

// Note: Display functions would require more complex setup to test properly
// as they depend on actual UAA API responses and internal data structures.
// These tests focus on command structure validation instead.
