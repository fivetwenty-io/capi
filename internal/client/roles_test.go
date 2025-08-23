package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	internalhttp "github.com/fivetwenty-io/capi/internal/http"
	"github.com/fivetwenty-io/capi/pkg/capi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRolesClient_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/roles", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var request capi.RoleCreateRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		require.NoError(t, err)

		assert.Equal(t, "organization_auditor", request.Type)
		assert.Equal(t, "user-guid", request.Relationships.User.Data.GUID)
		assert.Equal(t, "org-guid", request.Relationships.Organization.Data.GUID)

		now := time.Now()
		role := capi.Role{
			Resource: capi.Resource{
				GUID:      "role-guid",
				CreatedAt: now,
				UpdatedAt: now,
			},
			Type:          request.Type,
			Relationships: request.Relationships,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(role)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	roles := NewRolesClient(client.httpClient)

	request := &capi.RoleCreateRequest{
		Type: "organization_auditor",
		Relationships: capi.RoleRelationships{
			User: capi.Relationship{
				Data: &capi.RelationshipData{
					GUID: "user-guid",
				},
			},
			Organization: &capi.Relationship{
				Data: &capi.RelationshipData{
					GUID: "org-guid",
				},
			},
		},
	}

	role, err := roles.Create(context.Background(), request)
	require.NoError(t, err)
	assert.NotNil(t, role)
	assert.Equal(t, "role-guid", role.GUID)
	assert.Equal(t, "organization_auditor", role.Type)
}

func TestRolesClient_CreateSpaceRole(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/roles", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var request capi.RoleCreateRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		require.NoError(t, err)

		assert.Equal(t, "space_developer", request.Type)
		assert.Equal(t, "user-guid", request.Relationships.User.Data.GUID)
		assert.Equal(t, "space-guid", request.Relationships.Space.Data.GUID)

		now := time.Now()
		role := capi.Role{
			Resource: capi.Resource{
				GUID:      "role-guid",
				CreatedAt: now,
				UpdatedAt: now,
			},
			Type:          request.Type,
			Relationships: request.Relationships,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(role)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	roles := NewRolesClient(client.httpClient)

	request := &capi.RoleCreateRequest{
		Type: "space_developer",
		Relationships: capi.RoleRelationships{
			User: capi.Relationship{
				Data: &capi.RelationshipData{
					GUID: "user-guid",
				},
			},
			Space: &capi.Relationship{
				Data: &capi.RelationshipData{
					GUID: "space-guid",
				},
			},
		},
	}

	role, err := roles.Create(context.Background(), request)
	require.NoError(t, err)
	assert.NotNil(t, role)
	assert.Equal(t, "role-guid", role.GUID)
	assert.Equal(t, "space_developer", role.Type)
}

func TestRolesClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/roles/role-guid", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		now := time.Now()
		role := capi.Role{
			Resource: capi.Resource{
				GUID:      "role-guid",
				CreatedAt: now,
				UpdatedAt: now,
			},
			Type: "organization_manager",
			Relationships: capi.RoleRelationships{
				User: capi.Relationship{
					Data: &capi.RelationshipData{
						GUID: "user-guid",
					},
				},
				Organization: &capi.Relationship{
					Data: &capi.RelationshipData{
						GUID: "org-guid",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(role)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	roles := NewRolesClient(client.httpClient)

	role, err := roles.Get(context.Background(), "role-guid")
	require.NoError(t, err)
	assert.NotNil(t, role)
	assert.Equal(t, "role-guid", role.GUID)
	assert.Equal(t, "organization_manager", role.Type)
}

func TestRolesClient_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/roles", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "organization_auditor,organization_manager", r.URL.Query().Get("types"))
		assert.Equal(t, "org-guid", r.URL.Query().Get("organization_guids"))

		now := time.Now()
		response := capi.ListResponse[capi.Role]{
			Pagination: capi.Pagination{
				TotalResults: 2,
				TotalPages:   1,
				First:        capi.Link{Href: "/v3/roles?page=1"},
				Last:         capi.Link{Href: "/v3/roles?page=1"},
			},
			Resources: []capi.Role{
				{
					Resource: capi.Resource{
						GUID:      "role-guid-1",
						CreatedAt: now,
						UpdatedAt: now,
					},
					Type: "organization_auditor",
					Relationships: capi.RoleRelationships{
						User: capi.Relationship{
							Data: &capi.RelationshipData{
								GUID: "user-guid-1",
							},
						},
						Organization: &capi.Relationship{
							Data: &capi.RelationshipData{
								GUID: "org-guid",
							},
						},
					},
				},
				{
					Resource: capi.Resource{
						GUID:      "role-guid-2",
						CreatedAt: now,
						UpdatedAt: now,
					},
					Type: "organization_manager",
					Relationships: capi.RoleRelationships{
						User: capi.Relationship{
							Data: &capi.RelationshipData{
								GUID: "user-guid-2",
							},
						},
						Organization: &capi.Relationship{
							Data: &capi.RelationshipData{
								GUID: "org-guid",
							},
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	roles := NewRolesClient(client.httpClient)

	params := &capi.QueryParams{
		Filters: map[string][]string{
			"types":              {"organization_auditor", "organization_manager"},
			"organization_guids": {"org-guid"},
		},
	}

	list, err := roles.List(context.Background(), params)
	require.NoError(t, err)
	assert.NotNil(t, list)
	assert.Equal(t, 2, list.Pagination.TotalResults)
	assert.Len(t, list.Resources, 2)
	assert.Equal(t, "role-guid-1", list.Resources[0].GUID)
	assert.Equal(t, "organization_auditor", list.Resources[0].Type)
	assert.Equal(t, "role-guid-2", list.Resources[1].GUID)
	assert.Equal(t, "organization_manager", list.Resources[1].Type)
}

func TestRolesClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/roles/role-guid", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	roles := NewRolesClient(client.httpClient)

	err := roles.Delete(context.Background(), "role-guid")
	require.NoError(t, err)
}

func TestRolesClient_GetNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/roles/role-guid", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	roles := NewRolesClient(client.httpClient)

	role, err := roles.Get(context.Background(), "role-guid")
	assert.Error(t, err)
	assert.Nil(t, role)
}
