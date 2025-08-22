package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cloudfoundry-community/go-uaa"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// createUsersCurlCommand creates the curl command for direct UAA API access
func createUsersCurlCommand() *cobra.Command {
	var method, data string
	var headers []string
	var outputFile string

	cmd := &cobra.Command{
		Use:   "curl <path>",
		Short: "Direct UAA API access",
		Long: `Make direct HTTP requests to UAA API endpoints.

This command allows you to make arbitrary HTTP requests to the UAA API,
automatically injecting authentication headers. The path should be relative
to the UAA endpoint (e.g., '/Users' or '/oauth/clients').`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := loadConfig()
			path := args[0]

			if config.UAAEndpoint == "" {
				return fmt.Errorf("no UAA endpoint configured. Use 'capi uaa target <url>' to set one")
			}

			// Create UAA client
			uaaClient, err := NewUAAClient(config)
			if err != nil {
				return fmt.Errorf("failed to create UAA client: %w", err)
			}

			if !uaaClient.IsAuthenticated() {
				return fmt.Errorf("not authenticated. Use a token command to authenticate first")
			}

			// Make the curl request
			body, responseHeaders, statusCode, err := uaaClient.Client().Curl(path, method, data, headers)
			if err != nil {
				return fmt.Errorf("failed to make request: %w", err)
			}

			// Handle output
			if outputFile != "" {
				// Write response body to file
				err := os.WriteFile(outputFile, []byte(body), 0600)
				if err != nil {
					return fmt.Errorf("failed to write output to file: %w", err)
				}
				fmt.Printf("Response written to %s\n", outputFile)
				fmt.Printf("Status: %d\n", statusCode)
				return nil
			}

			// Display response
			fmt.Printf("Status: %d\n", statusCode)
			if responseHeaders != "" {
				fmt.Printf("Headers:\n%s\n", responseHeaders)
			}
			if body != "" {
				fmt.Printf("Response:\n%s\n", body)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&method, "method", "X", "GET", "HTTP method")
	cmd.Flags().StringVarP(&data, "data", "d", "", "Request body data")
	cmd.Flags().StringSliceVarP(&headers, "header", "H", nil, "HTTP headers (format: 'Key: Value')")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Write response body to file")

	return cmd
}

// createUsersUserinfoCommand creates the userinfo command
func createUsersUserinfoCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "userinfo",
		Short: "Display current user claims",
		Long: `Show claims about the currently authenticated user.

This command displays information about the user associated with the
current access token, including user attributes and granted scopes.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			config := loadConfig()

			if config.UAAEndpoint == "" {
				return fmt.Errorf("no UAA endpoint configured. Use 'capi uaa target <url>' to set one")
			}

			// Create UAA client
			uaaClient, err := NewUAAClient(config)
			if err != nil {
				return fmt.Errorf("failed to create UAA client: %w", err)
			}

			if !uaaClient.IsAuthenticated() {
				return fmt.Errorf("not authenticated. Use a token command to authenticate first")
			}

			// Get user info
			userInfo, err := uaaClient.Client().GetMe()
			if err != nil {
				return fmt.Errorf("failed to get user info: %w", err)
			}

			// Display user info
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(userInfo)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(userInfo)
			default:
				return displayUserInfoTable(userInfo)
			}
		},
	}
}

// Helper function for userinfo display
func displayUserInfoTable(userInfo *uaa.UserInfo) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header("Property", "Value")

	if userInfo.UserID != "" {
		_ = table.Append("User ID", userInfo.UserID)
	}
	if userInfo.Sub != "" {
		_ = table.Append("Subject", userInfo.Sub)
	}
	if userInfo.Username != "" {
		_ = table.Append("Username", userInfo.Username)
	}
	if userInfo.Name != "" {
		_ = table.Append("Name", userInfo.Name)
	}
	if userInfo.GivenName != "" {
		_ = table.Append("Given Name", userInfo.GivenName)
	}
	if userInfo.FamilyName != "" {
		_ = table.Append("Family Name", userInfo.FamilyName)
	}
	if userInfo.Email != "" {
		_ = table.Append("Email", userInfo.Email)
	}
	if userInfo.PhoneNumber != "" {
		_ = table.Append("Phone Number", userInfo.PhoneNumber)
	}
	if userInfo.PreviousLoginTime > 0 {
		_ = table.Append("Previous Login", fmt.Sprintf("%d", userInfo.PreviousLoginTime))
	}

	_ = table.Render()
	return nil
}
