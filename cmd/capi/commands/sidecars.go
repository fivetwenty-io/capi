package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fivetwenty-io/capi/v3/pkg/capi"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// NewSidecarsCommand creates the sidecars command group
func NewSidecarsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sidecars",
		Aliases: []string{"sidecar", "sc"},
		Short:   "Manage application sidecars",
		Long:    "View, update, and delete sidecars for applications",
	}

	cmd.AddCommand(newSidecarsGetCommand())
	cmd.AddCommand(newSidecarsUpdateCommand())
	cmd.AddCommand(newSidecarsDeleteCommand())
	cmd.AddCommand(newSidecarsListForProcessCommand())

	return cmd
}

func newSidecarsGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get SIDECAR_GUID",
		Short: "Get sidecar details",
		Long:  "Display detailed information about a specific sidecar",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sidecarGUID := args[0]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			sidecar, err := client.Sidecars().Get(ctx, sidecarGUID)
			if err != nil {
				return fmt.Errorf("failed to get sidecar: %w", err)
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(sidecar)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(sidecar)
			default:
				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Property", "Value")

				_ = table.Append("Name", sidecar.Name)
				_ = table.Append("GUID", sidecar.GUID)
				_ = table.Append("Command", sidecar.Command)
				_ = table.Append("Process Types", strings.Join(sidecar.ProcessTypes, ", "))

				if sidecar.MemoryInMB != nil {
					_ = table.Append("Memory", fmt.Sprintf("%d MB", *sidecar.MemoryInMB))
				} else {
					_ = table.Append("Memory", "default")
				}

				_ = table.Append("Origin", sidecar.Origin)
				_ = table.Append("Created", sidecar.CreatedAt.Format("2006-01-02 15:04:05"))
				_ = table.Append("Updated", sidecar.UpdatedAt.Format("2006-01-02 15:04:05"))

				if sidecar.Relationships.App.Data != nil {
					_ = table.Append("App GUID", sidecar.Relationships.App.Data.GUID)
				}

				fmt.Printf("Sidecar details:\n\n")
				_ = table.Render()
			}

			return nil
		},
	}
}

func newSidecarsUpdateCommand() *cobra.Command {
	var (
		name         string
		command      string
		processTypes []string
		memoryInMB   int
	)

	cmd := &cobra.Command{
		Use:   "update SIDECAR_GUID",
		Short: "Update a sidecar",
		Long:  "Update an existing sidecar configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sidecarGUID := args[0]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Build update request
			updateReq := &capi.SidecarUpdateRequest{}

			if name != "" {
				updateReq.Name = &name
			}

			if command != "" {
				updateReq.Command = &command
			}

			if len(processTypes) > 0 {
				updateReq.ProcessTypes = processTypes
			}

			if cmd.Flags().Changed("memory") {
				updateReq.MemoryInMB = &memoryInMB
			}

			// Update sidecar
			updatedSidecar, err := client.Sidecars().Update(ctx, sidecarGUID, updateReq)
			if err != nil {
				return fmt.Errorf("failed to update sidecar: %w", err)
			}

			fmt.Printf("Successfully updated sidecar '%s'\n", updatedSidecar.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "sidecar name")
	cmd.Flags().StringVar(&command, "command", "", "sidecar command")
	cmd.Flags().StringSliceVar(&processTypes, "process-types", nil, "process types (comma-separated)")
	cmd.Flags().IntVar(&memoryInMB, "memory", 0, "memory limit in MB")

	return cmd
}

func newSidecarsDeleteCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete SIDECAR_GUID",
		Short: "Delete a sidecar",
		Long:  "Delete a sidecar from an application",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sidecarGUID := args[0]

			if !force {
				fmt.Printf("Really delete sidecar '%s'? (y/N): ", sidecarGUID)
				var response string
				_, _ = fmt.Scanln(&response)
				if response != "y" && response != "Y" {
					fmt.Println("Cancelled")
					return nil
				}
			}

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Get sidecar name for confirmation
			sidecar, err := client.Sidecars().Get(ctx, sidecarGUID)
			if err != nil {
				return fmt.Errorf("failed to get sidecar: %w", err)
			}

			// Delete sidecar
			err = client.Sidecars().Delete(ctx, sidecarGUID)
			if err != nil {
				return fmt.Errorf("failed to delete sidecar: %w", err)
			}

			fmt.Printf("Successfully deleted sidecar '%s'\n", sidecar.Name)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "force deletion without confirmation")

	return cmd
}

func newSidecarsListForProcessCommand() *cobra.Command {
	var (
		allPages bool
		perPage  int
	)

	cmd := &cobra.Command{
		Use:   "list-for-process PROCESS_GUID",
		Short: "List sidecars for a process",
		Long:  "List all sidecars associated with a specific process",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			processGUID := args[0]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			params := capi.NewQueryParams()
			params.PerPage = perPage

			sidecars, err := client.Sidecars().ListForProcess(ctx, processGUID, params)
			if err != nil {
				return fmt.Errorf("failed to list sidecars for process: %w", err)
			}

			// Fetch all pages if requested
			allSidecars := sidecars.Resources
			if allPages && sidecars.Pagination.TotalPages > 1 {
				for page := 2; page <= sidecars.Pagination.TotalPages; page++ {
					params.Page = page
					moreSidecars, err := client.Sidecars().ListForProcess(ctx, processGUID, params)
					if err != nil {
						return fmt.Errorf("failed to fetch page %d: %w", page, err)
					}
					allSidecars = append(allSidecars, moreSidecars.Resources...)
				}
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(allSidecars)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(allSidecars)
			default:
				if len(allSidecars) == 0 {
					fmt.Printf("No sidecars found for process %s\n", processGUID)
					return nil
				}

				fmt.Printf("Sidecars for process %s:\n\n", processGUID)
				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Name", "GUID", "Command", "Process Types", "Memory", "Origin", "Created")

				for _, sidecar := range allSidecars {
					memoryStr := "default"
					if sidecar.MemoryInMB != nil {
						memoryStr = fmt.Sprintf("%d MB", *sidecar.MemoryInMB)
					}

					processTypesStr := strings.Join(sidecar.ProcessTypes, ", ")
					if len(processTypesStr) > 50 {
						processTypesStr = processTypesStr[:47] + "..."
					}

					commandStr := sidecar.Command
					if len(commandStr) > 40 {
						commandStr = commandStr[:37] + "..."
					}

					_ = table.Append(
						sidecar.Name,
						sidecar.GUID,
						commandStr,
						processTypesStr,
						memoryStr,
						sidecar.Origin,
						sidecar.CreatedAt.Format("2006-01-02"),
					)
				}

				_ = table.Render()

				if !allPages && sidecars.Pagination.TotalPages > 1 {
					fmt.Printf("\nShowing page 1 of %d. Use --all to fetch all pages.\n", sidecars.Pagination.TotalPages)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&allPages, "all", false, "fetch all pages")
	cmd.Flags().IntVar(&perPage, "per-page", 50, "results per page")

	return cmd
}
