package commands

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
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
	var showAll bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show token status and expiration",
		Long:  "Display information about the current authentication token including expiration time",
		RunE: func(cmd *cobra.Command, args []string) error {
			config := loadConfig()

			// If --all flag is specified, show all APIs
			if showAll {
				if len(config.APIs) == 0 {
					return fmt.Errorf("no APIs configured, use 'capi apis add' to add one")
				}
				return displayAllTokenStatus(config)
			}

			// If no API specified, show current API or all APIs
			if apiFlag == "" {
				if len(config.APIs) == 0 {
					return fmt.Errorf("no APIs configured, use 'capi apis add' to add one")
				}

				// If there's a current API, show just that one, otherwise show all
				if config.CurrentAPI != "" {
					if apiConfig, exists := config.APIs[config.CurrentAPI]; exists {
						return displayTokenStatus(apiConfig, config.CurrentAPI)
					}
				}

				// Show all APIs
				return displayAllTokenStatus(config)
			}

			// Get specific API config
			apiConfig, err := getAPIConfigByFlag(apiFlag)
			if err != nil {
				return err
			}

			// Find the API domain key by matching the endpoint or using the flag directly
			var apiDomain string

			// First try to use the flag directly as domain
			if _, exists := config.APIs[apiFlag]; exists {
				apiDomain = apiFlag
			} else {
				// Otherwise find by endpoint match
				for domain, cfg := range config.APIs {
					if cfg.Endpoint == apiConfig.Endpoint {
						apiDomain = domain
						break
					}
				}
			}

			if apiDomain == "" {
				return fmt.Errorf("could not determine API domain for '%s'", apiFlag)
			}

			return displayTokenStatus(apiConfig, apiDomain)
		},
	}

	cmd.Flags().StringVar(&apiFlag, "api", "", "show token status for specific API")
	cmd.Flags().BoolVar(&showAll, "all", false, "show token status for all configured APIs")

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

			// Find the API domain key by matching the endpoint or using the flag directly
			config := loadConfig()
			var apiDomain string

			// First try to use the flag directly as domain
			if apiFlag != "" {
				if _, exists := config.APIs[apiFlag]; exists {
					apiDomain = apiFlag
				} else {
					// Otherwise find by endpoint match
					for domain, cfg := range config.APIs {
						if cfg.Endpoint == apiConfig.Endpoint {
							apiDomain = domain
							break
						}
					}
				}
			} else {
				// No flag specified, use current API
				if config.CurrentAPI != "" {
					apiDomain = config.CurrentAPI
				}
			}

			if apiDomain == "" {
				return fmt.Errorf("could not determine API domain for '%s'", apiFlag)
			}

			return refreshToken(apiConfig, apiDomain)
		},
	}

	cmd.Flags().StringVar(&apiFlag, "api", "", "refresh token for specific API")

	return cmd
}

func displayTokenStatus(apiConfig *APIConfig, apiDomain string) error {
	output := viper.GetString("output")
	tokenStatus := buildTokenStatusData(apiConfig, apiDomain)

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

func displayAllTokenStatus(config *Config) error {
	output := viper.GetString("output")

	if output == "json" || output == "yaml" {
		// For structured output, show all APIs in one object
		allStatus := make(map[string]interface{})
		for domain, apiConfig := range config.APIs {
			tokenStatus := buildTokenStatusData(apiConfig, domain)
			allStatus[domain] = tokenStatus
		}

		switch output {
		case "json":
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			return encoder.Encode(allStatus)
		case "yaml":
			encoder := yaml.NewEncoder(os.Stdout)
			return encoder.Encode(allStatus)
		}
	}

	// For table output, show each API separately
	first := true
	for domain, apiConfig := range config.APIs {
		if !first {
			fmt.Println() // Add spacing between APIs
		}
		first = false

		if err := displayTokenStatus(apiConfig, domain); err != nil {
			return err
		}
	}

	return nil
}

func buildTokenStatusData(apiConfig *APIConfig, apiDomain string) map[string]interface{} {
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
		var expiresAt *time.Time
		if apiConfig.TokenExpiresAt != nil {
			expiresAt = apiConfig.TokenExpiresAt
		} else {
			// Try to decode from JWT token
			if jwtExp, err := decodeJWTExpiration(apiConfig.Token); err == nil {
				expiresAt = jwtExp
			}
		}

		if expiresAt != nil {
			tokenStatus["expires_at"] = expiresAt.Format(time.RFC3339)

			now := time.Now()
			timeUntilExpiry := expiresAt.Sub(now)

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

	return tokenStatus
}

// decodeJWTExpiration extracts expiration time from a JWT token
func decodeJWTExpiration(token string) (*time.Time, error) {
	// Split the JWT into its parts
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format")
	}

	// Decode the payload (second part)
	payload := parts[1]

	// Add padding if necessary
	if len(payload)%4 != 0 {
		payload += strings.Repeat("=", 4-len(payload)%4)
	}

	payloadBytes, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT payload: %w", err)
	}

	// Parse the JSON payload
	var claims struct {
		Exp int64 `json:"exp"`
	}

	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse JWT claims: %w", err)
	}

	if claims.Exp == 0 {
		return nil, fmt.Errorf("no expiration claim found")
	}

	expTime := time.Unix(claims.Exp, 0)
	return &expTime, nil
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
