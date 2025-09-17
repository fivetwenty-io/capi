package client_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/fivetwenty-io/capi/v3/internal/client"
	"github.com/fivetwenty-io/capi/v3/pkg/capi"
)

//nolint:funlen // Test functions can be longer for comprehensive testing
func TestDeploymentsClient_Create(t *testing.T) {
	t.Parallel()

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
			response: &capi.Deployment{
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
				Strategy: StringPtr("canary"),
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

	runCreateTestsForDeployments(t, tests)
}

func TestDeploymentsClient_Get(t *testing.T) {
	t.Parallel()

	tests := []TestGetOperation[capi.Deployment]{
		{
			Name:         "successful get",
			GUID:         "test-deployment-guid",
			ExpectedPath: "/v3/deployments/test-deployment-guid",
			StatusCode:   http.StatusOK,
			Response: &capi.Deployment{
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
			WantErr: false,
		},
		{
			Name:         "deployment not found",
			GUID:         "non-existent-guid",
			ExpectedPath: "/v3/deployments/non-existent-guid",
			StatusCode:   http.StatusNotFound,
			Response: &capi.Deployment{
				Resource: capi.Resource{
					GUID:      "test-deployment-guid",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				State: "DEPLOYED",
			},
			WantErr:    true,
			ErrMessage: "CF-ResourceNotFound",
		},
	}

	RunGetTests(t, tests, func(c *Client) func(context.Context, string) (*capi.Deployment, error) {
		return c.Deployments().Get
	})
}

//nolint:funlen // Test functions can be longer for comprehensive testing
func TestDeploymentsClient_List(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/deployments", request.URL.Path)
		assert.Equal(t, "GET", request.Method)

		// Check query parameters if present
		query := request.URL.Query()
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

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(writer).Encode(response)
	}))
	defer server.Close()

	client, err := New(context.Background(), &capi.Config{APIEndpoint: server.URL})
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

//nolint:dupl // Acceptable duplication - each test validates different endpoints with different request/response types
func TestDeploymentsClient_Update(t *testing.T) {
	t.Parallel()

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

	response := &capi.Deployment{
		Resource: capi.Resource{
			GUID:      "test-deployment-guid",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		State: "DEPLOYING",
		Metadata: &capi.Metadata{
			Labels: map[string]string{
				"version": "v1.0.1",
			},
			Annotations: map[string]string{
				"note": "Updated deployment",
			},
		},
	}

	RunStandardUpdateTest(t, "deployment", "test-deployment-guid", "/v3/deployments/test-deployment-guid", request, response,
		func(c *Client) func(context.Context, string, *capi.DeploymentUpdateRequest) (*capi.Deployment, error) {
			return c.Deployments().Update
		})
}

func TestDeploymentsClient_Cancel(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/deployments/test-deployment-guid/actions/cancel", request.URL.Path)
		assert.Equal(t, "POST", request.Method)
		writer.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c, err := New(context.Background(), &capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	err = c.Deployments().Cancel(context.Background(), "test-deployment-guid")
	require.NoError(t, err)
}

func TestDeploymentsClient_Continue(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/deployments/test-deployment-guid/actions/continue", request.URL.Path)
		assert.Equal(t, "POST", request.Method)
		writer.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c, err := New(context.Background(), &capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	err = c.Deployments().Continue(context.Background(), "test-deployment-guid")
	require.NoError(t, err)
}

// runCreateTestsForDeployments runs deployment create tests.
func runCreateTestsForDeployments(t *testing.T, tests []struct {
	name         string
	request      *capi.DeploymentCreateRequest
	response     interface{}
	statusCode   int
	expectedPath string
	wantErr      bool
	errMessage   string
}) {
	t.Helper()

	for _, testCase := range tests {
		RunCreateTestWithValidation(t, testCase.name, testCase.expectedPath, testCase.statusCode, testCase.response, testCase.wantErr, testCase.errMessage, func(c *Client) error {
			deployment, err := c.Deployments().Create(context.Background(), testCase.request)
			if err == nil {
				assert.NotEmpty(t, deployment.GUID)
				assert.NotEmpty(t, deployment.State)
			}

			if err != nil {
				return fmt.Errorf("failed to create deployment: %w", err)
			}

			return nil
		})
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}
