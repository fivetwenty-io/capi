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

// NewInfoCommand creates the info command
func NewInfoCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Display API endpoint information",
		Long:  "Display information about the Cloud Foundry API endpoint",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			info, err := client.GetInfo(ctx)
			if err != nil {
				return fmt.Errorf("failed to get API info: %w", err)
			}

			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(info)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(info)
			default:
				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Property", "Value")

				_ = table.Append("Name", info.Name)
				_ = table.Append("Build", info.Build)
				_ = table.Append("Version", fmt.Sprintf("%d", info.Version))
				_ = table.Append("Description", info.Description)
				_ = table.Append("CF on K8s", fmt.Sprintf("%t", info.CFOnK8s))
				_ = table.Append("CLI Minimum", info.CLIVersion.Minimum)
				_ = table.Append("CLI Recommended", info.CLIVersion.Recommended)

				// Add links
				if len(info.Links) > 0 {
					var linkStrings []string
					for name, link := range info.Links {
						linkStrings = append(linkStrings, fmt.Sprintf("%s: %s", name, link.Href))
					}
					_ = table.Append("Links", strings.Join(linkStrings, "\n"))
				}

				// Add custom fields if present
				if len(info.Custom) > 0 {
					var customStrings []string
					for key, value := range info.Custom {
						customStrings = append(customStrings, fmt.Sprintf("%s: %v", key, value))
					}
					_ = table.Append("Custom", strings.Join(customStrings, "\n"))
				}

				_ = table.Render()
			}

			return nil
		},
	}
}
