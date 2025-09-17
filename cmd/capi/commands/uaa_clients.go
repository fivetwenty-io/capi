package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/cloudfoundry-community/go-uaa"
	"github.com/fivetwenty-io/capi/v3/internal/constants"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

// createUsersCreateClientCommand creates the create client command.
func createUsersCreateClientCommand() *cobra.Command {
	var (
		secret, displayName                                                                string
		grantTypes, redirectURIs, scope, authorities, allowedProviders, requiredUserGroups []string
		accessTokenValidity, refreshTokenValidity                                          int64
		autoApprove, allowPublic                                                           bool
	)

	cmd := &cobra.Command{
		Use:   "create-client <client-id>",
		Short: "Create OAuth client",
		Long: `Create an OAuth client registration in UAA.

OAuth clients are applications that can authenticate with UAA and obtain
access tokens on behalf of users or using their own credentials.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreateClientCommand(args[0], &secret, displayName, grantTypes, redirectURIs, scope, authorities, allowedProviders, requiredUserGroups, accessTokenValidity, refreshTokenValidity, autoApprove, allowPublic)
		},
	}

	setupCreateClientFlags(cmd, &secret, &displayName, &grantTypes, &redirectURIs, &scope, &authorities, &allowedProviders, &requiredUserGroups, &accessTokenValidity, &refreshTokenValidity, &autoApprove, &allowPublic)

	return cmd
}

func setupCreateClientFlags(cmd *cobra.Command, secret, displayName *string, grantTypes, redirectURIs, scope, authorities, allowedProviders, requiredUserGroups *[]string, accessTokenValidity, refreshTokenValidity *int64, autoApprove, allowPublic *bool) {
	cmd.Flags().StringVar(secret, "secret", "", "Client secret")
	cmd.Flags().StringVar(displayName, "name", "", "Client display name")
	cmd.Flags().StringSliceVar(grantTypes, "authorized-grant-types", []string{"authorization_code"}, "Authorized grant types")
	cmd.Flags().StringSliceVar(redirectURIs, "redirect-uri", nil, "Redirect URIs")
	cmd.Flags().StringSliceVar(scope, "scope", nil, "OAuth scopes")
	cmd.Flags().StringSliceVar(authorities, "authorities", nil, "Client authorities")
	cmd.Flags().StringSliceVar(allowedProviders, "allowed-providers", nil, "Allowed identity providers")
	cmd.Flags().StringSliceVar(requiredUserGroups, "required-user-groups", nil, "Required user groups")
	cmd.Flags().Int64Var(accessTokenValidity, "access-token-validity", constants.DefaultAccessTokenValidity, "Access token validity in seconds")
	cmd.Flags().Int64Var(refreshTokenValidity, "refresh-token-validity", constants.DefaultRefreshTokenValidity, "Refresh token validity in seconds")
	cmd.Flags().BoolVar(autoApprove, "auto-approve", false, "Auto-approve authorization requests")
	cmd.Flags().BoolVar(allowPublic, "allow-public", false, "Allow public client (no secret required)")
}

func runCreateClientCommand(clientID string, secret *string, displayName string, grantTypes, redirectURIs, scope, authorities, allowedProviders, requiredUserGroups []string, accessTokenValidity, refreshTokenValidity int64, autoApprove, allowPublic bool) error {
	config := loadConfig()

	if GetEffectiveUAAEndpoint(config) == "" {
		return constants.ErrNoUAAConfigured
	}

	uaaClient, err := NewUAAClient(config)
	if err != nil {
		return fmt.Errorf("failed to create UAA client: %w", err)
	}

	if !uaaClient.IsAuthenticated() {
		return constants.ErrNotAuthenticated
	}

	err = promptForClientSecretIfNeeded(secret)
	if err != nil {
		return err
	}

	client := buildClientFromCreateFlags(clientID, *secret, displayName, grantTypes, redirectURIs, scope, authorities, allowedProviders, requiredUserGroups, accessTokenValidity, refreshTokenValidity, autoApprove, allowPublic)

	createdClient, err := uaaClient.Client().CreateClient(client)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	return displayCreatedClient(createdClient)
}

func promptForClientSecretIfNeeded(secret *string) error {
	if *secret == "" {
		_, err := os.Stdout.WriteString("Client Secret: ")
		if err != nil {
			return fmt.Errorf("failed to write prompt: %w", err)
		}

		secretBytes, err := term.ReadPassword(syscall.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read client secret: %w", err)
		}

		*secret = string(secretBytes)

		_, _ = os.Stdout.WriteString("\n") // Add newline after password input
	}

	return nil
}

func buildClientFromCreateFlags(clientID, secret, displayName string, grantTypes, redirectURIs, scope, authorities, allowedProviders, requiredUserGroups []string, accessTokenValidity, refreshTokenValidity int64, autoApprove, allowPublic bool) uaa.Client {
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

	if autoApprove {
		client.AutoApproveRaw = true
	}

	return client
}

func displayCreatedClient(createdClient *uaa.Client) error {
	output := viper.GetString("output")
	switch output {
	case OutputFormatJSON:
		return displayCreatedClientJSON(createdClient)
	case OutputFormatYAML:
		return displayCreatedClientYAML(createdClient)
	default:
		return displayClientTable(createdClient, false)
	}
}

func displayCreatedClientJSON(createdClient *uaa.Client) error {
	displayClient := *createdClient
	displayClient.ClientSecret = Masked
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	err := encoder.Encode(displayClient)
	if err != nil {
		return fmt.Errorf("failed to encode client: %w", err)
	}

	return nil
}

func displayCreatedClientYAML(createdClient *uaa.Client) error {
	displayClient := *createdClient
	displayClient.ClientSecret = Masked
	encoder := yaml.NewEncoder(os.Stdout)

	err := encoder.Encode(displayClient)
	if err != nil {
		return fmt.Errorf("failed to encode client: %w", err)
	}

	return nil
}

// createUsersGetClientCommand creates the get client command.
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
				return constants.ErrNoUAAConfigured
			}

			// Create UAA client
			uaaClient, err := NewUAAClient(config)
			if err != nil {
				return fmt.Errorf("failed to create UAA client: %w", err)
			}

			if !uaaClient.IsAuthenticated() {
				return constants.ErrNotAuthenticated
			}

			// Get client
			client, err := uaaClient.Client().GetClient(clientID)
			if err != nil {
				return fmt.Errorf("failed to get client: %w", err)
			}

			// Display client
			output := viper.GetString("output")
			switch output {
			case OutputFormatJSON:
				if !showSecret {
					// Mask secret for display
					displayClient := *client
					displayClient.ClientSecret = Masked
					client = &displayClient
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")

				return encoder.Encode(client)
			case OutputFormatYAML:
				if !showSecret {
					// Mask secret for display
					displayClient := *client
					displayClient.ClientSecret = Masked
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

// createUsersListClientsCommand creates the list clients command.
func createUsersListClientsCommand() *cobra.Command {
	var (
		filter, sortBy    string
		sortOrder         string
		count, startIndex int
		all               bool
	)

	cmd := &cobra.Command{
		Use:   "list-clients",
		Short: "List OAuth clients",
		Long: `List all OAuth clients in the targeted UAA.

Client secrets are never displayed in list operations for security.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			config := loadConfig()

			if GetEffectiveUAAEndpoint(config) == "" {
				return constants.ErrNoUAAConfigured
			}

			// Create UAA client
			uaaClient, err := NewUAAClient(config)
			if err != nil {
				return fmt.Errorf("failed to create UAA client: %w", err)
			}

			if !uaaClient.IsAuthenticated() {
				return constants.ErrNotAuthenticated
			}

			// Convert sort order string to enum
			var uaaSortOrder uaa.SortOrder
			switch strings.ToLower(sortOrder) {
			case Descending, "desc":
				uaaSortOrder = uaa.SortDescending
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
				clients[i].ClientSecret = Masked
			}

			// Display clients
			output := viper.GetString("output")
			switch output {
			case OutputFormatJSON:
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")

				return encoder.Encode(clients)
			case OutputFormatYAML:
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
	cmd.Flags().IntVar(&count, "count", constants.StandardPageSize, "Number of results per page")
	cmd.Flags().IntVar(&startIndex, "start-index", 1, "Starting index for pagination")
	cmd.Flags().BoolVar(&all, "all", false, "Fetch all clients across all pages")

	return cmd
}

// createUsersUpdateClientCommand creates the update client command.
func createUsersUpdateClientCommand() *cobra.Command {
	var (
		displayName                                                                        string
		grantTypes, redirectURIs, scope, authorities, allowedProviders, requiredUserGroups []string
		accessTokenValidity, refreshTokenValidity                                          int64
		autoApprove, allowPublic                                                           *bool
		autoApproveStr, allowPublicStr                                                     string
	)

	cmd := &cobra.Command{
		Use:   "update-client <client-id>",
		Short: "Update OAuth client",
		Long: `Update an OAuth client registration.

Only the specified attributes will be updated. Unspecified attributes
will remain unchanged.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdateClientCommand(args[0], displayName, grantTypes, redirectURIs, scope, authorities, allowedProviders, requiredUserGroups, accessTokenValidity, refreshTokenValidity, autoApprove, allowPublic)
		},
	}

	setupUpdateClientFlags(cmd, &displayName, &grantTypes, &redirectURIs, &scope, &authorities, &allowedProviders, &requiredUserGroups, &accessTokenValidity, &refreshTokenValidity, &autoApproveStr, &allowPublicStr)

	// Pre-run to parse boolean flags
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		return parseUpdateClientBooleanFlags(autoApproveStr, allowPublicStr, &autoApprove, &allowPublic)
	}

	return cmd
}

func setupUpdateClientFlags(cmd *cobra.Command, displayName *string, grantTypes, redirectURIs, scope, authorities, allowedProviders, requiredUserGroups *[]string, accessTokenValidity, refreshTokenValidity *int64, autoApproveStr, allowPublicStr *string) {
	cmd.Flags().StringVar(displayName, "name", "", "Client display name")
	cmd.Flags().StringSliceVar(grantTypes, "authorized-grant-types", nil, "Authorized grant types")
	cmd.Flags().StringSliceVar(redirectURIs, "redirect-uri", nil, "Redirect URIs")
	cmd.Flags().StringSliceVar(scope, "scope", nil, "OAuth scopes")
	cmd.Flags().StringSliceVar(authorities, "authorities", nil, "Client authorities")
	cmd.Flags().StringSliceVar(allowedProviders, "allowed-providers", nil, "Allowed identity providers")
	cmd.Flags().StringSliceVar(requiredUserGroups, "required-user-groups", nil, "Required user groups")
	cmd.Flags().Int64Var(accessTokenValidity, "access-token-validity", 0, "Access token validity in seconds")
	cmd.Flags().Int64Var(refreshTokenValidity, "refresh-token-validity", 0, "Refresh token validity in seconds")

	// Use string flags for booleans to distinguish between false and unset
	cmd.Flags().StringVar(autoApproveStr, "auto-approve", "", "Auto-approve authorization requests (true/false)")
	cmd.Flags().StringVar(allowPublicStr, "allow-public", "", "Allow public client (true/false)")
}

func parseUpdateClientBooleanFlags(autoApproveStr, allowPublicStr string, autoApprove, allowPublic **bool) error {
	autoApproveParser := NewBooleanFlagParser(autoApproveStr, "auto-approve", constants.ErrInvalidAutoApprove)

	var err error

	*autoApprove, err = autoApproveParser.Parse()
	if err != nil {
		return err
	}

	allowPublicParser := NewBooleanFlagParser(allowPublicStr, "allow-public", constants.ErrInvalidAllowPublic)
	*allowPublic, err = allowPublicParser.Parse()

	return err
}

func runUpdateClientCommand(clientID, displayName string, grantTypes, redirectURIs, scope, authorities, allowedProviders, requiredUserGroups []string, accessTokenValidity, refreshTokenValidity int64, autoApprove, allowPublic *bool) error {
	config := loadConfig()

	if GetEffectiveUAAEndpoint(config) == "" {
		return constants.ErrNoUAAConfigured
	}

	// Create UAA client
	uaaClient, err := NewUAAClient(config)
	if err != nil {
		return fmt.Errorf("failed to create UAA client: %w", err)
	}

	if !uaaClient.IsAuthenticated() {
		return constants.ErrNotAuthenticated
	}

	// Get existing client
	existingClient, err := uaaClient.Client().GetClient(clientID)
	if err != nil {
		return fmt.Errorf("failed to get existing client: %w", err)
	}

	// Update specified fields
	updateClientFields(existingClient, displayName, grantTypes, redirectURIs, scope, authorities, allowedProviders, requiredUserGroups, accessTokenValidity, refreshTokenValidity, autoApprove, allowPublic)

	// Update client
	updatedClient, err := uaaClient.Client().UpdateClient(*existingClient)
	if err != nil {
		return fmt.Errorf("failed to update client: %w", err)
	}

	// Display updated client
	return displayUpdatedClient(updatedClient)
}

func updateClientFields(client *uaa.Client, displayName string, grantTypes, redirectURIs, scope, authorities, allowedProviders, requiredUserGroups []string, accessTokenValidity, refreshTokenValidity int64, autoApprove, allowPublic *bool) {
	if displayName != "" {
		client.DisplayName = displayName
	}

	if len(grantTypes) > 0 {
		client.AuthorizedGrantTypes = grantTypes
	}

	if len(redirectURIs) > 0 {
		client.RedirectURI = redirectURIs
	}

	if len(scope) > 0 {
		client.Scope = scope
	}

	if len(authorities) > 0 {
		client.Authorities = authorities
	}

	if len(allowedProviders) > 0 {
		client.AllowedProviders = allowedProviders
	}

	if len(requiredUserGroups) > 0 {
		client.RequiredUserGroups = requiredUserGroups
	}

	if accessTokenValidity > 0 {
		client.AccessTokenValidity = accessTokenValidity
	}

	if refreshTokenValidity > 0 {
		client.RefreshTokenValidity = refreshTokenValidity
	}

	if autoApprove != nil {
		client.AutoApproveRaw = *autoApprove
	}

	if allowPublic != nil {
		client.AllowPublic = *allowPublic
	}
}

func displayUpdatedClient(updatedClient *uaa.Client) error {
	output := viper.GetString("output")
	switch output {
	case OutputFormatJSON:
		return displayUpdatedClientJSON(updatedClient)
	case OutputFormatYAML:
		return displayUpdatedClientYAML(updatedClient)
	default:
		return displayClientTable(updatedClient, false)
	}
}

func displayUpdatedClientJSON(updatedClient *uaa.Client) error {
	displayClient := *updatedClient
	displayClient.ClientSecret = Masked
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	err := encoder.Encode(displayClient)
	if err != nil {
		return fmt.Errorf("failed to encode client: %w", err)
	}

	return nil
}

func displayUpdatedClientYAML(updatedClient *uaa.Client) error {
	displayClient := *updatedClient
	displayClient.ClientSecret = Masked
	encoder := yaml.NewEncoder(os.Stdout)

	err := encoder.Encode(displayClient)
	if err != nil {
		return fmt.Errorf("failed to encode client: %w", err)
	}

	return nil
}

// createUsersSetClientSecretCommand creates the set client secret command.
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
				return constants.ErrNoUAAConfigured
			}

			// Create UAA client
			uaaClient, err := NewUAAClient(config)
			if err != nil {
				return fmt.Errorf("failed to create UAA client: %w", err)
			}

			if !uaaClient.IsAuthenticated() {
				return constants.ErrNotAuthenticated
			}

			// Prompt for secret if not provided
			if secret == "" {
				_, err := os.Stdout.WriteString("New Client Secret: ")
				if err != nil {
					return fmt.Errorf("failed to write prompt: %w", err)
				}
				secretBytes, err := term.ReadPassword(syscall.Stdin)
				if err != nil {
					return fmt.Errorf("failed to read client secret: %w", err)
				}
				secret = string(secretBytes)
				_, _ = os.Stdout.WriteString("\n") // Add newline after password input
			}

			// Update client secret
			err = uaaClient.Client().ChangeClientSecret(clientID, secret)
			if err != nil {
				return fmt.Errorf("failed to update client secret: %w", err)
			}

			_, _ = fmt.Fprintf(os.Stdout, "Successfully updated secret for client '%s'\n", clientID)

			return nil
		},
	}

	cmd.Flags().StringVar(&secret, "secret", "", "New client secret")

	return cmd
}

// createUsersDeleteClientCommand creates the delete client command.
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
				return constants.ErrNoUAAConfigured
			}

			// Create UAA client
			uaaClient, err := NewUAAClient(config)
			if err != nil {
				return fmt.Errorf("failed to create UAA client: %w", err)
			}

			if !uaaClient.IsAuthenticated() {
				return constants.ErrNotAuthenticated
			}

			// Get client details for confirmation
			client, err := uaaClient.Client().GetClient(clientID)
			if err != nil {
				return fmt.Errorf("failed to get client: %w", err)
			}

			// Confirm deletion unless --force is used
			if !force {
				_, _ = fmt.Fprintf(os.Stdout, "Are you sure you want to delete client '%s'", clientID)
				if client.DisplayName != "" {
					_, _ = fmt.Fprintf(os.Stdout, " (%s)", client.DisplayName)
				}
				_, err := os.Stdout.WriteString("? [y/N]: ")
				if err != nil {
					return fmt.Errorf("failed to write prompt: %w", err)
				}
				var response string
				_, _ = fmt.Scanln(&response)
				if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
					_, _ = os.Stdout.WriteString("Client deletion cancelled\n")

					return nil
				}
			}

			// Delete client
			_, err = uaaClient.Client().DeleteClient(clientID)
			if err != nil {
				return fmt.Errorf("failed to delete client: %w", err)
			}

			_, _ = fmt.Fprintf(os.Stdout, "Client '%s' has been deleted\n", clientID)

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
		_ = table.Append("Client Secret", Masked)
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
	_ = table.Append("Allow Public", strconv.FormatBool(client.AllowPublic))

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
		if len(grantTypes) > constants.GrantTypesDisplayLength {
			grantTypes = grantTypes[:30] + "..."
		}

		scope := strings.Join(client.Scope, ",")
		if len(scope) > constants.GrantTypesDisplayLength {
			scope = scope[:30] + "..."
		}

		autoApprove := fmt.Sprintf("%v", client.AutoApprove())

		_ = table.Append(clientID, displayName, grantTypes, scope, autoApprove)
	}

	_ = table.Render()

	return nil
}
