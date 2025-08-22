package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// validateFilePath validates that a file path is safe to read
func validateFilePathUsers(filePath string) error {
	// Clean the path to resolve any path traversal attempts
	cleanPath := filepath.Clean(filePath)

	// Check for path traversal attempts
	if filepath.IsAbs(filePath) {
		// Allow absolute paths but ensure they're clean
		if cleanPath != filePath {
			return fmt.Errorf("invalid file path: potential path traversal attempt")
		}
	} else {
		// For relative paths, ensure they don't escape the current directory
		if len(cleanPath) > 0 && cleanPath[0] == '.' && len(cleanPath) > 1 && cleanPath[1] == '.' {
			return fmt.Errorf("invalid file path: path traversal not allowed")
		}
	}

	// Check if file exists and is readable
	if _, err := os.Stat(cleanPath); err != nil {
		return fmt.Errorf("file not accessible: %w", err)
	}

	return nil
}

// createUsersBatchImportCommand creates the batch import command
func createUsersBatchImportCommand() *cobra.Command {
	var inputFile string
	var parallel bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "batch-import",
		Short: "Import users in batch from JSON file",
		Long: `Import multiple users from a JSON file in batch mode.
		
The JSON file should contain an array of user objects with the same structure
as individual user creation. This command supports both sequential and parallel
processing for improved performance.`,
		Example: `  # Import users from file (sequential)
  capi uaa batch-import --file users.json

  # Import users in parallel for better performance
  capi uaa batch-import --file users.json --parallel

  # Dry run to validate without creating users
  capi uaa batch-import --file users.json --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			config := loadConfig()

			if config.UAAEndpoint == "" {
				return fmt.Errorf("no UAA endpoint configured. Use 'capi uaa target <url>' to set one")
			}

			// Read input file
			var inputData []byte
			var err error

			if inputFile == "" || inputFile == "-" {
				// Read from stdin
				inputData, err = io.ReadAll(os.Stdin)
			} else {
				// Validate file path before reading
				if err := validateFilePathUsers(inputFile); err != nil {
					return fmt.Errorf("invalid input file: %w", err)
				}
				// Read from file
				inputData, err = os.ReadFile(filepath.Clean(inputFile))
			}

			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}

			if dryRun {
				// Validate JSON without creating users
				var users []interface{}
				if err := json.Unmarshal(inputData, &users); err != nil {
					return fmt.Errorf("invalid JSON format: %w", err)
				}
				fmt.Printf("Validation successful: %d users ready for import\n", len(users))
				return nil
			}

			// Create UAA client
			uaaClient, err := NewUAAClient(config)
			if err != nil {
				return fmt.Errorf("failed to create UAA client: %w", err)
			}

			if !uaaClient.IsAuthenticated() {
				return fmt.Errorf("not authenticated. Use a token command to authenticate first")
			}

			// Perform batch import
			results, err := BulkUserImport(uaaClient, inputData, parallel)
			if err != nil {
				return fmt.Errorf("failed to import users: %w", err)
			}

			// Display results
			successful := 0
			failed := 0

			for _, result := range results {
				if result.Error != nil {
					failed++
					fmt.Printf("❌ Failed to create user: %v\n", result.Error)
				} else {
					successful++
				}
			}

			fmt.Printf("\nBatch import completed:\n")
			fmt.Printf("  ✅ Successful: %d\n", successful)
			fmt.Printf("  ❌ Failed: %d\n", failed)
			fmt.Printf("  📊 Total: %d\n", len(results))

			return nil
		},
	}

	cmd.Flags().StringVar(&inputFile, "file", "", "Input JSON file (use '-' for stdin)")
	cmd.Flags().BoolVar(&parallel, "parallel", false, "Process imports in parallel")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Validate input without creating users")

	return cmd
}

// createUsersPerformanceCommand creates the performance metrics command
func createUsersPerformanceCommand() *cobra.Command {
	var reset bool

	cmd := &cobra.Command{
		Use:   "performance",
		Short: "Display performance metrics",
		Long: `Display performance metrics for UAA operations including cache hit rates,
operation timings, and efficiency statistics.`,
		Example: `  # Show current performance metrics
  capi uaa performance

  # Show metrics in JSON format
  capi uaa performance --output json

  # Reset performance counters
  capi uaa performance --reset`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if reset {
				// Reset performance metrics
				performanceMetrics = &PerformanceMetrics{
					operations: make(map[string][]time.Duration),
				}
				globalCache.Clear()
				fmt.Println("Performance metrics and cache cleared")
				return nil
			}

			// Get current metrics
			metrics := performanceMetrics.GetMetrics()

			// Display metrics based on output format
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(metrics)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(metrics)
			default:
				return displayPerformanceMetrics(metrics)
			}
		},
	}

	cmd.Flags().BoolVar(&reset, "reset", false, "Reset performance counters and cache")

	return cmd
}

// createUsersCacheCommand creates the cache management command
func createUsersCacheCommand() *cobra.Command {
	var clear bool
	var stats bool

	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage UAA operation cache",
		Long: `Manage the UAA operation cache for improved performance.
		
The cache stores frequently accessed data like user lookups, group information,
and server info to reduce API calls and improve response times.`,
		Example: `  # Show cache statistics
  capi uaa cache --stats

  # Clear all cached data
  capi uaa cache --clear`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if clear {
				globalCache.Clear()
				fmt.Println("Cache cleared successfully")
				return nil
			}

			if stats {
				metrics := performanceMetrics.GetMetrics()

				fmt.Printf("Cache Statistics:\n")
				fmt.Printf("  Cache Hits: %v\n", metrics["cache_hits"])
				fmt.Printf("  Cache Misses: %v\n", metrics["cache_misses"])
				fmt.Printf("  Hit Rate: %.2f%%\n", metrics["cache_hit_rate"])
				fmt.Printf("  Total Operations: %v\n", metrics["total_operations"])

				return nil
			}

			// Default: show cache status
			fmt.Println("UAA Cache Status:")
			fmt.Println("  Status: Active")
			fmt.Println("  TTL: 10 minutes")
			fmt.Println("  Auto-cleanup: Enabled")
			fmt.Println("\nUse --stats to see detailed statistics")
			fmt.Println("Use --clear to clear cached data")

			return nil
		},
	}

	cmd.Flags().BoolVar(&clear, "clear", false, "Clear all cached data")
	cmd.Flags().BoolVar(&stats, "stats", false, "Show cache statistics")

	return cmd
}

// Helper function to display performance metrics in table format
func displayPerformanceMetrics(metrics map[string]interface{}) error {
	fmt.Println("UAA Performance Metrics:")
	fmt.Println("========================")

	fmt.Printf("Total Operations: %v\n", metrics["total_operations"])
	fmt.Printf("Cache Hits: %v\n", metrics["cache_hits"])
	fmt.Printf("Cache Misses: %v\n", metrics["cache_misses"])
	fmt.Printf("Cache Hit Rate: %.2f%%\n", metrics["cache_hit_rate"])

	fmt.Println("\nOperation Statistics:")
	fmt.Println("--------------------")

	if operations, ok := metrics["operations"].(map[string]interface{}); ok {
		for op, stats := range operations {
			if opStats, ok := stats.(map[string]interface{}); ok {
				fmt.Printf("\n%s:\n", op)
				fmt.Printf("  Count: %v\n", opStats["count"])
				fmt.Printf("  Average: %v\n", opStats["average"])
				fmt.Printf("  Min: %v\n", opStats["min"])
				fmt.Printf("  Max: %v\n", opStats["max"])
			}
		}
	}

	return nil
}
