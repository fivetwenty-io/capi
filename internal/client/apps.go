package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/fivetwenty-io/capi-client-go/internal/http"
	"github.com/fivetwenty-io/capi-client-go/pkg/capi"
)

// AppsClient implements capi.AppsClient
type AppsClient struct {
	httpClient *http.Client
}

// NewAppsClient creates a new apps client
func NewAppsClient(httpClient *http.Client) *AppsClient {
	return &AppsClient{
		httpClient: httpClient,
	}
}

// Create implements capi.AppsClient.Create
func (c *AppsClient) Create(ctx context.Context, request *capi.AppCreateRequest) (*capi.App, error) {
	resp, err := c.httpClient.Post(ctx, "/v3/apps", request)
	if err != nil {
		return nil, fmt.Errorf("creating app: %w", err)
	}

	var app capi.App
	if err := json.Unmarshal(resp.Body, &app); err != nil {
		return nil, fmt.Errorf("parsing app response: %w", err)
	}

	return &app, nil
}

// Get implements capi.AppsClient.Get
func (c *AppsClient) Get(ctx context.Context, guid string) (*capi.App, error) {
	path := fmt.Sprintf("/v3/apps/%s", guid)
	resp, err := c.httpClient.Get(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("getting app: %w", err)
	}

	var app capi.App
	if err := json.Unmarshal(resp.Body, &app); err != nil {
		return nil, fmt.Errorf("parsing app response: %w", err)
	}

	return &app, nil
}

// List implements capi.AppsClient.List
func (c *AppsClient) List(ctx context.Context, params *capi.QueryParams) (*capi.ListResponse[capi.App], error) {
	var query url.Values
	if params != nil {
		query = params.ToValues()
	}

	resp, err := c.httpClient.Get(ctx, "/v3/apps", query)
	if err != nil {
		return nil, fmt.Errorf("listing apps: %w", err)
	}

	var result capi.ListResponse[capi.App]
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("parsing apps list response: %w", err)
	}

	return &result, nil
}

// Update implements capi.AppsClient.Update
func (c *AppsClient) Update(ctx context.Context, guid string, request *capi.AppUpdateRequest) (*capi.App, error) {
	path := fmt.Sprintf("/v3/apps/%s", guid)
	resp, err := c.httpClient.Patch(ctx, path, request)
	if err != nil {
		return nil, fmt.Errorf("updating app: %w", err)
	}

	var app capi.App
	if err := json.Unmarshal(resp.Body, &app); err != nil {
		return nil, fmt.Errorf("parsing app response: %w", err)
	}

	return &app, nil
}

// Delete implements capi.AppsClient.Delete
func (c *AppsClient) Delete(ctx context.Context, guid string) error {
	path := fmt.Sprintf("/v3/apps/%s", guid)
	_, err := c.httpClient.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("deleting app: %w", err)
	}

	return nil
}

// Start implements capi.AppsClient.Start
func (c *AppsClient) Start(ctx context.Context, guid string) (*capi.App, error) {
	path := fmt.Sprintf("/v3/apps/%s/actions/start", guid)
	resp, err := c.httpClient.Post(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("starting app: %w", err)
	}

	var app capi.App
	if err := json.Unmarshal(resp.Body, &app); err != nil {
		return nil, fmt.Errorf("parsing app response: %w", err)
	}

	return &app, nil
}

// Stop implements capi.AppsClient.Stop
func (c *AppsClient) Stop(ctx context.Context, guid string) (*capi.App, error) {
	path := fmt.Sprintf("/v3/apps/%s/actions/stop", guid)
	resp, err := c.httpClient.Post(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("stopping app: %w", err)
	}

	var app capi.App
	if err := json.Unmarshal(resp.Body, &app); err != nil {
		return nil, fmt.Errorf("parsing app response: %w", err)
	}

	return &app, nil
}

// Restart implements capi.AppsClient.Restart
func (c *AppsClient) Restart(ctx context.Context, guid string) (*capi.App, error) {
	path := fmt.Sprintf("/v3/apps/%s/actions/restart", guid)
	resp, err := c.httpClient.Post(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("restarting app: %w", err)
	}

	var app capi.App
	if err := json.Unmarshal(resp.Body, &app); err != nil {
		return nil, fmt.Errorf("parsing app response: %w", err)
	}

	return &app, nil
}

// GetEnv implements capi.AppsClient.GetEnv
func (c *AppsClient) GetEnv(ctx context.Context, guid string) (*capi.AppEnvironment, error) {
	path := fmt.Sprintf("/v3/apps/%s/env", guid)
	resp, err := c.httpClient.Get(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("getting app environment: %w", err)
	}

	var env capi.AppEnvironment
	if err := json.Unmarshal(resp.Body, &env); err != nil {
		return nil, fmt.Errorf("parsing app environment response: %w", err)
	}

	return &env, nil
}

// GetEnvVars implements capi.AppsClient.GetEnvVars
func (c *AppsClient) GetEnvVars(ctx context.Context, guid string) (map[string]interface{}, error) {
	path := fmt.Sprintf("/v3/apps/%s/environment_variables", guid)
	resp, err := c.httpClient.Get(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("getting app environment variables: %w", err)
	}

	// The response has a 'var' field that contains the environment variables
	var result struct {
		Var map[string]interface{} `json:"var"`
	}
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("parsing environment variables response: %w", err)
	}

	return result.Var, nil
}

// UpdateEnvVars implements capi.AppsClient.UpdateEnvVars
func (c *AppsClient) UpdateEnvVars(ctx context.Context, guid string, envVars map[string]interface{}) (map[string]interface{}, error) {
	path := fmt.Sprintf("/v3/apps/%s/environment_variables", guid)

	// Wrap the variables in a 'var' field as required by the API
	body := map[string]interface{}{
		"var": envVars,
	}

	resp, err := c.httpClient.Patch(ctx, path, body)
	if err != nil {
		return nil, fmt.Errorf("updating app environment variables: %w", err)
	}

	var result struct {
		Var map[string]interface{} `json:"var"`
	}
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("parsing environment variables response: %w", err)
	}

	return result.Var, nil
}

// GetCurrentDroplet implements capi.AppsClient.GetCurrentDroplet
func (c *AppsClient) GetCurrentDroplet(ctx context.Context, guid string) (*capi.Droplet, error) {
	path := fmt.Sprintf("/v3/apps/%s/droplets/current", guid)
	resp, err := c.httpClient.Get(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("getting current droplet: %w", err)
	}

	var droplet capi.Droplet
	if err := json.Unmarshal(resp.Body, &droplet); err != nil {
		return nil, fmt.Errorf("parsing droplet response: %w", err)
	}

	return &droplet, nil
}

// SetCurrentDroplet implements capi.AppsClient.SetCurrentDroplet
func (c *AppsClient) SetCurrentDroplet(ctx context.Context, guid string, dropletGUID string) (*capi.Relationship, error) {
	path := fmt.Sprintf("/v3/apps/%s/relationships/current_droplet", guid)

	body := capi.Relationship{
		Data: &capi.RelationshipData{GUID: dropletGUID},
	}

	resp, err := c.httpClient.Patch(ctx, path, body)
	if err != nil {
		return nil, fmt.Errorf("setting current droplet: %w", err)
	}

	var relationship capi.Relationship
	if err := json.Unmarshal(resp.Body, &relationship); err != nil {
		return nil, fmt.Errorf("parsing relationship response: %w", err)
	}

	return &relationship, nil
}

// GetSSHEnabled implements capi.AppsClient.GetSSHEnabled
func (c *AppsClient) GetSSHEnabled(ctx context.Context, guid string) (*capi.AppSSHEnabled, error) {
	path := fmt.Sprintf("/v3/apps/%s/ssh_enabled", guid)
	resp, err := c.httpClient.Get(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("getting SSH enabled status: %w", err)
	}

	var sshEnabled capi.AppSSHEnabled
	if err := json.Unmarshal(resp.Body, &sshEnabled); err != nil {
		return nil, fmt.Errorf("parsing SSH enabled response: %w", err)
	}

	return &sshEnabled, nil
}

// GetPermissions implements capi.AppsClient.GetPermissions
func (c *AppsClient) GetPermissions(ctx context.Context, guid string) (*capi.AppPermissions, error) {
	path := fmt.Sprintf("/v3/apps/%s/permissions", guid)
	resp, err := c.httpClient.Get(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("getting app permissions: %w", err)
	}

	var permissions capi.AppPermissions
	if err := json.Unmarshal(resp.Body, &permissions); err != nil {
		return nil, fmt.Errorf("parsing permissions response: %w", err)
	}

	return &permissions, nil
}

// ClearBuildpackCache implements capi.AppsClient.ClearBuildpackCache
func (c *AppsClient) ClearBuildpackCache(ctx context.Context, guid string) error {
	path := fmt.Sprintf("/v3/apps/%s/actions/clear_buildpack_cache", guid)
	_, err := c.httpClient.Post(ctx, path, nil)
	if err != nil {
		return fmt.Errorf("clearing buildpack cache: %w", err)
	}

	return nil
}

// GetManifest implements capi.AppsClient.GetManifest
func (c *AppsClient) GetManifest(ctx context.Context, guid string) (string, error) {
	path := fmt.Sprintf("/v3/apps/%s/manifest", guid)
	resp, err := c.httpClient.Get(ctx, path, nil)
	if err != nil {
		return "", fmt.Errorf("getting app manifest: %w", err)
	}

	// The manifest is returned as YAML, so we return it as a string
	return string(resp.Body), nil
}
