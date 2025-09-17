package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/fivetwenty-io/capi/v3/internal/constants"
	http_internal "github.com/fivetwenty-io/capi/v3/internal/http"
	"github.com/fivetwenty-io/capi/v3/pkg/capi"
)

// DropletsClient implements the capi.DropletsClient interface.
type DropletsClient struct {
	httpClient *http_internal.Client
}

// NewDropletsClient creates a new DropletsClient.
func NewDropletsClient(httpClient *http_internal.Client) *DropletsClient {
	return &DropletsClient{
		httpClient: httpClient,
	}
}

// Create creates a new droplet.
func (c *DropletsClient) Create(ctx context.Context, request *capi.DropletCreateRequest) (*capi.Droplet, error) {
	path := constants.APIPathDroplets

	resp, err := c.httpClient.Post(ctx, path, request)
	if err != nil {
		return nil, fmt.Errorf("creating droplet: %w", err)
	}

	var droplet capi.Droplet

	err = json.Unmarshal(resp.Body, &droplet)
	if err != nil {
		return nil, fmt.Errorf("parsing droplet response: %w", err)
	}

	return &droplet, nil
}

// Get retrieves a specific droplet.
func (c *DropletsClient) Get(ctx context.Context, guid string) (*capi.Droplet, error) {
	path := "/v3/droplets/" + guid

	resp, err := c.httpClient.Get(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("getting droplet: %w", err)
	}

	var droplet capi.Droplet

	err = json.Unmarshal(resp.Body, &droplet)
	if err != nil {
		return nil, fmt.Errorf("parsing droplet response: %w", err)
	}

	return &droplet, nil
}

// List lists all droplets.
func (c *DropletsClient) List(ctx context.Context, params *capi.QueryParams) (*capi.ListResponse[capi.Droplet], error) {
	path := constants.APIPathDroplets

	var queryParams url.Values
	if params != nil {
		queryParams = params.ToValues()
	}

	resp, err := c.httpClient.Get(ctx, path, queryParams)
	if err != nil {
		return nil, fmt.Errorf("listing droplets: %w", err)
	}

	var result capi.ListResponse[capi.Droplet]

	err = json.Unmarshal(resp.Body, &result)
	if err != nil {
		return nil, fmt.Errorf("parsing droplets list response: %w", err)
	}

	return &result, nil
}

// ListForApp lists droplets for a specific app.
func (c *DropletsClient) ListForApp(ctx context.Context, appGUID string, params *capi.QueryParams) (*capi.ListResponse[capi.Droplet], error) {
	path := fmt.Sprintf("/v3/apps/%s/droplets", appGUID)

	var queryParams url.Values
	if params != nil {
		queryParams = params.ToValues()
	}

	resp, err := c.httpClient.Get(ctx, path, queryParams)
	if err != nil {
		return nil, fmt.Errorf("listing droplets for app: %w", err)
	}

	var result capi.ListResponse[capi.Droplet]

	err = json.Unmarshal(resp.Body, &result)
	if err != nil {
		return nil, fmt.Errorf("parsing droplets list response: %w", err)
	}

	return &result, nil
}

// ListForPackage lists droplets for a specific package.
func (c *DropletsClient) ListForPackage(ctx context.Context, packageGUID string, params *capi.QueryParams) (*capi.ListResponse[capi.Droplet], error) {
	path := fmt.Sprintf("/v3/packages/%s/droplets", packageGUID)

	var queryParams url.Values
	if params != nil {
		queryParams = params.ToValues()
	}

	resp, err := c.httpClient.Get(ctx, path, queryParams)
	if err != nil {
		return nil, fmt.Errorf("listing droplets for package: %w", err)
	}

	var result capi.ListResponse[capi.Droplet]

	err = json.Unmarshal(resp.Body, &result)
	if err != nil {
		return nil, fmt.Errorf("parsing droplets list response: %w", err)
	}

	return &result, nil
}

// Update updates a droplet's metadata.
func (c *DropletsClient) Update(ctx context.Context, guid string, request *capi.DropletUpdateRequest) (*capi.Droplet, error) {
	path := "/v3/droplets/" + guid

	resp, err := c.httpClient.Patch(ctx, path, request)
	if err != nil {
		return nil, fmt.Errorf("updating droplet: %w", err)
	}

	var droplet capi.Droplet

	err = json.Unmarshal(resp.Body, &droplet)
	if err != nil {
		return nil, fmt.Errorf("parsing droplet response: %w", err)
	}

	return &droplet, nil
}

// Delete deletes a droplet.
func (c *DropletsClient) Delete(ctx context.Context, guid string) error {
	path := "/v3/droplets/" + guid

	_, err := c.httpClient.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("deleting droplet: %w", err)
	}

	return nil
}

// Copy copies a droplet to another app.
func (c *DropletsClient) Copy(ctx context.Context, sourceGUID string, request *capi.DropletCopyRequest) (*capi.Droplet, error) {
	path := constants.APIPathDroplets

	// Build query parameters
	queryParams := url.Values{}
	queryParams.Set("source_guid", sourceGUID)

	// Use Do method directly to pass query parameters properly
	resp, err := c.httpClient.Do(ctx, &http_internal.Request{
		Method: http.MethodPost,
		Path:   path,
		Query:  queryParams,
		Body:   request,
	})
	if err != nil {
		return nil, fmt.Errorf("copying droplet: %w", err)
	}

	var droplet capi.Droplet

	err = json.Unmarshal(resp.Body, &droplet)
	if err != nil {
		return nil, fmt.Errorf("parsing droplet response: %w", err)
	}

	return &droplet, nil
}

// Download downloads a droplet.
func (c *DropletsClient) Download(ctx context.Context, guid string) ([]byte, error) {
	path := fmt.Sprintf("/v3/droplets/%s/download", guid)

	resp, err := c.httpClient.Get(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("downloading droplet: %w", err)
	}

	// The response body contains the actual file content
	content, err := io.ReadAll(bytes.NewReader(resp.Body))
	if err != nil {
		return nil, fmt.Errorf("reading droplet content: %w", err)
	}

	return content, nil
}

// Upload uploads bits to a droplet.
func (c *DropletsClient) Upload(ctx context.Context, guid string, bits []byte) (*capi.Droplet, error) {
	path := fmt.Sprintf("/v3/droplets/%s/upload", guid)

	respBody, err := uploadMultipartFile(ctx, c.httpClient, path, "droplet.tgz", bits, "droplet")
	if err != nil {
		return nil, err
	}

	var droplet capi.Droplet

	err = json.Unmarshal(respBody, &droplet)
	if err != nil {
		return nil, fmt.Errorf("parsing droplet response: %w", err)
	}

	return &droplet, nil
}
