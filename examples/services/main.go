package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fivetwenty-io/capi-client/pkg/capi"
	"github.com/fivetwenty-io/capi-client/pkg/cfclient"
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

	// Example 1: List Service Offerings and Plans
	fmt.Println("=== Service Offerings and Plans ===")
	listServiceOfferingsAndPlans(client, ctx)

	// Example 2: Create Managed Service Instance
	fmt.Println("\n=== Create Managed Service Instance ===")
	managedInstance := createManagedServiceInstance(client, ctx)

	// Example 3: Create User-Provided Service Instance
	fmt.Println("\n=== Create User-Provided Service Instance ===")
	userProvidedInstance := createUserProvidedServiceInstance(client, ctx)

	// Example 4: Service Instance Management
	fmt.Println("\n=== Service Instance Management ===")
	if managedInstance != nil {
		manageServiceInstance(client, ctx, managedInstance)
	}

	// Example 5: Service Bindings
	fmt.Println("\n=== Service Bindings ===")
	if managedInstance != nil {
		manageServiceBindings(client, ctx, managedInstance)
	}

	// Example 6: Service Keys
	fmt.Println("\n=== Service Keys ===")
	if userProvidedInstance != nil {
		manageServiceKeys(client, ctx, userProvidedInstance)
	}

	// Example 7: Service Usage Events
	fmt.Println("\n=== Service Usage Events ===")
	listServiceUsageEvents(client, ctx)

	// Cleanup
	fmt.Println("\n=== Cleanup ===")
	cleanup(client, ctx, managedInstance, userProvidedInstance)
}

func listServiceOfferingsAndPlans(client capi.Client, ctx context.Context) {
	// List all service offerings
	offerings, err := client.ServiceOfferings().List(ctx, nil)
	if err != nil {
		log.Printf("Failed to list service offerings: %v", err)
		return
	}

	fmt.Printf("Found %d service offerings:\n", len(offerings.Resources))
	for _, offering := range offerings.Resources {
		fmt.Printf("  📦 %s (%s)\n", offering.Name, offering.GUID)
		fmt.Printf("     Description: %s\n", offering.Description)
		fmt.Printf("     Broker GUID: %s\n", offering.Relationships.ServiceBroker.Data.GUID)

		if offering.Metadata != nil && len(offering.Metadata.Labels) > 0 {
			fmt.Println("     Labels:")
			for key, value := range offering.Metadata.Labels {
				fmt.Printf("       %s: %s\n", key, value)
			}
		}

		// List plans for this offering
		params := capi.NewQueryParams().WithFilter("service_offering_guids", offering.GUID)
		plans, err := client.ServicePlans().List(ctx, params)
		if err != nil {
			log.Printf("Failed to list plans for offering %s: %v", offering.Name, err)
			continue
		}

		fmt.Printf("     Plans (%d):\n", len(plans.Resources))
		for _, plan := range plans.Resources {
			fmt.Printf("       • %s (%s)\n", plan.Name, plan.GUID)
			fmt.Printf("         Description: %s\n", plan.Description)
			fmt.Printf("         Free: %v\n", plan.Free)
			fmt.Printf("         Available: %v\n", plan.Available)

			if plan.Costs != nil && len(plan.Costs) > 0 {
				fmt.Println("         Costs:")
				for _, cost := range plan.Costs {
					fmt.Printf("           %s: %.2f %s\n", cost.Unit, cost.Amount, cost.Currency)
				}
			}
		}
		fmt.Println()
	}
}

func createManagedServiceInstance(client capi.Client, ctx context.Context) *capi.ServiceInstance {
	// Get a space to work with (you'll need to replace with your actual space GUID)
	spaceGUID := "your-space-guid"

	// Find a service plan to use (we'll use the first available plan)
	offerings, err := client.ServiceOfferings().List(ctx, nil)
	if err != nil || len(offerings.Resources) == 0 {
		log.Printf("No service offerings available: %v", err)
		return nil
	}

	offering := offerings.Resources[0]
	params := capi.NewQueryParams().WithFilter("service_offering_guids", offering.GUID)
	plans, err := client.ServicePlans().List(ctx, params)
	if err != nil || len(plans.Resources) == 0 {
		log.Printf("No service plans available for offering %s: %v", offering.Name, err)
		return nil
	}

	plan := plans.Resources[0]

	// Create managed service instance
	createReq := &capi.ServiceInstanceCreateRequest{
		Type: "managed",
		Name: "example-managed-service",
		Relationships: capi.ServiceInstanceRelationships{
			Space: capi.Relationship{
				Data: &capi.RelationshipData{GUID: spaceGUID},
			},
			ServicePlan: &capi.Relationship{
				Data: &capi.RelationshipData{GUID: plan.GUID},
			},
		},
		Parameters: map[string]interface{}{
			"example_param": "example_value",
		},
		Metadata: &capi.Metadata{
			Labels: map[string]string{
				"team":        "platform",
				"environment": "demo",
			},
			Annotations: map[string]string{
				"created-by": "service-example",
			},
		},
		Tags: []string{"database", "managed"},
	}

	instanceInterface, err := client.ServiceInstances().Create(ctx, createReq)
	if err != nil {
		log.Printf("Failed to create managed service instance: %v", err)
		return nil
	}

	instance, ok := instanceInterface.(*capi.ServiceInstance)
	if !ok {
		log.Printf("Unexpected return type from ServiceInstances().Create()")
		return nil
	}

	fmt.Printf("Created managed service instance: %s (GUID: %s)\n", instance.Name, instance.GUID)
	fmt.Printf("  Type: %s\n", instance.Type)
	fmt.Printf("  State: %s\n", instance.LastOperation.State)
	fmt.Printf("  Service Plan: %s\n", plan.Name)

	// Wait for the service instance to be ready
	if instance.LastOperation.State == "in progress" {
		fmt.Println("Waiting for service instance to be ready...")
		instance = waitForServiceInstanceReady(client, ctx, instance.GUID)
	}

	return instance
}

func createUserProvidedServiceInstance(client capi.Client, ctx context.Context) *capi.ServiceInstance {
	// Get a space to work with
	spaceGUID := "your-space-guid"

	// Create user-provided service instance
	createReq := &capi.ServiceInstanceCreateRequest{
		Type: "user-provided",
		Name: "example-user-provided-service",
		Relationships: capi.ServiceInstanceRelationships{
			Space: capi.Relationship{
				Data: &capi.RelationshipData{GUID: spaceGUID},
			},
		},
		Credentials: map[string]interface{}{
			"uri":      "https://external-api.example.com",
			"api_key":  "secret-api-key",
			"username": "service-user",
			"password": "service-password",
		},
		SyslogDrainURL:  &[]string{"https://logs.example.com/drain"}[0],
		RouteServiceURL: &[]string{"https://route-service.example.com"}[0],
		Tags:            []string{"external", "api"},
	}

	instanceInterface, err := client.ServiceInstances().Create(ctx, createReq)
	if err != nil {
		log.Printf("Failed to create user-provided service instance: %v", err)
		return nil
	}

	instance, ok := instanceInterface.(*capi.ServiceInstance)
	if !ok {
		log.Printf("Unexpected return type from ServiceInstances().Create()")
		return nil
	}

	fmt.Printf("Created user-provided service instance: %s (GUID: %s)\n", instance.Name, instance.GUID)
	fmt.Printf("  Type: %s\n", instance.Type)
	fmt.Printf("  Syslog Drain URL: %s\n", *instance.SyslogDrainURL)
	fmt.Printf("  Route Service URL: %s\n", *instance.RouteServiceURL)

	return instance
}

func manageServiceInstance(client capi.Client, ctx context.Context, instance *capi.ServiceInstance) {
	fmt.Printf("Managing service instance: %s\n", instance.Name)

	// Get updated service instance details
	updatedInstance, err := client.ServiceInstances().Get(ctx, instance.GUID)
	if err != nil {
		log.Printf("Failed to get service instance: %v", err)
		return
	}

	fmt.Printf("Service Instance Details:\n")
	fmt.Printf("  Name: %s\n", updatedInstance.Name)
	fmt.Printf("  Type: %s\n", updatedInstance.Type)
	fmt.Printf("  State: %s\n", updatedInstance.LastOperation.State)
	fmt.Printf("  Created: %s\n", updatedInstance.CreatedAt.Format(time.RFC3339))
	fmt.Printf("  Updated: %s\n", updatedInstance.UpdatedAt.Format(time.RFC3339))

	if updatedInstance.Tags != nil {
		fmt.Printf("  Tags: %v\n", updatedInstance.Tags)
	}

	// Update service instance
	newName := "updated-service-name"
	updateReq := &capi.ServiceInstanceUpdateRequest{
		Name: &newName,
		Metadata: &capi.Metadata{
			Labels: map[string]string{
				"version": "2.0",
				"updated": "true",
			},
		},
		Tags: []string{"database", "managed", "updated"},
	}

	updatedInstanceInterface, err := client.ServiceInstances().Update(ctx, instance.GUID, updateReq)
	if err != nil {
		log.Printf("Failed to update service instance: %v", err)
		return
	}

	updatedInstance, ok := updatedInstanceInterface.(*capi.ServiceInstance)
	if !ok {
		log.Printf("Unexpected return type from ServiceInstances().Update()")
		return
	}

	fmt.Printf("Updated service instance name to: %s\n", updatedInstance.Name)

	// Get service instance parameters (if available)
	parameters, err := client.ServiceInstances().GetParameters(ctx, instance.GUID)
	if err != nil {
		log.Printf("Failed to get parameters: %v", err)
	} else {
		fmt.Printf("Service parameters available: %v\n", len(parameters.Parameters) > 0)
	}
}

func manageServiceBindings(client capi.Client, ctx context.Context, serviceInstance *capi.ServiceInstance) {
	// Get an application to bind to (you'll need to replace with actual app GUID)
	appGUID := "your-app-guid"

	fmt.Printf("Managing service bindings for: %s\n", serviceInstance.Name)

	// Create service credential binding (app binding)
	bindingName := "example-binding"
	createBindingReq := &capi.ServiceCredentialBindingCreateRequest{
		Type: "app",
		Name: &bindingName,
		Relationships: capi.ServiceCredentialBindingRelationships{
			ServiceInstance: capi.Relationship{
				Data: &capi.RelationshipData{GUID: serviceInstance.GUID},
			},
			App: &capi.Relationship{
				Data: &capi.RelationshipData{GUID: appGUID},
			},
		},
		Parameters: map[string]interface{}{
			"permission": "read-write",
			"pool_size":  10,
		},
	}

	bindingInterface, err := client.ServiceCredentialBindings().Create(ctx, createBindingReq)
	if err != nil {
		log.Printf("Failed to create service binding: %v", err)
		return
	}

	binding, ok := bindingInterface.(*capi.ServiceCredentialBinding)
	if !ok {
		log.Printf("Unexpected return type from ServiceCredentialBindings().Create()")
		return
	}

	fmt.Printf("Created service binding: %s (GUID: %s)\n", binding.Name, binding.GUID)

	// Wait for binding to be ready
	if binding.LastOperation != nil && binding.LastOperation.State == "in progress" {
		fmt.Println("Waiting for service binding to be ready...")
		binding = waitForServiceBindingReady(client, ctx, binding.GUID)
	}

	// List all bindings for the service instance
	params := capi.NewQueryParams().WithFilter("service_instance_guids", serviceInstance.GUID)
	bindings, err := client.ServiceCredentialBindings().List(ctx, params)
	if err != nil {
		log.Printf("Failed to list service bindings: %v", err)
		return
	}

	fmt.Printf("Service instance has %d bindings:\n", len(bindings.Resources))
	for _, b := range bindings.Resources {
		fmt.Printf("  - %s (Type: %s, State: %s)\n", b.Name, b.Type, b.LastOperation.State)
	}

	// Get binding details
	bindingDetails, err := client.ServiceCredentialBindings().GetDetails(ctx, binding.GUID)
	if err != nil {
		log.Printf("Failed to get binding details: %v", err)
	} else {
		fmt.Printf("Binding has %d credential keys\n", len(bindingDetails.Credentials))
	}

	// Update binding
	updateBindingReq := &capi.ServiceCredentialBindingUpdateRequest{
		Metadata: &capi.Metadata{
			Labels: map[string]string{
				"updated": "true",
			},
		},
	}

	updatedBinding, err := client.ServiceCredentialBindings().Update(ctx, binding.GUID, updateBindingReq)
	if err != nil {
		log.Printf("Failed to update binding: %v", err)
	} else {
		fmt.Printf("Updated binding name to: %s\n", updatedBinding.Name)
	}

	// Cleanup: Delete the binding
	_, err = client.ServiceCredentialBindings().Delete(ctx, binding.GUID)
	if err != nil {
		log.Printf("Failed to delete service binding: %v", err)
	} else {
		fmt.Printf("Deleted service binding: %s\n", binding.Name)
	}
}

func manageServiceKeys(client capi.Client, ctx context.Context, serviceInstance *capi.ServiceInstance) {
	fmt.Printf("Managing service keys for: %s\n", serviceInstance.Name)

	// Create service key (credential binding without app)
	keyName := "example-service-key"
	createKeyReq := &capi.ServiceCredentialBindingCreateRequest{
		Type: "key",
		Name: &keyName,
		Relationships: capi.ServiceCredentialBindingRelationships{
			ServiceInstance: capi.Relationship{
				Data: &capi.RelationshipData{GUID: serviceInstance.GUID},
			},
		},
		Parameters: map[string]interface{}{
			"permissions": []string{"read", "write"},
		},
	}

	serviceKeyInterface, err := client.ServiceCredentialBindings().Create(ctx, createKeyReq)
	if err != nil {
		log.Printf("Failed to create service key: %v", err)
		return
	}

	serviceKey, ok := serviceKeyInterface.(*capi.ServiceCredentialBinding)
	if !ok {
		log.Printf("Unexpected return type from ServiceCredentialBindings().Create()")
		return
	}

	fmt.Printf("Created service key: %s (GUID: %s)\n", serviceKey.Name, serviceKey.GUID)

	// List service keys for the instance
	params := capi.NewQueryParams()
	params.WithFilter("service_instance_guids", serviceInstance.GUID)
	params.WithFilter("type", "key")

	serviceKeys, err := client.ServiceCredentialBindings().List(ctx, params)
	if err != nil {
		log.Printf("Failed to list service keys: %v", err)
		return
	}

	fmt.Printf("Service instance has %d keys:\n", len(serviceKeys.Resources))
	for _, key := range serviceKeys.Resources {
		fmt.Printf("  - %s (GUID: %s)\n", key.Name, key.GUID)
	}

	// Get service key credentials
	keyDetails, err := client.ServiceCredentialBindings().GetDetails(ctx, serviceKey.GUID)
	if err != nil {
		log.Printf("Failed to get service key details: %v", err)
	} else {
		fmt.Printf("Service key credentials available\n")
		for credKey := range keyDetails.Credentials {
			fmt.Printf("  - %s\n", credKey)
		}
	}

	// Cleanup: Delete the service key
	_, err = client.ServiceCredentialBindings().Delete(ctx, serviceKey.GUID)
	if err != nil {
		log.Printf("Failed to delete service key: %v", err)
	} else {
		fmt.Printf("Deleted service key: %s\n", serviceKey.Name)
	}
}

func listServiceUsageEvents(client capi.Client, ctx context.Context) {
	fmt.Println("=== Service Usage Events ===")
	fmt.Println("Service usage events are not implemented in this client yet")
	// TODO: Implement when ServiceUsageEvents client is available
}

func cleanup(client capi.Client, ctx context.Context, managedInstance, userProvidedInstance *capi.ServiceInstance) {
	if managedInstance != nil {
		fmt.Printf("Deleting managed service instance: %s\n", managedInstance.Name)
		_, err := client.ServiceInstances().Delete(ctx, managedInstance.GUID)
		if err != nil {
			log.Printf("Failed to delete managed service instance: %v", err)
		} else {
			fmt.Printf("Initiated deletion of managed service instance\n")
		}
	}

	if userProvidedInstance != nil {
		fmt.Printf("Deleting user-provided service instance: %s\n", userProvidedInstance.Name)
		_, err := client.ServiceInstances().Delete(ctx, userProvidedInstance.GUID)
		if err != nil {
			log.Printf("Failed to delete user-provided service instance: %v", err)
		} else {
			fmt.Printf("Deleted user-provided service instance\n")
		}
	}
}

func waitForServiceInstanceReady(client capi.Client, ctx context.Context, instanceGUID string) *capi.ServiceInstance {
	maxAttempts := 30
	for i := 0; i < maxAttempts; i++ {
		instance, err := client.ServiceInstances().Get(ctx, instanceGUID)
		if err != nil {
			log.Printf("Error checking service instance status: %v", err)
			return nil
		}

		if instance.LastOperation.State == "succeeded" {
			fmt.Printf("Service instance is ready!\n")
			return instance
		} else if instance.LastOperation.State == "failed" {
			fmt.Printf("Service instance creation failed: %s\n", instance.LastOperation.Description)
			return instance
		}

		fmt.Printf("Service instance status: %s (%s)\n",
			instance.LastOperation.State, instance.LastOperation.Description)
		time.Sleep(10 * time.Second)
	}

	fmt.Printf("Timeout waiting for service instance to be ready\n")
	return nil
}

func waitForServiceBindingReady(client capi.Client, ctx context.Context, bindingGUID string) *capi.ServiceCredentialBinding {
	maxAttempts := 20
	for i := 0; i < maxAttempts; i++ {
		binding, err := client.ServiceCredentialBindings().Get(ctx, bindingGUID)
		if err != nil {
			log.Printf("Error checking service binding status: %v", err)
			return nil
		}

		if binding.LastOperation.State == "succeeded" {
			fmt.Printf("Service binding is ready!\n")
			return binding
		} else if binding.LastOperation.State == "failed" {
			description := ""
			if binding.LastOperation.Description != nil {
				description = *binding.LastOperation.Description
			}
			fmt.Printf("Service binding creation failed: %s\n", description)
			return binding
		}

		fmt.Printf("Service binding status: %s\n", binding.LastOperation.State)
		time.Sleep(5 * time.Second)
	}

	fmt.Printf("Timeout waiting for service binding to be ready\n")
	return nil
}
