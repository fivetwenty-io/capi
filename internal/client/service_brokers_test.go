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

func TestServiceBrokersClient_Create(t *testing.T) {
	tests := []struct {
		name         string
		request      *capi.ServiceBrokerCreateRequest
		response     interface{}
		statusCode   int
		expectedPath string
		wantErr      bool
		errMessage   string
	}{
		{
			name:         "create global service broker",
			expectedPath: "/v3/service_brokers",
			statusCode:   http.StatusAccepted,
			request: &capi.ServiceBrokerCreateRequest{
				Name: "my-service-broker",
				URL:  "https://example.service-broker.com",
				Authentication: capi.ServiceBrokerAuthentication{
					Type: "basic",
					Credentials: capi.ServiceBrokerAuthenticationCredentials{
						Username: "admin",
						Password: "secretpassword",
					},
				},
				Metadata: &capi.Metadata{
					Labels: map[string]string{
						"type": "development",
					},
				},
			},
			response: capi.Job{
				Resource: capi.Resource{
					GUID:      "job-guid",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					Links: capi.Links{
						"self": capi.Link{
							Href: "https://api.example.org/v3/jobs/job-guid",
						},
					},
				},
				Operation: "service_broker.create",
				State:     "PROCESSING",
			},
			wantErr: false,
		},
		{
			name:         "create space-scoped service broker",
			expectedPath: "/v3/service_brokers",
			statusCode:   http.StatusAccepted,
			request: &capi.ServiceBrokerCreateRequest{
				Name: "space-broker",
				URL:  "https://space-broker.example.com",
				Authentication: capi.ServiceBrokerAuthentication{
					Type: "basic",
					Credentials: capi.ServiceBrokerAuthenticationCredentials{
						Username: "user",
						Password: "pass",
					},
				},
				Relationships: &capi.ServiceBrokerRelationships{
					Space: &capi.Relationship{
						Data: &capi.RelationshipData{
							GUID: "space-guid",
						},
					},
				},
			},
			response: capi.Job{
				Resource: capi.Resource{
					GUID: "job-guid",
				},
				Operation: "service_broker.create",
				State:     "PROCESSING",
			},
			wantErr: false,
		},
		{
			name:         "service broker already exists",
			expectedPath: "/v3/service_brokers",
			statusCode:   http.StatusUnprocessableEntity,
			request: &capi.ServiceBrokerCreateRequest{
				Name: "existing-broker",
				URL:  "https://existing.example.com",
				Authentication: capi.ServiceBrokerAuthentication{
					Type: "basic",
					Credentials: capi.ServiceBrokerAuthenticationCredentials{
						Username: "user",
						Password: "pass",
					},
				},
			},
			response: map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"code":   10008,
						"title":  "CF-UnprocessableEntity",
						"detail": "Service broker with name existing-broker already exists",
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

				var requestBody capi.ServiceBrokerCreateRequest
				err := json.NewDecoder(r.Body).Decode(&requestBody)
				require.NoError(t, err)

				w.Header().Set("Content-Type", "application/json")
				if tt.statusCode == http.StatusAccepted {
					w.Header().Set("Location", "https://api.example.org/v3/jobs/job-guid")
				}
				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client, err := New(&capi.Config{APIEndpoint: server.URL})
			require.NoError(t, err)

			job, err := client.ServiceBrokers().Create(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
				assert.Nil(t, job)
			} else {
				require.NoError(t, err)
				require.NotNil(t, job)
				assert.NotEmpty(t, job.GUID)
				assert.Equal(t, "service_broker.create", job.Operation)
			}
		})
	}
}

func TestServiceBrokersClient_Get(t *testing.T) {
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
			guid:         "test-broker-guid",
			expectedPath: "/v3/service_brokers/test-broker-guid",
			statusCode:   http.StatusOK,
			response: capi.ServiceBroker{
				Resource: capi.Resource{
					GUID:      "test-broker-guid",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					Links: capi.Links{
						"self": capi.Link{
							Href: "https://api.example.org/v3/service_brokers/test-broker-guid",
						},
						"service_offerings": capi.Link{
							Href: "https://api.example.org/v3/service_offerings?service_broker_guids=test-broker-guid",
						},
					},
				},
				Name: "my-service-broker",
				URL:  "https://example.service-broker.com",
				Relationships: capi.ServiceBrokerRelationships{
					Space: nil,
				},
				Metadata: &capi.Metadata{
					Labels: map[string]string{
						"type": "development",
					},
				},
			},
			wantErr: false,
		},
		{
			name:         "broker not found",
			guid:         "non-existent-guid",
			expectedPath: "/v3/service_brokers/non-existent-guid",
			statusCode:   http.StatusNotFound,
			response: map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"code":   10010,
						"title":  "CF-ResourceNotFound",
						"detail": "Service broker not found",
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

			broker, err := client.ServiceBrokers().Get(context.Background(), tt.guid)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
				assert.Nil(t, broker)
			} else {
				require.NoError(t, err)
				require.NotNil(t, broker)
				assert.Equal(t, tt.guid, broker.GUID)
				assert.Equal(t, "my-service-broker", broker.Name)
			}
		})
	}
}

func TestServiceBrokersClient_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_brokers", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		// Check query parameters if present
		query := r.URL.Query()
		if names := query.Get("names"); names != "" {
			assert.Equal(t, "broker1,broker2", names)
		}
		if spaceGuids := query.Get("space_guids"); spaceGuids != "" {
			assert.Equal(t, "space-1,space-2", spaceGuids)
		}

		response := capi.ListResponse[capi.ServiceBroker]{
			Pagination: capi.Pagination{
				TotalResults: 2,
				TotalPages:   1,
				First:        capi.Link{Href: "https://api.example.org/v3/service_brokers?page=1"},
				Last:         capi.Link{Href: "https://api.example.org/v3/service_brokers?page=1"},
				Next:         nil,
				Previous:     nil,
			},
			Resources: []capi.ServiceBroker{
				{
					Resource: capi.Resource{
						GUID:      "broker-1",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name: "global-broker",
					URL:  "https://global-broker.example.com",
				},
				{
					Resource: capi.Resource{
						GUID:      "broker-2",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name: "space-broker",
					URL:  "https://space-broker.example.com",
					Relationships: capi.ServiceBrokerRelationships{
						Space: &capi.Relationship{
							Data: &capi.RelationshipData{
								GUID: "space-guid",
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
	result, err := client.ServiceBrokers().List(context.Background(), nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 2, result.Pagination.TotalResults)
	assert.Len(t, result.Resources, 2)
	assert.Equal(t, "broker-1", result.Resources[0].GUID)
	assert.Equal(t, "global-broker", result.Resources[0].Name)

	// Test with filters
	params := &capi.QueryParams{
		Filters: map[string][]string{
			"names":       {"broker1", "broker2"},
			"space_guids": {"space-1", "space-2"},
		},
	}
	result, err = client.ServiceBrokers().List(context.Background(), params)
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestServiceBrokersClient_Update(t *testing.T) {
	tests := []struct {
		name         string
		guid         string
		request      *capi.ServiceBrokerUpdateRequest
		response     interface{}
		statusCode   int
		withJob      bool
		expectedPath string
		wantErr      bool
		errMessage   string
	}{
		{
			name:         "update with job (URL change)",
			guid:         "test-broker-guid",
			expectedPath: "/v3/service_brokers/test-broker-guid",
			statusCode:   http.StatusAccepted,
			withJob:      true,
			request: &capi.ServiceBrokerUpdateRequest{
				URL: stringPtr("https://new.service-broker.com"),
				Authentication: &capi.ServiceBrokerAuthentication{
					Type: "basic",
					Credentials: capi.ServiceBrokerAuthenticationCredentials{
						Username: "newuser",
						Password: "newpass",
					},
				},
			},
			response: capi.Job{
				Resource: capi.Resource{
					GUID: "job-guid",
				},
				Operation: "service_broker.update",
				State:     "PROCESSING",
			},
			wantErr: false,
		},
		{
			name:         "update without job (metadata only)",
			guid:         "test-broker-guid",
			expectedPath: "/v3/service_brokers/test-broker-guid",
			statusCode:   http.StatusOK,
			withJob:      false,
			request: &capi.ServiceBrokerUpdateRequest{
				Metadata: &capi.Metadata{
					Labels: map[string]string{
						"environment": "production",
					},
				},
			},
			response: capi.ServiceBroker{
				Resource: capi.Resource{
					GUID:      "test-broker-guid",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Name: "my-service-broker",
				URL:  "https://example.service-broker.com",
				Metadata: &capi.Metadata{
					Labels: map[string]string{
						"environment": "production",
					},
				},
			},
			wantErr: false,
		},
		{
			name:         "update with synchronization in progress",
			guid:         "test-broker-guid",
			expectedPath: "/v3/service_brokers/test-broker-guid",
			statusCode:   http.StatusUnprocessableEntity,
			request: &capi.ServiceBrokerUpdateRequest{
				URL: stringPtr("https://new.service-broker.com"),
			},
			response: map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"code":   10008,
						"title":  "CF-UnprocessableEntity",
						"detail": "Cannot update service broker while synchronization is in progress",
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
				assert.Equal(t, "PATCH", r.Method)

				var requestBody capi.ServiceBrokerUpdateRequest
				err := json.NewDecoder(r.Body).Decode(&requestBody)
				require.NoError(t, err)

				w.Header().Set("Content-Type", "application/json")
				if tt.withJob && tt.statusCode == http.StatusAccepted {
					w.Header().Set("Location", "https://api.example.org/v3/jobs/job-guid")
				}
				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client, err := New(&capi.Config{APIEndpoint: server.URL})
			require.NoError(t, err)

			result, err := client.ServiceBrokers().Update(context.Background(), tt.guid, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.withJob {
					job := result
					assert.Equal(t, "job-guid", job.GUID)
					assert.Equal(t, "service_broker.update", job.Operation)
				}
			}
		})
	}
}

func TestServiceBrokersClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_brokers/test-broker-guid", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		job := capi.Job{
			Resource: capi.Resource{
				GUID: "job-guid",
			},
			Operation: "service_broker.delete",
			State:     "PROCESSING",
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location", "https://api.example.org/v3/jobs/job-guid")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(job)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	job, err := client.ServiceBrokers().Delete(context.Background(), "test-broker-guid")
	require.NoError(t, err)
	require.NotNil(t, job)
	assert.Equal(t, "job-guid", job.GUID)
	assert.Equal(t, "service_broker.delete", job.Operation)
	assert.Equal(t, "PROCESSING", job.State)
}
