package commands

import (
	"github.com/spf13/cobra"
)

// NewBuildpacksCommand creates the buildpacks command group
func NewBuildpacksCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "buildpacks",
		Aliases: []string{"buildpack"},
		Short:   "Manage buildpacks",
		Long:    "List and manage Cloud Foundry buildpacks",
	}

	// TODO: Add subcommands for buildpack management
	cmd.Run = func(cmd *cobra.Command, args []string) {
		cmd.Help()
	}

	return cmd
}
