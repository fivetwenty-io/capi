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

func TestOrganizationsClient_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/organizations", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req capi.OrganizationCreateRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "test-org", req.Name)

		org := capi.Organization{
			Resource: capi.Resource{
				GUID:      "org-guid",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:      req.Name,
			Suspended: false,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(org)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	org, err := client.Organizations().Create(context.Background(), &capi.OrganizationCreateRequest{
		Name: "test-org",
	})

	require.NoError(t, err)
	assert.Equal(t, "org-guid", org.GUID)
	assert.Equal(t, "test-org", org.Name)
}

func TestOrganizationsClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/organizations/org-guid", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		org := capi.Organization{
			Resource: capi.Resource{
				GUID: "org-guid",
			},
			Name:      "test-org",
			Suspended: false,
		}

		_ = json.NewEncoder(w).Encode(org)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	org, err := client.Organizations().Get(context.Background(), "org-guid")
	require.NoError(t, err)
	assert.Equal(t, "org-guid", org.GUID)
	assert.Equal(t, "test-org", org.Name)
}

func TestOrganizationsClient_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/organizations", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "2", r.URL.Query().Get("page"))
		assert.Equal(t, "50", r.URL.Query().Get("per_page"))

		response := capi.ListResponse[capi.Organization]{
			Pagination: capi.Pagination{
				TotalResults: 2,
				TotalPages:   1,
				First:        capi.Link{Href: "/v3/organizations?page=1"},
				Last:         capi.Link{Href: "/v3/organizations?page=1"},
			},
			Resources: []capi.Organization{
				{
					Resource: capi.Resource{GUID: "org-1"},
					Name:     "org-1",
				},
				{
					Resource: capi.Resource{GUID: "org-2"},
					Name:     "org-2",
				},
			},
		}

		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	params := capi.NewQueryParams().WithPage(2).WithPerPage(50)
	result, err := client.Organizations().List(context.Background(), params)

	require.NoError(t, err)
	assert.Len(t, result.Resources, 2)
	assert.Equal(t, "org-1", result.Resources[0].Name)
	assert.Equal(t, "org-2", result.Resources[1].Name)
}

func TestOrganizationsClient_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/organizations/org-guid", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var req capi.OrganizationUpdateRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "updated-org", *req.Name)

		org := capi.Organization{
			Resource: capi.Resource{GUID: "org-guid"},
			Name:     *req.Name,
		}

		_ = json.NewEncoder(w).Encode(org)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	newName := "updated-org"
	org, err := client.Organizations().Update(context.Background(), "org-guid", &capi.OrganizationUpdateRequest{
		Name: &newName,
	})

	require.NoError(t, err)
	assert.Equal(t, "updated-org", org.Name)
}

func TestOrganizationsClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/organizations/org-guid", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		job := capi.Job{
			Resource:  capi.Resource{GUID: "job-guid"},
			Operation: "organization.delete",
			State:     "PROCESSING",
		}

		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(job)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	job, err := client.Organizations().Delete(context.Background(), "org-guid")
	require.NoError(t, err)
	assert.Equal(t, "job-guid", job.GUID)
	assert.Equal(t, "PROCESSING", job.State)
}

func TestOrganizationsClient_GetUsageSummary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/organizations/org-guid/usage_summary", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		summary := capi.OrganizationUsageSummary{}
		summary.UsageSummary.StartedInstances = 5
		summary.UsageSummary.MemoryInMB = 1024

		_ = json.NewEncoder(w).Encode(summary)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	summary, err := client.Organizations().GetUsageSummary(context.Background(), "org-guid")
	require.NoError(t, err)
	assert.Equal(t, 5, summary.UsageSummary.StartedInstances)
	assert.Equal(t, 1024, summary.UsageSummary.MemoryInMB)
}

func TestOrganizationsClient_ListUsers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/organizations/org-guid/users", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		response := capi.ListResponse[capi.User]{
			Pagination: capi.Pagination{
				TotalResults: 2,
				TotalPages:   1,
			},
			Resources: []capi.User{
				{
					Resource:         capi.Resource{GUID: "user-1"},
					Username:         "user1",
					PresentationName: "User One",
				},
				{
					Resource:         capi.Resource{GUID: "user-2"},
					Username:         "user2",
					PresentationName: "User Two",
				},
			},
		}

		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	result, err := client.Organizations().ListUsers(context.Background(), "org-guid", nil)
	require.NoError(t, err)
	assert.Len(t, result.Resources, 2)
	assert.Equal(t, "user1", result.Resources[0].Username)
	assert.Equal(t, "user2", result.Resources[1].Username)
}
