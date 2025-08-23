package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEnvVarGroupsCommand(t *testing.T) {
	cmd := NewEnvVarGroupsCommand()
	assert.Equal(t, "env-var-groups", cmd.Use)
	assert.Equal(t, []string{"environment-variable-groups", "env-groups", "evg"}, cmd.Aliases)
	assert.Equal(t, "Manage environment variable groups", cmd.Short)
	assert.Equal(t, "View and update environment variable groups (running and staging)", cmd.Long)

	// Check subcommands are added
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 2)

	var commandNames []string
	for _, subcmd := range subcommands {
		commandNames = append(commandNames, subcmd.Name())
	}

	assert.Contains(t, commandNames, "get")
	assert.Contains(t, commandNames, "update")
}

func TestEnvVarGroupsGetCommand(t *testing.T) {
	cmd := newEnvVarGroupsGetCommand()
	assert.Equal(t, "get GROUP_NAME", cmd.Use)
	assert.Equal(t, "Get environment variable group", cmd.Short)
	assert.Equal(t, "Display environment variables for a specific group (running or staging)", cmd.Long)
	assert.NotNil(t, cmd.RunE)
	assert.NotNil(t, cmd.Args)
}

func TestEnvVarGroupsUpdateCommand(t *testing.T) {
	cmd := newEnvVarGroupsUpdateCommand()
	assert.Equal(t, "update GROUP_NAME", cmd.Use)
	assert.Equal(t, "Update environment variable group", cmd.Short)
	assert.Equal(t, "Update environment variables for a specific group (running or staging)", cmd.Long)
	assert.NotNil(t, cmd.RunE)

	assert.NotNil(t, cmd.Args)
}
