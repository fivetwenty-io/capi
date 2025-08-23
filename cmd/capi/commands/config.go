package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fivetwenty-io/capi/v3/pkg/capi"
	"github.com/fivetwenty-io/capi/v3/pkg/cfclient"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config represents the CLI configuration
type Config struct {
	// Multi-API configuration
	APIs       map[string]*APIConfig `json:"apis,omitempty" yaml:"apis,omitempty"`
	CurrentAPI string                `json:"current_api,omitempty" yaml:"current_api,omitempty"`

	// Global settings
	Output  string `json:"output" yaml:"output"`
	NoColor bool   `json:"no_color" yaml:"no_color"`

	// Legacy fields for backward compatibility (will be migrated to APIs map)
	API               string            `json:"api,omitempty" yaml:"api,omitempty"`
	Token             string            `json:"token,omitempty" yaml:"token,omitempty"`
	RefreshToken      string            `json:"refresh_token,omitempty" yaml:"refresh_token,omitempty"`
	Username          string            `json:"username,omitempty" yaml:"username,omitempty"`
	Organization      string            `json:"organization,omitempty" yaml:"organization,omitempty"`
	OrganizationGUID  string            `json:"organization_guid,omitempty" yaml:"organization_guid,omitempty"`
	Space             string            `json:"space,omitempty" yaml:"space,omitempty"`
	SpaceGUID         string            `json:"space_guid,omitempty" yaml:"space_guid,omitempty"`
	SkipSSLValidation bool              `json:"skip_ssl_validation" yaml:"skip_ssl_validation"`
	Targets           map[string]Target `json:"targets,omitempty" yaml:"targets,omitempty"`
	CurrentTarget     string            `json:"current_target,omitempty" yaml:"current_target,omitempty"`
	UAAEndpoint       string            `json:"uaa_endpoint,omitempty" yaml:"uaa_endpoint,omitempty"`
	UAAToken          string            `json:"uaa_token,omitempty" yaml:"uaa_token,omitempty"`
	UAARefreshToken   string            `json:"uaa_refresh_token,omitempty" yaml:"uaa_refresh_token,omitempty"`
	UAAClientID       string            `json:"uaa_client_id,omitempty" yaml:"uaa_client_id,omitempty"`
	UAAClientSecret   string            `json:"uaa_client_secret,omitempty" yaml:"uaa_client_secret,omitempty"`
}

// APIConfig represents configuration for a single Cloud Foundry API endpoint
type APIConfig struct {
	Endpoint          string            `json:"endpoint" yaml:"endpoint"`
	Token             string            `json:"token,omitempty" yaml:"token,omitempty"`
	RefreshToken      string            `json:"refresh_token,omitempty" yaml:"refresh_token,omitempty"`
	Username          string            `json:"username,omitempty" yaml:"username,omitempty"`
	Organization      string            `json:"organization,omitempty" yaml:"organization,omitempty"`
	OrganizationGUID  string            `json:"organization_guid,omitempty" yaml:"organization_guid,omitempty"`
	Space             string            `json:"space,omitempty" yaml:"space,omitempty"`
	SpaceGUID         string            `json:"space_guid,omitempty" yaml:"space_guid,omitempty"`
	SkipSSLValidation bool              `json:"skip_ssl_validation" yaml:"skip_ssl_validation"`
	UAAEndpoint       string            `json:"uaa_endpoint,omitempty" yaml:"uaa_endpoint,omitempty"`
	UAAToken          string            `json:"uaa_token,omitempty" yaml:"uaa_token,omitempty"`
	UAARefreshToken   string            `json:"uaa_refresh_token,omitempty" yaml:"uaa_refresh_token,omitempty"`
	UAAClientID       string            `json:"uaa_client_id,omitempty" yaml:"uaa_client_id,omitempty"`
	UAAClientSecret   string            `json:"uaa_client_secret,omitempty" yaml:"uaa_client_secret,omitempty"`
	APILinks          map[string]string `json:"api_links,omitempty" yaml:"api_links,omitempty"`
}

// Target represents a saved CF target
type Target struct {
	API               string `json:"api" yaml:"api"`
	Token             string `json:"token,omitempty" yaml:"token,omitempty"`
	RefreshToken      string `json:"refresh_token,omitempty" yaml:"refresh_token,omitempty"`
	Username          string `json:"username,omitempty" yaml:"username,omitempty"`
	Organization      string `json:"organization,omitempty" yaml:"organization,omitempty"`
	Space             string `json:"space,omitempty" yaml:"space,omitempty"`
	SkipSSLValidation bool   `json:"skip_ssl_validation" yaml:"skip_ssl_validation"`
	// UAA-specific fields for targets
	UAAEndpoint     string `json:"uaa_endpoint,omitempty" yaml:"uaa_endpoint,omitempty"`
	UAAToken        string `json:"uaa_token,omitempty" yaml:"uaa_token,omitempty"`
	UAARefreshToken string `json:"uaa_refresh_token,omitempty" yaml:"uaa_refresh_token,omitempty"`
	UAAClientID     string `json:"uaa_client_id,omitempty" yaml:"uaa_client_id,omitempty"`
	UAAClientSecret string `json:"uaa_client_secret,omitempty" yaml:"uaa_client_secret,omitempty"`
}

// NewConfigCommand creates the config command group
func NewConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration",
		Long:  "Manage CAPI CLI configuration including targets and settings",
	}

	cmd.AddCommand(newConfigShowCommand())
	cmd.AddCommand(newConfigSetCommand())
	cmd.AddCommand(newConfigUnsetCommand())
	cmd.AddCommand(newConfigClearCommand())

	return cmd
}

func newConfigShowCommand() *cobra.Command {
	var apiFlag string

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Long:  "Display the current CLI configuration (global or API-specific)",
		RunE: func(cmd *cobra.Command, args []string) error {
			config := loadConfig()

			// If --api flag is provided, show only that API's configuration
			if apiFlag != "" {
				return showAPISpecificConfig(config, apiFlag)
			}

			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(config)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(config)
			default:
				return displayConfigTable(config)
			}
		},
	}

	cmd.Flags().StringVar(&apiFlag, "api", "", "show configuration for specific API")

	return cmd
}

func newConfigSetCommand() *cobra.Command {
	var apiFlag string

	cmd := &cobra.Command{
		Use:   "set KEY VALUE",
		Short: "Set a configuration value",
		Long:  "Set a specific configuration value (global or API-specific)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]

			config := loadConfig()

			// If --api flag is provided, set API-specific configuration
			if apiFlag != "" {
				return setAPISpecificConfig(config, apiFlag, key, value)
			}

			// Otherwise set global configuration
			return setGlobalConfig(config, key, value)
		},
	}

	cmd.Flags().StringVar(&apiFlag, "api", "", "target specific API for configuration")

	return cmd
}

func newConfigUnsetCommand() *cobra.Command {
	var apiFlag string

	cmd := &cobra.Command{
		Use:   "unset KEY",
		Short: "Unset a configuration value",
		Long:  "Remove a specific configuration value (global or API-specific)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			config := loadConfig()

			// If --api flag is provided, unset API-specific configuration
			if apiFlag != "" {
				return unsetAPISpecificConfig(config, apiFlag, key)
			}

			// Otherwise unset global configuration
			return unsetGlobalConfig(config, key)
		},
	}

	cmd.Flags().StringVar(&apiFlag, "api", "", "target specific API for configuration")

	return cmd
}

func newConfigClearCommand() *cobra.Command {
	var apiFlag string

	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear configuration",
		Long:  "Remove all configuration settings (global or API-specific)",
		RunE: func(cmd *cobra.Command, args []string) error {
			config := loadConfig()

			// If --api flag is provided, clear only that API's configuration
			if apiFlag != "" {
				return clearAPISpecificConfig(config, apiFlag)
			}

			// Otherwise clear all configuration
			configFile := viper.ConfigFileUsed()
			if configFile == "" {
				home, _ := os.UserHomeDir()
				configFile = filepath.Join(home, ".capi", "config.yml")
			}

			if err := os.Remove(configFile); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to remove config file: %w", err)
			}

			return outputConfigUpdateResult("Cleared", "all configuration", "", "")
		},
	}

	cmd.Flags().StringVar(&apiFlag, "api", "", "clear configuration for specific API only")

	return cmd
}

func loadConfig() *Config {
	config := &Config{
		// Global settings
		Output:  viper.GetString("output"),
		NoColor: viper.GetBool("no_color"),

		// Initialize APIs map
		APIs: make(map[string]*APIConfig),

		// Legacy fields for migration
		API:               viper.GetString("api"),
		Token:             viper.GetString("token"),
		RefreshToken:      viper.GetString("refresh_token"),
		Username:          viper.GetString("username"),
		Organization:      viper.GetString("organization"),
		OrganizationGUID:  viper.GetString("organization_guid"),
		Space:             viper.GetString("space"),
		SpaceGUID:         viper.GetString("space_guid"),
		SkipSSLValidation: viper.GetBool("skip_ssl_validation"),
		CurrentTarget:     viper.GetString("current_target"),
		Targets:           make(map[string]Target),
		UAAEndpoint:       viper.GetString("uaa_endpoint"),
		UAAToken:          viper.GetString("uaa_token"),
		UAARefreshToken:   viper.GetString("uaa_refresh_token"),
		UAAClientID:       viper.GetString("uaa_client_id"),
		UAAClientSecret:   viper.GetString("uaa_client_secret"),
	}

	// Load multi-API configuration if it exists
	config.CurrentAPI = viper.GetString("current_api")
	if apisRaw := viper.GetStringMap("apis"); apisRaw != nil {
		for domain, apiRaw := range apisRaw {
			if apiMap, ok := apiRaw.(map[string]interface{}); ok {
				apiConfig := &APIConfig{}
				if endpoint, ok := apiMap["endpoint"].(string); ok {
					apiConfig.Endpoint = endpoint
				}
				if token, ok := apiMap["token"].(string); ok {
					apiConfig.Token = token
				}
				if refreshToken, ok := apiMap["refresh_token"].(string); ok {
					apiConfig.RefreshToken = refreshToken
				}
				if username, ok := apiMap["username"].(string); ok {
					apiConfig.Username = username
				}
				if org, ok := apiMap["organization"].(string); ok {
					apiConfig.Organization = org
				}
				if orgGUID, ok := apiMap["organization_guid"].(string); ok {
					apiConfig.OrganizationGUID = orgGUID
				}
				if space, ok := apiMap["space"].(string); ok {
					apiConfig.Space = space
				}
				if spaceGUID, ok := apiMap["space_guid"].(string); ok {
					apiConfig.SpaceGUID = spaceGUID
				}
				if skipSSL, ok := apiMap["skip_ssl_validation"].(bool); ok {
					apiConfig.SkipSSLValidation = skipSSL
				}
				if uaaEndpoint, ok := apiMap["uaa_endpoint"].(string); ok {
					apiConfig.UAAEndpoint = uaaEndpoint
				}
				if uaaToken, ok := apiMap["uaa_token"].(string); ok {
					apiConfig.UAAToken = uaaToken
				}
				if uaaRefreshToken, ok := apiMap["uaa_refresh_token"].(string); ok {
					apiConfig.UAARefreshToken = uaaRefreshToken
				}
				if uaaClientID, ok := apiMap["uaa_client_id"].(string); ok {
					apiConfig.UAAClientID = uaaClientID
				}
				if uaaClientSecret, ok := apiMap["uaa_client_secret"].(string); ok {
					apiConfig.UAAClientSecret = uaaClientSecret
				}
				config.APIs[domain] = apiConfig
			}
		}
	}

	// Handle migration from legacy single-API config
	if config.API != "" && len(config.APIs) == 0 {
		config = migrateFromLegacyConfig(config)
	}

	// Convert legacy targets from viper (for backward compatibility)
	if targetsRaw := viper.GetStringMap("targets"); targetsRaw != nil {
		for name, targetRaw := range targetsRaw {
			if targetMap, ok := targetRaw.(map[string]interface{}); ok {
				target := Target{}
				if api, ok := targetMap["api"].(string); ok {
					target.API = api
				}
				if token, ok := targetMap["token"].(string); ok {
					target.Token = token
				}
				if refreshToken, ok := targetMap["refresh_token"].(string); ok {
					target.RefreshToken = refreshToken
				}
				if username, ok := targetMap["username"].(string); ok {
					target.Username = username
				}
				if org, ok := targetMap["organization"].(string); ok {
					target.Organization = org
				}
				if space, ok := targetMap["space"].(string); ok {
					target.Space = space
				}
				if skipSSL, ok := targetMap["skip_ssl_validation"].(bool); ok {
					target.SkipSSLValidation = skipSSL
				}
				if uaaEndpoint, ok := targetMap["uaa_endpoint"].(string); ok {
					target.UAAEndpoint = uaaEndpoint
				}
				if uaaToken, ok := targetMap["uaa_token"].(string); ok {
					target.UAAToken = uaaToken
				}
				if uaaRefreshToken, ok := targetMap["uaa_refresh_token"].(string); ok {
					target.UAARefreshToken = uaaRefreshToken
				}
				if uaaClientID, ok := targetMap["uaa_client_id"].(string); ok {
					target.UAAClientID = uaaClientID
				}
				if uaaClientSecret, ok := targetMap["uaa_client_secret"].(string); ok {
					target.UAAClientSecret = uaaClientSecret
				}
				config.Targets[name] = target
			}
		}
	}

	return config
}

func saveConfig() error {
	// Load current config to check for migration needs
	config := loadConfig()

	// Use the struct-based save which handles migration and backup
	return saveConfigStruct(config)
}

func saveConfigStruct(config *Config) error {
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		configDir := filepath.Join(home, ".capi")
		if err := os.MkdirAll(configDir, 0750); err != nil {
			return err
		}
		configFile = filepath.Join(configDir, "config.yml")
	}

	// Create backup before migration if APIs are being used and legacy fields exist
	if len(config.APIs) > 0 && config.API != "" {
		backupFile := configFile + ".backup"
		if _, err := os.Stat(backupFile); os.IsNotExist(err) {
			// Read current config and save as backup
			// configFile is securely constructed from user home dir and is not user-controllable
			// #nosec G304
			if currentData, err := os.ReadFile(configFile); err == nil {
				_ = os.WriteFile(backupFile, currentData, 0600)
			}
		}
	}

	// Clear legacy fields if migration occurred
	if len(config.APIs) > 0 {
		config.API = ""
		config.Token = ""
		config.RefreshToken = ""
		config.Username = ""
		config.Organization = ""
		config.OrganizationGUID = ""
		config.Space = ""
		config.SpaceGUID = ""
		config.SkipSSLValidation = false
		config.UAAEndpoint = ""
		config.UAAToken = ""
		config.UAARefreshToken = ""
		config.UAAClientID = ""
		config.UAAClientSecret = ""
		// Keep legacy targets for now but they're deprecated
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0600)
}

// migrateFromLegacyConfig converts legacy single-API config to new multi-API format
func migrateFromLegacyConfig(config *Config) *Config {
	if config.API == "" {
		return config
	}

	// Extract domain from API endpoint for use as key
	domain := extractDomainFromEndpoint(config.API)

	// Create new API config from legacy fields
	apiConfig := &APIConfig{
		Endpoint:          config.API,
		Token:             config.Token,
		RefreshToken:      config.RefreshToken,
		Username:          config.Username,
		Organization:      config.Organization,
		OrganizationGUID:  config.OrganizationGUID,
		Space:             config.Space,
		SpaceGUID:         config.SpaceGUID,
		SkipSSLValidation: config.SkipSSLValidation,
		UAAEndpoint:       config.UAAEndpoint,
		UAAToken:          config.UAAToken,
		UAARefreshToken:   config.UAARefreshToken,
		UAAClientID:       config.UAAClientID,
		UAAClientSecret:   config.UAAClientSecret,
	}

	// Add to APIs map and set as current
	config.APIs[domain] = apiConfig
	config.CurrentAPI = domain

	// Clear legacy fields after migration
	config.API = ""
	config.Token = ""
	config.RefreshToken = ""
	config.Username = ""
	config.Organization = ""
	config.OrganizationGUID = ""
	config.Space = ""
	config.SpaceGUID = ""
	config.SkipSSLValidation = false
	config.UAAEndpoint = ""
	config.UAAToken = ""
	config.UAARefreshToken = ""
	config.UAAClientID = ""
	config.UAAClientSecret = ""

	return config
}

// extractDomainFromEndpoint extracts the domain portion from a CF API endpoint
func extractDomainFromEndpoint(endpoint string) string {
	// Remove protocol if present
	domain := endpoint
	if strings.HasPrefix(domain, "https://") {
		domain = strings.TrimPrefix(domain, "https://")
	} else if strings.HasPrefix(domain, "http://") {
		domain = strings.TrimPrefix(domain, "http://")
	}

	// Remove path if present
	if idx := strings.Index(domain, "/"); idx != -1 {
		domain = domain[:idx]
	}

	// Remove port if present
	if idx := strings.Index(domain, ":"); idx != -1 {
		domain = domain[:idx]
	}

	return domain
}

// getCurrentAPIConfig returns the configuration for the currently targeted API
func getCurrentAPIConfig() (*APIConfig, error) {
	config := loadConfig()

	if config.CurrentAPI == "" {
		if len(config.APIs) == 0 {
			return nil, fmt.Errorf("no APIs configured, use 'capi apis add' to add one")
		}
		// If no current API set but APIs exist, use the first one
		for domain := range config.APIs {
			config.CurrentAPI = domain
			break
		}
	}

	apiConfig, exists := config.APIs[config.CurrentAPI]
	if !exists {
		return nil, fmt.Errorf("current API '%s' not found in configuration", config.CurrentAPI)
	}

	return apiConfig, nil
}

// getAPIConfigByFlag returns API config based on command line flag or current API
func getAPIConfigByFlag(apiFlag string) (*APIConfig, error) {
	config := loadConfig()

	// If --api flag is provided, use that specific API
	if apiFlag != "" {
		// First try to resolve it as a short name or endpoint
		resolvedEndpoint, err := ResolveAPIEndpoint(apiFlag)
		if err != nil {
			return nil, err
		}

		// Check if the original apiFlag is a short name in our config
		if apiConfig, exists := config.APIs[apiFlag]; exists {
			return apiConfig, nil
		}

		// Otherwise look for it by resolved endpoint
		for _, apiConfig := range config.APIs {
			if apiConfig.Endpoint == resolvedEndpoint {
				return apiConfig, nil
			}
		}

		return nil, fmt.Errorf("API '%s' not found in configuration, use 'capi apis list' to see available APIs", apiFlag)
	}

	// Otherwise use current API
	return getCurrentAPIConfig()
}

// ResolveAPIEndpoint resolves a short name or returns the endpoint if it's already a URL
func ResolveAPIEndpoint(apiNameOrEndpoint string) (string, error) {
	if apiNameOrEndpoint == "" {
		return "", fmt.Errorf("API name or endpoint is required")
	}

	config := loadConfig()

	// Check if it's a short name in the APIs map
	if apiConfig, exists := config.APIs[apiNameOrEndpoint]; exists {
		return apiConfig.Endpoint, nil
	}

	// If not found in config, treat as direct endpoint URL
	return apiNameOrEndpoint, nil
}

// GetEffectiveUAAEndpoint returns the effective UAA endpoint from either legacy config or current API
func GetEffectiveUAAEndpoint(config *Config) string {
	// Check legacy UAA endpoint first
	if config.UAAEndpoint != "" {
		return config.UAAEndpoint
	}

	// Check current API configuration for UAA endpoint
	if config.CurrentAPI != "" {
		if apiConfig, exists := config.APIs[config.CurrentAPI]; exists && apiConfig.UAAEndpoint != "" {
			return apiConfig.UAAEndpoint
		}
	}

	return ""
}

// CreateClientWithAPI creates a CAPI client using the specified API or current API
func CreateClientWithAPI(apiFlag string) (capi.Client, error) {
	// Get API config based on flag or current API
	apiConfig, err := getAPIConfigByFlag(apiFlag)
	if err != nil {
		return nil, err
	}

	if apiConfig.Endpoint == "" {
		return nil, fmt.Errorf("no API endpoint configured, use 'capi apis add' first")
	}

	// Set the API config values in viper so other commands can access them
	viper.Set("api", apiConfig.Endpoint)
	viper.Set("organization", apiConfig.Organization)
	viper.Set("organization_guid", apiConfig.OrganizationGUID)
	viper.Set("space", apiConfig.Space)
	viper.Set("space_guid", apiConfig.SpaceGUID)
	viper.Set("username", apiConfig.Username)
	viper.Set("uaa_endpoint", apiConfig.UAAEndpoint)

	config := &capi.Config{
		APIEndpoint:   apiConfig.Endpoint,
		AccessToken:   apiConfig.Token,
		SkipTLSVerify: apiConfig.SkipSSLValidation,
		Username:      apiConfig.Username,
	}

	// If we have no token and no username, require authentication
	if config.AccessToken == "" && config.Username == "" {
		return nil, fmt.Errorf("not authenticated, use 'capi login' first")
	}

	return cfclient.New(config)
}

// setGlobalConfig sets a global configuration value
func setGlobalConfig(config *Config, key, value string) error {
	switch key {
	case "output":
		config.Output = value
	case "no_color":
		if value == "true" || value == "1" {
			config.NoColor = true
		} else {
			config.NoColor = false
		}
	default:
		return fmt.Errorf("unknown global configuration key: %s. Use --api flag for API-specific settings", key)
	}

	if err := saveConfigStruct(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return outputConfigUpdateResult("Set global", key, value, "")
}

// setAPISpecificConfig sets configuration for a specific API
func setAPISpecificConfig(config *Config, apiDomain, key, value string) error {
	apiConfig, exists := config.APIs[apiDomain]
	if !exists {
		return fmt.Errorf("API '%s' not found. Use 'capi apis list' to see available APIs", apiDomain)
	}

	switch key {
	case "username":
		apiConfig.Username = value
	case "organization":
		apiConfig.Organization = value
	case "organization_guid":
		apiConfig.OrganizationGUID = value
	case "space":
		apiConfig.Space = value
	case "space_guid":
		apiConfig.SpaceGUID = value
	case "skip_ssl_validation":
		if value == "true" || value == "1" {
			apiConfig.SkipSSLValidation = true
		} else {
			apiConfig.SkipSSLValidation = false
		}
	case "uaa_endpoint":
		apiConfig.UAAEndpoint = value
	case "uaa_client_id":
		apiConfig.UAAClientID = value
	case "uaa_client_secret":
		apiConfig.UAAClientSecret = value
	default:
		return fmt.Errorf("unknown API-specific configuration key: %s", key)
	}

	// Update the API config in the main config
	config.APIs[apiDomain] = apiConfig

	if err := saveConfigStruct(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return outputConfigUpdateResult("Set", key, value, apiDomain)
}

// unsetGlobalConfig unsets a global configuration value
func unsetGlobalConfig(config *Config, key string) error {
	switch key {
	case "output":
		config.Output = "table"
	case "no_color":
		config.NoColor = false
	default:
		return fmt.Errorf("unknown global configuration key: %s. Use --api flag for API-specific settings", key)
	}

	if err := saveConfigStruct(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return outputConfigUpdateResult("Unset global", key, "", "")
}

// unsetAPISpecificConfig unsets configuration for a specific API
func unsetAPISpecificConfig(config *Config, apiDomain, key string) error {
	apiConfig, exists := config.APIs[apiDomain]
	if !exists {
		return fmt.Errorf("API '%s' not found. Use 'capi apis list' to see available APIs", apiDomain)
	}

	switch key {
	case "username":
		apiConfig.Username = ""
	case "organization":
		apiConfig.Organization = ""
	case "organization_guid":
		apiConfig.OrganizationGUID = ""
	case "space":
		apiConfig.Space = ""
	case "space_guid":
		apiConfig.SpaceGUID = ""
	case "skip_ssl_validation":
		apiConfig.SkipSSLValidation = false
	case "uaa_endpoint":
		apiConfig.UAAEndpoint = ""
	case "uaa_client_id":
		apiConfig.UAAClientID = ""
	case "uaa_client_secret":
		apiConfig.UAAClientSecret = ""
	// Token fields should not be unset via config command for security
	case "token", "refresh_token", "uaa_token", "uaa_refresh_token":
		return fmt.Errorf("token fields cannot be unset via config command. Use 'capi logout' instead")
	default:
		return fmt.Errorf("unknown API-specific configuration key: %s", key)
	}

	// Update the API config in the main config
	config.APIs[apiDomain] = apiConfig

	if err := saveConfigStruct(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return outputConfigUpdateResult("Unset", key, "", apiDomain)
}

// showAPISpecificConfig shows configuration for a specific API
func showAPISpecificConfig(config *Config, apiDomain string) error {
	apiConfig, exists := config.APIs[apiDomain]
	if !exists {
		return fmt.Errorf("API '%s' not found. Use 'capi apis list' to see available APIs", apiDomain)
	}

	output := viper.GetString("output")
	switch output {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(apiConfig)
	case "yaml":
		encoder := yaml.NewEncoder(os.Stdout)
		return encoder.Encode(apiConfig)
	default:
		return displayAPISpecificConfigTable(config, apiDomain, apiConfig)
	}
}

// clearAPISpecificConfig clears configuration for a specific API
func clearAPISpecificConfig(config *Config, apiDomain string) error {
	apiConfig, exists := config.APIs[apiDomain]
	if !exists {
		return fmt.Errorf("API '%s' not found. Use 'capi apis list' to see available APIs", apiDomain)
	}

	// Clear all configuration except the endpoint
	apiConfig.Token = ""
	apiConfig.RefreshToken = ""
	apiConfig.Username = ""
	apiConfig.Organization = ""
	apiConfig.OrganizationGUID = ""
	apiConfig.Space = ""
	apiConfig.SpaceGUID = ""
	apiConfig.SkipSSLValidation = false
	apiConfig.UAAEndpoint = ""
	apiConfig.UAAToken = ""
	apiConfig.UAARefreshToken = ""
	apiConfig.UAAClientID = ""
	apiConfig.UAAClientSecret = ""

	// Update the API config in the main config
	config.APIs[apiDomain] = apiConfig

	if err := saveConfigStruct(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return outputConfigUpdateResult("Cleared configuration for API", apiDomain, "", "")
}

// displayConfigTable displays configuration in a table format
func displayConfigTable(config *Config) error {
	// Global configuration table
	table := tablewriter.NewWriter(os.Stdout)
	table.Header("Property", "Value")

	if err := table.Append([]string{"Output", config.Output}); err != nil {
		return err
	}
	if err := table.Append([]string{"No Color", fmt.Sprintf("%v", config.NoColor)}); err != nil {
		return err
	}

	if config.CurrentAPI != "" {
		if err := table.Append([]string{"Current API", config.CurrentAPI}); err != nil {
			return err
		}
	}

	fmt.Println("Global Configuration:")
	if err := table.Render(); err != nil {
		return err
	}

	// APIs configuration table
	if len(config.APIs) > 0 {
		fmt.Println("\nConfigured APIs:")
		apiTable := tablewriter.NewWriter(os.Stdout)
		apiTable.Header("Domain", "Endpoint", "Username", "Organization", "Space", "Current", "UAA Endpoint")

		for domain, apiConfig := range config.APIs {
			current := ""
			if domain == config.CurrentAPI {
				current = "✓"
			}

			username := apiConfig.Username
			if username == "" {
				username = "-"
			}

			organization := apiConfig.Organization
			if organization == "" {
				organization = "-"
			}

			space := apiConfig.Space
			if space == "" {
				space = "-"
			}

			uaaEndpoint := apiConfig.UAAEndpoint
			if uaaEndpoint == "" {
				uaaEndpoint = "-"
			}

			if err := apiTable.Append([]string{domain, apiConfig.Endpoint, username, organization, space, current, uaaEndpoint}); err != nil {
				return err
			}
		}

		if err := apiTable.Render(); err != nil {
			return err
		}
	} else {
		fmt.Println("\nNo APIs configured. Use 'capi apis add' to add one.")
	}

	// Legacy configuration table (for migration purposes)
	if config.API != "" {
		fmt.Println("\nLegacy Configuration (will be migrated):")
		legacyTable := tablewriter.NewWriter(os.Stdout)
		legacyTable.Header("Property", "Value")

		if err := legacyTable.Append([]string{"API", config.API}); err != nil {
			return err
		}
		if config.Username != "" {
			if err := legacyTable.Append([]string{"Username", config.Username}); err != nil {
				return err
			}
		}
		if config.Organization != "" {
			if err := legacyTable.Append([]string{"Organization", config.Organization}); err != nil {
				return err
			}
		}
		if config.Space != "" {
			if err := legacyTable.Append([]string{"Space", config.Space}); err != nil {
				return err
			}
		}

		if err := legacyTable.Render(); err != nil {
			return err
		}
	}

	// Legacy targets table
	if len(config.Targets) > 0 {
		fmt.Println("\nLegacy Targets (deprecated):")
		targetsTable := tablewriter.NewWriter(os.Stdout)
		targetsTable.Header("Name", "API", "Username", "Organization", "Space")

		for name, target := range config.Targets {
			username := target.Username
			if username == "" {
				username = "-"
			}

			organization := target.Organization
			if organization == "" {
				organization = "-"
			}

			space := target.Space
			if space == "" {
				space = "-"
			}

			if err := targetsTable.Append([]string{name, target.API, username, organization, space}); err != nil {
				return err
			}
		}

		if err := targetsTable.Render(); err != nil {
			return err
		}
	}

	return nil
}

// displayAPISpecificConfigTable displays configuration for a specific API in table format
func displayAPISpecificConfigTable(config *Config, apiDomain string, apiConfig *APIConfig) error {
	current := ""
	if apiDomain == config.CurrentAPI {
		current = " (current)"
	}

	fmt.Printf("Configuration for API '%s'%s:\n", apiDomain, current)

	table := tablewriter.NewWriter(os.Stdout)
	table.Header("Property", "Value")

	if err := table.Append([]string{"Endpoint", apiConfig.Endpoint}); err != nil {
		return err
	}

	if apiConfig.Username != "" {
		if err := table.Append([]string{"Username", apiConfig.Username}); err != nil {
			return err
		}
	}

	if apiConfig.Organization != "" {
		if err := table.Append([]string{"Organization", apiConfig.Organization}); err != nil {
			return err
		}
	}

	if apiConfig.OrganizationGUID != "" {
		if err := table.Append([]string{"Org GUID", apiConfig.OrganizationGUID}); err != nil {
			return err
		}
	}

	if apiConfig.Space != "" {
		if err := table.Append([]string{"Space", apiConfig.Space}); err != nil {
			return err
		}
	}

	if apiConfig.SpaceGUID != "" {
		if err := table.Append([]string{"Space GUID", apiConfig.SpaceGUID}); err != nil {
			return err
		}
	}

	if apiConfig.SkipSSLValidation {
		if err := table.Append([]string{"Skip SSL", fmt.Sprintf("%v", apiConfig.SkipSSLValidation)}); err != nil {
			return err
		}
	}

	if apiConfig.UAAEndpoint != "" {
		if err := table.Append([]string{"UAA Endpoint", apiConfig.UAAEndpoint}); err != nil {
			return err
		}
	}

	if apiConfig.UAAClientID != "" {
		if err := table.Append([]string{"UAA Client ID", apiConfig.UAAClientID}); err != nil {
			return err
		}
	}

	// Note: tokens are not displayed for security
	if apiConfig.Token != "" {
		if err := table.Append([]string{"Token", "[REDACTED]"}); err != nil {
			return err
		}
	}

	if apiConfig.RefreshToken != "" {
		if err := table.Append([]string{"Refresh Token", "[REDACTED]"}); err != nil {
			return err
		}
	}

	if apiConfig.UAAToken != "" {
		if err := table.Append([]string{"UAA Token", "[REDACTED]"}); err != nil {
			return err
		}
	}

	if apiConfig.UAARefreshToken != "" {
		if err := table.Append([]string{"UAA Refresh Token", "[REDACTED]"}); err != nil {
			return err
		}
	}

	if err := table.Render(); err != nil {
		return err
	}

	return nil
}

// outputConfigUpdateResult outputs configuration update results in the requested format
func outputConfigUpdateResult(action, key, value, apiDomain string) error {
	output := viper.GetString("output")

	result := map[string]string{
		"action": action,
		"key":    key,
	}

	if value != "" {
		result["value"] = value
	}

	if apiDomain != "" {
		result["api_domain"] = apiDomain
	}

	switch output {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	case "yaml":
		encoder := yaml.NewEncoder(os.Stdout)
		return encoder.Encode(result)
	default:
		// Table format for update results
		table := tablewriter.NewWriter(os.Stdout)
		table.Header("Property", "Value")

		if err := table.Append([]string{"Action", action}); err != nil {
			return err
		}
		if err := table.Append([]string{"Key", key}); err != nil {
			return err
		}

		if value != "" {
			if err := table.Append([]string{"Value", value}); err != nil {
				return err
			}
		}

		if apiDomain != "" {
			if err := table.Append([]string{"API Domain", apiDomain}); err != nil {
				return err
			}
		}

		if err := table.Render(); err != nil {
			return err
		}
		return nil
	}
}
