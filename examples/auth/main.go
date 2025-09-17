package main

import (
	"context"
	"errors"
	"log"

	"github.com/fivetwenty-io/capi/v3/internal/constants"
	"github.com/fivetwenty-io/capi/v3/pkg/capi"
	"github.com/fivetwenty-io/capi/v3/pkg/cfclient"
)

func main() {
	ctx := context.Background()

	runAuthenticationExamples(ctx)
}

func runAuthenticationExamples(ctx context.Context) {
	userPassAuthExample(ctx)
	clientCredsAuthExample(ctx)
	accessTokenAuthExample(ctx)
	customConfigAuthExample(ctx)
	authErrorHandlingExample(ctx)
	environmentConfigExample()
}

func userPassAuthExample(ctx context.Context) {
	log.Println("=== Username/Password Authentication ===")

	client, err := cfclient.NewWithPassword(ctx,
		"https://api.your-cf-domain.com",
		"your-username",
		"your-password",
	)
	if err != nil {
		log.Fatalf("Failed to create client with username/password: %v", err)
	}

	testAuthentication(ctx, client, "username/password")
	log.Println()
}

func clientCredsAuthExample(ctx context.Context) {
	log.Println("=== Client Credentials Authentication ===")

	client, err := cfclient.NewWithClientCredentials(ctx,
		"https://api.your-cf-domain.com",
		"your-client-id",
		"your-client-secret",
	)
	if err != nil {
		log.Fatalf("Failed to create client with client credentials: %v", err)
	}

	testAuthenticationWithFallback(ctx, client, "client credentials")
	log.Println()
}

func accessTokenAuthExample(ctx context.Context) {
	log.Println("=== Access Token Authentication ===")
	// Note: You would typically get this token from another authentication flow
	accessToken := "your-access-token-here"

	client, err := cfclient.NewWithToken(ctx,
		"https://api.your-cf-domain.com",
		accessToken,
	)
	if err != nil {
		log.Fatalf("Failed to create client with access token: %v", err)
	}

	testAuthenticationWithFallback(ctx, client, "access token")
	log.Println()
}

func customConfigAuthExample(ctx context.Context) {
	log.Println("=== Custom Configuration ===")

	config := buildCustomConfig()

	client, err := cfclient.New(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create client with custom config: %v", err)
	}

	testAuthentication(ctx, client, "custom config")
	log.Println()
}

func buildCustomConfig() *capi.Config {
	return &capi.Config{
		APIEndpoint:   "https://api.your-cf-domain.com",
		Username:      "your-username",
		Password:      "your-password",
		SkipTLSVerify: false, // Set to true for self-signed certificates (not recommended for production)
		HTTPTimeout:   constants.DefaultHTTPTimeout,
		UserAgent:     "my-custom-app/1.0.0",
	}
}

func testAuthentication(ctx context.Context, client capi.Client, authType string) {
	info, err := client.GetInfo(ctx)
	if err != nil {
		log.Fatalf("Failed to authenticate with %s: %v", authType, err)
	}

	log.Printf("Successfully authenticated with %s! API Version: %d\n", authType, info.Version)
}

func testAuthenticationWithFallback(ctx context.Context, client capi.Client, authType string) {
	info, err := client.GetInfo(ctx)
	if err != nil {
		log.Printf("%s auth failed: %v", authType, err)
	} else {
		log.Printf("Successfully authenticated with %s! API Version: %d\n", authType, info.Version)
	}
}

func authErrorHandlingExample(ctx context.Context) {
	log.Println("=== Authentication Error Handling ===")

	badClient, err := cfclient.NewWithPassword(ctx,
		"https://api.your-cf-domain.com",
		"wrong-username",
		"wrong-password",
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	_, err = badClient.GetInfo(ctx)
	if err != nil {
		log.Printf("Authentication failed as expected: %v\n", err)
		handleAuthenticationError(err)
	}

	log.Println()
}

func handleAuthenticationError(err error) {
	capiErr := &capi.ResponseError{}
	if !errors.As(err, &capiErr) {
		return
	}

	log.Printf("CF Error Details:\n")
	log.Printf("  Errors: %d\n", len(capiErr.Errors))

	for _, errDetail := range capiErr.Errors {
		log.Printf("  Error: %d - %s: %s\n", errDetail.Code, errDetail.Title, errDetail.Detail)
	}
}

func environmentConfigExample() {
	log.Println("=== Environment-based Configuration ===")
	// This example shows how you might configure the client from environment variables
	// in a real application

	envConfig := buildEnvironmentConfig()
	printEnvironmentConfig(envConfig)
}

func buildEnvironmentConfig() *capi.Config {
	return &capi.Config{
		APIEndpoint:   getEnv("CF_API_ENDPOINT", "https://api.your-cf-domain.com"),
		Username:      getEnv("CF_USERNAME", ""),
		Password:      getEnv("CF_PASSWORD", ""),
		ClientID:      getEnv("CF_CLIENT_ID", ""),
		ClientSecret:  getEnv("CF_CLIENT_SECRET", ""),
		SkipTLSVerify: getEnv("CF_SKIP_SSL", "false") == "true",
	}
}

func printEnvironmentConfig(config *capi.Config) {
	log.Printf("Configuration from environment:\n")
	log.Printf("  API Endpoint: %s\n", config.APIEndpoint)
	log.Printf("  Username: %s\n", maskString(config.Username))
	log.Printf("  Client ID: %s\n", maskString(config.ClientID))
	log.Printf("  Skip TLS: %v\n", config.SkipTLSVerify)
}

// Helper function to get environment variable with default.
func getEnv(key, defaultValue string) string {
	// In a real application, you would use os.Getenv(key)
	// For this example, we return the default
	return defaultValue
}

// Helper function to mask sensitive information.
func maskString(str string) string {
	if str == "" {
		return "(not set)"
	}

	if len(str) <= constants.StringTruncationLimit {
		return "****"
	}

	return str[:2] + "****" + str[len(str)-2:]
}
