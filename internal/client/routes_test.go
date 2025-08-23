package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fivetwenty-io/capi/v3/pkg/capi"
)

func TestRoutesClient_Create(t *testing.T) {
	tests := []struct {
		name         string
		request      *capi.RouteCreateRequest
		response     interface{}
		statusCode   int
		expectedPath string
		wantErr      bool
		errMessage   string
	}{
		{
			name:         "create route with host",
			expectedPath: "/v3/routes",
			statusCode:   http.StatusCreated,
			request: &capi.RouteCreateRequest{
				Host: stringPtr("api"),
				Path: stringPtr("/v1"),
				Relationships: capi.RouteRelationships{
					Space:  capi.Relationship{Data: &capi.RelationshipData{GUID: "space-guid"}},
					Domain: capi.Relationship{Data: &capi.RelationshipData{GUID: "domain-guid"}},
				},
				Metadata: &capi.Metadata{
					Labels: map[string]string{
						"type": "api",
					},
				},
			},
			response: capi.Route{
				Resource: capi.Resource{
					GUID:      "route-guid",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					Links: capi.Links{
						"self": capi.Link{
							Href: "https://api.example.org/v3/routes/route-guid",
						},
						"space": capi.Link{
							Href: "https://api.example.org/v3/spaces/space-guid",
						},
						"domain": capi.Link{
							Href: "https://api.example.org/v3/domains/domain-guid",
						},
						"destinations": capi.Link{
							Href: "https://api.example.org/v3/routes/route-guid/destinations",
						},
					},
				},
				Protocol: "http",
				Host:     "api",
				Path:     "/v1",
				URL:      "api.example.com/v1",
				Relationships: capi.RouteRelationships{
					Space:  capi.Relationship{Data: &capi.RelationshipData{GUID: "space-guid"}},
					Domain: capi.Relationship{Data: &capi.RelationshipData{GUID: "domain-guid"}},
				},
				Metadata: &capi.Metadata{
					Labels: map[string]string{
						"type": "api",
					},
				},
			},
			wantErr: false,
		},
		{
			name:         "create route with port",
			expectedPath: "/v3/routes",
			statusCode:   http.StatusCreated,
			request: &capi.RouteCreateRequest{
				Port: intPtr(8080),
				Relationships: capi.RouteRelationships{
					Space:  capi.Relationship{Data: &capi.RelationshipData{GUID: "space-guid"}},
					Domain: capi.Relationship{Data: &capi.RelationshipData{GUID: "domain-guid"}},
				},
			},
			response: capi.Route{
				Resource: capi.Resource{
					GUID:      "route-guid",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Protocol: "tcp",
				Port:     intPtr(8080),
				URL:      "example.com:8080",
				Relationships: capi.RouteRelationships{
					Space:  capi.Relationship{Data: &capi.RelationshipData{GUID: "space-guid"}},
					Domain: capi.Relationship{Data: &capi.RelationshipData{GUID: "domain-guid"}},
				},
			},
			wantErr: false,
		},
		{
			name:         "route already exists",
			expectedPath: "/v3/routes",
			statusCode:   http.StatusUnprocessableEntity,
			request: &capi.RouteCreateRequest{
				Host: stringPtr("existing"),
				Relationships: capi.RouteRelationships{
					Space:  capi.Relationship{Data: &capi.RelationshipData{GUID: "space-guid"}},
					Domain: capi.Relationship{Data: &capi.RelationshipData{GUID: "domain-guid"}},
				},
			},
			response: map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"code":   10008,
						"title":  "CF-UnprocessableEntity",
						"detail": "Route already exists",
					},
				},
			},
			wantErr:    true,
			errMessage: "CF-UnprocessableEntity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.expectedPath, r.URL.Path)
				assert.Equal(t, "POST", r.Method)

				var requestBody capi.RouteCreateRequest
				err := json.NewDecoder(r.Body).Decode(&requestBody)
				require.NoError(t, err)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client, err := New(&capi.Config{APIEndpoint: server.URL})
			require.NoError(t, err)

			route, err := client.Routes().Create(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
				assert.Nil(t, route)
			} else {
				require.NoError(t, err)
				require.NotNil(t, route)
				assert.NotEmpty(t, route.GUID)
			}
		})
	}
}

func TestRoutesClient_Get(t *testing.T) {
	tests := []struct {
		name         string
		guid         string
		response     interface{}
		statusCode   int
		expectedPath string
		wantErr      bool
		errMessage   string
	}{
		{
			name:         "successful get",
			guid:         "test-route-guid",
			expectedPath: "/v3/routes/test-route-guid",
			statusCode:   http.StatusOK,
			response: capi.Route{
				Resource: capi.Resource{
					GUID:      "test-route-guid",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Protocol: "http",
				Host:     "api",
				Path:     "/v1",
				URL:      "api.example.com/v1",
			},
			wantErr: false,
		},
		{
			name:         "route not found",
			guid:         "non-existent-guid",
			expectedPath: "/v3/routes/non-existent-guid",
			statusCode:   http.StatusNotFound,
			response: map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"code":   10010,
						"title":  "CF-ResourceNotFound",
						"detail": "Route not found",
					},
				},
			},
			wantErr:    true,
			errMessage: "CF-ResourceNotFound",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.expectedPath, r.URL.Path)
				assert.Equal(t, "GET", r.Method)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client, err := New(&capi.Config{APIEndpoint: server.URL})
			require.NoError(t, err)

			route, err := client.Routes().Get(context.Background(), tt.guid)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
				assert.Nil(t, route)
			} else {
				require.NoError(t, err)
				require.NotNil(t, route)
				assert.Equal(t, tt.guid, route.GUID)
			}
		})
	}
}

func TestRoutesClient_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/routes", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		// Check query parameters if present
		query := r.URL.Query()
		if hosts := query.Get("hosts"); hosts != "" {
			assert.Equal(t, "api,www", hosts)
		}
		if spaceGuids := query.Get("space_guids"); spaceGuids != "" {
			assert.Equal(t, "space-1,space-2", spaceGuids)
		}

		response := capi.ListResponse[capi.Route]{
			Pagination: capi.Pagination{
				TotalResults: 2,
				TotalPages:   1,
				First:        capi.Link{Href: "https://api.example.org/v3/routes?page=1"},
				Last:         capi.Link{Href: "https://api.example.org/v3/routes?page=1"},
				Next:         nil,
				Previous:     nil,
			},
			Resources: []capi.Route{
				{
					Resource: capi.Resource{
						GUID:      "route-1",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Protocol: "http",
					Host:     "api",
					Path:     "/v1",
					URL:      "api.example.com/v1",
				},
				{
					Resource: capi.Resource{
						GUID:      "route-2",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Protocol: "tcp",
					Port:     intPtr(8080),
					URL:      "example.com:8080",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	// Test without filters
	result, err := client.Routes().List(context.Background(), nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 2, result.Pagination.TotalResults)
	assert.Len(t, result.Resources, 2)
	assert.Equal(t, "route-1", result.Resources[0].GUID)

	// Test with filters
	params := &capi.QueryParams{
		Filters: map[string][]string{
			"hosts":       {"api", "www"},
			"space_guids": {"space-1", "space-2"},
		},
	}
	result, err = client.Routes().List(context.Background(), params)
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestRoutesClient_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/routes/test-route-guid", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var requestBody capi.RouteUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		require.NoError(t, err)

		response := capi.Route{
			Resource: capi.Resource{
				GUID:      "test-route-guid",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Protocol: "http",
			Host:     "api",
			Path:     "/v1",
			URL:      "api.example.com/v1",
			Metadata: requestBody.Metadata,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	request := &capi.RouteUpdateRequest{
		Metadata: &capi.Metadata{
			Labels: map[string]string{
				"environment": "staging",
			},
			Annotations: map[string]string{
				"note": "Updated route",
			},
		},
	}

	route, err := client.Routes().Update(context.Background(), "test-route-guid", request)
	require.NoError(t, err)
	require.NotNil(t, route)
	assert.Equal(t, "test-route-guid", route.GUID)
	assert.Equal(t, "staging", route.Metadata.Labels["environment"])
}

func TestRoutesClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/routes/test-route-guid", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		job := capi.Job{
			Resource: capi.Resource{
				GUID: "job-guid",
			},
			Operation: "route.delete",
			State:     "PROCESSING",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(job)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	job, err := client.Routes().Delete(context.Background(), "test-route-guid")
	require.NoError(t, err)
	require.NotNil(t, job)
	assert.Equal(t, "job-guid", job.GUID)
	assert.Equal(t, "route.delete", job.Operation)
}

func TestRoutesClient_ListDestinations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/routes/test-route-guid/destinations", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		response := capi.RouteDestinations{
			Destinations: []capi.RouteDestination{
				{
					GUID: "dest-1",
					App: capi.RouteDestinationApp{
						GUID: "app-1",
						Process: &capi.Process{
							Resource: capi.Resource{GUID: "process-1"},
							Type:     "web",
						},
					},
					Port:     intPtr(8080),
					Protocol: stringPtr("http1"),
					Weight:   intPtr(100),
				},
				{
					GUID: "dest-2",
					App: capi.RouteDestinationApp{
						GUID: "app-2",
					},
				},
			},
			Links: capi.Links{
				"self": capi.Link{
					Href: "https://api.example.org/v3/routes/test-route-guid/destinations",
				},
				"route": capi.Link{
					Href: "https://api.example.org/v3/routes/test-route-guid",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	destinations, err := client.Routes().ListDestinations(context.Background(), "test-route-guid")
	require.NoError(t, err)
	require.NotNil(t, destinations)
	assert.Len(t, destinations.Destinations, 2)
	assert.Equal(t, "dest-1", destinations.Destinations[0].GUID)
	assert.Equal(t, "app-1", destinations.Destinations[0].App.GUID)
}

func TestRoutesClient_InsertDestinations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/routes/test-route-guid/destinations", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var requestBody struct {
			Destinations []capi.RouteDestination `json:"destinations"`
		}
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		require.NoError(t, err)
		assert.Len(t, requestBody.Destinations, 1)

		response := capi.RouteDestinations{
			Destinations: []capi.RouteDestination{
				{
					GUID: "dest-1",
					App: capi.RouteDestinationApp{
						GUID: "app-1",
					},
				},
				{
					GUID: "dest-2",
					App: capi.RouteDestinationApp{
						GUID: "app-2",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	newDestinations := []capi.RouteDestination{
		{
			App: capi.RouteDestinationApp{
				GUID: "app-2",
			},
		},
	}

	destinations, err := client.Routes().InsertDestinations(context.Background(), "test-route-guid", newDestinations)
	require.NoError(t, err)
	require.NotNil(t, destinations)
	assert.Len(t, destinations.Destinations, 2)
}

func TestRoutesClient_ReplaceDestinations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/routes/test-route-guid/destinations", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var requestBody struct {
			Destinations []capi.RouteDestination `json:"destinations"`
		}
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		require.NoError(t, err)
		assert.Len(t, requestBody.Destinations, 1)

		response := capi.RouteDestinations{
			Destinations: requestBody.Destinations,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	newDestinations := []capi.RouteDestination{
		{
			App: capi.RouteDestinationApp{
				GUID: "app-new",
			},
		},
	}

	destinations, err := client.Routes().ReplaceDestinations(context.Background(), "test-route-guid", newDestinations)
	require.NoError(t, err)
	require.NotNil(t, destinations)
	assert.Len(t, destinations.Destinations, 1)
}

func TestRoutesClient_UpdateDestination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/routes/test-route-guid/destinations/dest-guid", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var requestBody struct {
			Protocol string `json:"protocol"`
		}
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		require.NoError(t, err)
		assert.Equal(t, "http2", requestBody.Protocol)

		response := capi.RouteDestination{
			GUID: "dest-guid",
			App: capi.RouteDestinationApp{
				GUID: "app-1",
			},
			Protocol: stringPtr("http2"),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	destination, err := client.Routes().UpdateDestination(context.Background(), "test-route-guid", "dest-guid", "http2")
	require.NoError(t, err)
	require.NotNil(t, destination)
	assert.Equal(t, "dest-guid", destination.GUID)
	assert.Equal(t, "http2", *destination.Protocol)
}

func TestRoutesClient_RemoveDestination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/routes/test-route-guid/destinations/dest-guid", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	err = client.Routes().RemoveDestination(context.Background(), "test-route-guid", "dest-guid")
	require.NoError(t, err)
}

func TestRoutesClient_ShareWithSpace(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/routes/test-route-guid/relationships/shared_spaces", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var requestBody struct {
			Data []capi.RelationshipData `json:"data"`
		}
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		require.NoError(t, err)
		assert.Len(t, requestBody.Data, 2)

		response := capi.ToManyRelationship{
			Data: []capi.RelationshipData{
				{GUID: "space-1"},
				{GUID: "space-2"},
				{GUID: "space-3"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	relationship, err := client.Routes().ShareWithSpace(context.Background(), "test-route-guid", []string{"space-1", "space-2"})
	require.NoError(t, err)
	require.NotNil(t, relationship)
	assert.Len(t, relationship.Data, 3)
}

func TestRoutesClient_UnshareFromSpace(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/routes/test-route-guid/relationships/shared_spaces/space-guid", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	err = client.Routes().UnshareFromSpace(context.Background(), "test-route-guid", "space-guid")
	require.NoError(t, err)
}

func TestRoutesClient_TransferOwnership(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/routes/test-route-guid", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var requestBody struct {
			Relationships capi.RouteRelationships `json:"relationships"`
		}
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		require.NoError(t, err)
		assert.Equal(t, "new-space-guid", requestBody.Relationships.Space.Data.GUID)

		response := capi.Route{
			Resource: capi.Resource{
				GUID:      "test-route-guid",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Protocol: "http",
			Host:     "api",
			Path:     "/v1",
			URL:      "api.example.com/v1",
			Relationships: capi.RouteRelationships{
				Space:  capi.Relationship{Data: &capi.RelationshipData{GUID: "new-space-guid"}},
				Domain: capi.Relationship{Data: &capi.RelationshipData{GUID: "domain-guid"}},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	route, err := client.Routes().TransferOwnership(context.Background(), "test-route-guid", "new-space-guid")
	require.NoError(t, err)
	require.NotNil(t, route)
	assert.Equal(t, "test-route-guid", route.GUID)
	assert.Equal(t, "new-space-guid", route.Relationships.Space.Data.GUID)
}
