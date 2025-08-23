package cfclient_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fivetwenty-io/capi/v3/pkg/capi"
	"github.com/fivetwenty-io/capi/v3/pkg/cfclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("creates client with config", func(t *testing.T) {
		config := &capi.Config{
			APIEndpoint: "https://api.example.com",
		}

		client, err := cfclient.New(config)
		require.NoError(t, err)
		assert.NotNil(t, client)
	})
}

func TestNewWithEndpoint(t *testing.T) {
	client, err := cfclient.NewWithEndpoint("https://api.example.com")
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNewWithToken(t *testing.T) {
	client, err := cfclient.NewWithToken("https://api.example.com", "test-token")
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNewWithClientCredentials(t *testing.T) {
	t.Skip("Skipping test that requires network access")
	client, err := cfclient.NewWithClientCredentials("https://api.example.com", "client-id", "client-secret")
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNewWithPassword(t *testing.T) {
	t.Skip("Skipping test that requires network access")
	client, err := cfclient.NewWithPassword("https://api.example.com", "username", "password")
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestClientIntegration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v3/info":
			info := capi.Info{
				Name:    "Test CF",
				Version: 3,
			}
			json.NewEncoder(w).Encode(info)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := cfclient.NewWithEndpoint(server.URL)
	require.NoError(t, err)

	info, err := client.GetInfo(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "Test CF", info.Name)
	assert.Equal(t, 3, info.Version)
}
