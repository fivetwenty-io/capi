package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	internalhttp "github.com/fivetwenty-io/capi/internal/http"
	"github.com/fivetwenty-io/capi/pkg/capi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildpacksClient_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/buildpacks", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var request capi.BuildpackCreateRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		require.NoError(t, err)

		assert.Equal(t, "ruby_buildpack", request.Name)
		assert.NotNil(t, request.Stack)
		assert.Equal(t, "cflinuxfs4", *request.Stack)
		assert.NotNil(t, request.Position)
		assert.Equal(t, 42, *request.Position)

		now := time.Now()
		buildpack := capi.Buildpack{
			Resource: capi.Resource{
				GUID:      "buildpack-guid",
				CreatedAt: now,
				UpdatedAt: now,
			},
			Name:      request.Name,
			State:     "AWAITING_UPLOAD",
			Stack:     request.Stack,
			Position:  *request.Position,
			Lifecycle: "buildpack",
			Enabled:   true,
			Locked:    false,
			Metadata:  request.Metadata,
			Links: capi.Links{
				"self": capi.Link{
					Href: "https://api.example.org/v3/buildpacks/buildpack-guid",
				},
				"upload": capi.Link{
					Href:   "https://api.example.org/v3/buildpacks/buildpack-guid/upload",
					Method: "POST",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(buildpack)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	buildpacks := NewBuildpacksClient(client.httpClient)

	position := 42
	stack := "cflinuxfs4"
	request := &capi.BuildpackCreateRequest{
		Name:     "ruby_buildpack",
		Stack:    &stack,
		Position: &position,
		Metadata: &capi.Metadata{
			Labels: map[string]string{
				"env": "production",
			},
		},
	}

	buildpack, err := buildpacks.Create(context.Background(), request)
	require.NoError(t, err)
	assert.NotNil(t, buildpack)
	assert.Equal(t, "buildpack-guid", buildpack.GUID)
	assert.Equal(t, "ruby_buildpack", buildpack.Name)
	assert.Equal(t, "AWAITING_UPLOAD", buildpack.State)
	assert.Equal(t, 42, buildpack.Position)
}

func TestBuildpacksClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/buildpacks/buildpack-guid", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		now := time.Now()
		filename := "ruby_buildpack-v1.0.0.zip"
		stack := "cflinuxfs4"
		buildpack := capi.Buildpack{
			Resource: capi.Resource{
				GUID:      "buildpack-guid",
				CreatedAt: now,
				UpdatedAt: now,
			},
			Name:      "ruby_buildpack",
			State:     "READY",
			Filename:  &filename,
			Stack:     &stack,
			Position:  1,
			Lifecycle: "buildpack",
			Enabled:   true,
			Locked:    false,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(buildpack)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	buildpacks := NewBuildpacksClient(client.httpClient)

	buildpack, err := buildpacks.Get(context.Background(), "buildpack-guid")
	require.NoError(t, err)
	assert.NotNil(t, buildpack)
	assert.Equal(t, "buildpack-guid", buildpack.GUID)
	assert.Equal(t, "ruby_buildpack", buildpack.Name)
	assert.Equal(t, "READY", buildpack.State)
	assert.NotNil(t, buildpack.Filename)
	assert.Equal(t, "ruby_buildpack-v1.0.0.zip", *buildpack.Filename)
}

func TestBuildpacksClient_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/buildpacks", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "cflinuxfs4", r.URL.Query().Get("stacks"))
		assert.Equal(t, "ruby_buildpack,node_buildpack", r.URL.Query().Get("names"))

		now := time.Now()
		stack1 := "cflinuxfs4"
		stack2 := "cflinuxfs4"
		filename1 := "ruby_buildpack-v1.0.0.zip"
		filename2 := "node_buildpack-v2.0.0.zip"

		response := capi.ListResponse[capi.Buildpack]{
			Pagination: capi.Pagination{
				TotalResults: 2,
				TotalPages:   1,
				First:        capi.Link{Href: "/v3/buildpacks?page=1"},
				Last:         capi.Link{Href: "/v3/buildpacks?page=1"},
			},
			Resources: []capi.Buildpack{
				{
					Resource: capi.Resource{
						GUID:      "buildpack-guid-1",
						CreatedAt: now,
						UpdatedAt: now,
					},
					Name:      "ruby_buildpack",
					State:     "READY",
					Filename:  &filename1,
					Stack:     &stack1,
					Position:  1,
					Lifecycle: "buildpack",
					Enabled:   true,
					Locked:    false,
				},
				{
					Resource: capi.Resource{
						GUID:      "buildpack-guid-2",
						CreatedAt: now,
						UpdatedAt: now,
					},
					Name:      "node_buildpack",
					State:     "READY",
					Filename:  &filename2,
					Stack:     &stack2,
					Position:  2,
					Lifecycle: "buildpack",
					Enabled:   true,
					Locked:    false,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	buildpacks := NewBuildpacksClient(client.httpClient)

	params := &capi.QueryParams{
		Filters: map[string][]string{
			"stacks": {"cflinuxfs4"},
			"names":  {"ruby_buildpack", "node_buildpack"},
		},
	}

	list, err := buildpacks.List(context.Background(), params)
	require.NoError(t, err)
	assert.NotNil(t, list)
	assert.Equal(t, 2, list.Pagination.TotalResults)
	assert.Len(t, list.Resources, 2)
	assert.Equal(t, "buildpack-guid-1", list.Resources[0].GUID)
	assert.Equal(t, "ruby_buildpack", list.Resources[0].Name)
	assert.Equal(t, "buildpack-guid-2", list.Resources[1].GUID)
	assert.Equal(t, "node_buildpack", list.Resources[1].Name)
}

func TestBuildpacksClient_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/buildpacks/buildpack-guid", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var request capi.BuildpackUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		require.NoError(t, err)

		assert.NotNil(t, request.Position)
		assert.Equal(t, 5, *request.Position)
		assert.NotNil(t, request.Enabled)
		assert.Equal(t, false, *request.Enabled)

		now := time.Now()
		stack := "cflinuxfs4"
		buildpack := capi.Buildpack{
			Resource: capi.Resource{
				GUID:      "buildpack-guid",
				CreatedAt: now,
				UpdatedAt: now,
			},
			Name:      "ruby_buildpack",
			State:     "READY",
			Stack:     &stack,
			Position:  *request.Position,
			Lifecycle: "buildpack",
			Enabled:   *request.Enabled,
			Locked:    false,
			Metadata:  request.Metadata,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(buildpack)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	buildpacks := NewBuildpacksClient(client.httpClient)

	position := 5
	enabled := false
	request := &capi.BuildpackUpdateRequest{
		Position: &position,
		Enabled:  &enabled,
		Metadata: &capi.Metadata{
			Labels: map[string]string{
				"updated": "true",
			},
		},
	}

	buildpack, err := buildpacks.Update(context.Background(), "buildpack-guid", request)
	require.NoError(t, err)
	assert.NotNil(t, buildpack)
	assert.Equal(t, "buildpack-guid", buildpack.GUID)
	assert.Equal(t, 5, buildpack.Position)
	assert.Equal(t, false, buildpack.Enabled)
}

func TestBuildpacksClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/buildpacks/buildpack-guid", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		job := capi.Job{
			Resource: capi.Resource{
				GUID: "job-guid",
			},
			Operation: "buildpack.delete",
			State:     "PROCESSING",
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location", "/v3/jobs/job-guid")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(job)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	buildpacks := NewBuildpacksClient(client.httpClient)

	job, err := buildpacks.Delete(context.Background(), "buildpack-guid")
	require.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, "job-guid", job.GUID)
	assert.Equal(t, "buildpack.delete", job.Operation)
}

func TestBuildpacksClient_Upload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/buildpacks/buildpack-guid/upload", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")

		// Parse multipart form
		err := r.ParseMultipartForm(10 << 20) // 10 MB
		require.NoError(t, err)

		// Check that bits file is present
		file, header, err := r.FormFile("bits")
		require.NoError(t, err)
		defer file.Close()

		assert.NotNil(t, header)

		// Read the uploaded content
		uploadedContent, err := io.ReadAll(file)
		require.NoError(t, err)
		assert.Equal(t, "buildpack content", string(uploadedContent))

		now := time.Now()
		filename := "ruby_buildpack-v1.0.0.zip"
		stack := "cflinuxfs4"
		buildpack := capi.Buildpack{
			Resource: capi.Resource{
				GUID:      "buildpack-guid",
				CreatedAt: now,
				UpdatedAt: now,
			},
			Name:      "ruby_buildpack",
			State:     "READY",
			Filename:  &filename,
			Stack:     &stack,
			Position:  1,
			Lifecycle: "buildpack",
			Enabled:   true,
			Locked:    false,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(buildpack)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	buildpacks := NewBuildpacksClient(client.httpClient)

	// Create a reader with test content
	content := bytes.NewReader([]byte("buildpack content"))

	buildpack, err := buildpacks.Upload(context.Background(), "buildpack-guid", content)
	require.NoError(t, err)
	assert.NotNil(t, buildpack)
	assert.Equal(t, "buildpack-guid", buildpack.GUID)
	assert.Equal(t, "READY", buildpack.State)
	assert.NotNil(t, buildpack.Filename)
	assert.Equal(t, "ruby_buildpack-v1.0.0.zip", *buildpack.Filename)
}

func TestBuildpacksClient_GetNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/buildpacks/buildpack-guid", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	buildpacks := NewBuildpacksClient(client.httpClient)

	buildpack, err := buildpacks.Get(context.Background(), "buildpack-guid")
	assert.Error(t, err)
	assert.Nil(t, buildpack)
}

func TestBuildpacksClient_CreateWithCNBLifecycle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/buildpacks", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var request capi.BuildpackCreateRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		require.NoError(t, err)

		assert.Equal(t, "paketo_buildpack", request.Name)
		assert.NotNil(t, request.Lifecycle)
		assert.Equal(t, "cnb", *request.Lifecycle)

		now := time.Now()
		buildpack := capi.Buildpack{
			Resource: capi.Resource{
				GUID:      "buildpack-guid",
				CreatedAt: now,
				UpdatedAt: now,
			},
			Name:      request.Name,
			State:     "AWAITING_UPLOAD",
			Position:  1,
			Lifecycle: *request.Lifecycle,
			Enabled:   true,
			Locked:    false,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(buildpack)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	buildpacks := NewBuildpacksClient(client.httpClient)

	lifecycle := "cnb"
	request := &capi.BuildpackCreateRequest{
		Name:      "paketo_buildpack",
		Lifecycle: &lifecycle,
	}

	buildpack, err := buildpacks.Create(context.Background(), request)
	require.NoError(t, err)
	assert.NotNil(t, buildpack)
	assert.Equal(t, "buildpack-guid", buildpack.GUID)
	assert.Equal(t, "paketo_buildpack", buildpack.Name)
	assert.Equal(t, "cnb", buildpack.Lifecycle)
}
