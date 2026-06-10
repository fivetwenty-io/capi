package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	internalhttp "github.com/fivetwenty-io/capi/v3/internal/http"
	"github.com/fivetwenty-io/capi/v3/pkg/capi"
)

// Real CF answers async endpoints with 202 Accepted, an EMPTY body, and the
// job reference in the Location header (per the CF v3 OpenAPI spec). Every
// operation in this table previously json.Unmarshal'ed the empty body and
// failed with "parsing job response: unexpected end of JSON input"
// (cloudfoundry/stratos#5431).
func TestAsyncJobOperations_Empty202BodyWithLocationHeader(t *testing.T) {
	t.Parallel()

	const jobGUID = "job-from-location"

	newServer := func() *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			writer.Header().Set("Location", "/v3/jobs/"+jobGUID)
			writer.WriteHeader(http.StatusAccepted)
			// Empty body — what real CF sends on async responses.
		}))
	}

	managedSICreate := &capi.ServiceInstanceCreateRequest{
		Type: "managed",
		Name: "my-instance",
		Relationships: capi.ServiceInstanceRelationships{
			Space:       capi.Relationship{Data: &capi.RelationshipData{GUID: "space-guid"}},
			ServicePlan: &capi.Relationship{Data: &capi.RelationshipData{GUID: "plan-guid"}},
		},
	}

	cases := []struct {
		name string
		call func(t *testing.T, baseURL string) (interface{}, error)
	}{
		{
			name: "service instance create (managed)",
			call: func(_ *testing.T, baseURL string) (interface{}, error) {
				c := NewServiceInstancesClient(internalhttp.NewClient(baseURL, nil))
				return c.Create(context.Background(), managedSICreate)
			},
		},
		{
			name: "service instance update (managed)",
			call: func(_ *testing.T, baseURL string) (interface{}, error) {
				c := NewServiceInstancesClient(internalhttp.NewClient(baseURL, nil))
				return c.Update(context.Background(), "si-guid", &capi.ServiceInstanceUpdateRequest{
					Parameters: map[string]interface{}{"foo": "bar"},
				})
			},
		},
		{
			name: "service broker create",
			call: func(_ *testing.T, baseURL string) (interface{}, error) {
				c := NewServiceBrokersClient(internalhttp.NewClient(baseURL, nil))
				return c.Create(context.Background(), &capi.ServiceBrokerCreateRequest{Name: "broker"})
			},
		},
		{
			name: "service broker update (catalog sync)",
			call: func(_ *testing.T, baseURL string) (interface{}, error) {
				c := NewServiceBrokersClient(internalhttp.NewClient(baseURL, nil))
				name := "broker"
				return c.Update(context.Background(), "broker-guid", &capi.ServiceBrokerUpdateRequest{Name: &name})
			},
		},
		{
			name: "service credential binding create",
			call: func(_ *testing.T, baseURL string) (interface{}, error) {
				c := NewServiceCredentialBindingsClient(internalhttp.NewClient(baseURL, nil))
				return c.Create(context.Background(), &capi.ServiceCredentialBindingCreateRequest{Type: "app"})
			},
		},
		{
			name: "service route binding create",
			call: func(_ *testing.T, baseURL string) (interface{}, error) {
				c := NewServiceRouteBindingsClient(internalhttp.NewClient(baseURL, nil))
				return c.Create(context.Background(), &capi.ServiceRouteBindingCreateRequest{})
			},
		},
		{
			name: "space apply manifest",
			call: func(_ *testing.T, baseURL string) (interface{}, error) {
				c := NewSpacesClient(internalhttp.NewClient(baseURL, nil))
				return c.ApplyManifest(context.Background(), "space-guid", "applications: []")
			},
		},
		{
			name: "space delete unmapped routes",
			call: func(_ *testing.T, baseURL string) (interface{}, error) {
				c := NewSpacesClient(internalhttp.NewClient(baseURL, nil))
				return c.DeleteUnmappedRoutes(context.Background(), "space-guid")
			},
		},
		{
			name: "clear buildpack cache",
			call: func(t *testing.T, baseURL string) (interface{}, error) {
				c, err := New(context.Background(), &capi.Config{APIEndpoint: baseURL})
				require.NoError(t, err)
				return c.ClearBuildpackCache(context.Background())
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			server := newServer()
			defer server.Close()

			out, err := tc.call(t, server.URL)
			require.NoError(t, err)
			job, ok := out.(*capi.Job)
			require.True(t, ok, "expected *capi.Job, got %T", out)
			require.NotNil(t, job)
			assert.Equal(t, jobGUID, job.GUID)
		})
	}
}

// A 202 body carrying the Job resource (older CF / proxies / emulators) must
// keep working and win over the Location header — the body has full job
// state, the header only the GUID.
func TestAsyncJobOperations_202BodyStillPreferred(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.Header().Set("Location", "/v3/jobs/job-from-location")
		writer.WriteHeader(http.StatusAccepted)
		_, _ = writer.Write([]byte(`{"guid":"job-from-body","operation":"service_instance.update","state":"PROCESSING"}`))
	}))
	defer server.Close()

	c := NewServiceInstancesClient(internalhttp.NewClient(server.URL, nil))
	out, err := c.Update(context.Background(), "si-guid", &capi.ServiceInstanceUpdateRequest{
		Parameters: map[string]interface{}{"foo": "bar"},
	})
	require.NoError(t, err)
	job, ok := out.(*capi.Job)
	require.True(t, ok, "expected *capi.Job, got %T", out)
	assert.Equal(t, "job-from-body", job.GUID)
	assert.Equal(t, "PROCESSING", job.State)
}

// Neither a parseable body nor a Location header — surface a parse error so
// callers see the contract violation rather than a zero-value Job.
func TestAsyncJobOperations_Empty202NoLocationErrors(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	c := NewServiceInstancesClient(internalhttp.NewClient(server.URL, nil))
	_, err := c.Update(context.Background(), "si-guid", &capi.ServiceInstanceUpdateRequest{
		Parameters: map[string]interface{}{"foo": "bar"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing job response")
}
