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

// NewOrgQuotasCommand creates the organization quotas command group
func NewOrgQuotasCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "org-quotas",
		Aliases: []string{"organization-quotas", "org-quota", "quotas"},
		Short:   "Manage organization quotas",
		Long:    "List, create, update, delete, and apply organization quotas",
	}

	cmd.AddCommand(newOrgQuotasListCommand())
	cmd.AddCommand(newOrgQuotasGetCommand())
	cmd.AddCommand(newOrgQuotasCreateCommand())
	cmd.AddCommand(newOrgQuotasUpdateCommand())
	cmd.AddCommand(newOrgQuotasDeleteCommand())
	cmd.AddCommand(newOrgQuotasApplyCommand())

	return cmd
}

func newOrgQuotasListCommand() *cobra.Command {
	var (
		allPages bool
		perPage  int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List organization quotas",
		Long:  "List all organization quotas",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			params := capi.NewQueryParams()
			params.PerPage = perPage

			quotas, err := client.OrganizationQuotas().List(ctx, params)
			if err != nil {
				return fmt.Errorf("failed to list organization quotas: %w", err)
			}

			// Fetch all pages if requested
			allQuotas := quotas.Resources
			if allPages && quotas.Pagination.TotalPages > 1 {
				for page := 2; page <= quotas.Pagination.TotalPages; page++ {
					params.Page = page
					moreQuotas, err := client.OrganizationQuotas().List(ctx, params)
					if err != nil {
						return fmt.Errorf("failed to fetch page %d: %w", page, err)
					}
					allQuotas = append(allQuotas, moreQuotas.Resources...)
				}
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(allQuotas)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(allQuotas)
			default:
				if len(allQuotas) == 0 {
					fmt.Println("No organization quotas found")
					return nil
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Name", "GUID", "Total Memory", "Services", "Routes", "Created")

				for _, quota := range allQuotas {
					memoryStr := "unlimited"
					if quota.Apps != nil && quota.Apps.TotalMemoryInMB != nil {
						memoryStr = fmt.Sprintf("%d MB", *quota.Apps.TotalMemoryInMB)
					}

					servicesStr := "unlimited"
					if quota.Services != nil && quota.Services.TotalServiceInstances != nil {
						servicesStr = fmt.Sprintf("%d", *quota.Services.TotalServiceInstances)
					}

					routesStr := "unlimited"
					if quota.Routes != nil && quota.Routes.TotalRoutes != nil {
						routesStr = fmt.Sprintf("%d", *quota.Routes.TotalRoutes)
					}

					_ = table.Append(quota.Name, quota.GUID, memoryStr, servicesStr, routesStr,
						quota.CreatedAt.Format("2006-01-02"))
				}

				_ = table.Render()

				if !allPages && quotas.Pagination.TotalPages > 1 {
					fmt.Printf("\nShowing page 1 of %d. Use --all to fetch all pages.\n", quotas.Pagination.TotalPages)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&allPages, "all", false, "fetch all pages")
	cmd.Flags().IntVar(&perPage, "per-page", 50, "results per page")

	return cmd
}

func newOrgQuotasGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get QUOTA_NAME_OR_GUID",
		Short: "Get organization quota details",
		Long:  "Display detailed information about a specific organization quota",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			quotaClient := client.OrganizationQuotas()

			// Try to get by GUID first
			quota, err := quotaClient.Get(ctx, nameOrGUID)
			if err != nil {
				// If not found by GUID, try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)
				quotas, err := quotaClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find organization quota: %w", err)
				}
				if len(quotas.Resources) == 0 {
					return fmt.Errorf("organization quota '%s' not found", nameOrGUID)
				}
				quota = &quotas.Resources[0]
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(quota)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(quota)
			default:
				fmt.Printf("Organization Quota: %s\n", quota.Name)
				fmt.Printf("  GUID: %s\n", quota.GUID)
				fmt.Printf("  Created: %s\n", quota.CreatedAt.Format("2006-01-02 15:04:05"))
				fmt.Printf("  Updated: %s\n", quota.UpdatedAt.Format("2006-01-02 15:04:05"))

				if quota.Apps != nil {
					fmt.Println("  App Limits:")
					if quota.Apps.TotalMemoryInMB != nil {
						fmt.Printf("    Total Memory: %d MB\n", *quota.Apps.TotalMemoryInMB)
					} else {
						fmt.Printf("    Total Memory: unlimited\n")
					}
					if quota.Apps.TotalInstanceMemoryInMB != nil {
						fmt.Printf("    Instance Memory: %d MB\n", *quota.Apps.TotalInstanceMemoryInMB)
					} else {
						fmt.Printf("    Instance Memory: unlimited\n")
					}
					if quota.Apps.TotalInstances != nil {
						fmt.Printf("    Total Instances: %d\n", *quota.Apps.TotalInstances)
					} else {
						fmt.Printf("    Total Instances: unlimited\n")
					}
					if quota.Apps.TotalAppTasks != nil {
						fmt.Printf("    Total App Tasks: %d\n", *quota.Apps.TotalAppTasks)
					} else {
						fmt.Printf("    Total App Tasks: unlimited\n")
					}
				}

				if quota.Services != nil {
					fmt.Println("  Service Limits:")
					if quota.Services.PaidServicesAllowed != nil {
						fmt.Printf("    Paid Services: %t\n", *quota.Services.PaidServicesAllowed)
					}
					if quota.Services.TotalServiceInstances != nil {
						fmt.Printf("    Total Service Instances: %d\n", *quota.Services.TotalServiceInstances)
					} else {
						fmt.Printf("    Total Service Instances: unlimited\n")
					}
					if quota.Services.TotalServiceKeys != nil {
						fmt.Printf("    Total Service Keys: %d\n", *quota.Services.TotalServiceKeys)
					} else {
						fmt.Printf("    Total Service Keys: unlimited\n")
					}
				}

				if quota.Routes != nil {
					fmt.Println("  Route Limits:")
					if quota.Routes.TotalRoutes != nil {
						fmt.Printf("    Total Routes: %d\n", *quota.Routes.TotalRoutes)
					} else {
						fmt.Printf("    Total Routes: unlimited\n")
					}
					if quota.Routes.TotalReservedPorts != nil {
						fmt.Printf("    Total Reserved Ports: %d\n", *quota.Routes.TotalReservedPorts)
					} else {
						fmt.Printf("    Total Reserved Ports: unlimited\n")
					}
				}

				if quota.Domains != nil {
					fmt.Println("  Domain Limits:")
					if quota.Domains.TotalDomains != nil {
						fmt.Printf("    Total Domains: %d\n", *quota.Domains.TotalDomains)
					} else {
						fmt.Printf("    Total Domains: unlimited\n")
					}
				}
			}

			return nil
		},
	}
}

func newOrgQuotasCreateCommand() *cobra.Command {
	var (
		name                         string
		totalMemoryInMB              int
		totalInstanceMemoryInMB      int
		totalInstances               int
		totalAppTasks                int
		paidServicesAllowed          bool
		totalServiceInstances        int
		totalServiceKeys             int
		totalRoutes                  int
		totalReservedPorts           int
		totalDomains                 int
		logRateLimitInBytesPerSecond int
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new organization quota",
		Long:  "Create a new Cloud Foundry organization quota",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("quota name is required")
			}

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			createReq := &capi.OrganizationQuotaCreateRequest{
				Name: name,
			}

			// Build app limits if any app flags are set
			if cmd.Flags().Changed("total-memory") || cmd.Flags().Changed("instance-memory") ||
				cmd.Flags().Changed("instances") || cmd.Flags().Changed("app-tasks") ||
				cmd.Flags().Changed("log-rate-limit") {
				createReq.Apps = &capi.OrganizationQuotaApps{}
				if cmd.Flags().Changed("total-memory") {
					createReq.Apps.TotalMemoryInMB = &totalMemoryInMB
				}
				if cmd.Flags().Changed("instance-memory") {
					createReq.Apps.TotalInstanceMemoryInMB = &totalInstanceMemoryInMB
				}
				if cmd.Flags().Changed("instances") {
					createReq.Apps.TotalInstances = &totalInstances
				}
				if cmd.Flags().Changed("app-tasks") {
					createReq.Apps.TotalAppTasks = &totalAppTasks
				}
				if cmd.Flags().Changed("log-rate-limit") {
					createReq.Apps.LogRateLimitInBytesPerSecond = &logRateLimitInBytesPerSecond
				}
			}

			// Build service limits if any service flags are set
			if cmd.Flags().Changed("paid-services") || cmd.Flags().Changed("service-instances") ||
				cmd.Flags().Changed("service-keys") {
				createReq.Services = &capi.OrganizationQuotaServices{}
				if cmd.Flags().Changed("paid-services") {
					createReq.Services.PaidServicesAllowed = &paidServicesAllowed
				}
				if cmd.Flags().Changed("service-instances") {
					createReq.Services.TotalServiceInstances = &totalServiceInstances
				}
				if cmd.Flags().Changed("service-keys") {
					createReq.Services.TotalServiceKeys = &totalServiceKeys
				}
			}

			// Build route limits if any route flags are set
			if cmd.Flags().Changed("routes") || cmd.Flags().Changed("reserved-ports") {
				createReq.Routes = &capi.OrganizationQuotaRoutes{}
				if cmd.Flags().Changed("routes") {
					createReq.Routes.TotalRoutes = &totalRoutes
				}
				if cmd.Flags().Changed("reserved-ports") {
					createReq.Routes.TotalReservedPorts = &totalReservedPorts
				}
			}

			// Build domain limits if domain flags are set
			if cmd.Flags().Changed("domains") {
				createReq.Domains = &capi.OrganizationQuotaDomains{}
				createReq.Domains.TotalDomains = &totalDomains
			}

			quota, err := client.OrganizationQuotas().Create(ctx, createReq)
			if err != nil {
				return fmt.Errorf("failed to create organization quota: %w", err)
			}

			fmt.Printf("Successfully created organization quota '%s' with GUID %s\n", quota.Name, quota.GUID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "quota name (required)")
	cmd.Flags().IntVar(&totalMemoryInMB, "total-memory", 0, "total memory limit in MB")
	cmd.Flags().IntVar(&totalInstanceMemoryInMB, "instance-memory", 0, "instance memory limit in MB")
	cmd.Flags().IntVar(&totalInstances, "instances", 0, "total instances limit")
	cmd.Flags().IntVar(&totalAppTasks, "app-tasks", 0, "total app tasks limit")
	cmd.Flags().IntVar(&logRateLimitInBytesPerSecond, "log-rate-limit", 0, "log rate limit in bytes per second")
	cmd.Flags().BoolVar(&paidServicesAllowed, "paid-services", true, "allow paid services")
	cmd.Flags().IntVar(&totalServiceInstances, "service-instances", 0, "total service instances limit")
	cmd.Flags().IntVar(&totalServiceKeys, "service-keys", 0, "total service keys limit")
	cmd.Flags().IntVar(&totalRoutes, "routes", 0, "total routes limit")
	cmd.Flags().IntVar(&totalReservedPorts, "reserved-ports", 0, "total reserved ports limit")
	cmd.Flags().IntVar(&totalDomains, "domains", 0, "total domains limit")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newOrgQuotasUpdateCommand() *cobra.Command {
	var (
		newName                      string
		totalMemoryInMB              int
		totalInstanceMemoryInMB      int
		totalInstances               int
		totalAppTasks                int
		paidServicesAllowed          bool
		totalServiceInstances        int
		totalServiceKeys             int
		totalRoutes                  int
		totalReservedPorts           int
		totalDomains                 int
		logRateLimitInBytesPerSecond int
	)

	cmd := &cobra.Command{
		Use:   "update QUOTA_NAME_OR_GUID",
		Short: "Update an organization quota",
		Long:  "Update an existing Cloud Foundry organization quota",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			quotaClient := client.OrganizationQuotas()

			// Find quota
			var quotaGUID string
			quota, err := quotaClient.Get(ctx, nameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)
				quotas, err := quotaClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find organization quota: %w", err)
				}
				if len(quotas.Resources) == 0 {
					return fmt.Errorf("organization quota '%s' not found", nameOrGUID)
				}
				quotaGUID = quotas.Resources[0].GUID
			} else {
				quotaGUID = quota.GUID
			}

			// Build update request
			updateReq := &capi.OrganizationQuotaUpdateRequest{}

			if newName != "" {
				updateReq.Name = &newName
			}

			// Build app limits if any app flags are set
			if cmd.Flags().Changed("total-memory") || cmd.Flags().Changed("instance-memory") ||
				cmd.Flags().Changed("instances") || cmd.Flags().Changed("app-tasks") ||
				cmd.Flags().Changed("log-rate-limit") {
				updateReq.Apps = &capi.OrganizationQuotaApps{}
				if cmd.Flags().Changed("total-memory") {
					updateReq.Apps.TotalMemoryInMB = &totalMemoryInMB
				}
				if cmd.Flags().Changed("instance-memory") {
					updateReq.Apps.TotalInstanceMemoryInMB = &totalInstanceMemoryInMB
				}
				if cmd.Flags().Changed("instances") {
					updateReq.Apps.TotalInstances = &totalInstances
				}
				if cmd.Flags().Changed("app-tasks") {
					updateReq.Apps.TotalAppTasks = &totalAppTasks
				}
				if cmd.Flags().Changed("log-rate-limit") {
					updateReq.Apps.LogRateLimitInBytesPerSecond = &logRateLimitInBytesPerSecond
				}
			}

			// Build service limits if any service flags are set
			if cmd.Flags().Changed("paid-services") || cmd.Flags().Changed("service-instances") ||
				cmd.Flags().Changed("service-keys") {
				updateReq.Services = &capi.OrganizationQuotaServices{}
				if cmd.Flags().Changed("paid-services") {
					updateReq.Services.PaidServicesAllowed = &paidServicesAllowed
				}
				if cmd.Flags().Changed("service-instances") {
					updateReq.Services.TotalServiceInstances = &totalServiceInstances
				}
				if cmd.Flags().Changed("service-keys") {
					updateReq.Services.TotalServiceKeys = &totalServiceKeys
				}
			}

			// Build route limits if any route flags are set
			if cmd.Flags().Changed("routes") || cmd.Flags().Changed("reserved-ports") {
				updateReq.Routes = &capi.OrganizationQuotaRoutes{}
				if cmd.Flags().Changed("routes") {
					updateReq.Routes.TotalRoutes = &totalRoutes
				}
				if cmd.Flags().Changed("reserved-ports") {
					updateReq.Routes.TotalReservedPorts = &totalReservedPorts
				}
			}

			// Build domain limits if domain flags are set
			if cmd.Flags().Changed("domains") {
				updateReq.Domains = &capi.OrganizationQuotaDomains{}
				updateReq.Domains.TotalDomains = &totalDomains
			}

			// Update quota
			updatedQuota, err := quotaClient.Update(ctx, quotaGUID, updateReq)
			if err != nil {
				return fmt.Errorf("failed to update organization quota: %w", err)
			}

			fmt.Printf("Successfully updated organization quota '%s'\n", updatedQuota.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&newName, "name", "", "new quota name")
	cmd.Flags().IntVar(&totalMemoryInMB, "total-memory", 0, "total memory limit in MB")
	cmd.Flags().IntVar(&totalInstanceMemoryInMB, "instance-memory", 0, "instance memory limit in MB")
	cmd.Flags().IntVar(&totalInstances, "instances", 0, "total instances limit")
	cmd.Flags().IntVar(&totalAppTasks, "app-tasks", 0, "total app tasks limit")
	cmd.Flags().IntVar(&logRateLimitInBytesPerSecond, "log-rate-limit", 0, "log rate limit in bytes per second")
	cmd.Flags().BoolVar(&paidServicesAllowed, "paid-services", true, "allow paid services")
	cmd.Flags().IntVar(&totalServiceInstances, "service-instances", 0, "total service instances limit")
	cmd.Flags().IntVar(&totalServiceKeys, "service-keys", 0, "total service keys limit")
	cmd.Flags().IntVar(&totalRoutes, "routes", 0, "total routes limit")
	cmd.Flags().IntVar(&totalReservedPorts, "reserved-ports", 0, "total reserved ports limit")
	cmd.Flags().IntVar(&totalDomains, "domains", 0, "total domains limit")

	return cmd
}

func newOrgQuotasDeleteCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete QUOTA_NAME_OR_GUID",
		Short: "Delete an organization quota",
		Long:  "Delete a Cloud Foundry organization quota",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			if !force {
				fmt.Printf("Really delete organization quota '%s'? (y/N): ", nameOrGUID)
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
			quotaClient := client.OrganizationQuotas()

			// Find quota
			var quotaGUID string
			var quotaName string
			quota, err := quotaClient.Get(ctx, nameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)
				quotas, err := quotaClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find organization quota: %w", err)
				}
				if len(quotas.Resources) == 0 {
					return fmt.Errorf("organization quota '%s' not found", nameOrGUID)
				}
				quotaGUID = quotas.Resources[0].GUID
				quotaName = quotas.Resources[0].Name
			} else {
				quotaGUID = quota.GUID
				quotaName = quota.Name
			}

			// Delete quota
			err = quotaClient.Delete(ctx, quotaGUID)
			if err != nil {
				return fmt.Errorf("failed to delete organization quota: %w", err)
			}

			fmt.Printf("Successfully deleted organization quota '%s'\n", quotaName)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "force deletion without confirmation")

	return cmd
}

func newOrgQuotasApplyCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "apply QUOTA_NAME_OR_GUID ORG_NAME_OR_GUID...",
		Short: "Apply quota to organizations",
		Long:  "Apply an organization quota to one or more organizations",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			quotaNameOrGUID := args[0]
			orgNamesOrGUIDs := args[1:]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			quotaClient := client.OrganizationQuotas()
			orgsClient := client.Organizations()

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
					return fmt.Errorf("failed to find organization quota: %w", err)
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

			// Resolve organization GUIDs
			var orgGUIDs []string
			var orgNames []string
			for _, orgNameOrGUID := range orgNamesOrGUIDs {
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
					orgGUIDs = append(orgGUIDs, orgs.Resources[0].GUID)
					orgNames = append(orgNames, orgs.Resources[0].Name)
				} else {
					orgGUIDs = append(orgGUIDs, org.GUID)
					orgNames = append(orgNames, org.Name)
				}
			}

			// Apply quota to organizations
			_, err = quotaClient.ApplyToOrganizations(ctx, quotaGUID, orgGUIDs)
			if err != nil {
				return fmt.Errorf("failed to apply quota to organizations: %w", err)
			}

			fmt.Printf("Successfully applied quota '%s' to organizations: %s\n",
				quotaName, strings.Join(orgNames, ", "))

			return nil
		},
	}
}
