package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/fivetwenty-io/capi-client/pkg/capi"
	"github.com/fivetwenty-io/capi-client/pkg/cfclient"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// TargetInfo represents the current target information
type TargetInfo struct {
	API          string `json:"api,omitempty" yaml:"api,omitempty"`
	Endpoint     string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	User         string `json:"user,omitempty" yaml:"user,omitempty"`
	Organization string `json:"organization,omitempty" yaml:"organization,omitempty"`
	Space        string `json:"space,omitempty" yaml:"space,omitempty"`
}

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
			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Get current API config
			config := loadConfig()
			apiConfig, err := getCurrentAPIConfig()
			if err != nil {
				return err
			}

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
				apiConfig.Organization = org.Name
				apiConfig.OrganizationGUID = org.GUID
				fmt.Printf("Targeted organization: %s\n", org.Name)

				// Clear space if only org is being targeted
				if spaceName == "" {
					apiConfig.Space = ""
					apiConfig.SpaceGUID = ""

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
				if apiConfig.OrganizationGUID == "" {
					return fmt.Errorf("organization must be targeted first")
				}

				spacesClient := client.Spaces()
				spaceParams := capi.NewQueryParams()
				spaceParams.WithFilter("names", spaceName)
				spaceParams.WithFilter("organization_guids", apiConfig.OrganizationGUID)
				spaces, err := spacesClient.List(ctx, spaceParams)
				if err != nil {
					return fmt.Errorf("failed to find space: %w", err)
				}
				if len(spaces.Resources) == 0 {
					return fmt.Errorf("space '%s' not found in organization", spaceName)
				}

				space := spaces.Resources[0]
				apiConfig.Space = space.Name
				apiConfig.SpaceGUID = space.GUID
				fmt.Printf("Targeted space: %s\n", space.Name)
			}

			// Update config and save
			config.APIs[config.CurrentAPI] = apiConfig
			if err := saveConfigStruct(config); err != nil {
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
	config := loadConfig()

	if config.CurrentAPI == "" || len(config.APIs) == 0 {
		fmt.Println("No API targeted. Use 'capi apis add' to add an API endpoint.")
		return nil
	}

	apiConfig, exists := config.APIs[config.CurrentAPI]
	if !exists {
		fmt.Printf("Current API '%s' not found in configuration.\n", config.CurrentAPI)
		return nil
	}

	// Create target info struct
	targetInfo := TargetInfo{
		API:      config.CurrentAPI,
		Endpoint: apiConfig.Endpoint,
	}

	if apiConfig.Username != "" {
		targetInfo.User = apiConfig.Username
	}
	if apiConfig.Organization != "" {
		targetInfo.Organization = apiConfig.Organization
	}
	if apiConfig.Space != "" {
		targetInfo.Space = apiConfig.Space
	}

	// Output results
	output := viper.GetString("output")
	switch output {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(targetInfo)
	case "yaml":
		encoder := yaml.NewEncoder(os.Stdout)
		return encoder.Encode(targetInfo)
	default:
		table := tablewriter.NewWriter(os.Stdout)
		table.Header("Property", "Value")

		_ = table.Append("API", targetInfo.API)
		_ = table.Append("Endpoint", targetInfo.Endpoint)

		if targetInfo.User != "" {
			_ = table.Append("User", targetInfo.User)
		}
		if targetInfo.Organization != "" {
			_ = table.Append("Organization", targetInfo.Organization)
		}
		if targetInfo.Space != "" {
			_ = table.Append("Space", targetInfo.Space)
		}

		_ = table.Render()
	}

	return nil
}

func createClientWithAPI(apiFlag string) (capi.Client, error) {
	// Get API config based on flag or current API
	apiConfig, err := getAPIConfigByFlag(apiFlag)
	if err != nil {
		return nil, err
	}

	if apiConfig.Endpoint == "" {
		return nil, fmt.Errorf("no API endpoint configured, use 'capi apis add' first")
	}

	config := &capi.Config{
		APIEndpoint:   apiConfig.Endpoint,
		AccessToken:   apiConfig.Token,
		SkipTLSVerify: apiConfig.SkipSSLValidation,
		Username:      apiConfig.Username,
	}

	// If we have no token and no username, require authentication
	if config.AccessToken == "" && config.Username == "" {
		return nil, fmt.Errorf("not authenticated, use 'capi login' first")
	}

	return cfclient.New(config)
}
