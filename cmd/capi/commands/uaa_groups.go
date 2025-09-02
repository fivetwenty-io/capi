package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/cloudfoundry-community/go-uaa"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// createUsersCreateGroupCommand creates the create group command
func createUsersCreateGroupCommand() *cobra.Command {
	var description string
	var members []string

	cmd := &cobra.Command{
		Use:     "create-group <name>",
		Aliases: []string{"add-group", "new-group"},
		Short:   "Create a group",
		Long: `Create a new group in UAA.

Groups in UAA are used to organize users and manage permissions. You can
optionally specify initial members when creating the group.`,
		Example: `  # Create basic group
  capi uaa create-group developers

  # Create group with description
  capi uaa create-group qa-team --description "Quality Assurance Team"

  # Create group with initial members
  capi uaa create-group admins --description "Administrators" --members user1,user2`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := loadConfig()
			groupName := args[0]

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

			// Build group object
			group := uaa.Group{
				DisplayName: groupName,
				Description: description,
			}

			// Add initial members if specified
			if len(members) > 0 {
				groupMembers := make([]uaa.GroupMember, 0, len(members))
				for _, member := range members {
					// Assume members are user IDs for now
					groupMembers = append(groupMembers, uaa.GroupMember{
						Value:  member,
						Type:   "USER",
						Origin: "uaa", // Default origin
					})
				}
				group.Members = groupMembers
			}

			// Create group
			createdGroup, err := uaaClient.Client().CreateGroup(group)
			if err != nil {
				return fmt.Errorf("failed to create group: %w", err)
			}

			// Display created group
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(createdGroup)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(createdGroup)
			default:
				return displayGroupTable(createdGroup)
			}
		},
	}

	cmd.Flags().StringVar(&description, "description", "", "Group description")
	cmd.Flags().StringSliceVar(&members, "members", nil, "Initial group members (comma-separated user IDs)")

	return cmd
}

// createUsersGetGroupCommand creates the get group command
func createUsersGetGroupCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get-group <name>",
		Short: "Get group details",
		Long: `Look up a group by name and display detailed information.

The command will search for the group by name and display all available
group attributes including members and metadata.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := loadConfig()
			groupName := args[0]

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

			// Get group by name
			group, err := uaaClient.Client().GetGroupByName(groupName, "")
			if err != nil {
				return fmt.Errorf("failed to get group: %w", err)
			}

			// Display group
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(group)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(group)
			default:
				return displayGroupTable(group)
			}
		},
	}
}

// createUsersListGroupsCommand creates the list groups command
func createUsersListGroupsCommand() *cobra.Command {
	var filter, sortBy, attributes string
	var sortOrder string
	var count, startIndex int
	var all bool

	cmd := &cobra.Command{
		Use:   "list-groups",
		Short: "List groups",
		Long: `Search and list groups with optional SCIM filters.

SCIM filters allow complex queries on group attributes. Examples:
- displayName eq "admin"
- meta.created gt "2023-01-01T00:00:00.000Z"`,
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

			var groups []uaa.Group
			if all {
				// Get all groups across all pages
				groups, err = uaaClient.Client().ListAllGroups(filter, sortBy, attributes, uaaSortOrder)
			} else {
				// Get groups with pagination
				groups, _, err = uaaClient.Client().ListGroups(filter, sortBy, attributes, uaaSortOrder, startIndex, count)
			}

			if err != nil {
				return fmt.Errorf("failed to list groups: %w", err)
			}

			// Display groups
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(groups)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(groups)
			default:
				return displayGroupsTable(groups)
			}
		},
	}

	cmd.Flags().StringVar(&filter, "filter", "", "SCIM filter expression")
	cmd.Flags().StringVar(&sortBy, "sort-by", "", "Attribute to sort by")
	cmd.Flags().StringVar(&sortOrder, "sort-order", "ascending", "Sort order (ascending, descending)")
	cmd.Flags().StringVar(&attributes, "attributes", "", "Comma-separated list of attributes to return")
	cmd.Flags().IntVar(&count, "count", 50, "Number of results per page")
	cmd.Flags().IntVar(&startIndex, "start-index", 1, "Starting index for pagination")
	cmd.Flags().BoolVar(&all, "all", false, "Fetch all groups across all pages")

	return cmd
}

// createUsersAddMemberCommand creates the add member command
func createUsersAddMemberCommand() *cobra.Command {
	var origin, memberType string

	cmd := &cobra.Command{
		Use:   "add-member <group> <member>",
		Short: "Add user to group",
		Long: `Add a user to a group.

The group can be specified by name or ID, and the member can be a username or user ID.
By default, members are treated as users from the 'uaa' origin.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := loadConfig()
			groupIdentifier := args[0]
			memberIdentifier := args[1]

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

			// Resolve group to get ID
			var groupID string
			if isUUID(groupIdentifier) {
				groupID = groupIdentifier
			} else {
				group, err := uaaClient.Client().GetGroupByName(groupIdentifier, "")
				if err != nil {
					return fmt.Errorf("failed to find group '%s': %w", groupIdentifier, err)
				}
				groupID = group.ID
			}

			// Resolve member to get ID
			var memberID string
			if isUUID(memberIdentifier) {
				memberID = memberIdentifier
			} else {
				user, err := uaaClient.Client().GetUserByUsername(memberIdentifier, "", "")
				if err != nil {
					return fmt.Errorf("failed to find user '%s': %w", memberIdentifier, err)
				}
				memberID = user.ID
			}

			// Add member to group
			err = uaaClient.Client().AddGroupMember(groupID, memberID, memberType, origin)
			if err != nil {
				return fmt.Errorf("failed to add member to group: %w", err)
			}

			fmt.Printf("Successfully added member '%s' to group '%s'\n", memberIdentifier, groupIdentifier)
			return nil
		},
	}

	cmd.Flags().StringVar(&origin, "origin", "uaa", "Member origin (identity provider)")
	cmd.Flags().StringVar(&memberType, "type", "USER", "Member type (USER or GROUP)")

	return cmd
}

// createUsersRemoveMemberCommand creates the remove member command
func createUsersRemoveMemberCommand() *cobra.Command {
	var origin, memberType string

	cmd := &cobra.Command{
		Use:   "remove-member <group> <member>",
		Short: "Remove user from group",
		Long: `Remove a user from a group.

The group can be specified by name or ID, and the member can be a username or user ID.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := loadConfig()
			groupIdentifier := args[0]
			memberIdentifier := args[1]

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

			// Resolve group to get ID
			var groupID string
			if isUUID(groupIdentifier) {
				groupID = groupIdentifier
			} else {
				group, err := uaaClient.Client().GetGroupByName(groupIdentifier, "")
				if err != nil {
					return fmt.Errorf("failed to find group '%s': %w", groupIdentifier, err)
				}
				groupID = group.ID
			}

			// Resolve member to get ID
			var memberID string
			if isUUID(memberIdentifier) {
				memberID = memberIdentifier
			} else {
				user, err := uaaClient.Client().GetUserByUsername(memberIdentifier, "", "")
				if err != nil {
					return fmt.Errorf("failed to find user '%s': %w", memberIdentifier, err)
				}
				memberID = user.ID
			}

			// Remove member from group
			err = uaaClient.Client().RemoveGroupMember(groupID, memberID, memberType, origin)
			if err != nil {
				return fmt.Errorf("failed to remove member from group: %w", err)
			}

			fmt.Printf("Successfully removed member '%s' from group '%s'\n", memberIdentifier, groupIdentifier)
			return nil
		},
	}

	cmd.Flags().StringVar(&origin, "origin", "uaa", "Member origin (identity provider)")
	cmd.Flags().StringVar(&memberType, "type", "USER", "Member type (USER or GROUP)")

	return cmd
}

// createUsersMapGroupCommand creates the map group command
func createUsersMapGroupCommand() *cobra.Command {
	var group, externalGroup, origin string

	cmd := &cobra.Command{
		Use:   "map-group",
		Short: "Map external group to UAA group",
		Long: `Map an external group from an identity provider to a UAA group/scope.

This allows users from external identity providers to automatically
inherit UAA group memberships based on their external group memberships.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			config := loadConfig()

			if GetEffectiveUAAEndpoint(config) == "" {
				return fmt.Errorf("no UAA endpoint configured. Use 'capi uaa target <url>' to set one")
			}

			// Validate required flags
			if group == "" {
				return fmt.Errorf("--group flag is required")
			}
			if externalGroup == "" {
				return fmt.Errorf("--external-group flag is required")
			}
			if origin == "" {
				return fmt.Errorf("--origin flag is required")
			}

			// Create UAA client
			uaaClient, err := NewUAAClient(config)
			if err != nil {
				return fmt.Errorf("failed to create UAA client: %w", err)
			}

			if !uaaClient.IsAuthenticated() {
				return fmt.Errorf("not authenticated. Use a token command to authenticate first")
			}

			// Resolve group to get ID
			var groupID string
			if isUUID(group) {
				groupID = group
			} else {
				groupObj, err := uaaClient.Client().GetGroupByName(group, "")
				if err != nil {
					return fmt.Errorf("failed to find group '%s': %w", group, err)
				}
				groupID = groupObj.ID
			}

			// Map group
			err = uaaClient.Client().MapGroup(groupID, externalGroup, origin)
			if err != nil {
				return fmt.Errorf("failed to map group: %w", err)
			}

			fmt.Printf("Successfully mapped external group '%s' from origin '%s' to UAA group '%s'\n",
				externalGroup, origin, group)
			return nil
		},
	}

	cmd.Flags().StringVar(&group, "group", "", "UAA group name or ID (required)")
	cmd.Flags().StringVar(&externalGroup, "external-group", "", "External group name (required)")
	cmd.Flags().StringVar(&origin, "origin", "", "Identity provider origin (required)")

	return cmd
}

// createUsersUnmapGroupCommand creates the unmap group command
func createUsersUnmapGroupCommand() *cobra.Command {
	var group, externalGroup, origin string

	cmd := &cobra.Command{
		Use:   "unmap-group",
		Short: "Unmap external group from UAA group",
		Long: `Remove a mapping between an external group and a UAA group/scope.

This removes the automatic group membership inheritance for users
from the specified external identity provider.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			config := loadConfig()

			if GetEffectiveUAAEndpoint(config) == "" {
				return fmt.Errorf("no UAA endpoint configured. Use 'capi uaa target <url>' to set one")
			}

			// Validate required flags
			if group == "" {
				return fmt.Errorf("--group flag is required")
			}
			if externalGroup == "" {
				return fmt.Errorf("--external-group flag is required")
			}
			if origin == "" {
				return fmt.Errorf("--origin flag is required")
			}

			// Create UAA client
			uaaClient, err := NewUAAClient(config)
			if err != nil {
				return fmt.Errorf("failed to create UAA client: %w", err)
			}

			if !uaaClient.IsAuthenticated() {
				return fmt.Errorf("not authenticated. Use a token command to authenticate first")
			}

			// Resolve group to get ID
			var groupID string
			if isUUID(group) {
				groupID = group
			} else {
				groupObj, err := uaaClient.Client().GetGroupByName(group, "")
				if err != nil {
					return fmt.Errorf("failed to find group '%s': %w", group, err)
				}
				groupID = groupObj.ID
			}

			// Unmap group
			err = uaaClient.Client().UnmapGroup(groupID, externalGroup, origin)
			if err != nil {
				return fmt.Errorf("failed to unmap group: %w", err)
			}

			fmt.Printf("Successfully unmapped external group '%s' from origin '%s' from UAA group '%s'\n",
				externalGroup, origin, group)
			return nil
		},
	}

	cmd.Flags().StringVar(&group, "group", "", "UAA group name or ID (required)")
	cmd.Flags().StringVar(&externalGroup, "external-group", "", "External group name (required)")
	cmd.Flags().StringVar(&origin, "origin", "", "Identity provider origin (required)")

	return cmd
}

// createUsersListGroupMappingsCommand creates the list group mappings command
func createUsersListGroupMappingsCommand() *cobra.Command {
	var origin string
	var count, startIndex int

	cmd := &cobra.Command{
		Use:   "list-group-mappings",
		Short: "List group mappings",
		Long: `List all mappings between UAA groups and external groups.

This shows how external groups from identity providers are mapped
to UAA groups/scopes for automatic membership inheritance.`,
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

			var mappings []uaa.GroupMapping
			if origin != "" {
				// Get mappings for specific origin with pagination
				mappings, _, err = uaaClient.Client().ListGroupMappings(origin, startIndex, count)
			} else {
				// Get all mappings across all origins
				mappings, err = uaaClient.Client().ListAllGroupMappings("")
			}

			if err != nil {
				return fmt.Errorf("failed to list group mappings: %w", err)
			}

			// Display mappings
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(mappings)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(mappings)
			default:
				return displayGroupMappingsTable(mappings)
			}
		},
	}

	cmd.Flags().StringVar(&origin, "origin", "", "Filter by identity provider origin")
	cmd.Flags().IntVar(&count, "count", 50, "Number of results per page (ignored when no origin specified)")
	cmd.Flags().IntVar(&startIndex, "start-index", 1, "Starting index for pagination (ignored when no origin specified)")

	return cmd
}

// Helper functions for group display

func displayGroupTable(group *uaa.Group) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header("Property", "Value")

	if group.ID != "" {
		_ = table.Append("ID", group.ID)
	}
	if group.DisplayName != "" {
		_ = table.Append("Display Name", group.DisplayName)
	}
	if group.Description != "" {
		_ = table.Append("Description", group.Description)
	}
	if group.ZoneID != "" {
		_ = table.Append("Zone ID", group.ZoneID)
	}

	// Display metadata
	if group.Meta != nil {
		if group.Meta.Created != "" {
			_ = table.Append("Created", group.Meta.Created)
		}
		if group.Meta.LastModified != "" {
			_ = table.Append("Last Modified", group.Meta.LastModified)
		}
		if group.Meta.Version > 0 {
			_ = table.Append("Version", strconv.Itoa(group.Meta.Version))
		}
	}

	// Display members count
	if len(group.Members) > 0 {
		_ = table.Append("Members", fmt.Sprintf("%d members", len(group.Members)))
	}

	_ = table.Render()
	return nil
}

func displayGroupsTable(groups []uaa.Group) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header("Display Name", "Description", "Members", "Created")

	for _, group := range groups {
		displayName := group.DisplayName
		description := group.Description
		if len(description) > 50 {
			description = description[:50] + "..."
		}
		memberCount := fmt.Sprintf("%d", len(group.Members))
		created := ""
		if group.Meta != nil && group.Meta.Created != "" {
			created = group.Meta.Created
		}

		_ = table.Append(displayName, description, memberCount, created)
	}

	_ = table.Render()
	return nil
}

func displayGroupMappingsTable(mappings []uaa.GroupMapping) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header("UAA Group", "External Group", "Origin", "Created")

	for _, mapping := range mappings {
		uaaGroup := mapping.DisplayName
		externalGroup := mapping.ExternalGroup
		origin := mapping.Origin
		created := ""
		if mapping.Meta != nil && mapping.Meta.Created != "" {
			created = mapping.Meta.Created
		}

		_ = table.Append(uaaGroup, externalGroup, origin, created)
	}

	_ = table.Render()
	return nil
}

// isUUID checks if a string looks like a UUID
func isUUID(s string) bool {
	// Simple UUID format check: 8-4-4-4-12 hex digits
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
		} else {
			if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
				return false
			}
		}
	}
	return true
}
