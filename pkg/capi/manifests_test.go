package capi_test

import (
	"bytes"
	"encoding/json"
	"strings"
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

func TestManifestApplicationKeys(t *testing.T) {
	t.Parallel()

	t.Run("yaml round-trips app keys", func(t *testing.T) {
		t.Parallel()

		doc := "version: 1\napplications:\n- name: a\n  lifecycle: cnb\n  buildpack: legacy_bp\n" +
			"  default-route: true\n  features: {ssh: true, revisions: false}\n" +
			"  health-check-invocation-timeout: 7\n"

		manifest, out := yamlRoundTrip(t, doc)
		app := manifest.Applications[0]
		assert.Equal(t, "cnb", app.Lifecycle)
		assert.Equal(t, map[string]bool{"ssh": true, "revisions": false}, app.Features)
		assert.YAMLEq(t, doc, out)
	})

	t.Run("json round-trips app keys", func(t *testing.T) {
		t.Parallel()

		doc := `{"version":1,"applications":[{"name":"a","lifecycle":"cnb","buildpack":"legacy_bp",` +
			`"default_route":true,"features":{"ssh":true},"health_check_invocation_timeout":7}]}`

		_, out := jsonRoundTrip(t, doc)
		assert.JSONEq(t, doc, out)
	})
}

func TestManifestProcessAndRouteKeys(t *testing.T) {
	t.Parallel()

	t.Run("yaml round-trips process and route keys", func(t *testing.T) {
		t.Parallel()

		doc := "version: 1\napplications:\n- name: a\n  routes:\n" +
			"  - route: a.example.com\n    options: {loadbalancing: hash, hash_header: X-User, hash_balance: 1.5}\n" +
			"  processes:\n  - type: web\n    user: vcap\n    timeout: 10\n    health-check-invocation-timeout: 3\n"

		manifest, out := yamlRoundTrip(t, doc)
		app := manifest.Applications[0]
		assert.Equal(t, "vcap", app.Processes[0].User)
		assert.Equal(t, 10, *app.Processes[0].Timeout)
		assert.Equal(t, "hash", app.Routes[0].Options["loadbalancing"])
		assert.YAMLEq(t, doc, out)
	})

	t.Run("json round-trips process and route keys", func(t *testing.T) {
		t.Parallel()

		doc := `{"version":1,"applications":[{"name":"a",` +
			`"routes":[{"route":"a.example.com","options":{"loadbalancing":"hash","hash_balance":1.5}}],` +
			`"processes":[{"type":"web","user":"vcap","timeout":10,"health_check_invocation_timeout":3}]}]}`

		_, out := jsonRoundTrip(t, doc)
		assert.JSONEq(t, doc, out)
	})
}

// referenceManifest is the example manifest from the CF API 3.236.0 app
// manifest reference.
const referenceManifest = `version: 1
applications:
- name: app1
  buildpacks:
  - ruby_buildpack
  - java_buildpack
  env:
    VAR1: value1
    VAR2: value2
  features:
    ssh: true
    revisions: true
    service-binding-k8s: false
    file-based-vcap-services: false
  routes:
  - route: route.example.com
  - route: another-route.example.com
    protocol: http2
  - route: route-with-options.example.com
    options:
      loadbalancing: least-connection
  services:
  - my-service1
  - my-service2
  - name: my-service-with-arbitrary-params
    binding_name: my-binding
    parameters:
      key1: value1
      key2: value2
  stack: cflinuxfs4
  metadata:
    annotations:
      contact: "dev@example.com ops@example.com"
    labels:
      sensitive: true
  processes:
  - type: web
    command: start-web.sh
    disk_quota: 512M
    health-check-type: http
    health-check-http-endpoint: /healthcheck
    health-check-invocation-timeout: 10
    health-check-interval: 5
    readiness-health-check-type: http
    readiness-health-check-http-endpoint: /ready
    readiness-health-check-invocation-timeout: 10
    readiness-health-check-interval: 5
    instances: 3
    memory: 500M
    log-rate-limit-per-second: 1KB
    timeout: 10
    user: vcap
  - type: worker
    command: start-worker.sh
    disk_quota: 1G
    health-check-type: process
    instances: 2
    memory: 256M
    log-rate-limit-per-second: 1KB
    timeout: 15
- name: app2
  env:
    VAR1: value1
  processes:
  - type: web
    instances: 1
    memory: 256M
    log-rate-limit-per-second: 1KB
  sidecars:
  - name: authenticator
    process_types: [ 'web', 'worker' ]
    command: bundle exec run-authenticator
    memory: 800M
  - name: upcaser
    process_types: [ 'worker' ]
    command: ./tr-server
    memory: 2G
`

// generatedManifest is the example response of the CF API 3.236.0
// generate manifest endpoint.
const generatedManifest = `---
applications:
- name: my-app
  stack: cflinuxfs4
  features:
    ssh: true
    revisions: true
    service-binding-k8s: false
    file-based-vcap-services: false
  services:
  - my-service
  routes:
  - route: my-app.example.com
    protocol: http1
  processes:
  - type: web
    instances: 2
    memory: 512M
    log-rate-limit-per-second: 1KB
    disk_quota: 1024M
    health-check-type: http
    health-check-http-endpoint: /healthy
    health-check-invocation-timeout: 10
    health-check-interval: 5
    readiness-health-check-type: http
    readiness-health-check-http-endpoint: /ready
    readiness-health-check-invocation-timeout: 20
    readiness-health-check-interval: 5
`

func TestManifestReferenceExamples(t *testing.T) {
	t.Parallel()

	t.Run("the reference manifest decodes strictly and round-trips", func(t *testing.T) {
		t.Parallel()

		var manifest capi.Manifest

		dec := yaml.NewDecoder(strings.NewReader(referenceManifest))
		dec.KnownFields(true)
		require.NoError(t, dec.Decode(&manifest))

		out, err := yaml.Marshal(manifest)
		require.NoError(t, err)
		// Label values are strings, so the boolean label comes back quoted.
		want := strings.Replace(referenceManifest, "sensitive: true", `sensitive: "true"`, 1)
		assert.YAMLEq(t, want, string(out))
	})

	t.Run("the generated manifest decodes strictly and round-trips", func(t *testing.T) {
		t.Parallel()

		var manifest capi.Manifest

		dec := yaml.NewDecoder(strings.NewReader(generatedManifest))
		dec.KnownFields(true)
		require.NoError(t, dec.Decode(&manifest))

		out, err := yaml.Marshal(manifest)
		require.NoError(t, err)
		assert.YAMLEq(t, generatedManifest, string(out))
	})
}
