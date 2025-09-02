package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/cloudfoundry-community/go-uaa"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

// createUsersCreateClientCommand creates the create client command
func createUsersCreateClientCommand() *cobra.Command {
	var secret, displayName string
	var grantTypes, redirectURIs, scope, authorities, allowedProviders, requiredUserGroups []string
	var accessTokenValidity, refreshTokenValidity int64
	var autoApprove, allowPublic bool

	cmd := &cobra.Command{
		Use:   "create-client <client-id>",
		Short: "Create OAuth client",
		Long: `Create an OAuth client registration in UAA.

OAuth clients are applications that can authenticate with UAA and obtain
access tokens on behalf of users or using their own credentials.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := loadConfig()
			clientID := args[0]

			if GetEffectiveUAAEndpoint(config) == "" {
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

			// Prompt for secret if not provided
			if secret == "" {
				fmt.Print("Client Secret: ")
				secretBytes, err := term.ReadPassword(int(syscall.Stdin))
				if err != nil {
					return fmt.Errorf("failed to read client secret: %w", err)
				}
				secret = string(secretBytes)
				fmt.Println() // Add newline after password input
			}

			// Build client object
			client := uaa.Client{
				ClientID:             clientID,
				ClientSecret:         secret,
				DisplayName:          displayName,
				AuthorizedGrantTypes: grantTypes,
				RedirectURI:          redirectURIs,
				Scope:                scope,
				Authorities:          authorities,
				AccessTokenValidity:  accessTokenValidity,
				RefreshTokenValidity: refreshTokenValidity,
				AllowedProviders:     allowedProviders,
				RequiredUserGroups:   requiredUserGroups,
				AllowPublic:          allowPublic,
			}

			// Handle autoapprove
			if autoApprove {
				client.AutoApproveRaw = true
			}

			// Create client
			createdClient, err := uaaClient.Client().CreateClient(client)
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}

			// Display created client (mask secret)
			output := viper.GetString("output")
			switch output {
			case "json":
				// Mask secret for display
				displayClient := *createdClient
				displayClient.ClientSecret = "***"
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(displayClient)
			case "yaml":
				// Mask secret for display
				displayClient := *createdClient
				displayClient.ClientSecret = "***"
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(displayClient)
			default:
				return displayClientTable(createdClient, false)
			}
		},
	}

	cmd.Flags().StringVar(&secret, "secret", "", "Client secret")
	cmd.Flags().StringVar(&displayName, "name", "", "Client display name")
	cmd.Flags().StringSliceVar(&grantTypes, "authorized-grant-types", []string{"authorization_code"}, "Authorized grant types")
	cmd.Flags().StringSliceVar(&redirectURIs, "redirect-uri", nil, "Redirect URIs")
	cmd.Flags().StringSliceVar(&scope, "scope", nil, "OAuth scopes")
	cmd.Flags().StringSliceVar(&authorities, "authorities", nil, "Client authorities")
	cmd.Flags().StringSliceVar(&allowedProviders, "allowed-providers", nil, "Allowed identity providers")
	cmd.Flags().StringSliceVar(&requiredUserGroups, "required-user-groups", nil, "Required user groups")
	cmd.Flags().Int64Var(&accessTokenValidity, "access-token-validity", 43200, "Access token validity in seconds")
	cmd.Flags().Int64Var(&refreshTokenValidity, "refresh-token-validity", 2592000, "Refresh token validity in seconds")
	cmd.Flags().BoolVar(&autoApprove, "auto-approve", false, "Auto-approve authorization requests")
	cmd.Flags().BoolVar(&allowPublic, "allow-public", false, "Allow public client (no secret required)")

	return cmd
}

// createUsersGetClientCommand creates the get client command
func createUsersGetClientCommand() *cobra.Command {
	var showSecret bool

	cmd := &cobra.Command{
		Use:   "get-client <client-id>",
		Short: "Get client details",
		Long: `View OAuth client registration details.

By default, the client secret is masked for security. Use --show-secret
to display the actual secret value.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := loadConfig()
			clientID := args[0]

			if GetEffectiveUAAEndpoint(config) == "" {
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

			// Get client
			client, err := uaaClient.Client().GetClient(clientID)
			if err != nil {
				return fmt.Errorf("failed to get client: %w", err)
			}

			// Display client
			output := viper.GetString("output")
			switch output {
			case "json":
				if !showSecret {
					// Mask secret for display
					displayClient := *client
					displayClient.ClientSecret = "***"
					client = &displayClient
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(client)
			case "yaml":
				if !showSecret {
					// Mask secret for display
					displayClient := *client
					displayClient.ClientSecret = "***"
					client = &displayClient
				}
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(client)
			default:
				return displayClientTable(client, showSecret)
			}
		},
	}

	cmd.Flags().BoolVar(&showSecret, "show-secret", false, "Show client secret (default: masked)")

	return cmd
}

// createUsersListClientsCommand creates the list clients command
func createUsersListClientsCommand() *cobra.Command {
	var filter, sortBy string
	var sortOrder string
	var count, startIndex int
	var all bool

	cmd := &cobra.Command{
		Use:   "list-clients",
		Short: "List OAuth clients",
		Long: `List all OAuth clients in the targeted UAA.

Client secrets are never displayed in list operations for security.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			config := loadConfig()

			if GetEffectiveUAAEndpoint(config) == "" {
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

			// Convert sort order string to enum
			var uaaSortOrder uaa.SortOrder
			switch strings.ToLower(sortOrder) {
			case "descending", "desc":
				uaaSortOrder = "descending"
			default:
				uaaSortOrder = uaa.SortAscending
			}

			var clients []uaa.Client
			if all {
				// Get all clients across all pages
				clients, err = uaaClient.Client().ListAllClients(filter, sortBy, uaaSortOrder)
			} else {
				// Get clients with pagination
				clients, _, err = uaaClient.Client().ListClients(filter, sortBy, uaaSortOrder, startIndex, count)
			}

			if err != nil {
				return fmt.Errorf("failed to list clients: %w", err)
			}

			// Mask secrets for display
			for i := range clients {
				clients[i].ClientSecret = "***"
			}

			// Display clients
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(clients)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(clients)
			default:
				return displayClientsTable(clients)
			}
		},
	}

	cmd.Flags().StringVar(&filter, "filter", "", "SCIM filter expression")
	cmd.Flags().StringVar(&sortBy, "sort-by", "", "Attribute to sort by")
	cmd.Flags().StringVar(&sortOrder, "sort-order", "ascending", "Sort order (ascending, descending)")
	cmd.Flags().IntVar(&count, "count", 50, "Number of results per page")
	cmd.Flags().IntVar(&startIndex, "start-index", 1, "Starting index for pagination")
	cmd.Flags().BoolVar(&all, "all", false, "Fetch all clients across all pages")

	return cmd
}

// createUsersUpdateClientCommand creates the update client command
func createUsersUpdateClientCommand() *cobra.Command {
	var displayName string
	var grantTypes, redirectURIs, scope, authorities, allowedProviders, requiredUserGroups []string
	var accessTokenValidity, refreshTokenValidity int64
	var autoApprove, allowPublic *bool

	cmd := &cobra.Command{
		Use:   "update-client <client-id>",
		Short: "Update OAuth client",
		Long: `Update an OAuth client registration.

Only the specified attributes will be updated. Unspecified attributes
will remain unchanged.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := loadConfig()
			clientID := args[0]

			if GetEffectiveUAAEndpoint(config) == "" {
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

			// Get existing client
			existingClient, err := uaaClient.Client().GetClient(clientID)
			if err != nil {
				return fmt.Errorf("failed to get existing client: %w", err)
			}

			// Update specified fields
			if displayName != "" {
				existingClient.DisplayName = displayName
			}
			if len(grantTypes) > 0 {
				existingClient.AuthorizedGrantTypes = grantTypes
			}
			if len(redirectURIs) > 0 {
				existingClient.RedirectURI = redirectURIs
			}
			if len(scope) > 0 {
				existingClient.Scope = scope
			}
			if len(authorities) > 0 {
				existingClient.Authorities = authorities
			}
			if len(allowedProviders) > 0 {
				existingClient.AllowedProviders = allowedProviders
			}
			if len(requiredUserGroups) > 0 {
				existingClient.RequiredUserGroups = requiredUserGroups
			}
			if accessTokenValidity > 0 {
				existingClient.AccessTokenValidity = accessTokenValidity
			}
			if refreshTokenValidity > 0 {
				existingClient.RefreshTokenValidity = refreshTokenValidity
			}
			if autoApprove != nil {
				existingClient.AutoApproveRaw = *autoApprove
			}
			if allowPublic != nil {
				existingClient.AllowPublic = *allowPublic
			}

			// Update client
			updatedClient, err := uaaClient.Client().UpdateClient(*existingClient)
			if err != nil {
				return fmt.Errorf("failed to update client: %w", err)
			}

			// Display updated client
			output := viper.GetString("output")
			switch output {
			case "json":
				// Mask secret for display
				displayClient := *updatedClient
				displayClient.ClientSecret = "***"
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(displayClient)
			case "yaml":
				// Mask secret for display
				displayClient := *updatedClient
				displayClient.ClientSecret = "***"
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(displayClient)
			default:
				return displayClientTable(updatedClient, false)
			}
		},
	}

	cmd.Flags().StringVar(&displayName, "name", "", "Client display name")
	cmd.Flags().StringSliceVar(&grantTypes, "authorized-grant-types", nil, "Authorized grant types")
	cmd.Flags().StringSliceVar(&redirectURIs, "redirect-uri", nil, "Redirect URIs")
	cmd.Flags().StringSliceVar(&scope, "scope", nil, "OAuth scopes")
	cmd.Flags().StringSliceVar(&authorities, "authorities", nil, "Client authorities")
	cmd.Flags().StringSliceVar(&allowedProviders, "allowed-providers", nil, "Allowed identity providers")
	cmd.Flags().StringSliceVar(&requiredUserGroups, "required-user-groups", nil, "Required user groups")
	cmd.Flags().Int64Var(&accessTokenValidity, "access-token-validity", 0, "Access token validity in seconds")
	cmd.Flags().Int64Var(&refreshTokenValidity, "refresh-token-validity", 0, "Refresh token validity in seconds")

	// Use string flags for booleans to distinguish between false and unset
	var autoApproveStr, allowPublicStr string
	cmd.Flags().StringVar(&autoApproveStr, "auto-approve", "", "Auto-approve authorization requests (true/false)")
	cmd.Flags().StringVar(&allowPublicStr, "allow-public", "", "Allow public client (true/false)")

	// Pre-run to parse boolean flags
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if autoApproveStr != "" {
			val, err := strconv.ParseBool(autoApproveStr)
			if err != nil {
				return fmt.Errorf("invalid value for --auto-approve: %s", autoApproveStr)
			}
			autoApprove = &val
		}
		if allowPublicStr != "" {
			val, err := strconv.ParseBool(allowPublicStr)
			if err != nil {
				return fmt.Errorf("invalid value for --allow-public: %s", allowPublicStr)
			}
			allowPublic = &val
		}
		return nil
	}

	return cmd
}

// createUsersSetClientSecretCommand creates the set client secret command
func createUsersSetClientSecretCommand() *cobra.Command {
	var secret string

	cmd := &cobra.Command{
		Use:   "set-client-secret <client-id>",
		Short: "Update client secret",
		Long: `Update the secret for an OAuth client.

If the secret is not provided via the --secret flag, you will be prompted
to enter it securely.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := loadConfig()
			clientID := args[0]

			if GetEffectiveUAAEndpoint(config) == "" {
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

			// Prompt for secret if not provided
			if secret == "" {
				fmt.Print("New Client Secret: ")
				secretBytes, err := term.ReadPassword(int(syscall.Stdin))
				if err != nil {
					return fmt.Errorf("failed to read client secret: %w", err)
				}
				secret = string(secretBytes)
				fmt.Println() // Add newline after password input
			}

			// Update client secret
			err = uaaClient.Client().ChangeClientSecret(clientID, secret)
			if err != nil {
				return fmt.Errorf("failed to update client secret: %w", err)
			}

			fmt.Printf("Successfully updated secret for client '%s'\n", clientID)
			return nil
		},
	}

	cmd.Flags().StringVar(&secret, "secret", "", "New client secret")

	return cmd
}

// createUsersDeleteClientCommand creates the delete client command
func createUsersDeleteClientCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete-client <client-id>",
		Short: "Delete OAuth client",
		Long:  "Delete an OAuth client registration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := loadConfig()
			clientID := args[0]

			if GetEffectiveUAAEndpoint(config) == "" {
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

			// Get client details for confirmation
			client, err := uaaClient.Client().GetClient(clientID)
			if err != nil {
				return fmt.Errorf("failed to get client: %w", err)
			}

			// Confirm deletion unless --force is used
			if !force {
				fmt.Printf("Are you sure you want to delete client '%s'", clientID)
				if client.DisplayName != "" {
					fmt.Printf(" (%s)", client.DisplayName)
				}
				fmt.Print("? [y/N]: ")
				var response string
				_, _ = fmt.Scanln(&response)
				if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
					fmt.Println("Client deletion cancelled")
					return nil
				}
			}

			// Delete client
			_, err = uaaClient.Client().DeleteClient(clientID)
			if err != nil {
				return fmt.Errorf("failed to delete client: %w", err)
			}

			fmt.Printf("Client '%s' has been deleted\n", clientID)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force deletion without confirmation")

	return cmd
}

// Helper functions for client display

func displayClientTable(client *uaa.Client, showSecret bool) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header("Property", "Value")

	if client.ClientID != "" {
		_ = table.Append("Client ID", client.ClientID)
	}
	if client.DisplayName != "" {
		_ = table.Append("Display Name", client.DisplayName)
	}

	// Display secret (masked or actual)
	if showSecret && client.ClientSecret != "" {
		_ = table.Append("Client Secret", client.ClientSecret)
	} else if client.ClientSecret != "" {
		_ = table.Append("Client Secret", "***")
	}

	if len(client.AuthorizedGrantTypes) > 0 {
		_ = table.Append("Grant Types", strings.Join(client.AuthorizedGrantTypes, ", "))
	}
	if len(client.RedirectURI) > 0 {
		_ = table.Append("Redirect URIs", strings.Join(client.RedirectURI, ", "))
	}
	if len(client.Scope) > 0 {
		_ = table.Append("Scope", strings.Join(client.Scope, ", "))
	}
	if len(client.Authorities) > 0 {
		_ = table.Append("Authorities", strings.Join(client.Authorities, ", "))
	}

	if client.AccessTokenValidity > 0 {
		_ = table.Append("Access Token Validity", fmt.Sprintf("%d seconds", client.AccessTokenValidity))
	}
	if client.RefreshTokenValidity > 0 {
		_ = table.Append("Refresh Token Validity", fmt.Sprintf("%d seconds", client.RefreshTokenValidity))
	}

	_ = table.Append("Auto Approve", fmt.Sprintf("%v", client.AutoApprove()))
	_ = table.Append("Allow Public", fmt.Sprintf("%t", client.AllowPublic))

	if len(client.AllowedProviders) > 0 {
		_ = table.Append("Allowed Providers", strings.Join(client.AllowedProviders, ", "))
	}
	if len(client.RequiredUserGroups) > 0 {
		_ = table.Append("Required User Groups", strings.Join(client.RequiredUserGroups, ", "))
	}

	_ = table.Render()
	return nil
}

func displayClientsTable(clients []uaa.Client) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header("Client ID", "Display Name", "Grant Types", "Scope", "Auto Approve")

	for _, client := range clients {
		clientID := client.ClientID
		displayName := client.DisplayName
		grantTypes := strings.Join(client.AuthorizedGrantTypes, ",")
		if len(grantTypes) > 30 {
			grantTypes = grantTypes[:30] + "..."
		}
		scope := strings.Join(client.Scope, ",")
		if len(scope) > 30 {
			scope = scope[:30] + "..."
		}
		autoApprove := fmt.Sprintf("%v", client.AutoApprove())

		_ = table.Append(clientID, displayName, grantTypes, scope, autoApprove)
	}

	_ = table.Render()
	return nil
}
