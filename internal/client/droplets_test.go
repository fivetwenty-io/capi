package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fivetwenty-io/capi/v3/pkg/capi"
)

func TestDropletsClient_Create(t *testing.T) {
	tests := []struct {
		name         string
		request      *capi.DropletCreateRequest
		response     interface{}
		statusCode   int
		expectedPath string
		wantErr      bool
		errMessage   string
	}{
		{
			name:         "create droplet",
			expectedPath: "/v3/droplets",
			statusCode:   http.StatusCreated,
			request: &capi.DropletCreateRequest{
				Relationships: capi.DropletRelationships{
					App: &capi.Relationship{
						Data: &capi.RelationshipData{
							GUID: "app-guid",
						},
					},
				},
				ProcessTypes: map[string]string{
					"web":  "bundle exec rackup config.ru -p $PORT",
					"rake": "bundle exec rake",
				},
			},
			response: capi.Droplet{
				Resource: capi.Resource{
					GUID:      "droplet-guid",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					Links: capi.Links{
						"self": capi.Link{
							Href: "https://api.example.org/v3/droplets/droplet-guid",
						},
						"package": capi.Link{
							Href: "https://api.example.org/v3/packages/package-guid",
						},
						"app": capi.Link{
							Href: "https://api.example.org/v3/apps/app-guid",
						},
						"download": capi.Link{
							Href: "https://api.example.org/v3/droplets/droplet-guid/download",
						},
					},
				},
				State: "AWAITING_UPLOAD",
				Error: nil,
				Lifecycle: capi.Lifecycle{
					Type: "buildpack",
					Data: map[string]interface{}{},
				},
				ExecutionMetadata: "",
				ProcessTypes: map[string]string{
					"web":  "bundle exec rackup config.ru -p $PORT",
					"rake": "bundle exec rake",
				},
				Metadata: &capi.Metadata{
					Labels:      map[string]string{},
					Annotations: map[string]string{},
				},
				Relationships: &capi.DropletRelationships{
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
			name:         "missing app relationship",
			expectedPath: "/v3/droplets",
			statusCode:   http.StatusUnprocessableEntity,
			request: &capi.DropletCreateRequest{
				Relationships: capi.DropletRelationships{},
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

				var requestBody capi.DropletCreateRequest
				err := json.NewDecoder(r.Body).Decode(&requestBody)
				require.NoError(t, err)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				_ = json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client, err := New(&capi.Config{APIEndpoint: server.URL})
			require.NoError(t, err)

			droplet, err := client.Droplets().Create(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
				assert.Nil(t, droplet)
			} else {
				require.NoError(t, err)
				require.NotNil(t, droplet)
				assert.NotEmpty(t, droplet.GUID)
				assert.NotEmpty(t, droplet.State)
			}
		})
	}
}

func TestDropletsClient_Get(t *testing.T) {
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
			guid:         "test-droplet-guid",
			expectedPath: "/v3/droplets/test-droplet-guid",
			statusCode:   http.StatusOK,
			response: capi.Droplet{
				Resource: capi.Resource{
					GUID:      "test-droplet-guid",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				State: "STAGED",
				Error: nil,
				Lifecycle: capi.Lifecycle{
					Type: "buildpack",
					Data: map[string]interface{}{},
				},
				ProcessTypes: map[string]string{
					"web": "bundle exec rackup config.ru -p $PORT",
				},
				Checksum: &capi.DropletChecksum{
					Type:  "sha256",
					Value: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				},
				Buildpacks: []capi.DetectedBuildpack{
					{
						Name:          "ruby_buildpack",
						DetectOutput:  "ruby 2.7.2",
						Version:       stringPtr("1.8.0"),
						BuildpackName: stringPtr("ruby"),
					},
				},
				Stack: stringPtr("cflinuxfs4"),
			},
			wantErr: false,
		},
		{
			name:         "droplet not found",
			guid:         "non-existent-guid",
			expectedPath: "/v3/droplets/non-existent-guid",
			statusCode:   http.StatusNotFound,
			response: map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"code":   10010,
						"title":  "CF-ResourceNotFound",
						"detail": "Droplet not found",
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

			droplet, err := client.Droplets().Get(context.Background(), tt.guid)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
				assert.Nil(t, droplet)
			} else {
				require.NoError(t, err)
				require.NotNil(t, droplet)
				assert.Equal(t, tt.guid, droplet.GUID)
				assert.Equal(t, "STAGED", droplet.State)
			}
		})
	}
}

func TestDropletsClient_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/droplets", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		// Check query parameters if present
		query := r.URL.Query()
		if appGuids := query.Get("app_guids"); appGuids != "" {
			assert.Equal(t, "app-1,app-2", appGuids)
		}
		if states := query.Get("states"); states != "" {
			assert.Equal(t, "STAGED,FAILED", states)
		}

		response := capi.ListResponse[capi.Droplet]{
			Pagination: capi.Pagination{
				TotalResults: 2,
				TotalPages:   1,
				First:        capi.Link{Href: "https://api.example.org/v3/droplets?page=1"},
				Last:         capi.Link{Href: "https://api.example.org/v3/droplets?page=1"},
				Next:         nil,
				Previous:     nil,
			},
			Resources: []capi.Droplet{
				{
					Resource: capi.Resource{
						GUID:      "droplet-1",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					State: "STAGED",
					Lifecycle: capi.Lifecycle{
						Type: "buildpack",
						Data: map[string]interface{}{},
					},
				},
				{
					Resource: capi.Resource{
						GUID:      "droplet-2",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					State: "STAGED",
					Lifecycle: capi.Lifecycle{
						Type: "docker",
						Data: map[string]interface{}{},
					},
					Image: stringPtr("nginx:latest"),
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
	result, err := client.Droplets().List(context.Background(), nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 2, result.Pagination.TotalResults)
	assert.Len(t, result.Resources, 2)
	assert.Equal(t, "droplet-1", result.Resources[0].GUID)
	assert.Equal(t, "buildpack", result.Resources[0].Lifecycle.Type)

	// Test with filters
	params := &capi.QueryParams{
		Filters: map[string][]string{
			"app_guids": {"app-1", "app-2"},
			"states":    {"STAGED", "FAILED"},
		},
	}
	result, err = client.Droplets().List(context.Background(), params)
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestDropletsClient_ListForApp(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/apps/app-guid/droplets", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		response := capi.ListResponse[capi.Droplet]{
			Pagination: capi.Pagination{
				TotalResults: 1,
				TotalPages:   1,
			},
			Resources: []capi.Droplet{
				{
					Resource: capi.Resource{
						GUID: "droplet-for-app",
					},
					State: "STAGED",
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

	result, err := client.Droplets().ListForApp(context.Background(), "app-guid", nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.Pagination.TotalResults)
	assert.Equal(t, "droplet-for-app", result.Resources[0].GUID)
}

func TestDropletsClient_ListForPackage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/packages/package-guid/droplets", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		response := capi.ListResponse[capi.Droplet]{
			Pagination: capi.Pagination{
				TotalResults: 1,
				TotalPages:   1,
			},
			Resources: []capi.Droplet{
				{
					Resource: capi.Resource{
						GUID: "droplet-for-package",
					},
					State: "STAGED",
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

	result, err := client.Droplets().ListForPackage(context.Background(), "package-guid", nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.Pagination.TotalResults)
	assert.Equal(t, "droplet-for-package", result.Resources[0].GUID)
}

func TestDropletsClient_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/droplets/test-droplet-guid", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var requestBody capi.DropletUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		require.NoError(t, err)

		response := capi.Droplet{
			Resource: capi.Resource{
				GUID:      "test-droplet-guid",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			State:    "STAGED",
			Metadata: requestBody.Metadata,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	request := &capi.DropletUpdateRequest{
		Metadata: &capi.Metadata{
			Labels: map[string]string{
				"env": "production",
			},
			Annotations: map[string]string{
				"version": "1.0.0",
			},
		},
	}

	droplet, err := client.Droplets().Update(context.Background(), "test-droplet-guid", request)
	require.NoError(t, err)
	require.NotNil(t, droplet)
	assert.Equal(t, "test-droplet-guid", droplet.GUID)
	assert.Equal(t, "production", droplet.Metadata.Labels["env"])
	assert.Equal(t, "1.0.0", droplet.Metadata.Annotations["version"])
}

func TestDropletsClient_Delete(t *testing.T) {
	tests := []struct {
		name         string
		guid         string
		statusCode   int
		expectedPath string
		wantErr      bool
		errMessage   string
	}{
		{
			name:         "successful delete",
			guid:         "test-droplet-guid",
			expectedPath: "/v3/droplets/test-droplet-guid",
			statusCode:   http.StatusAccepted,
			wantErr:      false,
		},
		{
			name:         "droplet not found",
			guid:         "non-existent-guid",
			expectedPath: "/v3/droplets/non-existent-guid",
			statusCode:   http.StatusNotFound,
			wantErr:      true,
			errMessage:   "CF-ResourceNotFound",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.expectedPath, r.URL.Path)
				assert.Equal(t, "DELETE", r.Method)

				if tt.statusCode == http.StatusNotFound {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tt.statusCode)
					_ = json.NewEncoder(w).Encode(map[string]interface{}{
						"errors": []map[string]interface{}{
							{
								"code":   10010,
								"title":  "CF-ResourceNotFound",
								"detail": "Droplet not found",
							},
						},
					})
				} else {
					w.WriteHeader(tt.statusCode)
				}
			}))
			defer server.Close()

			client, err := New(&capi.Config{APIEndpoint: server.URL})
			require.NoError(t, err)

			err = client.Droplets().Delete(context.Background(), tt.guid)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDropletsClient_Copy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/v3/droplets"
		if r.URL.RawQuery != "" {
			expectedPath = expectedPath + "?" + r.URL.RawQuery
		}
		assert.Equal(t, "/v3/droplets?source_guid=source-droplet-guid", expectedPath)
		assert.Equal(t, "POST", r.Method)

		var requestBody capi.DropletCopyRequest
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		require.NoError(t, err)
		assert.Equal(t, "target-app-guid", requestBody.Relationships.App.Data.GUID)

		response := capi.Droplet{
			Resource: capi.Resource{
				GUID:      "new-droplet-guid",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			State: "COPYING",
			Relationships: &capi.DropletRelationships{
				App: &capi.Relationship{
					Data: &capi.RelationshipData{
						GUID: "target-app-guid",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	request := &capi.DropletCopyRequest{
		Relationships: capi.DropletRelationships{
			App: &capi.Relationship{
				Data: &capi.RelationshipData{
					GUID: "target-app-guid",
				},
			},
		},
	}

	droplet, err := client.Droplets().Copy(context.Background(), "source-droplet-guid", request)
	require.NoError(t, err)
	require.NotNil(t, droplet)
	assert.Equal(t, "new-droplet-guid", droplet.GUID)
	assert.Equal(t, "COPYING", droplet.State)
	assert.Equal(t, "target-app-guid", droplet.Relationships.App.Data.GUID)
}

func TestDropletsClient_Download(t *testing.T) {
	expectedContent := []byte("test droplet content")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/droplets/test-droplet-guid/download", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/zip")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(expectedContent)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	content, err := client.Droplets().Download(context.Background(), "test-droplet-guid")
	require.NoError(t, err)
	assert.Equal(t, expectedContent, content)
}

func TestDropletsClient_Upload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/droplets/test-droplet-guid/upload", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")

		// Read the uploaded file
		file, _, err := r.FormFile("bits")
		require.NoError(t, err)
		defer func() {
			if err := file.Close(); err != nil {
				t.Logf("Warning: failed to close file: %v", err)
			}
		}()

		uploadedContent, err := io.ReadAll(file)
		require.NoError(t, err)
		assert.Equal(t, []byte("test droplet content"), uploadedContent)

		response := capi.Droplet{
			Resource: capi.Resource{
				GUID:      "test-droplet-guid",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			State: "PROCESSING_UPLOAD",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	dropletContent := []byte("test droplet content")
	droplet, err := client.Droplets().Upload(context.Background(), "test-droplet-guid", dropletContent)
	require.NoError(t, err)
	require.NotNil(t, droplet)
	assert.Equal(t, "test-droplet-guid", droplet.GUID)
	assert.Equal(t, "PROCESSING_UPLOAD", droplet.State)
}
