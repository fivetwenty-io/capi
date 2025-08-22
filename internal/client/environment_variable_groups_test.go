package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fivetwenty-io/capi-client/pkg/capi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvironmentVariableGroupsClient_GetRunning(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/environment_variable_groups/running", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		envVarGroup := capi.EnvironmentVariableGroup{
			Name: "running",
			Var: map[string]interface{}{
				"LOG_LEVEL": "info",
				"TIMEOUT":   30,
			},
			UpdatedAt: &time.Time{},
		}

		json.NewEncoder(w).Encode(envVarGroup)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	envVarGroup, err := client.EnvironmentVariableGroups().Get(context.Background(), "running")
	require.NoError(t, err)
	assert.Equal(t, "running", envVarGroup.Name)
	assert.Equal(t, "info", envVarGroup.Var["LOG_LEVEL"])
	assert.Equal(t, float64(30), envVarGroup.Var["TIMEOUT"])
}

func TestEnvironmentVariableGroupsClient_GetStaging(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/environment_variable_groups/staging", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		envVarGroup := capi.EnvironmentVariableGroup{
			Name: "staging",
			Var: map[string]interface{}{
				"BUILD_ENV": "production",
				"CACHE":     true,
			},
			UpdatedAt: &time.Time{},
		}

		json.NewEncoder(w).Encode(envVarGroup)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	envVarGroup, err := client.EnvironmentVariableGroups().Get(context.Background(), "staging")
	require.NoError(t, err)
	assert.Equal(t, "staging", envVarGroup.Name)
	assert.Equal(t, "production", envVarGroup.Var["BUILD_ENV"])
	assert.Equal(t, true, envVarGroup.Var["CACHE"])
}

func TestEnvironmentVariableGroupsClient_UpdateRunning(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/environment_variable_groups/running", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)
		varMap := req["var"].(map[string]interface{})
		assert.Equal(t, "debug", varMap["LOG_LEVEL"])
		assert.Equal(t, float64(60), varMap["TIMEOUT"])

		envVarGroup := capi.EnvironmentVariableGroup{
			Name: "running",
			Var:  varMap,
		}

		json.NewEncoder(w).Encode(envVarGroup)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	envVarGroup, err := client.EnvironmentVariableGroups().Update(context.Background(), "running", map[string]interface{}{
		"LOG_LEVEL": "debug",
		"TIMEOUT":   60,
	})

	require.NoError(t, err)
	assert.Equal(t, "running", envVarGroup.Name)
	assert.Equal(t, "debug", envVarGroup.Var["LOG_LEVEL"])
	assert.Equal(t, float64(60), envVarGroup.Var["TIMEOUT"])
}

func TestEnvironmentVariableGroupsClient_UpdateStaging(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/environment_variable_groups/staging", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)
		varMap := req["var"].(map[string]interface{})
		assert.Equal(t, "development", varMap["BUILD_ENV"])
		assert.Equal(t, false, varMap["CACHE"])

		envVarGroup := capi.EnvironmentVariableGroup{
			Name: "staging",
			Var:  varMap,
		}

		json.NewEncoder(w).Encode(envVarGroup)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	envVarGroup, err := client.EnvironmentVariableGroups().Update(context.Background(), "staging", map[string]interface{}{
		"BUILD_ENV": "development",
		"CACHE":     false,
	})

	require.NoError(t, err)
	assert.Equal(t, "staging", envVarGroup.Name)
	assert.Equal(t, "development", envVarGroup.Var["BUILD_ENV"])
	assert.Equal(t, false, envVarGroup.Var["CACHE"])
}
