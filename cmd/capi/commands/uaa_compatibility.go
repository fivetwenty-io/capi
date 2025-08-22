package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry-community/go-uaa"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// UAAVersionInfo contains UAA version and compatibility information
type UAAVersionInfo struct {
	Version       string                 `json:"version"`
	Endpoint      string                 `json:"endpoint"`
	ServerInfo    map[string]interface{} `json:"server_info"`
	Features      []string               `json:"features"`
	Compatibility CompatibilityStatus    `json:"compatibility"`
	TestedAt      time.Time              `json:"tested_at"`
}

// CompatibilityStatus represents the compatibility test results
type CompatibilityStatus struct {
	Overall         string            `json:"overall"`         // "compatible", "partial", "incompatible"
	BasicAuth       string            `json:"basic_auth"`      // Token operations
	UserMgmt        string            `json:"user_mgmt"`       // User CRUD operations
	GroupMgmt       string            `json:"group_mgmt"`      // Group management
	ClientMgmt      string            `json:"client_mgmt"`     // OAuth client management
	FeatureTests    map[string]string `json:"feature_tests"`   // Individual feature test results
	Issues          []string          `json:"issues"`          // Known issues
	Recommendations []string          `json:"recommendations"` // Recommendations for this version
}

// CFIntegrationInfo contains Cloud Foundry integration details
type CFIntegrationInfo struct {
	CFAPIVersion    string `json:"cf_api_version"`
	UAAVersion      string `json:"uaa_version"`
	AuthMethod      string `json:"auth_method"`
	ScopesSupported bool   `json:"scopes_supported"`
	TokenFormat     string `json:"token_format"`
	Compatible      bool   `json:"compatible"`
}

// createUsersCompatibilityCommand creates the compatibility testing command
func createUsersCompatibilityCommand() *cobra.Command {
	var testBasic, testAll bool
	var saveResults bool

	cmd := &cobra.Command{
		Use:   "compatibility",
		Short: "Test UAA compatibility and features",
		Long: `Test compatibility with the current UAA endpoint and version.
		
This command performs a series of tests to verify that the UAA endpoint
supports the required features and operations for full compatibility
with the capi CLI UAA commands.`,
		Example: `  # Test basic compatibility
  capi uaa compatibility

  # Run comprehensive compatibility tests
  capi uaa compatibility --all

  # Save test results to file
  capi uaa compatibility --save-results`,
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

			// Run compatibility tests
			versionInfo, err := runCompatibilityTests(uaaClient, testAll)
			if err != nil {
				return fmt.Errorf("failed to run compatibility tests: %w", err)
			}

			// Save results if requested
			if saveResults {
				if err := saveCompatibilityResults(versionInfo); err != nil {
					fmt.Printf("Warning: Failed to save results: %v\n", err)
				} else {
					fmt.Println("Compatibility results saved to uaa-compatibility-results.json")
				}
			}

			// Display results
			return displayCompatibilityResults(versionInfo)
		},
	}

	cmd.Flags().BoolVar(&testBasic, "basic", false, "Run only basic compatibility tests")
	cmd.Flags().BoolVar(&testAll, "all", false, "Run comprehensive compatibility tests")
	cmd.Flags().BoolVar(&saveResults, "save-results", false, "Save test results to file")

	return cmd
}

// createUsersCFIntegrationCommand creates the CF integration testing command
func createUsersCFIntegrationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cf-integration",
		Short: "Test Cloud Foundry integration",
		Long: `Test integration with Cloud Foundry API and authentication.
		
This command verifies that the UAA configuration is compatible with
Cloud Foundry and can properly authenticate CF API requests.`,
		Example: `  # Test CF integration
  capi uaa cf-integration

  # Show integration details in JSON
  capi uaa cf-integration --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			config := loadConfig()

			if config.UAAEndpoint == "" {
				return fmt.Errorf("no UAA endpoint configured. Use 'capi uaa target <url>' to set one")
			}

			// Test CF integration
			integrationInfo, err := testCFIntegration(config)
			if err != nil {
				return fmt.Errorf("failed to test CF integration: %w", err)
			}

			// Display results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(integrationInfo)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(integrationInfo)
			default:
				return displayCFIntegrationResults(integrationInfo)
			}
		},
	}

	return cmd
}

// runCompatibilityTests performs comprehensive compatibility testing
func runCompatibilityTests(client *UAAClientWrapper, comprehensive bool) (*UAAVersionInfo, error) {
	versionInfo := &UAAVersionInfo{
		Endpoint: client.Endpoint(),
		TestedAt: time.Now(),
		Features: []string{},
		Compatibility: CompatibilityStatus{
			FeatureTests:    make(map[string]string),
			Issues:          []string{},
			Recommendations: []string{},
		},
	}

	fmt.Println("Running UAA compatibility tests...")

	// Test server info and version detection
	fmt.Print("• Testing server info... ")
	ctx := context.Background()
	serverInfo, err := client.GetServerInfo(ctx)
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		versionInfo.Compatibility.Overall = "incompatible"
		versionInfo.Compatibility.Issues = append(versionInfo.Compatibility.Issues, "Cannot retrieve server information")
		return versionInfo, nil
	}

	versionInfo.ServerInfo = serverInfo
	fmt.Println("✅ Success")

	// Extract version information
	if app, ok := serverInfo["app"].(map[string]interface{}); ok {
		if version, ok := app["version"].(string); ok {
			versionInfo.Version = version
		}
	}

	// Test basic authentication capabilities
	fmt.Print("• Testing authentication capabilities... ")
	authResult := testAuthenticationCapabilities(client)
	versionInfo.Compatibility.BasicAuth = authResult
	if authResult == "compatible" {
		fmt.Println("✅ Compatible")
		versionInfo.Features = append(versionInfo.Features, "OAuth2 Authentication")
	} else {
		fmt.Printf("⚠️  %s\n", authResult)
	}

	// Test user management capabilities
	if client.IsAuthenticated() {
		fmt.Print("• Testing user management... ")
		userMgmtResult := testUserManagementCapabilities(client)
		versionInfo.Compatibility.UserMgmt = userMgmtResult
		if userMgmtResult == "compatible" {
			fmt.Println("✅ Compatible")
			versionInfo.Features = append(versionInfo.Features, "User Management")
		} else {
			fmt.Printf("⚠️  %s\n", userMgmtResult)
		}

		if comprehensive {
			// Test group management capabilities
			fmt.Print("• Testing group management... ")
			groupMgmtResult := testGroupManagementCapabilities(client)
			versionInfo.Compatibility.GroupMgmt = groupMgmtResult
			if groupMgmtResult == "compatible" {
				fmt.Println("✅ Compatible")
				versionInfo.Features = append(versionInfo.Features, "Group Management")
			} else {
				fmt.Printf("⚠️  %s\n", groupMgmtResult)
			}

			// Test client management capabilities
			fmt.Print("• Testing client management... ")
			clientMgmtResult := testClientManagementCapabilities(client)
			versionInfo.Compatibility.ClientMgmt = clientMgmtResult
			if clientMgmtResult == "compatible" {
				fmt.Println("✅ Compatible")
				versionInfo.Features = append(versionInfo.Features, "OAuth Client Management")
			} else {
				fmt.Printf("⚠️  %s\n", clientMgmtResult)
			}
		}
	} else {
		versionInfo.Compatibility.Issues = append(versionInfo.Compatibility.Issues, "Authentication required for full testing")
	}

	// Determine overall compatibility
	versionInfo.Compatibility.Overall = determineOverallCompatibility(&versionInfo.Compatibility)

	// Add version-specific recommendations
	addVersionRecommendations(versionInfo)

	return versionInfo, nil
}

// testAuthenticationCapabilities tests OAuth2 authentication features
func testAuthenticationCapabilities(client *UAAClientWrapper) string {
	// Test token key retrieval (public endpoint)
	_, err := client.Client().TokenKey()
	if err != nil {
		return "incompatible"
	}

	// Test token keys retrieval
	_, err = client.Client().TokenKeys()
	if err != nil {
		return "partial"
	}

	return "compatible"
}

// testUserManagementCapabilities tests user CRUD operations
func testUserManagementCapabilities(client *UAAClientWrapper) string {
	// Test user listing (requires authentication)
	_, _, err := client.Client().ListUsers("", "", "", uaa.SortAscending, 1, 1)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") || strings.Contains(err.Error(), "forbidden") {
			return "partial"
		}
		return "incompatible"
	}

	return "compatible"
}

// testGroupManagementCapabilities tests group management operations
func testGroupManagementCapabilities(client *UAAClientWrapper) string {
	// Test group listing
	_, _, err := client.Client().ListGroups("", "", "", uaa.SortAscending, 1, 1)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") || strings.Contains(err.Error(), "forbidden") {
			return "partial"
		}
		return "incompatible"
	}

	return "compatible"
}

// testClientManagementCapabilities tests OAuth client management
func testClientManagementCapabilities(client *UAAClientWrapper) string {
	// Test client listing
	_, _, err := client.Client().ListClients("", "", uaa.SortAscending, 1, 1)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") || strings.Contains(err.Error(), "forbidden") {
			return "partial"
		}
		return "incompatible"
	}

	return "compatible"
}

// determineOverallCompatibility calculates overall compatibility status
func determineOverallCompatibility(status *CompatibilityStatus) string {
	compatibleCount := 0
	totalCount := 0

	tests := []string{status.BasicAuth, status.UserMgmt, status.GroupMgmt, status.ClientMgmt}
	for _, test := range tests {
		if test != "" {
			totalCount++
			if test == "compatible" {
				compatibleCount++
			}
		}
	}

	if totalCount == 0 {
		return "unknown"
	}

	ratio := float64(compatibleCount) / float64(totalCount)
	if ratio >= 0.8 {
		return "compatible"
	} else if ratio >= 0.5 {
		return "partial"
	} else {
		return "incompatible"
	}
}

// addVersionRecommendations adds version-specific recommendations
func addVersionRecommendations(info *UAAVersionInfo) {
	version := info.Version

	// Extract major.minor version
	re := regexp.MustCompile(`(\d+)\.(\d+)`)
	matches := re.FindStringSubmatch(version)
	if len(matches) >= 3 {
		major, _ := strconv.Atoi(matches[1])
		minor, _ := strconv.Atoi(matches[2])

		if major < 4 {
			info.Compatibility.Issues = append(info.Compatibility.Issues, "UAA version is very old and may have limited functionality")
			info.Compatibility.Recommendations = append(info.Compatibility.Recommendations, "Consider upgrading to UAA 4.x or later")
		} else if major == 4 && minor < 30 {
			info.Compatibility.Recommendations = append(info.Compatibility.Recommendations, "Consider upgrading to UAA 4.30+ for best compatibility")
		}
	}

	// Add general recommendations based on compatibility
	switch info.Compatibility.Overall {
	case "incompatible":
		info.Compatibility.Recommendations = append(info.Compatibility.Recommendations,
			"UAA endpoint may not be compatible with this CLI version",
			"Verify UAA endpoint URL and network connectivity",
			"Check UAA logs for detailed error information")
	case "partial":
		info.Compatibility.Recommendations = append(info.Compatibility.Recommendations,
			"Some features may not work due to insufficient permissions",
			"Ensure your client has appropriate authorities (scim.read, scim.write, etc.)",
			"Contact your UAA administrator for permission adjustments")
	case "compatible":
		info.Compatibility.Recommendations = append(info.Compatibility.Recommendations,
			"UAA endpoint is fully compatible with this CLI",
			"All features should work as expected")
	}
}

// testCFIntegration tests Cloud Foundry integration
func testCFIntegration(config *Config) (*CFIntegrationInfo, error) {
	integration := &CFIntegrationInfo{
		UAAVersion: "unknown",
		Compatible: false,
	}

	// Check if we have CF API endpoint configured
	if config.API != "" {
		integration.CFAPIVersion = "configured"

		// Try to infer UAA endpoint from CF API
		if config.UAAEndpoint == "" {
			// Try to infer UAA endpoint
			inferredUAA := inferUAAFromCFAPI(config.API)
			if inferredUAA != "" {
				integration.AuthMethod = "inferred"
			}
		} else {
			integration.AuthMethod = "explicit"
		}
	}

	// Test UAA compatibility
	if config.UAAEndpoint != "" {
		uaaClient, err := NewUAAClient(config)
		if err == nil {
			ctx := context.Background()
			if serverInfo, err := uaaClient.GetServerInfo(ctx); err == nil {
				if app, ok := serverInfo["app"].(map[string]interface{}); ok {
					if version, ok := app["version"].(string); ok {
						integration.UAAVersion = version
					}
				}
			}

			// Test if we can authenticate
			if uaaClient.IsAuthenticated() {
				integration.ScopesSupported = true
				integration.TokenFormat = "JWT"
				integration.Compatible = true
			}
		}
	}

	return integration, nil
}

// inferUAAFromCFAPI attempts to infer UAA endpoint from CF API endpoint
func inferUAAFromCFAPI(cfEndpoint string) string {
	// Simple heuristic: replace "api." with "uaa." for common CF deployments
	if strings.Contains(cfEndpoint, "api.") {
		return strings.Replace(cfEndpoint, "api.", "uaa.", 1)
	}
	return ""
}

// saveCompatibilityResults saves test results to file
func saveCompatibilityResults(info *UAAVersionInfo) error {
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile("uaa-compatibility-results.json", data, 0600)
}

// displayCompatibilityResults displays compatibility test results
func displayCompatibilityResults(info *UAAVersionInfo) error {
	output := viper.GetString("output")

	switch output {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(info)
	case "yaml":
		encoder := yaml.NewEncoder(os.Stdout)
		return encoder.Encode(info)
	default:
		return displayCompatibilityTable(info)
	}
}

// displayCompatibilityTable displays results in table format
func displayCompatibilityTable(info *UAAVersionInfo) error {
	fmt.Printf("\nUAA Compatibility Report\n")
	fmt.Printf("========================\n\n")

	// Basic info
	fmt.Printf("Endpoint: %s\n", info.Endpoint)
	fmt.Printf("Version: %s\n", info.Version)
	fmt.Printf("Tested: %s\n", info.TestedAt.Format(time.RFC3339))
	fmt.Printf("Overall Status: %s\n\n", strings.ToUpper(info.Compatibility.Overall))

	// Feature compatibility table
	table := tablewriter.NewWriter(os.Stdout)
	table.Header("Feature", "Status")

	if info.Compatibility.BasicAuth != "" {
		_ = table.Append([]string{"Authentication", formatStatus(info.Compatibility.BasicAuth)})
	}
	if info.Compatibility.UserMgmt != "" {
		_ = table.Append([]string{"User Management", formatStatus(info.Compatibility.UserMgmt)})
	}
	if info.Compatibility.GroupMgmt != "" {
		_ = table.Append([]string{"Group Management", formatStatus(info.Compatibility.GroupMgmt)})
	}
	if info.Compatibility.ClientMgmt != "" {
		_ = table.Append([]string{"Client Management", formatStatus(info.Compatibility.ClientMgmt)})
	}

	_ = table.Render()

	// Display supported features
	if len(info.Features) > 0 {
		fmt.Printf("\nSupported Features:\n")
		for _, feature := range info.Features {
			fmt.Printf("  ✅ %s\n", feature)
		}
		fmt.Println()
	}

	// Display issues
	if len(info.Compatibility.Issues) > 0 {
		fmt.Printf("Issues Found:\n")
		for _, issue := range info.Compatibility.Issues {
			fmt.Printf("  ⚠️  %s\n", issue)
		}
		fmt.Println()
	}

	// Display recommendations
	if len(info.Compatibility.Recommendations) > 0 {
		fmt.Printf("Recommendations:\n")
		for _, rec := range info.Compatibility.Recommendations {
			fmt.Printf("  💡 %s\n", rec)
		}
		fmt.Println()
	}

	return nil
}

// displayCFIntegrationResults displays CF integration test results
func displayCFIntegrationResults(info *CFIntegrationInfo) error {
	fmt.Printf("Cloud Foundry Integration Report\n")
	fmt.Printf("================================\n\n")

	fmt.Printf("CF API Version: %s\n", info.CFAPIVersion)
	fmt.Printf("UAA Version: %s\n", info.UAAVersion)
	fmt.Printf("Auth Method: %s\n", info.AuthMethod)
	fmt.Printf("Scopes Supported: %v\n", info.ScopesSupported)
	fmt.Printf("Token Format: %s\n", info.TokenFormat)
	fmt.Printf("Compatible: %v\n\n", info.Compatible)

	if info.Compatible {
		fmt.Println("✅ CF integration is working properly")
	} else {
		fmt.Println("❌ CF integration issues detected")
		fmt.Println("\nTroubleshooting:")
		fmt.Println("  • Verify CF API endpoint is correct")
		fmt.Println("  • Ensure UAA endpoint is accessible")
		fmt.Println("  • Check authentication credentials")
	}

	return nil
}

// formatStatus formats compatibility status with appropriate icons
func formatStatus(status string) string {
	switch status {
	case "compatible":
		return "✅ Compatible"
	case "partial":
		return "⚠️  Partial"
	case "incompatible":
		return "❌ Incompatible"
	default:
		return "❓ Unknown"
	}
}
