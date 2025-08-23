package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// NewManifestsCommand creates the manifests command group
func NewManifestsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "manifests",
		Short: "Manage application manifests",
		Long:  "Manage Cloud Foundry application manifests including applying, generating, and diffing manifests",
	}

	cmd.AddCommand(newManifestsApplyCommand())
	cmd.AddCommand(newManifestsGenerateCommand())
	cmd.AddCommand(newManifestsDiffCommand())

	return cmd
}

func newManifestsApplyCommand() *cobra.Command {
	var (
		manifestPath string
		wait         bool
	)

	cmd := &cobra.Command{
		Use:   "apply SPACE_GUID",
		Short: "Apply a manifest to a space",
		Long:  "Apply an application manifest to deploy or update applications in a space",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spaceGUID := args[0]

			// Read manifest file with path cleaning
			cleanPath := filepath.Clean(manifestPath)
			if !filepath.IsAbs(cleanPath) {
				// Make relative paths relative to current directory
				wd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get working directory: %w", err)
				}
				cleanPath = filepath.Join(wd, cleanPath)
			}
			manifestContent, err := os.ReadFile(cleanPath)
			if err != nil {
				return fmt.Errorf("failed to read manifest file: %w", err)
			}

			// Create client
			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}

			ctx := context.Background()

			// Apply manifest
			job, err := client.Manifests().ApplyManifest(ctx, spaceGUID, manifestContent)
			if err != nil {
				return fmt.Errorf("failed to apply manifest: %w", err)
			}

			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(job)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(job)
			default:
				fmt.Printf("Manifest application started\n")
				fmt.Printf("Job ID: %s\n", job.GUID)
				fmt.Printf("State: %s\n", job.State)

				if wait {
					fmt.Println("\nWaiting for job to complete...")
					completedJob, err := client.Jobs().PollUntilComplete(ctx, job.GUID)
					if err != nil {
						return fmt.Errorf("failed to wait for job: %w", err)
					}

					if completedJob.State == "COMPLETE" {
						fmt.Println("✓ Manifest applied successfully")
					} else {
						fmt.Printf("Job finished with state: %s\n", completedJob.State)
						if len(completedJob.Errors) > 0 {
							fmt.Println("\nErrors:")
							for _, e := range completedJob.Errors {
								fmt.Printf("  - %s: %s\n", e.Title, e.Detail)
							}
						}
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&manifestPath, "file", "f", "manifest.yml", "Path to manifest file")
	cmd.Flags().BoolVarP(&wait, "wait", "w", false, "Wait for the job to complete")

	return cmd
}

func newManifestsGenerateCommand() *cobra.Command {
	var outputFile string

	cmd := &cobra.Command{
		Use:   "generate APP_GUID",
		Short: "Generate a manifest for an app",
		Long:  "Generate a manifest file from an existing application's current configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			appGUID := args[0]

			// Create client
			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}

			ctx := context.Background()

			// Generate manifest
			manifestContent, err := client.Manifests().GenerateManifest(ctx, appGUID)
			if err != nil {
				return fmt.Errorf("failed to generate manifest: %w", err)
			}

			// Write to file or stdout
			if outputFile != "" {
				err = os.WriteFile(outputFile, manifestContent, 0600)
				if err != nil {
					return fmt.Errorf("failed to write manifest file: %w", err)
				}
				fmt.Printf("Manifest written to %s\n", outputFile)
			} else {
				fmt.Print(string(manifestContent))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFile, "output-file", "o", "", "Output file path (defaults to stdout)")

	return cmd
}

func newManifestsDiffCommand() *cobra.Command {
	var manifestPath string

	cmd := &cobra.Command{
		Use:   "diff SPACE_GUID",
		Short: "Create a diff between current and proposed manifest",
		Long:  "Compare the current state of applications in a space with a proposed manifest to see what would change",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spaceGUID := args[0]

			// Read manifest file with path cleaning
			cleanPath := filepath.Clean(manifestPath)
			if !filepath.IsAbs(cleanPath) {
				// Make relative paths relative to current directory
				wd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get working directory: %w", err)
				}
				cleanPath = filepath.Join(wd, cleanPath)
			}
			manifestContent, err := os.ReadFile(cleanPath)
			if err != nil {
				return fmt.Errorf("failed to read manifest file: %w", err)
			}

			// Create client
			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}

			ctx := context.Background()

			// Create diff
			diff, err := client.Manifests().CreateManifestDiff(ctx, spaceGUID, manifestContent)
			if err != nil {
				return fmt.Errorf("failed to create manifest diff: %w", err)
			}

			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(diff)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(diff)
			default:
				if len(diff.Diff) == 0 {
					fmt.Println("No differences found")
				} else {
					fmt.Printf("Found %d difference(s):\n\n", len(diff.Diff))

					table := tablewriter.NewWriter(os.Stdout)
					table.Header("Operation", "Path", "Current Value", "New Value")

					for _, entry := range diff.Diff {
						wasStr := formatValue(entry.Was)
						valueStr := formatValue(entry.Value)
						_ = table.Append(entry.Op, entry.Path, wasStr, valueStr)
					}

					if err := table.Render(); err != nil {
						return fmt.Errorf("failed to render table: %w", err)
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&manifestPath, "file", "f", "manifest.yml", "Path to manifest file")

	return cmd
}

// formatValue formats a value for display in the diff table
func formatValue(v interface{}) string {
	if v == nil {
		return "-"
	}

	switch val := v.(type) {
	case string:
		return val
	case bool:
		return fmt.Sprintf("%t", val)
	case float64:
		// Check if it's an integer
		if val == float64(int(val)) {
			return fmt.Sprintf("%d", int(val))
		}
		return fmt.Sprintf("%g", val)
	default:
		// For complex types, use JSON
		b, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprintf("%v", val)
		}
		return string(b)
	}
}
