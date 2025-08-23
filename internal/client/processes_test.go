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

	"github.com/fivetwenty-io/capi/pkg/capi"
)

func TestProcessesClient_Get(t *testing.T) {
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
			guid:         "test-process-guid",
			expectedPath: "/v3/processes/test-process-guid",
			statusCode:   http.StatusOK,
			response: capi.Process{
				Resource: capi.Resource{
					GUID:      "test-process-guid",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					Links: capi.Links{
						"self": capi.Link{
							Href: "https://api.example.org/v3/processes/test-process-guid",
						},
						"scale": capi.Link{
							Href:   "https://api.example.org/v3/processes/test-process-guid/actions/scale",
							Method: "POST",
						},
						"app": capi.Link{
							Href: "https://api.example.org/v3/apps/app-guid",
						},
					},
				},
				Metadata: &capi.Metadata{
					Labels: map[string]string{
						"environment": "production",
					},
					Annotations: map[string]string{
						"note": "web process",
					},
				},
				Type:                         "web",
				Command:                      stringPtr("bundle exec rackup"),
				Instances:                    5,
				MemoryInMB:                   256,
				DiskInMB:                     1024,
				LogRateLimitInBytesPerSecond: intPtr(1024),
				HealthCheck: &capi.HealthCheck{
					Type: "port",
					Data: &capi.HealthCheckData{
						Timeout: intPtr(60),
					},
				},
				ReadinessHealthCheck: &capi.ReadinessHealthCheck{
					Type: "process",
					Data: &capi.ReadinessHealthCheckData{
						InvocationTimeout: intPtr(10),
					},
				},
				Relationships: &capi.ProcessRelationships{
					App: &capi.Relationship{
						Data: &capi.RelationshipData{
							GUID: "app-guid",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:         "not found",
			guid:         "non-existent-guid",
			expectedPath: "/v3/processes/non-existent-guid",
			statusCode:   http.StatusNotFound,
			response: map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"code":   10010,
						"title":  "CF-ResourceNotFound",
						"detail": "Process not found",
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

			process, err := client.Processes().Get(context.Background(), tt.guid)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
				assert.Nil(t, process)
			} else {
				require.NoError(t, err)
				require.NotNil(t, process)
				assert.Equal(t, tt.guid, process.GUID)
				assert.Equal(t, "web", process.Type)
				assert.Equal(t, "bundle exec rackup", *process.Command)
				assert.Equal(t, 5, process.Instances)
				assert.Equal(t, 256, process.MemoryInMB)
				assert.Equal(t, 1024, process.DiskInMB)
				assert.NotNil(t, process.HealthCheck)
				assert.Equal(t, "port", process.HealthCheck.Type)
				assert.NotNil(t, process.ReadinessHealthCheck)
				assert.Equal(t, "process", process.ReadinessHealthCheck.Type)
				assert.NotNil(t, process.Relationships)
				assert.Equal(t, "app-guid", process.Relationships.App.Data.GUID)
			}
		})
	}
}

func TestProcessesClient_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/processes", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		// Check query parameters if present
		query := r.URL.Query()
		if appGuids := query.Get("app_guids"); appGuids != "" {
			assert.Equal(t, "app-1,app-2", appGuids)
		}

		response := capi.ListResponse[capi.Process]{
			Pagination: capi.Pagination{
				TotalResults: 2,
				TotalPages:   1,
				First:        capi.Link{Href: "https://api.example.org/v3/processes?page=1"},
				Last:         capi.Link{Href: "https://api.example.org/v3/processes?page=1"},
				Next:         nil,
				Previous:     nil,
			},
			Resources: []capi.Process{
				{
					Resource: capi.Resource{
						GUID:      "process-1",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Type:       "web",
					Command:    stringPtr("bundle exec rackup"),
					Instances:  3,
					MemoryInMB: 256,
					DiskInMB:   512,
				},
				{
					Resource: capi.Resource{
						GUID:      "process-2",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Type:       "worker",
					Command:    stringPtr("bundle exec sidekiq"),
					Instances:  1,
					MemoryInMB: 512,
					DiskInMB:   1024,
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
	result, err := client.Processes().List(context.Background(), nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 2, result.Pagination.TotalResults)
	assert.Len(t, result.Resources, 2)
	assert.Equal(t, "process-1", result.Resources[0].GUID)
	assert.Equal(t, "web", result.Resources[0].Type)
	assert.Equal(t, "process-2", result.Resources[1].GUID)
	assert.Equal(t, "worker", result.Resources[1].Type)

	// Test with filters
	params := &capi.QueryParams{
		Filters: map[string][]string{
			"app_guids": {"app-1", "app-2"},
			"types":     {"web"},
		},
	}
	result, err = client.Processes().List(context.Background(), params)
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestProcessesClient_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/processes/test-process-guid", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var requestBody capi.ProcessUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		require.NoError(t, err)

		response := capi.Process{
			Resource: capi.Resource{
				GUID:      "test-process-guid",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Type:       "web",
			Command:    requestBody.Command,
			Instances:  3,
			MemoryInMB: 256,
			DiskInMB:   512,
			Metadata:   requestBody.Metadata,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	request := &capi.ProcessUpdateRequest{
		Command: stringPtr("new command"),
		Metadata: &capi.Metadata{
			Labels: map[string]string{
				"env": "staging",
			},
		},
	}

	process, err := client.Processes().Update(context.Background(), "test-process-guid", request)
	require.NoError(t, err)
	require.NotNil(t, process)
	assert.Equal(t, "test-process-guid", process.GUID)
	assert.Equal(t, "new command", *process.Command)
	assert.Equal(t, "staging", process.Metadata.Labels["env"])
}

func TestProcessesClient_Scale(t *testing.T) {
	tests := []struct {
		name         string
		guid         string
		request      *capi.ProcessScaleRequest
		expectedPath string
	}{
		{
			name:         "scale instances and resources",
			guid:         "test-process-guid",
			expectedPath: "/v3/processes/test-process-guid/actions/scale",
			request: &capi.ProcessScaleRequest{
				Instances:  intPtr(10),
				MemoryInMB: intPtr(512),
				DiskInMB:   intPtr(2048),
			},
		},
		{
			name:         "scale with log rate limit",
			guid:         "test-process-guid",
			expectedPath: "/v3/processes/test-process-guid/actions/scale",
			request: &capi.ProcessScaleRequest{
				Instances:                    intPtr(5),
				LogRateLimitInBytesPerSecond: intPtr(2048),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.expectedPath, r.URL.Path)
				assert.Equal(t, "POST", r.Method)

				var requestBody capi.ProcessScaleRequest
				err := json.NewDecoder(r.Body).Decode(&requestBody)
				require.NoError(t, err)

				response := capi.Process{
					Resource: capi.Resource{
						GUID:      tt.guid,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Type: "web",
				}

				if requestBody.Instances != nil {
					response.Instances = *requestBody.Instances
				}
				if requestBody.MemoryInMB != nil {
					response.MemoryInMB = *requestBody.MemoryInMB
				}
				if requestBody.DiskInMB != nil {
					response.DiskInMB = *requestBody.DiskInMB
				}
				if requestBody.LogRateLimitInBytesPerSecond != nil {
					response.LogRateLimitInBytesPerSecond = requestBody.LogRateLimitInBytesPerSecond
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusAccepted)
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client, err := New(&capi.Config{APIEndpoint: server.URL})
			require.NoError(t, err)

			process, err := client.Processes().Scale(context.Background(), tt.guid, tt.request)
			require.NoError(t, err)
			require.NotNil(t, process)
			assert.Equal(t, tt.guid, process.GUID)

			if tt.name == "scale instances and resources" {
				assert.Equal(t, 10, process.Instances)
				assert.Equal(t, 512, process.MemoryInMB)
				assert.Equal(t, 2048, process.DiskInMB)
			} else if tt.name == "scale with log rate limit" {
				assert.Equal(t, 5, process.Instances)
				assert.Equal(t, 2048, *process.LogRateLimitInBytesPerSecond)
			}
		})
	}
}

func TestProcessesClient_GetStats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/processes/test-process-guid/stats", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		isolationSegment := "default"
		response := capi.ProcessStats{
			Pagination: &capi.Pagination{
				TotalResults: 2,
				TotalPages:   1,
			},
			Resources: []capi.ProcessStatsDetail{
				{
					Type:  "web",
					Index: 0,
					State: "RUNNING",
					Usage: &capi.ProcessUsage{
						Time:           time.Now().Format(time.RFC3339Nano),
						CPU:            0.15,
						CPUEntitlement: 0.2,
						Mem:            134217728,
						Disk:           268435456,
						LogRate:        100,
					},
					Host: "10.0.0.1",
					InstancePorts: []capi.ProcessInstancePort{
						{
							External:             61001,
							Internal:             8080,
							ExternalTLSProxyPort: 61443,
							InternalTLSProxyPort: 61002,
						},
					},
					Uptime:           3600,
					MemQuota:         268435456,
					DiskQuota:        1073741824,
					FdsQuota:         16384,
					IsolationSegment: &isolationSegment,
				},
				{
					Type:  "web",
					Index: 1,
					State: "RUNNING",
					Usage: &capi.ProcessUsage{
						Time:    time.Now().Format(time.RFC3339Nano),
						CPU:     0.12,
						Mem:     100000000,
						Disk:    200000000,
						LogRate: 50,
					},
					Host:      "10.0.0.2",
					Uptime:    3600,
					MemQuota:  268435456,
					DiskQuota: 1073741824,
					FdsQuota:  16384,
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

	stats, err := client.Processes().GetStats(context.Background(), "test-process-guid")
	require.NoError(t, err)
	require.NotNil(t, stats)
	assert.Equal(t, 2, stats.Pagination.TotalResults)
	assert.Len(t, stats.Resources, 2)

	// Check first instance
	instance0 := stats.Resources[0]
	assert.Equal(t, "web", instance0.Type)
	assert.Equal(t, 0, instance0.Index)
	assert.Equal(t, "RUNNING", instance0.State)
	assert.NotNil(t, instance0.Usage)
	assert.Equal(t, 0.15, instance0.Usage.CPU)
	assert.Equal(t, int64(134217728), instance0.Usage.Mem)
	assert.Equal(t, "10.0.0.1", instance0.Host)
	assert.Len(t, instance0.InstancePorts, 1)
	assert.Equal(t, 61001, instance0.InstancePorts[0].External)
	assert.Equal(t, 8080, instance0.InstancePorts[0].Internal)
	assert.Equal(t, "default", *instance0.IsolationSegment)

	// Check second instance
	instance1 := stats.Resources[1]
	assert.Equal(t, 1, instance1.Index)
	assert.Equal(t, "RUNNING", instance1.State)
}

func TestProcessesClient_TerminateInstance(t *testing.T) {
	tests := []struct {
		name         string
		guid         string
		index        int
		statusCode   int
		expectedPath string
		wantErr      bool
	}{
		{
			name:         "successful terminate",
			guid:         "test-process-guid",
			index:        0,
			expectedPath: "/v3/processes/test-process-guid/instances/0",
			statusCode:   http.StatusNoContent,
			wantErr:      false,
		},
		{
			name:         "terminate another instance",
			guid:         "test-process-guid",
			index:        3,
			expectedPath: "/v3/processes/test-process-guid/instances/3",
			statusCode:   http.StatusNoContent,
			wantErr:      false,
		},
		{
			name:         "process not found",
			guid:         "non-existent-guid",
			index:        0,
			expectedPath: "/v3/processes/non-existent-guid/instances/0",
			statusCode:   http.StatusNotFound,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.expectedPath, r.URL.Path)
				assert.Equal(t, "DELETE", r.Method)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)

				if tt.wantErr {
					response := map[string]interface{}{
						"errors": []map[string]interface{}{
							{
								"code":   10010,
								"title":  "CF-ResourceNotFound",
								"detail": "Resource not found",
							},
						},
					}
					json.NewEncoder(w).Encode(response)
				}
			}))
			defer server.Close()

			client, err := New(&capi.Config{APIEndpoint: server.URL})
			require.NoError(t, err)

			err = client.Processes().TerminateInstance(context.Background(), tt.guid, tt.index)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "CF-ResourceNotFound")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
