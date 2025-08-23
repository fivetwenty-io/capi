package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRevisionsCommand(t *testing.T) {
	cmd := NewRevisionsCommand()
	assert.Equal(t, "revisions", cmd.Use)
	assert.Equal(t, []string{"revision", "rev"}, cmd.Aliases)
	assert.Equal(t, "Manage application revisions", cmd.Short)
	assert.Equal(t, "View and manage application revisions", cmd.Long)

	// Check subcommands are added
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 3)

	var commandNames []string
	for _, subcmd := range subcommands {
		commandNames = append(commandNames, subcmd.Name())
	}

	assert.Contains(t, commandNames, "get")
	assert.Contains(t, commandNames, "update")
	assert.Contains(t, commandNames, "get-env")
}

func TestRevisionsGetCommand(t *testing.T) {
	cmd := newRevisionsGetCommand()
	assert.Equal(t, "get REVISION_GUID", cmd.Use)
	assert.Equal(t, "Get revision details", cmd.Short)
	assert.Equal(t, "Display detailed information about a specific revision", cmd.Long)
	assert.NotNil(t, cmd.RunE)
	assert.NotNil(t, cmd.Args)
}

func TestRevisionsUpdateCommand(t *testing.T) {
	cmd := newRevisionsUpdateCommand()
	assert.Equal(t, "update REVISION_GUID", cmd.Use)
	assert.Equal(t, "Update a revision", cmd.Short)
	assert.Equal(t, "Update a revision's metadata", cmd.Long)
	assert.NotNil(t, cmd.RunE)
	assert.NotNil(t, cmd.Args)

	// Check metadata flag
	metadataFlag := cmd.Flags().Lookup("metadata")
	assert.NotNil(t, metadataFlag)
}

func TestRevisionsGetEnvCommand(t *testing.T) {
	cmd := newRevisionsGetEnvCommand()
	assert.Equal(t, "get-env REVISION_GUID", cmd.Use)
	assert.Equal(t, "Get revision environment variables", cmd.Short)
	assert.Equal(t, "Display environment variables for a specific revision", cmd.Long)
	assert.NotNil(t, cmd.RunE)
	assert.NotNil(t, cmd.Args)
}
