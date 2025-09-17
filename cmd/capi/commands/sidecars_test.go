package commands_test

import (
	"testing"

	"github.com/fivetwenty-io/capi/v3/cmd/capi/commands"
	"github.com/stretchr/testify/assert"
)

func TestNewSidecarsCommand(t *testing.T) {
	t.Parallel()

	cmd := commands.NewSidecarsCommand()
	assert.Equal(t, "sidecars", cmd.Use)
	assert.Equal(t, []string{"sidecar", "sc"}, cmd.Aliases)
	assert.Equal(t, "Manage application sidecars", cmd.Short)
	assert.Equal(t, "View, update, and delete sidecars for applications", cmd.Long)

	// Check subcommands are added
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 4)

	commandNames := make([]string, 0, len(subcommands))
	for _, subcmd := range subcommands {
		commandNames = append(commandNames, subcmd.Name())
	}

	assert.Contains(t, commandNames, "get")
	assert.Contains(t, commandNames, "update")
	assert.Contains(t, commandNames, "delete")
	assert.Contains(t, commandNames, "list-for-process")
}

// Note: Tests for unexported functions (newSidecarsGetCommand, etc.)
// are not included since they cannot be accessed from the commands_test package.
// These functions are tested indirectly through the main command.
