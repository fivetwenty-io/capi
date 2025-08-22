package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fivetwenty-io/capi-client/pkg/capi"
	"github.com/fivetwenty-io/capi-client/pkg/cfclient"
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

	fmt.Println("=== Cloud Foundry Usage Monitoring Examples ===")
	fmt.Println()

	// Application Usage Events
	fmt.Println("1. Application Usage Events")
	demonstrateAppUsageEvents(ctx, client)

	// Service Usage Events
	fmt.Println("\n2. Service Usage Events")
	demonstrateServiceUsageEvents(ctx, client)

	// Audit Events
	fmt.Println("\n3. Audit Events")
	demonstrateAuditEvents(ctx, client)

	// Environment Variable Groups
	fmt.Println("\n4. Environment Variable Groups")
	demonstrateEnvironmentVariableGroups(ctx, client)
}

func demonstrateAppUsageEvents(ctx context.Context, client capi.Client) {
	// List recent app usage events
	fmt.Println("   Listing recent application usage events...")
	params := capi.NewQueryParams()
	params.WithPerPage(10)

	events, err := client.AppUsageEvents().List(ctx, params)
	if err != nil {
		log.Printf("   Failed to list app usage events: %v", err)
		return
	}

	fmt.Printf("   Found %d app usage events\n", len(events.Resources))
	for i, event := range events.Resources {
		if i >= 3 { // Show only first 3 for demo
			break
		}

		previousState := "N/A"
		if event.PreviousState != nil {
			previousState = *event.PreviousState
		}

		fmt.Printf("     - App: %s, State: %s -> %s, Instances: %d, Memory: %d MB\n",
			event.AppName,
			previousState,
			event.State,
			event.InstanceCount,
			event.MemoryInMBPerInstance)
	}

	// Get detailed information for the first event
	if len(events.Resources) > 0 {
		fmt.Println("\n   Getting detailed event information...")
		event, err := client.AppUsageEvents().Get(ctx, events.Resources[0].GUID)
		if err != nil {
			log.Printf("   Failed to get app usage event: %v", err)
			return
		}

		fmt.Printf("   Event Details:\n")
		fmt.Printf("     GUID: %s\n", event.GUID)
		fmt.Printf("     App: %s (%s)\n", event.AppName, event.AppGUID)
		fmt.Printf("     Space: %s (%s)\n", event.SpaceName, event.SpaceGUID)
		fmt.Printf("     Organization: %s (%s)\n", event.OrganizationName, event.OrganizationGUID)
		fmt.Printf("     Process Type: %s\n", event.ProcessType)
		fmt.Printf("     Created: %s\n", event.CreatedAt.Format(time.RFC3339))

		if event.BuildpackName != nil {
			fmt.Printf("     Buildpack: %s\n", *event.BuildpackName)
		}
		if event.TaskName != nil {
			fmt.Printf("     Task: %s\n", *event.TaskName)
		}
	}

	// Demonstrate filtering by time range
	fmt.Println("\n   Filtering events by time range (last 24 hours)...")
	yesterday := time.Now().Add(-24 * time.Hour)
	timeParams := capi.NewQueryParams()
	timeParams.WithFilter("created_ats[gte]", yesterday.Format(time.RFC3339))
	timeParams.WithPerPage(5)

	recentEvents, err := client.AppUsageEvents().List(ctx, timeParams)
	if err != nil {
		log.Printf("   Failed to filter app usage events: %v", err)
		return
	}

	fmt.Printf("   Found %d events in the last 24 hours\n", len(recentEvents.Resources))
}

func demonstrateServiceUsageEvents(ctx context.Context, client capi.Client) {
	// List recent service usage events
	fmt.Println("   Listing recent service usage events...")
	params := capi.NewQueryParams()
	params.WithPerPage(10)

	events, err := client.ServiceUsageEvents().List(ctx, params)
	if err != nil {
		log.Printf("   Failed to list service usage events: %v", err)
		return
	}

	fmt.Printf("   Found %d service usage events\n", len(events.Resources))
	for i, event := range events.Resources {
		if i >= 3 { // Show only first 3 for demo
			break
		}

		fmt.Printf("     - Service: %s (%s), State: %s, Plan: %s\n",
			event.ServiceInstanceName,
			event.ServiceInstanceType,
			event.State,
			event.ServicePlanName)
	}

	// Get detailed information for the first event
	if len(events.Resources) > 0 {
		fmt.Println("\n   Getting detailed service event information...")
		event, err := client.ServiceUsageEvents().Get(ctx, events.Resources[0].GUID)
		if err != nil {
			log.Printf("   Failed to get service usage event: %v", err)
			return
		}

		fmt.Printf("   Service Event Details:\n")
		fmt.Printf("     GUID: %s\n", event.GUID)
		fmt.Printf("     Service Instance: %s (%s)\n", event.ServiceInstanceName, event.ServiceInstanceGUID)
		fmt.Printf("     Service Type: %s\n", event.ServiceInstanceType)
		fmt.Printf("     Service Offering: %s\n", event.ServiceOfferingName)
		fmt.Printf("     Service Plan: %s\n", event.ServicePlanName)
		fmt.Printf("     Service Broker: %s\n", event.ServiceBrokerName)
		fmt.Printf("     Space: %s (%s)\n", event.SpaceName, event.SpaceGUID)
		fmt.Printf("     Organization: %s (%s)\n", event.OrganizationName, event.OrganizationGUID)
	}
}

func demonstrateAuditEvents(ctx context.Context, client capi.Client) {
	// List recent audit events
	fmt.Println("   Listing recent audit events...")
	params := capi.NewQueryParams()
	params.WithPerPage(10)

	events, err := client.AuditEvents().List(ctx, params)
	if err != nil {
		log.Printf("   Failed to list audit events: %v", err)
		return
	}

	fmt.Printf("   Found %d audit events\n", len(events.Resources))
	for i, event := range events.Resources {
		if i >= 5 { // Show only first 5 for demo
			break
		}

		fmt.Printf("     - Type: %s, Actor: %s, Target: %s (%s)\n",
			event.Type,
			event.Actor.Name,
			event.Target.Name,
			event.Target.Type)
	}

	// Get detailed information for the first event
	if len(events.Resources) > 0 {
		fmt.Println("\n   Getting detailed audit event information...")
		event, err := client.AuditEvents().Get(ctx, events.Resources[0].GUID)
		if err != nil {
			log.Printf("   Failed to get audit event: %v", err)
			return
		}

		fmt.Printf("   Audit Event Details:\n")
		fmt.Printf("     GUID: %s\n", event.GUID)
		fmt.Printf("     Type: %s\n", event.Type)
		fmt.Printf("     Actor: %s (%s) - %s\n", event.Actor.Name, event.Actor.Type, event.Actor.GUID)
		fmt.Printf("     Target: %s (%s) - %s\n", event.Target.Name, event.Target.Type, event.Target.GUID)
		fmt.Printf("     Created: %s\n", event.CreatedAt.Format(time.RFC3339))

		if event.Space != nil {
			fmt.Printf("     Space: %s (%s)\n", event.Space.Name, event.Space.GUID)
		}
		if event.Organization != nil {
			fmt.Printf("     Organization: %s (%s)\n", event.Organization.Name, event.Organization.GUID)
		}

		// Show event data if available
		if len(event.Data) > 0 {
			fmt.Printf("     Event Data:\n")
			for key, value := range event.Data {
				fmt.Printf("       %s: %v\n", key, value)
			}
		}
	}

	// Demonstrate filtering by event type
	fmt.Println("\n   Filtering audit events by type (app events)...")
	appEventParams := capi.NewQueryParams()
	appEventParams.WithFilter("types", "audit.app.create,audit.app.update,audit.app.delete")
	appEventParams.WithPerPage(5)

	appEvents, err := client.AuditEvents().List(ctx, appEventParams)
	if err != nil {
		log.Printf("   Failed to filter audit events: %v", err)
		return
	}

	fmt.Printf("   Found %d app-related audit events\n", len(appEvents.Resources))
	for _, event := range appEvents.Resources {
		fmt.Printf("     - %s: %s\n", event.Type, event.Target.Name)
	}
}

func demonstrateEnvironmentVariableGroups(ctx context.Context, client capi.Client) {
	// Get running environment variables
	fmt.Println("   Getting running environment variables...")
	runningEnvVars, err := client.EnvironmentVariableGroups().Get(ctx, "running")
	if err != nil {
		log.Printf("   Failed to get running environment variables: %v", err)
		return
	}

	fmt.Printf("   Running Environment Variables (%d variables):\n", len(runningEnvVars.Var))
	for key, value := range runningEnvVars.Var {
		fmt.Printf("     %s=%v\n", key, value)
	}

	// Get staging environment variables
	fmt.Println("\n   Getting staging environment variables...")
	stagingEnvVars, err := client.EnvironmentVariableGroups().Get(ctx, "staging")
	if err != nil {
		log.Printf("   Failed to get staging environment variables: %v", err)
		return
	}

	fmt.Printf("   Staging Environment Variables (%d variables):\n", len(stagingEnvVars.Var))
	for key, value := range stagingEnvVars.Var {
		fmt.Printf("     %s=%v\n", key, value)
	}

	// Demonstrate updating environment variables (commented out to avoid affecting real environment)
	fmt.Println("\n   Environment variable update example (commented out for safety):")
	fmt.Println("   // Update running environment variables")
	fmt.Println("   // newVars := map[string]interface{}{")
	fmt.Println("   //     \"DEMO_LOG_LEVEL\": \"debug\",")
	fmt.Println("   //     \"DEMO_FEATURE_FLAG\": true,")
	fmt.Println("   // }")
	fmt.Println("   // runningEnvVars, err := client.EnvironmentVariableGroups().Update(ctx, \"running\", newVars)")

	/*
	// Uncomment this section if you want to actually update environment variables
	// WARNING: This will affect all applications in the CF deployment
	
	newVars := map[string]interface{}{
		"DEMO_LOG_LEVEL":    "debug",
		"DEMO_FEATURE_FLAG": true,
		"DEMO_TIMESTAMP":    time.Now().Format(time.RFC3339),
	}

	fmt.Println("   Updating running environment variables...")
	updatedRunningEnvVars, err := client.EnvironmentVariableGroups().Update(ctx, "running", newVars)
	if err != nil {
		log.Printf("   Failed to update running environment variables: %v", err)
		return
	}

	fmt.Printf("   Updated running environment variables (%d variables)\n", len(updatedRunningEnvVars.Var))
	
	// Restore original environment variables
	fmt.Println("   Restoring original environment variables...")
	_, err = client.EnvironmentVariableGroups().Update(ctx, "running", runningEnvVars.Var)
	if err != nil {
		log.Printf("   Failed to restore running environment variables: %v", err)
		return
	}
	
	fmt.Println("   Restored original environment variables")
	*/
}