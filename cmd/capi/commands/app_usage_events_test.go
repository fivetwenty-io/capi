package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAppUsageEventsCommand(t *testing.T) {
	cmd := NewAppUsageEventsCommand()
	assert.Equal(t, "app-usage-events", cmd.Use)
	assert.Equal(t, []string{"app-usage", "app-events", "aue"}, cmd.Aliases)
	assert.Equal(t, "Manage application usage events", cmd.Short)
	assert.Equal(t, "View and manage application usage events for monitoring and billing", cmd.Long)

	// Check subcommands are added
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 3)

	var commandNames []string
	for _, subcmd := range subcommands {
		commandNames = append(commandNames, subcmd.Name())
	}

	assert.Contains(t, commandNames, "list")
	assert.Contains(t, commandNames, "get")
	assert.Contains(t, commandNames, "purge-and-reseed")
}

func TestAppUsageEventsListCommand(t *testing.T) {
	cmd := newAppUsageEventsListCommand()
	assert.Equal(t, "list", cmd.Use)
	assert.Equal(t, "List application usage events", cmd.Short)
	assert.Equal(t, "List application usage events with optional filtering", cmd.Long)
	assert.NotNil(t, cmd.RunE)

	// Check all filtering flags exist
	flags := []string{
		"all", "per-page", "after-guid", "app-name",
		"space-name", "org-name", "start-time", "end-time",
	}

	for _, flagName := range flags {
		flag := cmd.Flags().Lookup(flagName)
		assert.NotNil(t, flag, "Flag %s should exist", flagName)
	}

	// Check default values
	perPageFlag := cmd.Flags().Lookup("per-page")
	assert.Equal(t, "50", perPageFlag.DefValue)

	allFlag := cmd.Flags().Lookup("all")
	assert.Equal(t, "false", allFlag.DefValue)
}

func TestAppUsageEventsGetCommand(t *testing.T) {
	cmd := newAppUsageEventsGetCommand()
	assert.Equal(t, "get EVENT_GUID", cmd.Use)
	assert.Equal(t, "Get app usage event details", cmd.Short)
	assert.Equal(t, "Display detailed information about a specific app usage event", cmd.Long)
	assert.NotNil(t, cmd.RunE)
	assert.NotNil(t, cmd.Args)
}

func TestAppUsageEventsPurgeReseedCommand(t *testing.T) {
	cmd := newAppUsageEventsPurgeReseedCommand()
	assert.Equal(t, "purge-and-reseed", cmd.Use)
	assert.Equal(t, "Purge and reseed app usage events", cmd.Short)
	assert.Equal(t, "Purge existing app usage events and reseed with current state", cmd.Long)
	assert.NotNil(t, cmd.RunE)

	// Check force flag
	forceFlag := cmd.Flags().Lookup("force")
	assert.NotNil(t, forceFlag)
	assert.Equal(t, "f", forceFlag.Shorthand)
	assert.Equal(t, "false", forceFlag.DefValue)
}
