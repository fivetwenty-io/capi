package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAuditEventsCommand(t *testing.T) {
	cmd := NewAuditEventsCommand()
	assert.Equal(t, "audit-events", cmd.Use)
	assert.Equal(t, []string{"audit", "events", "ae"}, cmd.Aliases)
	assert.Equal(t, "Manage audit events", cmd.Short)
	assert.Equal(t, "View audit events for tracking system changes and user actions", cmd.Long)

	// Check subcommands are added
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 2)
	
	var commandNames []string
	for _, subcmd := range subcommands {
		commandNames = append(commandNames, subcmd.Name())
	}
	
	assert.Contains(t, commandNames, "list")
	assert.Contains(t, commandNames, "get")
}

func TestAuditEventsListCommand(t *testing.T) {
	cmd := newAuditEventsListCommand()
	assert.Equal(t, "list", cmd.Use)
	assert.Equal(t, "List audit events", cmd.Short)
	assert.Equal(t, "List audit events with optional filtering", cmd.Long)
	assert.NotNil(t, cmd.RunE)

	// Check filtering flags
	flags := []string{
		"all", "per-page", "event-types", "target-types", "actor-types", 
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

func TestAuditEventsGetCommand(t *testing.T) {
	cmd := newAuditEventsGetCommand()
	assert.Equal(t, "get EVENT_GUID", cmd.Use)
	assert.Equal(t, "Get audit event details", cmd.Short)
	assert.Equal(t, "Display detailed information about a specific audit event", cmd.Long)
	assert.NotNil(t, cmd.RunE)
	assert.NotNil(t, cmd.Args)
}