package client

import (
	"github.com/fivetwenty-io/capi/v3/internal/http"
	"github.com/fivetwenty-io/capi/v3/pkg/capi"
)

// ServiceUsageEventsClient implements capi.ServiceUsageEventsClient.
type ServiceUsageEventsClient struct {
	*UsageEventsClient[capi.ServiceUsageEvent]
}

// NewServiceUsageEventsClient creates a new service usage events client.
func NewServiceUsageEventsClient(httpClient *http.Client) *ServiceUsageEventsClient {
	return &ServiceUsageEventsClient{
		UsageEventsClient: NewUsageEventsClient[capi.ServiceUsageEvent](
			httpClient,
			"/v3/service_usage_events",
			"service",
		),
	}
}
