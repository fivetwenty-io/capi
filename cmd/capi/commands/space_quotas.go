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

// NewSpaceQuotasCommand creates the space quotas command group
func NewSpaceQuotasCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "space-quotas",
		Aliases: []string{"space-quota", "sq"},
		Short:   "Manage space quotas",
		Long:    "List, create, update, delete, apply, and remove space quotas",
	}

	cmd.AddCommand(newSpaceQuotasListCommand())
	cmd.AddCommand(newSpaceQuotasGetCommand())
	cmd.AddCommand(newSpaceQuotasCreateCommand())
	cmd.AddCommand(newSpaceQuotasUpdateCommand())
	cmd.AddCommand(newSpaceQuotasDeleteCommand())
	cmd.AddCommand(newSpaceQuotasApplyCommand())
	cmd.AddCommand(newSpaceQuotasRemoveCommand())

	return cmd
}

func newSpaceQuotasListCommand() *cobra.Command {
	var (
		allPages bool
		perPage  int
		orgName  string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List space quotas",
		Long:  "List all space quotas",
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
				// Find organization by name
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

			quotas, err := client.SpaceQuotas().List(ctx, params)
			if err != nil {
				return fmt.Errorf("failed to list space quotas: %w", err)
			}

			// Fetch all pages if requested
			allQuotas := quotas.Resources
			if allPages && quotas.Pagination.TotalPages > 1 {
				for page := 2; page <= quotas.Pagination.TotalPages; page++ {
					params.Page = page
					moreQuotas, err := client.SpaceQuotas().List(ctx, params)
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
					fmt.Println("No space quotas found")
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
	cmd.Flags().StringVarP(&orgName, "org", "o", "", "filter by organization name")

	return cmd
}

func newSpaceQuotasGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get QUOTA_NAME_OR_GUID",
		Short: "Get space quota details",
		Long:  "Display detailed information about a specific space quota",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			quotaClient := client.SpaceQuotas()

			// Try to get by GUID first
			quota, err := quotaClient.Get(ctx, nameOrGUID)
			if err != nil {
				// If not found by GUID, try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)
				quotas, err := quotaClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find space quota: %w", err)
				}
				if len(quotas.Resources) == 0 {
					return fmt.Errorf("space quota '%s' not found", nameOrGUID)
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
				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Property", "Value")
				_ = table.Append("Name", quota.Name)
				_ = table.Append("GUID", quota.GUID)
				_ = table.Append("Created", quota.CreatedAt.Format("2006-01-02 15:04:05"))
				_ = table.Append("Updated", quota.UpdatedAt.Format("2006-01-02 15:04:05"))
				if err := table.Render(); err != nil {
					return fmt.Errorf("failed to render table: %w", err)
				}

				if quota.Apps != nil {
					fmt.Println("\nApp Limits:")
					appTable := tablewriter.NewWriter(os.Stdout)
					appTable.Header("Limit", "Value")

					if quota.Apps.TotalMemoryInMB != nil {
						_ = appTable.Append("Total Memory", fmt.Sprintf("%d MB", *quota.Apps.TotalMemoryInMB))
					} else {
						_ = appTable.Append("Total Memory", "unlimited")
					}
					if quota.Apps.TotalInstanceMemoryInMB != nil {
						_ = appTable.Append("Instance Memory", fmt.Sprintf("%d MB", *quota.Apps.TotalInstanceMemoryInMB))
					} else {
						_ = appTable.Append("Instance Memory", "unlimited")
					}
					if quota.Apps.TotalInstances != nil {
						_ = appTable.Append("Total Instances", fmt.Sprintf("%d", *quota.Apps.TotalInstances))
					} else {
						_ = appTable.Append("Total Instances", "unlimited")
					}
					if quota.Apps.TotalAppTasks != nil {
						_ = appTable.Append("Total App Tasks", fmt.Sprintf("%d", *quota.Apps.TotalAppTasks))
					} else {
						_ = appTable.Append("Total App Tasks", "unlimited")
					}
					if quota.Apps.LogRateLimitInBytesPerSecond != nil {
						_ = appTable.Append("Log Rate Limit", fmt.Sprintf("%d bytes/sec", *quota.Apps.LogRateLimitInBytesPerSecond))
					} else {
						_ = appTable.Append("Log Rate Limit", "unlimited")
					}
					if err := appTable.Render(); err != nil {
						return fmt.Errorf("failed to render app table: %w", err)
					}
				}

				if quota.Services != nil {
					fmt.Println("\nService Limits:")
					serviceTable := tablewriter.NewWriter(os.Stdout)
					serviceTable.Header("Limit", "Value")

					if quota.Services.PaidServicesAllowed != nil {
						_ = serviceTable.Append("Paid Services", fmt.Sprintf("%t", *quota.Services.PaidServicesAllowed))
					}
					if quota.Services.TotalServiceInstances != nil {
						_ = serviceTable.Append("Total Service Instances", fmt.Sprintf("%d", *quota.Services.TotalServiceInstances))
					} else {
						_ = serviceTable.Append("Total Service Instances", "unlimited")
					}
					if quota.Services.TotalServiceKeys != nil {
						_ = serviceTable.Append("Total Service Keys", fmt.Sprintf("%d", *quota.Services.TotalServiceKeys))
					} else {
						_ = serviceTable.Append("Total Service Keys", "unlimited")
					}
					if err := serviceTable.Render(); err != nil {
						return fmt.Errorf("failed to render service table: %w", err)
					}
				}

				if quota.Routes != nil {
					fmt.Println("\nRoute Limits:")
					routeTable := tablewriter.NewWriter(os.Stdout)
					routeTable.Header("Limit", "Value")

					if quota.Routes.TotalRoutes != nil {
						_ = routeTable.Append("Total Routes", fmt.Sprintf("%d", *quota.Routes.TotalRoutes))
					} else {
						_ = routeTable.Append("Total Routes", "unlimited")
					}
					if quota.Routes.TotalReservedPorts != nil {
						_ = routeTable.Append("Total Reserved Ports", fmt.Sprintf("%d", *quota.Routes.TotalReservedPorts))
					} else {
						_ = routeTable.Append("Total Reserved Ports", "unlimited")
					}
					if err := routeTable.Render(); err != nil {
						return fmt.Errorf("failed to render route table: %w", err)
					}
				}
			}

			return nil
		},
	}
}

func newSpaceQuotasCreateCommand() *cobra.Command {
	var (
		name                         string
		orgName                      string
		totalMemoryInMB              int
		totalInstanceMemoryInMB      int
		totalInstances               int
		totalAppTasks                int
		paidServicesAllowed          bool
		totalServiceInstances        int
		totalServiceKeys             int
		totalRoutes                  int
		totalReservedPorts           int
		logRateLimitInBytesPerSecond int
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new space quota",
		Long:  "Create a new Cloud Foundry space quota",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("quota name is required")
			}
			if orgName == "" {
				return fmt.Errorf("organization name is required")
			}

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find organization
			orgParams := capi.NewQueryParams()
			orgParams.WithFilter("names", orgName)
			orgs, err := client.Organizations().List(ctx, orgParams)
			if err != nil {
				return fmt.Errorf("failed to find organization: %w", err)
			}
			if len(orgs.Resources) == 0 {
				return fmt.Errorf("organization '%s' not found", orgName)
			}
			orgGUID := orgs.Resources[0].GUID

			createReq := &capi.SpaceQuotaV3CreateRequest{
				Name: name,
				Relationships: capi.SpaceQuotaRelationships{
					Organization: capi.Relationship{
						Data: &capi.RelationshipData{GUID: orgGUID},
					},
					Spaces: capi.ToManyRelationship{
						Data: []capi.RelationshipData{},
					},
				},
			}

			// Build app limits if any app flags are set
			if cmd.Flags().Changed("total-memory") || cmd.Flags().Changed("instance-memory") ||
				cmd.Flags().Changed("instances") || cmd.Flags().Changed("app-tasks") ||
				cmd.Flags().Changed("log-rate-limit") {
				createReq.Apps = &capi.SpaceQuotaApps{}
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
				createReq.Services = &capi.SpaceQuotaServices{}
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
				createReq.Routes = &capi.SpaceQuotaRoutes{}
				if cmd.Flags().Changed("routes") {
					createReq.Routes.TotalRoutes = &totalRoutes
				}
				if cmd.Flags().Changed("reserved-ports") {
					createReq.Routes.TotalReservedPorts = &totalReservedPorts
				}
			}

			quota, err := client.SpaceQuotas().Create(ctx, createReq)
			if err != nil {
				return fmt.Errorf("failed to create space quota: %w", err)
			}

			fmt.Printf("Successfully created space quota '%s' with GUID %s\n", quota.Name, quota.GUID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "quota name (required)")
	cmd.Flags().StringVarP(&orgName, "org", "o", "", "organization name (required)")
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
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("org")

	return cmd
}

func newSpaceQuotasUpdateCommand() *cobra.Command {
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
		logRateLimitInBytesPerSecond int
	)

	cmd := &cobra.Command{
		Use:   "update QUOTA_NAME_OR_GUID",
		Short: "Update a space quota",
		Long:  "Update an existing Cloud Foundry space quota",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			quotaClient := client.SpaceQuotas()

			// Find quota
			var quotaGUID string
			quota, err := quotaClient.Get(ctx, nameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)
				quotas, err := quotaClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find space quota: %w", err)
				}
				if len(quotas.Resources) == 0 {
					return fmt.Errorf("space quota '%s' not found", nameOrGUID)
				}
				quotaGUID = quotas.Resources[0].GUID
			} else {
				quotaGUID = quota.GUID
			}

			// Build update request
			updateReq := &capi.SpaceQuotaV3UpdateRequest{}

			if newName != "" {
				updateReq.Name = &newName
			}

			// Build app limits if any app flags are set
			if cmd.Flags().Changed("total-memory") || cmd.Flags().Changed("instance-memory") ||
				cmd.Flags().Changed("instances") || cmd.Flags().Changed("app-tasks") ||
				cmd.Flags().Changed("log-rate-limit") {
				updateReq.Apps = &capi.SpaceQuotaApps{}
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
				updateReq.Services = &capi.SpaceQuotaServices{}
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
				updateReq.Routes = &capi.SpaceQuotaRoutes{}
				if cmd.Flags().Changed("routes") {
					updateReq.Routes.TotalRoutes = &totalRoutes
				}
				if cmd.Flags().Changed("reserved-ports") {
					updateReq.Routes.TotalReservedPorts = &totalReservedPorts
				}
			}

			// Update quota
			updatedQuota, err := quotaClient.Update(ctx, quotaGUID, updateReq)
			if err != nil {
				return fmt.Errorf("failed to update space quota: %w", err)
			}

			fmt.Printf("Successfully updated space quota '%s'\n", updatedQuota.Name)
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

	return cmd
}

func newSpaceQuotasDeleteCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete QUOTA_NAME_OR_GUID",
		Short: "Delete a space quota",
		Long:  "Delete a Cloud Foundry space quota",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			if !force {
				fmt.Printf("Really delete space quota '%s'? (y/N): ", nameOrGUID)
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
			quotaClient := client.SpaceQuotas()

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
					return fmt.Errorf("failed to find space quota: %w", err)
				}
				if len(quotas.Resources) == 0 {
					return fmt.Errorf("space quota '%s' not found", nameOrGUID)
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
				return fmt.Errorf("failed to delete space quota: %w", err)
			}

			fmt.Printf("Successfully deleted space quota '%s'\n", quotaName)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "force deletion without confirmation")

	return cmd
}

func newSpaceQuotasApplyCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "apply QUOTA_NAME_OR_GUID SPACE_NAME_OR_GUID...",
		Short: "Apply quota to spaces",
		Long:  "Apply a space quota to one or more spaces",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			quotaNameOrGUID := args[0]
			spaceNamesOrGUIDs := args[1:]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			quotaClient := client.SpaceQuotas()
			spacesClient := client.Spaces()

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
					return fmt.Errorf("failed to find space quota: %w", err)
				}
				if len(quotas.Resources) == 0 {
					return fmt.Errorf("space quota '%s' not found", quotaNameOrGUID)
				}
				quotaGUID = quotas.Resources[0].GUID
				quotaName = quotas.Resources[0].Name
			} else {
				quotaGUID = quota.GUID
				quotaName = quota.Name
			}

			// Resolve space GUIDs
			var spaceGUIDs []string
			var spaceNames []string
			for _, spaceNameOrGUID := range spaceNamesOrGUIDs {
				space, err := spacesClient.Get(ctx, spaceNameOrGUID)
				if err != nil {
					// Try by name
					params := capi.NewQueryParams()
					params.WithFilter("names", spaceNameOrGUID)
					// Add org filter if targeted
					if orgGUID := viper.GetString("organization_guid"); orgGUID != "" {
						params.WithFilter("organization_guids", orgGUID)
					}
					spaces, err := spacesClient.List(ctx, params)
					if err != nil {
						return fmt.Errorf("failed to find space '%s': %w", spaceNameOrGUID, err)
					}
					if len(spaces.Resources) == 0 {
						return fmt.Errorf("space '%s' not found", spaceNameOrGUID)
					}
					spaceGUIDs = append(spaceGUIDs, spaces.Resources[0].GUID)
					spaceNames = append(spaceNames, spaces.Resources[0].Name)
				} else {
					spaceGUIDs = append(spaceGUIDs, space.GUID)
					spaceNames = append(spaceNames, space.Name)
				}
			}

			// Apply quota to spaces
			_, err = quotaClient.ApplyToSpaces(ctx, quotaGUID, spaceGUIDs)
			if err != nil {
				return fmt.Errorf("failed to apply quota to spaces: %w", err)
			}

			fmt.Printf("Successfully applied quota '%s' to spaces: %s\n",
				quotaName, strings.Join(spaceNames, ", "))

			return nil
		},
	}
}

func newSpaceQuotasRemoveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "remove QUOTA_NAME_OR_GUID SPACE_NAME_OR_GUID",
		Short: "Remove quota from space",
		Long:  "Remove a space quota from a specific space",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			quotaNameOrGUID := args[0]
			spaceNameOrGUID := args[1]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			quotaClient := client.SpaceQuotas()
			spacesClient := client.Spaces()

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
					return fmt.Errorf("failed to find space quota: %w", err)
				}
				if len(quotas.Resources) == 0 {
					return fmt.Errorf("space quota '%s' not found", quotaNameOrGUID)
				}
				quotaGUID = quotas.Resources[0].GUID
				quotaName = quotas.Resources[0].Name
			} else {
				quotaGUID = quota.GUID
				quotaName = quota.Name
			}

			// Find space
			var spaceGUID string
			var spaceName string
			space, err := spacesClient.Get(ctx, spaceNameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", spaceNameOrGUID)
				// Add org filter if targeted
				if orgGUID := viper.GetString("organization_guid"); orgGUID != "" {
					params.WithFilter("organization_guids", orgGUID)
				}
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

			// Remove quota from space
			err = quotaClient.RemoveFromSpace(ctx, quotaGUID, spaceGUID)
			if err != nil {
				return fmt.Errorf("failed to remove quota from space: %w", err)
			}

			fmt.Printf("Successfully removed quota '%s' from space '%s'\n", quotaName, spaceName)
			return nil
		},
	}
}
