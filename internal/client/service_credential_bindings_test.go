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

func TestServiceCredentialBindingsClient_Create_App(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_credential_bindings", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var request capi.ServiceCredentialBindingCreateRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		require.NoError(t, err)

		assert.Equal(t, "app", request.Type)
		assert.Equal(t, "my-binding", *request.Name)
		assert.Equal(t, "instance-guid", request.Relationships.ServiceInstance.Data.GUID)
		assert.Equal(t, "app-guid", request.Relationships.App.Data.GUID)

		// App bindings may return a job for async operations
		job := capi.Job{
			Resource: capi.Resource{
				GUID: "job-guid",
			},
			Operation: "service_credential_binding.create",
			State:     "PROCESSING",
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location", "/v3/jobs/job-guid")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(job)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	serviceBindings := NewServiceCredentialBindingsClient(client.httpClient)

	name := "my-binding"
	request := &capi.ServiceCredentialBindingCreateRequest{
		Type: "app",
		Name: &name,
		Parameters: map[string]interface{}{
			"foo": "bar",
		},
		Relationships: capi.ServiceCredentialBindingRelationships{
			ServiceInstance: capi.Relationship{
				Data: &capi.RelationshipData{
					GUID: "instance-guid",
				},
			},
			App: &capi.Relationship{
				Data: &capi.RelationshipData{
					GUID: "app-guid",
				},
			},
		},
	}

	result, err := serviceBindings.Create(context.Background(), request)
	require.NoError(t, err)

	job, ok := result.(*capi.Job)
	require.True(t, ok, "Expected *capi.Job for app binding")
	assert.Equal(t, "job-guid", job.GUID)
	assert.Equal(t, "service_credential_binding.create", job.Operation)
}

func TestServiceCredentialBindingsClient_Create_Key(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_credential_bindings", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var request capi.ServiceCredentialBindingCreateRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		require.NoError(t, err)

		assert.Equal(t, "key", request.Type)
		assert.Equal(t, "my-key", *request.Name)
		assert.Equal(t, "instance-guid", request.Relationships.ServiceInstance.Data.GUID)
		assert.Nil(t, request.Relationships.App)

		// Key bindings usually return the binding directly
		now := time.Now()
		binding := capi.ServiceCredentialBinding{
			Resource: capi.Resource{
				GUID:      "binding-guid",
				CreatedAt: now,
				UpdatedAt: now,
			},
			Name: "my-key",
			Type: "key",
			LastOperation: &capi.ServiceCredentialBindingLastOperation{
				Type:      "create",
				State:     "succeeded",
				CreatedAt: &now,
				UpdatedAt: &now,
			},
			Relationships: capi.ServiceCredentialBindingRelationships{
				ServiceInstance: capi.Relationship{
					Data: &capi.RelationshipData{
						GUID: "instance-guid",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(binding)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	serviceBindings := NewServiceCredentialBindingsClient(client.httpClient)

	name := "my-key"
	request := &capi.ServiceCredentialBindingCreateRequest{
		Type: "key",
		Name: &name,
		Relationships: capi.ServiceCredentialBindingRelationships{
			ServiceInstance: capi.Relationship{
				Data: &capi.RelationshipData{
					GUID: "instance-guid",
				},
			},
		},
	}

	result, err := serviceBindings.Create(context.Background(), request)
	require.NoError(t, err)

	binding, ok := result.(*capi.ServiceCredentialBinding)
	require.True(t, ok, "Expected *capi.ServiceCredentialBinding for key binding")
	assert.Equal(t, "binding-guid", binding.GUID)
	assert.Equal(t, "my-key", binding.Name)
	assert.Equal(t, "key", binding.Type)
}

func TestServiceCredentialBindingsClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_credential_bindings/binding-guid", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		now := time.Now()
		binding := capi.ServiceCredentialBinding{
			Resource: capi.Resource{
				GUID:      "binding-guid",
				CreatedAt: now,
				UpdatedAt: now,
			},
			Name: "my-binding",
			Type: "app",
			LastOperation: &capi.ServiceCredentialBindingLastOperation{
				Type:      "create",
				State:     "succeeded",
				CreatedAt: &now,
				UpdatedAt: &now,
			},
			Relationships: capi.ServiceCredentialBindingRelationships{
				ServiceInstance: capi.Relationship{
					Data: &capi.RelationshipData{
						GUID: "instance-guid",
					},
				},
				App: &capi.Relationship{
					Data: &capi.RelationshipData{
						GUID: "app-guid",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(binding)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	serviceBindings := NewServiceCredentialBindingsClient(client.httpClient)

	binding, err := serviceBindings.Get(context.Background(), "binding-guid")
	require.NoError(t, err)
	assert.NotNil(t, binding)
	assert.Equal(t, "binding-guid", binding.GUID)
	assert.Equal(t, "my-binding", binding.Name)
	assert.Equal(t, "app", binding.Type)
}

func TestServiceCredentialBindingsClient_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_credential_bindings", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "instance-guid", r.URL.Query().Get("service_instance_guids"))
		assert.Equal(t, "app-guid", r.URL.Query().Get("app_guids"))

		now := time.Now()
		response := capi.ListResponse[capi.ServiceCredentialBinding]{
			Pagination: capi.Pagination{
				TotalResults: 2,
				TotalPages:   1,
				First:        capi.Link{Href: "/v3/service_credential_bindings?page=1"},
				Last:         capi.Link{Href: "/v3/service_credential_bindings?page=1"},
			},
			Resources: []capi.ServiceCredentialBinding{
				{
					Resource: capi.Resource{
						GUID:      "binding-guid-1",
						CreatedAt: now,
						UpdatedAt: now,
					},
					Name: "binding-1",
					Type: "app",
				},
				{
					Resource: capi.Resource{
						GUID:      "binding-guid-2",
						CreatedAt: now,
						UpdatedAt: now,
					},
					Name: "binding-2",
					Type: "key",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	serviceBindings := NewServiceCredentialBindingsClient(client.httpClient)

	params := &capi.QueryParams{
		Filters: map[string][]string{
			"service_instance_guids": {"instance-guid"},
			"app_guids":              {"app-guid"},
		},
	}

	list, err := serviceBindings.List(context.Background(), params)
	require.NoError(t, err)
	assert.NotNil(t, list)
	assert.Equal(t, 2, list.Pagination.TotalResults)
	assert.Len(t, list.Resources, 2)
	assert.Equal(t, "binding-1", list.Resources[0].Name)
	assert.Equal(t, "app", list.Resources[0].Type)
	assert.Equal(t, "binding-2", list.Resources[1].Name)
	assert.Equal(t, "key", list.Resources[1].Type)
}

func TestServiceCredentialBindingsClient_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_credential_bindings/binding-guid", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var request capi.ServiceCredentialBindingUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		require.NoError(t, err)

		now := time.Now()
		binding := capi.ServiceCredentialBinding{
			Resource: capi.Resource{
				GUID:      "binding-guid",
				CreatedAt: now,
				UpdatedAt: now,
			},
			Name:     "my-binding",
			Type:     "app",
			Metadata: request.Metadata,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(binding)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	serviceBindings := NewServiceCredentialBindingsClient(client.httpClient)

	request := &capi.ServiceCredentialBindingUpdateRequest{
		Metadata: &capi.Metadata{
			Labels: map[string]string{
				"env": "production",
			},
			Annotations: map[string]string{
				"owner": "team-a",
			},
		},
	}

	binding, err := serviceBindings.Update(context.Background(), "binding-guid", request)
	require.NoError(t, err)
	assert.NotNil(t, binding)
	assert.Equal(t, "binding-guid", binding.GUID)
}

func TestServiceCredentialBindingsClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_credential_bindings/binding-guid", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		job := capi.Job{
			Resource: capi.Resource{
				GUID: "job-guid",
			},
			Operation: "service_credential_binding.delete",
			State:     "PROCESSING",
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location", "/v3/jobs/job-guid")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(job)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	serviceBindings := NewServiceCredentialBindingsClient(client.httpClient)

	job, err := serviceBindings.Delete(context.Background(), "binding-guid")
	require.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, "job-guid", job.GUID)
	assert.Equal(t, "service_credential_binding.delete", job.Operation)
}

func TestServiceCredentialBindingsClient_GetDetails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_credential_bindings/binding-guid/details", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		details := capi.ServiceCredentialBindingDetails{
			Credentials: map[string]interface{}{
				"username": "admin",
				"password": "secret",
				"uri":      "mysql://admin:secret@localhost:3306/mydb",
			},
			SyslogDrainURL: stringPtr("syslog://example.com"),
			VolumeMounts:   []interface{}{},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(details)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	serviceBindings := NewServiceCredentialBindingsClient(client.httpClient)

	details, err := serviceBindings.GetDetails(context.Background(), "binding-guid")
	require.NoError(t, err)
	assert.NotNil(t, details)
	assert.Equal(t, "admin", details.Credentials["username"])
	assert.Equal(t, "secret", details.Credentials["password"])
	assert.NotNil(t, details.SyslogDrainURL)
	assert.Equal(t, "syslog://example.com", *details.SyslogDrainURL)
}

func TestServiceCredentialBindingsClient_GetParameters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_credential_bindings/binding-guid/parameters", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		params := capi.ServiceCredentialBindingParameters{
			Parameters: map[string]interface{}{
				"foo":             "bar",
				"max_connections": 10,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(params)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	serviceBindings := NewServiceCredentialBindingsClient(client.httpClient)

	params, err := serviceBindings.GetParameters(context.Background(), "binding-guid")
	require.NoError(t, err)
	assert.NotNil(t, params)
	assert.Equal(t, "bar", params.Parameters["foo"])
	assert.Equal(t, float64(10), params.Parameters["max_connections"])
}

func TestServiceCredentialBindingsClient_GetNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_credential_bindings/binding-guid", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	serviceBindings := NewServiceCredentialBindingsClient(client.httpClient)

	binding, err := serviceBindings.Get(context.Background(), "binding-guid")
	assert.Error(t, err)
	assert.Nil(t, binding)
}

func TestServiceCredentialBindingsClient_CreateForbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/service_credential_bindings", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	serviceBindings := NewServiceCredentialBindingsClient(client.httpClient)

	name := "my-binding"
	request := &capi.ServiceCredentialBindingCreateRequest{
		Type: "app",
		Name: &name,
		Relationships: capi.ServiceCredentialBindingRelationships{
			ServiceInstance: capi.Relationship{
				Data: &capi.RelationshipData{
					GUID: "instance-guid",
				},
			},
			App: &capi.Relationship{
				Data: &capi.RelationshipData{
					GUID: "app-guid",
				},
			},
		},
	}

	result, err := serviceBindings.Create(context.Background(), request)
	assert.Error(t, err)
	assert.Nil(t, result)
}
