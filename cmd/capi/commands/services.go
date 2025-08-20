package commands

import (
	"github.com/spf13/cobra"
)

// NewServicesCommand creates the services command group
func NewServicesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "services",
		Aliases: []string{"service"},
		Short:   "Manage services",
		Long:    "List and manage Cloud Foundry services and service instances",
	}

	// TODO: Add subcommands for service management
	cmd.Run = func(cmd *cobra.Command, args []string) {
		cmd.Help()
	}

	return cmd
}
