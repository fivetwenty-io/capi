# Contributing to Cloud Foundry API v3 Client

We welcome contributions to the Cloud Foundry API v3 client library! This document provides guidelines for contributing code, documentation, and reporting issues.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Environment](#development-environment)
- [Code Style](#code-style)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)
- [Issue Reporting](#issue-reporting)
- [Documentation](#documentation)
- [Release Process](#release-process)

## Getting Started

### Prerequisites

- Go 1.19 or later
- Access to a Cloud Foundry environment for integration testing
- Git
- Make (optional, but recommended)

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/capi-client.git
   cd capi-client
   ```
3. Add the original repository as upstream:
   ```bash
   git remote add upstream https://github.com/fivetwenty-io/capi-client.git
   ```

## Development Environment

### Setup

```bash
# Install dependencies
go mod download

# Install development tools
make install-tools

# Run tests to verify setup
make test
```

### Project Structure

```
.
├── cmd/capi/           # CLI application
├── pkg/                # Public API packages
│   ├── capi/          # Core client interfaces and types
│   └── cfclient/      # Client implementation
├── internal/          # Internal packages
│   ├── auth/          # Authentication handling
│   ├── client/        # HTTP client implementation
│   └── errors/        # Error handling
├── examples/          # Example code
├── test/              # Test utilities and fixtures
└── docs/              # Documentation
```

### Environment Variables

Set these environment variables for development and testing:

```bash
export CF_API_ENDPOINT="https://api.your-cf-domain.com"
export CF_USERNAME="your-username"
export CF_PASSWORD="your-password"
export CF_SKIP_SSL_VALIDATION="false"
```

## Code Style

### Go Guidelines

We follow the standard Go code style and conventions:

- Use `go fmt` to format code
- Use `go vet` to check for common errors
- Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use meaningful variable and function names
- Add comments for exported functions and types

### Linting

We use several linters to maintain code quality:

```bash
# Run all linters
make lint

# Individual linters
golangci-lint run
go vet ./...
staticcheck ./...
```

### Code Organization

- Keep functions small and focused
- Use dependency injection for testability
- Separate concerns (HTTP handling, business logic, etc.)
- Use interfaces for external dependencies

### Error Handling

- Use the `capi.Error` type for CF API errors
- Wrap errors with context using `fmt.Errorf`
- Return errors from functions that can fail
- Handle errors at the appropriate level

```go
// Good error handling
func (c *client) GetOrganization(ctx context.Context, guid string) (*capi.Organization, error) {
    resp, err := c.httpClient.Get(ctx, "/v3/organizations/"+guid)
    if err != nil {
        return nil, fmt.Errorf("getting organization %s: %w", guid, err)
    }
    
    var org capi.Organization
    if err := json.Unmarshal(resp.Body, &org); err != nil {
        return nil, fmt.Errorf("parsing organization response: %w", err)
    }
    
    return &org, nil
}
```

## Testing

### Test Types

1. **Unit Tests**: Test individual functions and methods
2. **Integration Tests**: Test against real CF API
3. **Mock Tests**: Test with mocked dependencies

### Running Tests

```bash
# Run unit tests
make test

# Run tests with coverage
make test-coverage

# Run integration tests (requires CF environment)
make test-integration

# Run specific test package
go test ./pkg/capi/...

# Run tests with verbose output
go test -v ./...
```

### Writing Tests

#### Unit Tests

```go
func TestOrganizationList(t *testing.T) {
    tests := []struct {
        name        string
        params      *capi.QueryParams
        mockResp    string
        expectError bool
        expectCount int
    }{
        {
            name:        "successful list",
            params:      nil,
            mockResp:    `{"resources":[{"guid":"org-1","name":"test-org"}]}`,
            expectError: false,
            expectCount: 1,
        },
        {
            name:        "with filters",
            params:      capi.NewQueryParams().WithFilter("names", "test-org"),
            mockResp:    `{"resources":[]}`,
            expectError: false,
            expectCount: 0,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            client := newMockClient(tt.mockResp, 200)
            
            orgs, err := client.Organizations().List(context.Background(), tt.params)
            
            if tt.expectError {
                assert.Error(t, err)
                return
            }
            
            assert.NoError(t, err)
            assert.Len(t, orgs.Resources, tt.expectCount)
        })
    }
}
```

#### Integration Tests

```go
func TestOrganizationIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    client := newIntegrationClient(t)
    ctx := context.Background()
    
    // Create test organization
    org, err := client.Organizations().Create(ctx, &capi.OrganizationCreate{
        Name: "test-org-" + randomString(8),
    })
    require.NoError(t, err)
    
    defer func() {
        // Cleanup
        client.Organizations().Delete(ctx, org.GUID)
    }()
    
    // Test operations
    retrievedOrg, err := client.Organizations().Get(ctx, org.GUID)
    require.NoError(t, err)
    assert.Equal(t, org.Name, retrievedOrg.Name)
}
```

### Test Data

- Use the `test/fixtures/` directory for test data
- Create realistic but anonymized test data
- Use table-driven tests for multiple scenarios

### Mocks

We use generated mocks for testing:

```bash
# Generate mocks
make generate-mocks

# Manual mock generation
mockgen -source=pkg/capi/client.go -destination=test/mocks/mock_client.go
```

## Pull Request Process

### Before Submitting

1. Ensure all tests pass:
   ```bash
   make test
   make lint
   ```

2. Update documentation if needed
3. Add tests for new functionality
4. Update CHANGELOG.md if applicable

### Submission

1. Create a feature branch:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes and commit:
   ```bash
   git add .
   git commit -m "Add feature: your feature description"
   ```

3. Push to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

4. Create a pull request on GitHub

### Pull Request Template

```markdown
## Description

Brief description of changes.

## Type of Change

- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing

- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing completed

## Checklist

- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests pass locally
```

### Review Process

1. At least one maintainer must approve
2. All CI checks must pass
3. Address review feedback
4. Squash commits if requested

## Issue Reporting

### Bug Reports

Use the bug report template:

```markdown
**Describe the bug**
A clear description of the bug.

**To Reproduce**
Steps to reproduce the behavior:
1. Go to '...'
2. Click on '....'
3. See error

**Expected behavior**
What you expected to happen.

**Environment**
- Go version:
- Client version:
- CF API version:
- OS:

**Additional context**
Add any other context about the problem.
```

### Feature Requests

```markdown
**Is your feature request related to a problem?**
A clear description of what the problem is.

**Describe the solution you'd like**
A clear description of what you want to happen.

**Describe alternatives you've considered**
Other solutions you've considered.

**Additional context**
Add any other context about the feature request.
```

## Documentation

### API Documentation

- Document all public functions and types
- Include examples in doc comments
- Update API.md for significant changes

### Example Code

- Add examples for new features
- Keep examples simple and focused
- Test examples to ensure they work

### README Updates

- Update feature list for new capabilities
- Add new examples to quick start
- Update installation instructions if needed

## Release Process

### Version Numbering

We use [Semantic Versioning](https://semver.org/):

- MAJOR: Incompatible API changes
- MINOR: Backwards-compatible functionality
- PATCH: Backwards-compatible bug fixes

### Release Checklist

1. Update version in code
2. Update CHANGELOG.md
3. Create and test release candidate
4. Tag release
5. Update documentation
6. Announce release

### Changelog Format

```markdown
## [1.2.0] - 2023-12-01

### Added
- New feature X
- Support for Y

### Changed
- Improved performance of Z

### Fixed
- Bug in feature A

### Deprecated
- Feature B (will be removed in v2.0.0)

### Removed
- Deprecated feature C

### Security
- Fixed vulnerability in dependency D
```

## Community

### Communication

- GitHub Discussions for questions and ideas
- GitHub Issues for bugs and feature requests
- Code reviews for technical discussions

### Code of Conduct

This project adheres to a code of conduct. By participating, you are expected to uphold this code.

## Getting Help

- Check existing issues and documentation
- Ask questions in GitHub Discussions
- Reach out to maintainers if needed

## License

By contributing to this project, you agree that your contributions will be licensed under the Apache License 2.0.

## Thank You

Thank you for contributing to the Cloud Foundry API v3 client library! Your contributions help make this project better for everyone.