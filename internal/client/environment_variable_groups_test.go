package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/fivetwenty-io/capi/v3/internal/client"
	"github.com/fivetwenty-io/capi/v3/pkg/capi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvironmentVariableGroupsClient_GetRunning(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/environment_variable_groups/running", request.URL.Path)
		assert.Equal(t, "GET", request.Method)

		envVarGroup := capi.EnvironmentVariableGroup{
			Name: "running",
			Var: map[string]interface{}{
				"LOG_LEVEL": "info",
				"TIMEOUT":   30,
			},
			UpdatedAt: &time.Time{},
		}

		_ = json.NewEncoder(writer).Encode(envVarGroup)
	}))
	defer server.Close()

	client, err := New(context.Background(), &capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	envVarGroup, err := client.EnvironmentVariableGroups().Get(context.Background(), "running")
	require.NoError(t, err)
	assert.Equal(t, "running", envVarGroup.Name)
	assert.Equal(t, "info", envVarGroup.Var["LOG_LEVEL"])
	assert.InDelta(t, float64(30), envVarGroup.Var["TIMEOUT"], 0.01)
}

func TestEnvironmentVariableGroupsClient_GetStaging(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/environment_variable_groups/staging", request.URL.Path)
		assert.Equal(t, "GET", request.Method)

		envVarGroup := capi.EnvironmentVariableGroup{
			Name: "staging",
			Var: map[string]interface{}{
				"BUILD_ENV": "production",
				"CACHE":     true,
			},
			UpdatedAt: &time.Time{},
		}

		_ = json.NewEncoder(writer).Encode(envVarGroup)
	}))
	defer server.Close()

	client, err := New(context.Background(), &capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	envVarGroup, err := client.EnvironmentVariableGroups().Get(context.Background(), "staging")
	require.NoError(t, err)
	assert.Equal(t, "staging", envVarGroup.Name)
	assert.Equal(t, "production", envVarGroup.Var["BUILD_ENV"])
	assert.Equal(t, true, envVarGroup.Var["CACHE"])
}

func TestEnvironmentVariableGroupsClient_UpdateRunning(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/environment_variable_groups/running", request.URL.Path)
		assert.Equal(t, "PATCH", request.Method)

		var req map[string]interface{}

		_ = json.NewDecoder(request.Body).Decode(&req)

		varMap, ok := req["var"].(map[string]interface{})
		if !ok {
			t.Errorf("req[\"var\"] is not a map[string]interface{}")

			return
		}

		assert.Equal(t, "debug", varMap["LOG_LEVEL"])
		assert.InDelta(t, float64(60), varMap["TIMEOUT"], 0.01)

		envVarGroup := capi.EnvironmentVariableGroup{
			Name: "running",
			Var:  varMap,
		}

		_ = json.NewEncoder(writer).Encode(envVarGroup)
	}))
	defer server.Close()

	client, err := New(context.Background(), &capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	envVarGroup, err := client.EnvironmentVariableGroups().Update(context.Background(), "running", map[string]interface{}{
		"LOG_LEVEL": "debug",
		"TIMEOUT":   60,
	})

	require.NoError(t, err)
	assert.Equal(t, "running", envVarGroup.Name)
	assert.Equal(t, "debug", envVarGroup.Var["LOG_LEVEL"])
	assert.InDelta(t, float64(60), envVarGroup.Var["TIMEOUT"], 0.01)
}

func TestEnvironmentVariableGroupsClient_UpdateStaging(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/environment_variable_groups/staging", request.URL.Path)
		assert.Equal(t, "PATCH", request.Method)

		var req map[string]interface{}

		_ = json.NewDecoder(request.Body).Decode(&req)

		varMap, ok := req["var"].(map[string]interface{})
		if !ok {
			t.Errorf("req[\"var\"] is not a map[string]interface{}")

			return
		}

		assert.Equal(t, "development", varMap["BUILD_ENV"])
		assert.Equal(t, false, varMap["CACHE"])

		envVarGroup := capi.EnvironmentVariableGroup{
			Name: "staging",
			Var:  varMap,
		}

		_ = json.NewEncoder(writer).Encode(envVarGroup)
	}))
	defer server.Close()

	client, err := New(context.Background(), &capi.Config{APIEndpoint: server.URL})
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
