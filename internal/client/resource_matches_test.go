package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fivetwenty-io/capi/v3/pkg/capi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceMatchesClient_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/resource_matches", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req capi.ResourceMatchesRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.Len(t, req.Resources, 2)
		assert.Equal(t, "file1.txt", req.Resources[0].Path)
		assert.Equal(t, "checksum1", req.Resources[0].SHA1)
		assert.Equal(t, "file2.txt", req.Resources[1].Path)
		assert.Equal(t, "checksum2", req.Resources[1].SHA1)

		// Mock response - only the first file matches
		response := capi.ResourceMatches{
			Resources: []capi.ResourceMatch{
				{
					Path: "file1.txt",
					SHA1: "checksum1",
					Size: 1024,
					Mode: "0644",
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

	response, err := client.ResourceMatches().Create(context.Background(), &capi.ResourceMatchesRequest{
		Resources: []capi.ResourceMatch{
			{
				Path: "file1.txt",
				SHA1: "checksum1",
				Size: 1024,
				Mode: "0644",
			},
			{
				Path: "file2.txt",
				SHA1: "checksum2",
				Size: 2048,
				Mode: "0644",
			},
		},
	})

	require.NoError(t, err)
	assert.Len(t, response.Resources, 1)
	assert.Equal(t, "file1.txt", response.Resources[0].Path)
	assert.Equal(t, "checksum1", response.Resources[0].SHA1)
	assert.Equal(t, int64(1024), response.Resources[0].Size)
	assert.Equal(t, "0644", response.Resources[0].Mode)
}

func TestResourceMatchesClient_CreateEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/resource_matches", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req capi.ResourceMatchesRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.Len(t, req.Resources, 2)

		// Mock response - no files match
		response := capi.ResourceMatches{
			Resources: []capi.ResourceMatch{},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	response, err := client.ResourceMatches().Create(context.Background(), &capi.ResourceMatchesRequest{
		Resources: []capi.ResourceMatch{
			{
				Path: "new-file1.txt",
				SHA1: "new-checksum1",
				Size: 512,
				Mode: "0644",
			},
			{
				Path: "new-file2.txt",
				SHA1: "new-checksum2",
				Size: 1024,
				Mode: "0755",
			},
		},
	})

	require.NoError(t, err)
	assert.Len(t, response.Resources, 0)
}
