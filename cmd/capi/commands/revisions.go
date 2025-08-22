package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/fivetwenty-io/capi-client/pkg/capi"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// NewRevisionsCommand creates the revisions command group
func NewRevisionsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "revisions",
		Aliases: []string{"revision", "rev"},
		Short:   "Manage application revisions",
		Long:    "View and manage application revisions",
	}

	cmd.AddCommand(newRevisionsGetCommand())
	cmd.AddCommand(newRevisionsUpdateCommand())
	cmd.AddCommand(newRevisionsGetEnvCommand())

	return cmd
}

func newRevisionsGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get REVISION_GUID",
		Short: "Get revision details",
		Long:  "Display detailed information about a specific revision",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			revisionGUID := args[0]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			revision, err := client.Revisions().Get(ctx, revisionGUID)
			if err != nil {
				return fmt.Errorf("failed to get revision: %w", err)
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(revision)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(revision)
			default:
				fmt.Printf("Revision: %d\n", revision.Version)
				fmt.Printf("  GUID: %s\n", revision.GUID)
				fmt.Printf("  Deployable: %t\n", revision.Deployable)
				if revision.Description != nil {
					fmt.Printf("  Description: %s\n", *revision.Description)
				}
				fmt.Printf("  Created: %s\n", revision.CreatedAt.Format("2006-01-02 15:04:05"))
				fmt.Printf("  Updated: %s\n", revision.UpdatedAt.Format("2006-01-02 15:04:05"))

				fmt.Printf("  Droplet GUID: %s\n", revision.Droplet.GUID)

				if revision.Relationships.App.Data != nil {
					fmt.Printf("  App GUID: %s\n", revision.Relationships.App.Data.GUID)
				}

				if len(revision.Processes) > 0 {
					fmt.Println("  Processes:")
					for processType, process := range revision.Processes {
						fmt.Printf("    %s:\n", processType)
						fmt.Printf("      GUID: %s\n", process.GUID)
						fmt.Printf("      Type: %s\n", process.Type)
						fmt.Printf("      Instances: %d\n", process.Instances)
						fmt.Printf("      Memory: %d MB\n", process.MemoryInMB)
						fmt.Printf("      Disk: %d MB\n", process.DiskInMB)
						if process.Command != nil {
							fmt.Printf("      Command: %s\n", *process.Command)
						}
					}
				}

				if len(revision.Sidecars) > 0 {
					fmt.Println("  Sidecars:")
					for _, sidecar := range revision.Sidecars {
						fmt.Printf("    %s:\n", sidecar.Name)
						fmt.Printf("      GUID: %s\n", sidecar.GUID)
						fmt.Printf("      Command: %s\n", sidecar.Command)
						if sidecar.MemoryInMB != nil {
							fmt.Printf("      Memory: %d MB\n", *sidecar.MemoryInMB)
						}
					}
				}
			}

			return nil
		},
	}
}

func newRevisionsUpdateCommand() *cobra.Command {
	var (
		metadata map[string]string
	)

	cmd := &cobra.Command{
		Use:   "update REVISION_GUID",
		Short: "Update a revision",
		Long:  "Update a revision's metadata",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			revisionGUID := args[0]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Build update request
			updateReq := &capi.RevisionUpdateRequest{}

			if len(metadata) > 0 {
				updateReq.Metadata = &capi.Metadata{
					Labels: metadata,
				}
			}

			// Update revision
			updatedRevision, err := client.Revisions().Update(ctx, revisionGUID, updateReq)
			if err != nil {
				return fmt.Errorf("failed to update revision: %w", err)
			}

			fmt.Printf("Successfully updated revision %d\n", updatedRevision.Version)
			return nil
		},
	}

	cmd.Flags().StringToStringVar(&metadata, "metadata", nil, "metadata labels to apply (key=value)")

	return cmd
}

func newRevisionsGetEnvCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get-env REVISION_GUID",
		Short: "Get revision environment variables",
		Long:  "Display environment variables for a specific revision",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			revisionGUID := args[0]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			envVars, err := client.Revisions().GetEnvironmentVariables(ctx, revisionGUID)
			if err != nil {
				return fmt.Errorf("failed to get revision environment variables: %w", err)
			}

			// Collect environment variables for structured output
			type EnvVar struct {
				Name  string      `json:"name" yaml:"name"`
				Value interface{} `json:"value" yaml:"value"`
			}

			var envVarsList []EnvVar
			for key, value := range envVars {
				envVarsList = append(envVarsList, EnvVar{
					Name:  key,
					Value: value,
				})
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(envVarsList)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(envVarsList)
			default:
				if len(envVarsList) == 0 {
					fmt.Printf("No environment variables found for revision %s\n", revisionGUID)
					return nil
				}

				fmt.Printf("Environment variables for revision %s:\n\n", revisionGUID)
				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Name", "Value")

				for _, envVar := range envVarsList {
					valueStr := fmt.Sprintf("%v", envVar.Value)
					// Truncate long values for table display
					if len(valueStr) > 80 {
						valueStr = valueStr[:77] + "..."
					}
					_ = table.Append(envVar.Name, valueStr)
				}

				_ = table.Render()
			}

			return nil
		},
	}
}
