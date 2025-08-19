package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fivetwenty-io/capi-client-go/pkg/capi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppsClient_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/apps", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req capi.AppCreateRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "test-app", req.Name)
		assert.Equal(t, "space-guid", req.Relationships.Space.Data.GUID)

		app := capi.App{
			Resource: capi.Resource{
				GUID:      "app-guid",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:  req.Name,
			State: "STOPPED",
			Lifecycle: capi.Lifecycle{
				Type: "buildpack",
				Data: map[string]interface{}{},
			},
			Relationships: req.Relationships,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(app)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	app, err := client.Apps().Create(context.Background(), &capi.AppCreateRequest{
		Name: "test-app",
		Relationships: capi.AppRelationships{
			Space: capi.Relationship{
				Data: &capi.RelationshipData{GUID: "space-guid"},
			},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "app-guid", app.GUID)
	assert.Equal(t, "test-app", app.Name)
	assert.Equal(t, "STOPPED", app.State)
}

func TestAppsClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/apps/app-guid", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		app := capi.App{
			Resource: capi.Resource{
				GUID: "app-guid",
			},
			Name:  "test-app",
			State: "STARTED",
		}

		json.NewEncoder(w).Encode(app)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	app, err := client.Apps().Get(context.Background(), "app-guid")
	require.NoError(t, err)
	assert.Equal(t, "app-guid", app.GUID)
	assert.Equal(t, "test-app", app.Name)
	assert.Equal(t, "STARTED", app.State)
}

func TestAppsClient_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/apps", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "10", r.URL.Query().Get("per_page"))

		response := capi.ListResponse[capi.App]{
			Pagination: capi.Pagination{
				TotalResults: 2,
				TotalPages:   1,
			},
			Resources: []capi.App{
				{
					Resource: capi.Resource{GUID: "app-1"},
					Name:     "app-1",
					State:    "STARTED",
				},
				{
					Resource: capi.Resource{GUID: "app-2"},
					Name:     "app-2",
					State:    "STOPPED",
				},
			},
		}

		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	params := capi.NewQueryParams().WithPage(1).WithPerPage(10)
	result, err := client.Apps().List(context.Background(), params)

	require.NoError(t, err)
	assert.Len(t, result.Resources, 2)
	assert.Equal(t, "app-1", result.Resources[0].Name)
	assert.Equal(t, "app-2", result.Resources[1].Name)
}

func TestAppsClient_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/apps/app-guid", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var req capi.AppUpdateRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "updated-app", *req.Name)

		app := capi.App{
			Resource: capi.Resource{GUID: "app-guid"},
			Name:     *req.Name,
			State:    "STOPPED",
		}

		json.NewEncoder(w).Encode(app)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	newName := "updated-app"
	app, err := client.Apps().Update(context.Background(), "app-guid", &capi.AppUpdateRequest{
		Name: &newName,
	})

	require.NoError(t, err)
	assert.Equal(t, "updated-app", app.Name)
}

func TestAppsClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/apps/app-guid", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	err = client.Apps().Delete(context.Background(), "app-guid")
	require.NoError(t, err)
}

func TestAppsClient_Start(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/apps/app-guid/actions/start", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		app := capi.App{
			Resource: capi.Resource{GUID: "app-guid"},
			Name:     "test-app",
			State:    "STARTED",
		}

		json.NewEncoder(w).Encode(app)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	app, err := client.Apps().Start(context.Background(), "app-guid")
	require.NoError(t, err)
	assert.Equal(t, "STARTED", app.State)
}

func TestAppsClient_Stop(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/apps/app-guid/actions/stop", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		app := capi.App{
			Resource: capi.Resource{GUID: "app-guid"},
			Name:     "test-app",
			State:    "STOPPED",
		}

		json.NewEncoder(w).Encode(app)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	app, err := client.Apps().Stop(context.Background(), "app-guid")
	require.NoError(t, err)
	assert.Equal(t, "STOPPED", app.State)
}

func TestAppsClient_Restart(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/apps/app-guid/actions/restart", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		app := capi.App{
			Resource: capi.Resource{GUID: "app-guid"},
			Name:     "test-app",
			State:    "STARTED",
		}

		json.NewEncoder(w).Encode(app)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	app, err := client.Apps().Restart(context.Background(), "app-guid")
	require.NoError(t, err)
	assert.Equal(t, "STARTED", app.State)
}

func TestAppsClient_GetEnv(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/apps/app-guid/env", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		env := capi.AppEnvironment{
			StagingEnvJSON: map[string]interface{}{
				"STAGING_VAR": "staging_value",
			},
			RunningEnvJSON: map[string]interface{}{
				"RUNNING_VAR": "running_value",
			},
			EnvironmentVariables: map[string]interface{}{
				"USER_VAR": "user_value",
			},
			SystemEnvJSON: map[string]interface{}{
				"VCAP_SERVICES": map[string]interface{}{},
			},
			ApplicationEnvJSON: map[string]interface{}{
				"VCAP_APPLICATION": map[string]interface{}{
					"name": "test-app",
				},
			},
		}

		json.NewEncoder(w).Encode(env)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	env, err := client.Apps().GetEnv(context.Background(), "app-guid")
	require.NoError(t, err)
	assert.Equal(t, "user_value", env.EnvironmentVariables["USER_VAR"])
	assert.Equal(t, "staging_value", env.StagingEnvJSON["STAGING_VAR"])
}

func TestAppsClient_GetEnvVars(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/apps/app-guid/environment_variables", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		response := map[string]interface{}{
			"var": map[string]interface{}{
				"KEY1": "value1",
				"KEY2": "value2",
			},
		}

		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	vars, err := client.Apps().GetEnvVars(context.Background(), "app-guid")
	require.NoError(t, err)
	assert.Equal(t, "value1", vars["KEY1"])
	assert.Equal(t, "value2", vars["KEY2"])
}

func TestAppsClient_UpdateEnvVars(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/apps/app-guid/environment_variables", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "new_value", req["var"].(map[string]interface{})["NEW_KEY"])

		response := map[string]interface{}{
			"var": map[string]interface{}{
				"NEW_KEY": "new_value",
				"KEY1":    "value1",
			},
		}

		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	vars, err := client.Apps().UpdateEnvVars(context.Background(), "app-guid", map[string]interface{}{
		"NEW_KEY": "new_value",
	})
	require.NoError(t, err)
	assert.Equal(t, "new_value", vars["NEW_KEY"])
}

func TestAppsClient_GetCurrentDroplet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/apps/app-guid/droplets/current", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		droplet := capi.Droplet{
			Resource: capi.Resource{GUID: "droplet-guid"},
		}

		json.NewEncoder(w).Encode(droplet)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	droplet, err := client.Apps().GetCurrentDroplet(context.Background(), "app-guid")
	require.NoError(t, err)
	assert.Equal(t, "droplet-guid", droplet.GUID)
}

func TestAppsClient_SetCurrentDroplet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/apps/app-guid/relationships/current_droplet", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var req capi.Relationship
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "droplet-guid", req.Data.GUID)

		json.NewEncoder(w).Encode(req)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	rel, err := client.Apps().SetCurrentDroplet(context.Background(), "app-guid", "droplet-guid")
	require.NoError(t, err)
	assert.Equal(t, "droplet-guid", rel.Data.GUID)
}

func TestAppsClient_GetSSHEnabled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/apps/app-guid/ssh_enabled", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		sshEnabled := capi.AppSSHEnabled{
			Enabled: true,
			Reason:  "",
		}

		json.NewEncoder(w).Encode(sshEnabled)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	ssh, err := client.Apps().GetSSHEnabled(context.Background(), "app-guid")
	require.NoError(t, err)
	assert.True(t, ssh.Enabled)
}

func TestAppsClient_GetPermissions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/apps/app-guid/permissions", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		permissions := capi.AppPermissions{
			ReadBasicData:     true,
			ReadSensitiveData: false,
		}

		json.NewEncoder(w).Encode(permissions)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	perms, err := client.Apps().GetPermissions(context.Background(), "app-guid")
	require.NoError(t, err)
	assert.True(t, perms.ReadBasicData)
	assert.False(t, perms.ReadSensitiveData)
}

func TestAppsClient_ClearBuildpackCache(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/apps/app-guid/actions/clear_buildpack_cache", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	err = client.Apps().ClearBuildpackCache(context.Background(), "app-guid")
	require.NoError(t, err)
}

func TestAppsClient_GetManifest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/apps/app-guid/manifest", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		manifest := `applications:
- name: test-app
  memory: 512M
  instances: 2
  buildpack: nodejs_buildpack`

		w.Header().Set("Content-Type", "application/x-yaml")
		w.Write([]byte(manifest))
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	manifest, err := client.Apps().GetManifest(context.Background(), "app-guid")
	require.NoError(t, err)
	assert.Contains(t, manifest, "name: test-app")
	assert.Contains(t, manifest, "memory: 512M")
}
