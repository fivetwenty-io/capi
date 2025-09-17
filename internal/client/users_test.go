package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/fivetwenty-io/capi/v3/internal/client"
	internalhttp "github.com/fivetwenty-io/capi/v3/internal/http"
	"github.com/fivetwenty-io/capi/v3/pkg/capi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsersClient_Create_WithGUID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/users", request.URL.Path)
		assert.Equal(t, "POST", request.Method)

		var req capi.UserCreateRequest

		err := json.NewDecoder(request.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "user-guid", req.GUID)

		user := capi.User{
			Resource: capi.Resource{
				GUID:      "user-guid",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Links: capi.Links{
					"self": capi.Link{
						Href: "/v3/users/user-guid",
					},
				},
			},
			Username:         "test-user",
			PresentationName: "test-user",
			Origin:           "uaa",
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(writer).Encode(user)
	}))
	defer server.Close()

	httpClient := internalhttp.NewClient(server.URL, nil)
	users := NewUsersClient(httpClient)

	req := &capi.UserCreateRequest{
		GUID: "user-guid",
	}

	user, err := users.Create(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "user-guid", user.GUID)
	assert.Equal(t, "test-user", user.Username)
	assert.Equal(t, "uaa", user.Origin)
}

func TestUsersClient_Create_WithUsernameAndOrigin(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/users", request.URL.Path)
		assert.Equal(t, "POST", request.Method)

		var req capi.UserCreateRequest

		err := json.NewDecoder(request.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "test-user", req.Username)
		assert.Equal(t, "ldap", req.Origin)

		user := capi.User{
			Resource: capi.Resource{
				GUID:      "generated-guid",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Links: capi.Links{
					"self": capi.Link{
						Href: "/v3/users/generated-guid",
					},
				},
			},
			Username:         "test-user",
			PresentationName: "test-user",
			Origin:           "ldap",
			Metadata: &capi.Metadata{
				Labels:      map[string]string{},
				Annotations: map[string]string{},
			},
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(writer).Encode(user)
	}))
	defer server.Close()

	httpClient := internalhttp.NewClient(server.URL, nil)
	users := NewUsersClient(httpClient)

	req := &capi.UserCreateRequest{
		Username: "test-user",
		Origin:   "ldap",
		Metadata: &capi.Metadata{
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
	}

	user, err := users.Create(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "generated-guid", user.GUID)
	assert.Equal(t, "test-user", user.Username)
	assert.Equal(t, "ldap", user.Origin)
}

func TestUsersClient_Get(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/users/user-guid", request.URL.Path)
		assert.Equal(t, "GET", request.Method)

		user := capi.User{
			Resource: capi.Resource{
				GUID:      "user-guid",
				CreatedAt: time.Now().Add(-time.Hour),
				UpdatedAt: time.Now().Add(-30 * time.Minute),
				Links: capi.Links{
					"self": capi.Link{
						Href: "/v3/users/user-guid",
					},
				},
			},
			Username:         "test-user",
			PresentationName: "test-user",
			Origin:           "uaa",
			Metadata: &capi.Metadata{
				Labels: map[string]string{
					"environment": "production",
				},
				Annotations: map[string]string{
					"note": "admin user",
				},
			},
		}

		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(user)
	}))
	defer server.Close()

	httpClient := internalhttp.NewClient(server.URL, nil)
	users := NewUsersClient(httpClient)

	user, err := users.Get(context.Background(), "user-guid")
	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "user-guid", user.GUID)
	assert.Equal(t, "test-user", user.Username)
	assert.Equal(t, "uaa", user.Origin)
	assert.Equal(t, "production", user.Metadata.Labels["environment"])
}

func TestUsersClient_List(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/users", request.URL.Path)
		assert.Equal(t, "GET", request.Method)
		assert.Equal(t, "test-user", request.URL.Query().Get("usernames"))
		assert.Equal(t, "2", request.URL.Query().Get("per_page"))

		response := capi.ListResponse[capi.User]{
			Pagination: capi.Pagination{
				TotalResults: 2,
				TotalPages:   1,
				First: capi.Link{
					Href: "/v3/users?page=1&per_page=2",
				},
				Last: capi.Link{
					Href: "/v3/users?page=1&per_page=2",
				},
			},
			Resources: []capi.User{
				{
					Resource: capi.Resource{
						GUID:      "user-guid-1",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
						Links: capi.Links{
							"self": capi.Link{
								Href: "/v3/users/user-guid-1",
							},
						},
					},
					Username:         "test-user",
					PresentationName: "test-user",
					Origin:           "uaa",
				},
				{
					Resource: capi.Resource{
						GUID:      "client-id",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
						Links: capi.Links{
							"self": capi.Link{
								Href: "/v3/users/client-id",
							},
						},
					},
					Username:         "",
					PresentationName: "client-id",
					Origin:           "",
				},
			},
		}

		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(response)
	}))
	defer server.Close()

	httpClient := internalhttp.NewClient(server.URL, nil)
	users := NewUsersClient(httpClient)

	params := &capi.QueryParams{
		PerPage: 2,
		Filters: map[string][]string{
			"usernames": {"test-user"},
		},
	}

	list, err := users.List(context.Background(), params)
	require.NoError(t, err)
	assert.NotNil(t, list)
	assert.Equal(t, 2, list.Pagination.TotalResults)
	assert.Len(t, list.Resources, 2)
	assert.Equal(t, "test-user", list.Resources[0].Username)
	assert.Empty(t, list.Resources[1].Username) // UAA client
	assert.Equal(t, "client-id", list.Resources[1].PresentationName)
}

func TestUsersClient_Update(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/users/user-guid", request.URL.Path)
		assert.Equal(t, "PATCH", request.Method)

		var req capi.UserUpdateRequest

		err := json.NewDecoder(request.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "staging", req.Metadata.Labels["environment"])
		assert.Equal(t, "updated note", req.Metadata.Annotations["note"])

		user := capi.User{
			Resource: capi.Resource{
				GUID:      "user-guid",
				CreatedAt: time.Now().Add(-time.Hour),
				UpdatedAt: time.Now(),
				Links: capi.Links{
					"self": capi.Link{
						Href: "/v3/users/user-guid",
					},
				},
			},
			Username:         "test-user",
			PresentationName: "test-user",
			Origin:           "uaa",
			Metadata: &capi.Metadata{
				Labels: map[string]string{
					"environment": "staging",
				},
				Annotations: map[string]string{
					"note": "updated note",
				},
			},
		}

		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(user)
	}))
	defer server.Close()

	httpClient := internalhttp.NewClient(server.URL, nil)
	users := NewUsersClient(httpClient)

	req := &capi.UserUpdateRequest{
		Metadata: &capi.Metadata{
			Labels: map[string]string{
				"environment": "staging",
			},
			Annotations: map[string]string{
				"note": "updated note",
			},
		},
	}

	user, err := users.Update(context.Background(), "user-guid", req)
	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "user-guid", user.GUID)
	assert.Equal(t, "staging", user.Metadata.Labels["environment"])
	assert.Equal(t, "updated note", user.Metadata.Annotations["note"])
}

func TestUsersClient_Delete(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/users/user-guid", request.URL.Path)
		assert.Equal(t, "DELETE", request.Method)

		job := capi.Job{
			Resource: capi.Resource{
				GUID:      "job-guid",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Links: capi.Links{
					"self": capi.Link{
						Href: "/v3/jobs/job-guid",
					},
				},
			},
			Operation: "user.delete",
			State:     "PROCESSING",
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Header().Set("Location", "https://api.example.org/v3/jobs/job-guid")
		writer.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(writer).Encode(job)
	}))
	defer server.Close()

	httpClient := internalhttp.NewClient(server.URL, nil)
	users := NewUsersClient(httpClient)

	job, err := users.Delete(context.Background(), "user-guid")
	require.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, "job-guid", job.GUID)
	assert.Equal(t, "user.delete", job.Operation)
	assert.Equal(t, "PROCESSING", job.State)
}
