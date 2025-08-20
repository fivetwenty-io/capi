package commands

import (
	"context"
	"fmt"

	"github.com/fivetwenty-io/capi-client/pkg/capi"
	"github.com/fivetwenty-io/capi-client/pkg/cfclient"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewTargetCommand creates the target command
func NewTargetCommand() *cobra.Command {
	var (
		orgName   string
		spaceName string
	)

	cmd := &cobra.Command{
		Use:   "target",
		Short: "Set or show the targeted organization and space",
		Long:  "Set or display the currently targeted Cloud Foundry organization and space",
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no flags provided, show current target
			if orgName == "" && spaceName == "" {
				return showTarget()
			}

			// Create client
			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Target organization
			if orgName != "" {
				orgsClient := client.Organizations()
				params := capi.NewQueryParams()
				params.WithFilter("names", orgName)
				orgs, err := orgsClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find organization: %w", err)
				}
				if len(orgs.Resources) == 0 {
					return fmt.Errorf("organization '%s' not found", orgName)
				}

				org := orgs.Resources[0]
				viper.Set("organization", org.Name)
				viper.Set("organization_guid", org.GUID)
				fmt.Printf("Targeted organization: %s\n", org.Name)

				// Clear space if only org is being targeted
				if spaceName == "" {
					viper.Set("space", "")
					viper.Set("space_guid", "")

					// List available spaces
					spacesClient := client.Spaces()
					spaceParams := capi.NewQueryParams()
					spaceParams.WithFilter("organization_guids", org.GUID)
					spaces, err := spacesClient.List(ctx, spaceParams)
					if err == nil && len(spaces.Resources) > 0 {
						fmt.Println("\nAvailable spaces:")
						for _, space := range spaces.Resources {
							fmt.Printf("  - %s\n", space.Name)
						}
						fmt.Println("\nUse 'capi target -s <space>' to target a space")
					}
				}
			}

			// Target space (requires organization to be set)
			if spaceName != "" {
				orgGUID := viper.GetString("organization_guid")
				if orgGUID == "" {
					return fmt.Errorf("organization must be targeted first")
				}

				spacesClient := client.Spaces()
				spaceParams := capi.NewQueryParams()
				spaceParams.WithFilter("names", spaceName)
				spaceParams.WithFilter("organization_guids", orgGUID)
				spaces, err := spacesClient.List(ctx, spaceParams)
				if err != nil {
					return fmt.Errorf("failed to find space: %w", err)
				}
				if len(spaces.Resources) == 0 {
					return fmt.Errorf("space '%s' not found in organization", spaceName)
				}

				space := spaces.Resources[0]
				viper.Set("space", space.Name)
				viper.Set("space_guid", space.GUID)
				fmt.Printf("Targeted space: %s\n", space.Name)
			}

			// Save configuration
			if err := saveConfig(); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&orgName, "org", "o", "", "target organization")
	cmd.Flags().StringVarP(&spaceName, "space", "s", "", "target space")

	return cmd
}

func showTarget() error {
	api := viper.GetString("api")
	username := viper.GetString("username")
	org := viper.GetString("organization")
	space := viper.GetString("space")

	if api == "" {
		fmt.Println("Not logged in. Use 'capi login' to authenticate.")
		return nil
	}

	fmt.Println("Current target:")
	fmt.Printf("  API:          %s\n", api)
	if username != "" {
		fmt.Printf("  User:         %s\n", username)
	}
	if org != "" {
		fmt.Printf("  Organization: %s\n", org)
	}
	if space != "" {
		fmt.Printf("  Space:        %s\n", space)
	}

	return nil
}

func createClient() (capi.Client, error) {
	api := viper.GetString("api")
	if api == "" {
		return nil, fmt.Errorf("not logged in, use 'capi login' first")
	}

	config := &capi.Config{
		APIEndpoint:   api,
		AccessToken:   viper.GetString("token"),
		SkipTLSVerify: viper.GetBool("skip_ssl_validation"),
	}

	// Try to use stored credentials if no token
	if config.AccessToken == "" {
		config.Username = viper.GetString("username")
		config.Password = viper.GetString("password")
	}

	return cfclient.New(config)
}
