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

func TestOrganizationsClient_Create(t *testing.T) {
	t.Parallel()
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/organizations", request.URL.Path)
		assert.Equal(t, "POST", request.Method)

		var req capi.OrganizationCreateRequest

		_ = json.NewDecoder(request.Body).Decode(&req)
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

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(writer).Encode(org)
	}))
	defer server.Close()

	c, err := New(context.Background(), &capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	org, err := c.Organizations().Create(context.Background(), &capi.OrganizationCreateRequest{
		Name: "test-org",
	})

	require.NoError(t, err)
	assert.Equal(t, "org-guid", org.GUID)
	assert.Equal(t, "test-org", org.Name)
}

func TestOrganizationsClient_Get(t *testing.T) {
	t.Parallel()
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/organizations/org-guid", request.URL.Path)
		assert.Equal(t, "GET", request.Method)

		org := capi.Organization{
			Resource: capi.Resource{
				GUID: "org-guid",
			},
			Name:      "test-org",
			Suspended: false,
		}

		_ = json.NewEncoder(writer).Encode(org)
	}))
	defer server.Close()

	c, err := New(context.Background(), &capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	org, err := c.Organizations().Get(context.Background(), "org-guid")
	require.NoError(t, err)
	assert.Equal(t, "org-guid", org.GUID)
	assert.Equal(t, "test-org", org.Name)
}

func TestOrganizationsClient_List(t *testing.T) {
	t.Parallel()
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/organizations", request.URL.Path)
		assert.Equal(t, "GET", request.Method)
		assert.Equal(t, "2", request.URL.Query().Get("page"))
		assert.Equal(t, "50", request.URL.Query().Get("per_page"))

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

		_ = json.NewEncoder(writer).Encode(response)
	}))
	defer server.Close()

	c, err := New(context.Background(), &capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	params := capi.NewQueryParams().WithPage(2).WithPerPage(50)
	result, err := c.Organizations().List(context.Background(), params)

	require.NoError(t, err)
	assert.Len(t, result.Resources, 2)
	assert.Equal(t, "org-1", result.Resources[0].Name)
	assert.Equal(t, "org-2", result.Resources[1].Name)
}

//nolint:dupl // Acceptable duplication - each test validates different resource types
func TestOrganizationsClient_Update(t *testing.T) {
	t.Parallel()
	RunNameUpdateTest(t, NameUpdateTestCase[capi.OrganizationUpdateRequest, capi.Organization]{
		ResourceType: "organization",
		ResourceGUID: "org-guid",
		ResourcePath: "/v3/organizations/org-guid",
		NewName:      "updated-org",
		CreateRequest: func(name string) *capi.OrganizationUpdateRequest {
			return &capi.OrganizationUpdateRequest{Name: &name}
		},
		CreateResponse: func(guid, name string) *capi.Organization {
			return &capi.Organization{
				Resource: capi.Resource{GUID: guid},
				Name:     name,
			}
		},
		ExtractName:     func(req *capi.OrganizationUpdateRequest) string { return *req.Name },
		ExtractNameResp: func(resp *capi.Organization) string { return resp.Name },
		UpdateFunc: func(c *Client) func(context.Context, string, *capi.OrganizationUpdateRequest) (*capi.Organization, error) {
			return c.Organizations().Update
		},
	})
}

func TestOrganizationsClient_Delete(t *testing.T) {
	t.Parallel()
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/organizations/org-guid", request.URL.Path)
		assert.Equal(t, "DELETE", request.Method)

		job := capi.Job{
			Resource:  capi.Resource{GUID: "job-guid"},
			Operation: "organization.delete",
			State:     "PROCESSING",
		}

		writer.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(writer).Encode(job)
	}))
	defer server.Close()

	c, err := New(context.Background(), &capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	job, err := c.Organizations().Delete(context.Background(), "org-guid")
	require.NoError(t, err)
	assert.Equal(t, "job-guid", job.GUID)
	assert.Equal(t, "PROCESSING", job.State)
}

func TestOrganizationsClient_GetUsageSummary(t *testing.T) {
	t.Parallel()
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/organizations/org-guid/usage_summary", request.URL.Path)
		assert.Equal(t, "GET", request.Method)

		summary := capi.OrganizationUsageSummary{}
		summary.UsageSummary.StartedInstances = 5
		summary.UsageSummary.MemoryInMB = 1024

		_ = json.NewEncoder(writer).Encode(summary)
	}))
	defer server.Close()

	c, err := New(context.Background(), &capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	summary, err := c.Organizations().GetUsageSummary(context.Background(), "org-guid")
	require.NoError(t, err)
	assert.Equal(t, 5, summary.UsageSummary.StartedInstances)
	assert.Equal(t, 1024, summary.UsageSummary.MemoryInMB)
}

func TestOrganizationsClient_ListUsers(t *testing.T) {
	t.Parallel()
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/organizations/org-guid/users", request.URL.Path)
		assert.Equal(t, "GET", request.Method)

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

		_ = json.NewEncoder(writer).Encode(response)
	}))
	defer server.Close()

	c, err := New(context.Background(), &capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	result, err := c.Organizations().ListUsers(context.Background(), "org-guid", nil)
	require.NoError(t, err)
	assert.Len(t, result.Resources, 2)
	assert.Equal(t, "user1", result.Resources[0].Username)
	assert.Equal(t, "user2", result.Resources[1].Username)
}
