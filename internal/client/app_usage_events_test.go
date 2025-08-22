package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fivetwenty-io/capi-client/pkg/capi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppUsageEventsClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/app_usage_events/event-guid", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		previousState := "STOPPED"
		previousInstanceCount := 1
		previousMemoryInMB := 256
		buildpackName := "nodejs_buildpack"
		buildpackGUID := "buildpack-guid"
		taskName := "migrate"
		taskGUID := "task-guid"
		parentAppName := "parent-app"
		parentAppGUID := "parent-app-guid"

		event := capi.AppUsageEvent{
			Resource: capi.Resource{
				GUID:      "event-guid",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			State:                         "STARTED",
			PreviousState:                 &previousState,
			InstanceCount:                 2,
			PreviousInstanceCount:         &previousInstanceCount,
			MemoryInMBPerInstance:         512,
			PreviousMemoryInMBPerInstance: &previousMemoryInMB,
			AppName:                       "test-app",
			AppGUID:                       "app-guid",
			SpaceName:                     "test-space",
			SpaceGUID:                     "space-guid",
			OrganizationName:              "test-org",
			OrganizationGUID:              "org-guid",
			BuildpackName:                 &buildpackName,
			BuildpackGUID:                 &buildpackGUID,
			ProcessType:                   "web",
			TaskName:                      &taskName,
			TaskGUID:                      &taskGUID,
			ParentAppName:                 &parentAppName,
			ParentAppGUID:                 &parentAppGUID,
			Package: capi.AppUsageEventPackage{
				State: "READY",
			},
		}

		json.NewEncoder(w).Encode(event)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	event, err := client.AppUsageEvents().Get(context.Background(), "event-guid")
	require.NoError(t, err)
	assert.Equal(t, "event-guid", event.GUID)
	assert.Equal(t, "STARTED", event.State)
	assert.Equal(t, "STOPPED", *event.PreviousState)
	assert.Equal(t, 2, event.InstanceCount)
	assert.Equal(t, 1, *event.PreviousInstanceCount)
	assert.Equal(t, 512, event.MemoryInMBPerInstance)
	assert.Equal(t, 256, *event.PreviousMemoryInMBPerInstance)
	assert.Equal(t, "test-app", event.AppName)
	assert.Equal(t, "nodejs_buildpack", *event.BuildpackName)
}

func TestAppUsageEventsClient_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/app_usage_events", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "10", r.URL.Query().Get("per_page"))

		response := capi.ListResponse[capi.AppUsageEvent]{
			Pagination: capi.Pagination{
				TotalResults: 2,
				TotalPages:   1,
			},
			Resources: []capi.AppUsageEvent{
				{
					Resource:              capi.Resource{GUID: "event-1"},
					State:                 "STARTED",
					InstanceCount:         1,
					MemoryInMBPerInstance: 256,
					AppName:               "app-1",
					AppGUID:               "app-guid-1",
					SpaceName:             "space-1",
					SpaceGUID:             "space-guid-1",
					OrganizationName:      "org-1",
					OrganizationGUID:      "org-guid-1",
					ProcessType:           "web",
				},
				{
					Resource:              capi.Resource{GUID: "event-2"},
					State:                 "STOPPED",
					InstanceCount:         0,
					MemoryInMBPerInstance: 512,
					AppName:               "app-2",
					AppGUID:               "app-guid-2",
					SpaceName:             "space-2",
					SpaceGUID:             "space-guid-2",
					OrganizationName:      "org-2",
					OrganizationGUID:      "org-guid-2",
					ProcessType:           "worker",
				},
			},
		}

		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	params := capi.NewQueryParams().WithPage(1).WithPerPage(10)
	result, err := client.AppUsageEvents().List(context.Background(), params)

	require.NoError(t, err)
	assert.Len(t, result.Resources, 2)
	assert.Equal(t, "event-1", result.Resources[0].GUID)
	assert.Equal(t, "event-2", result.Resources[1].GUID)
	assert.Equal(t, "STARTED", result.Resources[0].State)
	assert.Equal(t, "STOPPED", result.Resources[1].State)
}

func TestAppUsageEventsClient_PurgeAndReseed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/app_usage_events/actions/destructively_purge_all_and_reseed", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("{}"))
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	err = client.AppUsageEvents().PurgeAndReseed(context.Background())
	require.NoError(t, err)
}
