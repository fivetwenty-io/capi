package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config represents the CLI configuration
type Config struct {
	API               string            `json:"api" yaml:"api"`
	Token             string            `json:"token,omitempty" yaml:"token,omitempty"`
	RefreshToken      string            `json:"refresh_token,omitempty" yaml:"refresh_token,omitempty"`
	Username          string            `json:"username,omitempty" yaml:"username,omitempty"`
	Organization      string            `json:"organization,omitempty" yaml:"organization,omitempty"`
	Space             string            `json:"space,omitempty" yaml:"space,omitempty"`
	SkipSSLValidation bool              `json:"skip_ssl_validation" yaml:"skip_ssl_validation"`
	Output            string            `json:"output" yaml:"output"`
	NoColor           bool              `json:"no_color" yaml:"no_color"`
	Targets           map[string]Target `json:"targets,omitempty" yaml:"targets,omitempty"`
	CurrentTarget     string            `json:"current_target,omitempty" yaml:"current_target,omitempty"`
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
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Long:  "Display the current CLI configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			config := loadConfig()

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
				fmt.Println("Current Configuration:")
				fmt.Printf("  API:                %s\n", config.API)
				fmt.Printf("  Username:           %s\n", config.Username)
				fmt.Printf("  Organization:       %s\n", config.Organization)
				fmt.Printf("  Space:              %s\n", config.Space)
				fmt.Printf("  Skip SSL:           %v\n", config.SkipSSLValidation)
				fmt.Printf("  Output:             %s\n", config.Output)
				fmt.Printf("  No Color:           %v\n", config.NoColor)
				if config.CurrentTarget != "" {
					fmt.Printf("  Current Target:     %s\n", config.CurrentTarget)
				}
				if len(config.Targets) > 0 {
					fmt.Println("\nSaved Targets:")
					for name, target := range config.Targets {
						fmt.Printf("  %s:\n", name)
						fmt.Printf("    API:              %s\n", target.API)
						fmt.Printf("    Username:         %s\n", target.Username)
						fmt.Printf("    Organization:     %s\n", target.Organization)
						fmt.Printf("    Space:            %s\n", target.Space)
					}
				}
			}
			return nil
		},
	}
}

func newConfigSetCommand() *cobra.Command {
	var key, value string

	cmd := &cobra.Command{
		Use:   "set KEY VALUE",
		Short: "Set a configuration value",
		Long:  "Set a specific configuration value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key = args[0]
			value = args[1]

			viper.Set(key, value)

			if err := saveConfig(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("Set %s = %s\n", key, value)
			return nil
		},
	}

	return cmd
}

func newConfigUnsetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "unset KEY",
		Short: "Unset a configuration value",
		Long:  "Remove a specific configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			// Check if key exists
			if !viper.IsSet(key) {
				return fmt.Errorf("key %s is not set", key)
			}

			// Can't directly unset in viper, so we need to manipulate the config
			config := loadConfig()
			switch key {
			case "api":
				config.API = ""
			case "token":
				config.Token = ""
			case "refresh_token":
				config.RefreshToken = ""
			case "username":
				config.Username = ""
			case "organization":
				config.Organization = ""
			case "space":
				config.Space = ""
			case "skip_ssl_validation":
				config.SkipSSLValidation = false
			case "output":
				config.Output = "table"
			case "no_color":
				config.NoColor = false
			default:
				return fmt.Errorf("unknown configuration key: %s", key)
			}

			if err := saveConfigStruct(config); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("Unset %s\n", key)
			return nil
		},
	}
}

func newConfigClearCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Clear all configuration",
		Long:  "Remove all configuration settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile := viper.ConfigFileUsed()
			if configFile == "" {
				home, _ := os.UserHomeDir()
				configFile = filepath.Join(home, ".capi", "config.yml")
			}

			if err := os.Remove(configFile); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to remove config file: %w", err)
			}

			fmt.Println("Configuration cleared")
			return nil
		},
	}
}

func loadConfig() *Config {
	config := &Config{
		API:               viper.GetString("api"),
		Token:             viper.GetString("token"),
		RefreshToken:      viper.GetString("refresh_token"),
		Username:          viper.GetString("username"),
		Organization:      viper.GetString("organization"),
		Space:             viper.GetString("space"),
		SkipSSLValidation: viper.GetBool("skip_ssl_validation"),
		Output:            viper.GetString("output"),
		NoColor:           viper.GetBool("no_color"),
		CurrentTarget:     viper.GetString("current_target"),
		Targets:           make(map[string]Target),
	}

	// Convert targets from viper
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
				config.Targets[name] = target
			}
		}
	}

	return config
}

func saveConfig() error {
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		configDir := filepath.Join(home, ".capi")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return err
		}
		configFile = filepath.Join(configDir, "config.yml")
	}

	return viper.WriteConfigAs(configFile)
}

func saveConfigStruct(config *Config) error {
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		configDir := filepath.Join(home, ".capi")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return err
		}
		configFile = filepath.Join(configDir, "config.yml")
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0600)
}
