// Package cfclient provides the main entry point for creating Cloud Foundry API clients
package cfclient

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/fivetwenty-io/capi/v3/internal/client"
	"github.com/fivetwenty-io/capi/v3/internal/constants"
	"github.com/fivetwenty-io/capi/v3/pkg/capi"
)

// New creates a new Cloud Foundry API client with automatic UAA discovery.
func New(ctx context.Context, config *capi.Config) (capi.Client, error) {
	if config == nil {
		return nil, capi.ErrConfigRequired
	}

	if config.APIEndpoint == "" {
		return nil, capi.ErrAPIEndpointRequired
	}

	// Normalize API endpoint
	apiEndpoint := strings.TrimSuffix(config.APIEndpoint, "/")
	if !strings.HasPrefix(apiEndpoint, "http://") && !strings.HasPrefix(apiEndpoint, "https://") {
		apiEndpoint = "https://" + apiEndpoint
	}

	config.APIEndpoint = apiEndpoint

	// If we need authentication and don't have a token URL, discover the UAA endpoint
	if needsAuth(config) && config.TokenURL == "" {
		uaaURL, err := discoverUAAEndpoint(ctx, apiEndpoint, config.SkipTLSVerify)
		if err != nil {
			return nil, fmt.Errorf("discovering UAA endpoint: %w", err)
		}

		// Set the token URL for OAuth2
		config.TokenURL = strings.TrimSuffix(uaaURL, "/") + "/oauth/token"
	}

	// Enable fetching API links on init for better log support
	config.FetchAPILinksOnInit = true

	// Use the internal client implementation
	client, err := client.New(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create new client: %w", err)
	}

	return client, nil
}

// needsAuth checks if the config requires authentication.
func needsAuth(config *capi.Config) bool {
	return config.AccessToken == "" &&
		(config.Username != "" || config.ClientID != "" || config.RefreshToken != "")
}

// isDevelopmentEnvironment checks if we're in a development environment.
func isDevelopmentEnvironment() bool {
	devMode := os.Getenv("CAPI_DEV_MODE")

	return devMode == "true" || devMode == "1"
}

// discoverUAAEndpoint discovers the UAA endpoint from the CF API root.
// createDiscoveryHTTPClient creates an HTTP client for UAA endpoint discovery.
func createDiscoveryHTTPClient(skipTLS bool) (*http.Client, error) {
	httpClient := &http.Client{
		Timeout: constants.ShortHTTPTimeout,
	}

	if skipTLS {
		// Only allow insecure TLS in explicit development environments
		if isDevelopmentEnvironment() {
			httpClient.Transport = &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // #nosec G402 -- Protected by development environment check above
			}
		} else {
			return nil, fmt.Errorf("%w (set CAPI_DEV_MODE=true)", capi.ErrSkipTLSOnlyInDev)
		}
	}

	return httpClient, nil
}

// fetchRootInfo fetches and parses the root info from the API endpoint.
func fetchRootInfo(ctx context.Context, httpClient *http.Client, apiEndpoint string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiEndpoint+"/", nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("getting root info: %w", err)
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			// Log error but don't return it to avoid masking original error
			fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		return "", fmt.Errorf("%w with status %d: %s", capi.ErrRootInfoRequestFailed, resp.StatusCode, string(body))
	}

	var rootInfo struct {
		Links struct {
			UAA struct {
				Href string `json:"href"`
			} `json:"uaa"`
			Login struct {
				Href string `json:"href"`
			} `json:"login"`
		} `json:"links"`
	}

	err = json.NewDecoder(resp.Body).Decode(&rootInfo)
	if err != nil {
		return "", fmt.Errorf("parsing root info: %w", err)
	}

	// Prefer UAA URL, fall back to login URL
	uaaURL := rootInfo.Links.UAA.Href
	if uaaURL == "" {
		uaaURL = rootInfo.Links.Login.Href
	}

	if uaaURL == "" {
		return "", capi.ErrNoUAAOrLoginURL
	}

	return uaaURL, nil
}

func discoverUAAEndpoint(ctx context.Context, apiEndpoint string, skipTLS bool) (string, error) {
	httpClient, err := createDiscoveryHTTPClient(skipTLS)
	if err != nil {
		return "", err
	}

	return fetchRootInfo(ctx, httpClient, apiEndpoint)
}

// NewWithEndpoint creates a new client with just an API endpoint (no auth).
func NewWithEndpoint(ctx context.Context, endpoint string) (capi.Client, error) {
	return New(ctx, &capi.Config{
		APIEndpoint: endpoint,
	})
}

// NewWithToken creates a new client with an API endpoint and access token.
func NewWithToken(ctx context.Context, endpoint, token string) (capi.Client, error) {
	return New(ctx, &capi.Config{
		APIEndpoint: endpoint,
		AccessToken: token,
	})
}

// NewWithClientCredentials creates a new client using OAuth2 client credentials.
func NewWithClientCredentials(ctx context.Context, endpoint, clientID, clientSecret string) (capi.Client, error) {
	return New(ctx, &capi.Config{
		APIEndpoint:  endpoint,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	})
}

// NewWithPassword creates a new client using username/password authentication.
func NewWithPassword(ctx context.Context, endpoint, username, password string) (capi.Client, error) {
	return New(ctx, &capi.Config{
		APIEndpoint: endpoint,
		Username:    username,
		Password:    password,
	})
}
