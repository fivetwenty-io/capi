package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateUsersCreateClientCommand(t *testing.T) {
	cmd := createUsersCreateClientCommand()
	assert.Equal(t, "create-client <client-id>", cmd.Use)
	assert.Equal(t, "Create OAuth client", cmd.Short)
	assert.NotNil(t, cmd.RunE)
	assert.Contains(t, cmd.Long, "Create an OAuth client")

	// Check flags
	assert.NotNil(t, cmd.Flags().Lookup("secret"))
	assert.NotNil(t, cmd.Flags().Lookup("name"))
	assert.NotNil(t, cmd.Flags().Lookup("authorized-grant-types"))
	assert.NotNil(t, cmd.Flags().Lookup("redirect-uri"))
	assert.NotNil(t, cmd.Flags().Lookup("scope"))
	assert.NotNil(t, cmd.Flags().Lookup("authorities"))
	assert.NotNil(t, cmd.Flags().Lookup("access-token-validity"))
	assert.NotNil(t, cmd.Flags().Lookup("refresh-token-validity"))
	assert.NotNil(t, cmd.Flags().Lookup("auto-approve"))
	assert.NotNil(t, cmd.Flags().Lookup("allow-public"))
}

func TestCreateUsersGetClientCommand(t *testing.T) {
	cmd := createUsersGetClientCommand()
	assert.Equal(t, "get-client <client-id>", cmd.Use)
	assert.Equal(t, "Get client details", cmd.Short)
	assert.NotNil(t, cmd.RunE)
	assert.Contains(t, cmd.Long, "View OAuth client")

	// Check flags
	assert.NotNil(t, cmd.Flags().Lookup("show-secret"))
}

func TestCreateUsersListClientsCommand(t *testing.T) {
	cmd := createUsersListClientsCommand()
	assert.Equal(t, "list-clients", cmd.Use)
	assert.Equal(t, "List OAuth clients", cmd.Short)
	assert.NotNil(t, cmd.RunE)
	assert.Contains(t, cmd.Long, "List all OAuth clients")

	// Check flags
	assert.NotNil(t, cmd.Flags().Lookup("filter"))
	assert.NotNil(t, cmd.Flags().Lookup("sort-by"))
	assert.NotNil(t, cmd.Flags().Lookup("sort-order"))
	assert.NotNil(t, cmd.Flags().Lookup("count"))
	assert.NotNil(t, cmd.Flags().Lookup("start-index"))
	assert.NotNil(t, cmd.Flags().Lookup("all"))
}

func TestCreateUsersUpdateClientCommand(t *testing.T) {
	cmd := createUsersUpdateClientCommand()
	assert.Equal(t, "update-client <client-id>", cmd.Use)
	assert.Equal(t, "Update OAuth client", cmd.Short)
	assert.NotNil(t, cmd.RunE)
	assert.Contains(t, cmd.Long, "Update an OAuth client")

	// Check flags
	assert.NotNil(t, cmd.Flags().Lookup("name"))
	assert.NotNil(t, cmd.Flags().Lookup("authorized-grant-types"))
	assert.NotNil(t, cmd.Flags().Lookup("redirect-uri"))
	assert.NotNil(t, cmd.Flags().Lookup("scope"))
	assert.NotNil(t, cmd.Flags().Lookup("authorities"))
	assert.NotNil(t, cmd.Flags().Lookup("access-token-validity"))
	assert.NotNil(t, cmd.Flags().Lookup("refresh-token-validity"))
	assert.NotNil(t, cmd.Flags().Lookup("auto-approve"))
	assert.NotNil(t, cmd.Flags().Lookup("allow-public"))
}

func TestCreateUsersSetClientSecretCommand(t *testing.T) {
	cmd := createUsersSetClientSecretCommand()
	assert.Equal(t, "set-client-secret <client-id>", cmd.Use)
	assert.Equal(t, "Update client secret", cmd.Short)
	assert.NotNil(t, cmd.RunE)
	assert.Contains(t, cmd.Long, "Update the secret")

	// Check flags
	assert.NotNil(t, cmd.Flags().Lookup("secret"))
}

func TestCreateUsersDeleteClientCommand(t *testing.T) {
	cmd := createUsersDeleteClientCommand()
	assert.Equal(t, "delete-client <client-id>", cmd.Use)
	assert.Equal(t, "Delete OAuth client", cmd.Short)
	assert.NotNil(t, cmd.RunE)
	assert.Contains(t, cmd.Long, "Delete an OAuth client")

	// Check flags
	assert.NotNil(t, cmd.Flags().Lookup("force"))
}

// Note: Display functions would require more complex setup to test properly
// as they depend on actual UAA API responses and internal data structures.
// These tests focus on command structure validation instead.
