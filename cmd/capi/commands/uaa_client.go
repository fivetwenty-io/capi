package commands

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/cloudfoundry-community/go-uaa"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

// UAAClientWrapper provides a convenient wrapper around the go-uaa client
// with configuration integration and token management
type UAAClientWrapper struct {
	client   *uaa.API
	config   *Config
	endpoint string
	skipSSL  bool
	verbose  bool
}

// UAATokenInfo represents token information for display
type UAATokenInfo struct {
	AccessToken  string    `json:"access_token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type,omitempty"`
	ExpiresIn    int       `json:"expires_in,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	Scope        string    `json:"scope,omitempty"`
	JTI          string    `json:"jti,omitempty"`
}

// isDevelopmentEnvironment checks if we're in a development environment
func isDevelopmentEnvironment() bool {
	devMode := os.Getenv("CAPI_DEV_MODE")
	return devMode == "true" || devMode == "1"
}

// NewUAAClient creates a new UAA client wrapper
func NewUAAClient(config *Config) (*UAAClientWrapper, error) {
	if config == nil {
		config = loadConfig()
	}

	wrapper := &UAAClientWrapper{
		config:  config,
		verbose: viper.GetBool("verbose"),
		skipSSL: config.SkipSSLValidation,
	}

	// Determine UAA endpoint
	uaaEndpoint, err := wrapper.discoverUAAEndpoint()
	if err != nil {
		return nil, fmt.Errorf("failed to discover UAA endpoint: %w", err)
	}
	wrapper.endpoint = uaaEndpoint

	// Create UAA API client
	client, err := wrapper.createUAAClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create UAA client: %w", err)
	}
	wrapper.client = client

	return wrapper, nil
}

// NewUAAClientWithEndpoint creates a UAA client with a specific endpoint
func NewUAAClientWithEndpoint(endpoint string, config *Config) (*UAAClientWrapper, error) {
	if config == nil {
		config = loadConfig()
	}

	wrapper := &UAAClientWrapper{
		config:   config,
		endpoint: endpoint,
		verbose:  viper.GetBool("verbose"),
		skipSSL:  config.SkipSSLValidation,
	}

	// Create UAA API client
	client, err := wrapper.createUAAClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create UAA client: %w", err)
	}
	wrapper.client = client

	return wrapper, nil
}

// Client returns the underlying UAA API client
func (w *UAAClientWrapper) Client() *uaa.API {
	return w.client
}

// Endpoint returns the current UAA endpoint
func (w *UAAClientWrapper) Endpoint() string {
	return w.endpoint
}

// IsAuthenticated checks if the client has valid authentication
func (w *UAAClientWrapper) IsAuthenticated() bool {
	return w.config.UAAToken != "" || w.config.Token != ""
}

// SetToken sets the authentication token for UAA requests
func (w *UAAClientWrapper) SetToken(token string) {
	w.config.UAAToken = token
	// Note: The go-uaa client handles authentication internally via its configuration
}

// GetToken retrieves the current authentication token
func (w *UAAClientWrapper) GetToken() string {
	if w.config.UAAToken != "" {
		return w.config.UAAToken
	}
	// Fallback to CF token if UAA token is not available
	return w.config.Token
}

// discoverUAAEndpoint attempts to discover the UAA endpoint from various sources
func (w *UAAClientWrapper) discoverUAAEndpoint() (string, error) {
	// 1. Check if UAA endpoint is explicitly configured
	if w.config.UAAEndpoint != "" {
		return w.config.UAAEndpoint, nil
	}

	// 2. Try to discover from CF API endpoint
	if w.config.API != "" {
		uaaEndpoint, err := w.discoverFromCFAPI()
		if err == nil && uaaEndpoint != "" {
			return uaaEndpoint, nil
		}
		if w.verbose {
			fmt.Printf("Failed to discover UAA from CF API: %v\n", err)
		}
	}

	// 3. Try to infer from CF API endpoint (common patterns)
	if w.config.API != "" {
		if inferredEndpoint := w.inferUAAEndpoint(); inferredEndpoint != "" {
			return inferredEndpoint, nil
		}
	}

	return "", fmt.Errorf("no UAA endpoint configured and unable to discover from CF API endpoint")
}

// discoverFromCFAPI attempts to discover UAA endpoint from CF API info endpoint
func (w *UAAClientWrapper) discoverFromCFAPI() (string, error) {
	if w.config.API == "" {
		return "", fmt.Errorf("no CF API endpoint configured")
	}

	// Create HTTP client with appropriate SSL settings
	var transport *http.Transport
	if w.skipSSL {
		if !isDevelopmentEnvironment() {
			return "", fmt.Errorf("skipSSL is only allowed in development environments (set CAPI_DEV_MODE=true)")
		}
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // #nosec G402 -- Protected by development environment check above
			},
		}
	} else {
		transport = &http.Transport{}
	}

	httpClient := &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}

	// Try to get CF API info
	infoURL := strings.TrimSuffix(w.config.API, "/") + "/v2/info"
	resp, err := httpClient.Get(infoURL)
	if err != nil {
		return "", fmt.Errorf("failed to get CF API info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("CF API info returned status %d", resp.StatusCode)
	}

	// Parse JSON response to extract authorization_endpoint
	// For now, we'll use a simple approach - in a full implementation,
	// we would parse the JSON properly
	// This is a simplified version that assumes standard CF deployment patterns

	return "", fmt.Errorf("JSON parsing not implemented in this version")
}

// inferUAAEndpoint attempts to infer UAA endpoint from CF API endpoint using common patterns
func (w *UAAClientWrapper) inferUAAEndpoint() string {
	if w.config.API == "" {
		return ""
	}

	parsedURL, err := url.Parse(w.config.API)
	if err != nil {
		return ""
	}

	// Common patterns for UAA endpoints
	patterns := []string{
		// Replace "api." with "uaa."
		strings.Replace(parsedURL.Host, "api.", "uaa.", 1),
		// Replace "api" with "uaa"
		strings.Replace(parsedURL.Host, "api", "uaa", 1),
		// Append ":8443" for UAA (if not HTTPS)
		parsedURL.Host + ":8443",
	}

	for _, pattern := range patterns {
		if pattern != parsedURL.Host && pattern != "" {
			uaaURL := &url.URL{
				Scheme: "https", // UAA typically uses HTTPS
				Host:   pattern,
			}
			return uaaURL.String()
		}
	}

	return ""
}

// createUAAClient creates the underlying UAA API client
func (w *UAAClientWrapper) createUAAClient() (*uaa.API, error) {
	if w.endpoint == "" {
		return nil, fmt.Errorf("no UAA endpoint specified")
	}

	// Create HTTP client with appropriate settings
	var transport *http.Transport
	if w.skipSSL {
		if !isDevelopmentEnvironment() {
			return nil, fmt.Errorf("skipSSL is only allowed in development environments (set CAPI_DEV_MODE=true)")
		}
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // #nosec G402 -- Protected by development environment check above
			},
		}
	} else {
		transport = &http.Transport{}
	}

	httpClient := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}

	// Determine authentication method
	var authOpt uaa.AuthenticationOption
	if token := w.GetToken(); token != "" {
		// Create oauth2.Token from string
		oauthToken := &oauth2.Token{
			AccessToken: token,
		}
		authOpt = uaa.WithToken(oauthToken)
	} else {
		// No authentication for now
		authOpt = uaa.WithNoAuthentication()
	}

	// Create UAA API client
	client, err := uaa.New(
		w.endpoint,
		authOpt,
		uaa.WithClient(httpClient),
		uaa.WithVerbosity(w.verbose),
		uaa.WithSkipSSLValidation(w.skipSSL),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create UAA client: %w", err)
	}

	return client, nil
}

// RefreshToken attempts to refresh the current access token
func (w *UAAClientWrapper) RefreshToken(ctx context.Context) (*UAATokenInfo, error) {
	// Use the client's built-in token refresh capability
	token, err := w.client.Token(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get/refresh token: %w", err)
	}

	// Update stored tokens
	w.config.UAAToken = token.AccessToken
	if token.RefreshToken != "" {
		w.config.UAARefreshToken = token.RefreshToken
	}

	// Convert to our token info structure
	tokenInfo := &UAATokenInfo{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
	}

	if !token.Expiry.IsZero() {
		tokenInfo.ExpiresAt = token.Expiry
		tokenInfo.ExpiresIn = int(time.Until(token.Expiry).Seconds())
	}

	return tokenInfo, nil
}

// TestConnection tests the connection to the UAA endpoint
func (w *UAAClientWrapper) TestConnection(ctx context.Context) error {
	if w.client == nil {
		return fmt.Errorf("UAA client not initialized")
	}

	// Try to get server info as a connectivity test
	_, err := w.client.GetInfo()
	if err != nil {
		return fmt.Errorf("failed to connect to UAA: %w", err)
	}

	return nil
}

// GetServerInfo retrieves UAA server information
func (w *UAAClientWrapper) GetServerInfo(ctx context.Context) (map[string]interface{}, error) {
	if w.client == nil {
		return nil, fmt.Errorf("UAA client not initialized")
	}

	info, err := w.client.GetInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get server info: %w", err)
	}

	// Convert to map for easier handling
	result := map[string]interface{}{
		"app":       info.App,
		"commit_id": info.CommitID,
		"timestamp": info.Timestamp,
		"links":     info.Links,
		"zone_name": info.ZoneName,
		"entity_id": info.EntityID,
	}

	return result, nil
}

// SaveConfig saves the current configuration including UAA tokens
func (w *UAAClientWrapper) SaveConfig() error {
	return saveConfigStruct(w.config)
}
