package commands

import (
	"github.com/spf13/cobra"
)

// NewUsersCommand creates the users command group
func NewUsersCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "users",
		Aliases: []string{"user"},
		Short:   "Manage users",
		Long:    "List and manage Cloud Foundry users",
	}

	// TODO: Add subcommands for user management
	cmd.Run = func(cmd *cobra.Command, args []string) {
		cmd.Help()
	}

	return cmd
}
