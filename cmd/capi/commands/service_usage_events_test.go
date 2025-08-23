package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServiceUsageEventsCommand(t *testing.T) {
	cmd := NewServiceUsageEventsCommand()
	assert.Equal(t, "service-usage-events", cmd.Use)
	assert.Equal(t, []string{"service-usage", "service-events", "sue"}, cmd.Aliases)
	assert.Equal(t, "Manage service usage events", cmd.Short)
	assert.Equal(t, "View and manage service usage events for monitoring and billing", cmd.Long)

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

func TestServiceUsageEventsListCommand(t *testing.T) {
	cmd := newServiceUsageEventsListCommand()
	assert.Equal(t, "list", cmd.Use)
	assert.Equal(t, "List service usage events", cmd.Short)
	assert.Equal(t, "List service usage events with optional filtering", cmd.Long)
	assert.NotNil(t, cmd.RunE)

	// Check filtering flags
	flags := []string{
		"all", "per-page", "after-guid", "service-instance-name",
		"service-offering-name", "service-broker-name", "space-name",
		"org-name", "start-time", "end-time",
	}

	for _, flagName := range flags {
		flag := cmd.Flags().Lookup(flagName)
		assert.NotNil(t, flag, "Flag %s should exist", flagName)
	}

	// Check default values
	perPageFlag := cmd.Flags().Lookup("per-page")
	assert.Equal(t, "50", perPageFlag.DefValue)
}

func TestServiceUsageEventsGetCommand(t *testing.T) {
	cmd := newServiceUsageEventsGetCommand()
	assert.Equal(t, "get EVENT_GUID", cmd.Use)
	assert.Equal(t, "Get service usage event details", cmd.Short)
	assert.Equal(t, "Display detailed information about a specific service usage event", cmd.Long)
	assert.NotNil(t, cmd.RunE)
	assert.NotNil(t, cmd.Args)
}

func TestServiceUsageEventsPurgeReseedCommand(t *testing.T) {
	cmd := newServiceUsageEventsPurgeReseedCommand()
	assert.Equal(t, "purge-and-reseed", cmd.Use)
	assert.Equal(t, "Purge and reseed service usage events", cmd.Short)
	assert.Equal(t, "Purge existing service usage events and reseed with current state", cmd.Long)
	assert.NotNil(t, cmd.RunE)

	// Check force flag
	forceFlag := cmd.Flags().Lookup("force")
	assert.NotNil(t, forceFlag)
	assert.Equal(t, "f", forceFlag.Shorthand)
	assert.Equal(t, "false", forceFlag.DefValue)
}
