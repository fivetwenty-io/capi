package commands

import (
	"github.com/spf13/cobra"
)

// NewUAAIntegrationCommand creates the integration sub-command group.
func NewUAAIntegrationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "integration",
		Short: "Integration and compatibility utilities",
		Long: `Integration and compatibility utilities for UAA.

This command group provides utilities for:
- Cloud Foundry integration
- Compatibility testing
- External system integration
- Migration and compatibility checks`,
		Example: `  # Check Cloud Foundry integration
  capi uaa integration cf --check-endpoints

  # Run compatibility tests
  capi uaa integration compatibility --version 4.30.0`,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	// Add integration commands with new naming
	cmd.AddCommand(createIntegrationCompatibilityCommand())
	cmd.AddCommand(createIntegrationCFCommand())

	return cmd
}

func createIntegrationCompatibilityCommand() *cobra.Command {
	cmd := createUsersCompatibilityCommand()
	cmd.Use = "compatibility"

	return cmd
}

func createIntegrationCFCommand() *cobra.Command {
	cmd := createUsersCFIntegrationCommand()
	cmd.Use = "cf"

	return cmd
}
