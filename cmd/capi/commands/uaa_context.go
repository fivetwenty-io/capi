package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// createUsersContextCommand creates the UAA context command
func createUsersContextCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "context",
		Aliases: []string{"ctx", "status"},
		Short:   "Display current UAA context",
		Long:    "Show information about the currently active UAA context including endpoint, authentication status, and user information",
		Example: `  # Show current UAA context
  capi uaa context

  # Show context in JSON format
  capi uaa context --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			config := loadConfig()
			output := viper.GetString("output")

			// Create UAA client to get context information
			uaaClient, err := NewUAAClient(config)
			if err != nil {
				// Show basic context even if UAA client creation fails
				if output == "json" {
					return showContextJSON(config, err)
				} else if output == "yaml" {
					return showContextYAML(config, err)
				} else {
					return showContextTable(config, err)
				}
			}

			// Test connection and get server info with caching
			ctx := context.Background()
			var serverInfo map[string]interface{}
			var connectionError error

			if err := uaaClient.TestConnection(ctx); err != nil {
				connectionError = err
			} else {
				_ = WithPerformanceTracking("get-server-info", func() error {
					var infoErr error
					serverInfo, infoErr = CachedServerInfo(uaaClient)
					return infoErr
				})
			}

			// Display context based on output format
			switch output {
			case "json":
				return showContextWithServerInfoJSON(config, uaaClient, serverInfo, connectionError)
			case "yaml":
				return showContextWithServerInfoYAML(config, uaaClient, serverInfo, connectionError)
			default:
				return showContextWithServerInfoTable(config, uaaClient, serverInfo, connectionError)
			}
		},
	}
}

// createUsersTargetCommand creates the UAA target command
func createUsersTargetCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "target <url>",
		Aliases: []string{"t", "set-target"},
		Short:   "Set UAA endpoint URL",
		Long:    "Set the URL of the UAA service to target for user management operations",
		Example: `  # Target UAA with full URL
  capi uaa target https://uaa.your-domain.com

  # Target UAA with hostname (https:// added automatically)
  capi uaa target uaa.your-domain.com`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetURL := args[0]

			// Validate URL format
			if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
				targetURL = "https://" + targetURL
			}

			// Update configuration
			viper.Set("uaa_endpoint", targetURL)
			config := loadConfig()
			config.UAAEndpoint = targetURL

			// Test connection to the new endpoint
			uaaClient, err := NewUAAClientWithEndpoint(targetURL, config)
			if err != nil {
				return fmt.Errorf("failed to create UAA client: %w", err)
			}

			ctx := context.Background()
			if err := uaaClient.TestConnection(ctx); err != nil {
				fmt.Printf("Warning: Failed to connect to UAA at %s: %v\n", targetURL, err)
				fmt.Println("Target set but connection test failed. Please verify the URL and network connectivity.")
			} else {
				fmt.Printf("Successfully targeted UAA at %s\n", targetURL)
			}

			// Save configuration
			if err := saveConfigStruct(config); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}

			return nil
		},
	}
}

// createUsersInfoCommand creates the UAA info command
func createUsersInfoCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Display UAA server information",
		Long:  "Show version and configuration information for the targeted UAA server",
		RunE: func(cmd *cobra.Command, args []string) error {
			config := loadConfig()

			if config.UAAEndpoint == "" {
				return fmt.Errorf("no UAA endpoint configured. Use 'capi uaa target <url>' to set one")
			}

			uaaClient, err := NewUAAClient(config)
			if err != nil {
				return fmt.Errorf("failed to create UAA client: %w", err)
			}

			ctx := context.Background()
			serverInfo, err := uaaClient.GetServerInfo(ctx)
			if err != nil {
				return fmt.Errorf("failed to get server info: %w", err)
			}

			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(serverInfo)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(serverInfo)
			default:
				return displayServerInfoTable(serverInfo)
			}
		},
	}
}

// createUsersVersionCommand creates the UAA version command
func createUsersVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Display UAA server version",
		Long:  "Show the version of the targeted UAA server",
		RunE: func(cmd *cobra.Command, args []string) error {
			config := loadConfig()

			if config.UAAEndpoint == "" {
				return fmt.Errorf("no UAA endpoint configured. Use 'capi uaa target <url>' to set one")
			}

			uaaClient, err := NewUAAClient(config)
			if err != nil {
				return fmt.Errorf("failed to create UAA client: %w", err)
			}

			ctx := context.Background()
			serverInfo, err := uaaClient.GetServerInfo(ctx)
			if err != nil {
				return fmt.Errorf("failed to get server info: %w", err)
			}

			// Extract version information
			version := "unknown"
			if app, ok := serverInfo["app"].(map[string]interface{}); ok {
				if v, ok := app["version"].(string); ok && v != "" {
					version = v
				}
			}

			output := viper.GetString("output")
			switch output {
			case "json":
				result := map[string]string{
					"version":  version,
					"endpoint": config.UAAEndpoint,
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(result)
			case "yaml":
				result := map[string]string{
					"version":  version,
					"endpoint": config.UAAEndpoint,
				}
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(result)
			default:
				fmt.Printf("UAA Version: %s\n", version)
				fmt.Printf("Endpoint: %s\n", config.UAAEndpoint)
				return nil
			}
		},
	}
}

// Helper functions for context display

func showContextJSON(config *Config, clientError error) error {
	context := map[string]interface{}{
		"uaa_endpoint":  config.UAAEndpoint,
		"authenticated": config.UAAToken != "" || config.Token != "",
		"client_error":  nil,
	}

	if clientError != nil {
		context["client_error"] = clientError.Error()
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(context)
}

func showContextYAML(config *Config, clientError error) error {
	context := map[string]interface{}{
		"uaa_endpoint":  config.UAAEndpoint,
		"authenticated": config.UAAToken != "" || config.Token != "",
		"client_error":  nil,
	}

	if clientError != nil {
		context["client_error"] = clientError.Error()
	}

	encoder := yaml.NewEncoder(os.Stdout)
	return encoder.Encode(context)
}

func showContextTable(config *Config, clientError error) error {
	fmt.Println("UAA Context:")
	fmt.Printf("  Endpoint:       %s\n", getValueOrEmpty(config.UAAEndpoint))
	fmt.Printf("  Authenticated:  %v\n", config.UAAToken != "" || config.Token != "")

	if clientError != nil {
		fmt.Printf("  Error:          %s\n", clientError.Error())
	}

	return nil
}

func showContextWithServerInfoJSON(config *Config, client *UAAClientWrapper, serverInfo map[string]interface{}, connectionError error) error {
	context := map[string]interface{}{
		"uaa_endpoint":      config.UAAEndpoint,
		"authenticated":     client.IsAuthenticated(),
		"connection_status": connectionError == nil,
		"server_info":       serverInfo,
	}

	if connectionError != nil {
		context["connection_error"] = connectionError.Error()
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(context)
}

func showContextWithServerInfoYAML(config *Config, client *UAAClientWrapper, serverInfo map[string]interface{}, connectionError error) error {
	context := map[string]interface{}{
		"uaa_endpoint":      config.UAAEndpoint,
		"authenticated":     client.IsAuthenticated(),
		"connection_status": connectionError == nil,
		"server_info":       serverInfo,
	}

	if connectionError != nil {
		context["connection_error"] = connectionError.Error()
	}

	encoder := yaml.NewEncoder(os.Stdout)
	return encoder.Encode(context)
}

func showContextWithServerInfoTable(config *Config, client *UAAClientWrapper, serverInfo map[string]interface{}, connectionError error) error {
	fmt.Println("UAA Context:")
	fmt.Printf("  Endpoint:       %s\n", getValueOrEmpty(config.UAAEndpoint))
	fmt.Printf("  Authenticated:  %v\n", client.IsAuthenticated())
	fmt.Printf("  Connected:      %v\n", connectionError == nil)

	if connectionError != nil {
		fmt.Printf("  Error:          %s\n", connectionError.Error())
	}

	if len(serverInfo) > 0 {
		fmt.Println("\nServer Information:")
		if app, ok := serverInfo["app"].(map[string]interface{}); ok {
			if name, ok := app["name"].(string); ok {
				fmt.Printf("  Name:           %s\n", name)
			}
			if version, ok := app["version"].(string); ok {
				fmt.Printf("  Version:        %s\n", version)
			}
		}
		if commitID, ok := serverInfo["commit_id"].(string); ok && commitID != "" {
			fmt.Printf("  Commit ID:      %s\n", commitID)
		}
		if zoneName, ok := serverInfo["zone_name"].(string); ok && zoneName != "" {
			fmt.Printf("  Zone Name:      %s\n", zoneName)
		}
	}

	return nil
}

func displayServerInfoTable(serverInfo map[string]interface{}) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header("Property", "Value")

	// Add server info to table
	for key, value := range serverInfo {
		if valueStr, ok := value.(string); ok && valueStr != "" {
			_ = table.Append(key, valueStr)
		} else if valueMap, ok := value.(map[string]interface{}); ok {
			// Handle nested objects like "app"
			for nestedKey, nestedValue := range valueMap {
				if nestedStr, ok := nestedValue.(string); ok && nestedStr != "" {
					_ = table.Append(fmt.Sprintf("%s.%s", key, nestedKey), nestedStr)
				}
			}
		}
	}

	_ = table.Render()
	return nil
}

func getValueOrEmpty(value string) string {
	if value == "" {
		return "(not set)"
	}
	return value
}
