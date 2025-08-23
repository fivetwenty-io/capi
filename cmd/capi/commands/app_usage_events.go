package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/fivetwenty-io/capi/pkg/capi"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// NewAppUsageEventsCommand creates the app usage events command group
func NewAppUsageEventsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "app-usage-events",
		Aliases: []string{"app-usage", "app-events", "aue"},
		Short:   "Manage application usage events",
		Long:    "View and manage application usage events for monitoring and billing",
	}

	cmd.AddCommand(newAppUsageEventsListCommand())
	cmd.AddCommand(newAppUsageEventsGetCommand())
	cmd.AddCommand(newAppUsageEventsPurgeReseedCommand())

	return cmd
}

func newAppUsageEventsListCommand() *cobra.Command {
	var (
		allPages  bool
		perPage   int
		afterGUID string
		appName   string
		spaceName string
		orgName   string
		startTime string
		endTime   string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List application usage events",
		Long:  "List application usage events with optional filtering",
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

			if appName != "" {
				params.WithFilter("app_names", appName)
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

			events, err := client.AppUsageEvents().List(ctx, params)
			if err != nil {
				return fmt.Errorf("failed to list app usage events: %w", err)
			}

			// Fetch all pages if requested
			allEvents := events.Resources
			if allPages && events.Pagination.TotalPages > 1 {
				for page := 2; page <= events.Pagination.TotalPages; page++ {
					params.Page = page
					moreEvents, err := client.AppUsageEvents().List(ctx, params)
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
					fmt.Println("No app usage events found")
					return nil
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.Header("GUID", "App Name", "Space", "Organization", "State", "Instances", "Memory MB", "Created")

				for _, event := range allEvents {
					// Format previous state
					previousState := "N/A"
					if event.PreviousState != nil {
						previousState = *event.PreviousState
					}

					stateTransition := fmt.Sprintf("%s -> %s", previousState, event.State)

					_ = table.Append(
						event.GUID,
						event.AppName,
						event.SpaceName,
						event.OrganizationName,
						stateTransition,
						fmt.Sprintf("%d", event.InstanceCount),
						fmt.Sprintf("%d", event.MemoryInMBPerInstance),
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
	cmd.Flags().StringVar(&appName, "app-name", "", "filter by application name")
	cmd.Flags().StringVar(&spaceName, "space-name", "", "filter by space name")
	cmd.Flags().StringVar(&orgName, "org-name", "", "filter by organization name")
	cmd.Flags().StringVar(&startTime, "start-time", "", "filter events after this time (RFC3339 format)")
	cmd.Flags().StringVar(&endTime, "end-time", "", "filter events before this time (RFC3339 format)")

	return cmd
}

func newAppUsageEventsGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get EVENT_GUID",
		Short: "Get app usage event details",
		Long:  "Display detailed information about a specific app usage event",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			eventGUID := args[0]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			event, err := client.AppUsageEvents().Get(ctx, eventGUID)
			if err != nil {
				return fmt.Errorf("failed to get app usage event: %w", err)
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
				fmt.Printf("App Usage Event: %s\n", event.GUID)
				fmt.Printf("  Created: %s\n", event.CreatedAt.Format("2006-01-02 15:04:05"))
				fmt.Printf("  Updated: %s\n", event.UpdatedAt.Format("2006-01-02 15:04:05"))
				fmt.Println()

				fmt.Println("Application Information:")
				fmt.Printf("  App Name: %s\n", event.AppName)
				fmt.Printf("  App GUID: %s\n", event.AppGUID)
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
				fmt.Println()

				fmt.Println("Resource Usage:")
				fmt.Printf("  Instance Count: %d\n", event.InstanceCount)
				if event.PreviousInstanceCount != nil {
					fmt.Printf("  Previous Instance Count: %d\n", *event.PreviousInstanceCount)
				}
				fmt.Printf("  Memory per Instance: %d MB\n", event.MemoryInMBPerInstance)
				if event.PreviousMemoryInMBPerInstance != nil {
					fmt.Printf("  Previous Memory per Instance: %d MB\n", *event.PreviousMemoryInMBPerInstance)
				}
				fmt.Println()

				fmt.Println("Build Information:")
				if event.BuildpackName != nil {
					fmt.Printf("  Buildpack Name: %s\n", *event.BuildpackName)
				}
				if event.BuildpackGUID != nil {
					fmt.Printf("  Buildpack GUID: %s\n", *event.BuildpackGUID)
				}
				fmt.Printf("  Package State: %s\n", event.Package.State)
				fmt.Println()

				fmt.Println("Process Information:")
				fmt.Printf("  Process Type: %s\n", event.ProcessType)
				if event.TaskName != nil {
					fmt.Printf("  Task Name: %s\n", *event.TaskName)
				}
				if event.TaskGUID != nil {
					fmt.Printf("  Task GUID: %s\n", *event.TaskGUID)
				}
				if event.ParentAppName != nil {
					fmt.Printf("  Parent App Name: %s\n", *event.ParentAppName)
				}
				if event.ParentAppGUID != nil {
					fmt.Printf("  Parent App GUID: %s\n", *event.ParentAppGUID)
				}
			}

			return nil
		},
	}
}

func newAppUsageEventsPurgeReseedCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "purge-and-reseed",
		Short: "Purge and reseed app usage events",
		Long:  "Purge existing app usage events and reseed with current state",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				fmt.Print("This will purge all existing app usage events and reseed with current state.\n")
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

			fmt.Println("Purging and reseeding app usage events...")
			start := time.Now()

			err = client.AppUsageEvents().PurgeAndReseed(ctx)
			if err != nil {
				return fmt.Errorf("failed to purge and reseed app usage events: %w", err)
			}

			duration := time.Since(start)
			fmt.Printf("Successfully purged and reseeded app usage events in %v\n", duration)
			fmt.Println("New events will reflect the current state of all applications")

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "skip confirmation prompt")

	return cmd
}
