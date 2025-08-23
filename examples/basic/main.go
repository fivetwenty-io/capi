package main

import (
	"context"
	"fmt"
	"log"

	"github.com/fivetwenty-io/capi/pkg/capi"
	"github.com/fivetwenty-io/capi/pkg/cfclient"
)

func main() {
	// Create a client with username/password authentication
	client, err := cfclient.NewWithPassword(
		"https://api.your-cf-domain.com",
		"your-username",
		"your-password",
	)
	if err != nil {
		log.Fatalf("Failed to create CF client: %v", err)
	}

	ctx := context.Background()

	// Example 1: Get API info
	fmt.Println("=== API Info ===")
	info, err := client.GetInfo(ctx)
	if err != nil {
		log.Fatalf("Failed to get API info: %v", err)
	}
	fmt.Printf("API Version: %d\n", info.Version)
	fmt.Printf("API Description: %s\n", info.Description)
	fmt.Println()

	// Example 2: List organizations
	fmt.Println("=== Organizations ===")
	orgs, err := client.Organizations().List(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to list organizations: %v", err)
	}

	fmt.Printf("Found %d organizations:\n", len(orgs.Resources))
	for _, org := range orgs.Resources {
		fmt.Printf("  - %s (GUID: %s)\n", org.Name, org.GUID)

		// Show organization metadata if present
		if org.Metadata != nil && len(org.Metadata.Labels) > 0 {
			fmt.Println("    Labels:")
			for key, value := range org.Metadata.Labels {
				fmt.Printf("      %s: %s\n", key, value)
			}
		}
	}
	fmt.Println()

	// Example 3: List spaces (if we have organizations)
	if len(orgs.Resources) > 0 {
		fmt.Println("=== Spaces ===")
		firstOrgGUID := orgs.Resources[0].GUID

		// Filter spaces by organization
		params := capi.NewQueryParams()
		params.WithFilter("organization_guids", firstOrgGUID)

		spaces, err := client.Spaces().List(ctx, params)
		if err != nil {
			log.Fatalf("Failed to list spaces: %v", err)
		}

		fmt.Printf("Found %d spaces in organization '%s':\n",
			len(spaces.Resources), orgs.Resources[0].Name)
		for _, space := range spaces.Resources {
			fmt.Printf("  - %s (GUID: %s)\n", space.Name, space.GUID)
		}
		fmt.Println()
	}

	// Example 4: List applications with pagination
	fmt.Println("=== Applications (with pagination) ===")
	params := capi.NewQueryParams()
	params.WithPerPage(5) // Small page size for demonstration

	// Use List for pagination (simplified for example)
	appList, err := client.Apps().List(ctx, params)
	if err != nil {
		log.Fatalf("Failed to list applications: %v", err)
	}

	allApps := appList.Resources

	fmt.Printf("Total applications found: %d (first page only)\n", len(allApps))
	for _, app := range allApps {
		fmt.Printf("  - %s (State: %s)\n", app.Name, app.State)
	}
}
