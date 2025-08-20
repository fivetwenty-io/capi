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
			client, err := createClient()
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
					table.Append(org.Name, org.GUID, status,
						org.CreatedAt.Format("2006-01-02"),
						org.UpdatedAt.Format("2006-01-02"))
				}

				table.Render()

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

			client, err := createClient()
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
				fmt.Printf("Organization: %s\n", org.Name)
				fmt.Printf("  GUID:       %s\n", org.GUID)
				fmt.Printf("  Status:     %s\n", func() string {
					if org.Suspended {
						return "suspended"
					}
					return "active"
				}())
				fmt.Printf("  Created:    %s\n", org.CreatedAt.Format("2006-01-02 15:04:05"))
				fmt.Printf("  Updated:    %s\n", org.UpdatedAt.Format("2006-01-02 15:04:05"))

				if org.Metadata.Labels != nil && len(org.Metadata.Labels) > 0 {
					fmt.Println("  Labels:")
					for k, v := range org.Metadata.Labels {
						fmt.Printf("    %s: %s\n", k, v)
					}
				}

				if org.Metadata.Annotations != nil && len(org.Metadata.Annotations) > 0 {
					fmt.Println("  Annotations:")
					for k, v := range org.Metadata.Annotations {
						fmt.Printf("    %s: %s\n", k, v)
					}
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

			client, err := createClient()
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
	cmd.MarkFlagRequired("name")

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

			client, err := createClient()
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
