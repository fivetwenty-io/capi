package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fivetwenty-io/capi-client/pkg/capi"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// validateFilePath validates that a file path is safe to read
func validateFilePathApps(filePath string) error {
	// Clean the path to resolve any path traversal attempts
	cleanPath := filepath.Clean(filePath)

	// Check for path traversal attempts
	if filepath.IsAbs(filePath) {
		// Allow absolute paths but ensure they're clean
		if cleanPath != filePath {
			return fmt.Errorf("invalid file path: potential path traversal attempt")
		}
	} else {
		// For relative paths, ensure they don't escape the current directory
		if len(cleanPath) > 0 && cleanPath[0] == '.' && len(cleanPath) > 1 && cleanPath[1] == '.' {
			return fmt.Errorf("invalid file path: path traversal not allowed")
		}
	}

	// Check if file exists and is readable
	if _, err := os.Stat(cleanPath); err != nil {
		return fmt.Errorf("file not accessible: %w", err)
	}

	return nil
}

// NewAppsCommand creates the apps command group
func NewAppsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "apps",
		Aliases: []string{"app", "applications"},
		Short:   "Manage applications",
		Long:    "List, create, and manage Cloud Foundry applications",
	}

	cmd.AddCommand(newAppsListCommand())
	cmd.AddCommand(newAppsStartCommand())
	cmd.AddCommand(newAppsStopCommand())
	cmd.AddCommand(newAppsRestartCommand())
	cmd.AddCommand(newAppsRestageCommand())
	cmd.AddCommand(newAppsScaleCommand())
	cmd.AddCommand(newAppsEnvCommand())
	cmd.AddCommand(newAppsSetEnvCommand())
	cmd.AddCommand(newAppsUnsetEnvCommand())
	cmd.AddCommand(newAppsLogsCommand())
	cmd.AddCommand(newAppsSSHCommand())
	cmd.AddCommand(newAppsProcessesCommand())
	cmd.AddCommand(newAppsManifestCommand())
	cmd.AddCommand(newAppsStatsCommand())
	cmd.AddCommand(newAppsEventsCommand())
	cmd.AddCommand(newAppsHealthCheckCommand())

	return cmd
}

func newAppsListCommand() *cobra.Command {
	var spaceName string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List applications",
		Long:  "List all applications the user has access to",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			params := capi.NewQueryParams()

			// Filter by space if specified
			if spaceName != "" {
				// Find space by name
				spaceParams := capi.NewQueryParams()
				spaceParams.WithFilter("names", spaceName)

				// Add org filter if targeted
				if orgGUID := viper.GetString("organization_guid"); orgGUID != "" {
					spaceParams.WithFilter("organization_guids", orgGUID)
				}

				spaces, err := client.Spaces().List(ctx, spaceParams)
				if err != nil {
					return fmt.Errorf("failed to find space: %w", err)
				}
				if len(spaces.Resources) == 0 {
					return fmt.Errorf("space '%s' not found", spaceName)
				}

				params.WithFilter("space_guids", spaces.Resources[0].GUID)
			} else if spaceGUID := viper.GetString("space_guid"); spaceGUID != "" {
				// Use targeted space
				params.WithFilter("space_guids", spaceGUID)
			}

			apps, err := client.Apps().List(ctx, params)
			if err != nil {
				return fmt.Errorf("failed to list applications: %w", err)
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(apps.Resources)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(apps.Resources)
			default:
				if len(apps.Resources) == 0 {
					fmt.Println("No applications found")
					return nil
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Name", "GUID", "State", "Lifecycle", "Buildpacks", "Stack", "Created", "Updated")

				for _, app := range apps.Resources {
					lifecycle := "buildpack"
					if app.Lifecycle.Type == "docker" {
						lifecycle = "docker"
					}

					buildpacks := ""
					if app.Lifecycle.Data != nil {
						if bps, ok := app.Lifecycle.Data["buildpacks"].([]interface{}); ok {
							var bpStrs []string
							for _, bp := range bps {
								if bpStr, ok := bp.(string); ok {
									bpStrs = append(bpStrs, bpStr)
								}
							}
							buildpacks = strings.Join(bpStrs, ", ")
						}
					}

					stack := ""
					if app.Lifecycle.Data != nil {
						if s, ok := app.Lifecycle.Data["stack"].(string); ok {
							stack = s
						}
					}

					created := ""
					if !app.CreatedAt.IsZero() {
						created = app.CreatedAt.Format("2006-01-02 15:04:05")
					}

					updated := ""
					if !app.UpdatedAt.IsZero() {
						updated = app.UpdatedAt.Format("2006-01-02 15:04:05")
					}

					_ = table.Append(app.Name, app.GUID, app.State, lifecycle, buildpacks, stack, created, updated)
				}

				_ = table.Render()
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&spaceName, "space", "s", "", "filter by space name")

	return cmd
}

func newAppsStartCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "start APP_NAME_OR_GUID",
		Short: "Start an application",
		Long:  "Start a Cloud Foundry application",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find application
			appGUID, appName, err := resolveApp(ctx, client, nameOrGUID)
			if err != nil {
				return err
			}

			// Start application
			app, err := client.Apps().Start(ctx, appGUID)
			if err != nil {
				return fmt.Errorf("failed to start application: %w", err)
			}

			fmt.Printf("Successfully started application '%s'\n", app.Name)
			_ = appName // Use appName if needed
			return nil
		},
	}
}

func newAppsStopCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stop APP_NAME_OR_GUID",
		Short: "Stop an application",
		Long:  "Stop a Cloud Foundry application",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find application
			appGUID, _, err := resolveApp(ctx, client, nameOrGUID)
			if err != nil {
				return err
			}

			// Stop application
			app, err := client.Apps().Stop(ctx, appGUID)
			if err != nil {
				return fmt.Errorf("failed to stop application: %w", err)
			}

			fmt.Printf("Successfully stopped application '%s'\n", app.Name)
			return nil
		},
	}
}

func newAppsRestartCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "restart APP_NAME_OR_GUID",
		Short: "Restart an application",
		Long:  "Restart a Cloud Foundry application",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find application
			appGUID, _, err := resolveApp(ctx, client, nameOrGUID)
			if err != nil {
				return err
			}

			// Restart application
			app, err := client.Apps().Restart(ctx, appGUID)
			if err != nil {
				return fmt.Errorf("failed to restart application: %w", err)
			}

			fmt.Printf("Successfully restarted application '%s'\n", app.Name)
			return nil
		},
	}
}

func newAppsRestageCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "restage APP_NAME_OR_GUID",
		Short: "Restage an application",
		Long:  "Restage a Cloud Foundry application to create a new build",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find application
			appGUID, appName, err := resolveApp(ctx, client, nameOrGUID)
			if err != nil {
				return err
			}

			// Restage application
			build, err := client.Apps().Restage(ctx, appGUID)
			if err != nil {
				return fmt.Errorf("failed to restage application: %w", err)
			}

			fmt.Printf("Successfully initiated restage of application '%s'\n", appName)
			fmt.Printf("Build GUID: %s\n", build.GUID)
			fmt.Printf("Build State: %s\n", build.State)
			return nil
		},
	}
}

// Helper function to resolve app name or GUID
func resolveApp(ctx context.Context, client capi.Client, nameOrGUID string) (guid string, name string, err error) {
	// Try to get by GUID first
	app, err := client.Apps().Get(ctx, nameOrGUID)
	if err == nil {
		return app.GUID, app.Name, nil
	}

	// Try by name
	params := capi.NewQueryParams()
	params.WithFilter("names", nameOrGUID)

	// Add space filter if targeted
	if spaceGUID := viper.GetString("space_guid"); spaceGUID != "" {
		params.WithFilter("space_guids", spaceGUID)
	}

	apps, err := client.Apps().List(ctx, params)
	if err != nil {
		return "", "", fmt.Errorf("failed to find application: %w", err)
	}
	if len(apps.Resources) == 0 {
		return "", "", fmt.Errorf("application '%s' not found", nameOrGUID)
	}

	return apps.Resources[0].GUID, apps.Resources[0].Name, nil
}

func newAppsScaleCommand() *cobra.Command {
	var instances int
	var memory int
	var disk int
	var processType string

	cmd := &cobra.Command{
		Use:   "scale APP_NAME_OR_GUID",
		Short: "Scale an application",
		Long:  "Scale a Cloud Foundry application instances, memory, or disk",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find application
			appGUID, appName, err := resolveApp(ctx, client, nameOrGUID)
			if err != nil {
				return err
			}

			// Get processes for the app
			processParams := capi.NewQueryParams().WithFilter("app_guids", appGUID)
			processes, err := client.Processes().List(ctx, processParams)
			if err != nil {
				return fmt.Errorf("failed to list processes: %w", err)
			}

			if len(processes.Resources) == 0 {
				return fmt.Errorf("no processes found for application '%s'", appName)
			}

			// Find the target process (default to first process, usually 'web')
			var targetProcess *capi.Process
			for _, process := range processes.Resources {
				if processType == "" || process.Type == processType {
					targetProcess = &process
					break
				}
			}

			if targetProcess == nil {
				return fmt.Errorf("process type '%s' not found for application '%s'", processType, appName)
			}

			// Build scale request with only the flags that were set
			scaleReq := &capi.ProcessScaleRequest{}

			if cmd.Flags().Changed("instances") {
				scaleReq.Instances = &instances
			}
			if cmd.Flags().Changed("memory") {
				scaleReq.MemoryInMB = &memory
			}
			if cmd.Flags().Changed("disk") {
				scaleReq.DiskInMB = &disk
			}

			// Check if any scaling parameters were provided
			if scaleReq.Instances == nil && scaleReq.MemoryInMB == nil && scaleReq.DiskInMB == nil {
				// No scaling parameters provided, show current scale
				fmt.Printf("Application '%s' process '%s' current scale:\n", appName, targetProcess.Type)
				fmt.Printf("  Instances: %d\n", targetProcess.Instances)
				fmt.Printf("  Memory: %d MB\n", targetProcess.MemoryInMB)
				fmt.Printf("  Disk: %d MB\n", targetProcess.DiskInMB)
				return nil
			}

			// Scale the process
			scaledProcess, err := client.Processes().Scale(ctx, targetProcess.GUID, scaleReq)
			if err != nil {
				return fmt.Errorf("failed to scale application: %w", err)
			}

			fmt.Printf("Successfully scaled application '%s' process '%s':\n", appName, scaledProcess.Type)
			fmt.Printf("  Instances: %d\n", scaledProcess.Instances)
			fmt.Printf("  Memory: %d MB\n", scaledProcess.MemoryInMB)
			fmt.Printf("  Disk: %d MB\n", scaledProcess.DiskInMB)

			return nil
		},
	}

	cmd.Flags().IntVarP(&instances, "instances", "i", 0, "Number of instances")
	cmd.Flags().IntVarP(&memory, "memory", "m", 0, "Memory in MB")
	cmd.Flags().IntVarP(&disk, "disk", "d", 0, "Disk in MB")
	cmd.Flags().StringVarP(&processType, "process", "p", "", "Process type (defaults to first process)")

	return cmd
}

// flattenJSON recursively flattens a JSON object into a map with dot-separated keys
func flattenJSON(obj interface{}, prefix string) map[string]interface{} {
	result := make(map[string]interface{})

	switch v := obj.(type) {
	case map[string]interface{}:
		for key, value := range v {
			fullKey := key
			if prefix != "" {
				fullKey = prefix + "." + key
			}

			// Recursively flatten nested objects
			if nested := flattenJSON(value, fullKey); len(nested) > 0 {
				for k, v := range nested {
					result[k] = v
				}
			} else {
				result[fullKey] = value
			}
		}
	case []interface{}:
		for i, item := range v {
			fullKey := fmt.Sprintf("%s[%d]", prefix, i)
			if nested := flattenJSON(item, fullKey); len(nested) > 0 {
				for k, v := range nested {
					result[k] = v
				}
			} else {
				result[fullKey] = item
			}
		}
	default:
		if prefix != "" {
			result[prefix] = v
		}
	}

	return result
}

func newAppsEnvCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "env APP_NAME_OR_GUID",
		Short: "Show application environment variables",
		Long:  "Display all environment variables for a Cloud Foundry application",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find application
			appGUID, appName, err := resolveApp(ctx, client, nameOrGUID)
			if err != nil {
				return err
			}

			// Get environment variables
			env, err := client.Apps().GetEnv(ctx, appGUID)
			if err != nil {
				return fmt.Errorf("failed to get environment variables: %w", err)
			}

			// Collect environment variables for structured output
			type EnvVar struct {
				Name   string      `json:"name" yaml:"name"`
				Value  interface{} `json:"value" yaml:"value"`
				Source string      `json:"source" yaml:"source"`
			}

			var envVars []EnvVar
			var vcapServices interface{}
			var vcapApplication interface{}

			// Collect user-provided environment variables
			for key, value := range env.EnvironmentVariables {
				envVars = append(envVars, EnvVar{
					Name:   key,
					Value:  value,
					Source: "user-provided",
				})
			}

			// Collect system environment variables and extract VCAP_SERVICES
			for key, value := range env.SystemEnvJSON {
				if key == "VCAP_SERVICES" {
					vcapServices = value
				} else {
					envVars = append(envVars, EnvVar{
						Name:   key,
						Value:  value,
						Source: "system",
					})
				}
			}

			// Collect staging environment variables
			for key, value := range env.StagingEnvJSON {
				envVars = append(envVars, EnvVar{
					Name:   key,
					Value:  value,
					Source: "staging",
				})
			}

			// Collect running environment variables
			for key, value := range env.RunningEnvJSON {
				envVars = append(envVars, EnvVar{
					Name:   key,
					Value:  value,
					Source: "running",
				})
			}

			// Collect application environment variables and extract VCAP_APPLICATION
			for key, value := range env.ApplicationEnvJSON {
				if key == "VCAP_APPLICATION" {
					vcapApplication = value
				} else {
					envVars = append(envVars, EnvVar{
						Name:   key,
						Value:  value,
						Source: "application",
					})
				}
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				result := map[string]interface{}{
					"environment_variables": envVars,
					"vcap_services":         vcapServices,
					"vcap_application":      vcapApplication,
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(result)
			case "yaml":
				result := map[string]interface{}{
					"environment_variables": envVars,
					"vcap_services":         vcapServices,
					"vcap_application":      vcapApplication,
				}
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(result)
			default:
				fmt.Printf("Environment variables for application '%s':\n\n", appName)

				// Main environment variables table
				if len(envVars) > 0 {
					table := tablewriter.NewWriter(os.Stdout)
					table.Header("Name", "Value", "Source")

					for _, envVar := range envVars {
						valueStr := fmt.Sprintf("%v", envVar.Value)
						// Truncate long values for table display
						if len(valueStr) > 80 {
							valueStr = valueStr[:77] + "..."
						}
						_ = table.Append(envVar.Name, valueStr, envVar.Source)
					}

					_ = table.Render()
					fmt.Println()
				}

				// VCAP_SERVICES table
				if vcapServices != nil {
					vcapServicesFlattened := flattenJSON(vcapServices, "")
					if len(vcapServicesFlattened) > 0 {
						fmt.Println("VCAP_SERVICES:")
						fmt.Println()
						servicesTable := tablewriter.NewWriter(os.Stdout)
						servicesTable.Header("Key", "Value")

						for key, value := range vcapServicesFlattened {
							valueStr := fmt.Sprintf("%v", value)
							_ = servicesTable.Append(key, valueStr)
						}

						_ = servicesTable.Render()
						fmt.Println()
					}
				}

				// VCAP_APPLICATION table
				if vcapApplication != nil {
					vcapApplicationFlattened := flattenJSON(vcapApplication, "")
					if len(vcapApplicationFlattened) > 0 {
						fmt.Println("VCAP_APPLICATION:")
						fmt.Println()
						appTable := tablewriter.NewWriter(os.Stdout)
						appTable.Header("Key", "Value")

						for key, value := range vcapApplicationFlattened {
							valueStr := fmt.Sprintf("%v", value)
							_ = appTable.Append(key, valueStr)
						}

						_ = appTable.Render()
					}
				}

				if len(envVars) == 0 && vcapServices == nil && vcapApplication == nil {
					fmt.Printf("No environment variables found for application '%s'\n", appName)
				}
			}

			return nil
		},
	}
}

func newAppsSetEnvCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set-env APP_NAME_OR_GUID ENV_VAR_NAME ENV_VAR_VALUE",
		Short: "Set an environment variable for an application",
		Long:  "Set a user-provided environment variable for a Cloud Foundry application",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]
			envVarName := args[1]
			envVarValue := args[2]

			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find application
			appGUID, appName, err := resolveApp(ctx, client, nameOrGUID)
			if err != nil {
				return err
			}

			// Get current environment variables
			currentEnvVars, err := client.Apps().GetEnvVars(ctx, appGUID)
			if err != nil {
				return fmt.Errorf("failed to get current environment variables: %w", err)
			}

			// Update environment variables with the new one
			if currentEnvVars == nil {
				currentEnvVars = make(map[string]interface{})
			}
			currentEnvVars[envVarName] = envVarValue

			// Update environment variables
			_, err = client.Apps().UpdateEnvVars(ctx, appGUID, currentEnvVars)
			if err != nil {
				return fmt.Errorf("failed to set environment variable: %w", err)
			}

			fmt.Printf("Successfully set environment variable '%s' for application '%s'\n", envVarName, appName)
			fmt.Printf("  %s=%s\n", envVarName, envVarValue)
			fmt.Println("\nNote: You may need to restart the application for the changes to take effect.")

			return nil
		},
	}
}

func newAppsUnsetEnvCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "unset-env APP_NAME_OR_GUID ENV_VAR_NAME",
		Short: "Unset an environment variable for an application",
		Long:  "Remove a user-provided environment variable from a Cloud Foundry application",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]
			envVarName := args[1]

			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find application
			appGUID, appName, err := resolveApp(ctx, client, nameOrGUID)
			if err != nil {
				return err
			}

			// Get current environment variables
			currentEnvVars, err := client.Apps().GetEnvVars(ctx, appGUID)
			if err != nil {
				return fmt.Errorf("failed to get current environment variables: %w", err)
			}

			// Check if the environment variable exists
			if currentEnvVars == nil || currentEnvVars[envVarName] == nil {
				return fmt.Errorf("environment variable '%s' not found for application '%s'", envVarName, appName)
			}

			// Remove the environment variable
			delete(currentEnvVars, envVarName)

			// Update environment variables
			_, err = client.Apps().UpdateEnvVars(ctx, appGUID, currentEnvVars)
			if err != nil {
				return fmt.Errorf("failed to unset environment variable: %w", err)
			}

			fmt.Printf("Successfully unset environment variable '%s' for application '%s'\n", envVarName, appName)
			fmt.Println("\nNote: You may need to restart the application for the changes to take effect.")

			return nil
		},
	}
}

func newAppsLogsCommand() *cobra.Command {
	var follow bool
	var recent bool
	var numLines int

	cmd := &cobra.Command{
		Use:   "logs APP_NAME_OR_GUID",
		Short: "Show application logs",
		Long:  "Display recent logs or stream logs for a Cloud Foundry application",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find application
			appGUID, appName, err := resolveApp(ctx, client, nameOrGUID)
			if err != nil {
				return err
			}

			if recent || !follow {
				// Get recent logs using the logs API
				logs, err := client.Apps().GetRecentLogs(ctx, appGUID, numLines)
				if err != nil {
					return fmt.Errorf("failed to get recent logs: %w", err)
				}

				fmt.Printf("Recent logs for application '%s':\n\n", appName)
				for _, logMsg := range logs.Messages {
					timestamp := logMsg.Timestamp.Format("2006-01-02T15:04:05.000Z")
					fmt.Printf("[%s/%s] %s %s %s\n",
						logMsg.SourceType, logMsg.SourceID, logMsg.MessageType, timestamp, logMsg.Message)
				}

				fmt.Printf("\nNote: Logs streaming requires WebSocket/SSE connection to CF API.\n")
				fmt.Printf("Application GUID: %s\n", appGUID)
			}

			if follow {
				fmt.Printf("Streaming logs for application '%s'...\n", appName)
				fmt.Println("Press Ctrl+C to stop streaming.")

				// Create a context that can be cancelled with Ctrl+C
				ctx, cancel := context.WithCancel(ctx)
				defer cancel()

				// Start streaming logs
				logChan, err := client.Apps().StreamLogs(ctx, appGUID)
				if err != nil {
					return fmt.Errorf("failed to start log streaming: %w", err)
				}

				// Process streaming logs
				for logMsg := range logChan {
					timestamp := logMsg.Timestamp.Format("2006-01-02T15:04:05.000Z")
					fmt.Printf("[%s/%s] %s %s %s\n",
						logMsg.SourceType, logMsg.SourceID, logMsg.MessageType, timestamp, logMsg.Message)
				}

				fmt.Println("\nLog streaming stopped.")
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Stream logs continuously")
	cmd.Flags().BoolVarP(&recent, "recent", "r", false, "Show recent logs only")
	cmd.Flags().IntVarP(&numLines, "lines", "n", 50, "Number of recent log lines to show")

	return cmd
}

func newAppsSSHCommand() *cobra.Command {
	var index int
	var processType string
	var command string

	cmd := &cobra.Command{
		Use:   "ssh APP_NAME_OR_GUID",
		Short: "SSH into an application instance",
		Long:  "Open an SSH connection to a Cloud Foundry application instance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find application
			appGUID, appName, err := resolveApp(ctx, client, nameOrGUID)
			if err != nil {
				return err
			}

			// Get processes for the app to validate process type and index
			processParams := capi.NewQueryParams().WithFilter("app_guids", appGUID)
			processes, err := client.Processes().List(ctx, processParams)
			if err != nil {
				return fmt.Errorf("failed to list processes: %w", err)
			}

			if len(processes.Resources) == 0 {
				return fmt.Errorf("no processes found for application '%s'", appName)
			}

			// Find the target process (default to first process, usually 'web')
			var targetProcess *capi.Process
			for _, process := range processes.Resources {
				if processType == "" || process.Type == processType {
					targetProcess = &process
					break
				}
			}

			if targetProcess == nil {
				return fmt.Errorf("process type '%s' not found for application '%s'", processType, appName)
			}

			// Validate instance index
			if index >= targetProcess.Instances {
				return fmt.Errorf("instance index %d is out of range (0-%d) for process '%s'",
					index, targetProcess.Instances-1, targetProcess.Type)
			}

			// Check if SSH is enabled for the app (this would be a real API call)
			fmt.Printf("Connecting to application '%s' instance %d via SSH...\n", appName, index)
			fmt.Printf("Process: %s/%d\n", targetProcess.Type, index)

			if command != "" {
				fmt.Printf("Command: %s\n", command)
			}

			// Note: Real implementation would:
			// 1. Get SSH endpoint info from Cloud Controller
			// 2. Get one-time SSH code
			// 3. Establish SSH connection to Diego cell
			fmt.Println("\nNote: SSH connection requires:")
			fmt.Println("1. SSH to be enabled for the space and application")
			fmt.Println("2. Authentication with Diego SSH proxy")
			fmt.Println("3. Network access to Diego cells")
			fmt.Printf("4. App GUID: %s\n", appGUID)
			fmt.Printf("5. Process GUID: %s\n", targetProcess.GUID)

			return nil
		},
	}

	cmd.Flags().IntVarP(&index, "index", "i", 0, "Instance index to connect to")
	cmd.Flags().StringVarP(&processType, "process", "p", "", "Process type (defaults to first process)")
	cmd.Flags().StringVar(&command, "command", "", "Command to run in the SSH session")

	return cmd
}

func newAppsProcessesCommand() *cobra.Command {
	var showStats bool

	cmd := &cobra.Command{
		Use:   "processes APP_NAME_OR_GUID",
		Short: "List application processes",
		Long:  "List all processes for a Cloud Foundry application with their current status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find application
			appGUID, appName, err := resolveApp(ctx, client, nameOrGUID)
			if err != nil {
				return err
			}

			// Get processes for the app
			processParams := capi.NewQueryParams().WithFilter("app_guids", appGUID)
			processes, err := client.Processes().List(ctx, processParams)
			if err != nil {
				return fmt.Errorf("failed to list processes: %w", err)
			}

			if len(processes.Resources) == 0 {
				output := viper.GetString("output")
				switch output {
				case "json":
					encoder := json.NewEncoder(os.Stdout)
					encoder.SetIndent("", "  ")
					return encoder.Encode([]capi.Process{})
				case "yaml":
					encoder := yaml.NewEncoder(os.Stdout)
					return encoder.Encode([]capi.Process{})
				default:
					fmt.Printf("No processes found for application '%s'\n", appName)
				}
				return nil
			}

			// Collect process data with stats if requested
			var processData []map[string]interface{}
			for _, process := range processes.Resources {
				processInfo := map[string]interface{}{
					"type":      process.Type,
					"guid":      process.GUID,
					"instances": process.Instances,
					"memory_mb": process.MemoryInMB,
					"disk_mb":   process.DiskInMB,
				}

				if process.LogRateLimitInBytesPerSecond != nil {
					processInfo["log_rate_limit_bytes_per_sec"] = *process.LogRateLimitInBytesPerSecond
				} else {
					processInfo["log_rate_limit_bytes_per_sec"] = -1
				}

				if process.Command != nil {
					processInfo["command"] = *process.Command
				}

				if process.HealthCheck != nil {
					healthCheck := map[string]interface{}{
						"type": process.HealthCheck.Type,
					}
					if process.HealthCheck.Data != nil {
						if process.HealthCheck.Data.Timeout != nil {
							healthCheck["timeout"] = *process.HealthCheck.Data.Timeout
						}
						if process.HealthCheck.Data.Endpoint != nil {
							healthCheck["endpoint"] = *process.HealthCheck.Data.Endpoint
						}
					}
					processInfo["health_check"] = healthCheck
				}

				// Get process stats if requested
				if showStats {
					stats, err := client.Processes().GetStats(ctx, process.GUID)
					if err == nil && len(stats.Resources) > 0 {
						var instanceStats []map[string]interface{}
						for _, stat := range stats.Resources {
							statInfo := map[string]interface{}{
								"index": stat.Index,
								"state": stat.State,
							}
							if stat.Usage != nil {
								statInfo["usage"] = map[string]interface{}{
									"cpu_percent":  stat.Usage.CPU * 100,
									"memory_bytes": stat.Usage.Mem,
									"disk_bytes":   stat.Usage.Disk,
								}
							}
							instanceStats = append(instanceStats, statInfo)
						}
						processInfo["instance_stats"] = instanceStats
					}
				}

				processData = append(processData, processInfo)
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(processData)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(processData)
			default:
				fmt.Printf("Processes for application '%s':\n\n", appName)

				for _, process := range processes.Resources {
					fmt.Printf("Process: %s\n", process.Type)
					fmt.Printf("  GUID: %s\n", process.GUID)
					fmt.Printf("  Instances: %d\n", process.Instances)
					fmt.Printf("  Memory: %d MB\n", process.MemoryInMB)
					fmt.Printf("  Disk: %d MB\n", process.DiskInMB)

					if process.LogRateLimitInBytesPerSecond != nil {
						fmt.Printf("  Log Rate Limit: %d bytes/sec\n", *process.LogRateLimitInBytesPerSecond)
					} else {
						fmt.Printf("  Log Rate Limit: -1 bytes/sec\n")
					}

					if process.Command != nil {
						fmt.Printf("  Command: %s\n", *process.Command)
					}

					if process.HealthCheck != nil {
						fmt.Printf("  Health Check: %s", process.HealthCheck.Type)
						if process.HealthCheck.Data != nil {
							if process.HealthCheck.Data.Timeout != nil {
								fmt.Printf(" (timeout: %ds)", *process.HealthCheck.Data.Timeout)
							}
							if process.HealthCheck.Data.Endpoint != nil {
								fmt.Printf(" (endpoint: %s)", *process.HealthCheck.Data.Endpoint)
							}
						}
						fmt.Println()
					}

					// Get process stats if requested
					if showStats {
						stats, err := client.Processes().GetStats(ctx, process.GUID)
						if err == nil && len(stats.Resources) > 0 {
							fmt.Printf("  Instance Stats:\n")
							for _, stat := range stats.Resources {
								fmt.Printf("    Instance %d: %s", stat.Index, stat.State)
								if stat.Usage != nil {
									fmt.Printf(" (CPU: %.2f%%, Memory: %d bytes, Disk: %d bytes)",
										stat.Usage.CPU*100, stat.Usage.Mem, stat.Usage.Disk)
								}
								fmt.Println()
							}
						}
					}

					fmt.Println()
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&showStats, "stats", "s", false, "Show detailed instance statistics")

	return cmd
}

func newAppsManifestCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "manifest",
		Short: "Manage application manifests",
		Long:  "View, apply, or diff application manifests",
	}

	cmd.AddCommand(newAppsManifestGetCommand())
	cmd.AddCommand(newAppsManifestApplyCommand())
	cmd.AddCommand(newAppsManifestDiffCommand())

	return cmd
}

func newAppsManifestGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get APP_NAME_OR_GUID",
		Short: "Get application manifest",
		Long:  "Retrieve the current manifest for a Cloud Foundry application",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find application
			appGUID, appName, err := resolveApp(ctx, client, nameOrGUID)
			if err != nil {
				return err
			}

			// Get application manifest
			manifest, err := client.Apps().GetManifest(ctx, appGUID)
			if err != nil {
				return fmt.Errorf("failed to get manifest: %w", err)
			}

			// Prepare structured output
			type ManifestInfo struct {
				AppName      string `json:"app_name" yaml:"app_name"`
				AppGUID      string `json:"app_guid" yaml:"app_guid"`
				ManifestYAML string `json:"manifest_yaml" yaml:"manifest_yaml"`
			}

			manifestInfo := ManifestInfo{
				AppName:      appName,
				AppGUID:      appGUID,
				ManifestYAML: manifest,
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(manifestInfo)
			case "yaml":
				// For YAML output, we'll output the manifest directly since it's already YAML
				// But we'll include metadata as comments
				fmt.Printf("# Application: %s\n", appName)
				fmt.Printf("# GUID: %s\n", appGUID)
				fmt.Printf("# Manifest:\n")
				fmt.Print(manifest)
				return nil
			default:
				// For table output, we'll show metadata in a table and then the manifest
				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Property", "Value")
				_ = table.Append("Application Name", appName)
				_ = table.Append("Application GUID", appGUID)

				fmt.Printf("Manifest information:\n\n")
				_ = table.Render()
				fmt.Printf("\nManifest content:\n\n")
				fmt.Print(manifest)
			}

			return nil
		},
	}
}

func newAppsManifestApplyCommand() *cobra.Command {
	var manifestFile string

	cmd := &cobra.Command{
		Use:   "apply SPACE_NAME_OR_GUID [MANIFEST_FILE]",
		Short: "Apply application manifest",
		Long:  "Apply a manifest to deploy or update applications in a space",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			spaceIdentifier := args[0]

			var manifestPath string
			if len(args) == 2 {
				manifestPath = args[1]
			} else if manifestFile != "" {
				manifestPath = manifestFile
			} else {
				manifestPath = "manifest.yml"
			}

			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find space
			var spaceGUID string
			var spaceName string

			// Try as GUID first
			space, err := client.Spaces().Get(ctx, spaceIdentifier)
			if err == nil {
				spaceGUID = space.GUID
				spaceName = space.Name
			} else {
				// Try by name
				spaceParams := capi.NewQueryParams()
				spaceParams.WithFilter("names", spaceIdentifier)

				// Add org filter if targeted
				if orgGUID := viper.GetString("organization_guid"); orgGUID != "" {
					spaceParams.WithFilter("organization_guids", orgGUID)
				}

				spaces, err := client.Spaces().List(ctx, spaceParams)
				if err != nil {
					return fmt.Errorf("failed to find space: %w", err)
				}
				if len(spaces.Resources) == 0 {
					return fmt.Errorf("space '%s' not found", spaceIdentifier)
				}

				spaceGUID = spaces.Resources[0].GUID
				spaceName = spaces.Resources[0].Name
			}

			// Read manifest file
			// Validate file path to prevent directory traversal
			if err := validateFilePathApps(manifestPath); err != nil {
				return fmt.Errorf("invalid manifest file: %w", err)
			}
			manifestContent, err := os.ReadFile(filepath.Clean(manifestPath))
			if err != nil {
				return fmt.Errorf("failed to read manifest file '%s': %w", manifestPath, err)
			}

			// Apply manifest
			job, err := client.Spaces().ApplyManifest(ctx, spaceGUID, string(manifestContent))
			if err != nil {
				return fmt.Errorf("failed to apply manifest: %w", err)
			}

			fmt.Printf("Successfully applied manifest to space '%s'\n", spaceName)
			if job != nil {
				fmt.Printf("Job GUID: %s\n", job.GUID)
				fmt.Printf("Monitor job status with: capi jobs get %s\n", job.GUID)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&manifestFile, "file", "f", "", "path to manifest file (default: manifest.yml)")

	return cmd
}

func newAppsManifestDiffCommand() *cobra.Command {
	var manifestFile string

	cmd := &cobra.Command{
		Use:   "diff SPACE_NAME_OR_GUID [MANIFEST_FILE]",
		Short: "Show manifest diff",
		Long:  "Show the differences between the current state and a manifest",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			spaceIdentifier := args[0]

			var manifestPath string
			if len(args) == 2 {
				manifestPath = args[1]
			} else if manifestFile != "" {
				manifestPath = manifestFile
			} else {
				manifestPath = "manifest.yml"
			}

			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find space
			var spaceGUID string
			var spaceName string

			// Try as GUID first
			space, err := client.Spaces().Get(ctx, spaceIdentifier)
			if err == nil {
				spaceGUID = space.GUID
				spaceName = space.Name
			} else {
				// Try by name
				spaceParams := capi.NewQueryParams()
				spaceParams.WithFilter("names", spaceIdentifier)

				// Add org filter if targeted
				if orgGUID := viper.GetString("organization_guid"); orgGUID != "" {
					spaceParams.WithFilter("organization_guids", orgGUID)
				}

				spaces, err := client.Spaces().List(ctx, spaceParams)
				if err != nil {
					return fmt.Errorf("failed to find space: %w", err)
				}
				if len(spaces.Resources) == 0 {
					return fmt.Errorf("space '%s' not found", spaceIdentifier)
				}

				spaceGUID = spaces.Resources[0].GUID
				spaceName = spaces.Resources[0].Name
			}

			// Read manifest file
			// Validate file path to prevent directory traversal
			if err := validateFilePathApps(manifestPath); err != nil {
				return fmt.Errorf("invalid manifest file: %w", err)
			}
			manifestContent, err := os.ReadFile(filepath.Clean(manifestPath))
			if err != nil {
				return fmt.Errorf("failed to read manifest file '%s': %w", manifestPath, err)
			}

			// Create manifest diff
			diff, err := client.Spaces().CreateManifestDiff(ctx, spaceGUID, string(manifestContent))
			if err != nil {
				return fmt.Errorf("failed to create manifest diff: %w", err)
			}

			fmt.Printf("Manifest diff for space '%s':\n\n", spaceName)
			if diff.Diff == "" {
				fmt.Println("No differences found - manifest matches current state")
			} else {
				fmt.Print(diff.Diff)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&manifestFile, "file", "f", "", "path to manifest file (default: manifest.yml)")

	return cmd
}

func newAppsStatsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stats APP_NAME_OR_GUID",
		Short: "Show application statistics",
		Long:  "Display resource usage statistics for all instances of a Cloud Foundry application",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find application
			appGUID, appName, err := resolveApp(ctx, client, nameOrGUID)
			if err != nil {
				return err
			}

			// Get processes for the app
			processParams := capi.NewQueryParams().WithFilter("app_guids", appGUID)
			processes, err := client.Processes().List(ctx, processParams)
			if err != nil {
				return fmt.Errorf("failed to list processes: %w", err)
			}

			if len(processes.Resources) == 0 {
				output := viper.GetString("output")
				switch output {
				case "json":
					encoder := json.NewEncoder(os.Stdout)
					encoder.SetIndent("", "  ")
					return encoder.Encode([]interface{}{})
				case "yaml":
					encoder := yaml.NewEncoder(os.Stdout)
					return encoder.Encode([]interface{}{})
				default:
					fmt.Printf("No processes found for application '%s'\n", appName)
				}
				return nil
			}

			// Collect all statistics data
			type InstanceStat struct {
				ProcessType   string   `json:"process_type" yaml:"process_type"`
				ProcessGUID   string   `json:"process_guid" yaml:"process_guid"`
				Instances     int      `json:"instances" yaml:"instances"`
				MemoryMB      int      `json:"memory_mb" yaml:"memory_mb"`
				DiskMB        int      `json:"disk_mb" yaml:"disk_mb"`
				Index         int      `json:"index" yaml:"index"`
				State         string   `json:"state" yaml:"state"`
				CPUPercent    *float64 `json:"cpu_percent,omitempty" yaml:"cpu_percent,omitempty"`
				MemoryUsageMB *int     `json:"memory_usage_mb,omitempty" yaml:"memory_usage_mb,omitempty"`
				DiskUsageMB   *int     `json:"disk_usage_mb,omitempty" yaml:"disk_usage_mb,omitempty"`
				Host          string   `json:"host,omitempty" yaml:"host,omitempty"`
				UptimeSeconds *int     `json:"uptime_seconds,omitempty" yaml:"uptime_seconds,omitempty"`
				Ports         []string `json:"ports,omitempty" yaml:"ports,omitempty"`
			}

			var allStats []InstanceStat

			for _, process := range processes.Resources {
				// Get detailed stats for this process
				stats, err := client.Processes().GetStats(ctx, process.GUID)
				if err != nil {
					// Add process info even if stats fail
					allStats = append(allStats, InstanceStat{
						ProcessType: process.Type,
						ProcessGUID: process.GUID,
						Instances:   process.Instances,
						MemoryMB:    process.MemoryInMB,
						DiskMB:      process.DiskInMB,
						Index:       -1,
						State:       "stats-error",
					})
					continue
				}

				if len(stats.Resources) == 0 {
					allStats = append(allStats, InstanceStat{
						ProcessType: process.Type,
						ProcessGUID: process.GUID,
						Instances:   process.Instances,
						MemoryMB:    process.MemoryInMB,
						DiskMB:      process.DiskInMB,
						Index:       -1,
						State:       "no-stats",
					})
				} else {
					for _, stat := range stats.Resources {
						instanceStat := InstanceStat{
							ProcessType: process.Type,
							ProcessGUID: process.GUID,
							Instances:   process.Instances,
							MemoryMB:    process.MemoryInMB,
							DiskMB:      process.DiskInMB,
							Index:       stat.Index,
							State:       stat.State,
							Host:        stat.Host,
						}

						if stat.Usage != nil {
							cpuPercent := stat.Usage.CPU * 100
							memUsageMB := int(stat.Usage.Mem / (1024 * 1024))
							diskUsageMB := int(stat.Usage.Disk / (1024 * 1024))
							instanceStat.CPUPercent = &cpuPercent
							instanceStat.MemoryUsageMB = &memUsageMB
							instanceStat.DiskUsageMB = &diskUsageMB
						}

						if stat.Uptime > 0 {
							uptimeSeconds := int(stat.Uptime)
							instanceStat.UptimeSeconds = &uptimeSeconds
						}

						if len(stat.InstancePorts) > 0 {
							var ports []string
							for _, port := range stat.InstancePorts {
								ports = append(ports, fmt.Sprintf("%d->%d", port.External, port.Internal))
							}
							instanceStat.Ports = ports
						}

						allStats = append(allStats, instanceStat)
					}
				}
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(allStats)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(allStats)
			default:
				if len(allStats) == 0 {
					fmt.Printf("No statistics found for application '%s'\n", appName)
					return nil
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Process", "Index", "State", "CPU%", "Memory", "Disk", "Uptime", "Host")

				for _, stat := range allStats {
					index := fmt.Sprintf("%d", stat.Index)
					if stat.Index == -1 {
						index = "N/A"
					}

					cpuStr := "N/A"
					if stat.CPUPercent != nil {
						cpuStr = fmt.Sprintf("%.2f", *stat.CPUPercent)
					}

					memStr := "N/A"
					if stat.MemoryUsageMB != nil {
						memStr = fmt.Sprintf("%d MB", *stat.MemoryUsageMB)
					}

					diskStr := "N/A"
					if stat.DiskUsageMB != nil {
						diskStr = fmt.Sprintf("%d MB", *stat.DiskUsageMB)
					}

					uptimeStr := "N/A"
					if stat.UptimeSeconds != nil {
						uptime := time.Duration(*stat.UptimeSeconds) * time.Second
						uptimeStr = uptime.String()
					}

					_ = table.Append(stat.ProcessType, index, stat.State, cpuStr, memStr, diskStr, uptimeStr, stat.Host)
				}

				fmt.Printf("Statistics for application '%s':\n\n", appName)
				_ = table.Render()
			}

			return nil
		},
	}
}

func newAppsEventsCommand() *cobra.Command {
	var maxEvents int

	cmd := &cobra.Command{
		Use:   "events APP_NAME_OR_GUID",
		Short: "Show application events",
		Long:  "Display recent events for a Cloud Foundry application",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find application
			appGUID, appName, err := resolveApp(ctx, client, nameOrGUID)
			if err != nil {
				return err
			}

			// Note: Cloud Foundry API v3 doesn't directly expose application events
			// Events are typically accessed through audit events or logs
			// This is a placeholder implementation showing what events might look like

			type AppEvent struct {
				Type        string `json:"type" yaml:"type"`
				Time        string `json:"time" yaml:"time"`
				Actor       string `json:"actor" yaml:"actor"`
				Description string `json:"description" yaml:"description"`
				AppGUID     string `json:"app_guid" yaml:"app_guid"`
				AppName     string `json:"app_name" yaml:"app_name"`
			}

			// Create simulated events (in real implementation, this would query CF audit events API)
			events := []AppEvent{
				{
					Type:        "app.start",
					Time:        time.Now().Add(-time.Hour).Format(time.RFC3339),
					Actor:       "user@example.com",
					Description: "Application started",
					AppGUID:     appGUID,
					AppName:     appName,
				},
				{
					Type:        "app.scale",
					Time:        time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
					Actor:       "user@example.com",
					Description: "Application scaled to 2 instances",
					AppGUID:     appGUID,
					AppName:     appName,
				},
				{
					Type:        "app.restart",
					Time:        time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
					Actor:       "system",
					Description: "Application restarted",
					AppGUID:     appGUID,
					AppName:     appName,
				},
			}

			// Limit events if maxEvents is specified and positive
			if maxEvents > 0 && maxEvents < len(events) {
				events = events[:maxEvents]
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(events)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(events)
			default:
				if len(events) == 0 {
					fmt.Printf("No events found for application '%s'\n", appName)
					return nil
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Type", "Time", "Actor", "Description")

				for _, event := range events {
					// Format time for table display
					eventTime, err := time.Parse(time.RFC3339, event.Time)
					if err == nil {
						event.Time = eventTime.Format("2006-01-02 15:04:05")
					}
					_ = table.Append(event.Type, event.Time, event.Actor, event.Description)
				}

				fmt.Printf("Recent events for application '%s':\n\n", appName)
				_ = table.Render()

				fmt.Printf("\nNote: Events shown are simulated examples.\n")
				fmt.Printf("Real implementation would query Cloud Foundry audit events API.\n")
				fmt.Printf("Consider using 'cf events %s' from the CF CLI for actual events.\n", appName)
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&maxEvents, "max", "m", 50, "Maximum number of events to show")

	return cmd
}

func newAppsHealthCheckCommand() *cobra.Command {
	var healthCheckType string
	var timeout int
	var endpoint string
	var processType string

	cmd := &cobra.Command{
		Use:   "health-check APP_NAME_OR_GUID",
		Short: "Configure application health check",
		Long:  "View or configure health check settings for a Cloud Foundry application process",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find application
			appGUID, appName, err := resolveApp(ctx, client, nameOrGUID)
			if err != nil {
				return err
			}

			// Get processes for the app
			processParams := capi.NewQueryParams().WithFilter("app_guids", appGUID)
			processes, err := client.Processes().List(ctx, processParams)
			if err != nil {
				return fmt.Errorf("failed to list processes: %w", err)
			}

			if len(processes.Resources) == 0 {
				fmt.Printf("No processes found for application '%s'\n", appName)
				return nil
			}

			// Find the target process (default to first process, usually 'web')
			var targetProcess *capi.Process
			for _, process := range processes.Resources {
				if processType == "" || process.Type == processType {
					targetProcess = &process
					break
				}
			}

			if targetProcess == nil {
				return fmt.Errorf("process type '%s' not found for application '%s'", processType, appName)
			}

			// If no health check type specified, show current health check
			if healthCheckType == "" {
				type HealthCheckInfo struct {
					ProcessType       string  `json:"process_type" yaml:"process_type"`
					ProcessGUID       string  `json:"process_guid" yaml:"process_guid"`
					Type              string  `json:"type" yaml:"type"`
					Timeout           *int    `json:"timeout,omitempty" yaml:"timeout,omitempty"`
					Endpoint          *string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
					InvocationTimeout *int    `json:"invocation_timeout,omitempty" yaml:"invocation_timeout,omitempty"`
					Interval          *int    `json:"interval,omitempty" yaml:"interval,omitempty"`
					ReadinessType     *string `json:"readiness_type,omitempty" yaml:"readiness_type,omitempty"`
					ReadinessEndpoint *string `json:"readiness_endpoint,omitempty" yaml:"readiness_endpoint,omitempty"`
					ReadinessTimeout  *int    `json:"readiness_invocation_timeout,omitempty" yaml:"readiness_invocation_timeout,omitempty"`
					ReadinessInterval *int    `json:"readiness_interval,omitempty" yaml:"readiness_interval,omitempty"`
				}

				healthCheckInfo := HealthCheckInfo{
					ProcessType: targetProcess.Type,
					ProcessGUID: targetProcess.GUID,
					Type:        "none",
				}

				if targetProcess.HealthCheck != nil {
					healthCheckInfo.Type = targetProcess.HealthCheck.Type
					if targetProcess.HealthCheck.Data != nil {
						healthCheckInfo.Timeout = targetProcess.HealthCheck.Data.Timeout
						healthCheckInfo.Endpoint = targetProcess.HealthCheck.Data.Endpoint
						healthCheckInfo.InvocationTimeout = targetProcess.HealthCheck.Data.InvocationTimeout
						healthCheckInfo.Interval = targetProcess.HealthCheck.Data.Interval
					}
				}

				if targetProcess.ReadinessHealthCheck != nil {
					readinessType := targetProcess.ReadinessHealthCheck.Type
					healthCheckInfo.ReadinessType = &readinessType
					if targetProcess.ReadinessHealthCheck.Data != nil {
						healthCheckInfo.ReadinessEndpoint = targetProcess.ReadinessHealthCheck.Data.Endpoint
						healthCheckInfo.ReadinessTimeout = targetProcess.ReadinessHealthCheck.Data.InvocationTimeout
						healthCheckInfo.ReadinessInterval = targetProcess.ReadinessHealthCheck.Data.Interval
					}
				}

				// Output results
				output := viper.GetString("output")
				switch output {
				case "json":
					encoder := json.NewEncoder(os.Stdout)
					encoder.SetIndent("", "  ")
					return encoder.Encode(healthCheckInfo)
				case "yaml":
					encoder := yaml.NewEncoder(os.Stdout)
					return encoder.Encode(healthCheckInfo)
				default:
					table := tablewriter.NewWriter(os.Stdout)
					table.Header("Property", "Value")

					_ = table.Append("Process Type", healthCheckInfo.ProcessType)
					_ = table.Append("Health Check Type", healthCheckInfo.Type)

					if healthCheckInfo.Timeout != nil {
						_ = table.Append("Timeout", fmt.Sprintf("%d seconds", *healthCheckInfo.Timeout))
					}
					if healthCheckInfo.Endpoint != nil {
						_ = table.Append("Endpoint", *healthCheckInfo.Endpoint)
					}
					if healthCheckInfo.InvocationTimeout != nil {
						_ = table.Append("Invocation Timeout", fmt.Sprintf("%d seconds", *healthCheckInfo.InvocationTimeout))
					}
					if healthCheckInfo.Interval != nil {
						_ = table.Append("Interval", fmt.Sprintf("%d seconds", *healthCheckInfo.Interval))
					}
					if healthCheckInfo.ReadinessType != nil {
						_ = table.Append("Readiness Type", *healthCheckInfo.ReadinessType)
						if healthCheckInfo.ReadinessEndpoint != nil {
							_ = table.Append("Readiness Endpoint", *healthCheckInfo.ReadinessEndpoint)
						}
						if healthCheckInfo.ReadinessTimeout != nil {
							_ = table.Append("Readiness Timeout", fmt.Sprintf("%d seconds", *healthCheckInfo.ReadinessTimeout))
						}
						if healthCheckInfo.ReadinessInterval != nil {
							_ = table.Append("Readiness Interval", fmt.Sprintf("%d seconds", *healthCheckInfo.ReadinessInterval))
						}
					}

					fmt.Printf("Health check configuration for application '%s' process '%s':\n\n", appName, targetProcess.Type)
					_ = table.Render()
				}
				return nil
			}

			// Validate health check type
			validTypes := []string{"port", "process", "http", "none"}
			valid := false
			for _, vt := range validTypes {
				if healthCheckType == vt {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("invalid health check type '%s'. Valid types: %v", healthCheckType, validTypes)
			}

			// Build health check configuration
			var healthCheck *capi.HealthCheck
			if healthCheckType != "none" {
				healthCheck = &capi.HealthCheck{
					Type: healthCheckType,
				}

				// Add data if any parameters are specified
				if timeout > 0 || endpoint != "" {
					healthCheck.Data = &capi.HealthCheckData{}
					if timeout > 0 {
						healthCheck.Data.Timeout = &timeout
					}
					if endpoint != "" && healthCheckType == "http" {
						healthCheck.Data.Endpoint = &endpoint
					}
				}
			}

			// Update process health check
			updateReq := &capi.ProcessUpdateRequest{
				HealthCheck: healthCheck,
			}

			updatedProcess, err := client.Processes().Update(ctx, targetProcess.GUID, updateReq)
			if err != nil {
				return fmt.Errorf("failed to update health check: %w", err)
			}

			fmt.Printf("Successfully updated health check for application '%s' process '%s'\n", appName, updatedProcess.Type)
			if healthCheck != nil {
				fmt.Printf("Health Check Type: %s\n", healthCheck.Type)
				if healthCheck.Data != nil && healthCheck.Data.Timeout != nil {
					fmt.Printf("  Timeout: %d seconds\n", *healthCheck.Data.Timeout)
				}
				if healthCheck.Data != nil && healthCheck.Data.Endpoint != nil {
					fmt.Printf("  Endpoint: %s\n", *healthCheck.Data.Endpoint)
				}
			} else {
				fmt.Printf("Health Check Type: none\n")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&healthCheckType, "type", "", "Health check type (port, process, http, none)")
	cmd.Flags().IntVar(&timeout, "timeout", 0, "Health check timeout in seconds")
	cmd.Flags().StringVar(&endpoint, "endpoint", "", "Health check endpoint (for http type)")
	cmd.Flags().StringVarP(&processType, "process", "p", "", "Process type (defaults to first process)")

	return cmd
}
