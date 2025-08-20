package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	internalhttp "github.com/fivetwenty-io/capi-client/internal/http"
	"github.com/fivetwenty-io/capi-client/pkg/capi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceInstancesClient_Create_Managed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_instances", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var request capi.ServiceInstanceCreateRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		require.NoError(t, err)

		assert.Equal(t, "managed", request.Type)
		assert.Equal(t, "my-instance", request.Name)
		assert.Equal(t, "space-guid", request.Relationships.Space.Data.GUID)
		assert.Equal(t, "plan-guid", request.Relationships.ServicePlan.Data.GUID)

		// Managed instances return a job
		job := capi.Job{
			Resource: capi.Resource{
				GUID: "job-guid",
			},
			Operation: "service_instance.create",
			State:     "PROCESSING",
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location", "/v3/jobs/job-guid")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(job)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	serviceInstances := NewServiceInstancesClient(client.httpClient)

	request := &capi.ServiceInstanceCreateRequest{
		Type: "managed",
		Name: "my-instance",
		Parameters: map[string]interface{}{
			"foo": "bar",
		},
		Tags: []string{"tag1", "tag2"},
		Relationships: capi.ServiceInstanceRelationships{
			Space: capi.Relationship{
				Data: &capi.RelationshipData{
					GUID: "space-guid",
				},
			},
			ServicePlan: &capi.Relationship{
				Data: &capi.RelationshipData{
					GUID: "plan-guid",
				},
			},
		},
	}

	result, err := serviceInstances.Create(context.Background(), request)
	require.NoError(t, err)

	job, ok := result.(*capi.Job)
	require.True(t, ok, "Expected *capi.Job for managed instance")
	assert.Equal(t, "job-guid", job.GUID)
	assert.Equal(t, "service_instance.create", job.Operation)
}

func TestServiceInstancesClient_Create_UserProvided(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_instances", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var request capi.ServiceInstanceCreateRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		require.NoError(t, err)

		assert.Equal(t, "user-provided", request.Type)
		assert.Equal(t, "my-ups", request.Name)
		assert.Equal(t, "space-guid", request.Relationships.Space.Data.GUID)

		// User-provided instances return the instance directly
		now := time.Now()
		instance := capi.ServiceInstance{
			Resource: capi.Resource{
				GUID:      "instance-guid",
				CreatedAt: now,
				UpdatedAt: now,
			},
			Name: "my-ups",
			Type: "user-provided",
			Tags: []string{"tag1"},
			LastOperation: &capi.ServiceInstanceLastOperation{
				Type:        "create",
				State:       "succeeded",
				Description: "Operation succeeded",
				CreatedAt:   &now,
				UpdatedAt:   &now,
			},
			SyslogDrainURL:  request.SyslogDrainURL,
			RouteServiceURL: request.RouteServiceURL,
			Relationships: capi.ServiceInstanceRelationships{
				Space: capi.Relationship{
					Data: &capi.RelationshipData{
						GUID: "space-guid",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(instance)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	serviceInstances := NewServiceInstancesClient(client.httpClient)

	syslogURL := "https://syslog.example.com"
	routeURL := "https://route.example.com"
	request := &capi.ServiceInstanceCreateRequest{
		Type: "user-provided",
		Name: "my-ups",
		Credentials: map[string]interface{}{
			"username": "admin",
			"password": "secret",
		},
		Tags:            []string{"tag1"},
		SyslogDrainURL:  &syslogURL,
		RouteServiceURL: &routeURL,
		Relationships: capi.ServiceInstanceRelationships{
			Space: capi.Relationship{
				Data: &capi.RelationshipData{
					GUID: "space-guid",
				},
			},
		},
	}

	result, err := serviceInstances.Create(context.Background(), request)
	require.NoError(t, err)

	instance, ok := result.(*capi.ServiceInstance)
	require.True(t, ok, "Expected *capi.ServiceInstance for user-provided instance")
	assert.Equal(t, "instance-guid", instance.GUID)
	assert.Equal(t, "my-ups", instance.Name)
	assert.Equal(t, "user-provided", instance.Type)
}

func TestServiceInstancesClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_instances/instance-guid", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		now := time.Now()
		instance := capi.ServiceInstance{
			Resource: capi.Resource{
				GUID:      "instance-guid",
				CreatedAt: now,
				UpdatedAt: now,
			},
			Name: "my-instance",
			Type: "managed",
			Tags: []string{"database", "postgresql"},
			MaintenanceInfo: &capi.ServiceInstanceMaintenance{
				Version: "1.0.0",
			},
			UpgradeAvailable: false,
			DashboardURL:     stringPtr("https://dashboard.example.com"),
			LastOperation: &capi.ServiceInstanceLastOperation{
				Type:        "create",
				State:       "succeeded",
				Description: "Instance created",
				CreatedAt:   &now,
				UpdatedAt:   &now,
			},
			Relationships: capi.ServiceInstanceRelationships{
				Space: capi.Relationship{
					Data: &capi.RelationshipData{
						GUID: "space-guid",
					},
				},
				ServicePlan: &capi.Relationship{
					Data: &capi.RelationshipData{
						GUID: "plan-guid",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(instance)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	serviceInstances := NewServiceInstancesClient(client.httpClient)

	instance, err := serviceInstances.Get(context.Background(), "instance-guid")
	require.NoError(t, err)
	assert.NotNil(t, instance)
	assert.Equal(t, "instance-guid", instance.GUID)
	assert.Equal(t, "my-instance", instance.Name)
	assert.Equal(t, "managed", instance.Type)
	assert.Contains(t, instance.Tags, "database")
	assert.Contains(t, instance.Tags, "postgresql")
}

func TestServiceInstancesClient_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_instances", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "space-guid", r.URL.Query().Get("space_guids"))
		assert.Equal(t, "my-instance", r.URL.Query().Get("names"))

		now := time.Now()
		response := capi.ListResponse[capi.ServiceInstance]{
			Pagination: capi.Pagination{
				TotalResults: 2,
				TotalPages:   1,
				First:        capi.Link{Href: "/v3/service_instances?page=1"},
				Last:         capi.Link{Href: "/v3/service_instances?page=1"},
			},
			Resources: []capi.ServiceInstance{
				{
					Resource: capi.Resource{
						GUID:      "instance-guid-1",
						CreatedAt: now,
						UpdatedAt: now,
					},
					Name: "my-instance-1",
					Type: "managed",
				},
				{
					Resource: capi.Resource{
						GUID:      "instance-guid-2",
						CreatedAt: now,
						UpdatedAt: now,
					},
					Name: "my-instance-2",
					Type: "user-provided",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	serviceInstances := NewServiceInstancesClient(client.httpClient)

	params := &capi.QueryParams{
		Filters: map[string][]string{
			"space_guids": {"space-guid"},
			"names":       {"my-instance"},
		},
	}

	list, err := serviceInstances.List(context.Background(), params)
	require.NoError(t, err)
	assert.NotNil(t, list)
	assert.Equal(t, 2, list.Pagination.TotalResults)
	assert.Len(t, list.Resources, 2)
	assert.Equal(t, "my-instance-1", list.Resources[0].Name)
	assert.Equal(t, "managed", list.Resources[0].Type)
	assert.Equal(t, "my-instance-2", list.Resources[1].Name)
	assert.Equal(t, "user-provided", list.Resources[1].Type)
}

func TestServiceInstancesClient_Update_Managed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_instances/instance-guid", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var request capi.ServiceInstanceUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		require.NoError(t, err)

		// Managed instances return a job for updates
		job := capi.Job{
			Resource: capi.Resource{
				GUID: "job-guid",
			},
			Operation: "service_instance.update",
			State:     "PROCESSING",
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location", "/v3/jobs/job-guid")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(job)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	serviceInstances := NewServiceInstancesClient(client.httpClient)

	newName := "updated-instance"
	request := &capi.ServiceInstanceUpdateRequest{
		Name: &newName,
		Parameters: map[string]interface{}{
			"max_connections": 100,
		},
		Tags: []string{"updated", "tags"},
	}

	result, err := serviceInstances.Update(context.Background(), "instance-guid", request)
	require.NoError(t, err)

	job, ok := result.(*capi.Job)
	require.True(t, ok, "Expected *capi.Job for managed instance update")
	assert.Equal(t, "job-guid", job.GUID)
	assert.Equal(t, "service_instance.update", job.Operation)
}

func TestServiceInstancesClient_Update_UserProvided(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_instances/instance-guid", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		// Check for user-provided instance by examining the request
		var request map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&request)
		require.NoError(t, err)

		// User-provided instances return the updated instance directly
		now := time.Now()
		instance := capi.ServiceInstance{
			Resource: capi.Resource{
				GUID:      "instance-guid",
				CreatedAt: now,
				UpdatedAt: now,
			},
			Name: "updated-ups",
			Type: "user-provided",
			Tags: []string{"updated"},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(instance)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	serviceInstances := NewServiceInstancesClient(client.httpClient)

	newName := "updated-ups"
	request := &capi.ServiceInstanceUpdateRequest{
		Name: &newName,
		Credentials: map[string]interface{}{
			"username": "newuser",
		},
		Tags: []string{"updated"},
	}

	result, err := serviceInstances.Update(context.Background(), "instance-guid", request)
	require.NoError(t, err)

	instance, ok := result.(*capi.ServiceInstance)
	require.True(t, ok, "Expected *capi.ServiceInstance for user-provided instance update")
	assert.Equal(t, "instance-guid", instance.GUID)
	assert.Equal(t, "updated-ups", instance.Name)
}

func TestServiceInstancesClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_instances/instance-guid", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "true", r.URL.Query().Get("purge"))

		job := capi.Job{
			Resource: capi.Resource{
				GUID: "job-guid",
			},
			Operation: "service_instance.delete",
			State:     "PROCESSING",
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location", "/v3/jobs/job-guid")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(job)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	serviceInstances := NewServiceInstancesClient(client.httpClient)

	job, err := serviceInstances.Delete(context.Background(), "instance-guid")
	require.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, "job-guid", job.GUID)
	assert.Equal(t, "service_instance.delete", job.Operation)
}

func TestServiceInstancesClient_GetParameters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_instances/instance-guid/parameters", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		params := capi.ServiceInstanceParameters{
			Parameters: map[string]interface{}{
				"max_connections": 100,
				"enable_ssl":      true,
				"database_name":   "mydb",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(params)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	serviceInstances := NewServiceInstancesClient(client.httpClient)

	params, err := serviceInstances.GetParameters(context.Background(), "instance-guid")
	require.NoError(t, err)
	assert.NotNil(t, params)
	assert.Equal(t, float64(100), params.Parameters["max_connections"])
	assert.Equal(t, true, params.Parameters["enable_ssl"])
	assert.Equal(t, "mydb", params.Parameters["database_name"])
}

func TestServiceInstancesClient_ListSharedSpaces(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_instances/instance-guid/relationships/shared_spaces", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		relationships := capi.ServiceInstanceSharedSpacesRelationships{
			Data: []capi.Relationship{
				{Data: &capi.RelationshipData{GUID: "space-guid-1"}},
				{Data: &capi.RelationshipData{GUID: "space-guid-2"}},
			},
			Links: capi.Links{
				"self": capi.Link{
					Href: "/v3/service_instances/instance-guid/relationships/shared_spaces",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(relationships)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	serviceInstances := NewServiceInstancesClient(client.httpClient)

	sharedSpaces, err := serviceInstances.ListSharedSpaces(context.Background(), "instance-guid")
	require.NoError(t, err)
	assert.NotNil(t, sharedSpaces)
	assert.Len(t, sharedSpaces.Data, 2)
	assert.Equal(t, "space-guid-1", sharedSpaces.Data[0].Data.GUID)
	assert.Equal(t, "space-guid-2", sharedSpaces.Data[1].Data.GUID)
}

func TestServiceInstancesClient_ShareWithSpaces(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_instances/instance-guid/relationships/shared_spaces", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var request capi.ServiceInstanceShareRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		require.NoError(t, err)
		assert.Len(t, request.Data, 2)

		relationships := capi.ServiceInstanceSharedSpacesRelationships{
			Data: []capi.Relationship{
				{Data: &capi.RelationshipData{GUID: "space-guid-1"}},
				{Data: &capi.RelationshipData{GUID: "space-guid-2"}},
				{Data: &capi.RelationshipData{GUID: "space-guid-3"}},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(relationships)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	serviceInstances := NewServiceInstancesClient(client.httpClient)

	request := &capi.ServiceInstanceShareRequest{
		Data: []capi.Relationship{
			{Data: &capi.RelationshipData{GUID: "space-guid-2"}},
			{Data: &capi.RelationshipData{GUID: "space-guid-3"}},
		},
	}

	sharedSpaces, err := serviceInstances.ShareWithSpaces(context.Background(), "instance-guid", request)
	require.NoError(t, err)
	assert.NotNil(t, sharedSpaces)
	assert.Len(t, sharedSpaces.Data, 3)
}

func TestServiceInstancesClient_UnshareFromSpace(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_instances/instance-guid/relationships/shared_spaces/space-guid", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	serviceInstances := NewServiceInstancesClient(client.httpClient)

	err := serviceInstances.UnshareFromSpace(context.Background(), "instance-guid", "space-guid")
	require.NoError(t, err)
}

func TestServiceInstancesClient_GetNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_instances/instance-guid", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	serviceInstances := NewServiceInstancesClient(client.httpClient)

	instance, err := serviceInstances.Get(context.Background(), "instance-guid")
	assert.Error(t, err)
	assert.Nil(t, instance)
}

func TestServiceInstancesClient_DeleteWithBindings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_instances/instance-guid", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		w.WriteHeader(http.StatusUnprocessableEntity)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	serviceInstances := NewServiceInstancesClient(client.httpClient)

	job, err := serviceInstances.Delete(context.Background(), "instance-guid")
	assert.Error(t, err)
	assert.Nil(t, job)
}
