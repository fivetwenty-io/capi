package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/fivetwenty-io/capi/pkg/capi"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// NewFeatureFlagsCommand creates the feature-flags command group
func NewFeatureFlagsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "feature-flags",
		Aliases: []string{"feature-flag", "ff", "flags"},
		Short:   "Manage feature flags",
		Long:    "List and manage Cloud Foundry feature flags",
	}

	cmd.AddCommand(newFeatureFlagsListCommand())
	cmd.AddCommand(newFeatureFlagsGetCommand())
	cmd.AddCommand(newFeatureFlagsUpdateCommand())
	cmd.AddCommand(newFeatureFlagsEnableCommand())
	cmd.AddCommand(newFeatureFlagsDisableCommand())

	return cmd
}

func newFeatureFlagsListCommand() *cobra.Command {
	var (
		allPages bool
		perPage  int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List feature flags",
		Long:  "List all Cloud Foundry feature flags",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			params := capi.NewQueryParams()
			if perPage > 0 {
				params.PerPage = perPage
			}

			flags, err := client.FeatureFlags().List(ctx, params)
			if err != nil {
				return fmt.Errorf("failed to list feature flags: %w", err)
			}

			// Fetch all pages if requested
			allFlags := flags.Resources
			if allPages && flags.Pagination.TotalPages > 1 {
				for page := 2; page <= flags.Pagination.TotalPages; page++ {
					params.Page = page
					moreFlags, err := client.FeatureFlags().List(ctx, params)
					if err != nil {
						return fmt.Errorf("failed to fetch page %d: %w", page, err)
					}
					allFlags = append(allFlags, moreFlags.Resources...)
				}
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(allFlags)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				encoder.SetIndent(2)
				return encoder.Encode(allFlags)
			default:
				if len(allFlags) == 0 {
					fmt.Println("No feature flags found")
					return nil
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Name", "Enabled", "Updated", "Error Message")

				for _, flag := range allFlags {
					enabled := "false"
					if flag.Enabled {
						enabled = "true"
					}

					updatedAt := ""
					if flag.UpdatedAt != nil && !flag.UpdatedAt.IsZero() {
						updatedAt = flag.UpdatedAt.Format("2006-01-02 15:04:05")
					}

					errorMessage := ""
					if flag.CustomErrorMessage != nil {
						errorMessage = *flag.CustomErrorMessage
					}

					_ = table.Append(flag.Name, enabled, updatedAt, errorMessage)
				}

				_ = table.Render()

				if !allPages && flags.Pagination.TotalPages > 1 {
					fmt.Printf("\nShowing page 1 of %d. Use --all to fetch all pages.\n", flags.Pagination.TotalPages)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&allPages, "all", false, "fetch all pages")
	cmd.Flags().IntVar(&perPage, "per-page", 50, "results per page")

	return cmd
}

func newFeatureFlagsGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get FEATURE_FLAG_NAME",
		Short: "Get feature flag details",
		Long:  "Display detailed information about a specific feature flag",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			flagName := args[0]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			flag, err := client.FeatureFlags().Get(ctx, flagName)
			if err != nil {
				return fmt.Errorf("failed to get feature flag '%s': %w", flagName, err)
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(flag)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(flag)
			default:
				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Property", "Value")

				_ = table.Append("Name", flag.Name)
				_ = table.Append("Enabled", fmt.Sprintf("%t", flag.Enabled))

				if flag.CustomErrorMessage != nil && *flag.CustomErrorMessage != "" {
					_ = table.Append("Error Message", *flag.CustomErrorMessage)
				}

				if flag.UpdatedAt != nil && !flag.UpdatedAt.IsZero() {
					_ = table.Append("Updated", flag.UpdatedAt.Format("2006-01-02 15:04:05"))
				}

				fmt.Printf("Feature Flag: %s\n\n", flag.Name)
				_ = table.Render()
			}

			return nil
		},
	}
}

func newFeatureFlagsUpdateCommand() *cobra.Command {
	var (
		enabled      *bool
		errorMessage string
	)

	cmd := &cobra.Command{
		Use:   "update FEATURE_FLAG_NAME",
		Short: "Update a feature flag",
		Long:  "Update the state or error message of a Cloud Foundry feature flag",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			flagName := args[0]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Check if any parameters were provided
			if enabled == nil && errorMessage == "" {
				// Show current flag if no update parameters
				flag, err := client.FeatureFlags().Get(ctx, flagName)
				if err != nil {
					return fmt.Errorf("failed to get feature flag '%s': %w", flagName, err)
				}

				fmt.Printf("Feature flag '%s' current state:\n", flag.Name)
				fmt.Printf("  Enabled: %t\n", flag.Enabled)
				if flag.CustomErrorMessage != nil && *flag.CustomErrorMessage != "" {
					fmt.Printf("  Error Message: %s\n", *flag.CustomErrorMessage)
				}
				return nil
			}

			// Build update request
			updateReq := &capi.FeatureFlagUpdateRequest{}

			if enabled != nil {
				updateReq.Enabled = *enabled
			}

			if errorMessage != "" {
				updateReq.CustomErrorMessage = &errorMessage
			}

			// Update feature flag
			updatedFlag, err := client.FeatureFlags().Update(ctx, flagName, updateReq)
			if err != nil {
				return fmt.Errorf("failed to update feature flag '%s': %w", flagName, err)
			}

			fmt.Printf("Successfully updated feature flag '%s'\n", updatedFlag.Name)
			fmt.Printf("  Enabled: %t\n", updatedFlag.Enabled)
			if updatedFlag.CustomErrorMessage != nil && *updatedFlag.CustomErrorMessage != "" {
				fmt.Printf("  Error Message: %s\n", *updatedFlag.CustomErrorMessage)
			}

			return nil
		},
	}

	// Use a helper for the enabled flag to distinguish between not set and false
	var enabledStr string
	cmd.Flags().StringVar(&enabledStr, "enabled", "", "enable or disable the feature flag (true/false)")
	cmd.Flags().StringVar(&errorMessage, "error-message", "", "custom error message when flag is disabled")

	// Custom validation for enabled flag
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if enabledStr != "" {
			switch enabledStr {
			case "true":
				trueVal := true
				enabled = &trueVal
			case "false":
				falseVal := false
				enabled = &falseVal
			default:
				return fmt.Errorf("enabled flag must be 'true' or 'false', got '%s'", enabledStr)
			}
		}
		return nil
	}

	return cmd
}

func newFeatureFlagsEnableCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "enable FEATURE_FLAG_NAME",
		Short: "Enable a feature flag",
		Long:  "Enable a Cloud Foundry feature flag",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			flagName := args[0]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			updateReq := &capi.FeatureFlagUpdateRequest{
				Enabled: true,
			}

			updatedFlag, err := client.FeatureFlags().Update(ctx, flagName, updateReq)
			if err != nil {
				return fmt.Errorf("failed to enable feature flag '%s': %w", flagName, err)
			}

			fmt.Printf("Successfully enabled feature flag '%s'\n", updatedFlag.Name)
			return nil
		},
	}
}

func newFeatureFlagsDisableCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "disable FEATURE_FLAG_NAME",
		Short: "Disable a feature flag",
		Long:  "Disable a Cloud Foundry feature flag",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			flagName := args[0]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			updateReq := &capi.FeatureFlagUpdateRequest{
				Enabled: false,
			}

			updatedFlag, err := client.FeatureFlags().Update(ctx, flagName, updateReq)
			if err != nil {
				return fmt.Errorf("failed to disable feature flag '%s': %w", flagName, err)
			}

			fmt.Printf("Successfully disabled feature flag '%s'\n", updatedFlag.Name)
			return nil
		},
	}
}
