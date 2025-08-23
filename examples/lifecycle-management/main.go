package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fivetwenty-io/capi/v3/pkg/capi"
	"github.com/fivetwenty-io/capi/v3/pkg/cfclient"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <cf-api-endpoint>")
		fmt.Println("Example: go run main.go https://api.cf.example.com")
		os.Exit(1)
	}

	endpoint := os.Args[1]

	client, err := cfclient.NewWithPassword(
		endpoint,
		os.Getenv("CF_USERNAME"),
		os.Getenv("CF_PASSWORD"),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	fmt.Println("=== Cloud Foundry Application Lifecycle Management Examples ===")
	fmt.Println()

	// Revisions
	fmt.Println("1. Application Revisions")
	demonstrateRevisions(ctx, client)

	// Sidecars
	fmt.Println("\n2. Application Sidecars")
	demonstrateSidecars(ctx, client)

	// Resource Matches
	fmt.Println("\n3. Resource Matches")
	demonstrateResourceMatches(ctx, client)
}

func demonstrateRevisions(ctx context.Context, client capi.Client) {
	// First, find an application to work with
	fmt.Println("   Finding an application...")
	apps, err := client.Apps().List(ctx, capi.NewQueryParams().WithPerPage(1))
	if err != nil {
		log.Printf("   Failed to list applications: %v", err)
		return
	}

	if len(apps.Resources) == 0 {
		fmt.Println("   No applications found. Skipping revision demo.")
		return
	}

	app := apps.Resources[0]
	fmt.Printf("   Using application: %s (%s)\n", app.Name, app.GUID)

	// List revisions for the application
	fmt.Println("\n   Listing revisions for application...")
	params := capi.NewQueryParams()
	params.WithFilter("app_guids", app.GUID)
	params.WithPerPage(5)

	// Note: We would need to implement a revisions list method for apps
	// For now, we'll demonstrate individual revision operations

	// If we had a revision GUID, we could demonstrate:
	fmt.Println("\n   Revision operations example (requires revision GUID):")
	fmt.Println("   // Get revision details")
	fmt.Println("   // revision, err := client.Revisions().Get(ctx, \"revision-guid\")")
	fmt.Printf("   // fmt.Printf(\"Version: %%d, Deployable: %%t\", revision.Version, revision.Deployable)\n")

	fmt.Println("   // Get revision environment variables")
	fmt.Println("   // envVars, err := client.Revisions().GetEnvironmentVariables(ctx, \"revision-guid\")")
	fmt.Println("   // for key, value := range envVars {")
	fmt.Printf("   //     fmt.Printf(\"%%s=%%v\", key, value)\n")
	fmt.Println("   // }")

	fmt.Println("   // Update revision metadata")
	fmt.Println("   // updateReq := &capi.RevisionUpdateRequest{")
	fmt.Println("   //     Metadata: &capi.Metadata{")
	fmt.Println("   //         Labels: map[string]string{")
	fmt.Println("   //             \"version\": \"1.2.0\",")
	fmt.Println("   //             \"team\":    \"backend\",")
	fmt.Println("   //         },")
	fmt.Println("   //     },")
	fmt.Println("   // }")
	fmt.Println("   // revision, err := client.Revisions().Update(ctx, \"revision-guid\", updateReq)")
}

func demonstrateSidecars(ctx context.Context, client capi.Client) {
	// First, find a process to work with
	fmt.Println("   Finding an application process...")
	apps, err := client.Apps().List(ctx, capi.NewQueryParams().WithPerPage(1))
	if err != nil {
		log.Printf("   Failed to list applications: %v", err)
		return
	}

	if len(apps.Resources) == 0 {
		fmt.Println("   No applications found. Skipping sidecar demo.")
		return
	}

	app := apps.Resources[0]
	fmt.Printf("   Using application: %s (%s)\n", app.Name, app.GUID)

	// Get processes for the application
	processes, err := client.Processes().List(ctx, capi.NewQueryParams().WithFilter("app_guids", app.GUID))
	if err != nil {
		log.Printf("   Failed to list processes: %v", err)
		return
	}

	if len(processes.Resources) == 0 {
		fmt.Println("   No processes found for application. Skipping sidecar demo.")
		return
	}

	process := processes.Resources[0]
	fmt.Printf("   Using process: %s (%s)\n", process.Type, process.GUID)

	// List sidecars for the process
	fmt.Println("\n   Listing sidecars for process...")
	sidecars, err := client.Sidecars().ListForProcess(ctx, process.GUID, capi.NewQueryParams().WithPerPage(10))
	if err != nil {
		log.Printf("   Failed to list sidecars: %v", err)
		return
	}

	fmt.Printf("   Found %d sidecars for process\n", len(sidecars.Resources))
	for _, sidecar := range sidecars.Resources {
		memoryStr := "default"
		if sidecar.MemoryInMB != nil {
			memoryStr = fmt.Sprintf("%d MB", *sidecar.MemoryInMB)
		}

		fmt.Printf("     - %s: %s (Memory: %s, Types: %v)\n",
			sidecar.Name,
			sidecar.Command,
			memoryStr,
			sidecar.ProcessTypes)
	}

	// Demonstrate sidecar operations if sidecars exist
	if len(sidecars.Resources) > 0 {
		sidecar := sidecars.Resources[0]
		fmt.Printf("\n   Getting detailed sidecar information for: %s\n", sidecar.Name)

		detailedSidecar, err := client.Sidecars().Get(ctx, sidecar.GUID)
		if err != nil {
			log.Printf("   Failed to get sidecar: %v", err)
			return
		}

		fmt.Printf("   Sidecar Details:\n")
		fmt.Printf("     Name: %s\n", detailedSidecar.Name)
		fmt.Printf("     GUID: %s\n", detailedSidecar.GUID)
		fmt.Printf("     Command: %s\n", detailedSidecar.Command)
		fmt.Printf("     Process Types: %v\n", detailedSidecar.ProcessTypes)
		fmt.Printf("     Origin: %s\n", detailedSidecar.Origin)
		fmt.Printf("     Created: %s\n", detailedSidecar.CreatedAt.Format(time.RFC3339))

		if detailedSidecar.MemoryInMB != nil {
			fmt.Printf("     Memory: %d MB\n", *detailedSidecar.MemoryInMB)
		}
	} else {
		fmt.Println("\n   No sidecars found for demonstration.")
		fmt.Println("   Sidecar operations example:")
		fmt.Println("   // Get sidecar")
		fmt.Println("   // sidecar, err := client.Sidecars().Get(ctx, \"sidecar-guid\")")

		fmt.Println("   // Update sidecar")
		fmt.Println("   // newName := \"updated-sidecar\"")
		fmt.Println("   // newCommand := \"./updated-command\"")
		fmt.Println("   // newMemory := 256")
		fmt.Println("   // updateReq := &capi.SidecarUpdateRequest{")
		fmt.Println("   //     Name:         &newName,")
		fmt.Println("   //     Command:      &newCommand,")
		fmt.Println("   //     ProcessTypes: []string{\"web\", \"worker\"},")
		fmt.Println("   //     MemoryInMB:   &newMemory,")
		fmt.Println("   // }")
		fmt.Println("   // sidecar, err := client.Sidecars().Update(ctx, \"sidecar-guid\", updateReq)")
	}
}

func demonstrateResourceMatches(ctx context.Context, client capi.Client) {
	fmt.Println("   Demonstrating resource matches...")

	// Create example resource list
	resources := []capi.ResourceMatch{
		{
			Path: "app.js",
			SHA1: "da39a3ee5e6b4b0d3255bfef95601890afd80709", // Empty file SHA1
			Size: 1024,
			Mode: "0644",
		},
		{
			Path: "package.json",
			SHA1: "356a192b7913b04c54574d18c28d46e6395428ab", // "1" SHA1
			Size: 512,
			Mode: "0644",
		},
		{
			Path: "server.js",
			SHA1: "da4b9237bacccdf19c0760cab7aec4a8359010b0", // "hello" SHA1
			Size: 2048,
			Mode: "0644",
		},
	}

	createReq := &capi.ResourceMatchesRequest{
		Resources: resources,
	}

	fmt.Printf("   Checking %d resources for matches...\n", len(resources))
	for _, resource := range resources {
		fmt.Printf("     - %s (SHA1: %s, Size: %d bytes)\n",
			resource.Path, resource.SHA1, resource.Size)
	}

	// Create resource matches request
	matches, err := client.ResourceMatches().Create(ctx, createReq)
	if err != nil {
		log.Printf("   Failed to create resource matches: %v", err)
		return
	}

	// Calculate which resources need to be uploaded
	matchedCount := len(matches.Resources)
	totalCount := len(resources)
	uploadCount := totalCount - matchedCount

	fmt.Printf("\n   Resource Match Results:\n")
	fmt.Printf("     Total resources: %d\n", totalCount)
	fmt.Printf("     Matched resources: %d\n", matchedCount)
	fmt.Printf("     Resources to upload: %d\n", uploadCount)

	if matchedCount > 0 {
		fmt.Printf("   Matched resources:\n")
		for _, match := range matches.Resources {
			fmt.Printf("     - %s (already exists on platform)\n", match.Path)
		}
	}

	// Show which resources would need to be uploaded
	if uploadCount > 0 {
		fmt.Printf("   Resources that need to be uploaded:\n")
		for _, resource := range resources {
			found := false
			for _, match := range matches.Resources {
				if resource.SHA1 == match.SHA1 {
					found = true
					break
				}
			}
			if !found {
				fmt.Printf("     - %s (new file)\n", resource.Path)
			}
		}
	}

	fmt.Printf("\n   This optimization could save %.1f%% of upload bandwidth!\n",
		float64(matchedCount)/float64(totalCount)*100)
}
