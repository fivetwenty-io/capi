package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/fivetwenty-io/capi/pkg/capi"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// NewAdminCommand creates the admin command group
func NewAdminCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "admin",
		Short: "Administrative operations",
		Long:  "Administrative operations for Cloud Foundry platform management",
	}

	cmd.AddCommand(newAdminClearCacheCommand())
	cmd.AddCommand(newAdminUsageSummaryCommand())
	cmd.AddCommand(newAdminInfoCommand())

	return cmd
}

func newAdminClearCacheCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear-cache",
		Short: "Clear platform buildpack cache",
		Long:  "Clear the buildpack cache for the entire Cloud Foundry platform",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Clear buildpack cache
			job, err := client.ClearBuildpackCache(ctx)
			if err != nil {
				return fmt.Errorf("clearing buildpack cache: %w", err)
			}

			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(job)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(job)
			default:
				fmt.Printf("✓ Buildpack cache clear initiated\n")
				if job != nil {
					fmt.Printf("Job GUID: %s\n", job.GUID)
					fmt.Printf("Monitor job status with: capi jobs get %s\n", job.GUID)
				}
			}

			return nil
		},
	}

	return cmd
}

func newAdminUsageSummaryCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "usage-summary",
		Short: "Get platform usage summary",
		Long:  "Get usage summary for the entire Cloud Foundry platform",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Get usage summary
			summary, err := client.GetUsageSummary(ctx)
			if err != nil {
				return fmt.Errorf("getting usage summary: %w", err)
			}

			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(summary)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(summary)
			default:
				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Metric", "Value")

				_ = table.Append("Memory (MB)", fmt.Sprintf("%d", summary.UsageSummary.MemoryInMB))
				_ = table.Append("Started Instances", fmt.Sprintf("%d", summary.UsageSummary.StartedInstances))

				fmt.Println("Platform Usage Summary:")
				fmt.Println()
				if err := table.Render(); err != nil {
					return fmt.Errorf("failed to render table: %w", err)
				}
			}

			return nil
		},
	}
}

func newAdminInfoCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Get extended platform information",
		Long:  "Get extended information about the Cloud Foundry platform",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Get platform info
			info, err := client.GetInfo(ctx)
			if err != nil {
				return fmt.Errorf("getting platform info: %w", err)
			}

			// Get usage summary for additional context
			summary, err := client.GetUsageSummary(ctx)
			if err != nil {
				return fmt.Errorf("getting usage summary: %w", err)
			}

			// Combine info for extended view
			type ExtendedInfo struct {
				Name        string                `json:"name" yaml:"name"`
				Build       string                `json:"build" yaml:"build"`
				Version     int                   `json:"version" yaml:"version"`
				Description string                `json:"description" yaml:"description"`
				Links       capi.Links            `json:"links" yaml:"links"`
				Usage       capi.UsageSummaryData `json:"usage" yaml:"usage"`
			}

			extendedInfo := ExtendedInfo{
				Name:        info.Name,
				Build:       info.Build,
				Version:     info.Version,
				Description: info.Description,
				Links:       info.Links,
				Usage:       summary.UsageSummary,
			}

			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(extendedInfo)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(extendedInfo)
			default:
				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Property", "Value")

				_ = table.Append("Name", info.Name)
				_ = table.Append("Build", info.Build)
				_ = table.Append("API Version", fmt.Sprintf("%d", info.Version))
				_ = table.Append("Description", info.Description)

				_ = table.Append("Memory Usage (MB)", fmt.Sprintf("%d", summary.UsageSummary.MemoryInMB))
				_ = table.Append("Started Instances", fmt.Sprintf("%d", summary.UsageSummary.StartedInstances))

				// Add important links
				if authURL, exists := info.Links["authorization_endpoint"]; exists {
					_ = table.Append("Auth Endpoint", authURL.Href)
				}
				if tokenURL, exists := info.Links["token_endpoint"]; exists {
					_ = table.Append("Token Endpoint", tokenURL.Href)
				}
				if logURL, exists := info.Links["logging_endpoint"]; exists {
					_ = table.Append("Logging Endpoint", logURL.Href)
				}

				fmt.Println("Extended Platform Information:")
				fmt.Println()
				if err := table.Render(); err != nil {
					return fmt.Errorf("failed to render table: %w", err)
				}
			}

			return nil
		},
	}
}
