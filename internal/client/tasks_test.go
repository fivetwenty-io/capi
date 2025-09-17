package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/fivetwenty-io/capi/v3/internal/client"
	"github.com/fivetwenty-io/capi/v3/pkg/capi"
)

//nolint:funlen // Test functions can be longer for detailed testing
func TestTasksClient_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		appGUID      string
		request      *capi.TaskCreateRequest
		response     interface{}
		statusCode   int
		expectedPath string
		wantErr      bool
		errMessage   string
	}{
		{
			name:         "create task with command",
			appGUID:      "test-app-guid",
			expectedPath: "/v3/apps/test-app-guid/tasks",
			statusCode:   http.StatusAccepted,
			request: &capi.TaskCreateRequest{
				Command:    StringPtr("rake db:migrate"),
				Name:       StringPtr("migrate"),
				MemoryInMB: intPtr(512),
				DiskInMB:   intPtr(1024),
			},
			response: capi.Task{
				Resource: capi.Resource{
					GUID:      "task-guid",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					Links: capi.Links{
						"self": capi.Link{
							Href: "https://api.example.org/v3/tasks/task-guid",
						},
						"app": capi.Link{
							Href: "https://api.example.org/v3/apps/test-app-guid",
						},
						"cancel": capi.Link{
							Href:   "https://api.example.org/v3/tasks/task-guid/actions/cancel",
							Method: "POST",
						},
						"droplet": capi.Link{
							Href: "https://api.example.org/v3/droplets/droplet-guid",
						},
					},
				},
				SequenceID:  1,
				Name:        "migrate",
				Command:     "rake db:migrate",
				User:        StringPtr("vcap"),
				State:       "RUNNING",
				MemoryInMB:  512,
				DiskInMB:    1024,
				DropletGUID: "droplet-guid",
				Result: &capi.TaskResult{
					FailureReason: nil,
				},
				Metadata: &capi.Metadata{
					Labels:      map[string]string{},
					Annotations: map[string]string{},
				},
				Relationships: &capi.TaskRelationships{
					App: &capi.Relationship{
						Data: &capi.RelationshipData{
							GUID: "test-app-guid",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:         "create task with template",
			appGUID:      "test-app-guid",
			expectedPath: "/v3/apps/test-app-guid/tasks",
			statusCode:   http.StatusAccepted,
			request: &capi.TaskCreateRequest{
				Template: &capi.TaskTemplate{
					Process: &capi.TaskTemplateProcess{
						GUID: "process-guid",
					},
				},
			},
			response: capi.Task{
				Resource: capi.Resource{
					GUID:      "task-guid-2",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				SequenceID:  2,
				Name:        "task",
				Command:     "bundle exec rackup",
				State:       "PENDING",
				MemoryInMB:  256,
				DiskInMB:    512,
				DropletGUID: "droplet-guid",
			},
			wantErr: false,
		},
		{
			name:         "app not found",
			appGUID:      "non-existent-app",
			expectedPath: "/v3/apps/non-existent-app/tasks",
			statusCode:   http.StatusNotFound,
			request: &capi.TaskCreateRequest{
				Command: StringPtr("ls"),
			},
			response: map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"code":   10010,
						"title":  "CF-ResourceNotFound",
						"detail": "App not found",
					},
				},
			},
			wantErr:    true,
			errMessage: "CF-ResourceNotFound",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				assert.Equal(t, testCase.expectedPath, request.URL.Path)
				assert.Equal(t, "POST", request.Method)

				var requestBody capi.TaskCreateRequest

				err := json.NewDecoder(request.Body).Decode(&requestBody)
				assert.NoError(t, err)

				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(testCase.statusCode)
				_ = json.NewEncoder(writer).Encode(testCase.response)
			}))
			defer server.Close()

			client, err := New(context.Background(), &capi.Config{APIEndpoint: server.URL})
			require.NoError(t, err)

			task, err := client.Tasks().Create(context.Background(), testCase.appGUID, testCase.request)

			if testCase.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.errMessage)
				assert.Nil(t, task)
			} else {
				require.NoError(t, err)
				require.NotNil(t, task)
				assert.NotEmpty(t, task.GUID)
				assert.NotEmpty(t, task.State)
			}
		})
	}
}

func TestTasksClient_Get(t *testing.T) {
	t.Parallel()

	tests := []TestGetOperation[capi.Task]{
		{
			Name:         "successful get",
			GUID:         "test-task-guid",
			ExpectedPath: "/v3/tasks/test-task-guid",
			StatusCode:   http.StatusOK,
			Response: &capi.Task{
				Resource: capi.Resource{
					GUID:      "test-task-guid",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				SequenceID:                   1,
				Name:                         "migrate",
				Command:                      "rake db:migrate",
				User:                         StringPtr("vcap"),
				State:                        "SUCCEEDED",
				MemoryInMB:                   512,
				DiskInMB:                     1024,
				LogRateLimitInBytesPerSecond: intPtr(1024),
				DropletGUID:                  "droplet-guid",
				Result: &capi.TaskResult{
					FailureReason: nil,
				},
			},
			WantErr: false,
		},
		{
			Name:         "task not found",
			GUID:         "non-existent-guid",
			ExpectedPath: "/v3/tasks/non-existent-guid",
			StatusCode:   http.StatusNotFound,
			Response: &capi.Task{
				Resource: capi.Resource{
					GUID:      "test-task-guid",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				State: "SUCCEEDED",
			},
			WantErr:    true,
			ErrMessage: "CF-ResourceNotFound",
		},
	}

	RunGetTests(t, tests, func(c *Client) func(context.Context, string) (*capi.Task, error) {
		return c.Tasks().Get
	})
}

//nolint:funlen // Test functions can be longer for detailed testing
func TestTasksClient_List(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/tasks", request.URL.Path)
		assert.Equal(t, "GET", request.Method)

		// Check query parameters if present
		query := request.URL.Query()
		if appGuids := query.Get("app_guids"); appGuids != "" {
			assert.Equal(t, "app-1,app-2", appGuids)
		}

		if states := query.Get("states"); states != "" {
			assert.Equal(t, "RUNNING,PENDING", states)
		}

		response := capi.ListResponse[capi.Task]{
			Pagination: capi.Pagination{
				TotalResults: 2,
				TotalPages:   1,
				First:        capi.Link{Href: "https://api.example.org/v3/tasks?page=1"},
				Last:         capi.Link{Href: "https://api.example.org/v3/tasks?page=1"},
				Next:         nil,
				Previous:     nil,
			},
			Resources: []capi.Task{
				{
					Resource: capi.Resource{
						GUID:      "task-1",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					SequenceID:  1,
					Name:        "migrate",
					Command:     "rake db:migrate",
					State:       "RUNNING",
					MemoryInMB:  512,
					DiskInMB:    1024,
					DropletGUID: "droplet-1",
				},
				{
					Resource: capi.Resource{
						GUID:      "task-2",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					SequenceID:  2,
					Name:        "seed",
					Command:     "rake db:seed",
					State:       "PENDING",
					MemoryInMB:  256,
					DiskInMB:    512,
					DropletGUID: "droplet-2",
				},
			},
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(writer).Encode(response)
	}))
	defer server.Close()

	client, err := New(context.Background(), &capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	// Test without filters
	result, err := client.Tasks().List(context.Background(), nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 2, result.Pagination.TotalResults)
	assert.Len(t, result.Resources, 2)
	assert.Equal(t, "task-1", result.Resources[0].GUID)
	assert.Equal(t, "RUNNING", result.Resources[0].State)

	// Test with filters
	params := &capi.QueryParams{
		Filters: map[string][]string{
			"app_guids": {"app-1", "app-2"},
			"states":    {"RUNNING", "PENDING"},
		},
	}
	result, err = client.Tasks().List(context.Background(), params)
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestTasksClient_Update(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/tasks/test-task-guid", request.URL.Path)
		assert.Equal(t, "PATCH", request.Method)

		var requestBody capi.TaskUpdateRequest

		err := json.NewDecoder(request.Body).Decode(&requestBody)
		assert.NoError(t, err)

		response := capi.Task{
			Resource: capi.Resource{
				GUID:      "test-task-guid",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			SequenceID:  1,
			Name:        "migrate",
			Command:     "rake db:migrate",
			State:       "RUNNING",
			MemoryInMB:  512,
			DiskInMB:    1024,
			DropletGUID: "droplet-guid",
			Metadata:    requestBody.Metadata,
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(writer).Encode(response)
	}))
	defer server.Close()

	client, err := New(context.Background(), &capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	request := &capi.TaskUpdateRequest{
		Metadata: &capi.Metadata{
			Labels: map[string]string{
				"env": "production",
			},
			Annotations: map[string]string{
				"note": "database migration",
			},
		},
	}

	task, err := client.Tasks().Update(context.Background(), "test-task-guid", request)
	require.NoError(t, err)
	require.NotNil(t, task)
	assert.Equal(t, "test-task-guid", task.GUID)
	assert.Equal(t, "production", task.Metadata.Labels["env"])
	assert.Equal(t, "database migration", task.Metadata.Annotations["note"])
}

//nolint:funlen // Test functions can be longer for detailed testing
func TestTasksClient_Cancel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		guid         string
		response     interface{}
		statusCode   int
		expectedPath string
		wantErr      bool
		errMessage   string
	}{
		{
			name:         "successful cancel",
			guid:         "test-task-guid",
			expectedPath: "/v3/tasks/test-task-guid/actions/cancel",
			statusCode:   http.StatusAccepted,
			response: capi.Task{
				Resource: capi.Resource{
					GUID:      "test-task-guid",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				SequenceID:  1,
				Name:        "migrate",
				Command:     "rake db:migrate",
				State:       "CANCELING",
				MemoryInMB:  512,
				DiskInMB:    1024,
				DropletGUID: "droplet-guid",
			},
			wantErr: false,
		},
		{
			name:         "task already completed",
			guid:         "completed-task-guid",
			expectedPath: "/v3/tasks/completed-task-guid/actions/cancel",
			statusCode:   http.StatusUnprocessableEntity,
			response: map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"code":   10008,
						"title":  "CF-UnprocessableEntity",
						"detail": "Task has already been completed",
					},
				},
			},
			wantErr:    true,
			errMessage: "CF-UnprocessableEntity",
		},
		{
			name:         "task not found",
			guid:         "non-existent-guid",
			expectedPath: "/v3/tasks/non-existent-guid/actions/cancel",
			statusCode:   http.StatusNotFound,
			response: map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"code":   10010,
						"title":  "CF-ResourceNotFound",
						"detail": "Task not found",
					},
				},
			},
			wantErr:    true,
			errMessage: "CF-ResourceNotFound",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				assert.Equal(t, testCase.expectedPath, request.URL.Path)
				assert.Equal(t, "POST", request.Method)
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(testCase.statusCode)
				_ = json.NewEncoder(writer).Encode(testCase.response)
			}))
			defer server.Close()

			client, err := New(context.Background(), &capi.Config{APIEndpoint: server.URL})
			require.NoError(t, err)

			task, err := client.Tasks().Cancel(context.Background(), testCase.guid)

			if testCase.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.errMessage)
				assert.Nil(t, task)
			} else {
				require.NoError(t, err)
				require.NotNil(t, task)
				assert.Equal(t, testCase.guid, task.GUID)
				assert.Equal(t, "CANCELING", task.State)
			}
		})
	}
}
