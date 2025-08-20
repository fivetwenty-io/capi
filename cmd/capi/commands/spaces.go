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

// NewSpacesCommand creates the spaces command group
func NewSpacesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "spaces",
		Aliases: []string{"space"},
		Short:   "Manage spaces",
		Long:    "List and manage Cloud Foundry spaces",
	}

	cmd.AddCommand(newSpacesListCommand())
	cmd.AddCommand(newSpacesGetCommand())
	cmd.AddCommand(newSpacesCreateCommand())
	cmd.AddCommand(newSpacesUpdateCommand())
	cmd.AddCommand(newSpacesDeleteCommand())
	cmd.AddCommand(newSpacesSetQuotaCommand())
	cmd.AddCommand(newSpacesListUsersCommand())
	cmd.AddCommand(newSpacesSetRoleCommand())
	cmd.AddCommand(newSpacesUnsetRoleCommand())
	cmd.AddCommand(newSpacesListAppsCommand())
	cmd.AddCommand(newSpacesListServicesCommand())
	cmd.AddCommand(newSpacesApplyManifestCommand())
	return cmd
}

func newSpacesListCommand() *cobra.Command {
	var (
		orgName  string
		allPages bool
		perPage  int
	)

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
			params.PerPage = perPage

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

			// Fetch all pages if requested
			allSpaces := spaces.Resources
			if allPages && spaces.Pagination.TotalPages > 1 {
				for page := 2; page <= spaces.Pagination.TotalPages; page++ {
					params.Page = page
					moreSpaces, err := client.Spaces().List(ctx, params)
					if err != nil {
						return fmt.Errorf("failed to fetch page %d: %w", page, err)
					}
					allSpaces = append(allSpaces, moreSpaces.Resources...)
				}
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(allSpaces)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(allSpaces)
			default:
				if len(allSpaces) == 0 {
					fmt.Println("No spaces found")
					return nil
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Name", "GUID", "Organization", "Created", "Updated")

				for _, space := range allSpaces {
					// Get org name if available
					orgName := ""
					if space.Relationships.Organization.Data != nil {
						org, _ := client.Organizations().Get(ctx, space.Relationships.Organization.Data.GUID)
						if org != nil {
							orgName = org.Name
						}
					}

					table.Append(space.Name, space.GUID, orgName,
						space.CreatedAt.Format("2006-01-02"),
						space.UpdatedAt.Format("2006-01-02"))
				}

				table.Render()

				if !allPages && spaces.Pagination.TotalPages > 1 {
					fmt.Printf("\nShowing page 1 of %d. Use --all to fetch all pages.\n", spaces.Pagination.TotalPages)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&orgName, "org", "o", "", "filter by organization name")
	cmd.Flags().BoolVar(&allPages, "all", false, "fetch all pages")
	cmd.Flags().IntVar(&perPage, "per-page", 50, "results per page")

	return cmd
}

func newSpacesGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get SPACE_NAME_OR_GUID",
		Short: "Get space details",
		Long:  "Display detailed information about a specific space",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			spacesClient := client.Spaces()

			// Try to get by GUID first
			space, err := spacesClient.Get(ctx, nameOrGUID)
			if err != nil {
				// If not found by GUID, try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)
				spaces, err := spacesClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find space: %w", err)
				}
				if len(spaces.Resources) == 0 {
					return fmt.Errorf("space '%s' not found", nameOrGUID)
				}
				space = &spaces.Resources[0]
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(space)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(space)
			default:
				fmt.Printf("Space: %s\n", space.Name)
				fmt.Printf("  GUID:       %s\n", space.GUID)
				fmt.Printf("  Created:    %s\n", space.CreatedAt.Format("2006-01-02 15:04:05"))
				fmt.Printf("  Updated:    %s\n", space.UpdatedAt.Format("2006-01-02 15:04:05"))

				if space.Relationships.Organization.Data != nil {
					org, _ := client.Organizations().Get(ctx, space.Relationships.Organization.Data.GUID)
					if org != nil {
						fmt.Printf("  Organization: %s (%s)\n", org.Name, org.GUID)
					}
				}

				if space.Metadata.Labels != nil && len(space.Metadata.Labels) > 0 {
					fmt.Println("  Labels:")
					for k, v := range space.Metadata.Labels {
						fmt.Printf("    %s: %s\n", k, v)
					}
				}

				if space.Metadata.Annotations != nil && len(space.Metadata.Annotations) > 0 {
					fmt.Println("  Annotations:")
					for k, v := range space.Metadata.Annotations {
						fmt.Printf("    %s: %s\n", k, v)
					}
				}
			}

			return nil
		},
	}
}

func newSpacesCreateCommand() *cobra.Command {
	var (
		name    string
		orgName string
		labels  map[string]string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new space",
		Long:  "Create a new Cloud Foundry space",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("space name is required")
			}

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find organization
			var orgGUID string
			if orgName != "" {
				params := capi.NewQueryParams()
				params.WithFilter("names", orgName)
				orgs, err := client.Organizations().List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find organization: %w", err)
				}
				if len(orgs.Resources) == 0 {
					return fmt.Errorf("organization '%s' not found", orgName)
				}
				orgGUID = orgs.Resources[0].GUID
			} else {
				return fmt.Errorf("organization is required (use --org)")
			}

			createReq := &capi.SpaceCreateRequest{
				Name: name,
				Relationships: capi.SpaceRelationships{
					Organization: capi.Relationship{
						Data: &capi.RelationshipData{GUID: orgGUID},
					},
				},
			}

			if labels != nil {
				createReq.Metadata = &capi.Metadata{
					Labels: labels,
				}
			}

			space, err := client.Spaces().Create(ctx, createReq)
			if err != nil {
				return fmt.Errorf("failed to create space: %w", err)
			}

			fmt.Printf("Successfully created space '%s' with GUID %s\n", space.Name, space.GUID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "space name (required)")
	cmd.Flags().StringVarP(&orgName, "org", "o", "", "organization name (required)")
	cmd.Flags().StringToStringVar(&labels, "labels", nil, "labels to apply (key=value)")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("org")

	return cmd
}

func newSpacesUpdateCommand() *cobra.Command {
	var (
		newName string
		labels  map[string]string
	)

	cmd := &cobra.Command{
		Use:   "update SPACE_NAME_OR_GUID",
		Short: "Update a space",
		Long:  "Update an existing Cloud Foundry space",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			spacesClient := client.Spaces()

			// Find space
			var spaceGUID string
			space, err := spacesClient.Get(ctx, nameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)
				spaces, err := spacesClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find space: %w", err)
				}
				if len(spaces.Resources) == 0 {
					return fmt.Errorf("space '%s' not found", nameOrGUID)
				}
				spaceGUID = spaces.Resources[0].GUID
			} else {
				spaceGUID = space.GUID
			}

			// Build update request
			updateReq := &capi.SpaceUpdateRequest{}

			if newName != "" {
				updateReq.Name = &newName
			}

			if labels != nil {
				updateReq.Metadata = &capi.Metadata{
					Labels: labels,
				}
			}

			// Update space
			updatedSpace, err := spacesClient.Update(ctx, spaceGUID, updateReq)
			if err != nil {
				return fmt.Errorf("failed to update space: %w", err)
			}

			fmt.Printf("Successfully updated space '%s'\n", updatedSpace.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&newName, "name", "", "new space name")
	cmd.Flags().StringToStringVar(&labels, "labels", nil, "labels to apply (key=value)")

	return cmd
}

func newSpacesDeleteCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete SPACE_NAME_OR_GUID",
		Short: "Delete a space",
		Long:  "Delete a Cloud Foundry space",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			if !force {
				fmt.Printf("Really delete space '%s'? (y/N): ", nameOrGUID)
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
			spacesClient := client.Spaces()

			// Find space
			var spaceGUID string
			var spaceName string
			space, err := spacesClient.Get(ctx, nameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)
				spaces, err := spacesClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find space: %w", err)
				}
				if len(spaces.Resources) == 0 {
					return fmt.Errorf("space '%s' not found", nameOrGUID)
				}
				spaceGUID = spaces.Resources[0].GUID
				spaceName = spaces.Resources[0].Name
			} else {
				spaceGUID = space.GUID
				spaceName = space.Name
			}

			// Delete space
			job, err := spacesClient.Delete(ctx, spaceGUID)
			if err != nil {
				return fmt.Errorf("failed to delete space: %w", err)
			}

			if job != nil {
				fmt.Printf("Deleting space '%s'... (job: %s)\n", spaceName, job.GUID)
			} else {
				fmt.Printf("Successfully deleted space '%s'\n", spaceName)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "force deletion without confirmation")

	return cmd
}

func newSpacesSetQuotaCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set-quota SPACE_NAME_OR_GUID QUOTA_NAME_OR_GUID",
		Short: "Set space quota",
		Long:  "Assign a quota to a space",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement when quota management API is available
			fmt.Println("Space quota management is not yet implemented in the client library")
			return nil
		},
	}
}

func newSpacesListUsersCommand() *cobra.Command {
	var (
		allPages bool
		perPage  int
	)

	cmd := &cobra.Command{
		Use:   "list-users SPACE_NAME_OR_GUID",
		Short: "List users in space",
		Long:  "List all users with roles in the space",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spaceNameOrGUID := args[0]

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			spacesClient := client.Spaces()

			// Find space
			var spaceGUID string
			var spaceName string
			space, err := spacesClient.Get(ctx, spaceNameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", spaceNameOrGUID)
				spaces, err := spacesClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find space: %w", err)
				}
				if len(spaces.Resources) == 0 {
					return fmt.Errorf("space '%s' not found", spaceNameOrGUID)
				}
				spaceGUID = spaces.Resources[0].GUID
				spaceName = spaces.Resources[0].Name
			} else {
				spaceGUID = space.GUID
				spaceName = space.Name
			}

			// Get users with roles in space
			params := capi.NewQueryParams()
			params.PerPage = perPage
			params.WithFilter("space_guids", spaceGUID)

			roles, err := client.Roles().List(ctx, params)
			if err != nil {
				return fmt.Errorf("failed to list space roles: %w", err)
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
					fmt.Printf("No users found in space '%s'\n", spaceName)
					return nil
				}

				fmt.Printf("Users in space '%s':\n\n", spaceName)
				table := tablewriter.NewWriter(os.Stdout)
				table.Header("User GUID", "Role", "Created")

				// Group by user
				userRoles := make(map[string][]string)
				for _, role := range allRoles {
					if role.Relationships.User.Data != nil {
						userGUID := role.Relationships.User.Data.GUID
						userRoles[userGUID] = append(userRoles[userGUID], role.Type)
					}
				}

				for userGUID, roles := range userRoles {
					rolesStr := strings.Join(roles, ", ")
					// Get user details if possible
					user, _ := client.Users().Get(ctx, userGUID)
					username := userGUID
					if user != nil {
						username = user.Username
					}
					table.Append(username, rolesStr, "")
				}

				table.Render()

				if !allPages && roles.Pagination.TotalPages > 1 {
					fmt.Printf("\nShowing page 1 of %d. Use --all to fetch all pages.\n", roles.Pagination.TotalPages)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&allPages, "all", false, "fetch all pages")
	cmd.Flags().IntVar(&perPage, "per-page", 50, "results per page")

	return cmd
}

func newSpacesSetRoleCommand() *cobra.Command {
	var role string

	cmd := &cobra.Command{
		Use:   "set-role SPACE_NAME_OR_GUID USERNAME_OR_GUID",
		Short: "Set user role in space",
		Long:  "Assign a role to a user in a space",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			spaceNameOrGUID := args[0]
			userNameOrGUID := args[1]

			if role == "" {
				role = "space_developer"
			}

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find space
			spacesClient := client.Spaces()
			var spaceGUID string
			space, err := spacesClient.Get(ctx, spaceNameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", spaceNameOrGUID)
				spaces, err := spacesClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find space: %w", err)
				}
				if len(spaces.Resources) == 0 {
					return fmt.Errorf("space '%s' not found", spaceNameOrGUID)
				}
				spaceGUID = spaces.Resources[0].GUID
			} else {
				spaceGUID = space.GUID
			}

			// Find user
			usersClient := client.Users()
			var userGUID string
			user, err := usersClient.Get(ctx, userNameOrGUID)
			if err != nil {
				// Try by username
				params := capi.NewQueryParams()
				params.WithFilter("usernames", userNameOrGUID)
				users, err := usersClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find user: %w", err)
				}
				if len(users.Resources) == 0 {
					return fmt.Errorf("user '%s' not found", userNameOrGUID)
				}
				userGUID = users.Resources[0].GUID
			} else {
				userGUID = user.GUID
			}

			// Create role
			roleReq := &capi.RoleCreateRequest{
				Type: role,
				Relationships: capi.RoleRelationships{
					Space: &capi.Relationship{
						Data: &capi.RelationshipData{GUID: spaceGUID},
					},
					User: capi.Relationship{
						Data: &capi.RelationshipData{GUID: userGUID},
					},
				},
			}

			_, err = client.Roles().Create(ctx, roleReq)
			if err != nil {
				return fmt.Errorf("failed to set user role in space: %w", err)
			}

			fmt.Printf("Successfully set user role '%s' in space\n", role)
			return nil
		},
	}

	cmd.Flags().StringVarP(&role, "role", "r", "space_developer", "role to assign (space_developer, space_manager, space_auditor, space_supporter)")

	return cmd
}

func newSpacesUnsetRoleCommand() *cobra.Command {
	var role string

	cmd := &cobra.Command{
		Use:   "unset-role SPACE_NAME_OR_GUID USERNAME_OR_GUID",
		Short: "Remove user role from space",
		Long:  "Remove a user's role from a space",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			spaceNameOrGUID := args[0]
			userNameOrGUID := args[1]

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find space
			spacesClient := client.Spaces()
			var spaceGUID string
			space, err := spacesClient.Get(ctx, spaceNameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", spaceNameOrGUID)
				spaces, err := spacesClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find space: %w", err)
				}
				if len(spaces.Resources) == 0 {
					return fmt.Errorf("space '%s' not found", spaceNameOrGUID)
				}
				spaceGUID = spaces.Resources[0].GUID
			} else {
				spaceGUID = space.GUID
			}

			// Find user
			usersClient := client.Users()
			var userGUID string
			user, err := usersClient.Get(ctx, userNameOrGUID)
			if err != nil {
				// Try by username
				params := capi.NewQueryParams()
				params.WithFilter("usernames", userNameOrGUID)
				users, err := usersClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find user: %w", err)
				}
				if len(users.Resources) == 0 {
					return fmt.Errorf("user '%s' not found", userNameOrGUID)
				}
				userGUID = users.Resources[0].GUID
			} else {
				userGUID = user.GUID
			}

			// Find and delete role(s)
			rolesClient := client.Roles()
			params := capi.NewQueryParams()
			params.WithFilter("user_guids", userGUID)
			params.WithFilter("space_guids", spaceGUID)
			if role != "" {
				params.WithFilter("types", role)
			}

			roles, err := rolesClient.List(ctx, params)
			if err != nil {
				return fmt.Errorf("failed to list user roles: %w", err)
			}

			if len(roles.Resources) == 0 {
				fmt.Println("No roles found to remove")
				return nil
			}

			// Delete each role
			for _, r := range roles.Resources {
				err = rolesClient.Delete(ctx, r.GUID)
				if err != nil {
					return fmt.Errorf("failed to remove role '%s': %w", r.Type, err)
				}
				fmt.Printf("Removed role '%s'\n", r.Type)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&role, "role", "r", "", "specific role to remove (if not specified, removes all roles)")

	return cmd
}

func newSpacesListAppsCommand() *cobra.Command {
	var (
		allPages bool
		perPage  int
	)

	cmd := &cobra.Command{
		Use:   "list-apps SPACE_NAME_OR_GUID",
		Short: "List applications in space",
		Long:  "List all applications in a space",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spaceNameOrGUID := args[0]

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			spacesClient := client.Spaces()

			// Find space
			var spaceGUID string
			var spaceName string
			space, err := spacesClient.Get(ctx, spaceNameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", spaceNameOrGUID)
				spaces, err := spacesClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find space: %w", err)
				}
				if len(spaces.Resources) == 0 {
					return fmt.Errorf("space '%s' not found", spaceNameOrGUID)
				}
				spaceGUID = spaces.Resources[0].GUID
				spaceName = spaces.Resources[0].Name
			} else {
				spaceGUID = space.GUID
				spaceName = space.Name
			}

			// List apps in space
			params := capi.NewQueryParams()
			params.PerPage = perPage
			params.WithFilter("space_guids", spaceGUID)

			apps, err := client.Apps().List(ctx, params)
			if err != nil {
				return fmt.Errorf("failed to list applications: %w", err)
			}

			// Fetch all pages if requested
			allApps := apps.Resources
			if allPages && apps.Pagination.TotalPages > 1 {
				for page := 2; page <= apps.Pagination.TotalPages; page++ {
					params.Page = page
					moreApps, err := client.Apps().List(ctx, params)
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
					fmt.Printf("No applications found in space '%s'\n", spaceName)
					return nil
				}

				fmt.Printf("Applications in space '%s':\n\n", spaceName)
				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Name", "GUID", "State", "Created", "Updated")

				for _, app := range allApps {
					table.Append(app.Name, app.GUID, app.State,
						app.CreatedAt.Format("2006-01-02"),
						app.UpdatedAt.Format("2006-01-02"))
				}

				table.Render()

				if !allPages && apps.Pagination.TotalPages > 1 {
					fmt.Printf("\nShowing page 1 of %d. Use --all to fetch all pages.\n", apps.Pagination.TotalPages)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&allPages, "all", false, "fetch all pages")
	cmd.Flags().IntVar(&perPage, "per-page", 50, "results per page")

	return cmd
}

func newSpacesListServicesCommand() *cobra.Command {
	var (
		allPages bool
		perPage  int
	)

	cmd := &cobra.Command{
		Use:   "list-services SPACE_NAME_OR_GUID",
		Short: "List service instances in space",
		Long:  "List all service instances in a space",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spaceNameOrGUID := args[0]

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			spacesClient := client.Spaces()

			// Find space
			var spaceGUID string
			var spaceName string
			space, err := spacesClient.Get(ctx, spaceNameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", spaceNameOrGUID)
				spaces, err := spacesClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find space: %w", err)
				}
				if len(spaces.Resources) == 0 {
					return fmt.Errorf("space '%s' not found", spaceNameOrGUID)
				}
				spaceGUID = spaces.Resources[0].GUID
				spaceName = spaces.Resources[0].Name
			} else {
				spaceGUID = space.GUID
				spaceName = space.Name
			}

			// List service instances in space
			params := capi.NewQueryParams()
			params.PerPage = perPage
			params.WithFilter("space_guids", spaceGUID)

			services, err := client.ServiceInstances().List(ctx, params)
			if err != nil {
				return fmt.Errorf("failed to list service instances: %w", err)
			}

			// Fetch all pages if requested
			allServices := services.Resources
			if allPages && services.Pagination.TotalPages > 1 {
				for page := 2; page <= services.Pagination.TotalPages; page++ {
					params.Page = page
					moreServices, err := client.ServiceInstances().List(ctx, params)
					if err != nil {
						return fmt.Errorf("failed to fetch page %d: %w", page, err)
					}
					allServices = append(allServices, moreServices.Resources...)
				}
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(allServices)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(allServices)
			default:
				if len(allServices) == 0 {
					fmt.Printf("No service instances found in space '%s'\n", spaceName)
					return nil
				}

				fmt.Printf("Service instances in space '%s':\n\n", spaceName)
				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Name", "GUID", "Type", "State", "Created")

				for _, service := range allServices {
					state := "ready"
					if service.LastOperation != nil {
						state = service.LastOperation.State
					}
					table.Append(service.Name, service.GUID, service.Type, state,
						service.CreatedAt.Format("2006-01-02"))
				}

				table.Render()

				if !allPages && services.Pagination.TotalPages > 1 {
					fmt.Printf("\nShowing page 1 of %d. Use --all to fetch all pages.\n", services.Pagination.TotalPages)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&allPages, "all", false, "fetch all pages")
	cmd.Flags().IntVar(&perPage, "per-page", 50, "results per page")

	return cmd
}

func newSpacesApplyManifestCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "apply-manifest SPACE_NAME_OR_GUID MANIFEST_FILE",
		Short: "Apply manifest to space",
		Long:  "Apply an application manifest to a space",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement when manifest API is available
			fmt.Println("Manifest application is not yet implemented in the client library")
			return nil
		},
	}
}
