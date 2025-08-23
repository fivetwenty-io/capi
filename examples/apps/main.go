package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fivetwenty-io/capi/pkg/capi"
	"github.com/fivetwenty-io/capi/pkg/cfclient"
)

func main() {
	// Create authenticated client
	client, err := cfclient.NewWithPassword(
		"https://api.your-cf-domain.com",
		"your-username",
		"your-password",
	)
	if err != nil {
		log.Fatalf("Failed to create CF client: %v", err)
	}

	ctx := context.Background()

	// Get a space to work with (you'll need to replace this with your actual space GUID)
	spaceGUID := "your-space-guid"

	// Example 1: Create an Application
	fmt.Println("=== Creating Application ===")
	createAppReq := &capi.AppCreateRequest{
		Name: "example-app",
		Relationships: capi.AppRelationships{
			Space: capi.Relationship{
				Data: &capi.RelationshipData{GUID: spaceGUID},
			},
		},
		Metadata: &capi.Metadata{
			Labels: map[string]string{
				"team":        "platform",
				"environment": "demo",
			},
			Annotations: map[string]string{
				"created-by": "capi-client-example",
			},
		},
	}

	app, err := client.Apps().Create(ctx, createAppReq)
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}
	fmt.Printf("Created application: %s (GUID: %s)\n", app.Name, app.GUID)
	fmt.Println()

	// Example 2: Get Application Details
	fmt.Println("=== Getting Application Details ===")
	app, err = client.Apps().Get(ctx, app.GUID)
	if err != nil {
		log.Fatalf("Failed to get application: %v", err)
	}

	fmt.Printf("Application Details:\n")
	fmt.Printf("  Name: %s\n", app.Name)
	fmt.Printf("  GUID: %s\n", app.GUID)
	fmt.Printf("  State: %s\n", app.State)
	fmt.Printf("  Created: %s\n", app.CreatedAt.Format(time.RFC3339))
	fmt.Printf("  Updated: %s\n", app.UpdatedAt.Format(time.RFC3339))

	if app.Metadata != nil {
		if len(app.Metadata.Labels) > 0 {
			fmt.Println("  Labels:")
			for key, value := range app.Metadata.Labels {
				fmt.Printf("    %s: %s\n", key, value)
			}
		}
		if len(app.Metadata.Annotations) > 0 {
			fmt.Println("  Annotations:")
			for key, value := range app.Metadata.Annotations {
				fmt.Printf("    %s: %s\n", key, value)
			}
		}
	}
	fmt.Println()

	// Example 3: Update Application
	fmt.Println("=== Updating Application ===")
	newName := "example-app-updated"
	updateAppReq := &capi.AppUpdateRequest{
		Name: &newName,
		Metadata: &capi.Metadata{
			Labels: map[string]string{
				"version": "1.0.0",
				"updated": "true",
			},
			Annotations: map[string]string{
				"updated-by": "capi-client-example",
			},
		},
	}

	app, err = client.Apps().Update(ctx, app.GUID, updateAppReq)
	if err != nil {
		log.Fatalf("Failed to update application: %v", err)
	}
	fmt.Printf("Updated application name to: %s\n", app.Name)
	fmt.Println()

	// Example 4: List Application Processes
	fmt.Println("=== Listing Application Processes ===")
	processParams := capi.NewQueryParams().WithFilter("app_guids", app.GUID)
	processes, err := client.Processes().List(ctx, processParams)
	if err != nil {
		log.Fatalf("Failed to list processes: %v", err)
	}

	fmt.Printf("Found %d processes:\n", len(processes.Resources))
	for _, process := range processes.Resources {
		fmt.Printf("  - Type: %s, Instances: %d, Memory: %d MB, Disk: %d MB\n",
			process.Type, process.Instances, process.MemoryInMB, process.DiskInMB)
	}
	fmt.Println()

	// Example 5: Scale Application
	if len(processes.Resources) > 0 {
		fmt.Println("=== Scaling Application ===")
		webProcess := processes.Resources[0] // Usually the 'web' process

		instances := 2
		memory := 512
		disk := 1024
		scaleReq := &capi.ProcessScaleRequest{
			Instances:  &instances,
			MemoryInMB: &memory,
			DiskInMB:   &disk,
		}

		scaledProcess, err := client.Processes().Scale(ctx, webProcess.GUID, scaleReq)
		if err != nil {
			log.Fatalf("Failed to scale process: %v", err)
		}

		fmt.Printf("Scaled %s process:\n", scaledProcess.Type)
		fmt.Printf("  Instances: %d\n", scaledProcess.Instances)
		fmt.Printf("  Memory: %d MB\n", scaledProcess.MemoryInMB)
		fmt.Printf("  Disk: %d MB\n", scaledProcess.DiskInMB)
		fmt.Println()
	}

	// Example 6: List Application Environment Variables
	fmt.Println("=== Application Environment Variables ===")
	envVars, err := client.Apps().GetEnv(ctx, app.GUID)
	if err != nil {
		log.Fatalf("Failed to get environment variables: %v", err)
	}

	if len(envVars.SystemEnvJSON) > 0 {
		fmt.Println("System Environment Variables:")
		for key, value := range envVars.SystemEnvJSON {
			fmt.Printf("  %s: %v\n", key, value)
		}
	}

	if len(envVars.ApplicationEnvJSON) > 0 {
		fmt.Println("Application Environment Variables:")
		for key, value := range envVars.ApplicationEnvJSON {
			fmt.Printf("  %s: %v\n", key, value)
		}
	}
	fmt.Println()

	// Example 7: Set Environment Variables
	fmt.Println("=== Setting Environment Variables ===")
	newEnvVars := map[string]interface{}{
		"EXAMPLE_VAR": "example-value",
		"DEBUG":       "true",
	}
	_, err = client.Apps().UpdateEnvVars(ctx, app.GUID, newEnvVars)
	if err != nil {
		log.Fatalf("Failed to set environment variables: %v", err)
	}
	fmt.Println("Environment variables updated successfully")
	fmt.Println()

	// Example 8: Application Stats
	fmt.Println("=== Application Stats ===")
	// Get stats for the first process
	var stats *capi.ProcessStats
	if len(processes.Resources) > 0 {
		stats, err = client.Processes().GetStats(ctx, processes.Resources[0].GUID)
	} else {
		err = fmt.Errorf("no processes found")
	}
	if err != nil {
		log.Printf("Failed to get stats (app may not be running): %v", err)
	} else {
		fmt.Printf("Application Stats:\n")
		for _, stat := range stats.Resources {
			fmt.Printf("  Instance %d:\n", stat.Index)
			fmt.Printf("    State: %s\n", stat.State)
			if stat.Usage != nil {
				fmt.Printf("    CPU: %.2f%%\n", stat.Usage.CPU*100)
				fmt.Printf("    Memory: %d bytes\n", stat.Usage.Mem)
				fmt.Printf("    Disk: %d bytes\n", stat.Usage.Disk)
			}
		}
	}
	fmt.Println()

	// Example 9: Start/Stop Application
	fmt.Println("=== Starting Application ===")
	app, err = client.Apps().Start(ctx, app.GUID)
	if err != nil {
		log.Printf("Failed to start application: %v", err)
	} else {
		fmt.Printf("Application state changed to: %s\n", app.State)
	}

	// Wait a moment
	time.Sleep(2 * time.Second)

	fmt.Println("=== Stopping Application ===")
	app, err = client.Apps().Stop(ctx, app.GUID)
	if err != nil {
		log.Printf("Failed to stop application: %v", err)
	} else {
		fmt.Printf("Application state changed to: %s\n", app.State)
	}
	fmt.Println()

	// Example 10: Delete Application (cleanup)
	fmt.Println("=== Deleting Application ===")
	err = client.Apps().Delete(ctx, app.GUID)
	if err != nil {
		log.Fatalf("Failed to delete application: %v", err)
	}

	fmt.Println("Application deleted successfully!")
}
