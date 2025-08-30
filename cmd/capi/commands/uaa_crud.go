package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/cloudfoundry/go-uaa"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

// createUsersCreateUserCommand creates the create user command
func createUsersCreateUserCommand() *cobra.Command {
	var email, password, givenName, familyName, phoneNumber, origin string
	var active, verified bool

	cmd := &cobra.Command{
		Use:     "create-user <username>",
		Aliases: []string{"create", "add-user", "new-user"},
		Short:   "Create a new user",
		Long: `Create a new user in the UAA database.

The username is required as a positional argument. Other user attributes
can be specified using flags. If not provided via flags, you will be
prompted for required information.`,
		Example: `  # Create user with minimal information
  capi uaa create-user john.doe --email john@example.com

  # Create user with complete profile
  capi uaa create-user jane.smith \
    --email jane@example.com \
    --given-name Jane \
    --family-name Smith \
    --phone-number "+1-555-0123" \
    --verified \
    --active

  # Create user from specific identity provider
  capi uaa create-user ldap.user --email user@company.com --origin ldap`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := loadConfig()
			username := args[0]

			if GetEffectiveUAAEndpoint(config) == "" {
				err := NewEnhancedError("create user", fmt.Errorf("no UAA endpoint configured"))
				err.AddSuggestion("Run 'capi uaa target <url>' to set UAA endpoint")
				return err
			}

			// Create UAA client with progress indicator
			var uaaClient *UAAClientWrapper
			err := WrapWithProgress("Connecting to UAA", func() error {
				var err error
				uaaClient, err = NewUAAClient(config)
				return err
			})
			if err != nil {
				return CreateCommonUAAError("create user", err, config.UAAEndpoint)
			}

			if !uaaClient.IsAuthenticated() {
				return CreateCommonUAAError("create user", fmt.Errorf("not authenticated"), config.UAAEndpoint)
			}

			// Prompt for required fields if not provided
			if email == "" {
				fmt.Print("Email: ")
				_, _ = fmt.Scanln(&email)
			}
			if password == "" {
				fmt.Print("Password: ")
				passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
				if err != nil {
					return fmt.Errorf("failed to read password: %w", err)
				}
				password = string(passwordBytes)
				fmt.Println() // Add newline after password input
			}

			// Build user object
			user := uaa.User{
				Username: username,
				Password: password,
				Origin:   origin,
				Emails: []uaa.Email{
					{
						Value:   email,
						Primary: &[]bool{true}[0], // Pointer to true
					},
				},
				Active:   &active,
				Verified: &verified,
			}

			// Add name if provided
			if givenName != "" || familyName != "" {
				user.Name = &uaa.UserName{
					GivenName:  givenName,
					FamilyName: familyName,
				}
			}

			// Add phone number if provided
			if phoneNumber != "" {
				user.PhoneNumbers = []uaa.PhoneNumber{
					{
						Value: phoneNumber,
					},
				}
			}

			// Create user with progress indicator and performance tracking
			var createdUser *uaa.User
			err = WrapWithProgress(fmt.Sprintf("Creating user '%s'", username), func() error {
				return WithPerformanceTracking("create-user", func() error {
					var createErr error
					createdUser, createErr = uaaClient.Client().CreateUser(user)
					if createErr == nil {
						// Invalidate cache for user lookups
						InvalidateUserCache(username)
					}
					return createErr
				})
			})
			if err != nil {
				enhancedErr := NewEnhancedError("create user", err)
				enhancedErr.AddContext("UAA Endpoint", config.UAAEndpoint)
				enhancedErr.AddContext("Username", username)
				enhancedErr.AddContext("Email", email)
				enhancedErr.AddSuggestion("Verify that the username is unique")
				enhancedErr.AddSuggestion("Check that all required fields are provided")
				enhancedErr.AddSuggestion("Ensure your client has 'scim.write' authority")
				return enhancedErr
			}

			// Display created user
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(createdUser)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(createdUser)
			default:
				return displayUserTable(createdUser)
			}
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "User's email address (required)")
	cmd.Flags().StringVar(&password, "password", "", "User's password")
	cmd.Flags().StringVar(&givenName, "given-name", "", "User's given name (first name)")
	cmd.Flags().StringVar(&familyName, "family-name", "", "User's family name (last name)")
	cmd.Flags().StringVar(&phoneNumber, "phone-number", "", "User's phone number")
	cmd.Flags().StringVar(&origin, "origin", "uaa", "Identity provider origin")
	cmd.Flags().BoolVar(&active, "active", true, "User is active")
	cmd.Flags().BoolVar(&verified, "verified", false, "User is verified")

	return cmd
}

// createUsersGetUserCommand creates the get user command
func createUsersGetUserCommand() *cobra.Command {
	var attributes string

	cmd := &cobra.Command{
		Use:     "get-user <username>",
		Aliases: []string{"get", "show-user", "user"},
		Short:   "Get user details",
		Long: `Look up a user by username and display detailed information.

The command will search for the user by username and display all available
user attributes including groups, metadata, and authentication information.`,
		Example: `  # Get complete user information
  capi uaa get-user john.doe

  # Get specific user attributes only
  capi uaa get-user john.doe --attributes userName,email,active

  # Get user in JSON format for scripting
  capi uaa get-user john.doe --output json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := loadConfig()
			username := args[0]

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

			// Get user by username with caching
			var user *uaa.User
			err = WithPerformanceTracking("get-user", func() error {
				var getUserErr error
				user, getUserErr = CachedUserLookup(uaaClient, username)
				return getUserErr
			})
			if err != nil {
				return fmt.Errorf("failed to get user: %w", err)
			}

			// Display user
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(user)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(user)
			default:
				return displayUserTable(user)
			}
		},
	}

	cmd.Flags().StringVar(&attributes, "attributes", "", "Comma-separated list of attributes to return")

	return cmd
}

// createUsersListUsersCommand creates the list users command
func createUsersListUsersCommand() *cobra.Command {
	var filter, sortBy, attributes string
	var sortOrder string
	var count, startIndex int
	var all bool

	cmd := &cobra.Command{
		Use:     "list-users",
		Aliases: []string{"list", "users", "ls"},
		Short:   "List users",
		Long: `Search and list users with optional SCIM filters.

SCIM filters allow complex queries on user attributes. Examples:
- userName eq "john@example.com"
- active eq true
- origin eq "ldap"
- meta.created gt "2023-01-01T00:00:00.000Z"`,
		Example: `  # List all active users
  capi uaa list-users --filter 'active eq true'

  # List users from specific domain
  capi uaa list-users --filter 'email co "example.com"'

  # List recent users with specific attributes
  capi uaa list-users \
    --filter 'meta.created gt "2023-01-01T00:00:00.000Z"' \
    --attributes userName,email,meta.created \
    --sort-by meta.created \
    --sort-order desc

  # Get all users (auto-pagination)
  capi uaa list-users --all`,
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

			var users []uaa.User
			err = WithPerformanceTracking("list-users", func() error {
				var listErr error
				if all {
					// Use optimized pagination for better performance
					pagination := NewOptimizedPagination(uaaClient)
					users, listErr = pagination.GetAllUsers(filter, sortBy, attributes, uaaSortOrder)
				} else {
					// Get users with pagination
					users, _, listErr = uaaClient.Client().ListUsers(filter, sortBy, attributes, uaaSortOrder, startIndex, count)
				}
				return listErr
			})

			if err != nil {
				return fmt.Errorf("failed to list users: %w", err)
			}

			// Display users
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(users)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(users)
			default:
				return displayUsersTable(users)
			}
		},
	}

	cmd.Flags().StringVar(&filter, "filter", "", "SCIM filter expression")
	cmd.Flags().StringVar(&sortBy, "sort-by", "", "Attribute to sort by")
	cmd.Flags().StringVar(&sortOrder, "sort-order", "ascending", "Sort order (ascending, descending)")
	cmd.Flags().StringVar(&attributes, "attributes", "", "Comma-separated list of attributes to return")
	cmd.Flags().IntVar(&count, "count", 50, "Number of results per page")
	cmd.Flags().IntVar(&startIndex, "start-index", 1, "Starting index for pagination")
	cmd.Flags().BoolVar(&all, "all", false, "Fetch all users across all pages")

	return cmd
}

// createUsersUpdateUserCommand creates the update user command
func createUsersUpdateUserCommand() *cobra.Command {
	var email, givenName, familyName, phoneNumber string
	var active, verified *bool

	cmd := &cobra.Command{
		Use:   "update-user <username>",
		Short: "Update user attributes",
		Long: `Update attributes for an existing user.

Only the specified attributes will be updated. Unspecified attributes
will remain unchanged.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := loadConfig()
			username := args[0]

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

			// Get existing user
			existingUser, err := uaaClient.Client().GetUserByUsername(username, "", "")
			if err != nil {
				return fmt.Errorf("failed to get existing user: %w", err)
			}

			// Update specified fields
			if email != "" {
				existingUser.Emails = []uaa.Email{
					{
						Value:   email,
						Primary: &[]bool{true}[0],
					},
				}
			}

			if givenName != "" || familyName != "" {
				if existingUser.Name == nil {
					existingUser.Name = &uaa.UserName{}
				}
				if givenName != "" {
					existingUser.Name.GivenName = givenName
				}
				if familyName != "" {
					existingUser.Name.FamilyName = familyName
				}
			}

			if phoneNumber != "" {
				existingUser.PhoneNumbers = []uaa.PhoneNumber{
					{
						Value: phoneNumber,
					},
				}
			}

			if active != nil {
				existingUser.Active = active
			}

			if verified != nil {
				existingUser.Verified = verified
			}

			// Update user
			updatedUser, err := uaaClient.Client().UpdateUser(*existingUser)
			if err != nil {
				return fmt.Errorf("failed to update user: %w", err)
			}

			// Display updated user
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(updatedUser)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(updatedUser)
			default:
				return displayUserTable(updatedUser)
			}
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "User's email address")
	cmd.Flags().StringVar(&givenName, "given-name", "", "User's given name (first name)")
	cmd.Flags().StringVar(&familyName, "family-name", "", "User's family name (last name)")
	cmd.Flags().StringVar(&phoneNumber, "phone-number", "", "User's phone number")

	// Use string flags for booleans to distinguish between false and unset
	var activeStr, verifiedStr string
	cmd.Flags().StringVar(&activeStr, "active", "", "User is active (true/false)")
	cmd.Flags().StringVar(&verifiedStr, "verified", "", "User is verified (true/false)")

	// Pre-run to parse boolean flags
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if activeStr != "" {
			val, err := strconv.ParseBool(activeStr)
			if err != nil {
				return fmt.Errorf("invalid value for --active: %s", activeStr)
			}
			active = &val
		}
		if verifiedStr != "" {
			val, err := strconv.ParseBool(verifiedStr)
			if err != nil {
				return fmt.Errorf("invalid value for --verified: %s", verifiedStr)
			}
			verified = &val
		}
		return nil
	}

	return cmd
}

// createUsersActivateUserCommand creates the activate user command
func createUsersActivateUserCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "activate-user <username>",
		Short: "Activate a user account",
		Long:  "Activate a user account by username",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := loadConfig()
			username := args[0]

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

			// Get user to get ID and version
			user, err := uaaClient.Client().GetUserByUsername(username, "", "")
			if err != nil {
				return fmt.Errorf("failed to get user: %w", err)
			}

			// Activate user (requires user ID and version)
			version := 0
			if user.Meta != nil {
				version = user.Meta.Version
			}

			err = uaaClient.Client().ActivateUser(user.ID, version)
			if err != nil {
				return fmt.Errorf("failed to activate user: %w", err)
			}

			fmt.Printf("User '%s' has been activated\n", username)
			return nil
		},
	}
}

// createUsersDeactivateUserCommand creates the deactivate user command
func createUsersDeactivateUserCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "deactivate-user <username>",
		Short: "Deactivate a user account",
		Long:  "Deactivate a user account by username",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := loadConfig()
			username := args[0]

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

			// Get user to get ID and version
			user, err := uaaClient.Client().GetUserByUsername(username, "", "")
			if err != nil {
				return fmt.Errorf("failed to get user: %w", err)
			}

			// Deactivate user (requires user ID and version)
			version := 0
			if user.Meta != nil {
				version = user.Meta.Version
			}

			err = uaaClient.Client().DeactivateUser(user.ID, version)
			if err != nil {
				return fmt.Errorf("failed to deactivate user: %w", err)
			}

			fmt.Printf("User '%s' has been deactivated\n", username)
			return nil
		},
	}
}

// createUsersDeleteUserCommand creates the delete user command
func createUsersDeleteUserCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:     "delete-user <username>",
		Aliases: []string{"delete", "remove-user", "rm"},
		Short:   "Delete a user",
		Long:    "Delete a user by username. This action is irreversible.",
		Example: `  # Delete user with confirmation prompt
  capi uaa delete-user john.doe

  # Force delete without confirmation
  capi uaa delete-user john.doe --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := loadConfig()
			username := args[0]

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

			// Get user to get ID and display info
			user, err := uaaClient.Client().GetUserByUsername(username, "", "")
			if err != nil {
				return fmt.Errorf("failed to get user: %w", err)
			}

			// Confirm deletion unless --force is used
			if !ConfirmAction(fmt.Sprintf("Are you sure you want to delete user '%s' (ID: %s)?", username, user.ID), force) {
				fmt.Println("User deletion cancelled")
				return nil
			}

			// Delete user
			_, err = uaaClient.Client().DeleteUser(user.ID)
			if err != nil {
				return fmt.Errorf("failed to delete user: %w", err)
			}

			fmt.Printf("User '%s' has been deleted\n", username)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force deletion without confirmation")

	return cmd
}

// Helper functions for user display

func displayUserTable(user *uaa.User) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header("Property", "Value")

	if user.ID != "" {
		_ = table.Append("ID", user.ID)
	}
	if user.Username != "" {
		_ = table.Append("Username", user.Username)
	}
	if user.Origin != "" {
		_ = table.Append("Origin", user.Origin)
	}

	// Display name
	if user.Name != nil {
		if user.Name.GivenName != "" {
			_ = table.Append("Given Name", user.Name.GivenName)
		}
		if user.Name.FamilyName != "" {
			_ = table.Append("Family Name", user.Name.FamilyName)
		}
	}

	// Display primary email
	if len(user.Emails) > 0 {
		_ = table.Append("Email", user.Emails[0].Value)
	}

	// Display phone numbers
	if len(user.PhoneNumbers) > 0 {
		_ = table.Append("Phone", user.PhoneNumbers[0].Value)
	}

	// Display status
	if user.Active != nil {
		_ = table.Append("Active", fmt.Sprintf("%t", *user.Active))
	}
	if user.Verified != nil {
		_ = table.Append("Verified", fmt.Sprintf("%t", *user.Verified))
	}

	// Display metadata
	if user.Meta != nil {
		if user.Meta.Created != "" {
			_ = table.Append("Created", user.Meta.Created)
		}
		if user.Meta.LastModified != "" {
			_ = table.Append("Last Modified", user.Meta.LastModified)
		}
		if user.Meta.Version > 0 {
			_ = table.Append("Version", fmt.Sprintf("%d", user.Meta.Version))
		}
	}

	// Display groups count
	if len(user.Groups) > 0 {
		_ = table.Append("Groups", fmt.Sprintf("%d groups", len(user.Groups)))
	}

	_ = table.Render()
	return nil
}

func displayUsersTable(users []uaa.User) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header("Username", "Email", "Given Name", "Active", "Verified", "Origin", "Created")

	for _, user := range users {
		username := user.Username
		email := ""
		if len(user.Emails) > 0 {
			email = user.Emails[0].Value
		}
		givenName := ""
		if user.Name != nil {
			givenName = user.Name.GivenName
		}
		active := "unknown"
		if user.Active != nil {
			active = fmt.Sprintf("%t", *user.Active)
		}
		verified := "unknown"
		if user.Verified != nil {
			verified = fmt.Sprintf("%t", *user.Verified)
		}
		origin := user.Origin
		created := ""
		if user.Meta != nil && user.Meta.Created != "" {
			created = user.Meta.Created
		}

		_ = table.Append(username, email, givenName, active, verified, origin, created)
	}

	_ = table.Render()
	return nil
}
