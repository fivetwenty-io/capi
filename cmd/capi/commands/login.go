package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/fivetwenty-io/capi/v3/pkg/capi"
	"github.com/fivetwenty-io/capi/v3/pkg/cfclient"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

// NewLoginCommand creates the login command
func NewLoginCommand() *cobra.Command {
	var (
		apiEndpoint  string
		username     string
		password     string
		clientID     string
		clientSecret string
		ssoCode      string
		ssoPasscode  string
	)

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to Cloud Foundry",
		Long:  "Authenticate with a Cloud Foundry API endpoint",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get API endpoint
			originalInput := apiEndpoint
			if apiEndpoint == "" {
				apiEndpoint = viper.GetString("api")
				originalInput = apiEndpoint
			}

			// If still no API endpoint, try to use current API from config
			if apiEndpoint == "" {
				config := loadConfig()
				if config.CurrentAPI != "" {
					if _, exists := config.APIs[config.CurrentAPI]; exists {
						apiEndpoint = config.CurrentAPI // Use the short name, it will be resolved below
						originalInput = config.CurrentAPI
					}
				}
			}

			if apiEndpoint == "" {
				reader := bufio.NewReader(os.Stdin)
				fmt.Print("API endpoint (or short name): ")
				apiEndpoint, _ = reader.ReadString('\n')
				apiEndpoint = strings.TrimSpace(apiEndpoint)
				originalInput = apiEndpoint
			}

			// Validate API endpoint
			if apiEndpoint == "" {
				return fmt.Errorf("API endpoint is required")
			}

			// Resolve short name to endpoint if applicable
			resolvedEndpoint, err := ResolveAPIEndpoint(apiEndpoint)
			if err != nil {
				return err
			}
			apiEndpoint = resolvedEndpoint

			// Get skip SSL validation setting
			skipSSL := viper.GetBool("skip_ssl_validation")

			// Create config for client
			config := &capi.Config{
				APIEndpoint:   apiEndpoint,
				SkipTLSVerify: skipSSL,
			}

			// Determine authentication method
			if clientID != "" && clientSecret != "" {
				// Client credentials flow
				config.ClientID = clientID
				config.ClientSecret = clientSecret
			} else if ssoCode != "" || ssoPasscode != "" {
				// SSO flow - would need additional implementation
				// For now, use the access token if available
				config.AccessToken = ssoPasscode
			} else {
				// Username/password flow
				if username == "" {
					reader := bufio.NewReader(os.Stdin)
					fmt.Print("Username: ")
					username, _ = reader.ReadString('\n')
					username = strings.TrimSpace(username)
				}

				if password == "" {
					fmt.Print("Password: ")
					bytePassword, err := term.ReadPassword(int(syscall.Stdin))
					if err != nil {
						return fmt.Errorf("failed to read password: %w", err)
					}
					password = string(bytePassword)
					fmt.Println()
				}

				config.Username = username
				config.Password = password
			}

			// Create client
			client, err := cfclient.New(config)
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}

			// Test connection by getting info
			ctx := context.Background()
			info, err := client.GetInfo(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect to API: %w", err)
			}

			// Get root info to fetch all available API endpoints
			rootInfo, err := client.GetRootInfo(ctx)
			if err != nil {
				// Non-fatal: log warning but continue
				fmt.Printf("Warning: could not fetch API endpoints: %v\n", err)
			}

			// Normalize endpoint
			normalizedEndpoint, err := normalizeEndpoint(apiEndpoint)
			if err != nil {
				return fmt.Errorf("invalid API endpoint: %w", err)
			}

			// Determine the key to use for storing the API config
			// If the original input was a short name, preserve it
			var configKey string
			currentConfig := loadConfig()
			if _, exists := currentConfig.APIs[originalInput]; exists {
				// The original input was a short name that exists in config
				configKey = originalInput
			} else {
				// Extract domain for use as key (for new APIs or direct URLs)
				configKey = extractDomainFromEndpoint(normalizedEndpoint)
			}

			// Load current configuration
			configStruct := loadConfig()

			// Initialize APIs map if needed
			if configStruct.APIs == nil {
				configStruct.APIs = make(map[string]*APIConfig)
			}

			// Get or create API config
			apiConfig, exists := configStruct.APIs[configKey]
			if !exists {
				apiConfig = &APIConfig{
					Endpoint: normalizedEndpoint,
				}
				configStruct.APIs[configKey] = apiConfig
			}

			// Store authentication information (tokens only, not passwords)
			apiConfig.Username = username
			apiConfig.SkipSSLValidation = skipSSL

			// Save token if available
			if tokenGetter, ok := client.(interface {
				GetToken(context.Context) (string, error)
			}); ok {
				if token, err := tokenGetter.GetToken(ctx); err == nil && token != "" {
					apiConfig.Token = token
				}
			}

			// Store API links from root info if available
			if rootInfo != nil && rootInfo.Links != nil {
				apiConfig.APILinks = make(map[string]string)
				for key, link := range rootInfo.Links {
					apiConfig.APILinks[key] = link.Href
				}
			}

			// Set as current API if this is the first one or no current API is set
			if configStruct.CurrentAPI == "" || len(configStruct.APIs) == 1 {
				configStruct.CurrentAPI = configKey
			}

			// Save configuration
			if err := saveConfigStruct(configStruct); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}

			// Display success message
			isFirstAPI := len(configStruct.APIs) == 1
			fmt.Printf("Successfully logged in to %s\n", normalizedEndpoint)
			if isFirstAPI {
				fmt.Printf("API '%s' set as current target\n", configKey)
			}
			fmt.Printf("API version: %d\n", info.Version)

			// List organizations if available
			if orgsClient, ok := client.(interface {
				Organizations() capi.OrganizationsClient
			}); ok {
				orgs, err := orgsClient.Organizations().List(ctx, nil)
				if err == nil && len(orgs.Resources) > 0 {
					fmt.Println("\nAvailable organizations:")
					for _, org := range orgs.Resources {
						fmt.Printf("  - %s\n", org.Name)
					}
					fmt.Println("\nUse 'capi target -o <org>' to target an organization")
				}
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&apiEndpoint, "api", "a", "", "API endpoint URL or short name from config")
	cmd.Flags().StringVarP(&username, "username", "u", "", "username for authentication")
	cmd.Flags().StringVarP(&password, "password", "p", "", "password for authentication")
	cmd.Flags().StringVar(&clientID, "client-id", "", "OAuth2 client ID")
	cmd.Flags().StringVar(&clientSecret, "client-secret", "", "OAuth2 client secret")
	cmd.Flags().StringVar(&ssoCode, "sso-code", "", "SSO authorization code")
	cmd.Flags().StringVar(&ssoPasscode, "sso-passcode", "", "SSO one-time passcode")
	cmd.Flags().Bool("skip-ssl-validation", false, "skip SSL certificate validation")

	return cmd
}

// NewLogoutCommand creates the logout command
func NewLogoutCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Logout from Cloud Foundry",
		Long:  "Clear authentication credentials and logout from Cloud Foundry",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Clear authentication data
			viper.Set("token", "")
			viper.Set("refresh_token", "")
			viper.Set("username", "")
			viper.Set("password", "")
			viper.Set("organization", "")
			viper.Set("space", "")

			if err := saveConfig(); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}

			fmt.Println("Successfully logged out")
			return nil
		},
	}
}
