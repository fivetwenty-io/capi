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
	t.Parallel()
	t.Run("creates client with config", func(t *testing.T) {
		t.Parallel()

		config := &capi.Config{
			APIEndpoint: "https://api.example.com",
		}

		client, err := cfclient.New(context.Background(), config)
		require.NoError(t, err)
		assert.NotNil(t, client)
	})
}

func TestNewWithEndpoint(t *testing.T) {
	t.Parallel()

	client, err := cfclient.NewWithEndpoint(context.Background(), "https://api.example.com")
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNewWithToken(t *testing.T) {
	t.Parallel()

	client, err := cfclient.NewWithToken(context.Background(), "https://api.example.com", "test-token")
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNewWithClientCredentials(t *testing.T) {
	t.Parallel()
	t.Skip("Skipping test that requires network access")

	client, err := cfclient.NewWithClientCredentials(context.Background(), "https://api.example.com", "client-id", "client-secret")
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNewWithPassword(t *testing.T) {
	t.Parallel()
	t.Skip("Skipping test that requires network access")

	client, err := cfclient.NewWithPassword(context.Background(), "https://api.example.com", "username", "password")
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestClientIntegration(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/v3/info":
			info := capi.Info{
				Name:    "Test CF",
				Version: 3,
			}
			_ = json.NewEncoder(writer).Encode(info)
		default:
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := cfclient.NewWithEndpoint(context.Background(), server.URL)
	require.NoError(t, err)

	info, err := client.GetInfo(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "Test CF", info.Name)
	assert.Equal(t, 3, info.Version)
}
