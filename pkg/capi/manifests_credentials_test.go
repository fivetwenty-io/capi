package capi_test

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"

	"github.com/fivetwenty-io/capi/v3/pkg/capi"
)

// generatedRedactedManifest is what GET /v3/apps/:guid/manifest returns for
// an app with stored buildpack credentials (cloud_controller_ng PR 5547).
const generatedRedactedManifest = `applications:
- name: a
  lifecycle: buildpack
  buildpacks: [https://github.com/acme/bp]
  stack: cflinuxfs4
  credentials: '***'
`

func credentialHosts() map[string]capi.ManifestCredential {
	return map[string]capi.ManifestCredential{
		"github.com": {Username: "bot", Password: "pw"},
		"ghcr.io":    {Token: "tok"},
	}
}

func TestManifestCredentialsJSONMarshal(t *testing.T) {
	t.Parallel()

	t.Run("omitted key stays omitted", func(t *testing.T) {
		t.Parallel()

		data, err := json.Marshal(capi.ManifestApplication{Name: "a"})
		require.NoError(t, err)
		assert.JSONEq(t, `{"name":"a"}`, string(data))
	})

	t.Run("clear marshals explicit null", func(t *testing.T) {
		t.Parallel()

		app := capi.ManifestApplication{Name: "a", Credentials: capi.ClearManifestCredentials()}
		data, err := json.Marshal(app)
		require.NoError(t, err)
		assert.JSONEq(t, `{"name":"a","credentials":null}`, string(data))
	})

	t.Run("hosts marshal as object on both keys", func(t *testing.T) {
		t.Parallel()

		app := capi.ManifestApplication{
			Name:           "a",
			Credentials:    capi.NewManifestCredentials(credentialHosts()),
			CNBCredentials: capi.NewManifestCredentials(credentialHosts()),
		}
		data, err := json.Marshal(app)
		require.NoError(t, err)
		assert.JSONEq(t,
			`{"name":"a",
			"credentials":{"github.com":{"username":"bot","password":"pw"},"ghcr.io":{"token":"tok"}},
			"cnb_credentials":{"github.com":{"username":"bot","password":"pw"},"ghcr.io":{"token":"tok"}}}`,
			string(data))
	})

	t.Run("zero value refuses to marshal", func(t *testing.T) {
		t.Parallel()

		type unomitted struct {
			C capi.ManifestCredentials `json:"credentials"`
		}

		_, err := json.Marshal(capi.ManifestCredentials{})
		require.ErrorIs(t, err, capi.ErrManifestCredentialsUnset)

		_, err = json.Marshal(unomitted{})
		require.ErrorIs(t, err, capi.ErrManifestCredentialsUnset)
	})
}

func TestManifestCredentialsJSONUnmarshal(t *testing.T) {
	t.Parallel()

	t.Run("unmarshal tells null from absent", func(t *testing.T) {
		t.Parallel()

		var absent, null, set capi.ManifestApplication

		require.NoError(t, json.Unmarshal([]byte(`{"name":"a"}`), &absent))
		require.NoError(t, json.Unmarshal([]byte(`{"name":"a","credentials":null}`), &null))
		require.NoError(t, json.Unmarshal(
			[]byte(`{"name":"a","cnb_credentials":{"ghcr.io":{"token":"tok"}}}`), &set))

		assert.True(t, absent.Credentials.IsZero())
		assert.True(t, null.Credentials.Clear)
		assert.Nil(t, null.Credentials.Hosts)
		assert.False(t, set.CNBCredentials.Clear)
		assert.Equal(t, "tok", set.CNBCredentials.Hosts["ghcr.io"].Token)
	})

	t.Run("redacted value decodes and round-trips", func(t *testing.T) {
		t.Parallel()

		doc := `{"name":"a","credentials":"***","memory":"1G"}`

		var app capi.ManifestApplication

		require.NoError(t, json.Unmarshal([]byte(doc), &app))
		assert.True(t, app.Credentials.Redacted)
		assert.Equal(t, "1G", app.Memory)

		out, err := json.Marshal(app)
		require.NoError(t, err)
		assert.JSONEq(t, doc, string(out))
	})

	t.Run("unmarshal rejects other strings", func(t *testing.T) {
		t.Parallel()

		var app capi.ManifestApplication

		require.Error(t, json.Unmarshal([]byte(`{"name":"a","credentials":"x"}`), &app))
	})

	t.Run("empty object round-trips", func(t *testing.T) {
		t.Parallel()

		doc := `{"name":"a","credentials":{}}`

		var app capi.ManifestApplication

		require.NoError(t, json.Unmarshal([]byte(doc), &app))
		assert.NotNil(t, app.Credentials.Hosts)
		assert.False(t, app.Credentials.IsZero())

		out, err := json.Marshal(app)
		require.NoError(t, err)
		assert.JSONEq(t, doc, string(out))

		built, err := json.Marshal(capi.ManifestApplication{
			Name:        "a",
			Credentials: capi.NewManifestCredentials(map[string]capi.ManifestCredential{}),
		})
		require.NoError(t, err)
		assert.JSONEq(t, doc, string(built))
	})
}

func TestManifestCredentialsYAMLMarshal(t *testing.T) {
	t.Parallel()

	t.Run("omitted key stays omitted", func(t *testing.T) {
		t.Parallel()

		data, err := yaml.Marshal(capi.ManifestApplication{Name: "a"})
		require.NoError(t, err)
		assert.YAMLEq(t, "name: a\n", string(data))
	})

	t.Run("clear marshals explicit null", func(t *testing.T) {
		t.Parallel()

		app := capi.ManifestApplication{Name: "a", CNBCredentials: capi.ClearManifestCredentials()}
		data, err := yaml.Marshal(app)
		require.NoError(t, err)
		assert.Contains(t, string(data), "cnb-credentials: null")
	})

	t.Run("hosts marshal as mapping", func(t *testing.T) {
		t.Parallel()

		app := capi.ManifestApplication{Name: "a", Credentials: capi.NewManifestCredentials(credentialHosts())}
		data, err := yaml.Marshal(app)
		require.NoError(t, err)
		assert.YAMLEq(t,
			"name: a\ncredentials:\n  github.com: {username: bot, password: pw}\n  ghcr.io: {token: tok}\n",
			string(data))
	})

	t.Run("zero value refuses to marshal", func(t *testing.T) {
		t.Parallel()

		type unomitted struct {
			C capi.ManifestCredentials `yaml:"credentials"`
		}

		_, err := yaml.Marshal(map[string]any{"credentials": capi.ManifestCredentials{}})
		require.ErrorIs(t, err, capi.ErrManifestCredentialsUnset)

		_, err = yaml.Marshal(unomitted{})
		require.ErrorIs(t, err, capi.ErrManifestCredentialsUnset)
	})
}

func TestManifestCredentialsYAMLUnmarshal(t *testing.T) {
	t.Parallel()

	t.Run("hosts decode", func(t *testing.T) {
		t.Parallel()

		var app capi.ManifestApplication

		require.NoError(t, yaml.Unmarshal(
			[]byte("name: a\ncredentials:\n  github.com:\n    username: bot\n    password: pw\n"), &app))
		assert.Equal(t, "bot", app.Credentials.Hosts["github.com"].Username)
		assert.False(t, app.Credentials.Clear)
	})

	t.Run("generated redacted manifest decodes and round-trips", func(t *testing.T) {
		t.Parallel()

		var manifest capi.Manifest

		require.NoError(t, yaml.Unmarshal([]byte(generatedRedactedManifest), &manifest))
		require.Len(t, manifest.Applications, 1)

		app := manifest.Applications[0]
		assert.Equal(t, "a", app.Name)
		assert.Equal(t, testStackName, app.Stack)
		assert.True(t, app.Credentials.Redacted)
		assert.Nil(t, app.Credentials.Hosts)

		out, err := yaml.Marshal(app)
		require.NoError(t, err)
		assert.Contains(t, string(out), `credentials: '***'`)
	})

	t.Run("empty mapping round-trips", func(t *testing.T) {
		t.Parallel()

		var app capi.ManifestApplication

		require.NoError(t, yaml.Unmarshal([]byte("name: a\ncredentials: {}\n"), &app))
		assert.NotNil(t, app.Credentials.Hosts)
		assert.False(t, app.Credentials.IsZero())

		out, err := yaml.Marshal(app)
		require.NoError(t, err)
		assert.YAMLEq(t, "name: a\ncredentials: {}\n", string(out))
	})

	t.Run("non-mapping is a type error that keeps other fields", func(t *testing.T) {
		t.Parallel()

		var (
			app     capi.ManifestApplication
			typeErr *yaml.TypeError
		)

		err := yaml.Unmarshal([]byte("name: a\ncredentials: [x]\nmemory: 1G\n"), &app)
		require.ErrorAs(t, err, &typeErr)
		assert.Equal(t, "a", app.Name)
		assert.Equal(t, "1G", app.Memory)
	})
}

// TestManifestCredentialsYAMLNullForms locks in the documented limitation:
// YAML never hands a null to the field's unmarshaler, so every null form
// decodes as the zero value (key omitted) rather than as Clear.
func TestManifestCredentialsYAMLNullForms(t *testing.T) {
	t.Parallel()

	docs := map[string]string{
		"plain":       "applications:\n- name: a\n  credentials: null\n",
		"tilde":       "applications:\n- name: a\n  credentials: ~\n",
		"empty":       "applications:\n- name: a\n  credentials:\n",
		"tagged":      "applications:\n- name: a\n  credentials: !!null ''\n",
		"alias":       "x: &n null\napplications:\n- name: a\n  credentials: *n\n",
		"merge":       "x: &b {credentials: null}\napplications:\n- <<: *b\n  name: a\n",
		"aliased app": "x: &b {name: a, credentials: null}\napplications:\n- *b\n",
	}

	for name, doc := range docs {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var manifest capi.Manifest

			require.NoError(t, yaml.Unmarshal([]byte(doc), &manifest))
			require.Len(t, manifest.Applications, 1)
			assert.Equal(t, "a", manifest.Applications[0].Name)
			assert.True(t, manifest.Applications[0].Credentials.IsZero())
		})
	}
}

// extendedApp is how a caller adds keys of its own on top of the library
// type.
type extendedApp struct {
	capi.ManifestApplication `yaml:",inline"`

	Extra string `yaml:"extra"`
}

func TestManifestApplicationYAMLDecoding(t *testing.T) {
	t.Parallel()

	t.Run("no custom unmarshaler on the application", func(t *testing.T) {
		t.Parallel()

		_, found := reflect.TypeFor[*capi.ManifestApplication]().MethodByName("UnmarshalYAML")
		assert.False(t, found)
	})

	t.Run("credentials use the decoder-agnostic unmarshaler", func(t *testing.T) {
		t.Parallel()

		// Every yaml.v3 module path (and yaml.v2) calls this shape, while
		// UnmarshalYAML(*yaml.Node) binds to one module's Node type.
		var creds any = &capi.ManifestCredentials{}

		_, ok := creds.(interface {
			UnmarshalYAML(unmarshal func(any) error) error
		})
		assert.True(t, ok)
	})

	t.Run("type errors aggregate across applications", func(t *testing.T) {
		t.Parallel()

		var (
			manifest capi.Manifest
			typeErr  *yaml.TypeError
		)

		doc := "applications:\n- name: a\n  instances: abc\n  memory: 1G\n" +
			"- name: b\n  credentials: [x]\n  memory: 2G\n- name: c\n  memory: 3G\n"
		err := yaml.Unmarshal([]byte(doc), &manifest)
		require.ErrorAs(t, err, &typeErr)
		assert.Len(t, typeErr.Errors, 2)
		require.Len(t, manifest.Applications, 3)

		for idx, want := range []struct{ name, memory string }{{"a", "1G"}, {"b", "2G"}, {"c", "3G"}} {
			assert.Equal(t, want.name, manifest.Applications[idx].Name)
			assert.Equal(t, want.memory, manifest.Applications[idx].Memory)
		}
	})

	t.Run("embedding keeps outer fields", func(t *testing.T) {
		t.Parallel()

		var app extendedApp

		require.NoError(t, yaml.Unmarshal([]byte("name: a\nextra: kept\ncredentials: {h: {token: t}}\n"), &app))
		assert.Equal(t, "a", app.Name)
		assert.Equal(t, "kept", app.Extra)
		assert.Equal(t, "t", app.Credentials.Hosts["h"].Token)
	})
}

func TestManifestApplicationYAMLStrictAndOverlay(t *testing.T) {
	t.Parallel()

	t.Run("strict decoding reaches applications and credentials", func(t *testing.T) {
		t.Parallel()

		docs := map[string]string{
			"application key": "applications:\n- name: a\n  bogus: 1\n",
			"credential key":  "applications:\n- name: a\n  credentials: {h: {usr: u}}\n",
		}

		for name, doc := range docs {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				var manifest capi.Manifest

				dec := yaml.NewDecoder(bytes.NewReader([]byte(doc)))
				dec.KnownFields(true)
				require.Error(t, dec.Decode(&manifest))
			})
		}
	})

	t.Run("decoding overlays a prefilled application", func(t *testing.T) {
		t.Parallel()

		app := capi.ManifestApplication{Name: "x", Memory: "256M", Stack: testStackName}

		require.NoError(t, yaml.Unmarshal([]byte("name: y\n"), &app))
		assert.Equal(t, "y", app.Name)
		assert.Equal(t, "256M", app.Memory)
		assert.Equal(t, testStackName, app.Stack)
	})
}
