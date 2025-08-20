package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fivetwenty-io/capi-client/pkg/capi"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// NewBuildpacksCommand creates the buildpacks command group
func NewBuildpacksCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "buildpacks",
		Aliases: []string{"buildpack"},
		Short:   "Manage buildpacks",
		Long:    "List and manage Cloud Foundry buildpacks",
	}

	cmd.AddCommand(newBuildpacksListCommand())
	cmd.AddCommand(newBuildpacksGetCommand())
	cmd.AddCommand(newBuildpacksCreateCommand())
	cmd.AddCommand(newBuildpacksUpdateCommand())
	cmd.AddCommand(newBuildpacksDeleteCommand())
	cmd.AddCommand(newBuildpacksUploadCommand())

	return cmd
}

func newBuildpacksListCommand() *cobra.Command {
	var (
		allPages bool
		perPage  int
		enabled  bool
		stack    string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List buildpacks",
		Long:  "List all buildpacks",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			params := capi.NewQueryParams()
			params.PerPage = perPage

			// Filter by enabled state if specified
			if cmd.Flags().Changed("enabled") {
				if enabled {
					params.WithFilter("enabled", "true")
				} else {
					params.WithFilter("enabled", "false")
				}
			}

			// Filter by stack if specified
			if stack != "" {
				params.WithFilter("stacks", stack)
			}

			buildpacks, err := client.Buildpacks().List(ctx, params)
			if err != nil {
				return fmt.Errorf("failed to list buildpacks: %w", err)
			}

			// Fetch all pages if requested
			allBuildpacks := buildpacks.Resources
			if allPages && buildpacks.Pagination.TotalPages > 1 {
				for page := 2; page <= buildpacks.Pagination.TotalPages; page++ {
					params.Page = page
					moreBuildpacks, err := client.Buildpacks().List(ctx, params)
					if err != nil {
						return fmt.Errorf("failed to fetch page %d: %w", page, err)
					}
					allBuildpacks = append(allBuildpacks, moreBuildpacks.Resources...)
				}
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(allBuildpacks)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(allBuildpacks)
			default:
				if len(allBuildpacks) == 0 {
					fmt.Println("No buildpacks found")
					return nil
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Position", "Name", "Stack", "State", "Enabled", "Locked", "Filename")

				for _, bp := range allBuildpacks {
					stack := "any"
					if bp.Stack != nil {
						stack = *bp.Stack
					}

					enabled := "yes"
					if !bp.Enabled {
						enabled = "no"
					}

					locked := "no"
					if bp.Locked {
						locked = "yes"
					}

					filename := ""
					if bp.Filename != nil {
						filename = *bp.Filename
					}

					table.Append(fmt.Sprintf("%d", bp.Position), bp.Name, stack, bp.State, enabled, locked, filename)
				}

				table.Render()

				if !allPages && buildpacks.Pagination.TotalPages > 1 {
					fmt.Printf("\nShowing page 1 of %d. Use --all to fetch all pages.\n", buildpacks.Pagination.TotalPages)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&allPages, "all", false, "fetch all pages")
	cmd.Flags().IntVar(&perPage, "per-page", 50, "results per page")
	cmd.Flags().BoolVar(&enabled, "enabled", false, "filter by enabled buildpacks")
	cmd.Flags().StringVar(&stack, "stack", "", "filter by stack")

	return cmd
}

func newBuildpacksGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get BUILDPACK_NAME_OR_GUID",
		Short: "Get buildpack details",
		Long:  "Display detailed information about a specific buildpack",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Try to get by GUID first
			bp, err := client.Buildpacks().Get(ctx, nameOrGUID)
			if err != nil {
				// If not found by GUID, try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)
				buildpacks, err := client.Buildpacks().List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find buildpack: %w", err)
				}
				if len(buildpacks.Resources) == 0 {
					return fmt.Errorf("buildpack '%s' not found", nameOrGUID)
				}
				bp = &buildpacks.Resources[0]
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(bp)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(bp)
			default:
				fmt.Printf("Buildpack: %s\n", bp.Name)
				fmt.Printf("  GUID:      %s\n", bp.GUID)
				fmt.Printf("  Position:  %d\n", bp.Position)
				fmt.Printf("  State:     %s\n", bp.State)
				fmt.Printf("  Enabled:   %t\n", bp.Enabled)
				fmt.Printf("  Locked:    %t\n", bp.Locked)
				fmt.Printf("  Lifecycle: %s\n", bp.Lifecycle)
				fmt.Printf("  Created:   %s\n", bp.CreatedAt.Format("2006-01-02 15:04:05"))
				fmt.Printf("  Updated:   %s\n", bp.UpdatedAt.Format("2006-01-02 15:04:05"))

				if bp.Stack != nil {
					fmt.Printf("  Stack:     %s\n", *bp.Stack)
				}

				if bp.Filename != nil {
					fmt.Printf("  Filename:  %s\n", *bp.Filename)
				}
			}

			return nil
		},
	}
}

func newBuildpacksCreateCommand() *cobra.Command {
	var (
		name      string
		stack     string
		position  int
		enabled   bool
		locked    bool
		lifecycle string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a buildpack",
		Long:  "Create a new buildpack",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("buildpack name is required")
			}

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()

			createReq := &capi.BuildpackCreateRequest{
				Name: name,
			}

			if stack != "" {
				createReq.Stack = &stack
			}

			if cmd.Flags().Changed("position") {
				createReq.Position = &position
			}

			if cmd.Flags().Changed("enabled") {
				createReq.Enabled = &enabled
			}

			if cmd.Flags().Changed("locked") {
				createReq.Locked = &locked
			}

			if lifecycle != "" {
				createReq.Lifecycle = &lifecycle
			}

			bp, err := client.Buildpacks().Create(ctx, createReq)
			if err != nil {
				return fmt.Errorf("failed to create buildpack: %w", err)
			}

			fmt.Printf("Successfully created buildpack '%s'\n", bp.Name)
			fmt.Printf("  GUID:     %s\n", bp.GUID)
			fmt.Printf("  Position: %d\n", bp.Position)

			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "buildpack name (required)")
	cmd.Flags().StringVar(&stack, "stack", "", "stack name")
	cmd.Flags().IntVarP(&position, "position", "p", 0, "buildpack position")
	cmd.Flags().BoolVar(&enabled, "enabled", true, "enable the buildpack")
	cmd.Flags().BoolVar(&locked, "locked", false, "lock the buildpack")
	cmd.Flags().StringVar(&lifecycle, "lifecycle", "", "lifecycle type")
	cmd.MarkFlagRequired("name")

	return cmd
}

func newBuildpacksUpdateCommand() *cobra.Command {
	var (
		newName   string
		stack     string
		position  int
		enabled   bool
		locked    bool
		lifecycle string
	)

	cmd := &cobra.Command{
		Use:   "update BUILDPACK_NAME_OR_GUID",
		Short: "Update a buildpack",
		Long:  "Update an existing buildpack",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find buildpack
			var bpGUID string
			bp, err := client.Buildpacks().Get(ctx, nameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)
				buildpacks, err := client.Buildpacks().List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find buildpack: %w", err)
				}
				if len(buildpacks.Resources) == 0 {
					return fmt.Errorf("buildpack '%s' not found", nameOrGUID)
				}
				bp = &buildpacks.Resources[0]
			}
			bpGUID = bp.GUID

			// Build update request
			updateReq := &capi.BuildpackUpdateRequest{}

			if newName != "" {
				updateReq.Name = &newName
			}

			if stack != "" {
				updateReq.Stack = &stack
			}

			if cmd.Flags().Changed("position") {
				updateReq.Position = &position
			}

			if cmd.Flags().Changed("enabled") {
				updateReq.Enabled = &enabled
			}

			if cmd.Flags().Changed("locked") {
				updateReq.Locked = &locked
			}

			updatedBP, err := client.Buildpacks().Update(ctx, bpGUID, updateReq)
			if err != nil {
				return fmt.Errorf("failed to update buildpack: %w", err)
			}

			fmt.Printf("Successfully updated buildpack '%s'\n", updatedBP.Name)

			return nil
		},
	}

	cmd.Flags().StringVar(&newName, "name", "", "new buildpack name")
	cmd.Flags().StringVar(&stack, "stack", "", "stack name")
	cmd.Flags().IntVarP(&position, "position", "p", 0, "buildpack position")
	cmd.Flags().BoolVar(&enabled, "enabled", false, "enable the buildpack")
	cmd.Flags().BoolVar(&locked, "locked", false, "lock the buildpack")
	cmd.Flags().StringVar(&lifecycle, "lifecycle", "", "lifecycle type")

	return cmd
}

func newBuildpacksDeleteCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete BUILDPACK_NAME_OR_GUID",
		Short: "Delete a buildpack",
		Long:  "Delete a buildpack",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			if !force {
				fmt.Printf("Really delete buildpack '%s'? (y/N): ", nameOrGUID)
				var response string
				fmt.Scanln(&response)
				if response != "y" && response != "Y" {
					fmt.Println("Cancelled")
					return nil
				}
			}

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find buildpack
			var bpGUID string
			var bpName string
			bp, err := client.Buildpacks().Get(ctx, nameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)
				buildpacks, err := client.Buildpacks().List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find buildpack: %w", err)
				}
				if len(buildpacks.Resources) == 0 {
					return fmt.Errorf("buildpack '%s' not found", nameOrGUID)
				}
				bp = &buildpacks.Resources[0]
			}
			bpGUID = bp.GUID
			bpName = bp.Name

			job, err := client.Buildpacks().Delete(ctx, bpGUID)
			if err != nil {
				return fmt.Errorf("failed to delete buildpack: %w", err)
			}

			if job != nil {
				fmt.Printf("Deleting buildpack '%s'... (job: %s)\n", bpName, job.GUID)
				fmt.Printf("Monitor with: capi jobs get %s\n", job.GUID)
			} else {
				fmt.Printf("Successfully deleted buildpack '%s'\n", bpName)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "force deletion without confirmation")

	return cmd
}

func newBuildpacksUploadCommand() *cobra.Command {
	var (
		buildpackFile string
	)

	cmd := &cobra.Command{
		Use:   "upload BUILDPACK_NAME_OR_GUID BUILDPACK_FILE",
		Short: "Upload buildpack bits",
		Long:  "Upload a buildpack zip file to an existing buildpack",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]
			buildpackFile = args[1]

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find buildpack
			var bpGUID string
			bp, err := client.Buildpacks().Get(ctx, nameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)
				buildpacks, err := client.Buildpacks().List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find buildpack: %w", err)
				}
				if len(buildpacks.Resources) == 0 {
					return fmt.Errorf("buildpack '%s' not found", nameOrGUID)
				}
				bp = &buildpacks.Resources[0]
			}
			bpGUID = bp.GUID

			// Read buildpack file
			// Validate file path to prevent directory traversal
			if strings.Contains(buildpackFile, "..") {
				return fmt.Errorf("invalid file path: directory traversal not allowed")
			}
			buildpackBits, err := os.Open(buildpackFile) //nolint:gosec // G304: User-specified file path is intentional for CLI tool
			if err != nil {
				return fmt.Errorf("failed to open buildpack file: %w", err)
			}
			defer buildpackBits.Close()

			// Upload buildpack
			updatedBP, err := client.Buildpacks().Upload(ctx, bpGUID, buildpackBits)
			if err != nil {
				return fmt.Errorf("failed to upload buildpack: %w", err)
			}

			fmt.Printf("Successfully uploaded buildpack bits to '%s'\n", updatedBP.Name)
			fmt.Printf("  State: %s\n", updatedBP.State)
			if updatedBP.Filename != nil {
				fmt.Printf("  Filename: %s\n", *updatedBP.Filename)
			}

			return nil
		},
	}

	return cmd
}
