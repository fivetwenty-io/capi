package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/fivetwenty-io/capi/v3/pkg/capi"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// NewServiceUsageEventsCommand creates the service usage events command group
func NewServiceUsageEventsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "service-usage-events",
		Aliases: []string{"service-usage", "service-events", "sue"},
		Short:   "Manage service usage events",
		Long:    "View and manage service usage events for monitoring and billing",
	}

	cmd.AddCommand(newServiceUsageEventsListCommand())
	cmd.AddCommand(newServiceUsageEventsGetCommand())
	cmd.AddCommand(newServiceUsageEventsPurgeReseedCommand())

	return cmd
}

func newServiceUsageEventsListCommand() *cobra.Command {
	var (
		allPages            bool
		perPage             int
		afterGUID           string
		serviceInstanceName string
		servicePlanName     string
		serviceOfferingName string
		serviceBrokerName   string
		spaceName           string
		orgName             string
		startTime           string
		endTime             string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List service usage events",
		Long:  "List service usage events with optional filtering",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			params := capi.NewQueryParams()
			params.PerPage = perPage

			// Add filters
			if afterGUID != "" {
				params.WithFilter("guids", afterGUID)
			}

			if serviceInstanceName != "" {
				params.WithFilter("service_instance_names", serviceInstanceName)
			}

			if servicePlanName != "" {
				params.WithFilter("service_plan_names", servicePlanName)
			}

			if serviceOfferingName != "" {
				params.WithFilter("service_offering_names", serviceOfferingName)
			}

			if serviceBrokerName != "" {
				params.WithFilter("service_broker_names", serviceBrokerName)
			}

			if spaceName != "" {
				params.WithFilter("space_names", spaceName)
			}

			if orgName != "" {
				params.WithFilter("organization_names", orgName)
			}

			if startTime != "" {
				params.WithFilter("created_ats[gte]", startTime)
			}

			if endTime != "" {
				params.WithFilter("created_ats[lte]", endTime)
			}

			events, err := client.ServiceUsageEvents().List(ctx, params)
			if err != nil {
				return fmt.Errorf("failed to list service usage events: %w", err)
			}

			// Fetch all pages if requested
			allEvents := events.Resources
			if allPages && events.Pagination.TotalPages > 1 {
				for page := 2; page <= events.Pagination.TotalPages; page++ {
					params.Page = page
					moreEvents, err := client.ServiceUsageEvents().List(ctx, params)
					if err != nil {
						return fmt.Errorf("failed to fetch page %d: %w", page, err)
					}
					allEvents = append(allEvents, moreEvents.Resources...)
				}
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(allEvents)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(allEvents)
			default:
				if len(allEvents) == 0 {
					fmt.Println("No service usage events found")
					return nil
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.Header("GUID", "Service Instance", "Service Plan", "Service Offering", "State", "Space", "Organization", "Created")

				for _, event := range allEvents {
					// Format previous state
					previousState := "N/A"
					if event.PreviousState != nil {
						previousState = *event.PreviousState
					}

					stateTransition := fmt.Sprintf("%s -> %s", previousState, event.State)

					_ = table.Append(
						event.GUID,
						event.ServiceInstanceName,
						event.ServicePlanName,
						event.ServiceOfferingName,
						stateTransition,
						event.SpaceName,
						event.OrganizationName,
						event.CreatedAt.Format("2006-01-02 15:04:05"),
					)
				}

				_ = table.Render()

				if !allPages && events.Pagination.TotalPages > 1 {
					fmt.Printf("\nShowing page 1 of %d. Use --all to fetch all pages.\n", events.Pagination.TotalPages)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&allPages, "all", false, "fetch all pages")
	cmd.Flags().IntVar(&perPage, "per-page", 50, "results per page")
	cmd.Flags().StringVar(&afterGUID, "after-guid", "", "return events after this GUID")
	cmd.Flags().StringVar(&serviceInstanceName, "service-instance-name", "", "filter by service instance name")
	cmd.Flags().StringVar(&servicePlanName, "service-plan-name", "", "filter by service plan name")
	cmd.Flags().StringVar(&serviceOfferingName, "service-offering-name", "", "filter by service offering name")
	cmd.Flags().StringVar(&serviceBrokerName, "service-broker-name", "", "filter by service broker name")
	cmd.Flags().StringVar(&spaceName, "space-name", "", "filter by space name")
	cmd.Flags().StringVar(&orgName, "org-name", "", "filter by organization name")
	cmd.Flags().StringVar(&startTime, "start-time", "", "filter events after this time (RFC3339 format)")
	cmd.Flags().StringVar(&endTime, "end-time", "", "filter events before this time (RFC3339 format)")

	return cmd
}

func newServiceUsageEventsGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get EVENT_GUID",
		Short: "Get service usage event details",
		Long:  "Display detailed information about a specific service usage event",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			eventGUID := args[0]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			event, err := client.ServiceUsageEvents().Get(ctx, eventGUID)
			if err != nil {
				return fmt.Errorf("failed to get service usage event: %w", err)
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(event)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(event)
			default:
				fmt.Printf("Service Usage Event: %s\n", event.GUID)
				fmt.Printf("  Created: %s\n", event.CreatedAt.Format("2006-01-02 15:04:05"))
				fmt.Printf("  Updated: %s\n", event.UpdatedAt.Format("2006-01-02 15:04:05"))
				fmt.Println()

				fmt.Println("Service Instance Information:")
				fmt.Printf("  Service Instance Name: %s\n", event.ServiceInstanceName)
				fmt.Printf("  Service Instance GUID: %s\n", event.ServiceInstanceGUID)
				fmt.Printf("  Service Instance Type: %s\n", event.ServiceInstanceType)
				fmt.Println()

				fmt.Println("Service Plan Information:")
				fmt.Printf("  Service Plan Name: %s\n", event.ServicePlanName)
				fmt.Printf("  Service Plan GUID: %s\n", event.ServicePlanGUID)
				fmt.Println()

				fmt.Println("Service Offering Information:")
				fmt.Printf("  Service Offering Name: %s\n", event.ServiceOfferingName)
				fmt.Printf("  Service Offering GUID: %s\n", event.ServiceOfferingGUID)
				fmt.Println()

				fmt.Println("Service Broker Information:")
				fmt.Printf("  Service Broker Name: %s\n", event.ServiceBrokerName)
				fmt.Printf("  Service Broker GUID: %s\n", event.ServiceBrokerGUID)
				fmt.Println()

				fmt.Println("Location Information:")
				fmt.Printf("  Space Name: %s\n", event.SpaceName)
				fmt.Printf("  Space GUID: %s\n", event.SpaceGUID)
				fmt.Printf("  Organization Name: %s\n", event.OrganizationName)
				fmt.Printf("  Organization GUID: %s\n", event.OrganizationGUID)
				fmt.Println()

				fmt.Println("State Information:")
				fmt.Printf("  Current State: %s\n", event.State)
				if event.PreviousState != nil {
					fmt.Printf("  Previous State: %s\n", *event.PreviousState)
				}
			}

			return nil
		},
	}
}

func newServiceUsageEventsPurgeReseedCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "purge-and-reseed",
		Short: "Purge and reseed service usage events",
		Long:  "Purge existing service usage events and reseed with current state",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				fmt.Print("This will purge all existing service usage events and reseed with current state.\n")
				fmt.Print("This action cannot be undone. Continue? (y/N): ")
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

			fmt.Println("Purging and reseeding service usage events...")
			start := time.Now()

			err = client.ServiceUsageEvents().PurgeAndReseed(ctx)
			if err != nil {
				return fmt.Errorf("failed to purge and reseed service usage events: %w", err)
			}

			duration := time.Since(start)
			fmt.Printf("Successfully purged and reseeded service usage events in %v\n", duration)
			fmt.Println("New events will reflect the current state of all service instances")

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "skip confirmation prompt")

	return cmd
}
