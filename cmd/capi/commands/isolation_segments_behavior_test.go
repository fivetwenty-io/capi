//nolint:testpackage // RunE behavior tests need the unexported newClientFunc seam and command constructors
package commands

import (
	"testing"

	"github.com/fivetwenty-io/capi/v3/pkg/capi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsolationSegmentsGet_ByGUID(t *testing.T) { //nolint:paralleltest // serial: swaps process-global os.Stdout, viper, and newClientFunc
	withOutputFormat(t, OutputFormatJSON)

	recorder := &recordingIsoSegmentsClient{
		getResult: &capi.IsolationSegment{
			Resource: capi.Resource{GUID: "iso-guid-1"},
			Name:     "segment-one",
		},
	}
	withStubClient(t, &fakeClient{isolationSegments: recorder})

	out, err := runCommand(t, newIsolationSegmentsGetCommand(), "iso-guid-1")
	require.NoError(t, err)

	// Found directly by GUID — the name-based List fallback must not run.
	assert.Equal(t, "iso-guid-1", recorder.getGUID)
	assert.False(t, recorder.listCalled, "List fallback must not run when Get succeeds")
	assert.Contains(t, out, "segment-one")
}

func TestIsolationSegmentsGet_FallsBackToNameLookup(t *testing.T) { //nolint:paralleltest // serial: swaps process-global os.Stdout, viper, and newClientFunc
	withOutputFormat(t, OutputFormatJSON)

	recorder := &recordingIsoSegmentsClient{
		getErr: errClientBoom, // Get-by-GUID fails, forcing the name lookup.
		listResult: &capi.ListResponse[capi.IsolationSegment]{
			Resources: []capi.IsolationSegment{
				{Resource: capi.Resource{GUID: "iso-guid-2"}, Name: "by-name"},
			},
		},
	}
	withStubClient(t, &fakeClient{isolationSegments: recorder})

	out, err := runCommand(t, newIsolationSegmentsGetCommand(), "by-name")
	require.NoError(t, err)

	assert.Equal(t, "by-name", recorder.getGUID)
	assert.True(t, recorder.listCalled, "List fallback must run when Get fails")
	assert.Contains(t, out, "by-name")
}

func TestIsolationSegmentsGet_NotFoundReturnsError(t *testing.T) { //nolint:paralleltest // serial: swaps process-global os.Stdout, viper, and newClientFunc
	withOutputFormat(t, OutputFormatJSON)

	recorder := &recordingIsoSegmentsClient{
		getErr:     errClientBoom,
		listResult: &capi.ListResponse[capi.IsolationSegment]{Resources: nil},
	}
	withStubClient(t, &fakeClient{isolationSegments: recorder})

	_, err := runCommand(t, newIsolationSegmentsGetCommand(), "ghost")

	require.Error(t, err)
	require.ErrorIs(t, err, ErrIsolationSegmentNotFound)
}

func TestIsolationSegmentsGet_ListErrorPropagates(t *testing.T) { //nolint:paralleltest // serial: swaps process-global os.Stdout, viper, and newClientFunc
	withOutputFormat(t, OutputFormatJSON)

	recorder := &recordingIsoSegmentsClient{
		getErr:  errClientBoom,
		listErr: errClientBoom,
	}
	withStubClient(t, &fakeClient{isolationSegments: recorder})

	_, err := runCommand(t, newIsolationSegmentsGetCommand(), "anything")

	require.Error(t, err)
	require.ErrorIs(t, err, errClientBoom)
}
