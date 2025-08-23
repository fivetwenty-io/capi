package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fivetwenty-io/capi/v3/pkg/capi"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// NewDomainsCommand creates the domains command group
func NewDomainsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "domains",
		Aliases: []string{"domain"},
		Short:   "Manage domains",
		Long:    "List and manage Cloud Foundry domains",
	}

	cmd.AddCommand(newDomainsListCommand())
	cmd.AddCommand(newDomainsGetCommand())
	cmd.AddCommand(newDomainsCreateCommand())
	cmd.AddCommand(newDomainsUpdateCommand())
	cmd.AddCommand(newDomainsDeleteCommand())
	cmd.AddCommand(newDomainsShareCommand())
	cmd.AddCommand(newDomainsUnshareCommand())

	return cmd
}

func newDomainsListCommand() *cobra.Command {
	var (
		allPages bool
		perPage  int
		orgName  string
		internal bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List domains",
		Long:  "List all domains the user has access to",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			params := capi.NewQueryParams()
			params.PerPage = perPage

			// Filter by organization if specified
			if orgName != "" {
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
			}

			// Filter by internal/external if specified
			if cmd.Flags().Changed("internal") {
				if internal {
					params.WithFilter("internal", "true")
				} else {
					params.WithFilter("internal", "false")
				}
			}

			domains, err := client.Domains().List(ctx, params)
			if err != nil {
				return fmt.Errorf("failed to list domains: %w", err)
			}

			// Fetch all pages if requested
			allDomains := domains.Resources
			if allPages && domains.Pagination.TotalPages > 1 {
				for page := 2; page <= domains.Pagination.TotalPages; page++ {
					params.Page = page
					moreDomains, err := client.Domains().List(ctx, params)
					if err != nil {
						return fmt.Errorf("failed to fetch page %d: %w", page, err)
					}
					allDomains = append(allDomains, moreDomains.Resources...)
				}
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(allDomains)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(allDomains)
			default:
				if len(allDomains) == 0 {
					fmt.Println("No domains found")
					return nil
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Name", "Type", "Protocols", "Router Group", "Created")

				for _, domain := range allDomains {
					domainType := "shared"
					if domain.Internal {
						domainType = "internal"
					}

					protocols := strings.Join(domain.SupportedProtocols, ", ")

					routerGroup := "default"
					if domain.RouterGroup != nil {
						routerGroup = *domain.RouterGroup
					}

					_ = table.Append(domain.Name, domainType, protocols, routerGroup, domain.CreatedAt.Format("2006-01-02"))
				}

				_ = table.Render()

				if !allPages && domains.Pagination.TotalPages > 1 {
					fmt.Printf("\nShowing page 1 of %d. Use --all to fetch all pages.\n", domains.Pagination.TotalPages)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&orgName, "org", "o", "", "filter by organization name")
	cmd.Flags().BoolVar(&allPages, "all", false, "fetch all pages")
	cmd.Flags().IntVar(&perPage, "per-page", 50, "results per page")
	cmd.Flags().BoolVar(&internal, "internal", false, "filter by internal domains")

	return cmd
}

func newDomainsGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get DOMAIN_NAME_OR_GUID",
		Short: "Get domain details",
		Long:  "Display detailed information about a specific domain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Try to get by GUID first
			domain, err := client.Domains().Get(ctx, nameOrGUID)
			if err != nil {
				// If not found by GUID, try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)
				domains, err := client.Domains().List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find domain: %w", err)
				}
				if len(domains.Resources) == 0 {
					return fmt.Errorf("domain '%s' not found", nameOrGUID)
				}
				domain = &domains.Resources[0]
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(domain)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(domain)
			default:
				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Property", "Value")

				_ = table.Append("Name", domain.Name)
				_ = table.Append("GUID", domain.GUID)
				_ = table.Append("Internal", fmt.Sprintf("%t", domain.Internal))
				_ = table.Append("Protocols", strings.Join(domain.SupportedProtocols, ", "))
				_ = table.Append("Created", domain.CreatedAt.Format("2006-01-02 15:04:05"))
				_ = table.Append("Updated", domain.UpdatedAt.Format("2006-01-02 15:04:05"))

				if domain.RouterGroup != nil {
					_ = table.Append("Router Group", *domain.RouterGroup)
				}

				// Organization info
				if domain.Relationships.Organization != nil && domain.Relationships.Organization.Data != nil {
					org, _ := client.Organizations().Get(ctx, domain.Relationships.Organization.Data.GUID)
					if org != nil {
						_ = table.Append("Organization", org.Name)
					}
				} else {
					_ = table.Append("Organization", "shared (platform)")
				}

				// Shared organizations
				if domain.Relationships.SharedOrganizations != nil && len(domain.Relationships.SharedOrganizations.Data) > 0 {
					var sharedOrgs []string
					for _, orgData := range domain.Relationships.SharedOrganizations.Data {
						org, _ := client.Organizations().Get(ctx, orgData.GUID)
						if org != nil {
							sharedOrgs = append(sharedOrgs, org.Name)
						}
					}
					if len(sharedOrgs) > 0 {
						_ = table.Append("Shared With", strings.Join(sharedOrgs, ", "))
					}
				}

				fmt.Printf("Domain: %s\n\n", domain.Name)
				_ = table.Render()
			}

			return nil
		},
	}
}

func newDomainsCreateCommand() *cobra.Command {
	var (
		name        string
		internal    bool
		routerGroup string
		orgName     string
		labels      map[string]string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a domain",
		Long:  "Create a new Cloud Foundry domain",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("domain name is required")
			}

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			createReq := &capi.DomainCreateRequest{
				Name:     name,
				Internal: &internal,
			}

			if routerGroup != "" {
				createReq.RouterGroup = &routerGroup
			}

			// Add organization relationship if specified
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

				createReq.Relationships = &capi.DomainRelationships{
					Organization: &capi.Relationship{
						Data: &capi.RelationshipData{GUID: orgs.Resources[0].GUID},
					},
				}
			}

			if labels != nil {
				createReq.Metadata = &capi.Metadata{
					Labels: labels,
				}
			}

			domain, err := client.Domains().Create(ctx, createReq)
			if err != nil {
				return fmt.Errorf("failed to create domain: %w", err)
			}

			fmt.Printf("Successfully created domain '%s'\n", domain.Name)
			fmt.Printf("  GUID: %s\n", domain.GUID)
			fmt.Printf("  Type: %s\n", func() string {
				if domain.Internal {
					return "internal"
				}
				return "shared"
			}())

			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "domain name (required)")
	cmd.Flags().BoolVar(&internal, "internal", false, "create as internal domain")
	cmd.Flags().StringVar(&routerGroup, "router-group", "", "router group for TCP domains")
	cmd.Flags().StringVarP(&orgName, "org", "o", "", "organization name for private domains")
	cmd.Flags().StringToStringVar(&labels, "labels", nil, "labels to apply (key=value)")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newDomainsUpdateCommand() *cobra.Command {
	var labels map[string]string

	cmd := &cobra.Command{
		Use:   "update DOMAIN_NAME_OR_GUID",
		Short: "Update a domain",
		Long:  "Update domain metadata",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find domain
			var domainGUID string
			domain, err := client.Domains().Get(ctx, nameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)
				domains, err := client.Domains().List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find domain: %w", err)
				}
				if len(domains.Resources) == 0 {
					return fmt.Errorf("domain '%s' not found", nameOrGUID)
				}
				domain = &domains.Resources[0]
			}
			domainGUID = domain.GUID

			// Build update request
			updateReq := &capi.DomainUpdateRequest{}

			if labels != nil {
				updateReq.Metadata = &capi.Metadata{
					Labels: labels,
				}
			}

			updatedDomain, err := client.Domains().Update(ctx, domainGUID, updateReq)
			if err != nil {
				return fmt.Errorf("failed to update domain: %w", err)
			}

			fmt.Printf("Successfully updated domain '%s'\n", updatedDomain.Name)

			return nil
		},
	}

	cmd.Flags().StringToStringVar(&labels, "labels", nil, "labels to apply (key=value)")

	return cmd
}

func newDomainsDeleteCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete DOMAIN_NAME_OR_GUID",
		Short: "Delete a domain",
		Long:  "Delete a Cloud Foundry domain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			if !force {
				fmt.Printf("Really delete domain '%s'? (y/N): ", nameOrGUID)
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

			// Find domain
			var domainGUID string
			var domainName string
			domain, err := client.Domains().Get(ctx, nameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)
				domains, err := client.Domains().List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find domain: %w", err)
				}
				if len(domains.Resources) == 0 {
					return fmt.Errorf("domain '%s' not found", nameOrGUID)
				}
				domain = &domains.Resources[0]
			}
			domainGUID = domain.GUID
			domainName = domain.Name

			job, err := client.Domains().Delete(ctx, domainGUID)
			if err != nil {
				return fmt.Errorf("failed to delete domain: %w", err)
			}

			if job != nil {
				fmt.Printf("Deleting domain '%s'... (job: %s)\n", domainName, job.GUID)
				fmt.Printf("Monitor with: capi jobs get %s\n", job.GUID)
			} else {
				fmt.Printf("Successfully deleted domain '%s'\n", domainName)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "force deletion without confirmation")

	return cmd
}

func newDomainsShareCommand() *cobra.Command {
	var orgNames []string

	cmd := &cobra.Command{
		Use:   "share DOMAIN_NAME_OR_GUID",
		Short: "Share a domain with organizations",
		Long:  "Share a private domain with specified organizations",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			if len(orgNames) == 0 {
				return fmt.Errorf("at least one organization must be specified")
			}

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find domain
			var domainGUID string
			domain, err := client.Domains().Get(ctx, nameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)
				domains, err := client.Domains().List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find domain: %w", err)
				}
				if len(domains.Resources) == 0 {
					return fmt.Errorf("domain '%s' not found", nameOrGUID)
				}
				domain = &domains.Resources[0]
			}
			domainGUID = domain.GUID

			// Find organizations to share with
			var orgGUIDs []string
			for _, orgName := range orgNames {
				params := capi.NewQueryParams()
				params.WithFilter("names", orgName)
				orgs, err := client.Organizations().List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find organization '%s': %w", orgName, err)
				}
				if len(orgs.Resources) == 0 {
					return fmt.Errorf("organization '%s' not found", orgName)
				}
				orgGUIDs = append(orgGUIDs, orgs.Resources[0].GUID)
			}

			// Share domain with organizations
			_, err = client.Domains().ShareWithOrganization(ctx, domainGUID, orgGUIDs)
			if err != nil {
				return fmt.Errorf("failed to share domain: %w", err)
			}

			fmt.Printf("Successfully shared domain '%s' with organizations: %s\n", domain.Name, strings.Join(orgNames, ", "))

			return nil
		},
	}

	cmd.Flags().StringArrayVarP(&orgNames, "orgs", "o", nil, "organizations to share with (required)")
	_ = cmd.MarkFlagRequired("orgs")

	return cmd
}

func newDomainsUnshareCommand() *cobra.Command {
	var orgName string

	cmd := &cobra.Command{
		Use:   "unshare DOMAIN_NAME_OR_GUID",
		Short: "Unshare a domain from an organization",
		Long:  "Remove sharing of a domain from a specified organization",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			if orgName == "" {
				return fmt.Errorf("organization name is required")
			}

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find domain
			var domainGUID string
			domain, err := client.Domains().Get(ctx, nameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)
				domains, err := client.Domains().List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find domain: %w", err)
				}
				if len(domains.Resources) == 0 {
					return fmt.Errorf("domain '%s' not found", nameOrGUID)
				}
				domain = &domains.Resources[0]
			}
			domainGUID = domain.GUID

			// Find organization to unshare from
			params := capi.NewQueryParams()
			params.WithFilter("names", orgName)
			orgs, err := client.Organizations().List(ctx, params)
			if err != nil {
				return fmt.Errorf("failed to find organization: %w", err)
			}
			if len(orgs.Resources) == 0 {
				return fmt.Errorf("organization '%s' not found", orgName)
			}
			orgGUID := orgs.Resources[0].GUID

			// Unshare domain from organization
			err = client.Domains().UnshareFromOrganization(ctx, domainGUID, orgGUID)
			if err != nil {
				return fmt.Errorf("failed to unshare domain: %w", err)
			}

			fmt.Printf("Successfully unshared domain '%s' from organization '%s'\n", domain.Name, orgName)

			return nil
		},
	}

	cmd.Flags().StringVarP(&orgName, "org", "o", "", "organization to unshare from (required)")
	_ = cmd.MarkFlagRequired("org")

	return cmd
}
