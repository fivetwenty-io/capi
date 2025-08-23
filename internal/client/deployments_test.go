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

func TestDeploymentsClient_Create(t *testing.T) {
	tests := []struct {
		name         string
		request      *capi.DeploymentCreateRequest
		response     interface{}
		statusCode   int
		expectedPath string
		wantErr      bool
		errMessage   string
	}{
		{
			name:         "create deployment with droplet",
			expectedPath: "/v3/deployments",
			statusCode:   http.StatusCreated,
			request: &capi.DeploymentCreateRequest{
				Droplet: &capi.DeploymentDropletRef{
					GUID: "droplet-guid",
				},
				Relationships: capi.DeploymentRelationships{
					App: &capi.Relationship{
						Data: &capi.RelationshipData{
							GUID: "app-guid",
						},
					},
				},
				Metadata: &capi.Metadata{
					Labels: map[string]string{
						"version": "v1.0.0",
					},
				},
			},
			response: capi.Deployment{
				Resource: capi.Resource{
					GUID:      "deployment-guid",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					Links: capi.Links{
						"self": capi.Link{
							Href: "https://api.example.org/v3/deployments/deployment-guid",
						},
						"app": capi.Link{
							Href: "https://api.example.org/v3/apps/app-guid",
						},
						"cancel": capi.Link{
							Href:   "https://api.example.org/v3/deployments/deployment-guid/actions/cancel",
							Method: "POST",
						},
					},
				},
				State: "DEPLOYING",
				Status: capi.DeploymentStatus{
					Value:  "ACTIVE",
					Reason: "DEPLOYING",
				},
				Strategy: "rolling",
				Droplet: &capi.DeploymentDropletRef{
					GUID: "droplet-guid",
				},
				PreviousDroplet: &capi.DeploymentDropletRef{
					GUID: "previous-droplet-guid",
				},
				NewProcesses: []capi.DeploymentProcess{
					{
						GUID: "process-guid",
						Type: "web",
					},
				},
				Relationships: &capi.DeploymentRelationships{
					App: &capi.Relationship{
						Data: &capi.RelationshipData{
							GUID: "app-guid",
						},
					},
				},
				Metadata: &capi.Metadata{
					Labels: map[string]string{
						"version": "v1.0.0",
					},
				},
			},
			wantErr: false,
		},
		{
			name:         "create deployment with revision",
			expectedPath: "/v3/deployments",
			statusCode:   http.StatusCreated,
			request: &capi.DeploymentCreateRequest{
				Revision: &capi.DeploymentRevisionRef{
					GUID:    "revision-guid",
					Version: 42,
				},
				Strategy: stringPtr("canary"),
				Options: &capi.DeploymentOptions{
					MaxInFlight: intPtr(2),
					Canary: &capi.DeploymentCanaryOptions{
						Steps: []capi.DeploymentCanaryStep{
							{Instances: 1, WaitTime: 60},
							{Instances: 5, WaitTime: 120},
						},
					},
				},
				Relationships: capi.DeploymentRelationships{
					App: &capi.Relationship{
						Data: &capi.RelationshipData{
							GUID: "app-guid",
						},
					},
				},
			},
			response: capi.Deployment{
				Resource: capi.Resource{
					GUID:      "deployment-guid",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				State: "DEPLOYING",
				Status: capi.DeploymentStatus{
					Value:  "ACTIVE",
					Reason: "DEPLOYING",
					Canary: &capi.DeploymentCanaryStatus{
						Steps: capi.DeploymentCanarySteps{
							Current: 1,
							Total:   2,
						},
					},
				},
				Strategy: "canary",
				Revision: &capi.DeploymentRevisionRef{
					GUID:    "revision-guid",
					Version: 42,
				},
			},
			wantErr: false,
		},
		{
			name:         "missing app relationship",
			expectedPath: "/v3/deployments",
			statusCode:   http.StatusUnprocessableEntity,
			request: &capi.DeploymentCreateRequest{
				Droplet: &capi.DeploymentDropletRef{
					GUID: "droplet-guid",
				},
				Relationships: capi.DeploymentRelationships{},
			},
			response: map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"code":   10008,
						"title":  "CF-UnprocessableEntity",
						"detail": "App relationship is required",
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

				var requestBody capi.DeploymentCreateRequest
				err := json.NewDecoder(r.Body).Decode(&requestBody)
				require.NoError(t, err)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client, err := New(&capi.Config{APIEndpoint: server.URL})
			require.NoError(t, err)

			deployment, err := client.Deployments().Create(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
				assert.Nil(t, deployment)
			} else {
				require.NoError(t, err)
				require.NotNil(t, deployment)
				assert.NotEmpty(t, deployment.GUID)
				assert.NotEmpty(t, deployment.State)
			}
		})
	}
}

func TestDeploymentsClient_Get(t *testing.T) {
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
			guid:         "test-deployment-guid",
			expectedPath: "/v3/deployments/test-deployment-guid",
			statusCode:   http.StatusOK,
			response: capi.Deployment{
				Resource: capi.Resource{
					GUID:      "test-deployment-guid",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				State: "DEPLOYED",
				Status: capi.DeploymentStatus{
					Value:  "FINALIZED",
					Reason: "DEPLOYED",
					Details: &capi.DeploymentStatusDetails{
						LastHealthyAt: timePtr(time.Now()),
					},
				},
				Strategy: "rolling",
				Droplet: &capi.DeploymentDropletRef{
					GUID: "droplet-guid",
				},
			},
			wantErr: false,
		},
		{
			name:         "deployment not found",
			guid:         "non-existent-guid",
			expectedPath: "/v3/deployments/non-existent-guid",
			statusCode:   http.StatusNotFound,
			response: map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"code":   10010,
						"title":  "CF-ResourceNotFound",
						"detail": "Deployment not found",
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

			deployment, err := client.Deployments().Get(context.Background(), tt.guid)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
				assert.Nil(t, deployment)
			} else {
				require.NoError(t, err)
				require.NotNil(t, deployment)
				assert.Equal(t, tt.guid, deployment.GUID)
				assert.Equal(t, "DEPLOYED", deployment.State)
			}
		})
	}
}

func TestDeploymentsClient_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/deployments", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		// Check query parameters if present
		query := r.URL.Query()
		if appGuids := query.Get("app_guids"); appGuids != "" {
			assert.Equal(t, "app-1,app-2", appGuids)
		}
		if states := query.Get("states"); states != "" {
			assert.Equal(t, "DEPLOYING,DEPLOYED", states)
		}
		if statusReasons := query.Get("status_reasons"); statusReasons != "" {
			assert.Equal(t, "DEPLOYING,DEPLOYED", statusReasons)
		}
		if statusValues := query.Get("status_values"); statusValues != "" {
			assert.Equal(t, "ACTIVE,FINALIZED", statusValues)
		}

		response := capi.ListResponse[capi.Deployment]{
			Pagination: capi.Pagination{
				TotalResults: 2,
				TotalPages:   1,
				First:        capi.Link{Href: "https://api.example.org/v3/deployments?page=1"},
				Last:         capi.Link{Href: "https://api.example.org/v3/deployments?page=1"},
				Next:         nil,
				Previous:     nil,
			},
			Resources: []capi.Deployment{
				{
					Resource: capi.Resource{
						GUID:      "deployment-1",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					State: "DEPLOYING",
					Status: capi.DeploymentStatus{
						Value:  "ACTIVE",
						Reason: "DEPLOYING",
					},
					Strategy: "rolling",
				},
				{
					Resource: capi.Resource{
						GUID:      "deployment-2",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					State: "DEPLOYED",
					Status: capi.DeploymentStatus{
						Value:  "FINALIZED",
						Reason: "DEPLOYED",
					},
					Strategy: "rolling",
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
	result, err := client.Deployments().List(context.Background(), nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 2, result.Pagination.TotalResults)
	assert.Len(t, result.Resources, 2)
	assert.Equal(t, "deployment-1", result.Resources[0].GUID)
	assert.Equal(t, "DEPLOYING", result.Resources[0].State)

	// Test with filters
	params := &capi.QueryParams{
		Filters: map[string][]string{
			"app_guids":      {"app-1", "app-2"},
			"states":         {"DEPLOYING", "DEPLOYED"},
			"status_reasons": {"DEPLOYING", "DEPLOYED"},
			"status_values":  {"ACTIVE", "FINALIZED"},
		},
	}
	result, err = client.Deployments().List(context.Background(), params)
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestDeploymentsClient_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/deployments/test-deployment-guid", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var requestBody capi.DeploymentUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		require.NoError(t, err)

		response := capi.Deployment{
			Resource: capi.Resource{
				GUID:      "test-deployment-guid",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			State:    "DEPLOYING",
			Metadata: requestBody.Metadata,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	request := &capi.DeploymentUpdateRequest{
		Metadata: &capi.Metadata{
			Labels: map[string]string{
				"version": "v1.0.1",
			},
			Annotations: map[string]string{
				"note": "Updated deployment",
			},
		},
	}

	deployment, err := client.Deployments().Update(context.Background(), "test-deployment-guid", request)
	require.NoError(t, err)
	require.NotNil(t, deployment)
	assert.Equal(t, "test-deployment-guid", deployment.GUID)
	assert.Equal(t, "v1.0.1", deployment.Metadata.Labels["version"])
	assert.Equal(t, "Updated deployment", deployment.Metadata.Annotations["note"])
}

func TestDeploymentsClient_Cancel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/deployments/test-deployment-guid/actions/cancel", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	err = client.Deployments().Cancel(context.Background(), "test-deployment-guid")
	require.NoError(t, err)
}

func TestDeploymentsClient_Continue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/deployments/test-deployment-guid/actions/continue", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	err = client.Deployments().Continue(context.Background(), "test-deployment-guid")
	require.NoError(t, err)
}

func timePtr(t time.Time) *time.Time {
	return &t
}
