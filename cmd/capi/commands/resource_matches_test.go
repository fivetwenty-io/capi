package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewResourceMatchesCommand(t *testing.T) {
	cmd := NewResourceMatchesCommand()
	assert.Equal(t, "resource-matches", cmd.Use)
	assert.Equal(t, []string{"resources", "resource", "rm"}, cmd.Aliases)
	assert.Equal(t, "Manage resource matches", cmd.Short)
	assert.Equal(t, "Create resource matches for optimizing package uploads", cmd.Long)

	// Check subcommands are added
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 1)
	
	var commandNames []string
	for _, subcmd := range subcommands {
		commandNames = append(commandNames, subcmd.Name())
	}
	
	assert.Contains(t, commandNames, "create")
}

func TestResourceMatchesCreateCommand(t *testing.T) {
	cmd := newResourceMatchesCreateCommand()
	assert.Equal(t, "create", cmd.Use)
	assert.Equal(t, "Create resource matches", cmd.Short)
	assert.Equal(t, "Create resource matches to check which resources already exist on the platform", cmd.Long)
	assert.NotNil(t, cmd.RunE)
}