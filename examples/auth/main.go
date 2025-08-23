package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fivetwenty-io/capi/v3/pkg/capi"
	"github.com/fivetwenty-io/capi/v3/pkg/cfclient"
)

func main() {
	ctx := context.Background()

	// Example 1: Username/Password Authentication
	fmt.Println("=== Username/Password Authentication ===")
	userPassClient, err := cfclient.NewWithPassword(
		"https://api.your-cf-domain.com",
		"your-username",
		"your-password",
	)
	if err != nil {
		log.Fatalf("Failed to create client with username/password: %v", err)
	}

	// Test the connection
	info, err := userPassClient.GetInfo(ctx)
	if err != nil {
		log.Fatalf("Failed to authenticate with username/password: %v", err)
	}
	fmt.Printf("Successfully authenticated! API Version: %d\n", info.Version)
	fmt.Println()

	// Example 2: Client Credentials Authentication
	fmt.Println("=== Client Credentials Authentication ===")
	clientCredsClient, err := cfclient.NewWithClientCredentials(
		"https://api.your-cf-domain.com",
		"your-client-id",
		"your-client-secret",
	)
	if err != nil {
		log.Fatalf("Failed to create client with client credentials: %v", err)
	}

	info, err = clientCredsClient.GetInfo(ctx)
	if err != nil {
		log.Printf("Client credentials auth failed: %v", err)
	} else {
		fmt.Printf("Successfully authenticated with client credentials! API Version: %d\n", info.Version)
	}
	fmt.Println()

	// Example 3: Access Token Authentication
	fmt.Println("=== Access Token Authentication ===")
	// Note: You would typically get this token from another authentication flow
	accessToken := "your-access-token-here"

	tokenClient, err := cfclient.NewWithToken(
		"https://api.your-cf-domain.com",
		accessToken,
	)
	if err != nil {
		log.Fatalf("Failed to create client with access token: %v", err)
	}

	info, err = tokenClient.GetInfo(ctx)
	if err != nil {
		log.Printf("Access token auth failed (expected if token is invalid): %v", err)
	} else {
		fmt.Printf("Successfully authenticated with access token! API Version: %d\n", info.Version)
	}
	fmt.Println()

	// Example 4: Custom Configuration
	fmt.Println("=== Custom Configuration ===")
	config := &capi.Config{
		APIEndpoint:   "https://api.your-cf-domain.com",
		Username:      "your-username",
		Password:      "your-password",
		SkipTLSVerify: false, // Set to true for self-signed certificates (not recommended for production)
		HTTPTimeout:   30 * time.Second,
		UserAgent:     "my-custom-app/1.0.0",
	}

	customClient, err := cfclient.New(config)
	if err != nil {
		log.Fatalf("Failed to create client with custom config: %v", err)
	}

	info, err = customClient.GetInfo(ctx)
	if err != nil {
		log.Fatalf("Failed to authenticate with custom config: %v", err)
	}
	fmt.Printf("Successfully authenticated with custom config! API Version: %d\n", info.Version)
	fmt.Println()

	// Example 5: Authentication Error Handling
	fmt.Println("=== Authentication Error Handling ===")
	badClient, err := cfclient.NewWithPassword(
		"https://api.your-cf-domain.com",
		"wrong-username",
		"wrong-password",
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	_, err = badClient.GetInfo(ctx)
	if err != nil {
		fmt.Printf("Authentication failed as expected: %v\n", err)

		// Check if it's a CF API error
		if capiErr, ok := err.(*capi.ErrorResponse); ok {
			fmt.Printf("CF Error Details:\n")
			fmt.Printf("  Errors: %d\n", len(capiErr.Errors))

			for _, errDetail := range capiErr.Errors {
				fmt.Printf("  Error: %d - %s: %s\n", errDetail.Code, errDetail.Title, errDetail.Detail)
			}
		}
	}

	// Example 6: Environment-based Configuration
	fmt.Println("\n=== Environment-based Configuration ===")
	// This example shows how you might configure the client from environment variables
	// in a real application

	envConfig := &capi.Config{
		APIEndpoint:   getEnv("CF_API_ENDPOINT", "https://api.your-cf-domain.com"),
		Username:      getEnv("CF_USERNAME", ""),
		Password:      getEnv("CF_PASSWORD", ""),
		ClientID:      getEnv("CF_CLIENT_ID", ""),
		ClientSecret:  getEnv("CF_CLIENT_SECRET", ""),
		SkipTLSVerify: getEnv("CF_SKIP_SSL", "false") == "true",
	}

	fmt.Printf("Configuration from environment:\n")
	fmt.Printf("  API Endpoint: %s\n", envConfig.APIEndpoint)
	fmt.Printf("  Username: %s\n", maskString(envConfig.Username))
	fmt.Printf("  Client ID: %s\n", maskString(envConfig.ClientID))
	fmt.Printf("  Skip TLS: %v\n", envConfig.SkipTLSVerify)
}

// Helper function to get environment variable with default
func getEnv(key, defaultValue string) string {
	// In a real application, you would use os.Getenv(key)
	// For this example, we return the default
	return defaultValue
}

// Helper function to mask sensitive information
func maskString(s string) string {
	if s == "" {
		return "(not set)"
	}
	if len(s) <= 4 {
		return "****"
	}
	return s[:2] + "****" + s[len(s)-2:]
}
