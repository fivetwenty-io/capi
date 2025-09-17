package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/fivetwenty-io/capi/v3/internal/client"
	"github.com/fivetwenty-io/capi/v3/pkg/capi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestAppUsageEvent() capi.AppUsageEvent {
	previousState := "STOPPED"
	previousInstanceCount := 1
	previousMemoryInMB := 256
	buildpackName := "nodejs_buildpack"
	buildpackGUID := "buildpack-guid"
	taskName := "migrate"
	taskGUID := "task-guid"
	parentAppName := "parent-app"
	parentAppGUID := "parent-app-guid"

	return capi.AppUsageEvent{
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
}

func TestAppUsageEventsClient_Get(t *testing.T) {
	t.Parallel()

	event := createTestAppUsageEvent()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/app_usage_events/event-guid", request.URL.Path)
		assert.Equal(t, "GET", request.Method)

		_ = json.NewEncoder(writer).Encode(event)
	}))
	defer server.Close()

	client, err := New(context.Background(), &capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	result, err := client.AppUsageEvents().Get(context.Background(), "event-guid")
	require.NoError(t, err)
	assert.Equal(t, "event-guid", result.GUID)
	assert.Equal(t, "STARTED", result.State)
	assert.Equal(t, "STOPPED", *result.PreviousState)
	assert.Equal(t, 2, result.InstanceCount)
	assert.Equal(t, 1, *result.PreviousInstanceCount)
	assert.Equal(t, 512, result.MemoryInMBPerInstance)
	assert.Equal(t, 256, *result.PreviousMemoryInMBPerInstance)
	assert.Equal(t, "test-app", result.AppName)
	assert.Equal(t, "nodejs_buildpack", *result.BuildpackName)
}

func TestAppUsageEventsClient_List(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/app_usage_events", request.URL.Path)
		assert.Equal(t, "GET", request.Method)
		assert.Equal(t, "1", request.URL.Query().Get("page"))
		assert.Equal(t, "10", request.URL.Query().Get("per_page"))

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

		_ = json.NewEncoder(writer).Encode(response)
	}))
	defer server.Close()

	client, err := New(context.Background(), &capi.Config{APIEndpoint: server.URL})
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
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/app_usage_events/actions/destructively_purge_all_and_reseed", request.URL.Path)
		assert.Equal(t, "POST", request.Method)

		writer.WriteHeader(http.StatusAccepted)
		_, _ = writer.Write([]byte("{}"))
	}))
	defer server.Close()

	client, err := New(context.Background(), &capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	err = client.AppUsageEvents().PurgeAndReseed(context.Background())
	require.NoError(t, err)
}
