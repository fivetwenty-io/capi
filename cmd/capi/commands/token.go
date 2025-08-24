package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/fivetwenty-io/capi/v3/internal/auth"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// NewTokenCommand creates the token command group
func NewTokenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Manage authentication tokens",
		Long:  "Commands for managing authentication tokens including status and refresh",
	}

	cmd.AddCommand(newTokenStatusCommand())
	cmd.AddCommand(newTokenRefreshCommand())

	return cmd
}

func newTokenStatusCommand() *cobra.Command {
	var apiFlag string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show token status and expiration",
		Long:  "Display information about the current authentication token including expiration time",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get API config based on flag or current API
			apiConfig, err := getAPIConfigByFlag(apiFlag)
			if err != nil {
				return err
			}

			// Determine the API domain key for this config
			config := loadConfig()
			var apiDomain string
			for domain, cfg := range config.APIs {
				if cfg == apiConfig {
					apiDomain = domain
					break
				}
			}
			if apiDomain == "" {
				return fmt.Errorf("could not determine API domain for configuration")
			}

			return displayTokenStatus(apiConfig, apiDomain)
		},
	}

	cmd.Flags().StringVar(&apiFlag, "api", "", "show token status for specific API")

	return cmd
}

func newTokenRefreshCommand() *cobra.Command {
	var apiFlag string

	cmd := &cobra.Command{
		Use:   "refresh",
		Short: "Manually refresh authentication token",
		Long:  "Force refresh the authentication token using the stored refresh token",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get API config based on flag or current API
			apiConfig, err := getAPIConfigByFlag(apiFlag)
			if err != nil {
				return err
			}

			// Check if we have a refresh token
			if apiConfig.RefreshToken == "" {
				return fmt.Errorf("no refresh token available for this API, please run 'capi login' again")
			}

			// Determine the API domain key for this config
			config := loadConfig()
			var apiDomain string
			for domain, cfg := range config.APIs {
				if cfg == apiConfig {
					apiDomain = domain
					break
				}
			}
			if apiDomain == "" {
				return fmt.Errorf("could not determine API domain for configuration")
			}

			return refreshToken(apiConfig, apiDomain)
		},
	}

	cmd.Flags().StringVar(&apiFlag, "api", "", "refresh token for specific API")

	return cmd
}

func displayTokenStatus(apiConfig *APIConfig, apiDomain string) error {
	output := viper.GetString("output")

	tokenStatus := map[string]interface{}{
		"api_domain": apiDomain,
		"endpoint":   apiConfig.Endpoint,
	}

	// Check if we have a token
	if apiConfig.Token == "" {
		tokenStatus["status"] = "No token"
		tokenStatus["authenticated"] = false
	} else {
		tokenStatus["status"] = "Token present"
		tokenStatus["authenticated"] = true

		// Add expiration info if available
		if apiConfig.TokenExpiresAt != nil {
			tokenStatus["expires_at"] = apiConfig.TokenExpiresAt.Format(time.RFC3339)

			now := time.Now()
			timeUntilExpiry := apiConfig.TokenExpiresAt.Sub(now)

			if timeUntilExpiry <= 0 {
				tokenStatus["expiry_status"] = "Expired"
			} else if timeUntilExpiry <= 5*time.Minute {
				tokenStatus["expiry_status"] = "Expires soon"
			} else {
				tokenStatus["expiry_status"] = "Valid"
			}

			tokenStatus["time_until_expiry"] = timeUntilExpiry.String()
		} else {
			tokenStatus["expiry_status"] = "Unknown expiration"
		}

		// Add last refresh info if available
		if apiConfig.LastRefreshed != nil {
			tokenStatus["last_refreshed"] = apiConfig.LastRefreshed.Format(time.RFC3339)
		}

		// Check if we have a refresh token
		tokenStatus["refresh_token_available"] = apiConfig.RefreshToken != ""
	}

	switch output {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(tokenStatus)
	case "yaml":
		encoder := yaml.NewEncoder(os.Stdout)
		return encoder.Encode(tokenStatus)
	default:
		return displayTokenStatusTable(tokenStatus)
	}
}

func displayTokenStatusTable(tokenStatus map[string]interface{}) error {
	fmt.Printf("Token Status for API: %s\n", tokenStatus["api_domain"])
	fmt.Printf("Endpoint: %s\n\n", tokenStatus["endpoint"])

	table := tablewriter.NewWriter(os.Stdout)
	table.Header("Property", "Value")

	if err := table.Append([]string{"Authenticated", fmt.Sprintf("%v", tokenStatus["authenticated"])}); err != nil {
		return err
	}
	if err := table.Append([]string{"Status", fmt.Sprintf("%v", tokenStatus["status"])}); err != nil {
		return err
	}

	if expiryStatus, ok := tokenStatus["expiry_status"]; ok {
		if err := table.Append([]string{"Expiry Status", fmt.Sprintf("%v", expiryStatus)}); err != nil {
			return err
		}
	}

	if expiresAt, ok := tokenStatus["expires_at"]; ok {
		if err := table.Append([]string{"Expires At", fmt.Sprintf("%v", expiresAt)}); err != nil {
			return err
		}
	}

	if timeUntilExpiry, ok := tokenStatus["time_until_expiry"]; ok {
		if err := table.Append([]string{"Time Until Expiry", fmt.Sprintf("%v", timeUntilExpiry)}); err != nil {
			return err
		}
	}

	if lastRefreshed, ok := tokenStatus["last_refreshed"]; ok {
		if err := table.Append([]string{"Last Refreshed", fmt.Sprintf("%v", lastRefreshed)}); err != nil {
			return err
		}
	}

	if refreshAvailable, ok := tokenStatus["refresh_token_available"]; ok {
		if err := table.Append([]string{"Refresh Token Available", fmt.Sprintf("%v", refreshAvailable)}); err != nil {
			return err
		}
	}

	return table.Render()
}

func refreshToken(apiConfig *APIConfig, apiDomain string) error {
	fmt.Printf("Refreshing token for API: %s\n", apiDomain)

	// Create OAuth2 config for refresh
	uaaEndpoint := apiConfig.UAAEndpoint
	if uaaEndpoint == "" {
		// Try to discover from API endpoint
		var err error
		uaaEndpoint, err = discoverUAAEndpoint(apiConfig.Endpoint, apiConfig.SkipSSLValidation)
		if err != nil {
			return fmt.Errorf("could not determine UAA endpoint: %w", err)
		}
	}

	oauth2Config := &auth.OAuth2Config{
		TokenURL:     uaaEndpoint + "/oauth/token",
		ClientID:     "cf", // Default CF CLI client ID
		ClientSecret: "",
		RefreshToken: apiConfig.RefreshToken,
	}

	// Create token manager
	tokenManager := auth.NewOAuth2TokenManager(oauth2Config)

	// Force refresh
	ctx := context.Background()
	if err := tokenManager.RefreshToken(ctx); err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	// Get the new token
	newToken := tokenManager.GetTokenStore().Get()
	if newToken == nil {
		return fmt.Errorf("failed to retrieve refreshed token")
	}

	// Update config
	config := loadConfig()
	if config.APIs[apiDomain] == nil {
		return fmt.Errorf("API configuration not found")
	}

	config.APIs[apiDomain].Token = newToken.AccessToken
	if newToken.RefreshToken != "" {
		config.APIs[apiDomain].RefreshToken = newToken.RefreshToken
	}
	if !newToken.ExpiresAt.IsZero() {
		config.APIs[apiDomain].TokenExpiresAt = &newToken.ExpiresAt
	}
	now := time.Now()
	config.APIs[apiDomain].LastRefreshed = &now

	// Save config
	if err := saveConfigStruct(config); err != nil {
		return fmt.Errorf("failed to save updated token to config: %w", err)
	}

	fmt.Println("Token refreshed successfully!")

	// Show new expiration
	if !newToken.ExpiresAt.IsZero() {
		fmt.Printf("New token expires at: %s\n", newToken.ExpiresAt.Format(time.RFC3339))
		timeUntilExpiry := time.Until(newToken.ExpiresAt)
		fmt.Printf("Time until expiry: %s\n", timeUntilExpiry.String())
	}

	return nil
}
