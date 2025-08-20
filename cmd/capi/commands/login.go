package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/fivetwenty-io/capi-client/pkg/capi"
	"github.com/fivetwenty-io/capi-client/pkg/cfclient"
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
			if apiEndpoint == "" {
				apiEndpoint = viper.GetString("api")
			}
			if apiEndpoint == "" {
				reader := bufio.NewReader(os.Stdin)
				fmt.Print("API endpoint: ")
				apiEndpoint, _ = reader.ReadString('\n')
				apiEndpoint = strings.TrimSpace(apiEndpoint)
			}

			// Validate API endpoint
			if apiEndpoint == "" {
				return fmt.Errorf("API endpoint is required")
			}

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

			// Save configuration
			viper.Set("api", apiEndpoint)
			viper.Set("username", username)
			viper.Set("password", password) // In production, use a secure credential store
			viper.Set("skip_ssl_validation", skipSSL)

			// Save token if available (though it's managed internally)
			if tokenGetter, ok := client.(interface{ GetToken() string }); ok {
				viper.Set("token", tokenGetter.GetToken())
			}

			if err := saveConfig(); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}

			// Display success message
			fmt.Printf("Successfully logged in to %s\n", apiEndpoint)
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
	cmd.Flags().StringVarP(&apiEndpoint, "api", "a", "", "API endpoint URL")
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
