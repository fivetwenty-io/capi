package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fivetwenty-io/capi-client/cmd/capi/commands"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "capi",
	Short: "Cloud Foundry API v3 CLI",
	Long: `A command-line interface for interacting with Cloud Foundry API v3.
	
This CLI provides comprehensive access to Cloud Foundry resources including
applications, spaces, organizations, services, and more.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default is $HOME/.capi/config.yml)")
	rootCmd.PersistentFlags().StringP("api", "a", "", "API endpoint URL")
	rootCmd.PersistentFlags().StringP("token", "t", "", "authentication token")
	rootCmd.PersistentFlags().String("output", "table", "output format (table, json, yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().Bool("no-color", false, "disable colored output")
	rootCmd.PersistentFlags().Bool("skip-ssl-validation", false, "skip SSL certificate validation")

	// Bind flags to viper
	viper.BindPFlag("api", rootCmd.PersistentFlags().Lookup("api"))
	viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("no-color", rootCmd.PersistentFlags().Lookup("no-color"))
	viper.BindPFlag("skip-ssl-validation", rootCmd.PersistentFlags().Lookup("skip-ssl-validation"))

	// Add commands
	rootCmd.AddCommand(commands.NewVersionCommand(version, commit, date))
	rootCmd.AddCommand(commands.NewLoginCommand())
	rootCmd.AddCommand(commands.NewLogoutCommand())
	rootCmd.AddCommand(commands.NewConfigCommand())
	rootCmd.AddCommand(commands.NewOrgsCommand())
	rootCmd.AddCommand(commands.NewSpacesCommand())
	rootCmd.AddCommand(commands.NewAppsCommand())
	rootCmd.AddCommand(commands.NewServicesCommand())
	rootCmd.AddCommand(commands.NewDomainsCommand())
	rootCmd.AddCommand(commands.NewRoutesCommand())
	rootCmd.AddCommand(commands.NewBuildpacksCommand())
	rootCmd.AddCommand(commands.NewStacksCommand())
	rootCmd.AddCommand(commands.NewUsersCommand())
	rootCmd.AddCommand(commands.NewRolesCommand())
	rootCmd.AddCommand(commands.NewTargetCommand())
	rootCmd.AddCommand(commands.NewInfoCommand())
}

func initConfig() {
	cfgFile := viper.GetString("config")

	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		// Create config directory if it doesn't exist
		configDir := filepath.Join(home, ".capi")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating config directory: %v\n", err)
		}

		// Search config in ~/.capi/config.yml
		viper.AddConfigPath(configDir)
		viper.SetConfigType("yml")
		viper.SetConfigName("config")
	}

	// Read in environment variables that match
	viper.SetEnvPrefix("CAPI")
	viper.AutomaticEnv()

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("verbose") {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
