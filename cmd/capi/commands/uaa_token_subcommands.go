package commands

import (
	"github.com/spf13/cobra"
)

// NewUAATokenCommand creates the token sub-command group
func NewUAATokenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Manage UAA OAuth tokens",
		Long: `Manage UAA OAuth tokens and token operations.

This command group provides comprehensive token management capabilities including:
- Obtaining tokens via various OAuth2 flows
- Refreshing expired tokens
- Retrieving token validation keys
- Token introspection and validation`,
		Example: `  # Get token using client credentials flow
  capi uaa token get-client-credentials --client-id admin --client-secret secret

  # Get token using authorization code flow
  capi uaa token get-authcode --client-id myapp --client-secret secret

  # Get token using password flow
  capi uaa token get-password --client-id myapp --username john.doe

  # Refresh an existing token
  capi uaa token refresh --refresh-token <refresh-token>

  # Get token signing keys
  capi uaa token get-keys`,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	// Add token management commands with new naming
	cmd.AddCommand(createTokenGetAuthcodeCommand())
	cmd.AddCommand(createTokenGetClientCredentialsCommand())
	cmd.AddCommand(createTokenGetPasswordCommand())
	cmd.AddCommand(createTokenGetImplicitCommand())
	cmd.AddCommand(createTokenRefreshCommand())
	cmd.AddCommand(createTokenGetKeyCommand())
	cmd.AddCommand(createTokenGetKeysCommand())

	return cmd
}

func createTokenGetAuthcodeCommand() *cobra.Command {
	cmd := createUsersGetAuthcodeTokenCommand()
	cmd.Use = "get-authcode"
	return cmd
}

func createTokenGetClientCredentialsCommand() *cobra.Command {
	cmd := createUsersGetClientCredentialsTokenCommand()
	cmd.Use = "get-client-credentials"
	return cmd
}

func createTokenGetPasswordCommand() *cobra.Command {
	cmd := createUsersGetPasswordTokenCommand()
	cmd.Use = "get-password"
	return cmd
}

func createTokenGetImplicitCommand() *cobra.Command {
	cmd := createUsersGetImplicitTokenCommand()
	cmd.Use = "get-implicit"
	return cmd
}

func createTokenRefreshCommand() *cobra.Command {
	cmd := createUsersRefreshTokenCommand()
	cmd.Use = "refresh"
	return cmd
}

func createTokenGetKeyCommand() *cobra.Command {
	cmd := createUsersGetTokenKeyCommand()
	cmd.Use = "get-key"
	return cmd
}

func createTokenGetKeysCommand() *cobra.Command {
	cmd := createUsersGetTokenKeysCommand()
	cmd.Use = "get-keys"
	return cmd
}
