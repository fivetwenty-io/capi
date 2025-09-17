package capi_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/fivetwenty-io/capi/v3/pkg/capi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResource_JSONMarshaling(t *testing.T) {
	t.Parallel()

	resource := capi.Resource{
		GUID:      "test-guid",
		CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		Links: capi.Links{
			"self": capi.Link{
				Href: "https://api.example.org/v3/resources/test-guid",
			},
			"related": capi.Link{
				Href:   "https://api.example.org/v3/related",
				Method: "POST",
			},
		},
	}

	data, err := json.Marshal(resource)
	require.NoError(t, err)

	var decoded capi.Resource

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, resource.GUID, decoded.GUID)
	assert.Equal(t, resource.CreatedAt.Unix(), decoded.CreatedAt.Unix())
	assert.Equal(t, resource.UpdatedAt.Unix(), decoded.UpdatedAt.Unix())
	assert.Equal(t, resource.Links["self"].Href, decoded.Links["self"].Href)
	assert.Equal(t, resource.Links["related"].Method, decoded.Links["related"].Method)
}

func TestMetadata_JSONMarshaling(t *testing.T) {
	t.Parallel()

	metadata := capi.Metadata{
		Labels: map[string]string{
			"environment": "production",
			"team":        "platform",
		},
		Annotations: map[string]string{
			"version": "1.0.0",
			"owner":   "team@example.com",
		},
	}

	data, err := json.Marshal(metadata)
	require.NoError(t, err)

	var decoded capi.Metadata

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, metadata.Labels, decoded.Labels)
	assert.Equal(t, metadata.Annotations, decoded.Annotations)
}

func TestRelationship_JSONMarshaling(t *testing.T) {
	t.Parallel()
	t.Run("with data", func(t *testing.T) {
		t.Parallel()

		rel := capi.Relationship{
			Data: &capi.RelationshipData{
				GUID: "related-guid",
			},
		}

		data, err := json.Marshal(rel)
		require.NoError(t, err)

		var decoded capi.Relationship

		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		require.NotNil(t, decoded.Data)
		assert.Equal(t, "related-guid", decoded.Data.GUID)
	})

	t.Run("without data", func(t *testing.T) {
		t.Parallel()

		rel := capi.Relationship{}

		data, err := json.Marshal(rel)
		require.NoError(t, err)

		var decoded capi.Relationship

		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Nil(t, decoded.Data)
	})
}

func TestToManyRelationship_JSONMarshaling(t *testing.T) {
	t.Parallel()

	rel := capi.ToManyRelationship{
		Data: []capi.RelationshipData{
			{GUID: "guid-1"},
			{GUID: "guid-2"},
			{GUID: "guid-3"},
		},
	}

	data, err := json.Marshal(rel)
	require.NoError(t, err)

	var decoded capi.ToManyRelationship

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Len(t, decoded.Data, 3)
	assert.Equal(t, "guid-1", decoded.Data[0].GUID)
	assert.Equal(t, "guid-2", decoded.Data[1].GUID)
	assert.Equal(t, "guid-3", decoded.Data[2].GUID)
}

func TestPagination_JSONMarshaling(t *testing.T) {
	t.Parallel()

	pagination := capi.Pagination{
		TotalResults: 100,
		TotalPages:   10,
		First: capi.Link{
			Href: "https://api.example.org/v3/resources?page=1",
		},
		Last: capi.Link{
			Href: "https://api.example.org/v3/resources?page=10",
		},
		Next: &capi.Link{
			Href: "https://api.example.org/v3/resources?page=2",
		},
		Previous: nil,
	}

	data, err := json.Marshal(pagination)
	require.NoError(t, err)

	var decoded capi.Pagination

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, pagination.TotalResults, decoded.TotalResults)
	assert.Equal(t, pagination.TotalPages, decoded.TotalPages)
	assert.Equal(t, pagination.First.Href, decoded.First.Href)
	assert.Equal(t, pagination.Last.Href, decoded.Last.Href)
	require.NotNil(t, decoded.Next)
	assert.Equal(t, pagination.Next.Href, decoded.Next.Href)
	assert.Nil(t, decoded.Previous)
}

func TestListResponse_JSONMarshaling(t *testing.T) {
	t.Parallel()

	type TestResource struct {
		capi.Resource

		Name string `json:"name"`
	}

	listResp := capi.ListResponse[TestResource]{
		Pagination: capi.Pagination{
			TotalResults: 2,
			TotalPages:   1,
			First: capi.Link{
				Href: "https://api.example.org/v3/test?page=1",
			},
			Last: capi.Link{
				Href: "https://api.example.org/v3/test?page=1",
			},
		},
		Resources: []TestResource{
			{
				Resource: capi.Resource{
					GUID: "guid-1",
				},
				Name: "test-1",
			},
			{
				Resource: capi.Resource{
					GUID: "guid-2",
				},
				Name: "test-2",
			},
		},
	}

	data, err := json.Marshal(listResp)
	require.NoError(t, err)

	var decoded capi.ListResponse[TestResource]

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, listResp.Pagination.TotalResults, decoded.Pagination.TotalResults)
	assert.Len(t, decoded.Resources, 2)
	assert.Equal(t, "guid-1", decoded.Resources[0].GUID)
	assert.Equal(t, "test-1", decoded.Resources[0].Name)
	assert.Equal(t, "guid-2", decoded.Resources[1].GUID)
	assert.Equal(t, "test-2", decoded.Resources[1].Name)
}
