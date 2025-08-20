package commands

import (
	"github.com/spf13/cobra"
)

// NewStacksCommand creates the stacks command group
func NewStacksCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "stacks",
		Aliases: []string{"stack"},
		Short:   "Manage stacks",
		Long:    "List and manage Cloud Foundry stacks",
	}

	// TODO: Add subcommands for stack management
	cmd.Run = func(cmd *cobra.Command, args []string) {
		cmd.Help()
	}

	return cmd
}
