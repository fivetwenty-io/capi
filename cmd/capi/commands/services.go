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

// NewServicesCommand creates the services command group
func NewServicesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "services",
		Aliases: []string{"service"},
		Short:   "Manage services",
		Long:    "List and manage Cloud Foundry services and service instances",
	}

	cmd.AddCommand(newServicesListCommand())
	cmd.AddCommand(newServicesGetCommand())
	cmd.AddCommand(newServicesCreateCommand())
	cmd.AddCommand(newServicesUpdateCommand())
	cmd.AddCommand(newServicesDeleteCommand())
	cmd.AddCommand(newServicesBindCommand())
	cmd.AddCommand(newServicesUnbindCommand())
	cmd.AddCommand(newServicesRenameCommand())
	cmd.AddCommand(newServicesShareCommand())
	cmd.AddCommand(newServicesUnshareCommand())
	cmd.AddCommand(newServicesListBindingsCommand())
	cmd.AddCommand(newServicesBrokersCommand())
	cmd.AddCommand(newServicesOfferingsCommand())
	cmd.AddCommand(newServicesPlansCommand())

	return cmd
}

func newServicesListCommand() *cobra.Command {
	var (
		spaceName string
		allPages  bool
		perPage   int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List service instances",
		Long:  "List all service instances the user has access to",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			params := capi.NewQueryParams()
			params.PerPage = perPage

			// Filter by space if specified
			if spaceName != "" {
				// Find space by name
				spaceParams := capi.NewQueryParams()
				spaceParams.WithFilter("names", spaceName)

				// Add org filter if targeted
				if orgGUID := viper.GetString("organization_guid"); orgGUID != "" {
					spaceParams.WithFilter("organization_guids", orgGUID)
				}

				spaces, err := client.Spaces().List(ctx, spaceParams)
				if err != nil {
					return fmt.Errorf("failed to find space: %w", err)
				}
				if len(spaces.Resources) == 0 {
					return fmt.Errorf("space '%s' not found", spaceName)
				}

				params.WithFilter("space_guids", spaces.Resources[0].GUID)
			} else if spaceGUID := viper.GetString("space_guid"); spaceGUID != "" {
				// Use targeted space
				params.WithFilter("space_guids", spaceGUID)
			}

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
					fmt.Println("No service instances found")
					return nil
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Name", "Service", "Plan", "Bound Apps", "State", "Type")

				for _, service := range allServices {
					// Get service plan and offering details if available
					var planName, offeringName string
					if service.Relationships.ServicePlan != nil && service.Relationships.ServicePlan.Data != nil {
						plan, _ := client.ServicePlans().Get(ctx, service.Relationships.ServicePlan.Data.GUID)
						if plan != nil {
							planName = plan.Name
							if plan.Relationships.ServiceOffering.Data != nil {
								offering, _ := client.ServiceOfferings().Get(ctx, plan.Relationships.ServiceOffering.Data.GUID)
								if offering != nil {
									offeringName = offering.Name
								}
							}
						}
					}

					// Get bound apps count
					bindingsParams := capi.NewQueryParams()
					bindingsParams.WithFilter("service_instance_guids", service.GUID)
					bindings, _ := client.ServiceCredentialBindings().List(ctx, bindingsParams)
					boundApps := 0
					if bindings != nil {
						boundApps = len(bindings.Resources)
					}

					state := "ready"
					if service.LastOperation != nil {
						state = service.LastOperation.State
					}

					_ = table.Append(service.Name, offeringName, planName, fmt.Sprintf("%d", boundApps), state, service.Type)
				}

				_ = table.Render()

				if !allPages && services.Pagination.TotalPages > 1 {
					fmt.Printf("\nShowing page 1 of %d. Use --all to fetch all pages.\n", services.Pagination.TotalPages)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&spaceName, "space", "s", "", "filter by space name")
	cmd.Flags().BoolVar(&allPages, "all", false, "fetch all pages")
	cmd.Flags().IntVar(&perPage, "per-page", 50, "results per page")

	return cmd
}

func newServicesGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get SERVICE_NAME_OR_GUID",
		Short: "Get service instance details",
		Long:  "Display detailed information about a specific service instance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			servicesClient := client.ServiceInstances()

			// Try to get by GUID first
			service, err := servicesClient.Get(ctx, nameOrGUID)
			if err != nil {
				// If not found by GUID, try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)

				// Filter by space if targeted
				if spaceGUID := viper.GetString("space_guid"); spaceGUID != "" {
					params.WithFilter("space_guids", spaceGUID)
				}

				services, err := servicesClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find service instance: %w", err)
				}
				if len(services.Resources) == 0 {
					return fmt.Errorf("service instance '%s' not found", nameOrGUID)
				}
				service = &services.Resources[0]
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(service)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(service)
			default:
				fmt.Printf("Service Instance: %s\n", service.Name)
				fmt.Printf("  GUID:        %s\n", service.GUID)
				fmt.Printf("  Type:        %s\n", service.Type)
				fmt.Printf("  Created:     %s\n", service.CreatedAt.Format("2006-01-02 15:04:05"))
				fmt.Printf("  Updated:     %s\n", service.UpdatedAt.Format("2006-01-02 15:04:05"))

				// Service plan and offering details
				if service.Relationships.ServicePlan != nil && service.Relationships.ServicePlan.Data != nil {
					plan, _ := client.ServicePlans().Get(ctx, service.Relationships.ServicePlan.Data.GUID)
					if plan != nil {
						fmt.Printf("  Plan:        %s\n", plan.Name)
						if plan.Relationships.ServiceOffering.Data != nil {
							offering, _ := client.ServiceOfferings().Get(ctx, plan.Relationships.ServiceOffering.Data.GUID)
							if offering != nil {
								fmt.Printf("  Service:     %s\n", offering.Name)
							}
						}
					}
				}

				// Space info
				if service.Relationships.Space.Data != nil {
					space, _ := client.Spaces().Get(ctx, service.Relationships.Space.Data.GUID)
					if space != nil {
						fmt.Printf("  Space:       %s\n", space.Name)
					}
				}

				// Last operation
				if service.LastOperation != nil {
					fmt.Printf("  Last Operation:\n")
					fmt.Printf("    Type:        %s\n", service.LastOperation.Type)
					fmt.Printf("    State:       %s\n", service.LastOperation.State)
					fmt.Printf("    Description: %s\n", service.LastOperation.Description)
					if service.LastOperation.UpdatedAt != nil {
						fmt.Printf("    Updated:     %s\n", service.LastOperation.UpdatedAt.Format("2006-01-02 15:04:05"))
					}
				}

				// Tags
				if len(service.Tags) > 0 {
					fmt.Printf("  Tags:        %s\n", strings.Join(service.Tags, ", "))
				}

				// Dashboard URL
				if service.DashboardURL != nil {
					fmt.Printf("  Dashboard:   %s\n", *service.DashboardURL)
				}

				// User-provided specific fields
				if service.Type == "user-provided" {
					if service.SyslogDrainURL != nil {
						fmt.Printf("  Syslog Drain: %s\n", *service.SyslogDrainURL)
					}
					if service.RouteServiceURL != nil {
						fmt.Printf("  Route Service: %s\n", *service.RouteServiceURL)
					}
				}
			}

			return nil
		},
	}
}

func newServicesCreateCommand() *cobra.Command {
	var (
		serviceName     string
		planName        string
		spaceName       string
		tags            []string
		parameters      map[string]string
		syslogDrainURL  string
		routeServiceURL string
		credentials     map[string]string
		userProvided    bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a service instance",
		Long:  "Create a new service instance (managed or user-provided)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if serviceName == "" {
				return fmt.Errorf("service instance name is required")
			}

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find space
			var spaceGUID string
			if spaceName != "" {
				spaceParams := capi.NewQueryParams()
				spaceParams.WithFilter("names", spaceName)

				// Add org filter if targeted
				if orgGUID := viper.GetString("organization_guid"); orgGUID != "" {
					spaceParams.WithFilter("organization_guids", orgGUID)
				}

				spaces, err := client.Spaces().List(ctx, spaceParams)
				if err != nil {
					return fmt.Errorf("failed to find space: %w", err)
				}
				if len(spaces.Resources) == 0 {
					return fmt.Errorf("space '%s' not found", spaceName)
				}
				spaceGUID = spaces.Resources[0].GUID
			} else if targetedSpaceGUID := viper.GetString("space_guid"); targetedSpaceGUID != "" {
				spaceGUID = targetedSpaceGUID
			} else {
				return fmt.Errorf("space is required (use --space or target a space)")
			}

			createReq := &capi.ServiceInstanceCreateRequest{
				Name: serviceName,
				Relationships: capi.ServiceInstanceRelationships{
					Space: capi.Relationship{
						Data: &capi.RelationshipData{GUID: spaceGUID},
					},
				},
				Tags: tags,
			}

			if userProvided {
				createReq.Type = "user-provided"

				// Convert parameters map to credentials
				if len(credentials) > 0 {
					credMap := make(map[string]interface{})
					for k, v := range credentials {
						credMap[k] = v
					}
					createReq.Credentials = credMap
				}

				if syslogDrainURL != "" {
					createReq.SyslogDrainURL = &syslogDrainURL
				}

				if routeServiceURL != "" {
					createReq.RouteServiceURL = &routeServiceURL
				}
			} else {
				createReq.Type = "managed"

				if planName == "" {
					return fmt.Errorf("service plan is required for managed services")
				}

				// Find service plan by name
				planParams := capi.NewQueryParams()
				planParams.WithFilter("names", planName)
				plans, err := client.ServicePlans().List(ctx, planParams)
				if err != nil {
					return fmt.Errorf("failed to find service plan: %w", err)
				}
				if len(plans.Resources) == 0 {
					return fmt.Errorf("service plan '%s' not found", planName)
				}

				createReq.Relationships.ServicePlan = &capi.Relationship{
					Data: &capi.RelationshipData{GUID: plans.Resources[0].GUID},
				}

				// Convert parameters
				if len(parameters) > 0 {
					paramMap := make(map[string]interface{})
					for k, v := range parameters {
						paramMap[k] = v
					}
					createReq.Parameters = paramMap
				}
			}

			result, err := client.ServiceInstances().Create(ctx, createReq)
			if err != nil {
				return fmt.Errorf("failed to create service instance: %w", err)
			}

			if userProvided {
				// User-provided services return the service instance directly
				if service, ok := result.(*capi.ServiceInstance); ok {
					fmt.Printf("Successfully created user-provided service instance '%s'\n", service.Name)
					fmt.Printf("  GUID: %s\n", service.GUID)
				}
			} else {
				// Managed services return a job
				if job, ok := result.(*capi.Job); ok {
					fmt.Printf("Successfully initiated creation of managed service instance '%s'\n", serviceName)
					fmt.Printf("  Job GUID: %s\n", job.GUID)
					fmt.Printf("  Monitor with: capi jobs get %s\n", job.GUID)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&serviceName, "name", "n", "", "service instance name (required)")
	cmd.Flags().StringVarP(&planName, "plan", "p", "", "service plan name (required for managed services)")
	cmd.Flags().StringVarP(&spaceName, "space", "s", "", "space name (defaults to targeted space)")
	cmd.Flags().StringArrayVarP(&tags, "tags", "t", nil, "tags for the service instance")
	cmd.Flags().StringToStringVar(&parameters, "parameters", nil, "parameters for managed service instances (key=value)")
	cmd.Flags().StringVar(&syslogDrainURL, "syslog-drain-url", "", "syslog drain URL for user-provided services")
	cmd.Flags().StringVar(&routeServiceURL, "route-service-url", "", "route service URL for user-provided services")
	cmd.Flags().StringToStringVar(&credentials, "credentials", nil, "credentials for user-provided services (key=value)")
	cmd.Flags().BoolVar(&userProvided, "user-provided", false, "create a user-provided service instance")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newServicesUpdateCommand() *cobra.Command {
	var (
		newName         string
		tags            []string
		parameters      map[string]string
		syslogDrainURL  string
		routeServiceURL string
		credentials     map[string]string
		planName        string
	)

	cmd := &cobra.Command{
		Use:   "update SERVICE_NAME_OR_GUID",
		Short: "Update a service instance",
		Long:  "Update an existing service instance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			servicesClient := client.ServiceInstances()

			// Find service instance
			var serviceGUID string
			var serviceName string
			var serviceType string
			service, err := servicesClient.Get(ctx, nameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)

				// Filter by space if targeted
				if spaceGUID := viper.GetString("space_guid"); spaceGUID != "" {
					params.WithFilter("space_guids", spaceGUID)
				}

				services, err := servicesClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find service instance: %w", err)
				}
				if len(services.Resources) == 0 {
					return fmt.Errorf("service instance '%s' not found", nameOrGUID)
				}
				service = &services.Resources[0]
			}

			serviceGUID = service.GUID
			serviceName = service.Name
			serviceType = service.Type

			// Build update request
			updateReq := &capi.ServiceInstanceUpdateRequest{}

			if newName != "" {
				updateReq.Name = &newName
			}

			if len(tags) > 0 {
				updateReq.Tags = tags
			}

			if serviceType == "user-provided" {
				// User-provided service updates
				if len(credentials) > 0 {
					credMap := make(map[string]interface{})
					for k, v := range credentials {
						credMap[k] = v
					}
					updateReq.Credentials = credMap
				}

				if cmd.Flags().Changed("syslog-drain-url") {
					updateReq.SyslogDrainURL = &syslogDrainURL
				}

				if cmd.Flags().Changed("route-service-url") {
					updateReq.RouteServiceURL = &routeServiceURL
				}
			} else {
				// Managed service updates
				if len(parameters) > 0 {
					paramMap := make(map[string]interface{})
					for k, v := range parameters {
						paramMap[k] = v
					}
					updateReq.Parameters = paramMap
				}

				// Plan change
				if planName != "" {
					planParams := capi.NewQueryParams()
					planParams.WithFilter("names", planName)
					plans, err := client.ServicePlans().List(ctx, planParams)
					if err != nil {
						return fmt.Errorf("failed to find service plan: %w", err)
					}
					if len(plans.Resources) == 0 {
						return fmt.Errorf("service plan '%s' not found", planName)
					}

					updateReq.Relationships = &capi.ServiceInstanceRelationships{
						ServicePlan: &capi.Relationship{
							Data: &capi.RelationshipData{GUID: plans.Resources[0].GUID},
						},
					}
				}
			}

			result, err := servicesClient.Update(ctx, serviceGUID, updateReq)
			if err != nil {
				return fmt.Errorf("failed to update service instance: %w", err)
			}

			if serviceType == "user-provided" {
				// User-provided services return the service instance directly
				if updatedService, ok := result.(*capi.ServiceInstance); ok {
					fmt.Printf("Successfully updated user-provided service instance '%s'\n", updatedService.Name)
				}
			} else {
				// Managed services return a job
				if job, ok := result.(*capi.Job); ok {
					fmt.Printf("Successfully initiated update of managed service instance '%s'\n", serviceName)
					fmt.Printf("  Job GUID: %s\n", job.GUID)
					fmt.Printf("  Monitor with: capi jobs get %s\n", job.GUID)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&newName, "name", "", "new service instance name")
	cmd.Flags().StringArrayVarP(&tags, "tags", "t", nil, "tags for the service instance")
	cmd.Flags().StringToStringVar(&parameters, "parameters", nil, "parameters for managed service instances (key=value)")
	cmd.Flags().StringVar(&syslogDrainURL, "syslog-drain-url", "", "syslog drain URL for user-provided services")
	cmd.Flags().StringVar(&routeServiceURL, "route-service-url", "", "route service URL for user-provided services")
	cmd.Flags().StringToStringVar(&credentials, "credentials", nil, "credentials for user-provided services (key=value)")
	cmd.Flags().StringVarP(&planName, "plan", "p", "", "new service plan name for managed services")

	return cmd
}

func newServicesDeleteCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete SERVICE_NAME_OR_GUID",
		Short: "Delete a service instance",
		Long:  "Delete a service instance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			if !force {
				fmt.Printf("Really delete service instance '%s'? (y/N): ", nameOrGUID)
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
			servicesClient := client.ServiceInstances()

			// Find service instance
			var serviceGUID string
			var serviceName string
			service, err := servicesClient.Get(ctx, nameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)

				// Filter by space if targeted
				if spaceGUID := viper.GetString("space_guid"); spaceGUID != "" {
					params.WithFilter("space_guids", spaceGUID)
				}

				services, err := servicesClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find service instance: %w", err)
				}
				if len(services.Resources) == 0 {
					return fmt.Errorf("service instance '%s' not found", nameOrGUID)
				}
				service = &services.Resources[0]
			}

			serviceGUID = service.GUID
			serviceName = service.Name

			// Delete service instance
			job, err := servicesClient.Delete(ctx, serviceGUID)
			if err != nil {
				return fmt.Errorf("failed to delete service instance: %w", err)
			}

			if job != nil {
				fmt.Printf("Deleting service instance '%s'... (job: %s)\n", serviceName, job.GUID)
				fmt.Printf("Monitor with: capi jobs get %s\n", job.GUID)
			} else {
				fmt.Printf("Successfully deleted service instance '%s'\n", serviceName)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "force deletion without confirmation")

	return cmd
}

func newServicesBindCommand() *cobra.Command {
	var (
		parameters  map[string]string
		bindingName string
	)

	cmd := &cobra.Command{
		Use:   "bind SERVICE_NAME_OR_GUID APP_NAME_OR_GUID",
		Short: "Bind a service instance to an application",
		Long:  "Create a service credential binding between a service instance and an application",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceNameOrGUID := args[0]
			appNameOrGUID := args[1]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find service instance
			var serviceGUID string
			service, err := client.ServiceInstances().Get(ctx, serviceNameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", serviceNameOrGUID)

				// Filter by space if targeted
				if spaceGUID := viper.GetString("space_guid"); spaceGUID != "" {
					params.WithFilter("space_guids", spaceGUID)
				}

				services, err := client.ServiceInstances().List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find service instance: %w", err)
				}
				if len(services.Resources) == 0 {
					return fmt.Errorf("service instance '%s' not found", serviceNameOrGUID)
				}
				service = &services.Resources[0]
			}
			serviceGUID = service.GUID

			// Find application
			var appGUID string
			app, err := client.Apps().Get(ctx, appNameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", appNameOrGUID)

				// Filter by space if targeted
				if spaceGUID := viper.GetString("space_guid"); spaceGUID != "" {
					params.WithFilter("space_guids", spaceGUID)
				}

				apps, err := client.Apps().List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find application: %w", err)
				}
				if len(apps.Resources) == 0 {
					return fmt.Errorf("application '%s' not found", appNameOrGUID)
				}
				app = &apps.Resources[0]
			}
			appGUID = app.GUID

			// Create service credential binding
			createReq := &capi.ServiceCredentialBindingCreateRequest{
				Type: "app",
				Relationships: capi.ServiceCredentialBindingRelationships{
					ServiceInstance: capi.Relationship{
						Data: &capi.RelationshipData{GUID: serviceGUID},
					},
					App: &capi.Relationship{
						Data: &capi.RelationshipData{GUID: appGUID},
					},
				},
			}

			if bindingName != "" {
				createReq.Name = &bindingName
			}

			if len(parameters) > 0 {
				paramMap := make(map[string]interface{})
				for k, v := range parameters {
					paramMap[k] = v
				}
				createReq.Parameters = paramMap
			}

			result, err := client.ServiceCredentialBindings().Create(ctx, createReq)
			if err != nil {
				return fmt.Errorf("failed to bind service instance: %w", err)
			}

			if binding, ok := result.(*capi.ServiceCredentialBinding); ok {
				fmt.Printf("Successfully bound service instance '%s' to application '%s'\n", service.Name, app.Name)
				fmt.Printf("  Binding GUID: %s\n", binding.GUID)
			} else if job, ok := result.(*capi.Job); ok {
				fmt.Printf("Successfully initiated binding of service instance '%s' to application '%s'\n", service.Name, app.Name)
				fmt.Printf("  Job GUID: %s\n", job.GUID)
				fmt.Printf("  Monitor with: capi jobs get %s\n", job.GUID)
			}

			return nil
		},
	}

	cmd.Flags().StringToStringVar(&parameters, "parameters", nil, "binding parameters (key=value)")
	cmd.Flags().StringVar(&bindingName, "name", "", "name for the service binding")

	return cmd
}

func newServicesUnbindCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "unbind SERVICE_NAME_OR_GUID APP_NAME_OR_GUID",
		Short: "Unbind a service instance from an application",
		Long:  "Remove a service credential binding between a service instance and an application",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceNameOrGUID := args[0]
			appNameOrGUID := args[1]

			if !force {
				fmt.Printf("Really unbind service instance '%s' from application '%s'? (y/N): ", serviceNameOrGUID, appNameOrGUID)
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

			// Find service instance
			var serviceGUID string
			service, err := client.ServiceInstances().Get(ctx, serviceNameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", serviceNameOrGUID)

				// Filter by space if targeted
				if spaceGUID := viper.GetString("space_guid"); spaceGUID != "" {
					params.WithFilter("space_guids", spaceGUID)
				}

				services, err := client.ServiceInstances().List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find service instance: %w", err)
				}
				if len(services.Resources) == 0 {
					return fmt.Errorf("service instance '%s' not found", serviceNameOrGUID)
				}
				service = &services.Resources[0]
			}
			serviceGUID = service.GUID

			// Find application
			var appGUID string
			app, err := client.Apps().Get(ctx, appNameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", appNameOrGUID)

				// Filter by space if targeted
				if spaceGUID := viper.GetString("space_guid"); spaceGUID != "" {
					params.WithFilter("space_guids", spaceGUID)
				}

				apps, err := client.Apps().List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find application: %w", err)
				}
				if len(apps.Resources) == 0 {
					return fmt.Errorf("application '%s' not found", appNameOrGUID)
				}
				app = &apps.Resources[0]
			}
			appGUID = app.GUID

			// Find the binding
			bindingParams := capi.NewQueryParams()
			bindingParams.WithFilter("service_instance_guids", serviceGUID)
			bindingParams.WithFilter("app_guids", appGUID)

			bindings, err := client.ServiceCredentialBindings().List(ctx, bindingParams)
			if err != nil {
				return fmt.Errorf("failed to find service binding: %w", err)
			}
			if len(bindings.Resources) == 0 {
				return fmt.Errorf("no binding found between service '%s' and app '%s'", service.Name, app.Name)
			}

			// Delete the binding
			binding := bindings.Resources[0]
			job, err := client.ServiceCredentialBindings().Delete(ctx, binding.GUID)
			if err != nil {
				return fmt.Errorf("failed to unbind service instance: %w", err)
			}

			if job != nil {
				fmt.Printf("Unbinding service instance '%s' from application '%s'... (job: %s)\n", service.Name, app.Name, job.GUID)
				fmt.Printf("Monitor with: capi jobs get %s\n", job.GUID)
			} else {
				fmt.Printf("Successfully unbound service instance '%s' from application '%s'\n", service.Name, app.Name)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "force unbinding without confirmation")

	return cmd
}

func newServicesRenameCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "rename SERVICE_NAME_OR_GUID NEW_NAME",
		Short: "Rename a service instance",
		Long:  "Change the name of a service instance",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]
			newName := args[1]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			servicesClient := client.ServiceInstances()

			// Find service instance
			var serviceGUID string
			service, err := servicesClient.Get(ctx, nameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)

				// Filter by space if targeted
				if spaceGUID := viper.GetString("space_guid"); spaceGUID != "" {
					params.WithFilter("space_guids", spaceGUID)
				}

				services, err := servicesClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find service instance: %w", err)
				}
				if len(services.Resources) == 0 {
					return fmt.Errorf("service instance '%s' not found", nameOrGUID)
				}
				service = &services.Resources[0]
			}
			serviceGUID = service.GUID

			// Update with new name
			updateReq := &capi.ServiceInstanceUpdateRequest{
				Name: &newName,
			}

			result, err := servicesClient.Update(ctx, serviceGUID, updateReq)
			if err != nil {
				return fmt.Errorf("failed to rename service instance: %w", err)
			}

			if service.Type == "user-provided" {
				// User-provided services return the service instance directly
				if updatedService, ok := result.(*capi.ServiceInstance); ok {
					fmt.Printf("Successfully renamed service instance to '%s'\n", updatedService.Name)
				}
			} else {
				// Managed services return a job
				if job, ok := result.(*capi.Job); ok {
					fmt.Printf("Successfully initiated rename of service instance to '%s'\n", newName)
					fmt.Printf("  Job GUID: %s\n", job.GUID)
					fmt.Printf("  Monitor with: capi jobs get %s\n", job.GUID)
				}
			}

			return nil
		},
	}
}

func newServicesShareCommand() *cobra.Command {
	var spaceNames []string

	cmd := &cobra.Command{
		Use:   "share SERVICE_NAME_OR_GUID",
		Short: "Share a service instance with other spaces",
		Long:  "Share a service instance with specified spaces",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			if len(spaceNames) == 0 {
				return fmt.Errorf("at least one space must be specified")
			}

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			servicesClient := client.ServiceInstances()

			// Find service instance
			var serviceGUID string
			service, err := servicesClient.Get(ctx, nameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)

				// Filter by space if targeted
				if spaceGUID := viper.GetString("space_guid"); spaceGUID != "" {
					params.WithFilter("space_guids", spaceGUID)
				}

				services, err := servicesClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find service instance: %w", err)
				}
				if len(services.Resources) == 0 {
					return fmt.Errorf("service instance '%s' not found", nameOrGUID)
				}
				service = &services.Resources[0]
			}
			serviceGUID = service.GUID

			// Find spaces to share with
			var spaceRelationships []capi.Relationship
			for _, spaceName := range spaceNames {
				params := capi.NewQueryParams()
				params.WithFilter("names", spaceName)

				// Add org filter if targeted
				if orgGUID := viper.GetString("organization_guid"); orgGUID != "" {
					params.WithFilter("organization_guids", orgGUID)
				}

				spaces, err := client.Spaces().List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find space '%s': %w", spaceName, err)
				}
				if len(spaces.Resources) == 0 {
					return fmt.Errorf("space '%s' not found", spaceName)
				}

				spaceRelationships = append(spaceRelationships, capi.Relationship{
					Data: &capi.RelationshipData{GUID: spaces.Resources[0].GUID},
				})
			}

			// Share with spaces
			shareReq := &capi.ServiceInstanceShareRequest{
				Data: spaceRelationships,
			}

			_, err = servicesClient.ShareWithSpaces(ctx, serviceGUID, shareReq)
			if err != nil {
				return fmt.Errorf("failed to share service instance: %w", err)
			}

			fmt.Printf("Successfully shared service instance '%s' with spaces: %s\n", service.Name, strings.Join(spaceNames, ", "))

			return nil
		},
	}

	cmd.Flags().StringArrayVarP(&spaceNames, "spaces", "s", nil, "spaces to share with (required)")
	_ = cmd.MarkFlagRequired("spaces")

	return cmd
}

func newServicesUnshareCommand() *cobra.Command {
	var spaceName string

	cmd := &cobra.Command{
		Use:   "unshare SERVICE_NAME_OR_GUID",
		Short: "Unshare a service instance from a space",
		Long:  "Remove sharing of a service instance from a specified space",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			if spaceName == "" {
				return fmt.Errorf("space name is required")
			}

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			servicesClient := client.ServiceInstances()

			// Find service instance
			var serviceGUID string
			service, err := servicesClient.Get(ctx, nameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)

				// Filter by space if targeted
				if spaceGUID := viper.GetString("space_guid"); spaceGUID != "" {
					params.WithFilter("space_guids", spaceGUID)
				}

				services, err := servicesClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find service instance: %w", err)
				}
				if len(services.Resources) == 0 {
					return fmt.Errorf("service instance '%s' not found", nameOrGUID)
				}
				service = &services.Resources[0]
			}
			serviceGUID = service.GUID

			// Find space to unshare from
			params := capi.NewQueryParams()
			params.WithFilter("names", spaceName)

			// Add org filter if targeted
			if orgGUID := viper.GetString("organization_guid"); orgGUID != "" {
				params.WithFilter("organization_guids", orgGUID)
			}

			spaces, err := client.Spaces().List(ctx, params)
			if err != nil {
				return fmt.Errorf("failed to find space: %w", err)
			}
			if len(spaces.Resources) == 0 {
				return fmt.Errorf("space '%s' not found", spaceName)
			}
			spaceGUID := spaces.Resources[0].GUID

			// Unshare from space
			err = servicesClient.UnshareFromSpace(ctx, serviceGUID, spaceGUID)
			if err != nil {
				return fmt.Errorf("failed to unshare service instance: %w", err)
			}

			fmt.Printf("Successfully unshared service instance '%s' from space '%s'\n", service.Name, spaceName)

			return nil
		},
	}

	cmd.Flags().StringVarP(&spaceName, "space", "s", "", "space to unshare from (required)")
	_ = cmd.MarkFlagRequired("space")

	return cmd
}

func newServicesListBindingsCommand() *cobra.Command {
	var (
		allPages bool
		perPage  int
	)

	cmd := &cobra.Command{
		Use:   "list-bindings SERVICE_NAME_OR_GUID",
		Short: "List service bindings for a service instance",
		Long:  "List all credential bindings for a service instance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find service instance
			var serviceGUID string
			var serviceName string
			service, err := client.ServiceInstances().Get(ctx, nameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)

				// Filter by space if targeted
				if spaceGUID := viper.GetString("space_guid"); spaceGUID != "" {
					params.WithFilter("space_guids", spaceGUID)
				}

				services, err := client.ServiceInstances().List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find service instance: %w", err)
				}
				if len(services.Resources) == 0 {
					return fmt.Errorf("service instance '%s' not found", nameOrGUID)
				}
				service = &services.Resources[0]
			}
			serviceGUID = service.GUID
			serviceName = service.Name

			// List bindings for this service instance
			params := capi.NewQueryParams()
			params.PerPage = perPage
			params.WithFilter("service_instance_guids", serviceGUID)

			bindings, err := client.ServiceCredentialBindings().List(ctx, params)
			if err != nil {
				return fmt.Errorf("failed to list service bindings: %w", err)
			}

			// Fetch all pages if requested
			allBindings := bindings.Resources
			if allPages && bindings.Pagination.TotalPages > 1 {
				for page := 2; page <= bindings.Pagination.TotalPages; page++ {
					params.Page = page
					moreBindings, err := client.ServiceCredentialBindings().List(ctx, params)
					if err != nil {
						return fmt.Errorf("failed to fetch page %d: %w", page, err)
					}
					allBindings = append(allBindings, moreBindings.Resources...)
				}
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(allBindings)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(allBindings)
			default:
				if len(allBindings) == 0 {
					fmt.Printf("No bindings found for service instance '%s'\n", serviceName)
					return nil
				}

				fmt.Printf("Bindings for service instance '%s':\n\n", serviceName)
				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Name", "Type", "App", "State", "Created")

				for _, binding := range allBindings {
					appName := ""
					if binding.Relationships.App != nil && binding.Relationships.App.Data != nil {
						app, _ := client.Apps().Get(ctx, binding.Relationships.App.Data.GUID)
						if app != nil {
							appName = app.Name
						}
					}

					state := "ready"
					if binding.LastOperation != nil {
						state = binding.LastOperation.State
					}

					_ = table.Append(binding.Name, binding.Type, appName, state, binding.CreatedAt.Format("2006-01-02"))
				}

				_ = table.Render()

				if !allPages && bindings.Pagination.TotalPages > 1 {
					fmt.Printf("\nShowing page 1 of %d. Use --all to fetch all pages.\n", bindings.Pagination.TotalPages)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&allPages, "all", false, "fetch all pages")
	cmd.Flags().IntVar(&perPage, "per-page", 50, "results per page")

	return cmd
}

func newServicesBrokersCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "brokers",
		Aliases: []string{"broker"},
		Short:   "Manage service brokers",
		Long:    "List and manage Cloud Foundry service brokers",
	}

	cmd.AddCommand(newServicesBrokersListCommand())
	cmd.AddCommand(newServicesBrokersGetCommand())

	return cmd
}

func newServicesBrokersListCommand() *cobra.Command {
	var (
		allPages bool
		perPage  int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List service brokers",
		Long:  "List all service brokers",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			params := capi.NewQueryParams()
			params.PerPage = perPage

			brokers, err := client.ServiceBrokers().List(ctx, params)
			if err != nil {
				return fmt.Errorf("failed to list service brokers: %w", err)
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(brokers.Resources)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(brokers.Resources)
			default:
				if len(brokers.Resources) == 0 {
					fmt.Println("No service brokers found")
					return nil
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Name", "URL", "GUID", "Space", "Created")

				for _, broker := range brokers.Resources {
					spaceName := "platform"
					if broker.Relationships.Space != nil && broker.Relationships.Space.Data != nil {
						space, _ := client.Spaces().Get(ctx, broker.Relationships.Space.Data.GUID)
						if space != nil {
							spaceName = space.Name
						}
					}

					_ = table.Append(broker.Name, broker.URL, broker.GUID, spaceName, broker.CreatedAt.Format("2006-01-02"))
				}

				_ = table.Render()
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&allPages, "all", false, "fetch all pages")
	cmd.Flags().IntVar(&perPage, "per-page", 50, "results per page")

	return cmd
}

func newServicesBrokersGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get BROKER_NAME_OR_GUID",
		Short: "Get service broker details",
		Long:  "Display detailed information about a specific service broker",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Try to get by GUID first
			broker, err := client.ServiceBrokers().Get(ctx, nameOrGUID)
			if err != nil {
				// If not found by GUID, try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)
				brokers, err := client.ServiceBrokers().List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find service broker: %w", err)
				}
				if len(brokers.Resources) == 0 {
					return fmt.Errorf("service broker '%s' not found", nameOrGUID)
				}
				broker = &brokers.Resources[0]
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(broker)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(broker)
			default:
				fmt.Printf("Service Broker: %s\n", broker.Name)
				fmt.Printf("  GUID:    %s\n", broker.GUID)
				fmt.Printf("  URL:     %s\n", broker.URL)
				fmt.Printf("  Created: %s\n", broker.CreatedAt.Format("2006-01-02 15:04:05"))
				fmt.Printf("  Updated: %s\n", broker.UpdatedAt.Format("2006-01-02 15:04:05"))

				if broker.Relationships.Space != nil && broker.Relationships.Space.Data != nil {
					space, _ := client.Spaces().Get(ctx, broker.Relationships.Space.Data.GUID)
					if space != nil {
						fmt.Printf("  Space:   %s\n", space.Name)
					}
				} else {
					fmt.Printf("  Space:   platform (global)\n")
				}
			}

			return nil
		},
	}
}

func newServicesOfferingsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "offerings",
		Aliases: []string{"offering"},
		Short:   "Manage service offerings",
		Long:    "List and manage Cloud Foundry service offerings",
	}

	cmd.AddCommand(newServicesOfferingsListCommand())
	cmd.AddCommand(newServicesOfferingsGetCommand())

	return cmd
}

func newServicesOfferingsListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List service offerings",
		Long:  "List all service offerings",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			params := capi.NewQueryParams()

			offerings, err := client.ServiceOfferings().List(ctx, params)
			if err != nil {
				return fmt.Errorf("failed to list service offerings: %w", err)
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(offerings.Resources)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(offerings.Resources)
			default:
				if len(offerings.Resources) == 0 {
					fmt.Println("No service offerings found")
					return nil
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Name", "Description", "Broker", "Available", "Plans")

				for _, offering := range offerings.Resources {
					// Get broker name
					brokerName := ""
					if offering.Relationships.ServiceBroker.Data != nil {
						broker, _ := client.ServiceBrokers().Get(ctx, offering.Relationships.ServiceBroker.Data.GUID)
						if broker != nil {
							brokerName = broker.Name
						}
					}

					// Count plans
					planParams := capi.NewQueryParams()
					planParams.WithFilter("service_offering_guids", offering.GUID)
					plans, _ := client.ServicePlans().List(ctx, planParams)
					planCount := 0
					if plans != nil {
						planCount = len(plans.Resources)
					}

					available := "yes"
					if !offering.Available {
						available = "no"
					}

					description := offering.Description
					if len(description) > 50 {
						description = description[:47] + "..."
					}

					_ = table.Append(offering.Name, description, brokerName, available, fmt.Sprintf("%d", planCount))
				}

				_ = table.Render()
			}

			return nil
		},
	}

	return cmd
}

func newServicesOfferingsGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get OFFERING_NAME_OR_GUID",
		Short: "Get service offering details",
		Long:  "Display detailed information about a specific service offering",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Try to get by GUID first
			offering, err := client.ServiceOfferings().Get(ctx, nameOrGUID)
			if err != nil {
				// If not found by GUID, try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)
				offerings, err := client.ServiceOfferings().List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find service offering: %w", err)
				}
				if len(offerings.Resources) == 0 {
					return fmt.Errorf("service offering '%s' not found", nameOrGUID)
				}
				offering = &offerings.Resources[0]
			}

			fmt.Printf("Service Offering: %s\n", offering.Name)
			fmt.Printf("  GUID:        %s\n", offering.GUID)
			fmt.Printf("  Description: %s\n", offering.Description)
			fmt.Printf("  Available:   %t\n", offering.Available)
			fmt.Printf("  Shareable:   %t\n", offering.Shareable)
			fmt.Printf("  Created:     %s\n", offering.CreatedAt.Format("2006-01-02 15:04:05"))

			if len(offering.Tags) > 0 {
				fmt.Printf("  Tags:        %s\n", strings.Join(offering.Tags, ", "))
			}

			return nil
		},
	}
}

func newServicesPlansCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "plans",
		Aliases: []string{"plan"},
		Short:   "Manage service plans",
		Long:    "List and manage Cloud Foundry service plans",
	}

	cmd.AddCommand(newServicesPlansListCommand())
	cmd.AddCommand(newServicesPlansGetCommand())

	return cmd
}

func newServicesPlansListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List service plans",
		Long:  "List all service plans",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			params := capi.NewQueryParams()

			plans, err := client.ServicePlans().List(ctx, params)
			if err != nil {
				return fmt.Errorf("failed to list service plans: %w", err)
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(plans.Resources)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(plans.Resources)
			default:
				if len(plans.Resources) == 0 {
					fmt.Println("No service plans found")
					return nil
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Name", "Service", "Description", "Free", "Available")

				for _, plan := range plans.Resources {
					// Get service offering name
					offeringName := ""
					if plan.Relationships.ServiceOffering.Data != nil {
						offering, _ := client.ServiceOfferings().Get(ctx, plan.Relationships.ServiceOffering.Data.GUID)
						if offering != nil {
							offeringName = offering.Name
						}
					}

					available := "yes"
					if !plan.Available {
						available = "no"
					}

					free := "yes"
					if !plan.Free {
						free = "no"
					}

					description := plan.Description
					if len(description) > 40 {
						description = description[:37] + "..."
					}

					_ = table.Append(plan.Name, offeringName, description, free, available)
				}

				_ = table.Render()
			}

			return nil
		},
	}

	return cmd
}

func newServicesPlansGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get PLAN_NAME_OR_GUID",
		Short: "Get service plan details",
		Long:  "Display detailed information about a specific service plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Try to get by GUID first
			plan, err := client.ServicePlans().Get(ctx, nameOrGUID)
			if err != nil {
				// If not found by GUID, try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)
				plans, err := client.ServicePlans().List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find service plan: %w", err)
				}
				if len(plans.Resources) == 0 {
					return fmt.Errorf("service plan '%s' not found", nameOrGUID)
				}
				plan = &plans.Resources[0]
			}

			fmt.Printf("Service Plan: %s\n", plan.Name)
			fmt.Printf("  GUID:        %s\n", plan.GUID)
			fmt.Printf("  Description: %s\n", plan.Description)
			fmt.Printf("  Free:        %t\n", plan.Free)
			fmt.Printf("  Available:   %t\n", plan.Available)
			fmt.Printf("  Visibility:  %s\n", plan.VisibilityType)
			fmt.Printf("  Created:     %s\n", plan.CreatedAt.Format("2006-01-02 15:04:05"))

			// Get service offering name
			if plan.Relationships.ServiceOffering.Data != nil {
				offering, _ := client.ServiceOfferings().Get(ctx, plan.Relationships.ServiceOffering.Data.GUID)
				if offering != nil {
					fmt.Printf("  Offering:    %s\n", offering.Name)
				}
			}

			return nil
		},
	}
}
