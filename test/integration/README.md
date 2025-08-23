# UAA Integration Tests

This directory contains integration tests for the UAA user management functionality in the capi CLI.

## Prerequisites

### UAA Server
You need access to a running UAA server for integration testing. This can be:
- A local UAA development server
- A Cloud Foundry UAA instance
- A dedicated UAA test environment

### Environment Variables

Set the following environment variables before running integration tests:

#### Required
- `UAA_ENDPOINT` - The UAA server URL (e.g., `https://uaa.example.com`)

#### Authentication (choose one method)

**Option 1: Admin User Credentials**
- `UAA_ADMIN_USER` - Admin username
- `UAA_ADMIN_PASSWORD` - Admin password
- `UAA_CLIENT_ID` - Client ID for password grant
- `UAA_CLIENT_SECRET` - Client secret for password grant

**Option 2: Client Credentials**
- `UAA_CLIENT_ID` - Client ID with admin privileges
- `UAA_CLIENT_SECRET` - Client secret

#### Optional
- `CAPI_BINARY_PATH` - Path to capi binary (defaults to `../../capi`)

### Example Environment Setup

```bash
export UAA_ENDPOINT="https://uaa.bosh-lite.com"
export UAA_CLIENT_ID="admin"
export UAA_CLIENT_SECRET="admin-secret"
export UAA_ADMIN_USER="admin"
export UAA_ADMIN_PASSWORD="admin-password"
```

## Running Integration Tests

### Build the capi binary first:
```bash
cd ~/w/fivetwenty-io/capi
make build
# or
go build -o capi ./cmd/capi
```

### Run all integration tests:
```bash
go test ./test/integration/... -tags=integration -v
```

### Run specific test suites:
```bash
# UAA integration tests only
go test ./test/integration/... -tags=integration -run TestUAAIntegrationSuite -v

# Command help tests only
go test ./test/integration/... -tags=integration -run TestUAACommandHelp -v
```

### Run with timeout:
```bash
go test ./test/integration/... -tags=integration -timeout 10m -v
```

## Test Structure

### UAAIntegrationTestSuite
Comprehensive end-to-end testing covering:

1. **Context Management** - Target setting, context display, server info
2. **Token Management** - OAuth2 grant flows, JWT key retrieval
3. **User Lifecycle** - Create, read, update, activate, deactivate, delete users
4. **Group Management** - Create groups, manage membership
5. **Client Management** - OAuth client CRUD operations
6. **Utility Commands** - Direct API access, user info
7. **Error Handling** - Unauthenticated requests, non-existent resources

### Individual Command Tests
- **Command Help Testing** - Verifies all commands have proper help text
- **Flag Validation** - Ensures all expected flags are present
- **Error Case Testing** - Tests proper error handling

## Test Data Cleanup

The test suite automatically:
- Generates unique names using timestamps to avoid conflicts
- Cleans up created test resources in `TearDownSuite`
- Uses `--force` flags to bypass confirmation prompts

## Security Considerations

Integration tests:
- Never commit credentials to version control
- Use temporary test accounts when possible
- Clean up all test data after completion
- Mask sensitive information in test output

## Troubleshooting

### Common Issues

1. **UAA Connection Errors**
   - Verify `UAA_ENDPOINT` is correct and accessible
   - Check SSL certificate issues (use `--skip-ssl-validation` if needed)

2. **Authentication Errors**
   - Verify client credentials have proper scopes
   - Ensure admin user has necessary permissions

3. **Binary Not Found**
   - Set `CAPI_BINARY_PATH` to correct location
   - Ensure binary is built and executable

4. **Permission Errors**
   - Verify UAA client has admin scopes
   - Check user has proper permissions for operations

### Debug Mode

Set `CAPI_VERBOSE=true` to enable verbose output during testing.
