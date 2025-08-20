package commands

import (
	"github.com/spf13/cobra"
)

// NewDomainsCommand creates the domains command group
func NewDomainsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "domains",
		Aliases: []string{"domain"},
		Short:   "Manage domains",
		Long:    "List and manage Cloud Foundry domains",
	}

	// TODO: Add subcommands for domain management
	cmd.Run = func(cmd *cobra.Command, args []string) {
		cmd.Help()
	}

	return cmd
}
