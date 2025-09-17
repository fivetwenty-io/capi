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
func TestBuildsClient_Create(t *testing.T) {
	t.Parallel()

	tests := []TestCreateOperation[capi.BuildCreateRequest, capi.Build]{
		{
			Name:         "create build",
			ExpectedPath: "/v3/builds",
			StatusCode:   http.StatusCreated,
			Request: &capi.BuildCreateRequest{
				Package: &capi.BuildPackageRef{
					GUID: "package-guid",
				},
				StagingMemoryInMB: intPtr(1024),
				StagingDiskInMB:   intPtr(1024),
				Metadata: &capi.Metadata{
					Labels: map[string]string{
						"env": "staging",
					},
				},
			},
			Response: &capi.Build{
				Resource: capi.Resource{
					GUID:      "build-guid",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					Links: capi.Links{
						"self": capi.Link{
							Href: "https://api.example.org/v3/builds/build-guid",
						},
						"app": capi.Link{
							Href: "https://api.example.org/v3/apps/app-guid",
						},
					},
				},
				State:             "STAGING",
				StagingMemoryInMB: 1024,
				StagingDiskInMB:   1024,
				Package: &capi.BuildPackageRef{
					GUID: "package-guid",
				},
				Droplet: nil,
				CreatedBy: &capi.UserRef{
					GUID:  "user-guid",
					Name:  "bill",
					Email: "bill@example.com",
				},
				Lifecycle: &capi.Lifecycle{
					Type: "buildpack",
					Data: map[string]interface{}{
						"buildpacks": []string{"ruby_buildpack"},
						"stack":      "cflinuxfs4",
					},
				},
				Relationships: &capi.BuildRelationships{
					App: &capi.Relationship{
						Data: &capi.RelationshipData{
							GUID: "app-guid",
						},
					},
				},
				Metadata: &capi.Metadata{
					Labels: map[string]string{
						"env": "staging",
					},
				},
			},
			WantErr: false,
		},
		{
			Name:         "missing package",
			ExpectedPath: "/v3/builds",
			StatusCode:   http.StatusUnprocessableEntity,
			Request:      &capi.BuildCreateRequest{},
			Response: map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"code":   10008,
						"title":  "CF-UnprocessableEntity",
						"detail": "The request is semantically invalid: Missing required field 'package'",
					},
				},
			},
			WantErr:    true,
			ErrMessage: "CF-UnprocessableEntity",
		},
	}

	RunCreateTests(t, tests,
		func(c *Client) func(context.Context, *capi.BuildCreateRequest) (*capi.Build, error) {
			return c.Builds().Create
		},
		func(request *http.Request) (*capi.BuildCreateRequest, error) {
			var requestBody capi.BuildCreateRequest

			err := json.NewDecoder(request.Body).Decode(&requestBody)
			if err != nil {
				return &requestBody, fmt.Errorf("failed to decode request body: %w", err)
			}

			return &requestBody, nil
		},
	)
}

func TestBuildsClient_Get(t *testing.T) {
	t.Parallel()

	tests := []TestGetOperation[capi.Build]{
		{
			Name:         "successful get",
			GUID:         "test-build-guid",
			ExpectedPath: "/v3/builds/test-build-guid",
			StatusCode:   http.StatusOK,
			Response: &capi.Build{
				Resource: capi.Resource{
					GUID:      "test-build-guid",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				State:             "STAGED",
				StagingMemoryInMB: 1024,
				StagingDiskInMB:   1024,
				Package: &capi.BuildPackageRef{
					GUID: "package-guid",
				},
				Droplet: &capi.BuildDropletRef{
					GUID: "droplet-guid",
				},
				Lifecycle: &capi.Lifecycle{
					Type: "buildpack",
					Data: map[string]interface{}{},
				},
			},
			WantErr: false,
		},
		{
			Name:         "build not found",
			GUID:         "non-existent-guid",
			ExpectedPath: "/v3/builds/non-existent-guid",
			StatusCode:   http.StatusNotFound,
			Response: &capi.Build{
				Resource: capi.Resource{
					GUID:      "test-build-guid",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				State:             "STAGED",
				StagingMemoryInMB: 1024,
				StagingDiskInMB:   1024,
			},
			WantErr:    true,
			ErrMessage: "CF-ResourceNotFound",
		},
	}

	RunGetTests(t, tests, func(c *Client) func(context.Context, string) (*capi.Build, error) {
		return c.Builds().Get
	})
}

//nolint:funlen // Test functions can be longer for comprehensive testing
func TestBuildsClient_List(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/builds", request.URL.Path)
		assert.Equal(t, "GET", request.Method)

		// Check query parameters if present
		query := request.URL.Query()
		if states := query.Get("states"); states != "" {
			assert.Equal(t, "STAGING,STAGED", states)
		}

		if packageGuids := query.Get("package_guids"); packageGuids != "" {
			assert.Equal(t, "package-1,package-2", packageGuids)
		}

		response := capi.ListResponse[capi.Build]{
			Pagination: capi.Pagination{
				TotalResults: 2,
				TotalPages:   1,
				First:        capi.Link{Href: "https://api.example.org/v3/builds?page=1"},
				Last:         capi.Link{Href: "https://api.example.org/v3/builds?page=1"},
				Next:         nil,
				Previous:     nil,
			},
			Resources: []capi.Build{
				{
					Resource: capi.Resource{
						GUID:      "build-1",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					State:             "STAGING",
					StagingMemoryInMB: 1024,
					Package: &capi.BuildPackageRef{
						GUID: "package-1",
					},
				},
				{
					Resource: capi.Resource{
						GUID:      "build-2",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					State:             "STAGED",
					StagingMemoryInMB: 2048,
					Package: &capi.BuildPackageRef{
						GUID: "package-2",
					},
					Droplet: &capi.BuildDropletRef{
						GUID: "droplet-2",
					},
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
	result, err := client.Builds().List(context.Background(), nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 2, result.Pagination.TotalResults)
	assert.Len(t, result.Resources, 2)
	assert.Equal(t, "build-1", result.Resources[0].GUID)
	assert.Equal(t, "STAGING", result.Resources[0].State)

	// Test with filters
	params := &capi.QueryParams{
		Filters: map[string][]string{
			"states":        {"STAGING", "STAGED"},
			"package_guids": {"package-1", "package-2"},
		},
	}
	result, err = client.Builds().List(context.Background(), params)
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestBuildsClient_ListForApp(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/apps/app-guid/builds", request.URL.Path)
		assert.Equal(t, "GET", request.Method)

		response := capi.ListResponse[capi.Build]{
			Pagination: capi.Pagination{
				TotalResults: 1,
				TotalPages:   1,
			},
			Resources: []capi.Build{
				{
					Resource: capi.Resource{
						GUID: "build-for-app",
					},
					State:             "STAGED",
					StagingMemoryInMB: 1024,
					Relationships: &capi.BuildRelationships{
						App: &capi.Relationship{
							Data: &capi.RelationshipData{
								GUID: "app-guid",
							},
						},
					},
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

	result, err := client.Builds().ListForApp(context.Background(), "app-guid", nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.Pagination.TotalResults)
	assert.Equal(t, "build-for-app", result.Resources[0].GUID)
}

//nolint:dupl // Acceptable duplication - each test validates different endpoints with different request/response types
func TestBuildsClient_Update(t *testing.T) {
	t.Parallel()

	request := &capi.BuildUpdateRequest{
		Metadata: &capi.Metadata{
			Labels: map[string]string{
				"env": "production",
			},
			Annotations: map[string]string{
				"version": "1.0.0",
			},
		},
	}

	response := &capi.Build{
		Resource: capi.Resource{
			GUID:      "test-build-guid",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		State: "STAGED",
		Metadata: &capi.Metadata{
			Labels: map[string]string{
				"env": "production",
			},
			Annotations: map[string]string{
				"version": "1.0.0",
			},
		},
	}

	RunStandardUpdateTest(t, "build", "test-build-guid", "/v3/builds/test-build-guid", request, response,
		func(c *Client) func(context.Context, string, *capi.BuildUpdateRequest) (*capi.Build, error) {
			return c.Builds().Update
		})
}
