package client

import (
	"encoding/json"
	"fmt"
	"strings"

	http_internal "github.com/fivetwenty-io/capi/v3/internal/http"
	"github.com/fivetwenty-io/capi/v3/pkg/capi"
)

// jobFromAsyncResponse extracts the async job reference from a CF v3
// 202 Accepted response.
//
// Real CF sends 202 with an EMPTY body and the job link in the Location
// header (per the CF v3 OpenAPI spec) — see /v3/service_instances PATCH,
// /v3/service_brokers POST, /v3/spaces/{guid}/actions/apply_manifest, etc.
// Some proxies, emulators and older servers instead include the Job
// resource as the response body.
//
// Resolution order:
//  1. Non-empty body that parses as a Job → return it (full job state:
//     GUID, operation, state).
//  2. Location header → return a Job with the GUID extracted from its
//     trailing path segment; callers poll via Jobs().Get.
//  3. Neither → the original body parse error, so contract violations
//     stay visible.
func jobFromAsyncResponse(resp *http_internal.Response, opLabel string) (*capi.Job, error) {
	var parseErr error
	if len(strings.TrimSpace(string(resp.Body))) > 0 {
		var job capi.Job
		parseErr = json.Unmarshal(resp.Body, &job)
		if parseErr == nil {
			return &job, nil
		}
	}

	location := resp.Headers.Get("Location")
	if location != "" {
		jobGUID := location
		if idx := strings.LastIndex(location, "/"); idx >= 0 {
			jobGUID = location[idx+1:]
		}
		if jobGUID != "" {
			return &capi.Job{Resource: capi.Resource{GUID: jobGUID}}, nil
		}
		return nil, fmt.Errorf("%s: malformed Location header %q", opLabel, location)
	}

	if parseErr == nil {
		// Empty body and no Location header — keep the historical error
		// shape so existing callers' error matching continues to work.
		var job capi.Job
		parseErr = json.Unmarshal(resp.Body, &job)
	}
	return nil, fmt.Errorf("parsing job response: %w", parseErr)
}
