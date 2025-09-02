package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fivetwenty-io/capi/v3/pkg/capi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSidecarsClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/sidecars/sidecar-guid", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		memoryInMB := 128
		sidecar := capi.Sidecar{
			Resource: capi.Resource{
				GUID:      "sidecar-guid",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:         "test-sidecar",
			Command:      "echo hello",
			ProcessTypes: []string{"web", "worker"},
			MemoryInMB:   &memoryInMB,
			Origin:       "user",
		}

		_ = json.NewEncoder(w).Encode(sidecar)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	sidecar, err := client.Sidecars().Get(context.Background(), "sidecar-guid")
	require.NoError(t, err)
	assert.Equal(t, "sidecar-guid", sidecar.GUID)
	assert.Equal(t, "test-sidecar", sidecar.Name)
	assert.Equal(t, "echo hello", sidecar.Command)
	assert.Equal(t, []string{"web", "worker"}, sidecar.ProcessTypes)
	assert.Equal(t, 128, *sidecar.MemoryInMB)
}

func TestSidecarsClient_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/sidecars/sidecar-guid", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var req capi.SidecarUpdateRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "updated-sidecar", *req.Name)
		assert.Equal(t, "echo updated", *req.Command)

		sidecar := capi.Sidecar{
			Resource: capi.Resource{GUID: "sidecar-guid"},
			Name:     *req.Name,
			Command:  *req.Command,
		}

		_ = json.NewEncoder(w).Encode(sidecar)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	newName := "updated-sidecar"
	newCommand := "echo updated"
	sidecar, err := client.Sidecars().Update(context.Background(), "sidecar-guid", &capi.SidecarUpdateRequest{
		Name:    &newName,
		Command: &newCommand,
	})

	require.NoError(t, err)
	assert.Equal(t, "updated-sidecar", sidecar.Name)
	assert.Equal(t, "echo updated", sidecar.Command)
}

func TestSidecarsClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/sidecars/sidecar-guid", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	err = client.Sidecars().Delete(context.Background(), "sidecar-guid")
	require.NoError(t, err)
}

func TestSidecarsClient_ListForProcess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/processes/process-guid/sidecars", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "10", r.URL.Query().Get("per_page"))

		memory1 := 64
		memory2 := 128
		response := capi.ListResponse[capi.Sidecar]{
			Pagination: capi.Pagination{
				TotalResults: 2,
				TotalPages:   1,
			},
			Resources: []capi.Sidecar{
				{
					Resource:     capi.Resource{GUID: "sidecar-1"},
					Name:         "sidecar-1",
					Command:      "echo hello1",
					ProcessTypes: []string{"web"},
					MemoryInMB:   &memory1,
					Origin:       "user",
				},
				{
					Resource:     capi.Resource{GUID: "sidecar-2"},
					Name:         "sidecar-2",
					Command:      "echo hello2",
					ProcessTypes: []string{"worker"},
					MemoryInMB:   &memory2,
					Origin:       "user",
				},
			},
		}

		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	params := capi.NewQueryParams().WithPage(1).WithPerPage(10)
	result, err := client.Sidecars().ListForProcess(context.Background(), "process-guid", params)

	require.NoError(t, err)
	assert.Len(t, result.Resources, 2)
	assert.Equal(t, "sidecar-1", result.Resources[0].Name)
	assert.Equal(t, "sidecar-2", result.Resources[1].Name)
	assert.Equal(t, "echo hello1", result.Resources[0].Command)
	assert.Equal(t, "echo hello2", result.Resources[1].Command)
}
