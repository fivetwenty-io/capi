package commands

import (
	"github.com/spf13/cobra"
)

// NewUAACommand creates the uaa command group with UAA integration
func NewUAACommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "uaa",
		Aliases: []string{},
		Short:   "Manage users via UAA",
		Long: `Manage Users & Auth{z/n} via UAA (User Account and Authentication).

This command group provides comprehensive user management capabilities including:
- Context and authentication management
- Token operations (OAuth2 flows)
- User CRUD operations
- Group management
- OAuth client management
- Direct UAA API access

All commands interact with the UAA service to manage users, groups, and authentication.`,
		Example: `  # Quick start - set UAA target and authenticate
  capi uaa target https://uaa.your-domain.com
  capi uaa get-client-credentials-token --client-id admin --client-secret secret

  # Check authentication status
  capi uaa context

  # User management
  capi uaa create-user john.doe --email john@example.com
  capi uaa list-users --filter 'active eq true'
  capi uaa get-user john.doe

  # Group management
  capi uaa create-group developers --description "Development team"
  capi uaa add-member developers john.doe

  # Get help for any specific command
  capi uaa create-user --help`,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	// Context Management Commands
	cmd.AddCommand(createUsersContextCommand())
	cmd.AddCommand(createUsersTargetCommand())
	cmd.AddCommand(createUsersInfoCommand())
	cmd.AddCommand(createUsersVersionCommand())

	// Token Management Commands
	cmd.AddCommand(createUsersGetAuthcodeTokenCommand())
	cmd.AddCommand(createUsersGetClientCredentialsTokenCommand())
	cmd.AddCommand(createUsersGetImplicitTokenCommand())
	cmd.AddCommand(createUsersGetPasswordTokenCommand())
	cmd.AddCommand(createUsersRefreshTokenCommand())
	cmd.AddCommand(createUsersGetTokenKeyCommand())
	cmd.AddCommand(createUsersGetTokenKeysCommand())

	// User Management Commands
	cmd.AddCommand(createUsersCreateUserCommand())
	cmd.AddCommand(createUsersGetUserCommand())
	cmd.AddCommand(createUsersListUsersCommand())
	cmd.AddCommand(createUsersUpdateUserCommand())
	cmd.AddCommand(createUsersActivateUserCommand())
	cmd.AddCommand(createUsersDeactivateUserCommand())
	cmd.AddCommand(createUsersDeleteUserCommand())

	// Group Management Commands
	cmd.AddCommand(createUsersCreateGroupCommand())
	cmd.AddCommand(createUsersGetGroupCommand())
	cmd.AddCommand(createUsersListGroupsCommand())
	cmd.AddCommand(createUsersAddMemberCommand())
	cmd.AddCommand(createUsersRemoveMemberCommand())
	cmd.AddCommand(createUsersMapGroupCommand())
	cmd.AddCommand(createUsersUnmapGroupCommand())
	cmd.AddCommand(createUsersListGroupMappingsCommand())

	// Client Management Commands
	cmd.AddCommand(createUsersCreateClientCommand())
	cmd.AddCommand(createUsersGetClientCommand())
	cmd.AddCommand(createUsersListClientsCommand())
	cmd.AddCommand(createUsersUpdateClientCommand())
	cmd.AddCommand(createUsersSetClientSecretCommand())
	cmd.AddCommand(createUsersDeleteClientCommand())

	// Utility Commands
	cmd.AddCommand(createUsersCurlCommand())
	cmd.AddCommand(createUsersUserinfoCommand())

	// Performance and Batch Commands
	cmd.AddCommand(createUsersBatchImportCommand())
	cmd.AddCommand(createUsersPerformanceCommand())
	cmd.AddCommand(createUsersCacheCommand())

	// Compatibility and Integration Commands
	cmd.AddCommand(createUsersCompatibilityCommand())
	cmd.AddCommand(createUsersCFIntegrationCommand())

	return cmd
}
