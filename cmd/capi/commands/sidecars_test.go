package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSidecarsCommand(t *testing.T) {
	cmd := NewSidecarsCommand()
	assert.Equal(t, "sidecars", cmd.Use)
	assert.Equal(t, []string{"sidecar", "sc"}, cmd.Aliases)
	assert.Equal(t, "Manage application sidecars", cmd.Short)
	assert.Equal(t, "View, update, and delete sidecars for applications", cmd.Long)

	// Check subcommands are added
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 4)
	
	var commandNames []string
	for _, subcmd := range subcommands {
		commandNames = append(commandNames, subcmd.Name())
	}
	
	assert.Contains(t, commandNames, "get")
	assert.Contains(t, commandNames, "update")
	assert.Contains(t, commandNames, "delete")
	assert.Contains(t, commandNames, "list-for-process")
}

func TestSidecarsGetCommand(t *testing.T) {
	cmd := newSidecarsGetCommand()
	assert.Equal(t, "get SIDECAR_GUID", cmd.Use)
	assert.Equal(t, "Get sidecar details", cmd.Short)
	assert.Equal(t, "Display detailed information about a specific sidecar", cmd.Long)
	assert.NotNil(t, cmd.RunE)
	assert.NotNil(t, cmd.Args)
}

func TestSidecarsUpdateCommand(t *testing.T) {
	cmd := newSidecarsUpdateCommand()
	assert.Equal(t, "update SIDECAR_GUID", cmd.Use)
	assert.Equal(t, "Update a sidecar", cmd.Short)
	assert.Equal(t, "Update an existing sidecar configuration", cmd.Long)
	assert.NotNil(t, cmd.RunE)
	assert.NotNil(t, cmd.Args)

	// Check flags
	flags := []string{"name", "command", "process-types", "memory"}
	for _, flagName := range flags {
		flag := cmd.Flags().Lookup(flagName)
		assert.NotNil(t, flag, "Flag %s should exist", flagName)
	}
}

func TestSidecarsDeleteCommand(t *testing.T) {
	cmd := newSidecarsDeleteCommand()
	assert.Equal(t, "delete SIDECAR_GUID", cmd.Use)
	assert.Equal(t, "Delete a sidecar", cmd.Short)
	assert.Equal(t, "Delete a sidecar from an application", cmd.Long)
	assert.NotNil(t, cmd.RunE)
	assert.NotNil(t, cmd.Args)

	// Check force flag
	forceFlag := cmd.Flags().Lookup("force")
	assert.NotNil(t, forceFlag)
	assert.Equal(t, "f", forceFlag.Shorthand)
	assert.Equal(t, "false", forceFlag.DefValue)
}

func TestSidecarsListForProcessCommand(t *testing.T) {
	cmd := newSidecarsListForProcessCommand()
	assert.Equal(t, "list-for-process PROCESS_GUID", cmd.Use)
	assert.Equal(t, "List sidecars for a process", cmd.Short)
	assert.Equal(t, "List all sidecars associated with a specific process", cmd.Long)
	assert.NotNil(t, cmd.RunE)
	assert.NotNil(t, cmd.Args)

	// Check pagination flags
	assert.NotNil(t, cmd.Flags().Lookup("all"))
	assert.NotNil(t, cmd.Flags().Lookup("per-page"))
	
	perPageFlag := cmd.Flags().Lookup("per-page")
	assert.Equal(t, "50", perPageFlag.DefValue)
}