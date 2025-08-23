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

// NewRolesCommand creates the roles command group
func NewRolesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "roles",
		Aliases: []string{"role"},
		Short:   "Manage roles",
		Long:    "List and manage Cloud Foundry user roles",
	}

	cmd.AddCommand(newRolesListCommand())
	cmd.AddCommand(newRolesGetCommand())
	cmd.AddCommand(newRolesCreateCommand())
	cmd.AddCommand(newRolesDeleteCommand())

	return cmd
}

func newRolesListCommand() *cobra.Command {
	var (
		allPages  bool
		perPage   int
		userGUID  string
		orgGUID   string
		spaceGUID string
		roleType  string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List roles",
		Long:  "List all roles the user has access to",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			params := capi.NewQueryParams()
			params.PerPage = perPage

			// Apply filters
			if userGUID != "" {
				params.WithFilter("user_guids", userGUID)
			}
			if orgGUID != "" {
				params.WithFilter("organization_guids", orgGUID)
			}
			if spaceGUID != "" {
				params.WithFilter("space_guids", spaceGUID)
			}
			if roleType != "" {
				params.WithFilter("types", roleType)
			}

			roles, err := client.Roles().List(ctx, params)
			if err != nil {
				return fmt.Errorf("failed to list roles: %w", err)
			}

			// Fetch all pages if requested
			allRoles := roles.Resources
			if allPages && roles.Pagination.TotalPages > 1 {
				for page := 2; page <= roles.Pagination.TotalPages; page++ {
					params.Page = page
					moreRoles, err := client.Roles().List(ctx, params)
					if err != nil {
						return fmt.Errorf("failed to fetch page %d: %w", page, err)
					}
					allRoles = append(allRoles, moreRoles.Resources...)
				}
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(allRoles)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(allRoles)
			default:
				if len(allRoles) == 0 {
					fmt.Println("No roles found")
					return nil
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.Header("GUID", "Type", "User GUID", "Organization GUID", "Space GUID", "Created", "Updated")

				for _, role := range allRoles {
					orgGUID := ""
					if role.Relationships.Organization != nil {
						orgGUID = role.Relationships.Organization.Data.GUID
					}

					spaceGUID := ""
					if role.Relationships.Space != nil {
						spaceGUID = role.Relationships.Space.Data.GUID
					}

					_ = table.Append([]string{
						role.GUID,
						role.Type,
						role.Relationships.User.Data.GUID,
						orgGUID,
						spaceGUID,
						role.CreatedAt.Format("2006-01-02 15:04:05"),
						role.UpdatedAt.Format("2006-01-02 15:04:05"),
					})
				}

				_ = table.Render()
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&allPages, "all-pages", false, "fetch all pages")
	cmd.Flags().IntVar(&perPage, "per-page", 50, "number of results per page")
	cmd.Flags().StringVar(&userGUID, "user", "", "filter by user GUID")
	cmd.Flags().StringVar(&orgGUID, "org", "", "filter by organization GUID")
	cmd.Flags().StringVar(&spaceGUID, "space", "", "filter by space GUID")
	cmd.Flags().StringVar(&roleType, "type", "", "filter by role type")

	return cmd
}

func newRolesGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get ROLE_GUID",
		Short: "Get role details",
		Long:  "Display detailed information about a specific role",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			roleGUID := args[0]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			role, err := client.Roles().Get(ctx, roleGUID)
			if err != nil {
				return fmt.Errorf("failed to get role: %w", err)
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(role)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(role)
			default:
				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Property", "Value")

				_ = table.Append("GUID", role.GUID)
				_ = table.Append("Type", role.Type)
				_ = table.Append("User GUID", role.Relationships.User.Data.GUID)

				if role.Relationships.Organization != nil {
					_ = table.Append("Organization GUID", role.Relationships.Organization.Data.GUID)
				}

				if role.Relationships.Space != nil {
					_ = table.Append("Space GUID", role.Relationships.Space.Data.GUID)
				}

				_ = table.Append("Created", role.CreatedAt.Format("2006-01-02 15:04:05"))
				_ = table.Append("Updated", role.UpdatedAt.Format("2006-01-02 15:04:05"))

				fmt.Printf("Role details:\n\n")
				_ = table.Render()
			}

			return nil
		},
	}
}

func newRolesCreateCommand() *cobra.Command {
	var (
		userGUID  string
		orgGUID   string
		spaceGUID string
	)

	cmd := &cobra.Command{
		Use:   "create ROLE_TYPE",
		Short: "Create a role",
		Long:  "Create a new role assignment for a user",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			roleType := args[0]

			if userGUID == "" {
				return fmt.Errorf("user GUID is required")
			}

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Build relationships
			relationships := capi.RoleRelationships{
				User: capi.Relationship{
					Data: &capi.RelationshipData{GUID: userGUID},
				},
			}

			if orgGUID != "" {
				relationships.Organization = &capi.Relationship{
					Data: &capi.RelationshipData{GUID: orgGUID},
				}
			}

			if spaceGUID != "" {
				relationships.Space = &capi.Relationship{
					Data: &capi.RelationshipData{GUID: spaceGUID},
				}
			}

			request := &capi.RoleCreateRequest{
				Type:          roleType,
				Relationships: relationships,
			}

			role, err := client.Roles().Create(ctx, request)
			if err != nil {
				return fmt.Errorf("failed to create role: %w", err)
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(role)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(role)
			default:
				fmt.Printf("Created role: %s\n", role.GUID)
				fmt.Printf("  Type:      %s\n", role.Type)
				fmt.Printf("  User GUID: %s\n", role.Relationships.User.Data.GUID)

				if role.Relationships.Organization != nil {
					fmt.Printf("  Org GUID:  %s\n", role.Relationships.Organization.Data.GUID)
				}

				if role.Relationships.Space != nil {
					fmt.Printf("  Space GUID: %s\n", role.Relationships.Space.Data.GUID)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&userGUID, "user", "", "user GUID (required)")
	cmd.Flags().StringVar(&orgGUID, "org", "", "organization GUID")
	cmd.Flags().StringVar(&spaceGUID, "space", "", "space GUID")
	_ = cmd.MarkFlagRequired("user")

	return cmd
}

func newRolesDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete ROLE_GUID",
		Short: "Delete a role",
		Long:  "Delete a role assignment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			roleGUID := args[0]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			err = client.Roles().Delete(ctx, roleGUID)
			if err != nil {
				return fmt.Errorf("failed to delete role: %w", err)
			}

			fmt.Printf("Role %s deleted successfully\n", roleGUID)

			return nil
		},
	}
}
