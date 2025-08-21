package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fivetwenty-io/capi-client/pkg/capi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("requires API endpoint", func(t *testing.T) {
		config := &capi.Config{}
		_, err := New(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API endpoint is required")
	})

	t.Run("creates client with access token", func(t *testing.T) {
		config := &capi.Config{
			APIEndpoint: "https://api.example.com",
			AccessToken: "test-token",
		}

		client, err := New(config)
		require.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("creates client with client credentials", func(t *testing.T) {
		config := &capi.Config{
			APIEndpoint:  "https://api.example.com",
			ClientID:     "client-id",
			ClientSecret: "client-secret",
		}

		client, err := New(config)
		require.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("creates client with username/password", func(t *testing.T) {
		config := &capi.Config{
			APIEndpoint: "https://api.example.com",
			Username:    "user",
			Password:    "pass",
		}

		client, err := New(config)
		require.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("creates client without authentication", func(t *testing.T) {
		config := &capi.Config{
			APIEndpoint: "https://api.example.com",
		}

		client, err := New(config)
		require.NoError(t, err)
		assert.NotNil(t, client)
	})
}

func TestClient_GetInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/info", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		info := capi.Info{
			Build:       "1.2.3",
			Name:        "Test CF",
			Version:     3,
			Description: "Test Cloud Foundry",
			CFOnK8s:     false,
			CLIVersion: capi.CLIVersion{
				Minimum:     "1.0.0",
				Recommended: "2.0.0",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
	}))
	defer server.Close()

	config := &capi.Config{
		APIEndpoint: server.URL,
	}

	client, err := New(config)
	require.NoError(t, err)

	info, err := client.GetInfo(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, "1.2.3", info.Build)
	assert.Equal(t, "Test CF", info.Name)
	assert.Equal(t, 3, info.Version)
}

func TestClient_GetRootInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		rootInfo := capi.RootInfo{
			Links: capi.Links{
				"self": capi.Link{
					Href: r.Host,
				},
				"cloud_controller_v2": capi.Link{
					Href: r.Host + "/v2",
				},
				"cloud_controller_v3": capi.Link{
					Href: r.Host + "/v3",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rootInfo)
	}))
	defer server.Close()

	config := &capi.Config{
		APIEndpoint: server.URL,
	}

	client, err := New(config)
	require.NoError(t, err)

	rootInfo, err := client.GetRootInfo(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, rootInfo)
	assert.NotNil(t, rootInfo.Links)
}

func TestClient_GetUsageSummary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/info/usage_summary", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		summary := capi.UsageSummary{
			UsageSummary: capi.UsageSummaryData{
				StartedInstances: 10,
				MemoryInMB:       2048,
			},
			Links: capi.Links{
				"self": capi.Link{
					Href: r.Host + "/v3/info/usage_summary",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(summary)
	}))
	defer server.Close()

	config := &capi.Config{
		APIEndpoint: server.URL,
	}

	client, err := New(config)
	require.NoError(t, err)

	summary, err := client.GetUsageSummary(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, 10, summary.UsageSummary.StartedInstances)
	assert.Equal(t, 2048, summary.UsageSummary.MemoryInMB)
}

func TestClient_ClearBuildpackCache(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/admin/actions/clear_buildpack_cache", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		job := capi.Job{
			Resource: capi.Resource{
				GUID: "job-guid",
			},
			Operation: "clear_buildpack_cache",
			State:     "PROCESSING",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(job)
	}))
	defer server.Close()

	config := &capi.Config{
		APIEndpoint: server.URL,
	}

	client, err := New(config)
	require.NoError(t, err)

	job, err := client.ClearBuildpackCache(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, "job-guid", job.GUID)
	assert.Equal(t, "clear_buildpack_cache", job.Operation)
	assert.Equal(t, "PROCESSING", job.State)
}

func TestClient_ResourceAccessors(t *testing.T) {
	config := &capi.Config{
		APIEndpoint: "https://api.example.com",
	}

	client, err := New(config)
	require.NoError(t, err)

	// Test that all accessors return their respective clients (or nil for now)
	assert.NotNil(t, client.Apps())                      // Apps client is implemented
	assert.NotNil(t, client.Organizations())             // Organizations client is implemented
	assert.NotNil(t, client.Spaces())                    // Spaces client is implemented
	assert.NotNil(t, client.Domains())                   // Domains client is implemented
	assert.NotNil(t, client.Routes())                    // Routes client is implemented
	assert.NotNil(t, client.ServiceBrokers())            // ServiceBrokers client is implemented
	assert.NotNil(t, client.ServiceOfferings())          // ServiceOfferings client is implemented
	assert.NotNil(t, client.ServicePlans())              // ServicePlans client is implemented
	assert.NotNil(t, client.ServiceInstances())          // ServiceInstances client is implemented
	assert.NotNil(t, client.ServiceCredentialBindings()) // ServiceCredentialBindings client is implemented
	assert.NotNil(t, client.ServiceRouteBindings())      // ServiceRouteBindings client is implemented
	assert.NotNil(t, client.Builds())                    // Builds client is implemented
	assert.NotNil(t, client.Buildpacks())                // Buildpacks client is implemented
	assert.NotNil(t, client.Deployments())               // Deployments client is implemented
	assert.NotNil(t, client.Droplets())                  // Droplets client is implemented
	assert.NotNil(t, client.Packages())                  // Packages client is implemented
	assert.NotNil(t, client.Processes())                 // Processes client is implemented
	assert.NotNil(t, client.Tasks())                     // Tasks client is implemented
	assert.NotNil(t, client.Stacks())                    // Stacks client is implemented
	assert.NotNil(t, client.Roles())                     // Roles client is implemented
	assert.NotNil(t, client.SecurityGroups())            // SecurityGroups client is implemented
	assert.NotNil(t, client.IsolationSegments())         // IsolationSegments client is implemented
	assert.NotNil(t, client.FeatureFlags())              // FeatureFlags client is implemented
	assert.NotNil(t, client.Jobs())                      // Jobs client is implemented
	assert.NotNil(t, client.Users())                     // Users client is implemented
}
