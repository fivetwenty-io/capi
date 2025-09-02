package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	internalhttp "github.com/fivetwenty-io/capi/v3/internal/http"
	"github.com/fivetwenty-io/capi/v3/pkg/capi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsolationSegmentsClient_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/isolation_segments", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var request capi.IsolationSegmentCreateRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		require.NoError(t, err)

		assert.Equal(t, "my-segment", request.Name)
		assert.NotNil(t, request.Metadata)
		assert.Equal(t, "value1", request.Metadata.Labels["key1"])

		now := time.Now()
		is := capi.IsolationSegment{
			Resource: capi.Resource{
				GUID:      "segment-guid",
				CreatedAt: now,
				UpdatedAt: now,
				Links: capi.Links{
					"self": capi.Link{
						Href: "/v3/isolation_segments/segment-guid",
					},
					"organizations": capi.Link{
						Href: "/v3/isolation_segments/segment-guid/organizations",
					},
				},
			},
			Name:     request.Name,
			Metadata: request.Metadata,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(is)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	isolationSegments := NewIsolationSegmentsClient(client.httpClient)

	request := &capi.IsolationSegmentCreateRequest{
		Name: "my-segment",
		Metadata: &capi.Metadata{
			Labels: map[string]string{
				"key1": "value1",
			},
			Annotations: map[string]string{},
		},
	}

	is, err := isolationSegments.Create(context.Background(), request)
	require.NoError(t, err)
	assert.NotNil(t, is)
	assert.Equal(t, "segment-guid", is.GUID)
	assert.Equal(t, "my-segment", is.Name)
	assert.Equal(t, "value1", is.Metadata.Labels["key1"])
}

func TestIsolationSegmentsClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/isolation_segments/segment-guid", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		now := time.Now()
		is := capi.IsolationSegment{
			Resource: capi.Resource{
				GUID:      "segment-guid",
				CreatedAt: now,
				UpdatedAt: now,
				Links: capi.Links{
					"self": capi.Link{
						Href: "/v3/isolation_segments/segment-guid",
					},
					"organizations": capi.Link{
						Href: "/v3/isolation_segments/segment-guid/organizations",
					},
				},
			},
			Name: "my-segment",
			Metadata: &capi.Metadata{
				Labels:      map[string]string{},
				Annotations: map[string]string{},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(is)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	isolationSegments := NewIsolationSegmentsClient(client.httpClient)

	is, err := isolationSegments.Get(context.Background(), "segment-guid")
	require.NoError(t, err)
	assert.NotNil(t, is)
	assert.Equal(t, "segment-guid", is.GUID)
	assert.Equal(t, "my-segment", is.Name)
}

func TestIsolationSegmentsClient_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/isolation_segments", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "segment1,segment2", r.URL.Query().Get("names"))
		assert.Equal(t, "org-guid", r.URL.Query().Get("organization_guids"))

		now := time.Now()
		response := capi.ListResponse[capi.IsolationSegment]{
			Pagination: capi.Pagination{
				TotalResults: 2,
				TotalPages:   1,
				First:        capi.Link{Href: "/v3/isolation_segments?page=1"},
				Last:         capi.Link{Href: "/v3/isolation_segments?page=1"},
			},
			Resources: []capi.IsolationSegment{
				{
					Resource: capi.Resource{
						GUID:      "segment-guid-1",
						CreatedAt: now,
						UpdatedAt: now,
					},
					Name: "segment1",
				},
				{
					Resource: capi.Resource{
						GUID:      "segment-guid-2",
						CreatedAt: now,
						UpdatedAt: now,
					},
					Name: "segment2",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	isolationSegments := NewIsolationSegmentsClient(client.httpClient)

	params := &capi.QueryParams{
		Filters: map[string][]string{
			"names":              {"segment1", "segment2"},
			"organization_guids": {"org-guid"},
		},
	}

	list, err := isolationSegments.List(context.Background(), params)
	require.NoError(t, err)
	assert.NotNil(t, list)
	assert.Equal(t, 2, list.Pagination.TotalResults)
	assert.Len(t, list.Resources, 2)
	assert.Equal(t, "segment1", list.Resources[0].Name)
	assert.Equal(t, "segment2", list.Resources[1].Name)
}

func TestIsolationSegmentsClient_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/isolation_segments/segment-guid", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var request capi.IsolationSegmentUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		require.NoError(t, err)

		assert.NotNil(t, request.Name)
		assert.Equal(t, "updated-segment", *request.Name)
		assert.NotNil(t, request.Metadata)
		assert.Equal(t, "value2", request.Metadata.Labels["key2"])

		now := time.Now()
		is := capi.IsolationSegment{
			Resource: capi.Resource{
				GUID:      "segment-guid",
				CreatedAt: now,
				UpdatedAt: now,
			},
			Name:     *request.Name,
			Metadata: request.Metadata,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(is)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	isolationSegments := NewIsolationSegmentsClient(client.httpClient)

	name := "updated-segment"
	request := &capi.IsolationSegmentUpdateRequest{
		Name: &name,
		Metadata: &capi.Metadata{
			Labels: map[string]string{
				"key2": "value2",
			},
			Annotations: map[string]string{},
		},
	}

	is, err := isolationSegments.Update(context.Background(), "segment-guid", request)
	require.NoError(t, err)
	assert.NotNil(t, is)
	assert.Equal(t, "segment-guid", is.GUID)
	assert.Equal(t, "updated-segment", is.Name)
	assert.Equal(t, "value2", is.Metadata.Labels["key2"])
}

func TestIsolationSegmentsClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/isolation_segments/segment-guid", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	isolationSegments := NewIsolationSegmentsClient(client.httpClient)

	err := isolationSegments.Delete(context.Background(), "segment-guid")
	require.NoError(t, err)
}

func TestIsolationSegmentsClient_EntitleOrganizations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/isolation_segments/segment-guid/relationships/organizations", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var request capi.IsolationSegmentEntitleOrganizationsRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		require.NoError(t, err)

		assert.Len(t, request.Data, 2)
		assert.Equal(t, "org-guid-1", request.Data[0].GUID)
		assert.Equal(t, "org-guid-2", request.Data[1].GUID)

		response := capi.ToManyRelationship(request)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	isolationSegments := NewIsolationSegmentsClient(client.httpClient)

	relationship, err := isolationSegments.EntitleOrganizations(context.Background(), "segment-guid", []string{"org-guid-1", "org-guid-2"})
	require.NoError(t, err)
	assert.NotNil(t, relationship)
	assert.Len(t, relationship.Data, 2)
	assert.Equal(t, "org-guid-1", relationship.Data[0].GUID)
}

func TestIsolationSegmentsClient_RevokeOrganization(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/isolation_segments/segment-guid/relationships/organizations/org-guid", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	isolationSegments := NewIsolationSegmentsClient(client.httpClient)

	err := isolationSegments.RevokeOrganization(context.Background(), "segment-guid", "org-guid")
	require.NoError(t, err)
}

func TestIsolationSegmentsClient_ListOrganizations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/isolation_segments/segment-guid/organizations", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		now := time.Now()
		response := capi.ListResponse[capi.Organization]{
			Pagination: capi.Pagination{
				TotalResults: 2,
				TotalPages:   1,
				First:        capi.Link{Href: "/v3/isolation_segments/segment-guid/organizations?page=1"},
				Last:         capi.Link{Href: "/v3/isolation_segments/segment-guid/organizations?page=1"},
			},
			Resources: []capi.Organization{
				{
					Resource: capi.Resource{
						GUID:      "org-guid-1",
						CreatedAt: now,
						UpdatedAt: now,
					},
					Name:      "org1",
					Suspended: false,
				},
				{
					Resource: capi.Resource{
						GUID:      "org-guid-2",
						CreatedAt: now,
						UpdatedAt: now,
					},
					Name:      "org2",
					Suspended: false,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	isolationSegments := NewIsolationSegmentsClient(client.httpClient)

	list, err := isolationSegments.ListOrganizations(context.Background(), "segment-guid")
	require.NoError(t, err)
	assert.NotNil(t, list)
	assert.Equal(t, 2, list.Pagination.TotalResults)
	assert.Len(t, list.Resources, 2)
	assert.Equal(t, "org1", list.Resources[0].Name)
	assert.Equal(t, "org2", list.Resources[1].Name)
}

func TestIsolationSegmentsClient_ListSpaces(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/isolation_segments/segment-guid/relationships/spaces", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		now := time.Now()
		response := capi.ListResponse[capi.Space]{
			Pagination: capi.Pagination{
				TotalResults: 2,
				TotalPages:   1,
				First:        capi.Link{Href: "/v3/isolation_segments/segment-guid/relationships/spaces?page=1"},
				Last:         capi.Link{Href: "/v3/isolation_segments/segment-guid/relationships/spaces?page=1"},
			},
			Resources: []capi.Space{
				{
					Resource: capi.Resource{
						GUID:      "space-guid-1",
						CreatedAt: now,
						UpdatedAt: now,
					},
					Name: "space1",
				},
				{
					Resource: capi.Resource{
						GUID:      "space-guid-2",
						CreatedAt: now,
						UpdatedAt: now,
					},
					Name: "space2",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	isolationSegments := NewIsolationSegmentsClient(client.httpClient)

	list, err := isolationSegments.ListSpaces(context.Background(), "segment-guid")
	require.NoError(t, err)
	assert.NotNil(t, list)
	assert.Equal(t, 2, list.Pagination.TotalResults)
	assert.Len(t, list.Resources, 2)
	assert.Equal(t, "space1", list.Resources[0].Name)
	assert.Equal(t, "space2", list.Resources[1].Name)
}
