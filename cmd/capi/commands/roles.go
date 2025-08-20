package commands

import (
	"github.com/spf13/cobra"
)

// NewRolesCommand creates the roles command group
func NewRolesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "roles",
		Aliases: []string{"role"},
		Short:   "Manage roles",
		Long:    "List and manage Cloud Foundry user roles",
	}

	// TODO: Add subcommands for role management
	cmd.Run = func(cmd *cobra.Command, args []string) {
		cmd.Help()
	}

	return cmd
}
