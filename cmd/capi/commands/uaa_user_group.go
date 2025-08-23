package commands

import (
	"github.com/spf13/cobra"
)

// NewUAAUserCommand creates the user sub-command group
func NewUAAUserCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Manage UAA users",
		Long: `Manage UAA users with CRUD operations.

This command group provides comprehensive user management capabilities including:
- Creating new users
- Retrieving user information
- Listing users with filtering
- Updating user attributes
- Activating/deactivating users
- Deleting users`,
		Example: `  # Create a new user
  capi uaa user create john.doe --email john@example.com

  # Get user information
  capi uaa user get john.doe

  # List all active users
  capi uaa user list --filter 'active eq true'

  # Update user email
  capi uaa user update john.doe --email newemail@example.com

  # Deactivate a user
  capi uaa user deactivate john.doe`,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	// Add user management commands with new naming
	cmd.AddCommand(createUserCreateCommand())
	cmd.AddCommand(createUserGetCommand())
	cmd.AddCommand(createUserListCommand())
	cmd.AddCommand(createUserUpdateCommand())
	cmd.AddCommand(createUserActivateCommand())
	cmd.AddCommand(createUserDeactivateCommand())
	cmd.AddCommand(createUserDeleteCommand())

	return cmd
}

func createUserCreateCommand() *cobra.Command {
	cmd := createUsersCreateUserCommand()
	cmd.Use = "create <username>"
	return cmd
}

func createUserGetCommand() *cobra.Command {
	cmd := createUsersGetUserCommand()
	cmd.Use = "get <username>"
	return cmd
}

func createUserListCommand() *cobra.Command {
	cmd := createUsersListUsersCommand()
	cmd.Use = "list"
	return cmd
}

func createUserUpdateCommand() *cobra.Command {
	cmd := createUsersUpdateUserCommand()
	cmd.Use = "update <username>"
	return cmd
}

func createUserActivateCommand() *cobra.Command {
	cmd := createUsersActivateUserCommand()
	cmd.Use = "activate <username>"
	return cmd
}

func createUserDeactivateCommand() *cobra.Command {
	cmd := createUsersDeactivateUserCommand()
	cmd.Use = "deactivate <username>"
	return cmd
}

func createUserDeleteCommand() *cobra.Command {
	cmd := createUsersDeleteUserCommand()
	cmd.Use = "delete <username>"
	return cmd
}
