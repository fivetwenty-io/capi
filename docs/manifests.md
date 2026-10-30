# Manifest Management

Manifest management provides declarative application deployment using YAML manifests, similar to the CF CLI.

## Commands

```bash
# Apply a manifest to deploy/update applications
capi manifests apply SPACE_GUID --file manifest.yml [--wait]

# Generate a manifest from an existing application  
capi manifests generate APP_GUID [--output-file manifest.yml]

# Create a diff between current state and proposed manifest
capi manifests diff SPACE_GUID --file manifest.yml
```

## Examples

**Basic manifest (manifest.yml):**
```yaml
applications:
- name: my-app
  memory: 1G
  instances: 2
  buildpacks:
  - go_buildpack
  env:
    NODE_ENV: production
  services:
  - postgres-service
  routes:
  - route: myapp.example.com
```

**Apply manifest:**
```bash
capi manifests apply space-123 --file manifest.yml --wait
```

**Generate manifest from existing app:**
```bash
capi manifests generate app-456 --output-file generated-manifest.yml
```

**Preview changes before applying:**
```bash
capi manifests diff space-123 --file manifest.yml
```

## Output Formats

All manifest commands support multiple output formats:
- `--output table` (default) - Human-readable tables
- `--output json` - JSON for scripting
- `--output yaml` - YAML for configuration management

## CI/CD Pipeline Integration

```bash
#!/bin/bash
# Deploy application using manifest
capi target -o my-org -s production
capi manifests apply $(capi target --space-guid) --file production-manifest.yml --wait

# Enable production features
capi apps features enable my-app revisions
capi spaces features enable production ssh
```

## Development Workflow

```bash
# Development setup
capi target -o my-org -s development

# Deploy app with development manifest
capi manifests apply $(capi target --space-guid) --file dev-manifest.yml

# Enable development features
capi apps features enable my-app ssh
capi spaces features enable development ssh
```

## Troubleshooting

### Common Issues

**Manifest Syntax Errors:**
```bash
Error: Invalid YAML: Manifest contains invalid YAML syntax  
Solution: Validate YAML syntax and check indentation
```

### Getting Help

```bash
# Get help for any command
capi manifests --help

# Get detailed help for subcommands
capi manifests apply --help
```

## Private Buildpack Credentials

CF API 3.236.0 adds a top-level `credentials` manifest key, keyed by host, for fetching private buildpacks (buildpack lifecycle) and buildpack images (cnb lifecycle). The existing `cnb-credentials` key remains as an alias, and giving both keys is a validation error. Buildpack entries take a username and password, while cnb entries take either a username and password or a token.

In the Go library, `ManifestApplication.Credentials` and `ManifestApplication.CNBCredentials` carry these values as a `capi.ManifestCredentials`, which has four states.

| State | How we get it | What it marshals as | What the API does with it |
|-------|---------------|---------------------|---------------------------|
| Unset | The zero value | The key is omitted | Keeps the stored credentials |
| Hosts | `capi.NewManifestCredentials(hosts)` | An object keyed by host, or `{}` for an empty map | Stores the credentials given |
| Clear | `capi.ClearManifestCredentials()` | `null` | Clears the stored credentials |
| Redacted | Decoding a manifest that holds `***` | `***` | Rejects the manifest, because credentials must be an object |

The parent struct's `omitempty` (YAML) and `omitzero` (JSON) tags are what drop an unset field. Marshaling an unset `ManifestCredentials` on its own returns `capi.ErrManifestCredentialsUnset` rather than writing a `null` that would clear the stored credentials by accident.

```go
app := capi.ManifestApplication{
    Name: "my-app",
    Credentials: capi.NewManifestCredentials(map[string]capi.ManifestCredential{
        "github.com": {Username: "acme-bot", Password: token},
    }),
}
```

The API never returns stored credentials. A generated manifest and an app's lifecycle data show `***` in their place, and the manifest diff shows `[PRIVATE DATA HIDDEN]`. A generated manifest therefore decodes with `Credentials.Redacted` set, and marshaling it again writes `***` back, which the API rejects on apply. Before we apply a generated manifest, we either set real credentials or reset the field to its zero value so that the stored ones are kept.

YAML decoding has one limitation. The YAML decoder never calls a field's unmarshaler for a null value, so it can't tell an explicit `null` from a missing key. That means `credentials: null`, `credentials: ~`, and a null that arrives through an alias or a merge key all decode as the unset state, and marshaling the result omits the key instead of clearing the credentials. JSON decoding doesn't have this gap, and a JSON `null` decodes as Clear. When we edit a manifest in Go and need the clear to survive, we set `capi.ClearManifestCredentials()` before marshaling. The `capi manifests apply` command sends the file exactly as written, so a `null` in that file always reaches the API.

In JSON the alias is spelled `cnb_credentials`, which matches the other snake_case JSON keys on these types. The YAML key stays `cnb-credentials`.

## API Compatibility

All implemented features are compatible with Cloud Foundry API v3.236.0 and follow the official CF API specification for:

- `/v3/spaces/{guid}/actions/apply_manifest` - Manifest application
- `/v3/apps/{guid}/manifest` - Manifest generation  
- `/v3/spaces/{guid}/manifest_diff` - Manifest diffing

For complete API documentation, see the [Cloud Foundry API v3 specification](https://v3-apidocs.cloudfoundry.org/).