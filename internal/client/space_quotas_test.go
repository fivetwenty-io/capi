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

func TestSpaceQuotasClient_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/space_quotas", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req capi.SpaceQuotaV3CreateRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "test-space-quota", req.Name)

		totalMemory := 512
		quota := capi.SpaceQuotaV3{
			Resource: capi.Resource{
				GUID:      "space-quota-guid",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name: req.Name,
			Apps: &capi.SpaceQuotaApps{
				TotalMemoryInMB: &totalMemory,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(quota)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	totalMemory := 512
	quota, err := client.SpaceQuotas().Create(context.Background(), &capi.SpaceQuotaV3CreateRequest{
		Name: "test-space-quota",
		Apps: &capi.SpaceQuotaApps{
			TotalMemoryInMB: &totalMemory,
		},
		Relationships: capi.SpaceQuotaRelationships{
			Organization: capi.Relationship{
				Data: &capi.RelationshipData{GUID: "org-guid"},
			},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "space-quota-guid", quota.GUID)
	assert.Equal(t, "test-space-quota", quota.Name)
	assert.Equal(t, 512, *quota.Apps.TotalMemoryInMB)
}

func TestSpaceQuotasClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/space_quotas/space-quota-guid", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		totalMemory := 1024
		quota := capi.SpaceQuotaV3{
			Resource: capi.Resource{
				GUID: "space-quota-guid",
			},
			Name: "test-space-quota",
			Apps: &capi.SpaceQuotaApps{
				TotalMemoryInMB: &totalMemory,
			},
		}

		_ = json.NewEncoder(w).Encode(quota)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	quota, err := client.SpaceQuotas().Get(context.Background(), "space-quota-guid")
	require.NoError(t, err)
	assert.Equal(t, "space-quota-guid", quota.GUID)
	assert.Equal(t, "test-space-quota", quota.Name)
	assert.Equal(t, 1024, *quota.Apps.TotalMemoryInMB)
}

func TestSpaceQuotasClient_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/space_quotas", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "10", r.URL.Query().Get("per_page"))

		totalMemory1 := 512
		totalMemory2 := 1024
		response := capi.ListResponse[capi.SpaceQuotaV3]{
			Pagination: capi.Pagination{
				TotalResults: 2,
				TotalPages:   1,
			},
			Resources: []capi.SpaceQuotaV3{
				{
					Resource: capi.Resource{GUID: "space-quota-1"},
					Name:     "space-quota-1",
					Apps: &capi.SpaceQuotaApps{
						TotalMemoryInMB: &totalMemory1,
					},
				},
				{
					Resource: capi.Resource{GUID: "space-quota-2"},
					Name:     "space-quota-2",
					Apps: &capi.SpaceQuotaApps{
						TotalMemoryInMB: &totalMemory2,
					},
				},
			},
		}

		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	params := capi.NewQueryParams().WithPage(1).WithPerPage(10)
	result, err := client.SpaceQuotas().List(context.Background(), params)

	require.NoError(t, err)
	assert.Len(t, result.Resources, 2)
	assert.Equal(t, "space-quota-1", result.Resources[0].Name)
	assert.Equal(t, "space-quota-2", result.Resources[1].Name)
}

func TestSpaceQuotasClient_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/space_quotas/space-quota-guid", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var req capi.SpaceQuotaV3UpdateRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "updated-space-quota", *req.Name)

		quota := capi.SpaceQuotaV3{
			Resource: capi.Resource{GUID: "space-quota-guid"},
			Name:     *req.Name,
		}

		_ = json.NewEncoder(w).Encode(quota)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	newName := "updated-space-quota"
	quota, err := client.SpaceQuotas().Update(context.Background(), "space-quota-guid", &capi.SpaceQuotaV3UpdateRequest{
		Name: &newName,
	})

	require.NoError(t, err)
	assert.Equal(t, "updated-space-quota", quota.Name)
}

func TestSpaceQuotasClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/space_quotas/space-quota-guid", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	err = client.SpaceQuotas().Delete(context.Background(), "space-quota-guid")
	require.NoError(t, err)
}

func TestSpaceQuotasClient_ApplyToSpaces(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/space_quotas/space-quota-guid/relationships/spaces", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req capi.ToManyRelationship
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.Len(t, req.Data, 2)
		assert.Equal(t, "space-1", req.Data[0].GUID)
		assert.Equal(t, "space-2", req.Data[1].GUID)

		response := capi.ToManyRelationship{
			Data: req.Data,
		}

		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	rel, err := client.SpaceQuotas().ApplyToSpaces(context.Background(), "space-quota-guid", []string{"space-1", "space-2"})
	require.NoError(t, err)
	assert.Len(t, rel.Data, 2)
	assert.Equal(t, "space-1", rel.Data[0].GUID)
	assert.Equal(t, "space-2", rel.Data[1].GUID)
}

func TestSpaceQuotasClient_RemoveFromSpace(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/space_quotas/space-quota-guid/relationships/spaces/space-guid", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	err = client.SpaceQuotas().RemoveFromSpace(context.Background(), "space-quota-guid", "space-guid")
	require.NoError(t, err)
}
