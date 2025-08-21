package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fivetwenty-io/capi-client/pkg/capi"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// validateFilePath validates that a file path is safe to read
func validateFilePathSecurity(filePath string) error {
	// Clean the path to resolve any path traversal attempts
	cleanPath := filepath.Clean(filePath)

	// Check for path traversal attempts
	if filepath.IsAbs(filePath) {
		// Allow absolute paths but ensure they're clean
		if cleanPath != filePath {
			return fmt.Errorf("invalid file path: potential path traversal attempt")
		}
	} else {
		// For relative paths, ensure they don't escape the current directory
		if len(cleanPath) > 0 && cleanPath[0] == '.' && len(cleanPath) > 1 && cleanPath[1] == '.' {
			return fmt.Errorf("invalid file path: path traversal not allowed")
		}
	}

	// Check if file exists and is readable
	if _, err := os.Stat(cleanPath); err != nil {
		return fmt.Errorf("file not accessible: %w", err)
	}

	return nil
}

// NewSecurityGroupsCommand creates the security groups command group
func NewSecurityGroupsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "security-groups",
		Aliases: []string{"security-group", "sg"},
		Short:   "Manage security groups",
		Long:    "List and manage Cloud Foundry security groups",
	}

	cmd.AddCommand(newSecurityGroupsListCommand())
	cmd.AddCommand(newSecurityGroupsGetCommand())
	cmd.AddCommand(newSecurityGroupsCreateCommand())
	cmd.AddCommand(newSecurityGroupsUpdateCommand())
	cmd.AddCommand(newSecurityGroupsDeleteCommand())
	cmd.AddCommand(newSecurityGroupsBindCommand())
	cmd.AddCommand(newSecurityGroupsUnbindCommand())
	cmd.AddCommand(newSecurityGroupsRunningCommand())
	cmd.AddCommand(newSecurityGroupsStagingCommand())

	return cmd
}

func newSecurityGroupsListCommand() *cobra.Command {
	var (
		allPages        bool
		perPage         int
		globallyEnabled bool
		runningSpaces   bool
		stagingSpaces   bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List security groups",
		Long:  "List all security groups",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			params := capi.NewQueryParams()
			params.PerPage = perPage

			// Apply filters based on flags
			if globallyEnabled {
				params.WithFilter("globally_enabled_running", "true")
			}

			securityGroups, err := client.SecurityGroups().List(ctx, params)
			if err != nil {
				return fmt.Errorf("failed to list security groups: %w", err)
			}

			// Fetch all pages if requested
			allGroups := securityGroups.Resources
			if allPages && securityGroups.Pagination.TotalPages > 1 {
				for page := 2; page <= securityGroups.Pagination.TotalPages; page++ {
					params.Page = page
					moreGroups, err := client.SecurityGroups().List(ctx, params)
					if err != nil {
						return fmt.Errorf("failed to fetch page %d: %w", page, err)
					}
					allGroups = append(allGroups, moreGroups.Resources...)
				}
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(allGroups)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(allGroups)
			default:
				if len(allGroups) == 0 {
					fmt.Println("No security groups found")
					return nil
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Name", "GUID", "Running Spaces", "Staging Spaces", "Global Running", "Global Staging")

				for _, sg := range allGroups {
					runningSpacesCount := len(sg.Relationships.RunningSpaces.Data)
					stagingSpacesCount := len(sg.Relationships.StagingSpaces.Data)

					globalRunning := "no"
					if sg.GloballyEnabled.Running {
						globalRunning = "yes"
					}

					globalStaging := "no"
					if sg.GloballyEnabled.Staging {
						globalStaging = "yes"
					}

					_ = table.Append(sg.Name, sg.GUID,
						fmt.Sprintf("%d", runningSpacesCount),
						fmt.Sprintf("%d", stagingSpacesCount),
						globalRunning, globalStaging)
				}

				_ = table.Render()

				if !allPages && securityGroups.Pagination.TotalPages > 1 {
					fmt.Printf("\nShowing page 1 of %d. Use --all to fetch all pages.\n", securityGroups.Pagination.TotalPages)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&allPages, "all", false, "fetch all pages")
	cmd.Flags().IntVar(&perPage, "per-page", 50, "results per page")
	cmd.Flags().BoolVar(&globallyEnabled, "globally-enabled", false, "filter by globally enabled for running")
	cmd.Flags().BoolVar(&runningSpaces, "running-spaces", false, "show running spaces bindings")
	cmd.Flags().BoolVar(&stagingSpaces, "staging-spaces", false, "show staging spaces bindings")

	return cmd
}

func newSecurityGroupsGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get SECURITY_GROUP_NAME_OR_GUID",
		Short: "Get security group details",
		Long:  "Display detailed information about a specific security group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Try to get by GUID first
			sg, err := client.SecurityGroups().Get(ctx, nameOrGUID)
			if err != nil {
				// If not found by GUID, try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)
				groups, err := client.SecurityGroups().List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find security group: %w", err)
				}
				if len(groups.Resources) == 0 {
					return fmt.Errorf("security group '%s' not found", nameOrGUID)
				}
				sg = &groups.Resources[0]
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(sg)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(sg)
			default:
				fmt.Printf("Security Group: %s\n", sg.Name)
				fmt.Printf("  GUID:     %s\n", sg.GUID)
				fmt.Printf("  Created:  %s\n", sg.CreatedAt.Format("2006-01-02 15:04:05"))
				fmt.Printf("  Updated:  %s\n", sg.UpdatedAt.Format("2006-01-02 15:04:05"))

				fmt.Printf("  Globally Enabled:\n")
				fmt.Printf("    Running: %t\n", sg.GloballyEnabled.Running)
				fmt.Printf("    Staging: %t\n", sg.GloballyEnabled.Staging)

				if len(sg.Rules) > 0 {
					fmt.Printf("  Rules:\n")
					for i, rule := range sg.Rules {
						fmt.Printf("    Rule %d:\n", i+1)
						fmt.Printf("      Protocol:    %s\n", rule.Protocol)
						fmt.Printf("      Destination: %s\n", rule.Destination)
						if rule.Ports != nil {
							fmt.Printf("      Ports:       %s\n", *rule.Ports)
						}
						if rule.Description != nil {
							fmt.Printf("      Description: %s\n", *rule.Description)
						}
						if rule.Log != nil {
							fmt.Printf("      Log:         %t\n", *rule.Log)
						}
					}
				}

				// Show bound spaces
				runningSpacesCount := len(sg.Relationships.RunningSpaces.Data)
				stagingSpacesCount := len(sg.Relationships.StagingSpaces.Data)
				fmt.Printf("  Bound to %d running spaces, %d staging spaces\n", runningSpacesCount, stagingSpacesCount)
			}

			return nil
		},
	}
}

func newSecurityGroupsCreateCommand() *cobra.Command {
	var (
		name          string
		rulesFile     string
		globalRunning bool
		globalStaging bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a security group",
		Long:  "Create a new security group with rules",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("security group name is required")
			}

			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			createReq := &capi.SecurityGroupCreateRequest{
				Name: name,
			}

			// Set global enablement
			if globalRunning || globalStaging {
				createReq.GloballyEnabled = &capi.SecurityGroupGloballyEnabled{
					Running: globalRunning,
					Staging: globalStaging,
				}
			}

			// Load rules from file if specified
			if rulesFile != "" {
				// Validate file path to prevent directory traversal
				if err := validateFilePathSecurity(rulesFile); err != nil {
					return fmt.Errorf("invalid rules file: %w", err)
				}
				rulesContent, err := os.ReadFile(filepath.Clean(rulesFile))
				if err != nil {
					return fmt.Errorf("failed to read rules file: %w", err)
				}

				var rules []capi.SecurityGroupRule
				if err := json.Unmarshal(rulesContent, &rules); err != nil {
					return fmt.Errorf("failed to parse rules JSON: %w", err)
				}

				createReq.Rules = rules
			}

			sg, err := client.SecurityGroups().Create(ctx, createReq)
			if err != nil {
				return fmt.Errorf("failed to create security group: %w", err)
			}

			fmt.Printf("Successfully created security group '%s'\n", sg.Name)
			fmt.Printf("  GUID: %s\n", sg.GUID)
			fmt.Printf("  Rules: %d\n", len(sg.Rules))

			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "security group name (required)")
	cmd.Flags().StringVarP(&rulesFile, "rules", "r", "", "JSON file containing security group rules")
	cmd.Flags().BoolVar(&globalRunning, "global-running", false, "enable globally for running applications")
	cmd.Flags().BoolVar(&globalStaging, "global-staging", false, "enable globally for staging applications")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newSecurityGroupsUpdateCommand() *cobra.Command {
	var (
		newName       string
		rulesFile     string
		globalRunning bool
		globalStaging bool
	)

	cmd := &cobra.Command{
		Use:   "update SECURITY_GROUP_NAME_OR_GUID",
		Short: "Update a security group",
		Long:  "Update an existing security group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find security group
			var sgGUID string
			sg, err := client.SecurityGroups().Get(ctx, nameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)
				groups, err := client.SecurityGroups().List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find security group: %w", err)
				}
				if len(groups.Resources) == 0 {
					return fmt.Errorf("security group '%s' not found", nameOrGUID)
				}
				sg = &groups.Resources[0]
			}
			sgGUID = sg.GUID

			// Build update request
			updateReq := &capi.SecurityGroupUpdateRequest{}

			if newName != "" {
				updateReq.Name = &newName
			}

			// Set global enablement if specified
			if cmd.Flags().Changed("global-running") || cmd.Flags().Changed("global-staging") {
				updateReq.GloballyEnabled = &capi.SecurityGroupGloballyEnabled{
					Running: sg.GloballyEnabled.Running, // Keep existing value
					Staging: sg.GloballyEnabled.Staging, // Keep existing value
				}
				if cmd.Flags().Changed("global-running") {
					updateReq.GloballyEnabled.Running = globalRunning
				}
				if cmd.Flags().Changed("global-staging") {
					updateReq.GloballyEnabled.Staging = globalStaging
				}
			}

			// Load rules from file if specified
			if rulesFile != "" {
				// Validate file path to prevent directory traversal
				if err := validateFilePathSecurity(rulesFile); err != nil {
					return fmt.Errorf("invalid rules file: %w", err)
				}
				rulesContent, err := os.ReadFile(filepath.Clean(rulesFile))
				if err != nil {
					return fmt.Errorf("failed to read rules file: %w", err)
				}

				var rules []capi.SecurityGroupRule
				if err := json.Unmarshal(rulesContent, &rules); err != nil {
					return fmt.Errorf("failed to parse rules JSON: %w", err)
				}

				updateReq.Rules = rules
			}

			updatedSG, err := client.SecurityGroups().Update(ctx, sgGUID, updateReq)
			if err != nil {
				return fmt.Errorf("failed to update security group: %w", err)
			}

			fmt.Printf("Successfully updated security group '%s'\n", updatedSG.Name)

			return nil
		},
	}

	cmd.Flags().StringVar(&newName, "name", "", "new security group name")
	cmd.Flags().StringVarP(&rulesFile, "rules", "r", "", "JSON file containing security group rules")
	cmd.Flags().BoolVar(&globalRunning, "global-running", false, "enable globally for running applications")
	cmd.Flags().BoolVar(&globalStaging, "global-staging", false, "enable globally for staging applications")

	return cmd
}

func newSecurityGroupsDeleteCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete SECURITY_GROUP_NAME_OR_GUID",
		Short: "Delete a security group",
		Long:  "Delete a security group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			if !force {
				fmt.Printf("Really delete security group '%s'? (y/N): ", nameOrGUID)
				var response string
				_, _ = fmt.Scanln(&response)
				if response != "y" && response != "Y" {
					fmt.Println("Cancelled")
					return nil
				}
			}

			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find security group
			var sgGUID string
			var sgName string
			sg, err := client.SecurityGroups().Get(ctx, nameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)
				groups, err := client.SecurityGroups().List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find security group: %w", err)
				}
				if len(groups.Resources) == 0 {
					return fmt.Errorf("security group '%s' not found", nameOrGUID)
				}
				sg = &groups.Resources[0]
			}
			sgGUID = sg.GUID
			sgName = sg.Name

			job, err := client.SecurityGroups().Delete(ctx, sgGUID)
			if err != nil {
				return fmt.Errorf("failed to delete security group: %w", err)
			}

			if job != nil {
				fmt.Printf("Deleting security group '%s'... (job: %s)\n", sgName, job.GUID)
				fmt.Printf("Monitor with: capi jobs get %s\n", job.GUID)
			} else {
				fmt.Printf("Successfully deleted security group '%s'\n", sgName)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "force deletion without confirmation")

	return cmd
}

func newSecurityGroupsBindCommand() *cobra.Command {
	var (
		spaceNames []string
		running    bool
		staging    bool
	)

	cmd := &cobra.Command{
		Use:   "bind SECURITY_GROUP_NAME_OR_GUID",
		Short: "Bind security group to spaces",
		Long:  "Bind a security group to spaces for running or staging",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			if len(spaceNames) == 0 {
				return fmt.Errorf("at least one space must be specified")
			}

			if !running && !staging {
				return fmt.Errorf("must specify --running or --staging (or both)")
			}

			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find security group
			var sgGUID string
			sg, err := client.SecurityGroups().Get(ctx, nameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)
				groups, err := client.SecurityGroups().List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find security group: %w", err)
				}
				if len(groups.Resources) == 0 {
					return fmt.Errorf("security group '%s' not found", nameOrGUID)
				}
				sg = &groups.Resources[0]
			}
			sgGUID = sg.GUID

			// Find spaces to bind to
			var spaceGUIDs []string
			for _, spaceName := range spaceNames {
				params := capi.NewQueryParams()
				params.WithFilter("names", spaceName)

				// Add org filter if targeted
				if orgGUID := viper.GetString("organization_guid"); orgGUID != "" {
					params.WithFilter("organization_guids", orgGUID)
				}

				spaces, err := client.Spaces().List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find space '%s': %w", spaceName, err)
				}
				if len(spaces.Resources) == 0 {
					return fmt.Errorf("space '%s' not found", spaceName)
				}
				spaceGUIDs = append(spaceGUIDs, spaces.Resources[0].GUID)
			}

			// Bind to running spaces
			if running {
				_, err = client.SecurityGroups().BindRunningSpaces(ctx, sgGUID, spaceGUIDs)
				if err != nil {
					return fmt.Errorf("failed to bind security group to running spaces: %w", err)
				}
				fmt.Printf("Successfully bound security group '%s' to running spaces: %s\n", sg.Name, strings.Join(spaceNames, ", "))
			}

			// Bind to staging spaces
			if staging {
				_, err = client.SecurityGroups().BindStagingSpaces(ctx, sgGUID, spaceGUIDs)
				if err != nil {
					return fmt.Errorf("failed to bind security group to staging spaces: %w", err)
				}
				fmt.Printf("Successfully bound security group '%s' to staging spaces: %s\n", sg.Name, strings.Join(spaceNames, ", "))
			}

			return nil
		},
	}

	cmd.Flags().StringArrayVarP(&spaceNames, "spaces", "s", nil, "spaces to bind to (required)")
	cmd.Flags().BoolVar(&running, "running", false, "bind for running applications")
	cmd.Flags().BoolVar(&staging, "staging", false, "bind for staging applications")
	_ = cmd.MarkFlagRequired("spaces")

	return cmd
}

func newSecurityGroupsUnbindCommand() *cobra.Command {
	var (
		spaceName string
		running   bool
		staging   bool
	)

	cmd := &cobra.Command{
		Use:   "unbind SECURITY_GROUP_NAME_OR_GUID",
		Short: "Unbind security group from a space",
		Long:  "Remove a security group binding from a space",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrGUID := args[0]

			if spaceName == "" {
				return fmt.Errorf("space name is required")
			}

			if !running && !staging {
				return fmt.Errorf("must specify --running or --staging (or both)")
			}

			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Find security group
			var sgGUID string
			sg, err := client.SecurityGroups().Get(ctx, nameOrGUID)
			if err != nil {
				// Try by name
				params := capi.NewQueryParams()
				params.WithFilter("names", nameOrGUID)
				groups, err := client.SecurityGroups().List(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to find security group: %w", err)
				}
				if len(groups.Resources) == 0 {
					return fmt.Errorf("security group '%s' not found", nameOrGUID)
				}
				sg = &groups.Resources[0]
			}
			sgGUID = sg.GUID

			// Find space to unbind from
			params := capi.NewQueryParams()
			params.WithFilter("names", spaceName)

			// Add org filter if targeted
			if orgGUID := viper.GetString("organization_guid"); orgGUID != "" {
				params.WithFilter("organization_guids", orgGUID)
			}

			spaces, err := client.Spaces().List(ctx, params)
			if err != nil {
				return fmt.Errorf("failed to find space: %w", err)
			}
			if len(spaces.Resources) == 0 {
				return fmt.Errorf("space '%s' not found", spaceName)
			}
			spaceGUID := spaces.Resources[0].GUID

			// Unbind from running spaces
			if running {
				err = client.SecurityGroups().UnbindRunningSpace(ctx, sgGUID, spaceGUID)
				if err != nil {
					return fmt.Errorf("failed to unbind security group from running space: %w", err)
				}
				fmt.Printf("Successfully unbound security group '%s' from running space '%s'\n", sg.Name, spaceName)
			}

			// Unbind from staging spaces
			if staging {
				err = client.SecurityGroups().UnbindStagingSpace(ctx, sgGUID, spaceGUID)
				if err != nil {
					return fmt.Errorf("failed to unbind security group from staging space: %w", err)
				}
				fmt.Printf("Successfully unbound security group '%s' from staging space '%s'\n", sg.Name, spaceName)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&spaceName, "space", "s", "", "space to unbind from (required)")
	cmd.Flags().BoolVar(&running, "running", false, "unbind from running applications")
	cmd.Flags().BoolVar(&staging, "staging", false, "unbind from staging applications")
	_ = cmd.MarkFlagRequired("space")

	return cmd
}

func newSecurityGroupsRunningCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "running",
		Short: "List running security groups",
		Long:  "List all security groups that are globally enabled for running applications",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			params := capi.NewQueryParams()
			params.WithFilter("globally_enabled_running", "true")

			securityGroups, err := client.SecurityGroups().List(ctx, params)
			if err != nil {
				return fmt.Errorf("failed to list running security groups: %w", err)
			}

			if len(securityGroups.Resources) == 0 {
				fmt.Println("No globally enabled running security groups found")
				return nil
			}

			fmt.Println("Globally enabled running security groups:")
			for _, sg := range securityGroups.Resources {
				fmt.Printf("  - %s (%s)\n", sg.Name, sg.GUID)
			}

			return nil
		},
	}
}

func newSecurityGroupsStagingCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "staging",
		Short: "List staging security groups",
		Long:  "List all security groups that are globally enabled for staging applications",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := createClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()
			params := capi.NewQueryParams()
			params.WithFilter("globally_enabled_staging", "true")

			securityGroups, err := client.SecurityGroups().List(ctx, params)
			if err != nil {
				return fmt.Errorf("failed to list staging security groups: %w", err)
			}

			if len(securityGroups.Resources) == 0 {
				fmt.Println("No globally enabled staging security groups found")
				return nil
			}

			fmt.Println("Globally enabled staging security groups:")
			for _, sg := range securityGroups.Resources {
				fmt.Printf("  - %s (%s)\n", sg.Name, sg.GUID)
			}

			return nil
		},
	}
}
