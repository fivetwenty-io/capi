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

func TestSpacesClient_Create(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/spaces", request.URL.Path)
		assert.Equal(t, "POST", request.Method)

		var req capi.SpaceCreateRequest

		_ = json.NewDecoder(request.Body).Decode(&req)
		assert.Equal(t, "test-space", req.Name)

		space := capi.Space{
			Resource: capi.Resource{
				GUID:      "space-guid",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name: req.Name,
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(writer).Encode(space)
	}))
	defer server.Close()

	c, err := New(context.Background(), &capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	space, err := c.Spaces().Create(context.Background(), &capi.SpaceCreateRequest{
		Name: "test-space",
		Relationships: capi.SpaceRelationships{
			Organization: capi.Relationship{
				Data: &capi.RelationshipData{GUID: "org-guid"},
			},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "space-guid", space.GUID)
	assert.Equal(t, "test-space", space.Name)
}

func TestSpacesClient_Get(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/spaces/space-guid", request.URL.Path)
		assert.Equal(t, "GET", request.Method)

		space := capi.Space{
			Resource: capi.Resource{
				GUID: "space-guid",
			},
			Name: "test-space",
		}

		_ = json.NewEncoder(writer).Encode(space)
	}))
	defer server.Close()

	c, err := New(context.Background(), &capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	space, err := c.Spaces().Get(context.Background(), "space-guid")
	require.NoError(t, err)
	assert.Equal(t, "space-guid", space.GUID)
	assert.Equal(t, "test-space", space.Name)
}

func TestSpacesClient_List(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/spaces", request.URL.Path)
		assert.Equal(t, "GET", request.Method)
		assert.Equal(t, "2", request.URL.Query().Get("page"))
		assert.Equal(t, "50", request.URL.Query().Get("per_page"))

		response := capi.ListResponse[capi.Space]{
			Pagination: capi.Pagination{
				TotalResults: 2,
				TotalPages:   1,
				First:        capi.Link{Href: "/v3/spaces?page=1"},
				Last:         capi.Link{Href: "/v3/spaces?page=1"},
			},
			Resources: []capi.Space{
				{
					Resource: capi.Resource{GUID: "space-1"},
					Name:     "space-1",
				},
				{
					Resource: capi.Resource{GUID: "space-2"},
					Name:     "space-2",
				},
			},
		}

		_ = json.NewEncoder(writer).Encode(response)
	}))
	defer server.Close()

	c, err := New(context.Background(), &capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	params := capi.NewQueryParams().WithPage(2).WithPerPage(50)
	result, err := c.Spaces().List(context.Background(), params)

	require.NoError(t, err)
	assert.Len(t, result.Resources, 2)
	assert.Equal(t, "space-1", result.Resources[0].Name)
	assert.Equal(t, "space-2", result.Resources[1].Name)
}

//nolint:dupl // Acceptable duplication - each test validates different resource types
func TestSpacesClient_Update(t *testing.T) {
	t.Parallel()
	RunNameUpdateTest(t, NameUpdateTestCase[capi.SpaceUpdateRequest, capi.Space]{
		ResourceType: "space",
		ResourceGUID: "space-guid",
		ResourcePath: "/v3/spaces/space-guid",
		NewName:      "updated-space",
		CreateRequest: func(name string) *capi.SpaceUpdateRequest {
			return &capi.SpaceUpdateRequest{Name: &name}
		},
		CreateResponse: func(guid, name string) *capi.Space {
			return &capi.Space{
				Resource: capi.Resource{GUID: guid},
				Name:     name,
			}
		},
		ExtractName:     func(req *capi.SpaceUpdateRequest) string { return *req.Name },
		ExtractNameResp: func(resp *capi.Space) string { return resp.Name },
		UpdateFunc: func(c *Client) func(context.Context, string, *capi.SpaceUpdateRequest) (*capi.Space, error) {
			return c.Spaces().Update
		},
	})
}

func TestSpacesClient_Delete(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/spaces/space-guid", request.URL.Path)
		assert.Equal(t, "DELETE", request.Method)

		job := capi.Job{
			Resource:  capi.Resource{GUID: "job-guid"},
			Operation: "space.delete",
			State:     "PROCESSING",
		}

		writer.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(writer).Encode(job)
	}))
	defer server.Close()

	c, err := New(context.Background(), &capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	job, err := c.Spaces().Delete(context.Background(), "space-guid")
	require.NoError(t, err)
	assert.Equal(t, "job-guid", job.GUID)
	assert.Equal(t, "PROCESSING", job.State)
}

func TestSpacesClient_GetIsolationSegment(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/spaces/space-guid/relationships/isolation_segment", request.URL.Path)
		assert.Equal(t, "GET", request.Method)

		relationship := capi.Relationship{
			Data: &capi.RelationshipData{GUID: "iso-seg-guid"},
		}

		_ = json.NewEncoder(writer).Encode(relationship)
	}))
	defer server.Close()

	c, err := New(context.Background(), &capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	relationship, err := c.Spaces().GetIsolationSegment(context.Background(), "space-guid")
	require.NoError(t, err)
	assert.NotNil(t, relationship.Data)
	assert.Equal(t, "iso-seg-guid", relationship.Data.GUID)
}

func TestSpacesClient_SetIsolationSegment(t *testing.T) {
	t.Parallel()

	tests := []TestRelationshipOperation{
		{
			Name:         "set isolation segment",
			ResourceGUID: "space-guid",
			TargetGUID:   "new-iso-seg-guid",
			ExpectedPath: "/v3/spaces/space-guid/relationships/isolation_segment",
			RelationshipFunc: func(c *Client) func(context.Context, string, string) (*capi.Relationship, error) {
				return c.Spaces().SetIsolationSegment
			},
		},
	}

	RunRelationshipTests(t, tests)
}

func TestSpacesClient_ListUsers(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/spaces/space-guid/users", request.URL.Path)
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

	result, err := c.Spaces().ListUsers(context.Background(), "space-guid", nil)
	require.NoError(t, err)
	assert.Len(t, result.Resources, 2)
	assert.Equal(t, "user1", result.Resources[0].Username)
	assert.Equal(t, "user2", result.Resources[1].Username)
}

func TestSpacesClient_ListManagers(t *testing.T) {
	t.Parallel()
	RunSpaceUserListTest(t, "list managers", "/v3/spaces/space-guid/managers", "manager-1", "manager1",
		func(c *Client) func(context.Context, string, *capi.QueryParams) (*capi.ListResponse[capi.User], error) {
			return c.Spaces().ListManagers
		},
	)
}

func TestSpacesClient_ListDevelopers(t *testing.T) {
	t.Parallel()
	RunSpaceUserListTest(t, "list developers", "/v3/spaces/space-guid/developers", "dev-1", "developer1",
		func(c *Client) func(context.Context, string, *capi.QueryParams) (*capi.ListResponse[capi.User], error) {
			return c.Spaces().ListDevelopers
		},
	)
}

func TestSpacesClient_GetFeature(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/spaces/space-guid/features/ssh", request.URL.Path)
		assert.Equal(t, "GET", request.Method)

		feature := capi.SpaceFeature{
			Name:        "ssh",
			Enabled:     true,
			Description: "Enable SSH access to apps",
		}

		_ = json.NewEncoder(writer).Encode(feature)
	}))
	defer server.Close()

	c, err := New(context.Background(), &capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	feature, err := c.Spaces().GetFeature(context.Background(), "space-guid", "ssh")
	require.NoError(t, err)
	assert.Equal(t, "ssh", feature.Name)
	assert.True(t, feature.Enabled)
}

func TestSpacesClient_UpdateFeature(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/spaces/space-guid/features/ssh", request.URL.Path)
		assert.Equal(t, "PATCH", request.Method)

		var req map[string]bool

		_ = json.NewDecoder(request.Body).Decode(&req)
		assert.False(t, req["enabled"])

		feature := capi.SpaceFeature{
			Name:        "ssh",
			Enabled:     false,
			Description: "Enable SSH access to apps",
		}

		_ = json.NewEncoder(writer).Encode(feature)
	}))
	defer server.Close()

	c, err := New(context.Background(), &capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	feature, err := c.Spaces().UpdateFeature(context.Background(), "space-guid", "ssh", false)
	require.NoError(t, err)
	assert.Equal(t, "ssh", feature.Name)
	assert.False(t, feature.Enabled)
}

func TestSpacesClient_GetQuota(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/spaces/space-guid/quota", request.URL.Path)
		assert.Equal(t, "GET", request.Method)

		totalMem := 1024
		totalInstances := 10
		quota := capi.SpaceQuota{
			Resource: capi.Resource{GUID: "quota-guid"},
			Name:     "test-quota",
			Apps: &capi.AppsQuota{
				TotalMemoryInMB: &totalMem,
				TotalInstances:  &totalInstances,
			},
		}

		_ = json.NewEncoder(writer).Encode(quota)
	}))
	defer server.Close()

	c, err := New(context.Background(), &capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	quota, err := c.Spaces().GetQuota(context.Background(), "space-guid")
	require.NoError(t, err)
	assert.Equal(t, "quota-guid", quota.GUID)
	assert.Equal(t, "test-quota", quota.Name)
	assert.Equal(t, 1024, *quota.Apps.TotalMemoryInMB)
}

func TestSpacesClient_ApplyQuota(t *testing.T) {
	t.Parallel()

	tests := []TestRelationshipOperation{
		{
			Name:         "apply quota",
			ResourceGUID: "space-guid",
			TargetGUID:   "quota-guid",
			ExpectedPath: "/v3/spaces/space-guid/relationships/quota",
			RelationshipFunc: func(c *Client) func(context.Context, string, string) (*capi.Relationship, error) {
				return c.Spaces().ApplyQuota
			},
		},
	}

	RunRelationshipTests(t, tests)
}

func TestSpacesClient_RemoveQuota(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/spaces/space-guid/relationships/quota", request.URL.Path)
		assert.Equal(t, "PATCH", request.Method)

		var req capi.Relationship

		_ = json.NewDecoder(request.Body).Decode(&req)
		assert.Nil(t, req.Data)

		relationship := capi.Relationship{
			Data: nil,
		}

		_ = json.NewEncoder(writer).Encode(relationship)
	}))
	defer server.Close()

	c, err := New(context.Background(), &capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	err = c.Spaces().RemoveQuota(context.Background(), "space-guid")
	require.NoError(t, err)
}
