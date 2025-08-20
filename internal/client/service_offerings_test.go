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

	"github.com/fivetwenty-io/capi-client/pkg/capi"
)

func TestServiceOfferingsClient_Get(t *testing.T) {
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
			guid:         "test-offering-guid",
			expectedPath: "/v3/service_offerings/test-offering-guid",
			statusCode:   http.StatusOK,
			response: capi.ServiceOffering{
				Resource: capi.Resource{
					GUID:      "test-offering-guid",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					Links: capi.Links{
						"self": capi.Link{
							Href: "https://api.example.org/v3/service_offerings/test-offering-guid",
						},
						"service_plans": capi.Link{
							Href: "https://api.example.org/v3/service_plans?service_offering_guids=test-offering-guid",
						},
						"service_broker": capi.Link{
							Href: "https://api.example.org/v3/service_brokers/broker-guid",
						},
					},
				},
				Name:             "my_service_offering",
				Description:      "Provides my service",
				Available:        true,
				Tags:             []string{"relational", "caching"},
				Requires:         []string{},
				Shareable:        true,
				DocumentationURL: stringPtr("https://some-documentation-link.io"),
				BrokerCatalog: capi.ServiceOfferingCatalog{
					ID: "db730a8c-11e5-11ea-838a-0f4fff3b1cfb",
					Metadata: map[string]interface{}{
						"shareable": true,
					},
					Features: capi.ServiceOfferingCatalogFeatures{
						PlanUpdateable:       true,
						Bindable:             true,
						InstancesRetrievable: true,
						BindingsRetrievable:  true,
						AllowContextUpdates:  false,
					},
				},
				Relationships: capi.ServiceOfferingRelationships{
					ServiceBroker: capi.Relationship{
						Data: &capi.RelationshipData{
							GUID: "broker-guid",
						},
					},
				},
				Metadata: &capi.Metadata{
					Labels: map[string]string{
						"type": "database",
					},
				},
			},
			wantErr: false,
		},
		{
			name:         "offering not found",
			guid:         "non-existent-guid",
			expectedPath: "/v3/service_offerings/non-existent-guid",
			statusCode:   http.StatusNotFound,
			response: map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"code":   10010,
						"title":  "CF-ResourceNotFound",
						"detail": "Service offering not found",
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

			offering, err := client.ServiceOfferings().Get(context.Background(), tt.guid)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
				assert.Nil(t, offering)
			} else {
				require.NoError(t, err)
				require.NotNil(t, offering)
				assert.Equal(t, tt.guid, offering.GUID)
				assert.Equal(t, "my_service_offering", offering.Name)
				assert.True(t, offering.Available)
			}
		})
	}
}

func TestServiceOfferingsClient_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_offerings", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		// Check query parameters if present
		query := r.URL.Query()
		if names := query.Get("names"); names != "" {
			assert.Equal(t, "offering1,offering2", names)
		}
		if brokerGuids := query.Get("service_broker_guids"); brokerGuids != "" {
			assert.Equal(t, "broker-1,broker-2", brokerGuids)
		}
		if available := query.Get("available"); available != "" {
			assert.Equal(t, "true", available)
		}

		response := capi.ListResponse[capi.ServiceOffering]{
			Pagination: capi.Pagination{
				TotalResults: 2,
				TotalPages:   1,
				First:        capi.Link{Href: "https://api.example.org/v3/service_offerings?page=1"},
				Last:         capi.Link{Href: "https://api.example.org/v3/service_offerings?page=1"},
				Next:         nil,
				Previous:     nil,
			},
			Resources: []capi.ServiceOffering{
				{
					Resource: capi.Resource{
						GUID:      "offering-1",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:        "database-service",
					Description: "A database service",
					Available:   true,
					Tags:        []string{"database", "sql"},
					Requires:    []string{},
					Shareable:   true,
					BrokerCatalog: capi.ServiceOfferingCatalog{
						ID: "catalog-id-1",
						Features: capi.ServiceOfferingCatalogFeatures{
							Bindable: true,
						},
					},
					Relationships: capi.ServiceOfferingRelationships{
						ServiceBroker: capi.Relationship{
							Data: &capi.RelationshipData{
								GUID: "broker-1",
							},
						},
					},
				},
				{
					Resource: capi.Resource{
						GUID:      "offering-2",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:        "cache-service",
					Description: "A caching service",
					Available:   false,
					Tags:        []string{"cache", "memory"},
					Requires:    []string{"route_forwarding"},
					Shareable:   false,
					BrokerCatalog: capi.ServiceOfferingCatalog{
						ID: "catalog-id-2",
						Features: capi.ServiceOfferingCatalogFeatures{
							Bindable:             true,
							InstancesRetrievable: true,
						},
					},
					Relationships: capi.ServiceOfferingRelationships{
						ServiceBroker: capi.Relationship{
							Data: &capi.RelationshipData{
								GUID: "broker-2",
							},
						},
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

	// Test without filters
	result, err := client.ServiceOfferings().List(context.Background(), nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 2, result.Pagination.TotalResults)
	assert.Len(t, result.Resources, 2)
	assert.Equal(t, "offering-1", result.Resources[0].GUID)
	assert.Equal(t, "database-service", result.Resources[0].Name)
	assert.True(t, result.Resources[0].Available)
	assert.Equal(t, "offering-2", result.Resources[1].GUID)
	assert.Equal(t, "cache-service", result.Resources[1].Name)
	assert.False(t, result.Resources[1].Available)

	// Test with filters
	params := &capi.QueryParams{
		Filters: map[string][]string{
			"names":                {"offering1", "offering2"},
			"service_broker_guids": {"broker-1", "broker-2"},
			"available":            {"true"},
		},
	}
	result, err = client.ServiceOfferings().List(context.Background(), params)
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestServiceOfferingsClient_Update(t *testing.T) {
	tests := []struct {
		name         string
		guid         string
		request      *capi.ServiceOfferingUpdateRequest
		response     interface{}
		statusCode   int
		expectedPath string
		wantErr      bool
		errMessage   string
	}{
		{
			name:         "successful update",
			guid:         "test-offering-guid",
			expectedPath: "/v3/service_offerings/test-offering-guid",
			statusCode:   http.StatusOK,
			request: &capi.ServiceOfferingUpdateRequest{
				Metadata: &capi.Metadata{
					Labels: map[string]string{
						"environment": "production",
					},
					Annotations: map[string]string{
						"note": "Updated offering",
					},
				},
			},
			response: capi.ServiceOffering{
				Resource: capi.Resource{
					GUID:      "test-offering-guid",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Name:        "my_service_offering",
				Description: "Provides my service",
				Available:   true,
				Tags:        []string{"relational", "caching"},
				Requires:    []string{},
				Shareable:   true,
				BrokerCatalog: capi.ServiceOfferingCatalog{
					ID: "catalog-id",
					Features: capi.ServiceOfferingCatalogFeatures{
						Bindable: true,
					},
				},
				Relationships: capi.ServiceOfferingRelationships{
					ServiceBroker: capi.Relationship{
						Data: &capi.RelationshipData{
							GUID: "broker-guid",
						},
					},
				},
				Metadata: &capi.Metadata{
					Labels: map[string]string{
						"environment": "production",
					},
					Annotations: map[string]string{
						"note": "Updated offering",
					},
				},
			},
			wantErr: false,
		},
		{
			name:         "offering not found",
			guid:         "non-existent-guid",
			expectedPath: "/v3/service_offerings/non-existent-guid",
			statusCode:   http.StatusNotFound,
			request: &capi.ServiceOfferingUpdateRequest{
				Metadata: &capi.Metadata{
					Labels: map[string]string{
						"test": "value",
					},
				},
			},
			response: map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"code":   10010,
						"title":  "CF-ResourceNotFound",
						"detail": "Service offering not found",
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
				assert.Equal(t, "PATCH", r.Method)

				var requestBody capi.ServiceOfferingUpdateRequest
				err := json.NewDecoder(r.Body).Decode(&requestBody)
				require.NoError(t, err)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client, err := New(&capi.Config{APIEndpoint: server.URL})
			require.NoError(t, err)

			offering, err := client.ServiceOfferings().Update(context.Background(), tt.guid, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
				assert.Nil(t, offering)
			} else {
				require.NoError(t, err)
				require.NotNil(t, offering)
				assert.Equal(t, tt.guid, offering.GUID)
				assert.Equal(t, "my_service_offering", offering.Name)
			}
		})
	}
}

func TestServiceOfferingsClient_Delete(t *testing.T) {
	tests := []struct {
		name         string
		guid         string
		statusCode   int
		expectedPath string
		wantErr      bool
		errMessage   string
		response     interface{}
	}{
		{
			name:         "successful delete",
			guid:         "test-offering-guid",
			expectedPath: "/v3/service_offerings/test-offering-guid",
			statusCode:   http.StatusNoContent,
			wantErr:      false,
		},
		{
			name:         "offering not found",
			guid:         "non-existent-guid",
			expectedPath: "/v3/service_offerings/non-existent-guid",
			statusCode:   http.StatusNotFound,
			response: map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"code":   10010,
						"title":  "CF-ResourceNotFound",
						"detail": "Service offering not found",
					},
				},
			},
			wantErr:    true,
			errMessage: "CF-ResourceNotFound",
		},
		{
			name:         "offering has service instances",
			guid:         "offering-with-instances",
			expectedPath: "/v3/service_offerings/offering-with-instances",
			statusCode:   http.StatusUnprocessableEntity,
			response: map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"code":   10008,
						"title":  "CF-UnprocessableEntity",
						"detail": "Service offering has service instances",
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
				assert.Equal(t, "DELETE", r.Method)

				if tt.response != nil {
					w.Header().Set("Content-Type", "application/json")
				}
				w.WriteHeader(tt.statusCode)
				if tt.response != nil {
					json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()

			client, err := New(&capi.Config{APIEndpoint: server.URL})
			require.NoError(t, err)

			err = client.ServiceOfferings().Delete(context.Background(), tt.guid)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
