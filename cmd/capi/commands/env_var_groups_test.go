package commands_test

import (
	"testing"

	"github.com/fivetwenty-io/capi/v3/cmd/capi/commands"
	"github.com/stretchr/testify/assert"
)

func TestNewEnvVarGroupsCommand(t *testing.T) {
	t.Parallel()

	cmd := commands.NewEnvVarGroupsCommand()
	assert.Equal(t, "env-var-groups", cmd.Use)
	assert.Equal(t, []string{"environment-variable-groups", "env-groups", "evg"}, cmd.Aliases)
	assert.Equal(t, "Manage environment variable groups", cmd.Short)
	assert.Equal(t, "View and update environment variable groups (running and staging)", cmd.Long)

	// Check subcommands are added
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 2)

	commandNames := make([]string, 0, len(subcommands))
	for _, subcmd := range subcommands {
		commandNames = append(commandNames, subcmd.Name())
	}

	assert.Contains(t, commandNames, "get")
	assert.Contains(t, commandNames, "update")
}

// Note: Tests for unexported functions (newEnvVarGroupsGetCommand, newEnvVarGroupsUpdateCommand)
// are not included since they cannot be accessed from the commands_test package.
// These functions are tested indirectly through the main command.
