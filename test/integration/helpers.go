// +build integration

package integration

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestConfig holds configuration for integration tests
type TestConfig struct {
	UAAEndpoint   string
	ClientID      string
	ClientSecret  string
	AdminUser     string
	AdminPassword string
	CapiPath      string
	Verbose       bool
}

// LoadTestConfig loads configuration from environment variables
func LoadTestConfig() *TestConfig {
	return &TestConfig{
		UAAEndpoint:   os.Getenv("UAA_ENDPOINT"),
		ClientID:      os.Getenv("UAA_CLIENT_ID"), 
		ClientSecret:  os.Getenv("UAA_CLIENT_SECRET"),
		AdminUser:     os.Getenv("UAA_ADMIN_USER"),
		AdminPassword: os.Getenv("UAA_ADMIN_PASSWORD"),
		CapiPath:      getCapiPath(),
		Verbose:       os.Getenv("CAPI_VERBOSE") == "true",
	}
}

// getCapiPath determines the path to the capi binary
func getCapiPath() string {
	if path := os.Getenv("CAPI_BINARY_PATH"); path != "" {
		return path
	}
	
	// Try common locations
	candidates := []string{
		"../../capi",
		"./capi",
		"../capi",
	}
	
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	
	return "capi" // Fallback to PATH
}

// SkipIfMissingConfig skips test if required config is missing
func (config *TestConfig) SkipIfMissingConfig(t *testing.T) {
	if config.UAAEndpoint == "" {
		t.Skip("UAA_ENDPOINT not set, skipping integration test")
	}
	
	if _, err := os.Stat(config.CapiPath); os.IsNotExist(err) {
		t.Skipf("capi binary not found at %s, skipping integration test", config.CapiPath)
	}
}

// CommandRunner provides utilities for running capi commands
type CommandRunner struct {
	config *TestConfig
	t      *testing.T
}

// NewCommandRunner creates a new command runner
func NewCommandRunner(config *TestConfig, t *testing.T) *CommandRunner {
	return &CommandRunner{
		config: config,
		t:      t,
	}
}

// Run executes a capi command and returns output
func (runner *CommandRunner) Run(args ...string) (stdout, stderr string, err error) {
	cmd := exec.Command(runner.config.CapiPath, args...)
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	
	if runner.config.Verbose {
		runner.t.Logf("Running: %s %s", runner.config.CapiPath, strings.Join(args, " "))
	}
	
	err = cmd.Run()
	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()
	
	if runner.config.Verbose && err != nil {
		runner.t.Logf("Command failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}
	
	return stdout, stderr, err
}

// RunWithInput executes a capi command with stdin input
func (runner *CommandRunner) RunWithInput(input string, args ...string) (stdout, stderr string, err error) {
	cmd := exec.Command(runner.config.CapiPath, args...)
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	cmd.Stdin = strings.NewReader(input)
	
	if runner.config.Verbose {
		runner.t.Logf("Running with input: %s %s", runner.config.CapiPath, strings.Join(args, " "))
	}
	
	err = cmd.Run()
	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()
	
	if runner.config.Verbose && err != nil {
		runner.t.Logf("Command failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}
	
	return stdout, stderr, err
}

// SetupUAATarget configures the UAA endpoint
func (runner *CommandRunner) SetupUAATarget() error {
	_, stderr, err := runner.Run("users", "target", runner.config.UAAEndpoint)
	if err != nil {
		return fmt.Errorf("failed to set UAA target: %s", stderr)
	}
	return nil
}

// AuthenticateAdmin authenticates as admin for test operations
func (runner *CommandRunner) AuthenticateAdmin() error {
	if runner.config.AdminUser != "" && runner.config.AdminPassword != "" {
		// Use password grant
		_, stderr, err := runner.Run("users", "get-password-token",
			"--username", runner.config.AdminUser,
			"--password", runner.config.AdminPassword,
			"--client-id", runner.config.ClientID,
			"--client-secret", runner.config.ClientSecret)
		if err != nil {
			return fmt.Errorf("failed to authenticate with password grant: %s", stderr)
		}
	} else if runner.config.ClientID != "" && runner.config.ClientSecret != "" {
		// Use client credentials
		_, stderr, err := runner.Run("users", "get-client-credentials-token",
			"--client-id", runner.config.ClientID,
			"--client-secret", runner.config.ClientSecret)
		if err != nil {
			return fmt.Errorf("failed to authenticate with client credentials: %s", stderr)
		}
	} else {
		return fmt.Errorf("no authentication credentials provided")
	}
	return nil
}

// GenerateTestName creates a unique test resource name
func GenerateTestName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().Unix())
}

// CleanupResource attempts to delete a test resource
func (runner *CommandRunner) CleanupResource(resourceType, name string) {
	var args []string
	switch resourceType {
	case "user":
		args = []string{"users", "delete-user", name, "--force"}
	case "group":
		args = []string{"users", "delete-group", name, "--force"}
	case "client":
		args = []string{"users", "delete-client", name, "--force"}
	default:
		runner.t.Logf("Unknown resource type for cleanup: %s", resourceType)
		return
	}
	
	stdout, stderr, err := runner.Run(args...)
	if err != nil && runner.config.Verbose {
		runner.t.Logf("Cleanup warning for %s %s: %s\nStderr: %s", resourceType, name, stdout, stderr)
	}
}

// WaitForCondition waits for a condition to be met with timeout
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, message string) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	
	timeoutChan := time.After(timeout)
	
	for {
		select {
		case <-ticker.C:
			if condition() {
				return
			}
		case <-timeoutChan:
			t.Fatalf("Timeout waiting for condition: %s", message)
		}
	}
}

// AssertJSONOutput verifies command output is valid JSON
func AssertJSONOutput(t *testing.T, output string) {
	// Basic JSON validation
	output = strings.TrimSpace(output)
	if !strings.HasPrefix(output, "{") && !strings.HasPrefix(output, "[") {
		t.Errorf("Output does not appear to be JSON: %s", output)
	}
}

// AssertYAMLOutput verifies command output is valid YAML
func AssertYAMLOutput(t *testing.T, output string) {
	// Basic YAML validation
	output = strings.TrimSpace(output)
	if strings.Contains(output, "---") || strings.Contains(output, ":") {
		return // Looks like YAML
	}
	t.Errorf("Output does not appear to be YAML: %s", output)
}