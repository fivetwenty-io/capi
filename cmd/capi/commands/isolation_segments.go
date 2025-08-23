package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/fivetwenty-io/capi/v3/pkg/capi"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// NewIsolationSegmentsCommand creates the isolation-segments command group
func NewIsolationSegmentsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "isolation-segments",
		Aliases: []string{"isolation-segment", "iso-seg", "iso-segs"},
		Short:   "Manage isolation segments",
		Long:    "List and manage Cloud Foundry isolation segments",
	}

	cmd.AddCommand(newIsolationSegmentsListCommand())
	cmd.AddCommand(newIsolationSegmentsGetCommand())
	cmd.AddCommand(newIsolationSegmentsCreateCommand())
	cmd.AddCommand(newIsolationSegmentsUpdateCommand())
	cmd.AddCommand(newIsolationSegmentsDeleteCommand())
	cmd.AddCommand(newIsolationSegmentsEntitleOrgsCommand())
	cmd.AddCommand(newIsolationSegmentsRevokeOrgCommand())
	cmd.AddCommand(newIsolationSegmentsListOrgsCommand())
	cmd.AddCommand(newIsolationSegmentsListSpacesCommand())

	return cmd
}

func newIsolationSegmentsListCommand() *cobra.Command {
	var (
		allPages bool
		perPage  int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List isolation segments",
		Long:  "List all isolation segments the user has access to",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			params := capi.NewQueryParams()
			if perPage > 0 {
				params.PerPage = perPage
			}

			segments, err := client.IsolationSegments().List(ctx, params)
			if err != nil {
				return fmt.Errorf("failed to list isolation segments: %w", err)
			}

			// Fetch all pages if requested
			allSegments := segments.Resources
			if allPages && segments.Pagination.TotalPages > 1 {
				for page := 2; page <= segments.Pagination.TotalPages; page++ {
					params.Page = page
					moreSegments, err := client.IsolationSegments().List(ctx, params)
					if err != nil {
						return fmt.Errorf("failed to fetch page %d: %w", page, err)
					}
					allSegments = append(allSegments, moreSegments.Resources...)
				}
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(allSegments)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				encoder.SetIndent(2)
				return encoder.Encode(allSegments)
			default:
				if len(allSegments) == 0 {
					fmt.Println("No isolation segments found")
					return nil
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Name", "GUID", "Created", "Updated")

				for _, segment := range allSegments {
					createdAt := ""
					if !segment.CreatedAt.IsZero() {
						createdAt = segment.CreatedAt.Format("2006-01-02 15:04:05")
					}
					updatedAt := ""
					if !segment.UpdatedAt.IsZero() {
						updatedAt = segment.UpdatedAt.Format("2006-01-02 15:04:05")
					}

					_ = table.Append(segment.Name, segment.GUID, createdAt, updatedAt)
				}

				_ = table.Render()

				if !allPages && segments.Pagination.TotalPages > 1 {
					fmt.Printf("\nShowing page 1 of %d. Use --all to fetch all pages.\n", segments.Pagination.TotalPages)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&allPages, "all", false, "fetch all pages")
	cmd.Flags().IntVar(&perPage, "per-page", 50, "results per page")

	return cmd
}

func newIsolationSegmentsGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get ISOLATION_SEGMENT_NAME_OR_GUID",
		Short: "Get isolation segment details",
		Long:  "Display detailed information about a specific isolation segment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			segmentsClient := client.IsolationSegments()

			// Try to get by GUID first
			segment, err := segmentsClient.Get(ctx, nameOrGUID)
			if err != nil {
				// If not found by GUID, try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)
				segments, err := segmentsClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find isolation segment: %w", err)
				}
				if len(segments.Resources) == 0 {
					return fmt.Errorf("isolation segment '%s' not found", nameOrGUID)
				}
				segment = &segments.Resources[0]
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(segment)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(segment)
			default:
				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Property", "Value")

				_ = table.Append("Name", segment.Name)
				_ = table.Append("GUID", segment.GUID)
				_ = table.Append("Created", segment.CreatedAt.Format("2006-01-02 15:04:05"))
				_ = table.Append("Updated", segment.UpdatedAt.Format("2006-01-02 15:04:05"))

				fmt.Printf("Isolation Segment Details:\n\n")
				_ = table.Render()

				if len(segment.Metadata.Labels) > 0 {
					fmt.Println("\nLabels:")
					labelTable := tablewriter.NewWriter(os.Stdout)
					labelTable.Header("Key", "Value")
					for k, v := range segment.Metadata.Labels {
						_ = labelTable.Append(k, v)
					}
					_ = labelTable.Render()
				}

				if len(segment.Metadata.Annotations) > 0 {
					fmt.Println("\nAnnotations:")
					annotationTable := tablewriter.NewWriter(os.Stdout)
					annotationTable.Header("Key", "Value")
					for k, v := range segment.Metadata.Annotations {
						_ = annotationTable.Append(k, v)
					}
					_ = annotationTable.Render()
				}
			}

			return nil
		},
	}
}

func newIsolationSegmentsCreateCommand() *cobra.Command {
	var (
		name   string
		labels map[string]string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an isolation segment",
		Long:  "Create a new Cloud Foundry isolation segment",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("isolation segment name is required")
			}

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			createReq := &capi.IsolationSegmentCreateRequest{
				Name: name,
			}

			if labels != nil {
				createReq.Metadata = &capi.Metadata{
					Labels: labels,
				}
			}

			segment, err := client.IsolationSegments().Create(ctx, createReq)
			if err != nil {
				return fmt.Errorf("failed to create isolation segment: %w", err)
			}

			fmt.Printf("Successfully created isolation segment '%s' with GUID %s\n", segment.Name, segment.GUID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "isolation segment name (required)")
	cmd.Flags().StringToStringVar(&labels, "labels", nil, "labels to apply (key=value)")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newIsolationSegmentsUpdateCommand() *cobra.Command {
	var (
		newName string
		labels  map[string]string
	)

	cmd := &cobra.Command{
		Use:   "update ISOLATION_SEGMENT_NAME_OR_GUID",
		Short: "Update an isolation segment",
		Long:  "Update an existing Cloud Foundry isolation segment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			segmentsClient := client.IsolationSegments()

			// Find isolation segment
			var segmentGUID string
			segment, err := segmentsClient.Get(ctx, nameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)
				segments, err := segmentsClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find isolation segment: %w", err)
				}
				if len(segments.Resources) == 0 {
					return fmt.Errorf("isolation segment '%s' not found", nameOrGUID)
				}
				segmentGUID = segments.Resources[0].GUID
			} else {
				segmentGUID = segment.GUID
			}

			// Build update request
			updateReq := &capi.IsolationSegmentUpdateRequest{}

			if newName != "" {
				updateReq.Name = &newName
			}

			if labels != nil {
				updateReq.Metadata = &capi.Metadata{
					Labels: labels,
				}
			}

			// Update isolation segment
			updatedSegment, err := segmentsClient.Update(ctx, segmentGUID, updateReq)
			if err != nil {
				return fmt.Errorf("failed to update isolation segment: %w", err)
			}

			fmt.Printf("Successfully updated isolation segment '%s'\n", updatedSegment.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&newName, "name", "", "new isolation segment name")
	cmd.Flags().StringToStringVar(&labels, "labels", nil, "labels to apply (key=value)")

	return cmd
}

func newIsolationSegmentsDeleteCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete ISOLATION_SEGMENT_NAME_OR_GUID",
		Short: "Delete an isolation segment",
		Long:  "Delete a Cloud Foundry isolation segment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			if !force {
				fmt.Printf("Really delete isolation segment '%s'? (y/N): ", nameOrGUID)
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
			segmentsClient := client.IsolationSegments()

			// Find isolation segment
			var segmentGUID string
			var segmentName string
			segment, err := segmentsClient.Get(ctx, nameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)
				segments, err := segmentsClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find isolation segment: %w", err)
				}
				if len(segments.Resources) == 0 {
					return fmt.Errorf("isolation segment '%s' not found", nameOrGUID)
				}
				segmentGUID = segments.Resources[0].GUID
				segmentName = segments.Resources[0].Name
			} else {
				segmentGUID = segment.GUID
				segmentName = segment.Name
			}

			// Delete isolation segment
			err = segmentsClient.Delete(ctx, segmentGUID)
			if err != nil {
				return fmt.Errorf("failed to delete isolation segment: %w", err)
			}

			fmt.Printf("Successfully deleted isolation segment '%s'\n", segmentName)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "force deletion without confirmation")

	return cmd
}

func newIsolationSegmentsEntitleOrgsCommand() *cobra.Command {
	var orgNames []string

	cmd := &cobra.Command{
		Use:   "entitle-orgs ISOLATION_SEGMENT_NAME_OR_GUID",
		Short: "Entitle organizations to isolation segment",
		Long:  "Grant access to an isolation segment for specific organizations",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			if len(orgNames) == 0 {
				return fmt.Errorf("at least one organization name is required")
			}

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			segmentsClient := client.IsolationSegments()

			// Find isolation segment
			var segmentGUID string
			var segmentName string
			segment, err := segmentsClient.Get(ctx, nameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)
				segments, err := segmentsClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find isolation segment: %w", err)
				}
				if len(segments.Resources) == 0 {
					return fmt.Errorf("isolation segment '%s' not found", nameOrGUID)
				}
				segmentGUID = segments.Resources[0].GUID
				segmentName = segments.Resources[0].Name
			} else {
				segmentGUID = segment.GUID
				segmentName = segment.Name
			}

			// Find organizations by name
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

			// Entitle organizations
			_, err = segmentsClient.EntitleOrganizations(ctx, segmentGUID, orgGUIDs)
			if err != nil {
				return fmt.Errorf("failed to entitle organizations to isolation segment: %w", err)
			}

			fmt.Printf("Successfully entitled %d organizations to isolation segment '%s'\n", len(orgNames), segmentName)
			return nil
		},
	}

	cmd.Flags().StringSliceVar(&orgNames, "orgs", nil, "organization names to entitle (required)")
	_ = cmd.MarkFlagRequired("orgs")

	return cmd
}

func newIsolationSegmentsRevokeOrgCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "revoke-org ISOLATION_SEGMENT_NAME_OR_GUID ORG_NAME_OR_GUID",
		Short: "Revoke organization from isolation segment",
		Long:  "Remove organization's access to an isolation segment",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			segmentNameOrGUID := args[0]
			orgNameOrGUID := args[1]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			segmentsClient := client.IsolationSegments()

			// Find isolation segment
			var segmentGUID string
			var segmentName string
			segment, err := segmentsClient.Get(ctx, segmentNameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", segmentNameOrGUID)
				segments, err := segmentsClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find isolation segment: %w", err)
				}
				if len(segments.Resources) == 0 {
					return fmt.Errorf("isolation segment '%s' not found", segmentNameOrGUID)
				}
				segmentGUID = segments.Resources[0].GUID
				segmentName = segments.Resources[0].Name
			} else {
				segmentGUID = segment.GUID
				segmentName = segment.Name
			}

			// Find organization
			var orgGUID string
			var orgName string
			org, err := client.Organizations().Get(ctx, orgNameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", orgNameOrGUID)
				orgs, err := client.Organizations().List(ctx, params)
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

			// Revoke organization
			err = segmentsClient.RevokeOrganization(ctx, segmentGUID, orgGUID)
			if err != nil {
				return fmt.Errorf("failed to revoke organization from isolation segment: %w", err)
			}

			fmt.Printf("Successfully revoked organization '%s' from isolation segment '%s'\n", orgName, segmentName)
			return nil
		},
	}
}

func newIsolationSegmentsListOrgsCommand() *cobra.Command {
	var (
		allPages bool
		perPage  int
	)

	cmd := &cobra.Command{
		Use:   "list-orgs ISOLATION_SEGMENT_NAME_OR_GUID",
		Short: "List organizations entitled to isolation segment",
		Long:  "List all organizations that have access to an isolation segment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			segmentsClient := client.IsolationSegments()

			// Find isolation segment
			var segmentGUID string
			var segmentName string
			segment, err := segmentsClient.Get(ctx, nameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)
				segments, err := segmentsClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find isolation segment: %w", err)
				}
				if len(segments.Resources) == 0 {
					return fmt.Errorf("isolation segment '%s' not found", nameOrGUID)
				}
				segmentGUID = segments.Resources[0].GUID
				segmentName = segments.Resources[0].Name
			} else {
				segmentGUID = segment.GUID
				segmentName = segment.Name
			}

			// List organizations for isolation segment
			orgs, err := segmentsClient.ListOrganizations(ctx, segmentGUID)
			if err != nil {
				return fmt.Errorf("failed to list organizations for isolation segment: %w", err)
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(orgs.Resources)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				encoder.SetIndent(2)
				return encoder.Encode(orgs.Resources)
			default:
				if len(orgs.Resources) == 0 {
					fmt.Printf("No organizations found for isolation segment '%s'\n", segmentName)
					return nil
				}

				fmt.Printf("Organizations entitled to isolation segment '%s':\n\n", segmentName)
				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Name", "GUID", "Created", "Updated")

				for _, org := range orgs.Resources {
					createdAt := ""
					if !org.CreatedAt.IsZero() {
						createdAt = org.CreatedAt.Format("2006-01-02 15:04:05")
					}
					updatedAt := ""
					if !org.UpdatedAt.IsZero() {
						updatedAt = org.UpdatedAt.Format("2006-01-02 15:04:05")
					}

					_ = table.Append(org.Name, org.GUID, createdAt, updatedAt)
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

func newIsolationSegmentsListSpacesCommand() *cobra.Command {
	var (
		allPages bool
		perPage  int
	)

	cmd := &cobra.Command{
		Use:   "list-spaces ISOLATION_SEGMENT_NAME_OR_GUID",
		Short: "List spaces using isolation segment",
		Long:  "List all spaces that use an isolation segment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			segmentsClient := client.IsolationSegments()

			// Find isolation segment
			var segmentGUID string
			var segmentName string
			segment, err := segmentsClient.Get(ctx, nameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)
				segments, err := segmentsClient.List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find isolation segment: %w", err)
				}
				if len(segments.Resources) == 0 {
					return fmt.Errorf("isolation segment '%s' not found", nameOrGUID)
				}
				segmentGUID = segments.Resources[0].GUID
				segmentName = segments.Resources[0].Name
			} else {
				segmentGUID = segment.GUID
				segmentName = segment.Name
			}

			// List spaces for isolation segment
			spaces, err := segmentsClient.ListSpaces(ctx, segmentGUID)
			if err != nil {
				return fmt.Errorf("failed to list spaces for isolation segment: %w", err)
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(spaces.Resources)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				encoder.SetIndent(2)
				return encoder.Encode(spaces.Resources)
			default:
				if len(spaces.Resources) == 0 {
					fmt.Printf("No spaces found for isolation segment '%s'\n", segmentName)
					return nil
				}

				fmt.Printf("Spaces using isolation segment '%s':\n\n", segmentName)
				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Name", "GUID", "Organization", "Created")

				for _, space := range spaces.Resources {
					createdAt := ""
					if !space.CreatedAt.IsZero() {
						createdAt = space.CreatedAt.Format("2006-01-02 15:04:05")
					}

					// Get org name if available
					orgName := ""
					if space.Relationships.Organization.Data != nil {
						org, _ := client.Organizations().Get(ctx, space.Relationships.Organization.Data.GUID)
						if org != nil {
							orgName = org.Name
						}
					}

					_ = table.Append(space.Name, space.GUID, orgName, createdAt)
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
