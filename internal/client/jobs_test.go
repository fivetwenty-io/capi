package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	internalhttp "github.com/fivetwenty-io/capi-client/internal/http"
	"github.com/fivetwenty-io/capi-client/pkg/capi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJobsClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/jobs/job-guid", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		job := capi.Job{
			Resource: capi.Resource{
				GUID:      "job-guid",
				CreatedAt: time.Now().Add(-time.Minute),
				UpdatedAt: time.Now().Add(-30 * time.Second),
				Links: capi.Links{
					"self": capi.Link{
						Href: "/v3/jobs/job-guid",
					},
				},
			},
			Operation: "app.apply_manifest",
			State:     "COMPLETE",
			Errors:    []capi.APIError{},
			Warnings: []capi.Warning{
				{
					Detail: "Deprecated property detected: buildpack. App manifests must use buildpacks.",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(job)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	jobs := NewJobsClient(client.httpClient)

	job, err := jobs.Get(context.Background(), "job-guid")
	require.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, "job-guid", job.GUID)
	assert.Equal(t, "app.apply_manifest", job.Operation)
	assert.Equal(t, "COMPLETE", job.State)
	assert.Len(t, job.Warnings, 1)
	assert.Equal(t, "Deprecated property detected: buildpack. App manifests must use buildpacks.", job.Warnings[0].Detail)
}

func TestJobsClient_Get_Processing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/jobs/job-guid", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		job := capi.Job{
			Resource: capi.Resource{
				GUID:      "job-guid",
				CreatedAt: time.Now().Add(-time.Minute),
				UpdatedAt: time.Now(),
				Links: capi.Links{
					"self": capi.Link{
						Href: "/v3/jobs/job-guid",
					},
				},
			},
			Operation: "service_instance.create",
			State:     "PROCESSING",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(job)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	jobs := NewJobsClient(client.httpClient)

	job, err := jobs.Get(context.Background(), "job-guid")
	require.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, "job-guid", job.GUID)
	assert.Equal(t, "service_instance.create", job.Operation)
	assert.Equal(t, "PROCESSING", job.State)
	assert.Len(t, job.Errors, 0)
	assert.Len(t, job.Warnings, 0)
}

func TestJobsClient_Get_Failed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/jobs/job-guid", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		job := capi.Job{
			Resource: capi.Resource{
				GUID:      "job-guid",
				CreatedAt: time.Now().Add(-time.Minute),
				UpdatedAt: time.Now(),
				Links: capi.Links{
					"self": capi.Link{
						Href: "/v3/jobs/job-guid",
					},
				},
			},
			Operation: "service_broker.delete",
			State:     "FAILED",
			Errors: []capi.APIError{
				{
					Detail: "Service broker deletion failed: broker has service instances",
					Title:  "CF-ServiceBrokerNotRemovable",
					Code:   10001,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(job)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	jobs := NewJobsClient(client.httpClient)

	job, err := jobs.Get(context.Background(), "job-guid")
	require.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, "job-guid", job.GUID)
	assert.Equal(t, "service_broker.delete", job.Operation)
	assert.Equal(t, "FAILED", job.State)
	assert.Len(t, job.Errors, 1)
	assert.Equal(t, "Service broker deletion failed: broker has service instances", job.Errors[0].Detail)
	assert.Equal(t, "CF-ServiceBrokerNotRemovable", job.Errors[0].Title)
	assert.Equal(t, 10001, job.Errors[0].Code)
}

func TestJobsClient_PollUntilComplete_Success(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/jobs/job-guid", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		attempts++
		var job capi.Job

		// Simulate job transitioning from PROCESSING to COMPLETE
		if attempts <= 2 {
			job = capi.Job{
				Resource: capi.Resource{
					GUID:      "job-guid",
					CreatedAt: time.Now().Add(-time.Minute),
					UpdatedAt: time.Now(),
					Links: capi.Links{
						"self": capi.Link{
							Href: "/v3/jobs/job-guid",
						},
					},
				},
				Operation: "app.apply_manifest",
				State:     "PROCESSING",
			}
		} else {
			job = capi.Job{
				Resource: capi.Resource{
					GUID:      "job-guid",
					CreatedAt: time.Now().Add(-time.Minute),
					UpdatedAt: time.Now(),
					Links: capi.Links{
						"self": capi.Link{
							Href: "/v3/jobs/job-guid",
						},
					},
				},
				Operation: "app.apply_manifest",
				State:     "COMPLETE",
				Warnings: []capi.Warning{
					{
						Detail: "Manifest applied successfully",
					},
				},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(job)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	jobs := NewJobsClient(client.httpClient)

	// Use a shorter poll interval for testing
	jobs.pollInterval = 10 * time.Millisecond

	job, err := jobs.PollUntilComplete(context.Background(), "job-guid")
	require.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, "job-guid", job.GUID)
	assert.Equal(t, "app.apply_manifest", job.Operation)
	assert.Equal(t, "COMPLETE", job.State)
	assert.Equal(t, 3, attempts)
}

func TestJobsClient_PollUntilComplete_Failed(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/jobs/job-guid", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		attempts++
		var job capi.Job

		// Simulate job transitioning from PROCESSING to FAILED
		if attempts <= 1 {
			job = capi.Job{
				Resource: capi.Resource{
					GUID:      "job-guid",
					CreatedAt: time.Now().Add(-time.Minute),
					UpdatedAt: time.Now(),
					Links: capi.Links{
						"self": capi.Link{
							Href: "/v3/jobs/job-guid",
						},
					},
				},
				Operation: "service_instance.delete",
				State:     "PROCESSING",
			}
		} else {
			job = capi.Job{
				Resource: capi.Resource{
					GUID:      "job-guid",
					CreatedAt: time.Now().Add(-time.Minute),
					UpdatedAt: time.Now(),
					Links: capi.Links{
						"self": capi.Link{
							Href: "/v3/jobs/job-guid",
						},
					},
				},
				Operation: "service_instance.delete",
				State:     "FAILED",
				Errors: []capi.APIError{
					{
						Detail: "Service instance deletion failed",
						Title:  "CF-ServiceInstanceNotRemovable",
						Code:   10002,
					},
				},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(job)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	jobs := NewJobsClient(client.httpClient)

	// Use a shorter poll interval for testing
	jobs.pollInterval = 10 * time.Millisecond

	job, err := jobs.PollUntilComplete(context.Background(), "job-guid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "job failed")
	assert.NotNil(t, job)
	assert.Equal(t, "job-guid", job.GUID)
	assert.Equal(t, "service_instance.delete", job.Operation)
	assert.Equal(t, "FAILED", job.State)
	assert.Len(t, job.Errors, 1)
}

func TestJobsClient_PollUntilComplete_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/jobs/job-guid", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		// Always return PROCESSING
		job := capi.Job{
			Resource: capi.Resource{
				GUID:      "job-guid",
				CreatedAt: time.Now().Add(-time.Minute),
				UpdatedAt: time.Now(),
				Links: capi.Links{
					"self": capi.Link{
						Href: "/v3/jobs/job-guid",
					},
				},
			},
			Operation: "app.apply_manifest",
			State:     "PROCESSING",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(job)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	jobs := NewJobsClient(client.httpClient)

	// Use a shorter poll interval and timeout for testing
	jobs.pollInterval = 10 * time.Millisecond
	jobs.pollTimeout = 50 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	job, err := jobs.PollUntilComplete(ctx, "job-guid")
	require.Error(t, err)
	assert.True(t, err.Error() == "timeout waiting for job to complete: context deadline exceeded" ||
		strings.Contains(err.Error(), "context deadline exceeded"),
		"Expected timeout error, got: %v", err)
	if job != nil {
		assert.Equal(t, "PROCESSING", job.State)
	}
}
