package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSpaceQuotasCommand(t *testing.T) {
	cmd := NewSpaceQuotasCommand()
	assert.Equal(t, "space-quotas", cmd.Use)
	assert.Equal(t, []string{"space-quota", "sq"}, cmd.Aliases)
	assert.Equal(t, "Manage space quotas", cmd.Short)
	assert.Equal(t, "List, create, update, delete, apply, and remove space quotas", cmd.Long)

	// Check subcommands are added
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 7)

	var commandNames []string
	for _, subcmd := range subcommands {
		commandNames = append(commandNames, subcmd.Name())
	}

	assert.Contains(t, commandNames, "list")
	assert.Contains(t, commandNames, "get")
	assert.Contains(t, commandNames, "create")
	assert.Contains(t, commandNames, "update")
	assert.Contains(t, commandNames, "delete")
	assert.Contains(t, commandNames, "apply")
	assert.Contains(t, commandNames, "remove")
}

func TestSpaceQuotasListCommand(t *testing.T) {
	cmd := newSpaceQuotasListCommand()
	assert.Equal(t, "list", cmd.Use)
	assert.Equal(t, "List space quotas", cmd.Short)
	assert.Equal(t, "List all space quotas", cmd.Long)
	assert.NotNil(t, cmd.RunE)

	// Check flags
	assert.NotNil(t, cmd.Flags().Lookup("all"))
	assert.NotNil(t, cmd.Flags().Lookup("per-page"))
	assert.NotNil(t, cmd.Flags().Lookup("org"))

	// Check flag defaults
	allFlag := cmd.Flags().Lookup("all")
	assert.Equal(t, "false", allFlag.DefValue)

	perPageFlag := cmd.Flags().Lookup("per-page")
	assert.Equal(t, "50", perPageFlag.DefValue)

	orgFlag := cmd.Flags().Lookup("org")
	assert.Equal(t, "o", orgFlag.Shorthand)
}

func TestSpaceQuotasGetCommand(t *testing.T) {
	cmd := newSpaceQuotasGetCommand()
	assert.Equal(t, "get QUOTA_NAME_OR_GUID", cmd.Use)
	assert.Equal(t, "Get space quota details", cmd.Short)
	assert.Equal(t, "Display detailed information about a specific space quota", cmd.Long)
	assert.NotNil(t, cmd.RunE)
	assert.NotNil(t, cmd.Args)
}

func TestSpaceQuotasCreateCommand(t *testing.T) {
	cmd := newSpaceQuotasCreateCommand()
	assert.Equal(t, "create", cmd.Use)
	assert.Equal(t, "Create a new space quota", cmd.Short)
	assert.Equal(t, "Create a new Cloud Foundry space quota", cmd.Long)
	assert.NotNil(t, cmd.RunE)

	// Check all flags exist
	flags := []string{
		"name", "org", "total-memory", "instance-memory", "instances",
		"app-tasks", "log-rate-limit", "paid-services", "service-instances",
		"service-keys", "routes", "reserved-ports",
	}

	for _, flagName := range flags {
		flag := cmd.Flags().Lookup(flagName)
		assert.NotNil(t, flag, "Flag %s should exist", flagName)
	}

	// Check required flags
	nameFlag := cmd.Flags().Lookup("name")
	assert.NotNil(t, nameFlag)
	assert.Equal(t, "n", nameFlag.Shorthand)

	orgFlag := cmd.Flags().Lookup("org")
	assert.NotNil(t, orgFlag)
	assert.Equal(t, "o", orgFlag.Shorthand)
}

func TestSpaceQuotasUpdateCommand(t *testing.T) {
	cmd := newSpaceQuotasUpdateCommand()
	assert.Equal(t, "update QUOTA_NAME_OR_GUID", cmd.Use)
	assert.Equal(t, "Update a space quota", cmd.Short)
	assert.Equal(t, "Update an existing Cloud Foundry space quota", cmd.Long)
	assert.NotNil(t, cmd.RunE)
	assert.NotNil(t, cmd.Args)
}

func TestSpaceQuotasDeleteCommand(t *testing.T) {
	cmd := newSpaceQuotasDeleteCommand()
	assert.Equal(t, "delete QUOTA_NAME_OR_GUID", cmd.Use)
	assert.Equal(t, "Delete a space quota", cmd.Short)
	assert.Equal(t, "Delete a Cloud Foundry space quota", cmd.Long)
	assert.NotNil(t, cmd.RunE)
	assert.NotNil(t, cmd.Args)

	// Check force flag
	forceFlag := cmd.Flags().Lookup("force")
	assert.NotNil(t, forceFlag)
	assert.Equal(t, "f", forceFlag.Shorthand)
}

func TestSpaceQuotasApplyCommand(t *testing.T) {
	cmd := newSpaceQuotasApplyCommand()
	assert.Equal(t, "apply QUOTA_NAME_OR_GUID SPACE_NAME_OR_GUID...", cmd.Use)
	assert.Equal(t, "Apply quota to spaces", cmd.Short)
	assert.Equal(t, "Apply a space quota to one or more spaces", cmd.Long)
	assert.NotNil(t, cmd.RunE)

	assert.NotNil(t, cmd.Args)
}

func TestSpaceQuotasRemoveCommand(t *testing.T) {
	cmd := newSpaceQuotasRemoveCommand()
	assert.Equal(t, "remove QUOTA_NAME_OR_GUID SPACE_NAME_OR_GUID", cmd.Use)
	assert.Equal(t, "Remove quota from space", cmd.Short)
	assert.Equal(t, "Remove a space quota from a specific space", cmd.Long)
	assert.NotNil(t, cmd.RunE)
	assert.NotNil(t, cmd.Args)
}
