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

func TestPackagesClient_Create(t *testing.T) {
	tests := []struct {
		name         string
		request      *capi.PackageCreateRequest
		response     interface{}
		statusCode   int
		expectedPath string
		wantErr      bool
		errMessage   string
	}{
		{
			name:         "create bits package",
			expectedPath: "/v3/packages",
			statusCode:   http.StatusCreated,
			request: &capi.PackageCreateRequest{
				Type: "bits",
				Relationships: capi.PackageRelationships{
					App: &capi.Relationship{
						Data: &capi.RelationshipData{
							GUID: "app-guid",
						},
					},
				},
			},
			response: capi.Package{
				Resource: capi.Resource{
					GUID:      "package-guid",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					Links: capi.Links{
						"self": capi.Link{
							Href: "https://api.example.org/v3/packages/package-guid",
						},
						"upload": capi.Link{
							Href:   "https://api.example.org/v3/packages/package-guid/upload",
							Method: "POST",
						},
						"download": capi.Link{
							Href:   "https://api.example.org/v3/packages/package-guid/download",
							Method: "GET",
						},
						"app": capi.Link{
							Href: "https://api.example.org/v3/apps/app-guid",
						},
					},
				},
				Type:  "bits",
				State: "AWAITING_UPLOAD",
				Data: &capi.PackageData{
					Checksum: &capi.PackageChecksum{
						Type:  "sha256",
						Value: nil,
					},
					Error: nil,
				},
				Metadata: &capi.Metadata{
					Labels:      map[string]string{},
					Annotations: map[string]string{},
				},
				Relationships: &capi.PackageRelationships{
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
			name:         "create docker package",
			expectedPath: "/v3/packages",
			statusCode:   http.StatusCreated,
			request: &capi.PackageCreateRequest{
				Type: "docker",
				Relationships: capi.PackageRelationships{
					App: &capi.Relationship{
						Data: &capi.RelationshipData{
							GUID: "app-guid",
						},
					},
				},
				Data: &capi.PackageCreateData{
					Image:    stringPtr("nginx:latest"),
					Username: stringPtr("dockeruser"),
					Password: stringPtr("dockerpass"),
				},
			},
			response: capi.Package{
				Resource: capi.Resource{
					GUID:      "docker-package-guid",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Type:  "docker",
				State: "READY",
				Data: &capi.PackageData{
					Image:    stringPtr("nginx:latest"),
					Username: stringPtr("dockeruser"),
					Password: nil, // Password is not returned
				},
			},
			wantErr: false,
		},
		{
			name:         "missing app relationship",
			expectedPath: "/v3/packages",
			statusCode:   http.StatusUnprocessableEntity,
			request: &capi.PackageCreateRequest{
				Type:          "bits",
				Relationships: capi.PackageRelationships{},
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

				var requestBody capi.PackageCreateRequest
				err := json.NewDecoder(r.Body).Decode(&requestBody)
				require.NoError(t, err)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client, err := New(&capi.Config{APIEndpoint: server.URL})
			require.NoError(t, err)

			pkg, err := client.Packages().Create(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
				assert.Nil(t, pkg)
			} else {
				require.NoError(t, err)
				require.NotNil(t, pkg)
				assert.NotEmpty(t, pkg.GUID)
				assert.NotEmpty(t, pkg.Type)
			}
		})
	}
}

func TestPackagesClient_Get(t *testing.T) {
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
			guid:         "test-package-guid",
			expectedPath: "/v3/packages/test-package-guid",
			statusCode:   http.StatusOK,
			response: capi.Package{
				Resource: capi.Resource{
					GUID:      "test-package-guid",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Type:  "bits",
				State: "READY",
				Data: &capi.PackageData{
					Checksum: &capi.PackageChecksum{
						Type:  "sha256",
						Value: stringPtr("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"),
					},
				},
			},
			wantErr: false,
		},
		{
			name:         "package not found",
			guid:         "non-existent-guid",
			expectedPath: "/v3/packages/non-existent-guid",
			statusCode:   http.StatusNotFound,
			response: map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"code":   10010,
						"title":  "CF-ResourceNotFound",
						"detail": "Package not found",
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

			pkg, err := client.Packages().Get(context.Background(), tt.guid)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
				assert.Nil(t, pkg)
			} else {
				require.NoError(t, err)
				require.NotNil(t, pkg)
				assert.Equal(t, tt.guid, pkg.GUID)
				assert.Equal(t, "READY", pkg.State)
			}
		})
	}
}

func TestPackagesClient_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/packages", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		// Check query parameters if present
		query := r.URL.Query()
		if appGuids := query.Get("app_guids"); appGuids != "" {
			assert.Equal(t, "app-1,app-2", appGuids)
		}
		if states := query.Get("states"); states != "" {
			assert.Equal(t, "READY,FAILED", states)
		}

		response := capi.ListResponse[capi.Package]{
			Pagination: capi.Pagination{
				TotalResults: 2,
				TotalPages:   1,
				First:        capi.Link{Href: "https://api.example.org/v3/packages?page=1"},
				Last:         capi.Link{Href: "https://api.example.org/v3/packages?page=1"},
				Next:         nil,
				Previous:     nil,
			},
			Resources: []capi.Package{
				{
					Resource: capi.Resource{
						GUID:      "package-1",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Type:  "bits",
					State: "READY",
				},
				{
					Resource: capi.Resource{
						GUID:      "package-2",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Type:  "docker",
					State: "READY",
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
	result, err := client.Packages().List(context.Background(), nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 2, result.Pagination.TotalResults)
	assert.Len(t, result.Resources, 2)
	assert.Equal(t, "package-1", result.Resources[0].GUID)
	assert.Equal(t, "bits", result.Resources[0].Type)

	// Test with filters
	params := &capi.QueryParams{
		Filters: map[string][]string{
			"app_guids": {"app-1", "app-2"},
			"states":    {"READY", "FAILED"},
		},
	}
	result, err = client.Packages().List(context.Background(), params)
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestPackagesClient_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/packages/test-package-guid", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var requestBody capi.PackageUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		require.NoError(t, err)

		response := capi.Package{
			Resource: capi.Resource{
				GUID:      "test-package-guid",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Type:     "bits",
			State:    "READY",
			Metadata: requestBody.Metadata,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	request := &capi.PackageUpdateRequest{
		Metadata: &capi.Metadata{
			Labels: map[string]string{
				"env": "production",
			},
			Annotations: map[string]string{
				"version": "1.0.0",
			},
		},
	}

	pkg, err := client.Packages().Update(context.Background(), "test-package-guid", request)
	require.NoError(t, err)
	require.NotNil(t, pkg)
	assert.Equal(t, "test-package-guid", pkg.GUID)
	assert.Equal(t, "production", pkg.Metadata.Labels["env"])
	assert.Equal(t, "1.0.0", pkg.Metadata.Annotations["version"])
}

func TestPackagesClient_Delete(t *testing.T) {
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
			guid:         "test-package-guid",
			expectedPath: "/v3/packages/test-package-guid",
			statusCode:   http.StatusAccepted,
			wantErr:      false,
		},
		{
			name:         "package not found",
			guid:         "non-existent-guid",
			expectedPath: "/v3/packages/non-existent-guid",
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
					json.NewEncoder(w).Encode(map[string]interface{}{
						"errors": []map[string]interface{}{
							{
								"code":   10010,
								"title":  "CF-ResourceNotFound",
								"detail": "Package not found",
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

			err = client.Packages().Delete(context.Background(), tt.guid)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPackagesClient_Upload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/packages/test-package-guid/upload", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")

		// Read the uploaded file
		file, _, err := r.FormFile("bits")
		require.NoError(t, err)
		defer file.Close()

		uploadedContent, err := io.ReadAll(file)
		require.NoError(t, err)
		assert.Equal(t, []byte("test zip content"), uploadedContent)

		response := capi.Package{
			Resource: capi.Resource{
				GUID:      "test-package-guid",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Type:  "bits",
			State: "PROCESSING_UPLOAD",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	zipContent := []byte("test zip content")
	pkg, err := client.Packages().Upload(context.Background(), "test-package-guid", zipContent)
	require.NoError(t, err)
	require.NotNil(t, pkg)
	assert.Equal(t, "test-package-guid", pkg.GUID)
	assert.Equal(t, "PROCESSING_UPLOAD", pkg.State)
}

func TestPackagesClient_Download(t *testing.T) {
	expectedContent := []byte("test package content")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/packages/test-package-guid/download", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/zip")
		w.WriteHeader(http.StatusOK)
		w.Write(expectedContent)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	content, err := client.Packages().Download(context.Background(), "test-package-guid")
	require.NoError(t, err)
	assert.Equal(t, expectedContent, content)
}

func TestPackagesClient_Copy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/v3/packages"
		if r.URL.RawQuery != "" {
			expectedPath = expectedPath + "?" + r.URL.RawQuery
		}
		assert.Equal(t, "/v3/packages?source_guid=source-package-guid", expectedPath)
		assert.Equal(t, "POST", r.Method)

		var requestBody capi.PackageCopyRequest
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		require.NoError(t, err)
		assert.Equal(t, "target-app-guid", requestBody.Relationships.App.Data.GUID)

		response := capi.Package{
			Resource: capi.Resource{
				GUID:      "new-package-guid",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Type:  "bits",
			State: "COPYING",
			Relationships: &capi.PackageRelationships{
				App: &capi.Relationship{
					Data: &capi.RelationshipData{
						GUID: "target-app-guid",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	request := &capi.PackageCopyRequest{
		Relationships: capi.PackageRelationships{
			App: &capi.Relationship{
				Data: &capi.RelationshipData{
					GUID: "target-app-guid",
				},
			},
		},
	}

	pkg, err := client.Packages().Copy(context.Background(), "source-package-guid", request)
	require.NoError(t, err)
	require.NotNil(t, pkg)
	assert.Equal(t, "new-package-guid", pkg.GUID)
	assert.Equal(t, "COPYING", pkg.State)
	assert.Equal(t, "target-app-guid", pkg.Relationships.App.Data.GUID)
}
