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

// NewAuditEventsCommand creates the audit events command group
func NewAuditEventsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "audit-events",
		Aliases: []string{"audit", "events", "ae"},
		Short:   "Manage audit events",
		Long:    "View audit events for tracking system changes and user actions",
	}

	cmd.AddCommand(newAuditEventsListCommand())
	cmd.AddCommand(newAuditEventsGetCommand())

	return cmd
}

func newAuditEventsListCommand() *cobra.Command {
	var (
		allPages    bool
		perPage     int
		eventTypes  []string
		targetTypes []string
		actorTypes  []string
		spaceName   string
		orgName     string
		startTime   string
		endTime     string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List audit events",
		Long:  "List audit events with optional filtering",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			params := capi.NewQueryParams()
			params.PerPage = perPage

			// Add filters
			if len(eventTypes) > 0 {
				params.WithFilter("types", strings.Join(eventTypes, ","))
			}

			if len(targetTypes) > 0 {
				params.WithFilter("target_types", strings.Join(targetTypes, ","))
			}

			if len(actorTypes) > 0 {
				params.WithFilter("actor_types", strings.Join(actorTypes, ","))
			}

			if spaceName != "" {
				// Find space by name to get GUID
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
			}

			if orgName != "" {
				// Find organization by name to get GUID
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

			if startTime != "" {
				params.WithFilter("created_ats[gte]", startTime)
			}

			if endTime != "" {
				params.WithFilter("created_ats[lte]", endTime)
			}

			events, err := client.AuditEvents().List(ctx, params)
			if err != nil {
				return fmt.Errorf("failed to list audit events: %w", err)
			}

			// Fetch all pages if requested
			allEvents := events.Resources
			if allPages && events.Pagination.TotalPages > 1 {
				for page := 2; page <= events.Pagination.TotalPages; page++ {
					params.Page = page
					moreEvents, err := client.AuditEvents().List(ctx, params)
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
					fmt.Println("No audit events found")
					return nil
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.Header("GUID", "Type", "Actor", "Target", "Space", "Organization", "Created")

				for _, event := range allEvents {
					actorInfo := fmt.Sprintf("%s (%s)", event.Actor.Name, event.Actor.Type)
					if len(actorInfo) > 30 {
						actorInfo = actorInfo[:27] + "..."
					}

					targetInfo := fmt.Sprintf("%s (%s)", event.Target.Name, event.Target.Type)
					if len(targetInfo) > 30 {
						targetInfo = targetInfo[:27] + "..."
					}

					spaceName := "N/A"
					if event.Space != nil {
						spaceName = event.Space.Name
					}

					orgName := "N/A"
					if event.Organization != nil {
						orgName = event.Organization.Name
					}

					_ = table.Append(
						event.GUID,
						event.Type,
						actorInfo,
						targetInfo,
						spaceName,
						orgName,
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
	cmd.Flags().StringSliceVar(&eventTypes, "event-types", nil, "filter by event types (comma-separated)")
	cmd.Flags().StringSliceVar(&targetTypes, "target-types", nil, "filter by target types (comma-separated)")
	cmd.Flags().StringSliceVar(&actorTypes, "actor-types", nil, "filter by actor types (comma-separated)")
	cmd.Flags().StringVar(&spaceName, "space-name", "", "filter by space name")
	cmd.Flags().StringVar(&orgName, "org-name", "", "filter by organization name")
	cmd.Flags().StringVar(&startTime, "start-time", "", "filter events after this time (RFC3339 format)")
	cmd.Flags().StringVar(&endTime, "end-time", "", "filter events before this time (RFC3339 format)")

	return cmd
}

func newAuditEventsGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get EVENT_GUID",
		Short: "Get audit event details",
		Long:  "Display detailed information about a specific audit event",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			eventGUID := args[0]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			event, err := client.AuditEvents().Get(ctx, eventGUID)
			if err != nil {
				return fmt.Errorf("failed to get audit event: %w", err)
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
				fmt.Printf("Audit Event: %s\n", event.GUID)
				fmt.Printf("  Type: %s\n", event.Type)
				fmt.Printf("  Created: %s\n", event.CreatedAt.Format("2006-01-02 15:04:05"))
				fmt.Printf("  Updated: %s\n", event.UpdatedAt.Format("2006-01-02 15:04:05"))
				fmt.Println()

				fmt.Println("Actor Information:")
				fmt.Printf("  GUID: %s\n", event.Actor.GUID)
				fmt.Printf("  Type: %s\n", event.Actor.Type)
				fmt.Printf("  Name: %s\n", event.Actor.Name)
				fmt.Println()

				fmt.Println("Target Information:")
				fmt.Printf("  GUID: %s\n", event.Target.GUID)
				fmt.Printf("  Type: %s\n", event.Target.Type)
				fmt.Printf("  Name: %s\n", event.Target.Name)
				fmt.Println()

				if event.Space != nil {
					fmt.Println("Space Information:")
					fmt.Printf("  GUID: %s\n", event.Space.GUID)
					fmt.Printf("  Name: %s\n", event.Space.Name)
					fmt.Println()
				}

				if event.Organization != nil {
					fmt.Println("Organization Information:")
					fmt.Printf("  GUID: %s\n", event.Organization.GUID)
					fmt.Printf("  Name: %s\n", event.Organization.Name)
					fmt.Println()
				}

				if len(event.Data) > 0 {
					fmt.Println("Event Data:")
					for key, value := range event.Data {
						// Handle nested objects
						if valueMap, ok := value.(map[string]interface{}); ok {
							fmt.Printf("  %s:\n", key)
							for subKey, subValue := range valueMap {
								fmt.Printf("    %s: %v\n", subKey, subValue)
							}
						} else {
							fmt.Printf("  %s: %v\n", key, value)
						}
					}
				}
			}

			return nil
		},
	}
}
