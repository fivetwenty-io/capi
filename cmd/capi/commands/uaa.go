package commands

import (
	"github.com/spf13/cobra"
)

// NewUAACommand creates the uaa command group with UAA integration.
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
  capi uaa token get-client-credentials --client-id admin --client-secret secret

  # Check authentication status
  capi uaa context

  # User management (new structure)
  capi uaa user create john.doe --email john@example.com
  capi uaa user list --filter 'active eq true'
  capi uaa user get john.doe

  # Group management (new structure)
  capi uaa group create developers --description "Development team"
  capi uaa group add-member developers john.doe

  # Client management (new structure)
  capi uaa client create myapp --secret mysecret
  capi uaa client list

  # Token management (new structure)
  capi uaa token get-password --username john.doe

  # Get help for any specific command group
  capi uaa user --help
  capi uaa token --help`,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	addUAASubCommands(cmd)
	addUAAContextCommands(cmd)
	addUAAUtilityCommands(cmd)
	addLegacyUAACommands(cmd)

	return cmd
}

// addUAASubCommands adds the main UAA resource management sub-commands.
func addUAASubCommands(cmd *cobra.Command) {
	cmd.AddCommand(NewUAAUserCommand())
	cmd.AddCommand(NewUAAGroupCommand())
	cmd.AddCommand(NewUAAClientCommand())
	cmd.AddCommand(NewUAATokenCommand())
	cmd.AddCommand(NewUAABatchCommand())
	cmd.AddCommand(NewUAAIntegrationCommand())
}

// addUAAContextCommands adds context management commands.
func addUAAContextCommands(cmd *cobra.Command) {
	cmd.AddCommand(createUsersContextCommand())
	cmd.AddCommand(createUsersTargetCommand())
	cmd.AddCommand(createUsersInfoCommand())
	cmd.AddCommand(createUsersVersionCommand())
}

// addUAAUtilityCommands adds utility commands.
func addUAAUtilityCommands(cmd *cobra.Command) {
	cmd.AddCommand(createUsersCurlCommand())
	cmd.AddCommand(createUsersUserinfoCommand())
}

// addLegacyUAACommands adds all legacy UAA commands as hidden commands for backward compatibility.
func addLegacyUAACommands(cmd *cobra.Command) {
	addLegacyUserCommands(cmd)
	addLegacyGroupCommands(cmd)
	addLegacyClientCommands(cmd)
	addLegacyTokenCommands(cmd)
	addLegacyUtilityCommands(cmd)
}

// addLegacyUserCommands adds legacy user commands.
func addLegacyUserCommands(cmd *cobra.Command) {
	legacyUserCmd := createUsersCreateUserCommand()
	legacyUserCmd.Hidden = true
	cmd.AddCommand(legacyUserCmd)

	legacyGetUserCmd := createUsersGetUserCommand()
	legacyGetUserCmd.Hidden = true
	cmd.AddCommand(legacyGetUserCmd)

	legacyListUsersCmd := createUsersListUsersCommand()
	legacyListUsersCmd.Hidden = true
	cmd.AddCommand(legacyListUsersCmd)

	legacyUpdateUserCmd := createUsersUpdateUserCommand()
	legacyUpdateUserCmd.Hidden = true
	cmd.AddCommand(legacyUpdateUserCmd)

	legacyActivateUserCmd := createUsersActivateUserCommand()
	legacyActivateUserCmd.Hidden = true
	cmd.AddCommand(legacyActivateUserCmd)

	legacyDeactivateUserCmd := createUsersDeactivateUserCommand()
	legacyDeactivateUserCmd.Hidden = true
	cmd.AddCommand(legacyDeactivateUserCmd)

	legacyDeleteUserCmd := createUsersDeleteUserCommand()
	legacyDeleteUserCmd.Hidden = true
	cmd.AddCommand(legacyDeleteUserCmd)
}

// addLegacyGroupCommands adds legacy group commands.
func addLegacyGroupCommands(cmd *cobra.Command) {
	legacyCreateGroupCmd := createUsersCreateGroupCommand()
	legacyCreateGroupCmd.Hidden = true
	cmd.AddCommand(legacyCreateGroupCmd)

	legacyGetGroupCmd := createUsersGetGroupCommand()
	legacyGetGroupCmd.Hidden = true
	cmd.AddCommand(legacyGetGroupCmd)

	legacyListGroupsCmd := createUsersListGroupsCommand()
	legacyListGroupsCmd.Hidden = true
	cmd.AddCommand(legacyListGroupsCmd)

	legacyAddMemberCmd := createUsersAddMemberCommand()
	legacyAddMemberCmd.Hidden = true
	cmd.AddCommand(legacyAddMemberCmd)

	legacyRemoveMemberCmd := createUsersRemoveMemberCommand()
	legacyRemoveMemberCmd.Hidden = true
	cmd.AddCommand(legacyRemoveMemberCmd)

	legacyMapGroupCmd := createUsersMapGroupCommand()
	legacyMapGroupCmd.Hidden = true
	cmd.AddCommand(legacyMapGroupCmd)

	legacyUnmapGroupCmd := createUsersUnmapGroupCommand()
	legacyUnmapGroupCmd.Hidden = true
	cmd.AddCommand(legacyUnmapGroupCmd)

	legacyListGroupMappingsCmd := createUsersListGroupMappingsCommand()
	legacyListGroupMappingsCmd.Hidden = true
	cmd.AddCommand(legacyListGroupMappingsCmd)
}

// addLegacyClientCommands adds legacy client commands.
func addLegacyClientCommands(cmd *cobra.Command) {
	legacyCreateClientCmd := createUsersCreateClientCommand()
	legacyCreateClientCmd.Hidden = true
	cmd.AddCommand(legacyCreateClientCmd)

	legacyGetClientCmd := createUsersGetClientCommand()
	legacyGetClientCmd.Hidden = true
	cmd.AddCommand(legacyGetClientCmd)

	legacyListClientsCmd := createUsersListClientsCommand()
	legacyListClientsCmd.Hidden = true
	cmd.AddCommand(legacyListClientsCmd)

	legacyUpdateClientCmd := createUsersUpdateClientCommand()
	legacyUpdateClientCmd.Hidden = true
	cmd.AddCommand(legacyUpdateClientCmd)

	legacySetClientSecretCmd := createUsersSetClientSecretCommand()
	legacySetClientSecretCmd.Hidden = true
	cmd.AddCommand(legacySetClientSecretCmd)

	legacyDeleteClientCmd := createUsersDeleteClientCommand()
	legacyDeleteClientCmd.Hidden = true
	cmd.AddCommand(legacyDeleteClientCmd)
}

// addLegacyTokenCommands adds legacy token commands.
func addLegacyTokenCommands(cmd *cobra.Command) {
	legacyGetAuthcodeTokenCmd := createUsersGetAuthcodeTokenCommand()
	legacyGetAuthcodeTokenCmd.Hidden = true
	cmd.AddCommand(legacyGetAuthcodeTokenCmd)

	legacyGetClientCredentialsTokenCmd := createUsersGetClientCredentialsTokenCommand()
	legacyGetClientCredentialsTokenCmd.Hidden = true
	cmd.AddCommand(legacyGetClientCredentialsTokenCmd)

	legacyGetImplicitTokenCmd := createUsersGetImplicitTokenCommand()
	legacyGetImplicitTokenCmd.Hidden = true
	cmd.AddCommand(legacyGetImplicitTokenCmd)

	legacyGetPasswordTokenCmd := createUsersGetPasswordTokenCommand()
	legacyGetPasswordTokenCmd.Hidden = true
	cmd.AddCommand(legacyGetPasswordTokenCmd)

	legacyRefreshTokenCmd := createUsersRefreshTokenCommand()
	legacyRefreshTokenCmd.Hidden = true
	cmd.AddCommand(legacyRefreshTokenCmd)

	legacyGetTokenKeyCmd := createUsersGetTokenKeyCommand()
	legacyGetTokenKeyCmd.Hidden = true
	cmd.AddCommand(legacyGetTokenKeyCmd)

	legacyGetTokenKeysCmd := createUsersGetTokenKeysCommand()
	legacyGetTokenKeysCmd.Hidden = true
	cmd.AddCommand(legacyGetTokenKeysCmd)
}

// addLegacyUtilityCommands adds legacy utility commands.
func addLegacyUtilityCommands(cmd *cobra.Command) {
	legacyBatchImportCmd := createUsersBatchImportCommand()
	legacyBatchImportCmd.Hidden = true
	cmd.AddCommand(legacyBatchImportCmd)

	legacyPerformanceCmd := createUsersPerformanceCommand()
	legacyPerformanceCmd.Hidden = true
	cmd.AddCommand(legacyPerformanceCmd)

	legacyCacheCmd := createUsersCacheCommand()
	legacyCacheCmd.Hidden = true
	cmd.AddCommand(legacyCacheCmd)

	legacyCompatibilityCmd := createUsersCompatibilityCommand()
	legacyCompatibilityCmd.Hidden = true
	cmd.AddCommand(legacyCompatibilityCmd)

	legacyCFIntegrationCmd := createUsersCFIntegrationCommand()
	legacyCFIntegrationCmd.Hidden = true
	cmd.AddCommand(legacyCFIntegrationCmd)
}
