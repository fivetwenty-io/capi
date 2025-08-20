package commands

import (
	"context"
	"fmt"

	"github.com/fivetwenty-io/capi-client/pkg/capi"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewSpacesCommand creates the spaces command group
func NewSpacesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "spaces",
		Aliases: []string{"space"},
		Short:   "Manage spaces",
		Long:    "List and manage Cloud Foundry spaces",
	}

	cmd.AddCommand(newSpacesListCommand())
	return cmd
}

func newSpacesListCommand() *cobra.Command {
	var orgName string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List spaces",
		Long:  "List all spaces the user has access to",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			params := capi.NewQueryParams()

			// Filter by organization if specified
			if orgName != "" {
				// Find org by name
				orgParams := capi.NewQueryParams()
				orgParams.WithFilter("names", orgName)
				orgs, err := client.Organizations().List(ctx, orgParams)
				if err != nil {
					return fmt.Errorf("failed to find organization: %w", err)
				}
				if len(orgs.Resources) == 0 {
					return fmt.Errorf("organization '%s' not found", orgName)
				}

				params.WithFilter("organization_guids", orgs.Resources[0].GUID)
			} else if orgGUID := viper.GetString("organization_guid"); orgGUID != "" {
				// Use targeted organization
				params.WithFilter("organization_guids", orgGUID)
			}

			spaces, err := client.Spaces().List(ctx, params)
			if err != nil {
				return fmt.Errorf("failed to list spaces: %w", err)
			}

			if len(spaces.Resources) == 0 {
				fmt.Println("No spaces found")
				return nil
			}

			fmt.Println("Spaces:")
			for _, space := range spaces.Resources {
				fmt.Printf("  %s (%s)\n", space.Name, space.GUID)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&orgName, "org", "o", "", "filter by organization name")

	return cmd
}
