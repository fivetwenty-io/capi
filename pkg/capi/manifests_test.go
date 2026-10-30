package capi_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"

	"github.com/fivetwenty-io/capi/v3/pkg/capi"
)

// yamlRoundTrip decodes doc into a Manifest and marshals it back.
func yamlRoundTrip(t *testing.T, doc string) (capi.Manifest, string) {
	t.Helper()

	var manifest capi.Manifest

	require.NoError(t, yaml.Unmarshal([]byte(doc), &manifest))

	out, err := yaml.Marshal(manifest)
	require.NoError(t, err)

	return manifest, string(out)
}

// jsonRoundTrip decodes doc into a Manifest and marshals it back.
func jsonRoundTrip(t *testing.T, doc string) (capi.Manifest, string) {
	t.Helper()

	var manifest capi.Manifest

	require.NoError(t, json.Unmarshal([]byte(doc), &manifest))

	out, err := json.Marshal(manifest)
	require.NoError(t, err)

	return manifest, string(out)
}

func TestManifestServiceForms(t *testing.T) {
	t.Parallel()

	t.Run("yaml takes names and objects", func(t *testing.T) {
		t.Parallel()

		doc := "version: 1\napplications:\n- name: a\n  services:\n  - my-service1\n" +
			"  - name: s2\n    binding_name: b\n    parameters: {k: v}\n"

		manifest, out := yamlRoundTrip(t, doc)
		services := manifest.Applications[0].Services
		require.Len(t, services, 2)
		assert.Equal(t, capi.ManifestService{Name: "my-service1"}, services[0])
		assert.Equal(t, "b", services[1].BindingName)
		assert.YAMLEq(t, doc, out)
	})

	t.Run("json takes names and objects", func(t *testing.T) {
		t.Parallel()

		doc := `{"version":1,"applications":[{"name":"a","services":["my-service1",{"name":"s2","binding_name":"b"}]}]}`

		manifest, out := jsonRoundTrip(t, doc)
		services := manifest.Applications[0].Services
		require.Len(t, services, 2)
		assert.Equal(t, "my-service1", services[0].Name)
		assert.Equal(t, "b", services[1].BindingName)
		assert.JSONEq(t, doc, out)
	})

	t.Run("type errors keep the rest of the application", func(t *testing.T) {
		t.Parallel()

		var (
			manifest capi.Manifest
			typeErr  *yaml.TypeError
		)

		err := yaml.Unmarshal([]byte("applications:\n- name: a\n  services: [[x], s1]\n  memory: 1G\n"), &manifest)
		require.ErrorAs(t, err, &typeErr)
		assert.Equal(t, "1G", manifest.Applications[0].Memory)
		// The decoder drops sequence entries that fail, so s1 moves up.
		require.Len(t, manifest.Applications[0].Services, 1)
		assert.Equal(t, "s1", manifest.Applications[0].Services[0].Name)
	})

	t.Run("strict decoding reaches service objects", func(t *testing.T) {
		t.Parallel()

		var manifest capi.Manifest

		dec := yaml.NewDecoder(bytes.NewReader([]byte("applications:\n- name: a\n  services:\n  - {name: s, bogus: 1}\n")))
		dec.KnownFields(true)
		require.Error(t, dec.Decode(&manifest))
	})
}
