package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// NewEnvVarGroupsCommand creates the environment variable groups command group
func NewEnvVarGroupsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "env-var-groups",
		Aliases: []string{"environment-variable-groups", "env-groups", "evg"},
		Short:   "Manage environment variable groups",
		Long:    "View and update environment variable groups (running and staging)",
	}

	cmd.AddCommand(newEnvVarGroupsGetCommand())
	cmd.AddCommand(newEnvVarGroupsUpdateCommand())

	return cmd
}

func newEnvVarGroupsGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:       "get GROUP_NAME",
		Short:     "Get environment variable group",
		Long:      "Display environment variables for a specific group (running or staging)",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"running", "staging"},
		RunE: func(cmd *cobra.Command, args []string) error {
			groupName := args[0]

			// Validate group name
			if groupName != "running" && groupName != "staging" {
				return fmt.Errorf("invalid group name '%s'. Valid groups: running, staging", groupName)
			}

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			envVarGroup, err := client.EnvironmentVariableGroups().Get(ctx, groupName)
			if err != nil {
				return fmt.Errorf("failed to get environment variable group: %w", err)
			}

			// Collect environment variables for structured output
			type EnvVar struct {
				Name  string      `json:"name" yaml:"name"`
				Value interface{} `json:"value" yaml:"value"`
			}

			var envVarsList []EnvVar
			for key, value := range envVarGroup.Var {
				envVarsList = append(envVarsList, EnvVar{
					Name:  key,
					Value: value,
				})
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				result := map[string]interface{}{
					"name":                  envVarGroup.Name,
					"updated_at":            envVarGroup.UpdatedAt,
					"environment_variables": envVarsList,
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(result)
			case "yaml":
				result := map[string]interface{}{
					"name":                  envVarGroup.Name,
					"updated_at":            envVarGroup.UpdatedAt,
					"environment_variables": envVarsList,
				}
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(result)
			default:
				fmt.Printf("Environment Variable Group: %s\n", envVarGroup.Name)
				if envVarGroup.UpdatedAt != nil {
					fmt.Printf("  Last Updated: %s\n", envVarGroup.UpdatedAt.Format("2006-01-02 15:04:05"))
				}
				fmt.Println()

				if len(envVarsList) == 0 {
					fmt.Printf("No environment variables found in group '%s'\n", groupName)
					return nil
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Name", "Value")

				for _, envVar := range envVarsList {
					valueStr := fmt.Sprintf("%v", envVar.Value)
					// Truncate long values for table display
					if len(valueStr) > 80 {
						valueStr = valueStr[:77] + "..."
					}
					_ = table.Append(envVar.Name, valueStr)
				}

				_ = table.Render()
			}

			return nil
		},
	}
}

func newEnvVarGroupsUpdateCommand() *cobra.Command {
	var (
		envVars  []string
		unset    []string
		fromFile string
	)

	cmd := &cobra.Command{
		Use:       "update GROUP_NAME",
		Short:     "Update environment variable group",
		Long:      "Update environment variables for a specific group (running or staging)",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"running", "staging"},
		RunE: func(cmd *cobra.Command, args []string) error {
			groupName := args[0]

			// Validate group name
			if groupName != "running" && groupName != "staging" {
				return fmt.Errorf("invalid group name '%s'. Valid groups: running, staging", groupName)
			}

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Get current environment variables
			currentGroup, err := client.EnvironmentVariableGroups().Get(ctx, groupName)
			if err != nil {
				return fmt.Errorf("failed to get current environment variables: %w", err)
			}

			// Start with current environment variables
			updatedEnvVars := make(map[string]interface{})
			for k, v := range currentGroup.Var {
				updatedEnvVars[k] = v
			}

			// Load from file if specified
			if fromFile != "" {
				fileEnvVars, err := loadEnvVarsFromFile(fromFile)
				if err != nil {
					return fmt.Errorf("failed to load environment variables from file: %w", err)
				}
				for k, v := range fileEnvVars {
					updatedEnvVars[k] = v
				}
			}

			// Apply env var updates
			for _, envVar := range envVars {
				parts := strings.SplitN(envVar, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid environment variable format '%s'. Expected KEY=VALUE", envVar)
				}
				key := parts[0]
				value := parts[1]

				// Try to parse value as different types
				updatedEnvVars[key] = parseValue(value)
			}

			// Remove variables specified in --unset
			for _, key := range unset {
				delete(updatedEnvVars, key)
			}

			// Update environment variable group
			updatedGroup, err := client.EnvironmentVariableGroups().Update(ctx, groupName, updatedEnvVars)
			if err != nil {
				return fmt.Errorf("failed to update environment variable group: %w", err)
			}

			fmt.Printf("Successfully updated environment variable group '%s'\n", updatedGroup.Name)

			// Show summary of changes
			if len(envVars) > 0 {
				fmt.Printf("Set %d environment variable(s)\n", len(envVars))
			}
			if len(unset) > 0 {
				fmt.Printf("Unset %d environment variable(s): %s\n", len(unset), strings.Join(unset, ", "))
			}
			if fromFile != "" {
				fmt.Printf("Loaded environment variables from file: %s\n", fromFile)
			}

			return nil
		},
	}

	cmd.Flags().StringSliceVarP(&envVars, "set", "s", nil, "set environment variable (KEY=VALUE)")
	cmd.Flags().StringSliceVarP(&unset, "unset", "u", nil, "unset environment variable (KEY)")
	cmd.Flags().StringVarP(&fromFile, "from-file", "f", "", "load environment variables from file")

	return cmd
}

// parseValue attempts to parse a string value as the most appropriate type
func parseValue(value string) interface{} {
	// Try to parse as boolean
	if boolVal, err := strconv.ParseBool(value); err == nil {
		return boolVal
	}

	// Try to parse as integer
	if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
		return intVal
	}

	// Try to parse as float
	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		return floatVal
	}

	// Return as string if no other type matches
	return value
}

// loadEnvVarsFromFile loads environment variables from a file
// Supports .env format (KEY=VALUE per line) and JSON/YAML formats
func loadEnvVarsFromFile(filename string) (map[string]interface{}, error) {
	// Validate and clean the file path to prevent directory traversal attacks
	cleanPath := filepath.Clean(filename)
	if filepath.IsAbs(cleanPath) {
		// For absolute paths, ensure they don't escape the filesystem root
		absPath, err := filepath.Abs(cleanPath)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve absolute path: %w", err)
		}
		cleanPath = absPath
	} else {
		// For relative paths, resolve them relative to current working directory
		absPath, err := filepath.Abs(cleanPath)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve relative path: %w", err)
		}
		cleanPath = absPath
	}

	// Additional validation: ensure the path doesn't contain directory traversal sequences
	if strings.Contains(cleanPath, "..") {
		return nil, fmt.Errorf("path contains directory traversal sequences: %s", filename)
	}

	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	envVars := make(map[string]interface{})

	// Try to parse as JSON first
	if err := json.Unmarshal(data, &envVars); err == nil {
		return envVars, nil
	}

	// Try to parse as YAML
	if err := yaml.Unmarshal(data, &envVars); err == nil {
		return envVars, nil
	}

	// Parse as .env format (KEY=VALUE per line)
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid .env format on line %d: %s", i+1, line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
			(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			value = value[1 : len(value)-1]
		}

		envVars[key] = parseValue(value)
	}

	return envVars, nil
}
