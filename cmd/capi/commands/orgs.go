package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fivetwenty-io/capi/pkg/capi"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// NewOrgsCommand creates the organizations command group
func NewOrgsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "orgs",
		Aliases: []string{"organizations", "org"},
		Short:   "Manage organizations",
		Long:    "List, create, update, and delete Cloud Foundry organizations",
	}

	cmd.AddCommand(newOrgsListCommand())
	cmd.AddCommand(newOrgsGetCommand())
	cmd.AddCommand(newOrgsCreateCommand())
	cmd.AddCommand(newOrgsUpdateCommand())
	cmd.AddCommand(newOrgsDeleteCommand())
	cmd.AddCommand(newOrgsSetQuotaCommand())
	cmd.AddCommand(newOrgsListUsersCommand())
	cmd.AddCommand(newOrgsAddUserCommand())
	cmd.AddCommand(newOrgsRemoveUserCommand())
	cmd.AddCommand(newOrgsListSpacesCommand())

	return cmd
}

func newOrgsListCommand() *cobra.Command {
	var (
		allPages bool
		perPage  int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List organizations",
		Long:  "List all organizations the user has access to",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			params := capi.NewQueryParams()
			params.PerPage = perPage

			orgs, err := client.Organizations().List(ctx, params)
			if err != nil {
				return fmt.Errorf("failed to list organizations: %w", err)
			}

			// Fetch all pages if requested
			allOrgs := orgs.Resources
			if allPages && orgs.Pagination.TotalPages > 1 {
				for page := 2; page <= orgs.Pagination.TotalPages; page++ {
					params.Page = page
					moreOrgs, err := client.Organizations().List(ctx, params)
					if err != nil {
						return fmt.Errorf("failed to fetch page %d: %w", page, err)
					}
					allOrgs = append(allOrgs, moreOrgs.Resources...)
				}
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(allOrgs)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(allOrgs)
			default:
				if len(allOrgs) == 0 {
					fmt.Println("No organizations found")
					return nil
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Name", "GUID", "Status", "Created", "Updated")

				for _, org := range allOrgs {
					status := "active"
					if org.Suspended {
						status = "suspended"
					}
					_ = table.Append(org.Name, org.GUID, status,
						org.CreatedAt.Format("2006-01-02"),
						org.UpdatedAt.Format("2006-01-02"))
				}

				_ = table.Render()

				if !allPages && orgs.Pagination.TotalPages > 1 {
					fmt.Printf("\nShowing page 1 of %d. Use --all to fetch all pages.\n", orgs.Pagination.TotalPages)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&allPages, "all", false, "fetch all pages")
	cmd.Flags().IntVar(&perPage, "per-page", 50, "results per page")

	return cmd
}

func newOrgsGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get ORG_NAME_OR_GUID",
		Short: "Get organization details",
		Long:  "Display detailed information about a specific organization",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			orgsClient := client.Organizations()

			// Try to get by GUID first
			org, err := orgsClient.Get(ctx, nameOrGUID)
			if err != nil {
				// If not found by GUID, try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)
				orgs, err := orgsClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find organization: %w", err)
				}
				if len(orgs.Resources) == 0 {
					return fmt.Errorf("organization '%s' not found", nameOrGUID)
				}
				org = &orgs.Resources[0]
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(org)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(org)
			default:
				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Property", "Value")

				_ = table.Append("Name", org.Name)
				_ = table.Append("GUID", org.GUID)
				status := "active"
				if org.Suspended {
					status = "suspended"
				}
				_ = table.Append("Status", status)
				_ = table.Append("Created", org.CreatedAt.Format("2006-01-02 15:04:05"))
				_ = table.Append("Updated", org.UpdatedAt.Format("2006-01-02 15:04:05"))

				fmt.Printf("Organization details:\n\n")
				_ = table.Render()

				if len(org.Metadata.Labels) > 0 {
					fmt.Println("\nLabels:")
					labelTable := tablewriter.NewWriter(os.Stdout)
					labelTable.Header("Key", "Value")
					for k, v := range org.Metadata.Labels {
						_ = labelTable.Append(k, v)
					}
					_ = labelTable.Render()
				}

				if len(org.Metadata.Annotations) > 0 {
					fmt.Println("\nAnnotations:")
					annotationTable := tablewriter.NewWriter(os.Stdout)
					annotationTable.Header("Key", "Value")
					for k, v := range org.Metadata.Annotations {
						_ = annotationTable.Append(k, v)
					}
					_ = annotationTable.Render()
				}
			}

			return nil
		},
	}
}

func newOrgsCreateCommand() *cobra.Command {
	var (
		name   string
		labels map[string]string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new organization",
		Long:  "Create a new Cloud Foundry organization",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("organization name is required")
			}

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			createReq := &capi.OrganizationCreateRequest{
				Name: name,
			}

			if labels != nil {
				createReq.Metadata = &capi.Metadata{
					Labels: labels,
				}
			}

			org, err := client.Organizations().Create(ctx, createReq)
			if err != nil {
				return fmt.Errorf("failed to create organization: %w", err)
			}

			fmt.Printf("Successfully created organization '%s' with GUID %s\n", org.Name, org.GUID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "organization name (required)")
	cmd.Flags().StringToStringVar(&labels, "labels", nil, "labels to apply (key=value)")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newOrgsUpdateCommand() *cobra.Command {
	var (
		newName   string
		suspended bool
		labels    map[string]string
	)

	cmd := &cobra.Command{
		Use:   "update ORG_NAME_OR_GUID",
		Short: "Update an organization",
		Long:  "Update an existing Cloud Foundry organization",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			orgsClient := client.Organizations()

			// Find organization
			var orgGUID string
			org, err := orgsClient.Get(ctx, nameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)
				orgs, err := orgsClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find organization: %w", err)
				}
				if len(orgs.Resources) == 0 {
					return fmt.Errorf("organization '%s' not found", nameOrGUID)
				}
				orgGUID = orgs.Resources[0].GUID
			} else {
				orgGUID = org.GUID
			}

			// Build update request
			updateReq := &capi.OrganizationUpdateRequest{}

			if newName != "" {
				updateReq.Name = &newName
			}

			if cmd.Flags().Changed("suspended") {
				updateReq.Suspended = &suspended
			}

			if labels != nil {
				updateReq.Metadata = &capi.Metadata{
					Labels: labels,
				}
			}

			// Update organization
			updatedOrg, err := orgsClient.Update(ctx, orgGUID, updateReq)
			if err != nil {
				return fmt.Errorf("failed to update organization: %w", err)
			}

			fmt.Printf("Successfully updated organization '%s'\n", updatedOrg.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&newName, "name", "", "new organization name")
	cmd.Flags().BoolVar(&suspended, "suspended", false, "suspend the organization")
	cmd.Flags().StringToStringVar(&labels, "labels", nil, "labels to apply (key=value)")

	return cmd
}

func newOrgsDeleteCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete ORG_NAME_OR_GUID",
		Short: "Delete an organization",
		Long:  "Delete a Cloud Foundry organization",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			if !force {
				fmt.Printf("Really delete organization '%s'? (y/N): ", nameOrGUID)
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
			orgsClient := client.Organizations()

			// Find organization
			var orgGUID string
			var orgName string
			org, err := orgsClient.Get(ctx, nameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)
				orgs, err := orgsClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find organization: %w", err)
				}
				if len(orgs.Resources) == 0 {
					return fmt.Errorf("organization '%s' not found", nameOrGUID)
				}
				orgGUID = orgs.Resources[0].GUID
				orgName = orgs.Resources[0].Name
			} else {
				orgGUID = org.GUID
				orgName = org.Name
			}

			// Delete organization
			job, err := orgsClient.Delete(ctx, orgGUID)
			if err != nil {
				return fmt.Errorf("failed to delete organization: %w", err)
			}

			if job != nil {
				fmt.Printf("Deleting organization '%s'... (job: %s)\n", orgName, job.GUID)
				// Could poll job for completion here
			} else {
				fmt.Printf("Successfully deleted organization '%s'\n", orgName)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "force deletion without confirmation")

	return cmd
}

func newOrgsSetQuotaCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set-quota ORG_NAME_OR_GUID QUOTA_NAME_OR_GUID",
		Short: "Set organization quota",
		Long:  "Assign a quota to an organization",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			orgNameOrGUID := args[0]
			quotaNameOrGUID := args[1]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			quotaClient := client.OrganizationQuotas()
			orgsClient := client.Organizations()

			// Find organization
			var orgGUID string
			var orgName string
			org, err := orgsClient.Get(ctx, orgNameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", orgNameOrGUID)
				orgs, err := orgsClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find organization '%s': %w", orgNameOrGUID, err)
				}
				if len(orgs.Resources) == 0 {
					return fmt.Errorf("organization '%s' not found", orgNameOrGUID)
				}
				orgGUID = orgs.Resources[0].GUID
				orgName = orgs.Resources[0].Name
			} else {
				orgGUID = org.GUID
				orgName = org.Name
			}

			// Find quota
			var quotaGUID string
			var quotaName string
			quota, err := quotaClient.Get(ctx, quotaNameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", quotaNameOrGUID)
				quotas, err := quotaClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find organization quota '%s': %w", quotaNameOrGUID, err)
				}
				if len(quotas.Resources) == 0 {
					return fmt.Errorf("organization quota '%s' not found", quotaNameOrGUID)
				}
				quotaGUID = quotas.Resources[0].GUID
				quotaName = quotas.Resources[0].Name
			} else {
				quotaGUID = quota.GUID
				quotaName = quota.Name
			}

			// Apply quota to organization
			_, err = quotaClient.ApplyToOrganizations(ctx, quotaGUID, []string{orgGUID})
			if err != nil {
				return fmt.Errorf("failed to set quota for organization: %w", err)
			}

			type QuotaAssignmentResult struct {
				OrganizationName string `json:"organization_name" yaml:"organization_name"`
				OrganizationGUID string `json:"organization_guid" yaml:"organization_guid"`
				QuotaName        string `json:"quota_name" yaml:"quota_name"`
				QuotaGUID        string `json:"quota_guid" yaml:"quota_guid"`
				Operation        string `json:"operation" yaml:"operation"`
			}

			result := QuotaAssignmentResult{
				OrganizationName: orgName,
				OrganizationGUID: orgGUID,
				QuotaName:        quotaName,
				QuotaGUID:        quotaGUID,
				Operation:        "set-quota",
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(result)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(result)
			default:
				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Property", "Value")
				_ = table.Append("Organization", orgName)
				_ = table.Append("Quota", quotaName)
				_ = table.Append("Operation", "set-quota")

				fmt.Printf("Successfully set quota for organization:\n\n")
				_ = table.Render()
			}

			return nil
		},
	}
}

func newOrgsListUsersCommand() *cobra.Command {
	var (
		allPages bool
		perPage  int
	)

	cmd := &cobra.Command{
		Use:   "list-users ORG_NAME_OR_GUID",
		Short: "List users in organization",
		Long:  "List all users with roles in the organization",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			orgNameOrGUID := args[0]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			orgsClient := client.Organizations()

			// Find organization
			var orgGUID string
			var orgName string
			org, err := orgsClient.Get(ctx, orgNameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", orgNameOrGUID)
				orgs, err := orgsClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find organization: %w", err)
				}
				if len(orgs.Resources) == 0 {
					return fmt.Errorf("organization '%s' not found", orgNameOrGUID)
				}
				orgGUID = orgs.Resources[0].GUID
				orgName = orgs.Resources[0].Name
			} else {
				orgGUID = org.GUID
				orgName = org.Name
			}

			// Get users with roles in organization
			params := capi.NewQueryParams()
			params.PerPage = perPage

			users, err := orgsClient.ListUsers(ctx, orgGUID, params)
			if err != nil {
				return fmt.Errorf("failed to list users: %w", err)
			}

			// Fetch all pages if requested
			allUsers := users.Resources
			if allPages && users.Pagination.TotalPages > 1 {
				for page := 2; page <= users.Pagination.TotalPages; page++ {
					params.Page = page
					moreUsers, err := orgsClient.ListUsers(ctx, orgGUID, params)
					if err != nil {
						return fmt.Errorf("failed to fetch page %d: %w", page, err)
					}
					allUsers = append(allUsers, moreUsers.Resources...)
				}
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(allUsers)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(allUsers)
			default:
				if len(allUsers) == 0 {
					fmt.Printf("No users found in organization '%s'\n", orgName)
					return nil
				}

				fmt.Printf("Users in organization '%s':\n\n", orgName)
				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Username", "GUID", "Roles", "Created")

				for _, user := range allUsers {
					// Get roles for this user
					rolesParams := capi.NewQueryParams()
					rolesParams.WithFilter("user_guids", user.GUID)
					rolesParams.WithFilter("organization_guids", orgGUID)
					roles, _ := client.Roles().List(ctx, rolesParams)

					roleNames := []string{}
					for _, role := range roles.Resources {
						roleNames = append(roleNames, role.Type)
					}
					rolesStr := "none"
					if len(roleNames) > 0 {
						rolesStr = strings.Join(roleNames, ", ")
					}

					_ = table.Append(user.Username, user.GUID, rolesStr,
						user.CreatedAt.Format("2006-01-02"))
				}

				_ = table.Render()

				if !allPages && users.Pagination.TotalPages > 1 {
					fmt.Printf("\nShowing page 1 of %d. Use --all to fetch all pages.\n", users.Pagination.TotalPages)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&allPages, "all", false, "fetch all pages")
	cmd.Flags().IntVar(&perPage, "per-page", 50, "results per page")

	return cmd
}

func newOrgsAddUserCommand() *cobra.Command {
	var role string

	cmd := &cobra.Command{
		Use:   "add-user ORG_NAME_OR_GUID USERNAME_OR_GUID",
		Short: "Add user to organization",
		Long:  "Add a user to an organization with a specific role",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			orgNameOrGUID := args[0]
			userNameOrGUID := args[1]

			if role == "" {
				role = "organization_user"
			}

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find organization
			orgsClient := client.Organizations()
			var orgGUID string
			org, err := orgsClient.Get(ctx, orgNameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", orgNameOrGUID)
				orgs, err := orgsClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find organization: %w", err)
				}
				if len(orgs.Resources) == 0 {
					return fmt.Errorf("organization '%s' not found", orgNameOrGUID)
				}
				orgGUID = orgs.Resources[0].GUID
			} else {
				orgGUID = org.GUID
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
					Organization: &capi.Relationship{
						Data: &capi.RelationshipData{GUID: orgGUID},
					},
					User: capi.Relationship{
						Data: &capi.RelationshipData{GUID: userGUID},
					},
				},
			}

			_, err = client.Roles().Create(ctx, roleReq)
			if err != nil {
				return fmt.Errorf("failed to add user to organization: %w", err)
			}

			fmt.Printf("Successfully added user to organization with role '%s'\n", role)
			return nil
		},
	}

	cmd.Flags().StringVarP(&role, "role", "r", "organization_user", "role to assign (organization_user, organization_manager, organization_auditor, organization_billing_manager)")

	return cmd
}

func newOrgsRemoveUserCommand() *cobra.Command {
	var role string

	cmd := &cobra.Command{
		Use:   "remove-user ORG_NAME_OR_GUID USERNAME_OR_GUID",
		Short: "Remove user from organization",
		Long:  "Remove a user's role from an organization",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			orgNameOrGUID := args[0]
			userNameOrGUID := args[1]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find organization
			orgsClient := client.Organizations()
			var orgGUID string
			org, err := orgsClient.Get(ctx, orgNameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", orgNameOrGUID)
				orgs, err := orgsClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find organization: %w", err)
				}
				if len(orgs.Resources) == 0 {
					return fmt.Errorf("organization '%s' not found", orgNameOrGUID)
				}
				orgGUID = orgs.Resources[0].GUID
			} else {
				orgGUID = org.GUID
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
			params.WithFilter("organization_guids", orgGUID)
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

func newOrgsListSpacesCommand() *cobra.Command {
	var (
		allPages bool
		perPage  int
	)

	cmd := &cobra.Command{
		Use:   "list-spaces ORG_NAME_OR_GUID",
		Short: "List spaces in organization",
		Long:  "List all spaces within an organization",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			orgNameOrGUID := args[0]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			orgsClient := client.Organizations()

			// Find organization
			var orgGUID string
			var orgName string
			org, err := orgsClient.Get(ctx, orgNameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", orgNameOrGUID)
				orgs, err := orgsClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find organization: %w", err)
				}
				if len(orgs.Resources) == 0 {
					return fmt.Errorf("organization '%s' not found", orgNameOrGUID)
				}
				orgGUID = orgs.Resources[0].GUID
				orgName = orgs.Resources[0].Name
			} else {
				orgGUID = org.GUID
				orgName = org.Name
			}

			// List spaces in organization
			params := capi.NewQueryParams()
			params.PerPage = perPage
			params.WithFilter("organization_guids", orgGUID)

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
					fmt.Printf("No spaces found in organization '%s'\n", orgName)
					return nil
				}

				fmt.Printf("Spaces in organization '%s':\n\n", orgName)
				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Name", "GUID", "Created", "Updated")

				for _, space := range allSpaces {
					_ = table.Append(space.Name, space.GUID,
						space.CreatedAt.Format("2006-01-02"),
						space.UpdatedAt.Format("2006-01-02"))
				}

				_ = table.Render()

				if !allPages && spaces.Pagination.TotalPages > 1 {
					fmt.Printf("\nShowing page 1 of %d. Use --all to fetch all pages.\n", spaces.Pagination.TotalPages)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&allPages, "all", false, "fetch all pages")
	cmd.Flags().IntVar(&perPage, "per-page", 50, "results per page")

	return cmd
}
