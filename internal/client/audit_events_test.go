package client_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/fivetwenty-io/capi/v3/internal/client"
	"github.com/fivetwenty-io/capi/v3/pkg/capi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuditEventsClient_Get(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/v3/audit_events/event-guid", request.URL.Path)
		assert.Equal(t, "GET", request.Method)

		event := capi.AuditEvent{
			Resource: capi.Resource{
				GUID:      "event-guid",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Type: "audit.app.create",
			Actor: capi.AuditEventActor{
				GUID: "user-guid",
				Type: "user",
				Name: "test@example.com",
			},
			Target: capi.AuditEventTarget{
				GUID: "app-guid",
				Type: "app",
				Name: "test-app",
			},
			Data: map[string]interface{}{
				"request": map[string]interface{}{
					"name":       "test-app",
					"space_guid": "space-guid",
				},
			},
			Space: &capi.AuditEventSpace{
				GUID: "space-guid",
				Name: "test-space",
			},
			Organization: &capi.AuditEventOrganization{
				GUID: "org-guid",
				Name: "test-org",
			},
		}

		_ = json.NewEncoder(writer).Encode(event)
	}))
	defer server.Close()

	client, err := New(context.Background(), &capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	event, err := client.AuditEvents().Get(context.Background(), "event-guid")
	require.NoError(t, err)
	assert.Equal(t, "event-guid", event.GUID)
	assert.Equal(t, "audit.app.create", event.Type)
	assert.Equal(t, "user-guid", event.Actor.GUID)
	assert.Equal(t, "user", event.Actor.Type)
	assert.Equal(t, "test@example.com", event.Actor.Name)
	assert.Equal(t, "app-guid", event.Target.GUID)
	assert.Equal(t, "app", event.Target.Type)
	assert.Equal(t, "test-app", event.Target.Name)
	assert.Equal(t, "space-guid", event.Space.GUID)
	assert.Equal(t, "test-space", event.Space.Name)
}

//nolint:funlen // Test functions can be longer for comprehensive testing
func TestAuditEventsClient_List(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		params    *capi.QueryParams
		response  string
		wantErr   bool
		wantTotal int
	}{
		{
			name:   "successful list",
			params: nil,
			response: `{
				"pagination": {
					"total_results": 2,
					"total_pages": 1,
					"first": {"href": "/v3/audit_events?page=1&per_page=50"},
					"last": {"href": "/v3/audit_events?page=1&per_page=50"},
					"next": null,
					"previous": null
				},
				"resources": [
					{
						"guid": "event-1",
						"created_at": "2024-01-01T00:00:00Z",
						"updated_at": "2024-01-01T00:00:00Z",
						"type": "audit.app.update",
						"actor": {
							"guid": "user-1",
							"type": "user",
							"name": "admin"
						},
						"target": {
							"guid": "app-1",
							"type": "app",
							"name": "my-app"
						},
						"data": {},
						"space": {
							"guid": "space-1"
						},
						"organization": {
							"guid": "org-1"
						}
					},
					{
						"guid": "event-2",
						"created_at": "2024-01-02T00:00:00Z",
						"updated_at": "2024-01-02T00:00:00Z",
						"type": "audit.space.create",
						"actor": {
							"guid": "user-2",
							"type": "user",
							"name": "developer"
						},
						"target": {
							"guid": "space-2",
							"type": "space",
							"name": "dev-space"
						},
						"data": {}
					}
				]
			}`,
			wantErr:   false,
			wantTotal: 2,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/v3/audit_events", r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)

				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprint(writer, testCase.response)
			}))
			defer server.Close()

			client := NewTestClient(server.URL)
			ctx := context.Background()

			result, err := client.AuditEvents().List(ctx, testCase.params)
			if testCase.wantErr {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, testCase.wantTotal, result.Pagination.TotalResults)
		})
	}
}
