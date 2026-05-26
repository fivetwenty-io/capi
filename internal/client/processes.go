package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/fivetwenty-io/capi/v3/internal/http"
	"github.com/fivetwenty-io/capi/v3/pkg/capi"
)

// ProcessesClient implements the capi.ProcessesClient interface.
type ProcessesClient struct {
	httpClient *http.Client
}

// NewProcessesClient creates a new processes client.
func NewProcessesClient(httpClient *http.Client) *ProcessesClient {
	return &ProcessesClient{
		httpClient: httpClient,
	}
}

// Get retrieves a specific process by GUID.
func (c *ProcessesClient) Get(ctx context.Context, guid string) (*capi.Process, error) {
	path := "/v3/processes/" + guid

	resp, err := c.httpClient.Get(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("getting process: %w", err)
	}

	var process capi.Process

	err = json.Unmarshal(resp.Body, &process)
	if err != nil {
		return nil, fmt.Errorf("parsing process response: %w", err)
	}

	return &process, nil
}

// List retrieves all processes with optional filtering.
func (c *ProcessesClient) List(ctx context.Context, params *capi.QueryParams) (*capi.ListResponse[capi.Process], error) {
	var query url.Values
	if params != nil {
		query = params.ToValues()
	}

	resp, err := c.httpClient.Get(ctx, "/v3/processes", query)
	if err != nil {
		return nil, fmt.Errorf("listing processes: %w", err)
	}

	var result capi.ListResponse[capi.Process]

	err = json.Unmarshal(resp.Body, &result)
	if err != nil {
		return nil, fmt.Errorf("parsing processes list response: %w", err)
	}

	return &result, nil
}

// Update modifies a process.
func (c *ProcessesClient) Update(ctx context.Context, guid string, request *capi.ProcessUpdateRequest) (*capi.Process, error) {
	path := "/v3/processes/" + guid

	resp, err := c.httpClient.Patch(ctx, path, request)
	if err != nil {
		return nil, fmt.Errorf("updating process: %w", err)
	}

	var process capi.Process

	err = json.Unmarshal(resp.Body, &process)
	if err != nil {
		return nil, fmt.Errorf("parsing process response: %w", err)
	}

	return &process, nil
}

// Scale adjusts the instances, memory, disk, or log rate limit of a process.
//
// POST /v3/processes/{guid}/actions/scale is async on modern CF v3
// (202 + Location → /v3/jobs/{jobGuid}) and was synchronous on older
// versions (200 + updated Process body). We support both for forward-
// and backward-compatibility, matching AppsClient.postActionJob:
// Location header present → return *Job with extracted GUID; absent →
// return (nil, nil) signaling "action complete, no job to poll".
func (c *ProcessesClient) Scale(ctx context.Context, guid string, request *capi.ProcessScaleRequest) (*capi.Job, error) {
	path := fmt.Sprintf("/v3/processes/%s/actions/scale", guid)

	resp, err := c.httpClient.Post(ctx, path, request)
	if err != nil {
		return nil, fmt.Errorf("scaling process: %w", err)
	}

	location := resp.Headers.Get("Location")
	if location == "" {
		// Sync-complete (older CF) — no job to poll.
		return nil, nil
	}

	jobGUID := location
	if idx := strings.LastIndex(location, "/"); idx >= 0 {
		jobGUID = location[idx+1:]
	}
	if jobGUID == "" {
		return nil, fmt.Errorf("scaling process: malformed Location header %q", location)
	}

	return &capi.Job{Resource: capi.Resource{GUID: jobGUID}}, nil
}

// GetStats retrieves runtime statistics for all instances of a process.
func (c *ProcessesClient) GetStats(ctx context.Context, guid string) (*capi.ProcessStats, error) {
	path := fmt.Sprintf("/v3/processes/%s/stats", guid)

	resp, err := c.httpClient.Get(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("getting process stats: %w", err)
	}

	var stats capi.ProcessStats

	err = json.Unmarshal(resp.Body, &stats)
	if err != nil {
		return nil, fmt.Errorf("parsing process stats response: %w", err)
	}

	return &stats, nil
}

// ListInstances retrieves the instances for a process.
func (c *ProcessesClient) ListInstances(ctx context.Context, guid string) (*capi.ListResponse[capi.ProcessInstance], error) {
	path := fmt.Sprintf("/v3/processes/%s/process_instances", guid)

	resp, err := c.httpClient.Get(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("listing process instances: %w", err)
	}

	var result capi.ListResponse[capi.ProcessInstance]

	err = json.Unmarshal(resp.Body, &result)
	if err != nil {
		return nil, fmt.Errorf("parsing process instances response: %w", err)
	}

	return &result, nil
}

// TerminateInstance terminates a specific instance of a process.
func (c *ProcessesClient) TerminateInstance(ctx context.Context, guid string, index int) error {
	path := fmt.Sprintf("/v3/processes/%s/instances/%d", guid, index)

	_, err := c.httpClient.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("terminating process instance: %w", err)
	}

	return nil
}
