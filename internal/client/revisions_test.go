package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/fivetwenty-io/capi/v3/internal/client"
	"github.com/fivetwenty-io/capi/v3/pkg/capi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRevisionsClient_Get(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/revisions/revision-guid", request.URL.Path)
		assert.Equal(t, "GET", request.Method)

		description := "Test revision"
		revision := capi.Revision{
			Resource: capi.Resource{
				GUID:      "revision-guid",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Version:     1,
			Deployable:  true,
			Description: &description,
			Droplet: capi.RevisionDropletRef{
				GUID: "droplet-guid",
			},
			Processes: map[string]capi.Process{
				"web": {
					Resource: capi.Resource{
						GUID: "process-guid",
					},
					Type:       "web",
					Instances:  2,
					MemoryInMB: 512,
					DiskInMB:   1024,
				},
			},
		}

		_ = json.NewEncoder(writer).Encode(revision)
	}))
	defer server.Close()

	client, err := New(context.Background(), &capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	revision, err := client.Revisions().Get(context.Background(), "revision-guid")
	require.NoError(t, err)
	assert.Equal(t, "revision-guid", revision.GUID)
	assert.Equal(t, 1, revision.Version)
	assert.True(t, revision.Deployable)
	assert.Equal(t, "Test revision", *revision.Description)
	assert.Equal(t, "droplet-guid", revision.Droplet.GUID)
}

func TestRevisionsClient_Update(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/revisions/revision-guid", request.URL.Path)
		assert.Equal(t, "PATCH", request.Method)

		var req capi.RevisionUpdateRequest

		_ = json.NewDecoder(request.Body).Decode(&req)
		assert.NotNil(t, req.Metadata)
		assert.Equal(t, "value1", req.Metadata.Labels["key1"])

		revision := capi.Revision{
			Resource: capi.Resource{GUID: "revision-guid"},
			Version:  1,
			Metadata: req.Metadata,
		}

		_ = json.NewEncoder(writer).Encode(revision)
	}))
	defer server.Close()

	client, err := New(context.Background(), &capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	revision, err := client.Revisions().Update(context.Background(), "revision-guid", &capi.RevisionUpdateRequest{
		Metadata: &capi.Metadata{
			Labels: map[string]string{
				"key1": "value1",
			},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "revision-guid", revision.GUID)
	assert.Equal(t, "value1", revision.Metadata.Labels["key1"])
}

func TestRevisionsClient_GetEnvironmentVariables(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/revisions/revision-guid/environment_variables", request.URL.Path)
		assert.Equal(t, "GET", request.Method)

		envVars := map[string]interface{}{
			"DATABASE_URL": "postgres://localhost/myapp",
			"API_KEY":      "secret-key",
			"DEBUG":        true,
		}

		response := map[string]interface{}{
			"var": envVars,
		}

		_ = json.NewEncoder(writer).Encode(response)
	}))
	defer server.Close()

	client, err := New(context.Background(), &capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	envVars, err := client.Revisions().GetEnvironmentVariables(context.Background(), "revision-guid")
	require.NoError(t, err)
	assert.Equal(t, "postgres://localhost/myapp", envVars["DATABASE_URL"])
	assert.Equal(t, "secret-key", envVars["API_KEY"])
	assert.Equal(t, true, envVars["DEBUG"])
}
