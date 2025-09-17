package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/fivetwenty-io/capi/v3/internal/http"
	"github.com/fivetwenty-io/capi/v3/pkg/capi"
)

// UsageEvent represents a generic usage event interface.
type UsageEvent interface {
	capi.AppUsageEvent | capi.ServiceUsageEvent
}

// UsageEventsClient provides a generic client for usage events.
type UsageEventsClient[T UsageEvent] struct {
	httpClient   *http.Client
	resourcePath string
	eventType    string
}

// NewUsageEventsClient creates a new generic usage events client.
func NewUsageEventsClient[T UsageEvent](httpClient *http.Client, resourcePath, eventType string) *UsageEventsClient[T] {
	return &UsageEventsClient[T]{
		httpClient:   httpClient,
		resourcePath: resourcePath,
		eventType:    eventType,
	}
}

// Get retrieves a specific usage event by GUID.
func (c *UsageEventsClient[T]) Get(ctx context.Context, guid string) (*T, error) {
	path := c.resourcePath + "/" + guid

	resp, err := c.httpClient.Get(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("getting %s usage event: %w", c.eventType, err)
	}

	var event T

	err = json.Unmarshal(resp.Body, &event)
	if err != nil {
		return nil, fmt.Errorf("parsing %s usage event response: %w", c.eventType, err)
	}

	return &event, nil
}

// List retrieves a list of usage events.
func (c *UsageEventsClient[T]) List(ctx context.Context, params *capi.QueryParams) (*capi.ListResponse[T], error) {
	var query url.Values
	if params != nil {
		query = params.ToValues()
	}

	resp, err := c.httpClient.Get(ctx, c.resourcePath, query)
	if err != nil {
		return nil, fmt.Errorf("listing %s usage events: %w", c.eventType, err)
	}

	var result capi.ListResponse[T]

	err = json.Unmarshal(resp.Body, &result)
	if err != nil {
		return nil, fmt.Errorf("parsing %s usage events list response: %w", c.eventType, err)
	}

	return &result, nil
}

// PurgeAndReseed purges and reseeds usage events.
func (c *UsageEventsClient[T]) PurgeAndReseed(ctx context.Context) error {
	path := c.resourcePath + "/actions/destructively_purge_all_and_reseed"

	_, err := c.httpClient.Post(ctx, path, nil)
	if err != nil {
		return fmt.Errorf("purging and reseeding %s usage events: %w", c.eventType, err)
	}

	return nil
}
