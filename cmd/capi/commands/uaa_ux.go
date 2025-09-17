package commands

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fivetwenty-io/capi/v3/internal/constants"
)

// ProgressIndicator provides visual feedback for long-running operations.
type ProgressIndicator struct {
	message    string
	isRunning  bool
	stopChan   chan bool
	characters []string
	position   int
}

// NewProgressIndicator creates a new progress indicator.
func NewProgressIndicator(message string) *ProgressIndicator {
	return &ProgressIndicator{
		message:    message,
		isRunning:  false,
		stopChan:   make(chan bool),
		characters: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		position:   0,
	}
}

// Start begins the progress indicator animation.
func (p *ProgressIndicator) Start() {
	if p.isRunning {
		return
	}

	p.isRunning = true

	go func() {
		for {
			select {
			case <-p.stopChan:
				return
			default:
				_, _ = fmt.Fprintf(os.Stdout, "\r%s %s", p.characters[p.position], p.message)
				p.position = (p.position + 1) % len(p.characters)

				time.Sleep(constants.UIUpdateInterval)
			}
		}
	}()
}

// Stop ends the progress indicator and clears the line.
func (p *ProgressIndicator) Stop() {
	if !p.isRunning {
		return
	}

	p.isRunning = false
	p.stopChan <- true

	// Clear the progress line
	_, _ = fmt.Fprintf(os.Stdout, "\r%s\r", strings.Repeat(" ", len(p.message)+constants.UIMessageSpacing))
}

// Success displays a success message and stops the indicator.
func (p *ProgressIndicator) Success(message string) {
	p.Stop()

	_, _ = fmt.Fprintf(os.Stdout, "✓ %s\n", message)
}

// Error displays an error message and stops the indicator.
func (p *ProgressIndicator) Error(message string) {
	p.Stop()

	_, _ = fmt.Fprintf(os.Stdout, "✗ %s\n", message)
}

// Warning displays a warning message and stops the indicator.
func (p *ProgressIndicator) Warning(message string) {
	p.Stop()

	_, _ = fmt.Fprintf(os.Stdout, "⚠ %s\n", message)
}

// EnhancedError provides better error messages with context and suggestions.
type EnhancedError struct {
	Operation   string
	Cause       error
	Suggestions []string
	Context     map[string]string
}

// NewEnhancedError creates a new enhanced error with helpful context.
func NewEnhancedError(operation string, cause error) *EnhancedError {
	return &EnhancedError{
		Operation: operation,
		Cause:     cause,
		Context:   make(map[string]string),
	}
}

// Error implements the error interface.
func (e *EnhancedError) Error() string {
	var msg strings.Builder

	// Main error message
	msg.WriteString("Failed to " + e.Operation)

	if e.Cause != nil {
		msg.WriteString(": " + e.Cause.Error())
	}

	// Add context information
	if len(e.Context) > 0 {
		msg.WriteString("\n\nContext:")

		for key, value := range e.Context {
			msg.WriteString(fmt.Sprintf("\n  %s: %s", key, value))
		}
	}

	// Add suggestions
	if len(e.Suggestions) > 0 {
		msg.WriteString("\n\nSuggestions:")

		for _, suggestion := range e.Suggestions {
			msg.WriteString("\n  • " + suggestion)
		}
	}

	return msg.String()
}

// AddContext adds contextual information to the error.
func (e *EnhancedError) AddContext(key, value string) *EnhancedError {
	e.Context[key] = value

	return e
}

// AddSuggestion adds a helpful suggestion to resolve the error.
func (e *EnhancedError) AddSuggestion(suggestion string) *EnhancedError {
	e.Suggestions = append(e.Suggestions, suggestion)

	return e
}

// CreateCommonUAAError creates enhanced errors for common UAA scenarios.
func CreateCommonUAAError(operation string, cause error, endpoint string) error {
	enhancedErr := NewEnhancedError(operation, cause)

	if endpoint != "" {
		_ = enhancedErr.AddContext("UAA Endpoint", endpoint)
	}

	// Add common suggestions based on error type
	errMsg := strings.ToLower(cause.Error())

	switch {
	case strings.Contains(errMsg, "not authenticated") || strings.Contains(errMsg, "unauthorized"):
		_ = enhancedErr.AddSuggestion("Run 'capi uaa get-client-credentials-token' to authenticate")
		_ = enhancedErr.AddSuggestion("Or run 'capi uaa get-password-token' for user authentication")
		_ = enhancedErr.AddSuggestion("Check that your client has the required scopes/authorities")

	case strings.Contains(errMsg, "forbidden") || strings.Contains(errMsg, "insufficient"):
		_ = enhancedErr.AddSuggestion("Your client may not have sufficient authorities for this operation")
		_ = enhancedErr.AddSuggestion("Contact your UAA administrator to grant additional permissions")
		_ = enhancedErr.AddSuggestion("Try using a client with 'uaa.admin' authority")

	case strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "404"):
		_ = enhancedErr.AddSuggestion("Verify the resource name/ID is correct")
		_ = enhancedErr.AddSuggestion("Use list commands to find available resources")
		_ = enhancedErr.AddSuggestion("Check that the UAA endpoint is correct")

	case strings.Contains(errMsg, "connection") || strings.Contains(errMsg, "timeout"):
		_ = enhancedErr.AddSuggestion("Check network connectivity to the UAA endpoint")
		_ = enhancedErr.AddSuggestion("Verify the UAA endpoint URL is correct")
		_ = enhancedErr.AddSuggestion("Try using --skip-ssl-validation for development environments")

	case strings.Contains(errMsg, "certificate") || strings.Contains(errMsg, "ssl") || strings.Contains(errMsg, "tls"):
		_ = enhancedErr.AddSuggestion("SSL certificate verification failed")
		_ = enhancedErr.AddSuggestion("Use --skip-ssl-validation flag for development environments")
		_ = enhancedErr.AddSuggestion("Ensure the UAA endpoint has a valid SSL certificate")

	case strings.Contains(errMsg, "invalid") || strings.Contains(errMsg, "bad request"):
		_ = enhancedErr.AddSuggestion("Check that all required parameters are provided")
		_ = enhancedErr.AddSuggestion("Verify parameter formats and values")
		_ = enhancedErr.AddSuggestion("Use 'capi uaa <command> --help' for usage examples")

	default:
		_ = enhancedErr.AddSuggestion("Check 'capi uaa context' to verify authentication status")
		_ = enhancedErr.AddSuggestion("Ensure the UAA endpoint is accessible")
		_ = enhancedErr.AddSuggestion("Try re-authenticating with fresh credentials")
	}

	return enhancedErr
}

// WrapWithProgress wraps a function with a progress indicator.
func WrapWithProgress(message string, fn func() error) error {
	progress := NewProgressIndicator(message)
	progress.Start()

	err := fn()
	if err != nil {
		progress.Error("Operation failed")

		return err
	}

	progress.Success("Operation completed")

	return nil
}

// ConfirmAction prompts the user for confirmation before dangerous operations.
func ConfirmAction(message string, force bool) bool {
	if force {
		return true
	}

	_, _ = fmt.Fprintf(os.Stdout, "%s (y/N): ", message)

	var response string

	_, _ = fmt.Scanln(&response)

	response = strings.ToLower(strings.TrimSpace(response))

	return response == "y" || response == "yes"
}

// FormatResourceCount formats resource counts with appropriate pluralization.
func FormatResourceCount(count int, singular, plural string) string {
	if count == 1 {
		return fmt.Sprintf("%d %s", count, singular)
	}

	return fmt.Sprintf("%d %s", count, plural)
}

// TruncateString truncates a string to a specified length with ellipsis.
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}

	return s[:maxLength-3] + "..."
}

// StatusIndicator provides visual status indicators.
func StatusIndicator(status bool) string {
	if status {
		return "✓"
	}

	return "✗"
}

// FormatDuration formats a duration in a human-readable way.
func FormatDuration(duration time.Duration) string {
	if duration < time.Minute {
		return fmt.Sprintf("%.1fs", duration.Seconds())
	}

	if duration < time.Hour {
		return fmt.Sprintf("%.1fm", duration.Minutes())
	}

	return fmt.Sprintf("%.1fh", duration.Hours())
}
