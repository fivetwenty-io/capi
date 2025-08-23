package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fivetwenty-io/capi/v3/pkg/capi"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// NewResourceMatchesCommand creates the resource matches command group
func NewResourceMatchesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "resource-matches",
		Aliases: []string{"resources", "resource", "rm"},
		Short:   "Manage resource matches",
		Long:    "Create resource matches for optimizing package uploads",
	}

	cmd.AddCommand(newResourceMatchesCreateCommand())

	return cmd
}

// validateFilePathResourceMatches validates that a file path is safe to read
func validateFilePathResourceMatches(filePath string) error {
	// Clean the path to resolve any path traversal attempts
	cleanPath := filepath.Clean(filePath)

	// Check for path traversal attempts
	if filepath.IsAbs(filePath) {
		// Allow absolute paths but ensure they're clean
		if cleanPath != filePath {
			return fmt.Errorf("invalid file path: potential path traversal attempt")
		}
	} else {
		// For relative paths, ensure they don't escape the current directory
		if len(cleanPath) > 0 && cleanPath[0] == '.' && len(cleanPath) > 1 && cleanPath[1] == '.' {
			return fmt.Errorf("invalid file path: path traversal not allowed")
		}
	}

	// Check if file exists and is readable
	if _, err := os.Stat(cleanPath); err != nil {
		return fmt.Errorf("file not accessible: %w", err)
	}

	return nil
}

func newResourceMatchesCreateCommand() *cobra.Command {
	var (
		fromFile  string
		resources []string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create resource matches",
		Long:  "Create resource matches to check which resources already exist on the platform",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			var resourceList []capi.ResourceMatch

			// Load resources from file if specified
			if fromFile != "" {
				// Validate file path to prevent directory traversal
				if err := validateFilePathResourceMatches(fromFile); err != nil {
					return fmt.Errorf("invalid resource file: %w", err)
				}

				fileResources, err := loadResourcesFromFile(fromFile)
				if err != nil {
					return fmt.Errorf("failed to load resources from file: %w", err)
				}
				resourceList = append(resourceList, fileResources...)
			}

			// Parse individual resource entries
			for _, resourceStr := range resources {
				resource, err := parseResourceString(resourceStr)
				if err != nil {
					return fmt.Errorf("failed to parse resource '%s': %w", resourceStr, err)
				}
				resourceList = append(resourceList, resource)
			}

			if len(resourceList) == 0 {
				return fmt.Errorf("no resources specified. Use --from-file or --resource flags")
			}

			// Create resource matches request
			request := &capi.ResourceMatchesRequest{
				Resources: resourceList,
			}

			matches, err := client.ResourceMatches().Create(ctx, request)
			if err != nil {
				return fmt.Errorf("failed to create resource matches: %w", err)
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(matches)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(matches)
			default:
				if len(matches.Resources) == 0 {
					fmt.Println("No matching resources found on the platform")
					fmt.Printf("All %d resources will need to be uploaded\n", len(resourceList))
					return nil
				}

				fmt.Printf("Resource Matches Summary:\n")
				fmt.Printf("  Total resources checked: %d\n", len(resourceList))
				fmt.Printf("  Matching resources found: %d\n", len(matches.Resources))
				fmt.Printf("  Resources to upload: %d\n", len(resourceList)-len(matches.Resources))
				fmt.Println()

				table := tablewriter.NewWriter(os.Stdout)
				table.Header("SHA1", "Size", "Path", "Mode")

				for _, resource := range matches.Resources {
					_ = table.Append(
						resource.SHA1,
						fmt.Sprintf("%d", resource.Size),
						resource.Path,
						resource.Mode,
					)
				}

				fmt.Println("Matching resources (already exist on platform):")
				_ = table.Render()

				// Calculate size savings
				var totalSizeToUpload int64
				var totalSizeMatched int64

				for _, resource := range resourceList {
					totalSizeToUpload += resource.Size
				}

				for _, resource := range matches.Resources {
					totalSizeMatched += resource.Size
				}

				sizeSavings := totalSizeMatched
				fmt.Printf("\nUpload optimization:\n")
				fmt.Printf("  Total size without optimization: %d bytes\n", totalSizeToUpload)
				fmt.Printf("  Size of matching resources: %d bytes\n", totalSizeMatched)
				fmt.Printf("  Actual upload size needed: %d bytes\n", totalSizeToUpload-totalSizeMatched)
				if totalSizeToUpload > 0 {
					percentSaved := float64(sizeSavings) / float64(totalSizeToUpload) * 100
					fmt.Printf("  Bandwidth saved: %.1f%%\n", percentSaved)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&fromFile, "from-file", "f", "", "load resources from file (JSON or YAML format)")
	cmd.Flags().StringSliceVarP(&resources, "resource", "r", nil, "resource specification (sha1:size:path:mode)")

	return cmd
}

// parseResourceString parses a resource string in the format "sha1:size:path:mode"
func parseResourceString(resourceStr string) (capi.ResourceMatch, error) {
	parts := strings.Split(resourceStr, ":")
	if len(parts) != 4 {
		return capi.ResourceMatch{}, fmt.Errorf("invalid format. Expected sha1:size:path:mode")
	}

	sha1 := parts[0]
	sizeStr := parts[1]
	path := parts[2]
	mode := parts[3]

	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return capi.ResourceMatch{}, fmt.Errorf("invalid size '%s': %w", sizeStr, err)
	}

	return capi.ResourceMatch{
		SHA1: sha1,
		Size: size,
		Path: path,
		Mode: mode,
	}, nil
}

// loadResourcesFromFile loads resources from a JSON or YAML file
func loadResourcesFromFile(filename string) ([]capi.ResourceMatch, error) {
	data, err := os.ReadFile(filepath.Clean(filename))
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var resources []capi.ResourceMatch

	// Try to parse as JSON first
	if err := json.Unmarshal(data, &resources); err == nil {
		return resources, nil
	}

	// Try to parse as YAML
	if err := yaml.Unmarshal(data, &resources); err == nil {
		return resources, nil
	}

	// Try to parse as a ResourceMatchesRequest (with resources field)
	var request struct {
		Resources []capi.ResourceMatch `json:"resources" yaml:"resources"`
	}

	if err := json.Unmarshal(data, &request); err == nil && len(request.Resources) > 0 {
		return request.Resources, nil
	}

	if err := yaml.Unmarshal(data, &request); err == nil && len(request.Resources) > 0 {
		return request.Resources, nil
	}

	return nil, fmt.Errorf("failed to parse file as JSON or YAML resource list")
}
