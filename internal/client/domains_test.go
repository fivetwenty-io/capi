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

func boolPtr(b bool) *bool {
	return &b
}

func TestDomainsClient_Create(t *testing.T) {
	tests := []struct {
		name         string
		request      *capi.DomainCreateRequest
		response     interface{}
		statusCode   int
		expectedPath string
		wantErr      bool
		errMessage   string
	}{
		{
			name:         "create shared domain",
			expectedPath: "/v3/domains",
			statusCode:   http.StatusCreated,
			request: &capi.DomainCreateRequest{
				Name:     "example.com",
				Internal: boolPtr(false),
				Metadata: &capi.Metadata{
					Labels: map[string]string{
						"environment": "production",
					},
				},
			},
			response: capi.Domain{
				Resource: capi.Resource{
					GUID:      "domain-guid",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					Links: capi.Links{
						"self": capi.Link{
							Href: "https://api.example.org/v3/domains/domain-guid",
						},
						"route_reservations": capi.Link{
							Href: "https://api.example.org/v3/domains/domain-guid/route_reservations",
						},
						"shared_organizations": capi.Link{
							Href: "https://api.example.org/v3/domains/domain-guid/relationships/shared_organizations",
						},
					},
				},
				Name:               "example.com",
				Internal:           false,
				SupportedProtocols: []string{"http", "tcp"},
				Relationships:      capi.DomainRelationships{},
				Metadata: &capi.Metadata{
					Labels: map[string]string{
						"environment": "production",
					},
				},
			},
			wantErr: false,
		},
		{
			name:         "create private domain for organization",
			expectedPath: "/v3/domains",
			statusCode:   http.StatusCreated,
			request: &capi.DomainCreateRequest{
				Name: "apps.example.com",
				Relationships: &capi.DomainRelationships{
					Organization: &capi.Relationship{
						Data: &capi.RelationshipData{
							GUID: "org-guid",
						},
					},
				},
			},
			response: capi.Domain{
				Resource: capi.Resource{
					GUID:      "domain-guid",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Name:               "apps.example.com",
				Internal:           false,
				SupportedProtocols: []string{"http"},
				Relationships: capi.DomainRelationships{
					Organization: &capi.Relationship{
						Data: &capi.RelationshipData{
							GUID: "org-guid",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:         "create internal domain",
			expectedPath: "/v3/domains",
			statusCode:   http.StatusCreated,
			request: &capi.DomainCreateRequest{
				Name:     "apps.internal",
				Internal: boolPtr(true),
			},
			response: capi.Domain{
				Resource: capi.Resource{
					GUID:      "domain-guid",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Name:               "apps.internal",
				Internal:           true,
				SupportedProtocols: []string{"http"},
				Relationships:      capi.DomainRelationships{},
			},
			wantErr: false,
		},
		{
			name:         "domain already exists",
			expectedPath: "/v3/domains",
			statusCode:   http.StatusUnprocessableEntity,
			request: &capi.DomainCreateRequest{
				Name: "existing.com",
			},
			response: map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"code":   10008,
						"title":  "CF-UnprocessableEntity",
						"detail": "Domain name existing.com is already in use",
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

				var requestBody capi.DomainCreateRequest
				err := json.NewDecoder(r.Body).Decode(&requestBody)
				require.NoError(t, err)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				_ = json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client, err := New(&capi.Config{APIEndpoint: server.URL})
			require.NoError(t, err)

			domain, err := client.Domains().Create(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
				assert.Nil(t, domain)
			} else {
				require.NoError(t, err)
				require.NotNil(t, domain)
				assert.NotEmpty(t, domain.GUID)
				assert.Equal(t, tt.request.Name, domain.Name)
			}
		})
	}
}

func TestDomainsClient_Get(t *testing.T) {
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
			guid:         "test-domain-guid",
			expectedPath: "/v3/domains/test-domain-guid",
			statusCode:   http.StatusOK,
			response: capi.Domain{
				Resource: capi.Resource{
					GUID:      "test-domain-guid",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Name:               "example.com",
				Internal:           false,
				SupportedProtocols: []string{"http", "tcp"},
				Relationships:      capi.DomainRelationships{},
			},
			wantErr: false,
		},
		{
			name:         "domain not found",
			guid:         "non-existent-guid",
			expectedPath: "/v3/domains/non-existent-guid",
			statusCode:   http.StatusNotFound,
			response: map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"code":   10010,
						"title":  "CF-ResourceNotFound",
						"detail": "Domain not found",
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
				_ = json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client, err := New(&capi.Config{APIEndpoint: server.URL})
			require.NoError(t, err)

			domain, err := client.Domains().Get(context.Background(), tt.guid)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
				assert.Nil(t, domain)
			} else {
				require.NoError(t, err)
				require.NotNil(t, domain)
				assert.Equal(t, tt.guid, domain.GUID)
				assert.Equal(t, "example.com", domain.Name)
			}
		})
	}
}

func TestDomainsClient_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/domains", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		// Check query parameters if present
		query := r.URL.Query()
		if names := query.Get("names"); names != "" {
			assert.Equal(t, "example.com,test.com", names)
		}
		if orgGuids := query.Get("organization_guids"); orgGuids != "" {
			assert.Equal(t, "org-1,org-2", orgGuids)
		}

		response := capi.ListResponse[capi.Domain]{
			Pagination: capi.Pagination{
				TotalResults: 2,
				TotalPages:   1,
				First:        capi.Link{Href: "https://api.example.org/v3/domains?page=1"},
				Last:         capi.Link{Href: "https://api.example.org/v3/domains?page=1"},
				Next:         nil,
				Previous:     nil,
			},
			Resources: []capi.Domain{
				{
					Resource: capi.Resource{
						GUID:      "domain-1",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:               "example.com",
					Internal:           false,
					SupportedProtocols: []string{"http"},
				},
				{
					Resource: capi.Resource{
						GUID:      "domain-2",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:               "apps.internal",
					Internal:           true,
					SupportedProtocols: []string{"http"},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	// Test without filters
	result, err := client.Domains().List(context.Background(), nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 2, result.Pagination.TotalResults)
	assert.Len(t, result.Resources, 2)
	assert.Equal(t, "domain-1", result.Resources[0].GUID)
	assert.Equal(t, "example.com", result.Resources[0].Name)

	// Test with filters
	params := &capi.QueryParams{
		Filters: map[string][]string{
			"names":              {"example.com", "test.com"},
			"organization_guids": {"org-1", "org-2"},
		},
	}
	result, err = client.Domains().List(context.Background(), params)
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestDomainsClient_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/domains/test-domain-guid", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var requestBody capi.DomainUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		require.NoError(t, err)

		response := capi.Domain{
			Resource: capi.Resource{
				GUID:      "test-domain-guid",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:     "example.com",
			Metadata: requestBody.Metadata,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	request := &capi.DomainUpdateRequest{
		Metadata: &capi.Metadata{
			Labels: map[string]string{
				"environment": "staging",
			},
			Annotations: map[string]string{
				"note": "Updated domain",
			},
		},
	}

	domain, err := client.Domains().Update(context.Background(), "test-domain-guid", request)
	require.NoError(t, err)
	require.NotNil(t, domain)
	assert.Equal(t, "test-domain-guid", domain.GUID)
	assert.Equal(t, "staging", domain.Metadata.Labels["environment"])
	assert.Equal(t, "Updated domain", domain.Metadata.Annotations["note"])
}

func TestDomainsClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/domains/test-domain-guid", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		job := capi.Job{
			Resource: capi.Resource{
				GUID: "job-guid",
			},
			Operation: "domain.delete",
			State:     "PROCESSING",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(job)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	job, err := client.Domains().Delete(context.Background(), "test-domain-guid")
	require.NoError(t, err)
	require.NotNil(t, job)
	assert.Equal(t, "job-guid", job.GUID)
	assert.Equal(t, "domain.delete", job.Operation)
	assert.Equal(t, "PROCESSING", job.State)
}

func TestDomainsClient_ShareWithOrganization(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/domains/test-domain-guid/relationships/shared_organizations", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var requestBody struct {
			Data []capi.RelationshipData `json:"data"`
		}
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		require.NoError(t, err)
		assert.Len(t, requestBody.Data, 2)

		response := capi.ToManyRelationship{
			Data: []capi.RelationshipData{
				{GUID: "org-1"},
				{GUID: "org-2"},
				{GUID: "org-3"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	relationship, err := client.Domains().ShareWithOrganization(context.Background(), "test-domain-guid", []string{"org-1", "org-2"})
	require.NoError(t, err)
	require.NotNil(t, relationship)
	assert.Len(t, relationship.Data, 3)
}

func TestDomainsClient_UnshareFromOrganization(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/domains/test-domain-guid/relationships/shared_organizations/org-guid", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	err = client.Domains().UnshareFromOrganization(context.Background(), "test-domain-guid", "org-guid")
	require.NoError(t, err)
}

func TestDomainsClient_CheckRouteReservations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/domains/test-domain-guid/route_reservations", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		// Check query parameters
		query := r.URL.Query()
		assert.Equal(t, "api", query.Get("host"))
		assert.Equal(t, "/v1", query.Get("path"))

		response := capi.RouteReservation{
			MatchingRoute: &capi.Route{
				Resource: capi.Resource{
					GUID: "route-guid",
				},
				Host: "api",
				Path: "/v1",
				URL:  "api.example.com/v1",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	request := &capi.RouteReservationRequest{
		Host: "api",
		Path: "/v1",
	}

	reservation, err := client.Domains().CheckRouteReservations(context.Background(), "test-domain-guid", request)
	require.NoError(t, err)
	require.NotNil(t, reservation)
	assert.NotNil(t, reservation.MatchingRoute)
	assert.Equal(t, "route-guid", reservation.MatchingRoute.GUID)
	assert.Equal(t, "api", reservation.MatchingRoute.Host)
}
