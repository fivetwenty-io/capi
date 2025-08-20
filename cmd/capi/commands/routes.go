package commands

import (
	"github.com/spf13/cobra"
)

// NewRoutesCommand creates the routes command group
func NewRoutesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "routes",
		Aliases: []string{"route"},
		Short:   "Manage routes",
		Long:    "List and manage Cloud Foundry routes",
	}

	// TODO: Add subcommands for route management
	cmd.Run = func(cmd *cobra.Command, args []string) {
		cmd.Help()
	}

	return cmd
}
