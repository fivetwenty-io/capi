package commands

import (
	"github.com/spf13/cobra"
)

// NewUAABatchCommand creates the batch sub-command group.
func NewUAABatchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batch",
		Short: "Batch operations and performance utilities",
		Long: `Batch operations and performance utilities for UAA.

This command group provides utilities for:
- Bulk import operations
- Performance testing and monitoring
- Caching operations
- Batch processing of UAA resources`,
		Example: `  # Import users from file
  capi uaa batch import --file users.json

  # Run performance tests
  capi uaa batch performance --users 1000 --concurrent 10

  # Cache management
  capi uaa batch cache --clear`,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	// Add batch operation commands with new naming
	cmd.AddCommand(createBatchImportCommand())
	cmd.AddCommand(createBatchPerformanceCommand())
	cmd.AddCommand(createBatchCacheCommand())

	return cmd
}

func createBatchImportCommand() *cobra.Command {
	cmd := createUsersBatchImportCommand()
	cmd.Use = "import"

	return cmd
}

func createBatchPerformanceCommand() *cobra.Command {
	cmd := createUsersPerformanceCommand()
	cmd.Use = "performance"

	return cmd
}

func createBatchCacheCommand() *cobra.Command {
	cmd := createUsersCacheCommand()
	cmd.Use = "cache"

	return cmd
}
