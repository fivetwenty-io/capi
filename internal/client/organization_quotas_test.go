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

func TestOrganizationQuotasClient_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/organization_quotas", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req capi.OrganizationQuotaCreateRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "test-quota", req.Name)

		totalMemory := 1024
		quota := capi.OrganizationQuota{
			Resource: capi.Resource{
				GUID:      "quota-guid",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name: req.Name,
			Apps: &capi.OrganizationQuotaApps{
				TotalMemoryInMB: &totalMemory,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(quota)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	totalMemory := 1024
	quota, err := client.OrganizationQuotas().Create(context.Background(), &capi.OrganizationQuotaCreateRequest{
		Name: "test-quota",
		Apps: &capi.OrganizationQuotaApps{
			TotalMemoryInMB: &totalMemory,
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "quota-guid", quota.GUID)
	assert.Equal(t, "test-quota", quota.Name)
	assert.Equal(t, 1024, *quota.Apps.TotalMemoryInMB)
}

func TestOrganizationQuotasClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/organization_quotas/quota-guid", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		totalMemory := 2048
		quota := capi.OrganizationQuota{
			Resource: capi.Resource{
				GUID: "quota-guid",
			},
			Name: "test-quota",
			Apps: &capi.OrganizationQuotaApps{
				TotalMemoryInMB: &totalMemory,
			},
		}

		json.NewEncoder(w).Encode(quota)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	quota, err := client.OrganizationQuotas().Get(context.Background(), "quota-guid")
	require.NoError(t, err)
	assert.Equal(t, "quota-guid", quota.GUID)
	assert.Equal(t, "test-quota", quota.Name)
	assert.Equal(t, 2048, *quota.Apps.TotalMemoryInMB)
}

func TestOrganizationQuotasClient_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/organization_quotas", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "10", r.URL.Query().Get("per_page"))

		totalMemory1 := 1024
		totalMemory2 := 2048
		response := capi.ListResponse[capi.OrganizationQuota]{
			Pagination: capi.Pagination{
				TotalResults: 2,
				TotalPages:   1,
			},
			Resources: []capi.OrganizationQuota{
				{
					Resource: capi.Resource{GUID: "quota-1"},
					Name:     "quota-1",
					Apps: &capi.OrganizationQuotaApps{
						TotalMemoryInMB: &totalMemory1,
					},
				},
				{
					Resource: capi.Resource{GUID: "quota-2"},
					Name:     "quota-2",
					Apps: &capi.OrganizationQuotaApps{
						TotalMemoryInMB: &totalMemory2,
					},
				},
			},
		}

		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	params := capi.NewQueryParams().WithPage(1).WithPerPage(10)
	result, err := client.OrganizationQuotas().List(context.Background(), params)

	require.NoError(t, err)
	assert.Len(t, result.Resources, 2)
	assert.Equal(t, "quota-1", result.Resources[0].Name)
	assert.Equal(t, "quota-2", result.Resources[1].Name)
}

func TestOrganizationQuotasClient_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/organization_quotas/quota-guid", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var req capi.OrganizationQuotaUpdateRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "updated-quota", *req.Name)

		quota := capi.OrganizationQuota{
			Resource: capi.Resource{GUID: "quota-guid"},
			Name:     *req.Name,
		}

		json.NewEncoder(w).Encode(quota)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	newName := "updated-quota"
	quota, err := client.OrganizationQuotas().Update(context.Background(), "quota-guid", &capi.OrganizationQuotaUpdateRequest{
		Name: &newName,
	})

	require.NoError(t, err)
	assert.Equal(t, "updated-quota", quota.Name)
}

func TestOrganizationQuotasClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/organization_quotas/quota-guid", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	err = client.OrganizationQuotas().Delete(context.Background(), "quota-guid")
	require.NoError(t, err)
}

func TestOrganizationQuotasClient_ApplyToOrganizations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/organization_quotas/quota-guid/relationships/organizations", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req capi.ToManyRelationship
		json.NewDecoder(r.Body).Decode(&req)
		assert.Len(t, req.Data, 2)
		assert.Equal(t, "org-1", req.Data[0].GUID)
		assert.Equal(t, "org-2", req.Data[1].GUID)

		response := capi.ToManyRelationship{
			Data: req.Data,
		}

		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	rel, err := client.OrganizationQuotas().ApplyToOrganizations(context.Background(), "quota-guid", []string{"org-1", "org-2"})
	require.NoError(t, err)
	assert.Len(t, rel.Data, 2)
	assert.Equal(t, "org-1", rel.Data[0].GUID)
	assert.Equal(t, "org-2", rel.Data[1].GUID)
}
