package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fivetwenty-io/capi/v3/pkg/capi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuditEventsClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/audit_events/event-guid", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

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

		_ = json.NewEncoder(w).Encode(event)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
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

func TestAuditEventsClient_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/audit_events", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "10", r.URL.Query().Get("per_page"))

		response := capi.ListResponse[capi.AuditEvent]{
			Pagination: capi.Pagination{
				TotalResults: 2,
				TotalPages:   1,
			},
			Resources: []capi.AuditEvent{
				{
					Resource: capi.Resource{GUID: "event-1"},
					Type:     "audit.app.create",
					Actor: capi.AuditEventActor{
						GUID: "user-guid-1",
						Type: "user",
						Name: "user1@example.com",
					},
					Target: capi.AuditEventTarget{
						GUID: "app-guid-1",
						Type: "app",
						Name: "app-1",
					},
					Space: &capi.AuditEventSpace{
						GUID: "space-guid-1",
						Name: "space-1",
					},
					Organization: &capi.AuditEventOrganization{
						GUID: "org-guid-1",
						Name: "org-1",
					},
				},
				{
					Resource: capi.Resource{GUID: "event-2"},
					Type:     "audit.app.delete",
					Actor: capi.AuditEventActor{
						GUID: "user-guid-2",
						Type: "user",
						Name: "user2@example.com",
					},
					Target: capi.AuditEventTarget{
						GUID: "app-guid-2",
						Type: "app",
						Name: "app-2",
					},
					Space: &capi.AuditEventSpace{
						GUID: "space-guid-2",
						Name: "space-2",
					},
					Organization: &capi.AuditEventOrganization{
						GUID: "org-guid-2",
						Name: "org-2",
					},
				},
			},
		}

		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(&capi.Config{APIEndpoint: server.URL})
	require.NoError(t, err)

	params := capi.NewQueryParams().WithPage(1).WithPerPage(10)
	result, err := client.AuditEvents().List(context.Background(), params)

	require.NoError(t, err)
	assert.Len(t, result.Resources, 2)
	assert.Equal(t, "event-1", result.Resources[0].GUID)
	assert.Equal(t, "event-2", result.Resources[1].GUID)
	assert.Equal(t, "audit.app.create", result.Resources[0].Type)
	assert.Equal(t, "audit.app.delete", result.Resources[1].Type)
	assert.Equal(t, "user1@example.com", result.Resources[0].Actor.Name)
	assert.Equal(t, "user2@example.com", result.Resources[1].Actor.Name)
}
