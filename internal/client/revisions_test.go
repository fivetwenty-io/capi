package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fivetwenty-io/capi-client/pkg/capi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRevisionsClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/revisions/revision-guid", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

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

		json.NewEncoder(w).Encode(revision)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/revisions/revision-guid", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var req capi.RevisionUpdateRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.NotNil(t, req.Metadata)
		assert.Equal(t, "value1", req.Metadata.Labels["key1"])

		revision := capi.Revision{
			Resource: capi.Resource{GUID: "revision-guid"},
			Version:  1,
			Metadata: req.Metadata,
		}

		json.NewEncoder(w).Encode(revision)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/revisions/revision-guid/environment_variables", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		envVars := map[string]interface{}{
			"DATABASE_URL": "postgres://localhost/myapp",
			"API_KEY":      "secret-key",
			"DEBUG":        true,
		}

		response := map[string]interface{}{
			"var": envVars,
		}

		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	envVars, err := client.Revisions().GetEnvironmentVariables(context.Background(), "revision-guid")
	require.NoError(t, err)
	assert.Equal(t, "postgres://localhost/myapp", envVars["DATABASE_URL"])
	assert.Equal(t, "secret-key", envVars["API_KEY"])
	assert.Equal(t, true, envVars["DEBUG"])
}
