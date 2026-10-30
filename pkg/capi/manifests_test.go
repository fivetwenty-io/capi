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

func TestManifestVersion(t *testing.T) {
	t.Parallel()

	t.Run("unset version is omitted", func(t *testing.T) {
		t.Parallel()

		manifest := capi.Manifest{Applications: []capi.ManifestApplication{{Name: "a"}}}

		out, err := yaml.Marshal(manifest)
		require.NoError(t, err)
		assert.YAMLEq(t, "applications:\n- name: a\n", string(out))

		data, err := json.Marshal(manifest)
		require.NoError(t, err)
		assert.JSONEq(t, `{"applications":[{"name":"a"}]}`, string(data))
	})

	t.Run("version 1 is kept", func(t *testing.T) {
		t.Parallel()

		_, out := yamlRoundTrip(t, "version: 1\napplications:\n- name: a\n")
		assert.YAMLEq(t, "version: 1\napplications:\n- name: a\n", out)
	})
}

func TestManifestDocker(t *testing.T) {
	t.Parallel()

	t.Run("yaml object round-trips", func(t *testing.T) {
		t.Parallel()

		doc := "version: 1\napplications:\n- name: d\n  docker:\n    image: nginx\n    username: bob\n"

		manifest, out := yamlRoundTrip(t, doc)
		assert.Equal(t, &capi.ManifestDocker{Image: "nginx", Username: "bob"}, manifest.Applications[0].Docker)
		assert.YAMLEq(t, doc, out)
	})

	t.Run("json object round-trips", func(t *testing.T) {
		t.Parallel()

		doc := `{"version":1,"applications":[{"name":"d","docker":{"image":"nginx","username":"bob"}}]}`

		manifest, out := jsonRoundTrip(t, doc)
		assert.Equal(t, "nginx", manifest.Applications[0].Docker.Image)
		assert.JSONEq(t, doc, out)
	})

	t.Run("an empty object keeps the docker lifecycle", func(t *testing.T) {
		t.Parallel()

		doc := "version: 1\napplications:\n- name: d\n  docker: {}\n"

		manifest, out := yamlRoundTrip(t, doc)
		assert.NotNil(t, manifest.Applications[0].Docker)
		assert.YAMLEq(t, doc, out)
	})
}

func TestManifestMetadataNull(t *testing.T) {
	t.Parallel()

	t.Run("yaml null removes a key and round-trips", func(t *testing.T) {
		t.Parallel()

		doc := "version: 1\napplications:\n- name: a\n  metadata:\n" +
			"    labels: {team: a, gone: null}\n    annotations: {old: null}\n"

		manifest, out := yamlRoundTrip(t, doc)
		labels := manifest.Applications[0].Metadata.Labels
		require.Contains(t, labels, "gone")
		assert.Nil(t, labels["gone"])
		assert.Equal(t, "a", capi.StringValue(labels["team"]))
		assert.YAMLEq(t, doc, out)
	})

	t.Run("json null removes a key and round-trips", func(t *testing.T) {
		t.Parallel()

		doc := `{"version":1,"applications":[{"name":"a","metadata":{"labels":{"team":"a","gone":null}}}]}`

		_, out := jsonRoundTrip(t, doc)
		assert.JSONEq(t, doc, out)
	})
}

func TestManifestSidecarMemory(t *testing.T) {
	t.Parallel()

	t.Run("yaml keeps integers and unit strings", func(t *testing.T) {
		t.Parallel()

		doc := "version: 1\napplications:\n- name: a\n  sidecars:\n" +
			"  - {name: auth, process_types: [web], command: run-auth, memory: 800}\n" +
			"  - {name: up, process_types: [worker], command: ./up, memory: 2G}\n"

		manifest, out := yamlRoundTrip(t, doc)
		assert.Equal(t, capi.ManifestSidecarMemory("800"), manifest.Applications[0].Sidecars[0].Memory)
		assert.YAMLEq(t, doc, out)
	})

	t.Run("json keeps integers and unit strings", func(t *testing.T) {
		t.Parallel()

		doc := `{"version":1,"applications":[{"name":"a","sidecars":[` +
			`{"name":"auth","process_types":["web"],"command":"run-auth","memory":800},` +
			`{"name":"up","process_types":["worker"],"command":"./up","memory":"2G"}]}]}`

		_, out := jsonRoundTrip(t, doc)
		assert.JSONEq(t, doc, out)
	})

	t.Run("unset keys are omitted", func(t *testing.T) {
		t.Parallel()

		out, err := yaml.Marshal(capi.ManifestSidecar{Name: "s"})
		require.NoError(t, err)
		assert.YAMLEq(t, "name: s\n", string(out))

		data, err := json.Marshal(capi.ManifestSidecar{Name: "s"})
		require.NoError(t, err)
		assert.JSONEq(t, `{"name":"s"}`, string(data))
	})

	t.Run("leading zeros still marshal as a valid number", func(t *testing.T) {
		t.Parallel()

		data, err := json.Marshal(capi.ManifestSidecar{Name: "s", Memory: "0800"})
		require.NoError(t, err)
		assert.JSONEq(t, `{"name":"s","memory":800}`, string(data))
	})

	t.Run("json rejects other types", func(t *testing.T) {
		t.Parallel()

		var sidecar capi.ManifestSidecar

		require.Error(t, json.Unmarshal([]byte(`{"name":"s","memory":true}`), &sidecar))
	})
}
