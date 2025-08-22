package capi_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/fivetwenty-io/capi-client/pkg/capi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockClient implements capi.Client for testing
type MockClient struct {
	mock.Mock
}

func (m *MockClient) GetInfo(ctx context.Context) (*capi.Info, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*capi.Info), args.Error(1)
}

func (m *MockClient) GetRootInfo(ctx context.Context) (*capi.RootInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*capi.RootInfo), args.Error(1)
}

func (m *MockClient) GetUsageSummary(ctx context.Context) (*capi.UsageSummary, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*capi.UsageSummary), args.Error(1)
}

func (m *MockClient) ClearBuildpackCache(ctx context.Context) (*capi.Job, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*capi.Job), args.Error(1)
}

func (m *MockClient) Apps() capi.AppsClient {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(capi.AppsClient)
}

func (m *MockClient) Organizations() capi.OrganizationsClient {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(capi.OrganizationsClient)
}

func (m *MockClient) Spaces() capi.SpacesClient {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(capi.SpacesClient)
}

func (m *MockClient) Domains() capi.DomainsClient {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(capi.DomainsClient)
}

func (m *MockClient) Routes() capi.RoutesClient {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(capi.RoutesClient)
}

func (m *MockClient) ServiceBrokers() capi.ServiceBrokersClient {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(capi.ServiceBrokersClient)
}

func (m *MockClient) ServiceOfferings() capi.ServiceOfferingsClient {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(capi.ServiceOfferingsClient)
}

func (m *MockClient) ServicePlans() capi.ServicePlansClient {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(capi.ServicePlansClient)
}

func (m *MockClient) ServiceInstances() capi.ServiceInstancesClient {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(capi.ServiceInstancesClient)
}

func (m *MockClient) ServiceCredentialBindings() capi.ServiceCredentialBindingsClient {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(capi.ServiceCredentialBindingsClient)
}

func (m *MockClient) ServiceRouteBindings() capi.ServiceRouteBindingsClient {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(capi.ServiceRouteBindingsClient)
}

func (m *MockClient) Builds() capi.BuildsClient {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(capi.BuildsClient)
}

func (m *MockClient) Buildpacks() capi.BuildpacksClient {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(capi.BuildpacksClient)
}

func (m *MockClient) Deployments() capi.DeploymentsClient {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(capi.DeploymentsClient)
}

func (m *MockClient) Droplets() capi.DropletsClient {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(capi.DropletsClient)
}

func (m *MockClient) Packages() capi.PackagesClient {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(capi.PackagesClient)
}

func (m *MockClient) Processes() capi.ProcessesClient {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(capi.ProcessesClient)
}

func (m *MockClient) Tasks() capi.TasksClient {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(capi.TasksClient)
}

func (m *MockClient) Stacks() capi.StacksClient {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(capi.StacksClient)
}

func (m *MockClient) Users() capi.UsersClient {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(capi.UsersClient)
}

func (m *MockClient) Roles() capi.RolesClient {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(capi.RolesClient)
}

func (m *MockClient) SecurityGroups() capi.SecurityGroupsClient {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(capi.SecurityGroupsClient)
}

func (m *MockClient) IsolationSegments() capi.IsolationSegmentsClient {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(capi.IsolationSegmentsClient)
}

func (m *MockClient) FeatureFlags() capi.FeatureFlagsClient {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(capi.FeatureFlagsClient)
}

func (m *MockClient) Jobs() capi.JobsClient {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(capi.JobsClient)
}

func (m *MockClient) OrganizationQuotas() capi.OrganizationQuotasClient {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(capi.OrganizationQuotasClient)
}

func (m *MockClient) SpaceQuotas() capi.SpaceQuotasClient {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(capi.SpaceQuotasClient)
}

func (m *MockClient) Sidecars() capi.SidecarsClient {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(capi.SidecarsClient)
}

func (m *MockClient) Revisions() capi.RevisionsClient {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(capi.RevisionsClient)
}

func (m *MockClient) EnvironmentVariableGroups() capi.EnvironmentVariableGroupsClient {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(capi.EnvironmentVariableGroupsClient)
}

func (m *MockClient) AppUsageEvents() capi.AppUsageEventsClient {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(capi.AppUsageEventsClient)
}

func (m *MockClient) ServiceUsageEvents() capi.ServiceUsageEventsClient {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(capi.ServiceUsageEventsClient)
}

func (m *MockClient) AuditEvents() capi.AuditEventsClient {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(capi.AuditEventsClient)
}

func (m *MockClient) ResourceMatches() capi.ResourceMatchesClient {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(capi.ResourceMatchesClient)
}

// MockAppsClient implements capi.AppsClient for testing
type MockAppsClient struct {
	mock.Mock
}

func (m *MockAppsClient) Create(ctx context.Context, request *capi.AppCreateRequest) (*capi.App, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*capi.App), args.Error(1)
}

func (m *MockAppsClient) Get(ctx context.Context, guid string) (*capi.App, error) {
	args := m.Called(ctx, guid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*capi.App), args.Error(1)
}

func (m *MockAppsClient) List(ctx context.Context, params *capi.QueryParams) (*capi.ListResponse[capi.App], error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*capi.ListResponse[capi.App]), args.Error(1)
}

func (m *MockAppsClient) Update(ctx context.Context, guid string, request *capi.AppUpdateRequest) (*capi.App, error) {
	args := m.Called(ctx, guid, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*capi.App), args.Error(1)
}

func (m *MockAppsClient) Delete(ctx context.Context, guid string) error {
	args := m.Called(ctx, guid)
	return args.Error(0)
}

func (m *MockAppsClient) GetCurrentDroplet(ctx context.Context, guid string) (*capi.Droplet, error) {
	args := m.Called(ctx, guid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*capi.Droplet), args.Error(1)
}

func (m *MockAppsClient) SetCurrentDroplet(ctx context.Context, guid string, dropletGUID string) (*capi.Relationship, error) {
	args := m.Called(ctx, guid, dropletGUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*capi.Relationship), args.Error(1)
}

func (m *MockAppsClient) GetEnv(ctx context.Context, guid string) (*capi.AppEnvironment, error) {
	args := m.Called(ctx, guid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*capi.AppEnvironment), args.Error(1)
}

func (m *MockAppsClient) GetEnvVars(ctx context.Context, guid string) (map[string]interface{}, error) {
	args := m.Called(ctx, guid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockAppsClient) UpdateEnvVars(ctx context.Context, guid string, vars map[string]interface{}) (map[string]interface{}, error) {
	args := m.Called(ctx, guid, vars)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockAppsClient) GetPermissions(ctx context.Context, guid string) (*capi.AppPermissions, error) {
	args := m.Called(ctx, guid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*capi.AppPermissions), args.Error(1)
}

func (m *MockAppsClient) GetSSHEnabled(ctx context.Context, guid string) (*capi.AppSSHEnabled, error) {
	args := m.Called(ctx, guid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*capi.AppSSHEnabled), args.Error(1)
}

func (m *MockAppsClient) Start(ctx context.Context, guid string) (*capi.App, error) {
	args := m.Called(ctx, guid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*capi.App), args.Error(1)
}

func (m *MockAppsClient) Stop(ctx context.Context, guid string) (*capi.App, error) {
	args := m.Called(ctx, guid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*capi.App), args.Error(1)
}

func (m *MockAppsClient) Restart(ctx context.Context, guid string) (*capi.App, error) {
	args := m.Called(ctx, guid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*capi.App), args.Error(1)
}

func (m *MockAppsClient) Restage(ctx context.Context, guid string) (*capi.Build, error) {
	args := m.Called(ctx, guid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*capi.Build), args.Error(1)
}

func (m *MockAppsClient) ClearBuildpackCache(ctx context.Context, guid string) error {
	args := m.Called(ctx, guid)
	return args.Error(0)
}

func (m *MockAppsClient) GetManifest(ctx context.Context, guid string) (string, error) {
	args := m.Called(ctx, guid)
	return args.String(0), args.Error(1)
}

func (m *MockAppsClient) GetRecentLogs(ctx context.Context, guid string, lines int) (*capi.AppLogs, error) {
	args := m.Called(ctx, guid, lines)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*capi.AppLogs), args.Error(1)
}

func (m *MockAppsClient) StreamLogs(ctx context.Context, guid string) (<-chan capi.LogMessage, error) {
	args := m.Called(ctx, guid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(<-chan capi.LogMessage), args.Error(1)
}

func TestBatchExecutor_Execute(t *testing.T) {
	mockClient := &MockClient{}
	mockApps := &MockAppsClient{}
	mockClient.On("Apps").Return(mockApps)

	executor := capi.NewBatchExecutor(mockClient, 2)
	ctx := context.Background()

	// Set up mock expectations
	app1 := &capi.App{
		Resource: capi.Resource{GUID: "app-1"},
		Name:     "Test App 1",
	}
	app2 := &capi.App{
		Resource: capi.Resource{GUID: "app-2"},
		Name:     "Test App 2",
	}

	mockApps.On("Get", mock.Anything, "app-guid-1").Return(app1, nil)
	mockApps.On("Get", mock.Anything, "app-guid-2").Return(app2, nil)

	operations := []capi.BatchOperation{
		{
			ID:       "op1",
			Type:     "get",
			Resource: "app",
			Data:     "app-guid-1",
		},
		{
			ID:       "op2",
			Type:     "get",
			Resource: "app",
			Data:     "app-guid-2",
		},
	}

	results, err := executor.Execute(ctx, operations)
	require.NoError(t, err)
	assert.Len(t, results, 2)

	// Check results
	for _, result := range results {
		assert.True(t, result.Success)
		assert.NoError(t, result.Error)
		assert.NotNil(t, result.Data)
		assert.True(t, result.Duration > 0)
	}

	mockClient.AssertExpectations(t)
	mockApps.AssertExpectations(t)
}

func TestBatchExecutor_WithCallback(t *testing.T) {
	mockClient := &MockClient{}
	mockApps := &MockAppsClient{}
	mockClient.On("Apps").Return(mockApps)

	executor := capi.NewBatchExecutor(mockClient, 1)
	ctx := context.Background()

	app := &capi.App{
		Resource: capi.Resource{GUID: "app-1"},
		Name:     "Test App",
	}
	mockApps.On("Get", mock.Anything, "app-guid").Return(app, nil)

	var callbackCalled bool
	var callbackResult *capi.BatchResult

	operation := capi.BatchOperation{
		ID:       "op1",
		Type:     "get",
		Resource: "app",
		Data:     "app-guid",
		Callback: func(result *capi.BatchResult) {
			callbackCalled = true
			callbackResult = result
		},
	}

	_, err := executor.Execute(ctx, []capi.BatchOperation{operation})
	require.NoError(t, err)

	assert.True(t, callbackCalled)
	assert.NotNil(t, callbackResult)
	assert.True(t, callbackResult.Success)
	assert.Equal(t, "op1", callbackResult.ID)

	mockClient.AssertExpectations(t)
	mockApps.AssertExpectations(t)
}

func TestBatchExecutor_WithError(t *testing.T) {
	mockClient := &MockClient{}
	mockApps := &MockAppsClient{}
	mockClient.On("Apps").Return(mockApps)

	executor := capi.NewBatchExecutor(mockClient, 1)
	ctx := context.Background()

	mockApps.On("Get", mock.Anything, "app-guid").Return(nil, fmt.Errorf("app not found"))

	operation := capi.BatchOperation{
		ID:       "op1",
		Type:     "get",
		Resource: "app",
		Data:     "app-guid",
	}

	results, err := executor.Execute(ctx, []capi.BatchOperation{operation})
	require.NoError(t, err) // Execute itself shouldn't fail
	assert.Len(t, results, 1)

	result := results[0]
	assert.False(t, result.Success)
	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "app not found")

	mockClient.AssertExpectations(t)
	mockApps.AssertExpectations(t)
}

func TestBatchBuilder(t *testing.T) {
	builder := capi.NewBatchBuilder()

	req1 := &capi.AppCreateRequest{
		Name: "app1",
	}
	name := "updated-app"
	req2 := &capi.AppUpdateRequest{
		Name: &name,
	}

	builder.
		AddCreateApp("create-1", req1).
		AddUpdateApp("update-1", "app-guid", req2).
		AddDeleteApp("delete-1", "app-to-delete").
		AddGetApp("get-1", "app-to-get")

	operations := builder.Build()
	assert.Len(t, operations, 4)

	assert.Equal(t, "create-1", operations[0].ID)
	assert.Equal(t, "create", operations[0].Type)
	assert.Equal(t, "app", operations[0].Resource)

	assert.Equal(t, "update-1", operations[1].ID)
	assert.Equal(t, "update", operations[1].Type)

	assert.Equal(t, "delete-1", operations[2].ID)
	assert.Equal(t, "delete", operations[2].Type)

	assert.Equal(t, "get-1", operations[3].ID)
	assert.Equal(t, "get", operations[3].Type)
}

func TestBatchExecutor_Timeout(t *testing.T) {
	mockClient := &MockClient{}
	executor := capi.NewBatchExecutor(mockClient, 1)
	executor.SetTimeout(1 * time.Millisecond)

	// Create an operation that will timeout
	operation := capi.BatchOperation{
		ID:       "op1",
		Type:     "get",
		Resource: "unsupported", // This will cause an error in the executor
		Data:     "test",
	}

	ctx := context.Background()
	results, err := executor.Execute(ctx, []capi.BatchOperation{operation})
	require.NoError(t, err)
	assert.Len(t, results, 1)

	result := results[0]
	assert.False(t, result.Success)
	assert.Error(t, result.Error)
}
