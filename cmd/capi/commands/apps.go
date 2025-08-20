package commands

import (
	"context"
	"fmt"

	"github.com/fivetwenty-io/capi-client/pkg/capi"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewAppsCommand creates the apps command group
func NewAppsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "apps",
		Aliases: []string{"app", "applications"},
		Short:   "Manage applications",
		Long:    "List, create, and manage Cloud Foundry applications",
	}

	cmd.AddCommand(newAppsListCommand())
	cmd.AddCommand(newAppsStartCommand())
	cmd.AddCommand(newAppsStopCommand())
	cmd.AddCommand(newAppsRestartCommand())

	return cmd
}

func newAppsListCommand() *cobra.Command {
	var spaceName string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List applications",
		Long:  "List all applications the user has access to",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			params := capi.NewQueryParams()

			// Filter by space if specified
			if spaceName != "" {
				// Find space by name
				spaceParams := capi.NewQueryParams()
				spaceParams.WithFilter("names", spaceName)

				// Add org filter if targeted
				if orgGUID := viper.GetString("organization_guid"); orgGUID != "" {
					spaceParams.WithFilter("organization_guids", orgGUID)
				}

				spaces, err := client.Spaces().List(ctx, spaceParams)
				if err != nil {
					return fmt.Errorf("failed to find space: %w", err)
				}
				if len(spaces.Resources) == 0 {
					return fmt.Errorf("space '%s' not found", spaceName)
				}

				params.WithFilter("space_guids", spaces.Resources[0].GUID)
			} else if spaceGUID := viper.GetString("space_guid"); spaceGUID != "" {
				// Use targeted space
				params.WithFilter("space_guids", spaceGUID)
			}

			apps, err := client.Apps().List(ctx, params)
			if err != nil {
				return fmt.Errorf("failed to list applications: %w", err)
			}

			if len(apps.Resources) == 0 {
				fmt.Println("No applications found")
				return nil
			}

			fmt.Println("Applications:")
			for _, app := range apps.Resources {
				lifecycle := "buildpack"
				if app.Lifecycle.Type == "docker" {
					lifecycle = "docker"
				}
				fmt.Printf("  %s (%s) - %s [%s]\n", app.Name, app.GUID, app.State, lifecycle)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&spaceName, "space", "s", "", "filter by space name")

	return cmd
}

func newAppsStartCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "start APP_NAME_OR_GUID",
		Short: "Start an application",
		Long:  "Start a Cloud Foundry application",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find application
			appGUID, appName, err := resolveApp(ctx, client, nameOrGUID)
			if err != nil {
				return err
			}

			// Start application
			app, err := client.Apps().Start(ctx, appGUID)
			if err != nil {
				return fmt.Errorf("failed to start application: %w", err)
			}

			fmt.Printf("Successfully started application '%s'\n", app.Name)
			_ = appName // Use appName if needed
			return nil
		},
	}
}

func newAppsStopCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stop APP_NAME_OR_GUID",
		Short: "Stop an application",
		Long:  "Stop a Cloud Foundry application",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find application
			appGUID, _, err := resolveApp(ctx, client, nameOrGUID)
			if err != nil {
				return err
			}

			// Stop application
			app, err := client.Apps().Stop(ctx, appGUID)
			if err != nil {
				return fmt.Errorf("failed to stop application: %w", err)
			}

			fmt.Printf("Successfully stopped application '%s'\n", app.Name)
			return nil
		},
	}
}

func newAppsRestartCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "restart APP_NAME_OR_GUID",
		Short: "Restart an application",
		Long:  "Restart a Cloud Foundry application",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find application
			appGUID, _, err := resolveApp(ctx, client, nameOrGUID)
			if err != nil {
				return err
			}

			// Restart application
			app, err := client.Apps().Restart(ctx, appGUID)
			if err != nil {
				return fmt.Errorf("failed to restart application: %w", err)
			}

			fmt.Printf("Successfully restarted application '%s'\n", app.Name)
			return nil
		},
	}
}

// Helper function to resolve app name or GUID
func resolveApp(ctx context.Context, client capi.Client, nameOrGUID string) (guid string, name string, err error) {
	// Try to get by GUID first
	app, err := client.Apps().Get(ctx, nameOrGUID)
	if err == nil {
		return app.GUID, app.Name, nil
	}

	// Try by name
	params := capi.NewQueryParams()
	params.WithFilter("names", nameOrGUID)

	// Add space filter if targeted
	if spaceGUID := viper.GetString("space_guid"); spaceGUID != "" {
		params.WithFilter("space_guids", spaceGUID)
	}

	apps, err := client.Apps().List(ctx, params)
	if err != nil {
		return "", "", fmt.Errorf("failed to find application: %w", err)
	}
	if len(apps.Resources) == 0 {
		return "", "", fmt.Errorf("application '%s' not found", nameOrGUID)
	}

	return apps.Resources[0].GUID, apps.Resources[0].Name, nil
}
