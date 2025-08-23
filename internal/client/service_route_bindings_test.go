package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	internalhttp "github.com/fivetwenty-io/capi/internal/http"
	"github.com/fivetwenty-io/capi/pkg/capi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceRouteBindingsClient_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_route_bindings", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var request capi.ServiceRouteBindingCreateRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		require.NoError(t, err)

		assert.Equal(t, "instance-guid", request.Relationships.ServiceInstance.Data.GUID)
		assert.Equal(t, "route-guid", request.Relationships.Route.Data.GUID)

		// Service route bindings may return a job for async operations
		job := capi.Job{
			Resource: capi.Resource{
				GUID: "job-guid",
			},
			Operation: "service_route_binding.create",
			State:     "PROCESSING",
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location", "/v3/jobs/job-guid")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(job)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	serviceRouteBindings := NewServiceRouteBindingsClient(client.httpClient)

	request := &capi.ServiceRouteBindingCreateRequest{
		Parameters: map[string]interface{}{
			"rate_limit": 100,
		},
		Relationships: capi.ServiceRouteBindingRelationships{
			ServiceInstance: capi.Relationship{
				Data: &capi.RelationshipData{
					GUID: "instance-guid",
				},
			},
			Route: capi.Relationship{
				Data: &capi.RelationshipData{
					GUID: "route-guid",
				},
			},
		},
	}

	result, err := serviceRouteBindings.Create(context.Background(), request)
	require.NoError(t, err)

	job, ok := result.(*capi.Job)
	require.True(t, ok, "Expected *capi.Job for service route binding")
	assert.Equal(t, "job-guid", job.GUID)
	assert.Equal(t, "service_route_binding.create", job.Operation)
}

func TestServiceRouteBindingsClient_CreateSync(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_route_bindings", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var request capi.ServiceRouteBindingCreateRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		require.NoError(t, err)

		// Some service route bindings may return the binding directly
		now := time.Now()
		binding := capi.ServiceRouteBinding{
			Resource: capi.Resource{
				GUID:      "binding-guid",
				CreatedAt: now,
				UpdatedAt: now,
			},
			RouteServiceURL: stringPtr("https://route-service.example.com"),
			LastOperation: &capi.ServiceRouteBindingLastOperation{
				Type:      "create",
				State:     "succeeded",
				CreatedAt: &now,
				UpdatedAt: &now,
			},
			Relationships: capi.ServiceRouteBindingRelationships{
				ServiceInstance: capi.Relationship{
					Data: &capi.RelationshipData{
						GUID: "instance-guid",
					},
				},
				Route: capi.Relationship{
					Data: &capi.RelationshipData{
						GUID: "route-guid",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(binding)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	serviceRouteBindings := NewServiceRouteBindingsClient(client.httpClient)

	request := &capi.ServiceRouteBindingCreateRequest{
		Relationships: capi.ServiceRouteBindingRelationships{
			ServiceInstance: capi.Relationship{
				Data: &capi.RelationshipData{
					GUID: "instance-guid",
				},
			},
			Route: capi.Relationship{
				Data: &capi.RelationshipData{
					GUID: "route-guid",
				},
			},
		},
	}

	result, err := serviceRouteBindings.Create(context.Background(), request)
	require.NoError(t, err)

	binding, ok := result.(*capi.ServiceRouteBinding)
	require.True(t, ok, "Expected *capi.ServiceRouteBinding for synchronous creation")
	assert.Equal(t, "binding-guid", binding.GUID)
	assert.NotNil(t, binding.RouteServiceURL)
	assert.Equal(t, "https://route-service.example.com", *binding.RouteServiceURL)
}

func TestServiceRouteBindingsClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_route_bindings/binding-guid", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		now := time.Now()
		binding := capi.ServiceRouteBinding{
			Resource: capi.Resource{
				GUID:      "binding-guid",
				CreatedAt: now,
				UpdatedAt: now,
			},
			RouteServiceURL: stringPtr("https://route-service.example.com"),
			LastOperation: &capi.ServiceRouteBindingLastOperation{
				Type:      "create",
				State:     "succeeded",
				CreatedAt: &now,
				UpdatedAt: &now,
			},
			Relationships: capi.ServiceRouteBindingRelationships{
				ServiceInstance: capi.Relationship{
					Data: &capi.RelationshipData{
						GUID: "instance-guid",
					},
				},
				Route: capi.Relationship{
					Data: &capi.RelationshipData{
						GUID: "route-guid",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(binding)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	serviceRouteBindings := NewServiceRouteBindingsClient(client.httpClient)

	binding, err := serviceRouteBindings.Get(context.Background(), "binding-guid")
	require.NoError(t, err)
	assert.NotNil(t, binding)
	assert.Equal(t, "binding-guid", binding.GUID)
	assert.NotNil(t, binding.RouteServiceURL)
	assert.Equal(t, "https://route-service.example.com", *binding.RouteServiceURL)
}

func TestServiceRouteBindingsClient_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_route_bindings", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "instance-guid", r.URL.Query().Get("service_instance_guids"))
		assert.Equal(t, "route-guid", r.URL.Query().Get("route_guids"))

		now := time.Now()
		response := capi.ListResponse[capi.ServiceRouteBinding]{
			Pagination: capi.Pagination{
				TotalResults: 2,
				TotalPages:   1,
				First:        capi.Link{Href: "/v3/service_route_bindings?page=1"},
				Last:         capi.Link{Href: "/v3/service_route_bindings?page=1"},
			},
			Resources: []capi.ServiceRouteBinding{
				{
					Resource: capi.Resource{
						GUID:      "binding-guid-1",
						CreatedAt: now,
						UpdatedAt: now,
					},
					RouteServiceURL: stringPtr("https://route-service1.example.com"),
				},
				{
					Resource: capi.Resource{
						GUID:      "binding-guid-2",
						CreatedAt: now,
						UpdatedAt: now,
					},
					RouteServiceURL: stringPtr("https://route-service2.example.com"),
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	serviceRouteBindings := NewServiceRouteBindingsClient(client.httpClient)

	params := &capi.QueryParams{
		Filters: map[string][]string{
			"service_instance_guids": {"instance-guid"},
			"route_guids":            {"route-guid"},
		},
	}

	list, err := serviceRouteBindings.List(context.Background(), params)
	require.NoError(t, err)
	assert.NotNil(t, list)
	assert.Equal(t, 2, list.Pagination.TotalResults)
	assert.Len(t, list.Resources, 2)
	assert.Equal(t, "binding-guid-1", list.Resources[0].GUID)
	assert.Equal(t, "binding-guid-2", list.Resources[1].GUID)
}

func TestServiceRouteBindingsClient_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_route_bindings/binding-guid", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var request capi.ServiceRouteBindingUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		require.NoError(t, err)

		now := time.Now()
		binding := capi.ServiceRouteBinding{
			Resource: capi.Resource{
				GUID:      "binding-guid",
				CreatedAt: now,
				UpdatedAt: now,
			},
			RouteServiceURL: stringPtr("https://route-service.example.com"),
			Metadata:        request.Metadata,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(binding)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	serviceRouteBindings := NewServiceRouteBindingsClient(client.httpClient)

	request := &capi.ServiceRouteBindingUpdateRequest{
		Metadata: &capi.Metadata{
			Labels: map[string]string{
				"env": "production",
			},
			Annotations: map[string]string{
				"owner": "team-a",
			},
		},
	}

	binding, err := serviceRouteBindings.Update(context.Background(), "binding-guid", request)
	require.NoError(t, err)
	assert.NotNil(t, binding)
	assert.Equal(t, "binding-guid", binding.GUID)
}

func TestServiceRouteBindingsClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_route_bindings/binding-guid", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		job := capi.Job{
			Resource: capi.Resource{
				GUID: "job-guid",
			},
			Operation: "service_route_binding.delete",
			State:     "PROCESSING",
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location", "/v3/jobs/job-guid")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(job)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	serviceRouteBindings := NewServiceRouteBindingsClient(client.httpClient)

	job, err := serviceRouteBindings.Delete(context.Background(), "binding-guid")
	require.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, "job-guid", job.GUID)
	assert.Equal(t, "service_route_binding.delete", job.Operation)
}

func TestServiceRouteBindingsClient_GetParameters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_route_bindings/binding-guid/parameters", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		params := capi.ServiceRouteBindingParameters{
			Parameters: map[string]interface{}{
				"rate_limit": 100,
				"enabled":    true,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(params)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	serviceRouteBindings := NewServiceRouteBindingsClient(client.httpClient)

	params, err := serviceRouteBindings.GetParameters(context.Background(), "binding-guid")
	require.NoError(t, err)
	assert.NotNil(t, params)
	assert.Equal(t, float64(100), params.Parameters["rate_limit"])
	assert.Equal(t, true, params.Parameters["enabled"])
}

func TestServiceRouteBindingsClient_GetNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_route_bindings/binding-guid", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	serviceRouteBindings := NewServiceRouteBindingsClient(client.httpClient)

	binding, err := serviceRouteBindings.Get(context.Background(), "binding-guid")
	assert.Error(t, err)
	assert.Nil(t, binding)
}

func TestServiceRouteBindingsClient_CreateForbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_route_bindings", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	serviceRouteBindings := NewServiceRouteBindingsClient(client.httpClient)

	request := &capi.ServiceRouteBindingCreateRequest{
		Relationships: capi.ServiceRouteBindingRelationships{
			ServiceInstance: capi.Relationship{
				Data: &capi.RelationshipData{
					GUID: "instance-guid",
				},
			},
			Route: capi.Relationship{
				Data: &capi.RelationshipData{
					GUID: "route-guid",
				},
			},
		},
	}

	result, err := serviceRouteBindings.Create(context.Background(), request)
	assert.Error(t, err)
	assert.Nil(t, result)
}
