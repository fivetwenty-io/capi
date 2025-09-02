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

func TestFeatureFlagsClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/feature_flags/my_feature_flag", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		now := time.Now()
		customError := "error message the user sees"
		ff := capi.FeatureFlag{
			Name:               "my_feature_flag",
			Enabled:            true,
			UpdatedAt:          &now,
			CustomErrorMessage: &customError,
			Links: capi.Links{
				"self": capi.Link{
					Href: "/v3/feature_flags/my_feature_flag",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ff)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	featureFlags := NewFeatureFlagsClient(client.httpClient)

	ff, err := featureFlags.Get(context.Background(), "my_feature_flag")
	require.NoError(t, err)
	assert.NotNil(t, ff)
	assert.Equal(t, "my_feature_flag", ff.Name)
	assert.True(t, ff.Enabled)
	assert.NotNil(t, ff.UpdatedAt)
	assert.NotNil(t, ff.CustomErrorMessage)
	assert.Equal(t, "error message the user sees", *ff.CustomErrorMessage)
}

func TestFeatureFlagsClient_Get_NotConfigured(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/feature_flags/app_scaling", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		ff := capi.FeatureFlag{
			Name:               "app_scaling",
			Enabled:            true,
			UpdatedAt:          nil, // Not configured flags have null updated_at
			CustomErrorMessage: nil, // Not configured flags have null custom_error_message
			Links: capi.Links{
				"self": capi.Link{
					Href: "/v3/feature_flags/app_scaling",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ff)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	featureFlags := NewFeatureFlagsClient(client.httpClient)

	ff, err := featureFlags.Get(context.Background(), "app_scaling")
	require.NoError(t, err)
	assert.NotNil(t, ff)
	assert.Equal(t, "app_scaling", ff.Name)
	assert.True(t, ff.Enabled)
	assert.Nil(t, ff.UpdatedAt)
	assert.Nil(t, ff.CustomErrorMessage)
}

func TestFeatureFlagsClient_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/feature_flags", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "2", r.URL.Query().Get("per_page"))
		assert.Equal(t, "1", r.URL.Query().Get("page"))

		now := time.Now()
		customError := "error message the user sees"
		response := capi.ListResponse[capi.FeatureFlag]{
			Pagination: capi.Pagination{
				TotalResults: 3,
				TotalPages:   2,
				First:        capi.Link{Href: "/v3/feature_flags?page=1&per_page=2"},
				Last:         capi.Link{Href: "/v3/feature_flags?page=2&per_page=2"},
				Next:         &capi.Link{Href: "/v3/feature_flags?page=2&per_page=2"},
			},
			Resources: []capi.FeatureFlag{
				{
					Name:               "my_feature_flag",
					Enabled:            true,
					UpdatedAt:          &now,
					CustomErrorMessage: &customError,
					Links: capi.Links{
						"self": capi.Link{
							Href: "/v3/feature_flags/my_feature_flag",
						},
					},
				},
				{
					Name:               "my_second_feature_flag",
					Enabled:            false,
					UpdatedAt:          nil,
					CustomErrorMessage: nil,
					Links: capi.Links{
						"self": capi.Link{
							Href: "/v3/feature_flags/my_second_feature_flag",
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	featureFlags := NewFeatureFlagsClient(client.httpClient)

	params := &capi.QueryParams{
		Page:    1,
		PerPage: 2,
	}

	list, err := featureFlags.List(context.Background(), params)
	require.NoError(t, err)
	assert.NotNil(t, list)
	assert.Equal(t, 3, list.Pagination.TotalResults)
	assert.Equal(t, 2, list.Pagination.TotalPages)
	assert.Len(t, list.Resources, 2)
	assert.Equal(t, "my_feature_flag", list.Resources[0].Name)
	assert.True(t, list.Resources[0].Enabled)
	assert.Equal(t, "my_second_feature_flag", list.Resources[1].Name)
	assert.False(t, list.Resources[1].Enabled)
}

func TestFeatureFlagsClient_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/feature_flags/my_feature_flag", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var request capi.FeatureFlagUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		require.NoError(t, err)

		assert.False(t, request.Enabled)
		assert.NotNil(t, request.CustomErrorMessage)
		assert.Equal(t, "custom error", *request.CustomErrorMessage)

		now := time.Now()
		ff := capi.FeatureFlag{
			Name:               "my_feature_flag",
			Enabled:            request.Enabled,
			UpdatedAt:          &now,
			CustomErrorMessage: request.CustomErrorMessage,
			Links: capi.Links{
				"self": capi.Link{
					Href: "/v3/feature_flags/my_feature_flag",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ff)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	featureFlags := NewFeatureFlagsClient(client.httpClient)

	customError := "custom error"
	request := &capi.FeatureFlagUpdateRequest{
		Enabled:            false,
		CustomErrorMessage: &customError,
	}

	ff, err := featureFlags.Update(context.Background(), "my_feature_flag", request)
	require.NoError(t, err)
	assert.NotNil(t, ff)
	assert.Equal(t, "my_feature_flag", ff.Name)
	assert.False(t, ff.Enabled)
	assert.NotNil(t, ff.UpdatedAt)
	assert.NotNil(t, ff.CustomErrorMessage)
	assert.Equal(t, "custom error", *ff.CustomErrorMessage)
}

func TestFeatureFlagsClient_Update_EnableOnly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/feature_flags/diego_docker", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var request capi.FeatureFlagUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		require.NoError(t, err)

		assert.True(t, request.Enabled)
		assert.Nil(t, request.CustomErrorMessage)

		now := time.Now()
		ff := capi.FeatureFlag{
			Name:               "diego_docker",
			Enabled:            request.Enabled,
			UpdatedAt:          &now,
			CustomErrorMessage: nil,
			Links: capi.Links{
				"self": capi.Link{
					Href: "/v3/feature_flags/diego_docker",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ff)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	featureFlags := NewFeatureFlagsClient(client.httpClient)

	request := &capi.FeatureFlagUpdateRequest{
		Enabled: true,
	}

	ff, err := featureFlags.Update(context.Background(), "diego_docker", request)
	require.NoError(t, err)
	assert.NotNil(t, ff)
	assert.Equal(t, "diego_docker", ff.Name)
	assert.True(t, ff.Enabled)
	assert.NotNil(t, ff.UpdatedAt)
	assert.Nil(t, ff.CustomErrorMessage)
}
