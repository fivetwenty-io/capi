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

func TestSpacesClient_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/spaces", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req capi.SpaceCreateRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "test-space", req.Name)

		space := capi.Space{
			Resource: capi.Resource{
				GUID:      "space-guid",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name: req.Name,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(space)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	space, err := client.Spaces().Create(context.Background(), &capi.SpaceCreateRequest{
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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/spaces/space-guid", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		space := capi.Space{
			Resource: capi.Resource{
				GUID: "space-guid",
			},
			Name: "test-space",
		}

		json.NewEncoder(w).Encode(space)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	space, err := client.Spaces().Get(context.Background(), "space-guid")
	require.NoError(t, err)
	assert.Equal(t, "space-guid", space.GUID)
	assert.Equal(t, "test-space", space.Name)
}

func TestSpacesClient_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/spaces", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "2", r.URL.Query().Get("page"))
		assert.Equal(t, "50", r.URL.Query().Get("per_page"))

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

		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	params := capi.NewQueryParams().WithPage(2).WithPerPage(50)
	result, err := client.Spaces().List(context.Background(), params)

	require.NoError(t, err)
	assert.Len(t, result.Resources, 2)
	assert.Equal(t, "space-1", result.Resources[0].Name)
	assert.Equal(t, "space-2", result.Resources[1].Name)
}

func TestSpacesClient_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/spaces/space-guid", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var req capi.SpaceUpdateRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "updated-space", *req.Name)

		space := capi.Space{
			Resource: capi.Resource{GUID: "space-guid"},
			Name:     *req.Name,
		}

		json.NewEncoder(w).Encode(space)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	newName := "updated-space"
	space, err := client.Spaces().Update(context.Background(), "space-guid", &capi.SpaceUpdateRequest{
		Name: &newName,
	})

	require.NoError(t, err)
	assert.Equal(t, "updated-space", space.Name)
}

func TestSpacesClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/spaces/space-guid", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		job := capi.Job{
			Resource:  capi.Resource{GUID: "job-guid"},
			Operation: "space.delete",
			State:     "PROCESSING",
		}

		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(job)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	job, err := client.Spaces().Delete(context.Background(), "space-guid")
	require.NoError(t, err)
	assert.Equal(t, "job-guid", job.GUID)
	assert.Equal(t, "PROCESSING", job.State)
}

func TestSpacesClient_GetIsolationSegment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/spaces/space-guid/relationships/isolation_segment", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		relationship := capi.Relationship{
			Data: &capi.RelationshipData{GUID: "iso-seg-guid"},
		}

		json.NewEncoder(w).Encode(relationship)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	relationship, err := client.Spaces().GetIsolationSegment(context.Background(), "space-guid")
	require.NoError(t, err)
	assert.NotNil(t, relationship.Data)
	assert.Equal(t, "iso-seg-guid", relationship.Data.GUID)
}

func TestSpacesClient_SetIsolationSegment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/spaces/space-guid/relationships/isolation_segment", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var req capi.Relationship
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "new-iso-seg-guid", req.Data.GUID)

		relationship := capi.Relationship{
			Data: req.Data,
		}

		json.NewEncoder(w).Encode(relationship)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	relationship, err := client.Spaces().SetIsolationSegment(context.Background(), "space-guid", "new-iso-seg-guid")
	require.NoError(t, err)
	assert.NotNil(t, relationship.Data)
	assert.Equal(t, "new-iso-seg-guid", relationship.Data.GUID)
}

func TestSpacesClient_ListUsers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/spaces/space-guid/users", r.URL.Path)
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

		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	result, err := client.Spaces().ListUsers(context.Background(), "space-guid", nil)
	require.NoError(t, err)
	assert.Len(t, result.Resources, 2)
	assert.Equal(t, "user1", result.Resources[0].Username)
	assert.Equal(t, "user2", result.Resources[1].Username)
}

func TestSpacesClient_ListManagers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/spaces/space-guid/managers", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		response := capi.ListResponse[capi.User]{
			Resources: []capi.User{
				{
					Resource: capi.Resource{GUID: "manager-1"},
					Username: "manager1",
				},
			},
		}

		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	result, err := client.Spaces().ListManagers(context.Background(), "space-guid", nil)
	require.NoError(t, err)
	assert.Len(t, result.Resources, 1)
	assert.Equal(t, "manager1", result.Resources[0].Username)
}

func TestSpacesClient_ListDevelopers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/spaces/space-guid/developers", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		response := capi.ListResponse[capi.User]{
			Resources: []capi.User{
				{
					Resource: capi.Resource{GUID: "dev-1"},
					Username: "developer1",
				},
			},
		}

		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	result, err := client.Spaces().ListDevelopers(context.Background(), "space-guid", nil)
	require.NoError(t, err)
	assert.Len(t, result.Resources, 1)
	assert.Equal(t, "developer1", result.Resources[0].Username)
}

func TestSpacesClient_GetFeature(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/spaces/space-guid/features/ssh", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		feature := capi.SpaceFeature{
			Name:        "ssh",
			Enabled:     true,
			Description: "Enable SSH access to apps",
		}

		json.NewEncoder(w).Encode(feature)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	feature, err := client.Spaces().GetFeature(context.Background(), "space-guid", "ssh")
	require.NoError(t, err)
	assert.Equal(t, "ssh", feature.Name)
	assert.True(t, feature.Enabled)
}

func TestSpacesClient_UpdateFeature(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/spaces/space-guid/features/ssh", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var req map[string]bool
		json.NewDecoder(r.Body).Decode(&req)
		assert.False(t, req["enabled"])

		feature := capi.SpaceFeature{
			Name:        "ssh",
			Enabled:     false,
			Description: "Enable SSH access to apps",
		}

		json.NewEncoder(w).Encode(feature)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	feature, err := client.Spaces().UpdateFeature(context.Background(), "space-guid", "ssh", false)
	require.NoError(t, err)
	assert.Equal(t, "ssh", feature.Name)
	assert.False(t, feature.Enabled)
}

func TestSpacesClient_GetQuota(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/spaces/space-guid/quota", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

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

		json.NewEncoder(w).Encode(quota)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	quota, err := client.Spaces().GetQuota(context.Background(), "space-guid")
	require.NoError(t, err)
	assert.Equal(t, "quota-guid", quota.GUID)
	assert.Equal(t, "test-quota", quota.Name)
	assert.Equal(t, 1024, *quota.Apps.TotalMemoryInMB)
}

func TestSpacesClient_ApplyQuota(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/spaces/space-guid/relationships/quota", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var req capi.Relationship
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "quota-guid", req.Data.GUID)

		relationship := capi.Relationship{
			Data: req.Data,
		}

		json.NewEncoder(w).Encode(relationship)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	relationship, err := client.Spaces().ApplyQuota(context.Background(), "space-guid", "quota-guid")
	require.NoError(t, err)
	assert.NotNil(t, relationship.Data)
	assert.Equal(t, "quota-guid", relationship.Data.GUID)
}

func TestSpacesClient_RemoveQuota(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/spaces/space-guid/relationships/quota", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var req capi.Relationship
		json.NewDecoder(r.Body).Decode(&req)
		assert.Nil(t, req.Data)

		relationship := capi.Relationship{
			Data: nil,
		}

		json.NewEncoder(w).Encode(relationship)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	err = client.Spaces().RemoveQuota(context.Background(), "space-guid")
	require.NoError(t, err)
}
