package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/fivetwenty-io/capi-client/pkg/capi"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// NewRoutesCommand creates the routes command group
func NewRoutesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "routes",
		Aliases: []string{"route"},
		Short:   "Manage routes",
		Long:    "List and manage Cloud Foundry routes",
	}

	cmd.AddCommand(newRoutesListCommand())
	cmd.AddCommand(newRoutesCreateCommand())
	cmd.AddCommand(newRoutesDeleteCommand())

	return cmd
}

func newRoutesListCommand() *cobra.Command {
	var spaceName string
	var domainName string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List routes",
		Long:  "List all routes the user has access to",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			params := capi.NewQueryParams()

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

			// Filter by domain if specified
			if domainName != "" {
				// Find domain by name
				domainParams := capi.NewQueryParams()
				domainParams.WithFilter("names", domainName)

				domains, err := client.Domains().List(ctx, domainParams)
				if err != nil {
					return fmt.Errorf("failed to find domain: %w", err)
				}
				if len(domains.Resources) == 0 {
					return fmt.Errorf("domain '%s' not found", domainName)
				}

				params.WithFilter("domain_guids", domains.Resources[0].GUID)
			}

			routes, err := client.Routes().List(ctx, params)
			if err != nil {
				return fmt.Errorf("failed to list routes: %w", err)
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(routes.Resources)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(routes.Resources)
			default:
				if len(routes.Resources) == 0 {
					fmt.Println("No routes found")
					return nil
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.Header("URL", "GUID", "Protocol", "Host", "Path", "Port", "Destinations", "Created", "Updated")

				for _, route := range routes.Resources {
					path := route.Path
					if path == "" {
						path = "/"
					}

					port := ""
					if route.Port != nil {
						port = strconv.Itoa(*route.Port)
					}

					destinations := ""
					if len(route.Destinations) > 0 {
						destCount := len(route.Destinations)
						if destCount == 1 {
							destinations = "1 destination"
						} else {
							destinations = fmt.Sprintf("%d destinations", destCount)
						}
					}

					created := ""
					if !route.CreatedAt.IsZero() {
						created = route.CreatedAt.Format("2006-01-02 15:04:05")
					}

					updated := ""
					if !route.UpdatedAt.IsZero() {
						updated = route.UpdatedAt.Format("2006-01-02 15:04:05")
					}

					_ = table.Append(route.URL, route.GUID, route.Protocol, route.Host, path, port, destinations, created, updated)
				}

				_ = table.Render()
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&spaceName, "space", "s", "", "filter by space name")
	cmd.Flags().StringVarP(&domainName, "domain", "d", "", "filter by domain name")

	return cmd
}

func newRoutesCreateCommand() *cobra.Command {
	var spaceName string
	var hostname string
	var path string
	var port int

	cmd := &cobra.Command{
		Use:   "create DOMAIN_NAME",
		Short: "Create a route",
		Long:  "Create a new Cloud Foundry route in a space",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			domainName := args[0]

			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find domain by name
			domainParams := capi.NewQueryParams()
			domainParams.WithFilter("names", domainName)

			domains, err := client.Domains().List(ctx, domainParams)
			if err != nil {
				return fmt.Errorf("failed to find domain: %w", err)
			}
			if len(domains.Resources) == 0 {
				return fmt.Errorf("domain '%s' not found", domainName)
			}

			domain := domains.Resources[0]

			// Find space
			var spaceGUID string
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

				spaceGUID = spaces.Resources[0].GUID
			} else if targetedSpaceGUID := viper.GetString("space_guid"); targetedSpaceGUID != "" {
				spaceGUID = targetedSpaceGUID
			} else {
				return fmt.Errorf("no space specified and no space targeted")
			}

			// Create route request
			createReq := &capi.RouteCreateRequest{
				Relationships: capi.RouteRelationships{
					Space: capi.Relationship{
						Data: &capi.RelationshipData{GUID: spaceGUID},
					},
					Domain: capi.Relationship{
						Data: &capi.RelationshipData{GUID: domain.GUID},
					},
				},
			}

			if hostname != "" {
				createReq.Host = &hostname
			}

			if path != "" {
				createReq.Path = &path
			}

			if cmd.Flags().Changed("port") {
				createReq.Port = &port
			}

			// Create route
			route, err := client.Routes().Create(ctx, createReq)
			if err != nil {
				return fmt.Errorf("failed to create route: %w", err)
			}

			fmt.Printf("Successfully created route: %s\n", route.URL)
			fmt.Printf("  GUID: %s\n", route.GUID)
			fmt.Printf("  Protocol: %s\n", route.Protocol)
			fmt.Printf("  Host: %s\n", route.Host)
			if route.Path != "" {
				fmt.Printf("  Path: %s\n", route.Path)
			}
			if route.Port != nil {
				fmt.Printf("  Port: %d\n", *route.Port)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&spaceName, "space", "s", "", "space name (defaults to targeted space)")
	cmd.Flags().StringVar(&hostname, "hostname", "", "hostname for the route")
	cmd.Flags().StringVar(&path, "path", "", "path for the route")
	cmd.Flags().IntVar(&port, "port", 0, "port for the route (for TCP routes)")

	return cmd
}

func newRoutesDeleteCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete ROUTE_GUID_OR_URL",
		Short: "Delete a route",
		Long:  "Delete a Cloud Foundry route by GUID or URL",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			routeIdentifier := args[0]

			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Try to find route by GUID first, then by URL
			var routeGUID string
			var routeURL string

			// Try as GUID first
			route, err := client.Routes().Get(ctx, routeIdentifier)
			if err == nil {
				routeGUID = route.GUID
				routeURL = route.URL
			} else {
				// Try to find by URL (this would require listing and filtering)
				params := capi.NewQueryParams()
				routes, err := client.Routes().List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to list routes: %w", err)
				}

				for _, r := range routes.Resources {
					if r.URL == routeIdentifier {
						routeGUID = r.GUID
						routeURL = r.URL
						break
					}
				}

				if routeGUID == "" {
					return fmt.Errorf("route '%s' not found", routeIdentifier)
				}
			}

			// Confirm deletion unless forced
			if !force {
				fmt.Printf("Are you sure you want to delete route '%s'? (y/N): ", routeURL)
				var response string
				_, _ = fmt.Scanln(&response)
				if response != "y" && response != "Y" && response != "yes" && response != "YES" {
					fmt.Println("Deletion cancelled.")
					return nil
				}
			}

			// Delete route
			job, err := client.Routes().Delete(ctx, routeGUID)
			if err != nil {
				return fmt.Errorf("failed to delete route: %w", err)
			}

			if job != nil {
				fmt.Printf("Route deletion job created: %s\n", job.GUID)
				fmt.Printf("Successfully initiated deletion of route '%s'\n", routeURL)
			} else {
				fmt.Printf("Successfully deleted route '%s'\n", routeURL)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "force deletion without confirmation")

	return cmd
}
