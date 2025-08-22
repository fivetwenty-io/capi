package commands

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// NewAPIsCommand creates the apis command group
func NewAPIsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "apis",
		Aliases: []string{"api"},
		Short:   "Manage Cloud Foundry API endpoints",
		Long:    "Add, list, delete, and target Cloud Foundry API endpoints",
	}

	cmd.AddCommand(newAPIsAddCommand())
	cmd.AddCommand(newAPIsListCommand())
	cmd.AddCommand(newAPIsDeleteCommand())
	cmd.AddCommand(newAPIsTargetCommand())

	return cmd
}

func newAPIsAddCommand() *cobra.Command {
	var skipSSLValidation bool

	cmd := &cobra.Command{
		Use:   "add NAME ENDPOINT",
		Short: "Add a new Cloud Foundry API endpoint",
		Long:  "Add a new Cloud Foundry API endpoint to the configuration",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			endpoint := args[1]

			// Validate and normalize the endpoint
			normalizedEndpoint, err := normalizeEndpoint(endpoint)
			if err != nil {
				return fmt.Errorf("invalid endpoint: %w", err)
			}

			// Load current configuration
			config := loadConfig()

			// Initialize APIs map if it doesn't exist
			if config.APIs == nil {
				config.APIs = make(map[string]*APIConfig)
			}

			// Extract domain for use as key
			domain := extractDomainFromEndpoint(normalizedEndpoint)

			// Check if API already exists
			if _, exists := config.APIs[domain]; exists {
				return fmt.Errorf("API '%s' already exists for domain '%s'", name, domain)
			}

			// Create new API configuration
			apiConfig := &APIConfig{
				Endpoint:          normalizedEndpoint,
				SkipSSLValidation: skipSSLValidation,
			}

			// Add to configuration
			config.APIs[domain] = apiConfig

			// If this is the first API, make it current
			if config.CurrentAPI == "" {
				config.CurrentAPI = domain
				fmt.Printf("API '%s' (%s) added and set as current target\n", name, normalizedEndpoint)
			} else {
				fmt.Printf("API '%s' (%s) added\n", name, normalizedEndpoint)
			}

			// Save configuration
			if err := saveConfigStruct(config); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&skipSSLValidation, "skip-ssl-validation", false, "Skip SSL certificate validation")

	return cmd
}

func newAPIsListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all Cloud Foundry API endpoints",
		Long:  "Display all configured Cloud Foundry API endpoints",
		RunE: func(cmd *cobra.Command, args []string) error {
			config := loadConfig()

			if len(config.APIs) == 0 {
				fmt.Println("No APIs configured. Use 'capi apis add' to add one.")
				return nil
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				type APIInfo struct {
					Domain            string `json:"domain"`
					Endpoint          string `json:"endpoint"`
					Username          string `json:"username,omitempty"`
					Organization      string `json:"organization,omitempty"`
					Space             string `json:"space,omitempty"`
					SkipSSLValidation bool   `json:"skip_ssl_validation"`
					Current           bool   `json:"current"`
				}

				var apis []APIInfo
				for domain, apiConfig := range config.APIs {
					apis = append(apis, APIInfo{
						Domain:            domain,
						Endpoint:          apiConfig.Endpoint,
						Username:          apiConfig.Username,
						Organization:      apiConfig.Organization,
						Space:             apiConfig.Space,
						SkipSSLValidation: apiConfig.SkipSSLValidation,
						Current:           domain == config.CurrentAPI,
					})
				}

				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(apis)

			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(config.APIs)

			default:
				fmt.Println("Configured APIs:")
				for domain, apiConfig := range config.APIs {
					current := ""
					if domain == config.CurrentAPI {
						current = " (current)"
					}
					fmt.Printf("  %s%s\n", domain, current)
					fmt.Printf("    Endpoint: %s\n", apiConfig.Endpoint)
					if apiConfig.Username != "" {
						fmt.Printf("    User:     %s\n", apiConfig.Username)
					}
					if apiConfig.Organization != "" {
						fmt.Printf("    Org:      %s\n", apiConfig.Organization)
					}
					if apiConfig.Space != "" {
						fmt.Printf("    Space:    %s\n", apiConfig.Space)
					}
					if apiConfig.SkipSSLValidation {
						fmt.Printf("    Skip SSL: %v\n", apiConfig.SkipSSLValidation)
					}
					fmt.Println()
				}
			}

			return nil
		},
	}
}

func newAPIsDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete DOMAIN",
		Short: "Delete a Cloud Foundry API endpoint",
		Long:  "Remove a Cloud Foundry API endpoint from the configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			domain := args[0]

			// Load current configuration
			config := loadConfig()

			// Check if API exists
			if _, exists := config.APIs[domain]; !exists {
				return fmt.Errorf("API '%s' not found", domain)
			}

			// Don't allow deleting the last API if it's current
			if len(config.APIs) == 1 && config.CurrentAPI == domain {
				return fmt.Errorf("cannot delete the only configured API")
			}

			// Remove from configuration
			delete(config.APIs, domain)

			// If this was the current API, switch to another one
			if config.CurrentAPI == domain {
				if len(config.APIs) > 0 {
					// Set the first remaining API as current
					for newDomain := range config.APIs {
						config.CurrentAPI = newDomain
						break
					}
					fmt.Printf("API '%s' deleted. Current API switched to '%s'\n", domain, config.CurrentAPI)
				} else {
					config.CurrentAPI = ""
					fmt.Printf("API '%s' deleted. No APIs remaining.\n", domain)
				}
			} else {
				fmt.Printf("API '%s' deleted\n", domain)
			}

			// Save configuration
			if err := saveConfigStruct(config); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}

			return nil
		},
	}
}

func newAPIsTargetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "target DOMAIN",
		Short: "Target a Cloud Foundry API endpoint",
		Long:  "Set a Cloud Foundry API endpoint as the current target",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			domain := args[0]

			// Load current configuration
			config := loadConfig()

			// Check if API exists
			if _, exists := config.APIs[domain]; !exists {
				return fmt.Errorf("API '%s' not found. Use 'capi apis list' to see available APIs", domain)
			}

			// Set as current
			config.CurrentAPI = domain

			// Save configuration
			if err := saveConfigStruct(config); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}

			fmt.Printf("API '%s' is now the current target\n", domain)
			return nil
		},
	}
}

// normalizeEndpoint validates and normalizes a CF API endpoint URL
func normalizeEndpoint(endpoint string) (string, error) {
	// Add https:// if no protocol is specified
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "https://" + endpoint
	}

	// Parse URL to validate
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("invalid URL format: %w", err)
	}

	// Ensure we have a host
	if parsedURL.Host == "" {
		return "", fmt.Errorf("no host specified in URL")
	}

	// Remove trailing slash and path for consistency
	normalizedURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)

	return normalizedURL, nil
}
