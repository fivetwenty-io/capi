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

func TestBuildsClient_Create(t *testing.T) {
	tests := []struct {
		name         string
		request      *capi.BuildCreateRequest
		response     interface{}
		statusCode   int
		expectedPath string
		wantErr      bool
		errMessage   string
	}{
		{
			name:         "create build",
			expectedPath: "/v3/builds",
			statusCode:   http.StatusCreated,
			request: &capi.BuildCreateRequest{
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
			response: capi.Build{
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
			wantErr: false,
		},
		{
			name:         "missing package",
			expectedPath: "/v3/builds",
			statusCode:   http.StatusUnprocessableEntity,
			request:      &capi.BuildCreateRequest{},
			response: map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"code":   10008,
						"title":  "CF-UnprocessableEntity",
						"detail": "Package is required",
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

				var requestBody capi.BuildCreateRequest
				err := json.NewDecoder(r.Body).Decode(&requestBody)
				require.NoError(t, err)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client, err := New(&capi.Config{APIEndpoint: server.URL})
			require.NoError(t, err)

			build, err := client.Builds().Create(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
				assert.Nil(t, build)
			} else {
				require.NoError(t, err)
				require.NotNil(t, build)
				assert.NotEmpty(t, build.GUID)
				assert.NotEmpty(t, build.State)
			}
		})
	}
}

func TestBuildsClient_Get(t *testing.T) {
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
			guid:         "test-build-guid",
			expectedPath: "/v3/builds/test-build-guid",
			statusCode:   http.StatusOK,
			response: capi.Build{
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
			wantErr: false,
		},
		{
			name:         "build not found",
			guid:         "non-existent-guid",
			expectedPath: "/v3/builds/non-existent-guid",
			statusCode:   http.StatusNotFound,
			response: map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"code":   10010,
						"title":  "CF-ResourceNotFound",
						"detail": "Build not found",
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

			build, err := client.Builds().Get(context.Background(), tt.guid)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
				assert.Nil(t, build)
			} else {
				require.NoError(t, err)
				require.NotNil(t, build)
				assert.Equal(t, tt.guid, build.GUID)
				assert.Equal(t, "STAGED", build.State)
			}
		})
	}
}

func TestBuildsClient_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/builds", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		// Check query parameters if present
		query := r.URL.Query()
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

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/apps/app-guid/builds", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

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

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	result, err := client.Builds().ListForApp(context.Background(), "app-guid", nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.Pagination.TotalResults)
	assert.Equal(t, "build-for-app", result.Resources[0].GUID)
}

func TestBuildsClient_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/builds/test-build-guid", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var requestBody capi.BuildUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		require.NoError(t, err)

		response := capi.Build{
			Resource: capi.Resource{
				GUID:      "test-build-guid",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			State:    "STAGED",
			Metadata: requestBody.Metadata,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

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

	build, err := client.Builds().Update(context.Background(), "test-build-guid", request)
	require.NoError(t, err)
	require.NotNil(t, build)
	assert.Equal(t, "test-build-guid", build.GUID)
	assert.Equal(t, "production", build.Metadata.Labels["env"])
	assert.Equal(t, "1.0.0", build.Metadata.Annotations["version"])
}
