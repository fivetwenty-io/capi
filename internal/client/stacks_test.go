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

func TestStacksClient_Create(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/stacks", request.URL.Path)
		assert.Equal(t, "POST", request.Method)

		var requestBody capi.StackCreateRequest

		err := json.NewDecoder(request.Body).Decode(&requestBody)
		assert.NoError(t, err)

		assert.Equal(t, "cflinuxfs4", requestBody.Name)
		assert.Equal(t, "Ubuntu Jammy Stack", requestBody.Description)

		now := time.Now()
		stack := capi.Stack{
			Resource: capi.Resource{
				GUID:      "stack-guid",
				CreatedAt: now,
				UpdatedAt: now,
			},
			Name:             requestBody.Name,
			Description:      requestBody.Description,
			BuildRootfsImage: "cloudfoundry/cflinuxfs4",
			RunRootfsImage:   "cloudfoundry/cflinuxfs4",
			Default:          true,
			Metadata:         requestBody.Metadata,
			Links: capi.Links{
				"self": capi.Link{
					Href: "https://api.example.org/v3/stacks/stack-guid",
				},
			},
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(writer).Encode(stack)
	}))
	defer server.Close()

	httpClient := internalhttp.NewClient(server.URL, nil)
	stacks := NewStacksClient(httpClient)

	request := &capi.StackCreateRequest{
		Name:        "cflinuxfs4",
		Description: "Ubuntu Jammy Stack",
		Metadata: &capi.Metadata{
			Labels: map[string]string{
				"env": "production",
			},
		},
	}

	stack, err := stacks.Create(context.Background(), request)
	require.NoError(t, err)
	assert.NotNil(t, stack)
	assert.Equal(t, "stack-guid", stack.GUID)
	assert.Equal(t, "cflinuxfs4", stack.Name)
	assert.Equal(t, "Ubuntu Jammy Stack", stack.Description)
	assert.True(t, stack.Default)
}

func TestStacksClient_Get(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/stacks/stack-guid", request.URL.Path)
		assert.Equal(t, "GET", request.Method)

		now := time.Now()
		stack := capi.Stack{
			Resource: capi.Resource{
				GUID:      "stack-guid",
				CreatedAt: now,
				UpdatedAt: now,
			},
			Name:             "cflinuxfs3",
			Description:      "Ubuntu Bionic Stack",
			BuildRootfsImage: "cloudfoundry/cflinuxfs3",
			RunRootfsImage:   "cloudfoundry/cflinuxfs3",
			Default:          false,
		}

		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(stack)
	}))
	defer server.Close()

	httpClient := internalhttp.NewClient(server.URL, nil)
	stacks := NewStacksClient(httpClient)

	stack, err := stacks.Get(context.Background(), "stack-guid")
	require.NoError(t, err)
	assert.NotNil(t, stack)
	assert.Equal(t, "stack-guid", stack.GUID)
	assert.Equal(t, "cflinuxfs3", stack.Name)
	assert.Equal(t, "Ubuntu Bionic Stack", stack.Description)
	assert.False(t, stack.Default)
}

func TestStacksClient_List(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/stacks", request.URL.Path)
		assert.Equal(t, "GET", request.Method)
		assert.Equal(t, "cflinuxfs3,cflinuxfs4", request.URL.Query().Get("names"))

		now := time.Now()
		response := capi.ListResponse[capi.Stack]{
			Pagination: capi.Pagination{
				TotalResults: 2,
				TotalPages:   1,
				First:        capi.Link{Href: "/v3/stacks?page=1"},
				Last:         capi.Link{Href: "/v3/stacks?page=1"},
			},
			Resources: []capi.Stack{
				{
					Resource: capi.Resource{
						GUID:      "stack-guid-1",
						CreatedAt: now,
						UpdatedAt: now,
					},
					Name:             "cflinuxfs3",
					Description:      "Ubuntu Bionic Stack",
					BuildRootfsImage: "cloudfoundry/cflinuxfs3",
					RunRootfsImage:   "cloudfoundry/cflinuxfs3",
					Default:          false,
				},
				{
					Resource: capi.Resource{
						GUID:      "stack-guid-2",
						CreatedAt: now,
						UpdatedAt: now,
					},
					Name:             "cflinuxfs4",
					Description:      "Ubuntu Jammy Stack",
					BuildRootfsImage: "cloudfoundry/cflinuxfs4",
					RunRootfsImage:   "cloudfoundry/cflinuxfs4",
					Default:          true,
				},
			},
		}

		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(response)
	}))
	defer server.Close()

	httpClient := internalhttp.NewClient(server.URL, nil)
	stacks := NewStacksClient(httpClient)

	params := &capi.QueryParams{
		Filters: map[string][]string{
			"names": {"cflinuxfs3", "cflinuxfs4"},
		},
	}

	list, err := stacks.List(context.Background(), params)
	require.NoError(t, err)
	assert.NotNil(t, list)
	assert.Equal(t, 2, list.Pagination.TotalResults)
	assert.Len(t, list.Resources, 2)
	assert.Equal(t, "stack-guid-1", list.Resources[0].GUID)
	assert.Equal(t, "cflinuxfs3", list.Resources[0].Name)
	assert.Equal(t, "stack-guid-2", list.Resources[1].GUID)
	assert.Equal(t, "cflinuxfs4", list.Resources[1].Name)
}

func TestStacksClient_Update(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/stacks/stack-guid", request.URL.Path)
		assert.Equal(t, "PATCH", request.Method)

		var requestBody capi.StackUpdateRequest

		err := json.NewDecoder(request.Body).Decode(&requestBody)
		assert.NoError(t, err)

		assert.NotNil(t, requestBody.Metadata)
		assert.Equal(t, "true", requestBody.Metadata.Labels["updated"])

		now := time.Now()
		stack := capi.Stack{
			Resource: capi.Resource{
				GUID:      "stack-guid",
				CreatedAt: now,
				UpdatedAt: now,
			},
			Name:             "cflinuxfs4",
			Description:      "Ubuntu Jammy Stack",
			BuildRootfsImage: "cloudfoundry/cflinuxfs4",
			RunRootfsImage:   "cloudfoundry/cflinuxfs4",
			Default:          true,
			Metadata:         requestBody.Metadata,
		}

		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(stack)
	}))
	defer server.Close()

	httpClient := internalhttp.NewClient(server.URL, nil)
	stacks := NewStacksClient(httpClient)

	request := &capi.StackUpdateRequest{
		Metadata: &capi.Metadata{
			Labels: map[string]string{
				"updated": "true",
			},
			Annotations: map[string]string{
				"note": "Updated stack metadata",
			},
		},
	}

	stack, err := stacks.Update(context.Background(), "stack-guid", request)
	require.NoError(t, err)
	assert.NotNil(t, stack)
	assert.Equal(t, "stack-guid", stack.GUID)
	assert.NotNil(t, stack.Metadata)
	assert.Equal(t, "true", stack.Metadata.Labels["updated"])
}

func TestStacksClient_Delete(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/stacks/stack-guid", request.URL.Path)
		assert.Equal(t, "DELETE", request.Method)

		writer.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	httpClient := internalhttp.NewClient(server.URL, nil)
	stacks := NewStacksClient(httpClient)

	err := stacks.Delete(context.Background(), "stack-guid")
	require.NoError(t, err)
}

func TestStacksClient_ListApps(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/stacks/stack-guid/apps", request.URL.Path)
		assert.Equal(t, "GET", request.Method)

		now := time.Now()
		response := capi.ListResponse[capi.App]{
			Pagination: capi.Pagination{
				TotalResults: 2,
				TotalPages:   1,
				First:        capi.Link{Href: "/v3/stacks/stack-guid/apps?page=1"},
				Last:         capi.Link{Href: "/v3/stacks/stack-guid/apps?page=1"},
			},
			Resources: []capi.App{
				{
					Resource: capi.Resource{
						GUID:      "app-guid-1",
						CreatedAt: now,
						UpdatedAt: now,
					},
					Name:  "app1",
					State: "STARTED",
				},
				{
					Resource: capi.Resource{
						GUID:      "app-guid-2",
						CreatedAt: now,
						UpdatedAt: now,
					},
					Name:  "app2",
					State: "STOPPED",
				},
			},
		}

		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(response)
	}))
	defer server.Close()

	httpClient := internalhttp.NewClient(server.URL, nil)
	stacks := NewStacksClient(httpClient)

	list, err := stacks.ListApps(context.Background(), "stack-guid", nil)
	require.NoError(t, err)
	assert.NotNil(t, list)
	assert.Equal(t, 2, list.Pagination.TotalResults)
	assert.Len(t, list.Resources, 2)
	assert.Equal(t, "app-guid-1", list.Resources[0].GUID)
	assert.Equal(t, "app1", list.Resources[0].Name)
	assert.Equal(t, "app-guid-2", list.Resources[1].GUID)
	assert.Equal(t, "app2", list.Resources[1].Name)
}

func TestStacksClient_GetNotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/stacks/stack-guid", request.URL.Path)
		assert.Equal(t, "GET", request.Method)

		writer.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	httpClient := internalhttp.NewClient(server.URL, nil)
	stacks := NewStacksClient(httpClient)

	stack, err := stacks.Get(context.Background(), "stack-guid")
	require.Error(t, err)
	assert.Nil(t, stack)
}
