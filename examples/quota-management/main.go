package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/fivetwenty-io/capi/pkg/capi"
	"github.com/fivetwenty-io/capi/pkg/cfclient"
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

	fmt.Println("=== Cloud Foundry Quota Management Examples ===")
	fmt.Println()

	// Organization Quota Management
	fmt.Println("1. Organization Quota Management")
	demonstrateOrgQuotaManagement(ctx, client)

	// Space Quota Management
	fmt.Println("\n2. Space Quota Management")
	demonstrateSpaceQuotaManagement(ctx, client)
}

func demonstrateOrgQuotaManagement(ctx context.Context, client capi.Client) {
	// List existing organization quotas
	fmt.Println("   Listing existing organization quotas...")
	quotas, err := client.OrganizationQuotas().List(ctx, nil)
	if err != nil {
		log.Printf("   Failed to list organization quotas: %v", err)
		return
	}

	fmt.Printf("   Found %d organization quotas\n", len(quotas.Resources))
	for _, quota := range quotas.Resources {
		memoryStr := "unlimited"
		if quota.Apps != nil && quota.Apps.TotalMemoryInMB != nil {
			memoryStr = fmt.Sprintf("%d MB", *quota.Apps.TotalMemoryInMB)
		}

		servicesStr := "unlimited"
		if quota.Services != nil && quota.Services.TotalServiceInstances != nil {
			servicesStr = fmt.Sprintf("%d", *quota.Services.TotalServiceInstances)
		}

		fmt.Printf("     - %s: Memory=%s, Services=%s\n", quota.Name, memoryStr, servicesStr)
	}

	// Create a new organization quota
	fmt.Println("\n   Creating new organization quota...")
	totalMemory := 2048
	instanceMemory := 512
	totalInstances := 10
	totalAppTasks := 5
	totalServices := 10
	totalServiceKeys := 20
	totalRoutes := 50
	totalReservedPorts := 5
	totalDomains := 5
	paidServices := true

	createReq := &capi.OrganizationQuotaCreateRequest{
		Name: "demo-org-quota",
		Apps: &capi.OrganizationQuotaApps{
			TotalMemoryInMB:         &totalMemory,
			TotalInstanceMemoryInMB: &instanceMemory,
			TotalInstances:          &totalInstances,
			TotalAppTasks:           &totalAppTasks,
		},
		Services: &capi.OrganizationQuotaServices{
			PaidServicesAllowed:   &paidServices,
			TotalServiceInstances: &totalServices,
			TotalServiceKeys:      &totalServiceKeys,
		},
		Routes: &capi.OrganizationQuotaRoutes{
			TotalRoutes:        &totalRoutes,
			TotalReservedPorts: &totalReservedPorts,
		},
		Domains: &capi.OrganizationQuotaDomains{
			TotalDomains: &totalDomains,
		},
	}

	quota, err := client.OrganizationQuotas().Create(ctx, createReq)
	if err != nil {
		log.Printf("   Failed to create organization quota: %v", err)
		return
	}

	fmt.Printf("   Created organization quota: %s (GUID: %s)\n", quota.Name, quota.GUID)

	// Get the quota details
	fmt.Println("\n   Getting quota details...")
	quota, err = client.OrganizationQuotas().Get(ctx, quota.GUID)
	if err != nil {
		log.Printf("   Failed to get organization quota: %v", err)
		return
	}

	fmt.Printf("   Quota Details:\n")
	fmt.Printf("     Name: %s\n", quota.Name)
	fmt.Printf("     GUID: %s\n", quota.GUID)
	if quota.Apps != nil {
		if quota.Apps.TotalMemoryInMB != nil {
			fmt.Printf("     Total Memory: %d MB\n", *quota.Apps.TotalMemoryInMB)
		}
		if quota.Apps.TotalInstances != nil {
			fmt.Printf("     Total Instances: %d\n", *quota.Apps.TotalInstances)
		}
	}

	// Update the quota
	fmt.Println("\n   Updating quota...")
	newMemory := 4096
	newName := "demo-org-quota-updated"
	updateReq := &capi.OrganizationQuotaUpdateRequest{
		Name: &newName,
		Apps: &capi.OrganizationQuotaApps{
			TotalMemoryInMB: &newMemory,
		},
	}

	updatedQuota, err := client.OrganizationQuotas().Update(ctx, quota.GUID, updateReq)
	if err != nil {
		log.Printf("   Failed to update organization quota: %v", err)
		return
	}

	fmt.Printf("   Updated quota: %s (Memory: %d MB)\n", updatedQuota.Name, *updatedQuota.Apps.TotalMemoryInMB)

	// Clean up - delete the demo quota
	fmt.Println("\n   Cleaning up - deleting demo quota...")
	err = client.OrganizationQuotas().Delete(ctx, updatedQuota.GUID)
	if err != nil {
		log.Printf("   Failed to delete organization quota: %v", err)
		return
	}

	fmt.Printf("   Deleted quota: %s\n", updatedQuota.Name)
}

func demonstrateSpaceQuotaManagement(ctx context.Context, client capi.Client) {
	// First, we need to find an organization to create the space quota in
	fmt.Println("   Finding an organization...")
	orgs, err := client.Organizations().List(ctx, capi.NewQueryParams().WithPerPage(1))
	if err != nil {
		log.Printf("   Failed to list organizations: %v", err)
		return
	}

	if len(orgs.Resources) == 0 {
		fmt.Println("   No organizations found. Skipping space quota demo.")
		return
	}

	orgGUID := orgs.Resources[0].GUID
	orgName := orgs.Resources[0].Name
	fmt.Printf("   Using organization: %s (%s)\n", orgName, orgGUID)

	// List existing space quotas for this organization
	fmt.Println("\n   Listing space quotas for organization...")
	params := capi.NewQueryParams()
	params.WithFilter("organization_guids", orgGUID)
	spaceQuotas, err := client.SpaceQuotas().List(ctx, params)
	if err != nil {
		log.Printf("   Failed to list space quotas: %v", err)
		return
	}

	fmt.Printf("   Found %d space quotas in organization\n", len(spaceQuotas.Resources))

	// Create a new space quota
	fmt.Println("\n   Creating new space quota...")
	totalMemory := 1024
	totalInstances := 5
	totalRoutes := 20
	logRateLimit := 1000

	createReq := &capi.SpaceQuotaV3CreateRequest{
		Name: "demo-space-quota",
		Relationships: capi.SpaceQuotaRelationships{
			Organization: capi.Relationship{
				Data: &capi.RelationshipData{GUID: orgGUID},
			},
			Spaces: capi.ToManyRelationship{
				Data: []capi.RelationshipData{},
			},
		},
		Apps: &capi.SpaceQuotaApps{
			TotalMemoryInMB:              &totalMemory,
			TotalInstances:               &totalInstances,
			LogRateLimitInBytesPerSecond: &logRateLimit,
		},
		Routes: &capi.SpaceQuotaRoutes{
			TotalRoutes: &totalRoutes,
		},
	}

	spaceQuota, err := client.SpaceQuotas().Create(ctx, createReq)
	if err != nil {
		log.Printf("   Failed to create space quota: %v", err)
		return
	}

	fmt.Printf("   Created space quota: %s (GUID: %s)\n", spaceQuota.Name, spaceQuota.GUID)

	// Get the space quota details
	fmt.Println("\n   Getting space quota details...")
	spaceQuota, err = client.SpaceQuotas().Get(ctx, spaceQuota.GUID)
	if err != nil {
		log.Printf("   Failed to get space quota: %v", err)
		return
	}

	fmt.Printf("   Space Quota Details:\n")
	fmt.Printf("     Name: %s\n", spaceQuota.Name)
	if spaceQuota.Apps != nil {
		if spaceQuota.Apps.TotalMemoryInMB != nil {
			fmt.Printf("     Total Memory: %d MB\n", *spaceQuota.Apps.TotalMemoryInMB)
		}
		if spaceQuota.Apps.TotalInstances != nil {
			fmt.Printf("     Total Instances: %d\n", *spaceQuota.Apps.TotalInstances)
		}
	}

	// Clean up - delete the demo space quota
	fmt.Println("\n   Cleaning up - deleting demo space quota...")
	err = client.SpaceQuotas().Delete(ctx, spaceQuota.GUID)
	if err != nil {
		log.Printf("   Failed to delete space quota: %v", err)
		return
	}

	fmt.Printf("   Deleted space quota: %s\n", spaceQuota.Name)
}
