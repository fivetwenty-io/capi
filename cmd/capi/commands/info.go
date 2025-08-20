package commands

import (
	"context"
	"encoding/json"
	"fmt"

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
			client, err := createClient()
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
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				return encoder.Encode(info)
			case "yaml":
				encoder := yaml.NewEncoder(cmd.OutOrStdout())
				return encoder.Encode(info)
			default:
				fmt.Println("API Information:")
				fmt.Printf("  Name:                    %s\n", info.Name)
				fmt.Printf("  Build:                   %s\n", info.Build)
				fmt.Printf("  Version:                 %d\n", info.Version)
				fmt.Printf("  Description:             %s\n", info.Description)

				if len(info.Links) > 0 {
					fmt.Println("\nLinks:")
					for name, link := range info.Links {
						fmt.Printf("  %s: %s\n", name, link.Href)
					}
				}
			}

			return nil
		},
	}
}
