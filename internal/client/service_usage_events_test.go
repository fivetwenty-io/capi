package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fivetwenty-io/capi/v3/pkg/capi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceUsageEventsClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_usage_events/event-guid", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		event := capi.ServiceUsageEvent{
			Resource: capi.Resource{
				GUID:      "event-guid",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			State:               "CREATED",
			ServiceInstanceName: "my-db",
			ServiceInstanceGUID: "service-instance-guid",
			ServiceInstanceType: "managed_service_instance",
			ServicePlanName:     "premium",
			ServicePlanGUID:     "plan-guid-2",
			ServiceOfferingName: "postgres",
			ServiceOfferingGUID: "offering-guid",
			ServiceBrokerName:   "postgres-broker",
			ServiceBrokerGUID:   "broker-guid",
			SpaceName:           "test-space",
			SpaceGUID:           "space-guid",
			OrganizationName:    "test-org",
			OrganizationGUID:    "org-guid",
		}

		json.NewEncoder(w).Encode(event)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	event, err := client.ServiceUsageEvents().Get(context.Background(), "event-guid")
	require.NoError(t, err)
	assert.Equal(t, "event-guid", event.GUID)
	assert.Equal(t, "CREATED", event.State)
	assert.Equal(t, "my-db", event.ServiceInstanceName)
	assert.Equal(t, "premium", event.ServicePlanName)
	assert.Equal(t, "postgres", event.ServiceOfferingName)
	assert.Equal(t, "postgres-broker", event.ServiceBrokerName)
}

func TestServiceUsageEventsClient_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_usage_events", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "10", r.URL.Query().Get("per_page"))

		response := capi.ListResponse[capi.ServiceUsageEvent]{
			Pagination: capi.Pagination{
				TotalResults: 2,
				TotalPages:   1,
			},
			Resources: []capi.ServiceUsageEvent{
				{
					Resource:            capi.Resource{GUID: "event-1"},
					State:               "CREATED",
					ServiceInstanceName: "service-1",
					ServiceInstanceGUID: "service-instance-guid-1",
					ServiceInstanceType: "managed_service_instance",
					ServicePlanName:     "basic",
					ServicePlanGUID:     "plan-guid-1",
					ServiceOfferingName: "redis",
					ServiceOfferingGUID: "offering-guid-1",
					ServiceBrokerName:   "redis-broker",
					ServiceBrokerGUID:   "broker-guid-1",
					SpaceName:           "space-1",
					SpaceGUID:           "space-guid-1",
					OrganizationName:    "org-1",
					OrganizationGUID:    "org-guid-1",
				},
				{
					Resource:            capi.Resource{GUID: "event-2"},
					State:               "DELETED",
					ServiceInstanceName: "service-2",
					ServiceInstanceGUID: "service-instance-guid-2",
					ServiceInstanceType: "user_provided_service_instance",
					ServicePlanName:     "",
					ServicePlanGUID:     "",
					ServiceOfferingName: "",
					ServiceOfferingGUID: "",
					ServiceBrokerName:   "",
					ServiceBrokerGUID:   "",
					SpaceName:           "space-2",
					SpaceGUID:           "space-guid-2",
					OrganizationName:    "org-2",
					OrganizationGUID:    "org-guid-2",
				},
			},
		}

		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	params := capi.NewQueryParams().WithPage(1).WithPerPage(10)
	result, err := client.ServiceUsageEvents().List(context.Background(), params)

	require.NoError(t, err)
	assert.Len(t, result.Resources, 2)
	assert.Equal(t, "event-1", result.Resources[0].GUID)
	assert.Equal(t, "event-2", result.Resources[1].GUID)
	assert.Equal(t, "CREATED", result.Resources[0].State)
	assert.Equal(t, "DELETED", result.Resources[1].State)
	assert.Equal(t, "managed_service_instance", result.Resources[0].ServiceInstanceType)
	assert.Equal(t, "user_provided_service_instance", result.Resources[1].ServiceInstanceType)
}

func TestServiceUsageEventsClient_PurgeAndReseed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_usage_events/actions/destructively_purge_all_and_reseed", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("{}"))
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	err = client.ServiceUsageEvents().PurgeAndReseed(context.Background())
	require.NoError(t, err)
}
