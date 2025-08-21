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

// NewStacksCommand creates the stacks command group
func NewStacksCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "stacks",
		Aliases: []string{"stack"},
		Short:   "Manage stacks",
		Long:    "List and manage Cloud Foundry stacks",
	}

	cmd.AddCommand(newStacksListCommand())
	cmd.AddCommand(newStacksGetCommand())
	cmd.AddCommand(newStacksCreateCommand())
	cmd.AddCommand(newStacksUpdateCommand())
	cmd.AddCommand(newStacksDeleteCommand())
	cmd.AddCommand(newStacksListAppsCommand())

	return cmd
}

func newStacksListCommand() *cobra.Command {
	var (
		allPages bool
		perPage  int
		name     string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List stacks",
		Long:  "List all stacks available in the platform",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			params := capi.NewQueryParams()
			params.PerPage = perPage

			// Apply name filter if specified
			if name != "" {
				params.WithFilter("names", name)
			}

			stacks, err := client.Stacks().List(ctx, params)
			if err != nil {
				return fmt.Errorf("failed to list stacks: %w", err)
			}

			// Fetch all pages if requested
			allStacks := stacks.Resources
			if allPages && stacks.Pagination.TotalPages > 1 {
				for page := 2; page <= stacks.Pagination.TotalPages; page++ {
					params.Page = page
					moreStacks, err := client.Stacks().List(ctx, params)
					if err != nil {
						return fmt.Errorf("failed to fetch page %d: %w", page, err)
					}
					allStacks = append(allStacks, moreStacks.Resources...)
				}
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(allStacks)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(allStacks)
			default:
				if len(allStacks) == 0 {
					fmt.Println("No stacks found")
					return nil
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Name", "Description", "Default", "Build Image", "Run Image", "Created", "Updated")

				for _, stack := range allStacks {
					defaultStack := "false"
					if stack.Default {
						defaultStack = "true"
					}

					// Truncate long image names for table display
					buildImage := stack.BuildRootfsImage
					if len(buildImage) > 40 {
						buildImage = buildImage[:37] + "..."
					}

					runImage := stack.RunRootfsImage
					if len(runImage) > 40 {
						runImage = runImage[:37] + "..."
					}

					_ = table.Append([]string{
						stack.Name,
						stack.Description,
						defaultStack,
						buildImage,
						runImage,
						stack.CreatedAt.Format("2006-01-02 15:04:05"),
						stack.UpdatedAt.Format("2006-01-02 15:04:05"),
					})
				}

				_ = table.Render()
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&allPages, "all-pages", false, "fetch all pages")
	cmd.Flags().IntVar(&perPage, "per-page", 50, "number of results per page")
	cmd.Flags().StringVar(&name, "name", "", "filter by stack name")

	return cmd
}

func newStacksGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get STACK_GUID",
		Short: "Get stack details",
		Long:  "Display detailed information about a specific stack",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			stackGUID := args[0]

			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			stack, err := client.Stacks().Get(ctx, stackGUID)
			if err != nil {
				return fmt.Errorf("failed to get stack: %w", err)
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(stack)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(stack)
			default:
				fmt.Printf("Stack: %s\n", stack.GUID)
				fmt.Printf("  Name:        %s\n", stack.Name)
				fmt.Printf("  Description: %s\n", stack.Description)
				fmt.Printf("  Default:     %t\n", stack.Default)
				fmt.Printf("  Build Image: %s\n", stack.BuildRootfsImage)
				fmt.Printf("  Run Image:   %s\n", stack.RunRootfsImage)
				fmt.Printf("  Created:     %s\n", stack.CreatedAt.Format("2006-01-02 15:04:05"))
				fmt.Printf("  Updated:     %s\n", stack.UpdatedAt.Format("2006-01-02 15:04:05"))
			}

			return nil
		},
	}
}

func newStacksCreateCommand() *cobra.Command {
	var (
		description      string
		buildRootfsImage string
		runRootfsImage   string
	)

	cmd := &cobra.Command{
		Use:   "create STACK_NAME",
		Short: "Create a stack",
		Long:  "Create a new stack",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			stackName := args[0]

			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			request := &capi.StackCreateRequest{
				Name:             stackName,
				Description:      description,
				BuildRootfsImage: buildRootfsImage,
				RunRootfsImage:   runRootfsImage,
			}

			stack, err := client.Stacks().Create(ctx, request)
			if err != nil {
				return fmt.Errorf("failed to create stack: %w", err)
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(stack)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(stack)
			default:
				fmt.Printf("Created stack: %s\n", stack.GUID)
				fmt.Printf("  Name:        %s\n", stack.Name)
				fmt.Printf("  Description: %s\n", stack.Description)
				fmt.Printf("  Build Image: %s\n", stack.BuildRootfsImage)
				fmt.Printf("  Run Image:   %s\n", stack.RunRootfsImage)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&description, "description", "", "stack description")
	cmd.Flags().StringVar(&buildRootfsImage, "build-image", "", "build rootfs image")
	cmd.Flags().StringVar(&runRootfsImage, "run-image", "", "run rootfs image")

	return cmd
}

func newStacksUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update STACK_GUID",
		Short: "Update a stack",
		Long:  "Update stack metadata",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			stackGUID := args[0]

			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// For now, stacks only support metadata updates
			request := &capi.StackUpdateRequest{
				Metadata: &capi.Metadata{
					Labels:      make(map[string]string),
					Annotations: make(map[string]string),
				},
			}

			stack, err := client.Stacks().Update(ctx, stackGUID, request)
			if err != nil {
				return fmt.Errorf("failed to update stack: %w", err)
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(stack)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(stack)
			default:
				fmt.Printf("Updated stack: %s\n", stack.GUID)
				fmt.Printf("  Name:        %s\n", stack.Name)
				fmt.Printf("  Description: %s\n", stack.Description)
			}

			return nil
		},
	}

	return cmd
}

func newStacksDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete STACK_GUID",
		Short: "Delete a stack",
		Long:  "Delete a stack",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			stackGUID := args[0]

			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			err = client.Stacks().Delete(ctx, stackGUID)
			if err != nil {
				return fmt.Errorf("failed to delete stack: %w", err)
			}

			fmt.Printf("Stack %s deleted successfully\n", stackGUID)

			return nil
		},
	}
}

func newStacksListAppsCommand() *cobra.Command {
	var (
		allPages bool
		perPage  int
	)

	cmd := &cobra.Command{
		Use:   "list-apps STACK_GUID",
		Short: "List apps using a stack",
		Long:  "List all applications that are using the specified stack",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			stackGUID := args[0]

			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			params := capi.NewQueryParams()
			params.PerPage = perPage

			apps, err := client.Stacks().ListApps(ctx, stackGUID, params)
			if err != nil {
				return fmt.Errorf("failed to list apps for stack: %w", err)
			}

			// Fetch all pages if requested
			allApps := apps.Resources
			if allPages && apps.Pagination.TotalPages > 1 {
				for page := 2; page <= apps.Pagination.TotalPages; page++ {
					params.Page = page
					moreApps, err := client.Stacks().ListApps(ctx, stackGUID, params)
					if err != nil {
						return fmt.Errorf("failed to fetch page %d: %w", page, err)
					}
					allApps = append(allApps, moreApps.Resources...)
				}
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(allApps)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(allApps)
			default:
				if len(allApps) == 0 {
					fmt.Println("No apps found using this stack")
					return nil
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Name", "GUID", "State", "Lifecycle", "Created", "Updated")

				for _, app := range allApps {
					lifecycle := app.Lifecycle.Type
					if len(app.Lifecycle.Data) > 0 {
						if stackName, ok := app.Lifecycle.Data["stack"]; ok {
							lifecycle += fmt.Sprintf(" (%v)", stackName)
						}
					}

					_ = table.Append([]string{
						app.Name,
						app.GUID,
						app.State,
						lifecycle,
						app.CreatedAt.Format("2006-01-02 15:04:05"),
						app.UpdatedAt.Format("2006-01-02 15:04:05"),
					})
				}

				_ = table.Render()
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&allPages, "all-pages", false, "fetch all pages")
	cmd.Flags().IntVar(&perPage, "per-page", 50, "number of results per page")

	return cmd
}
