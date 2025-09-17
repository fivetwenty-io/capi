package capi

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/fivetwenty-io/capi/v3/internal/constants"
)

// Static errors for err113 compliance.
var (
	ErrUnsupportedResourceType        = errors.New("unsupported resource type")
	ErrUnsupportedOperationType       = errors.New("unsupported operation type")
	ErrInvalidDataTypeApp             = errors.New("invalid data type for app operation")
	ErrInvalidDataTypeSpace           = errors.New("invalid data type for space operation")
	ErrInvalidDataTypeOrg             = errors.New("invalid data type for org operation")
	ErrInvalidDataTypeRoute           = errors.New("invalid data type for route operation")
	ErrInvalidDataTypeServiceInstance = errors.New("invalid data type for service instance operation")
	ErrTransactionFailed              = errors.New("transaction failed")
)

// UpdateDataWrapper wraps update data with GUID for consistent handling.
type UpdateDataWrapper[T any] struct {
	GUID    string
	Request *T
}

// handleCrudOperation is a helper that handles common CRUD pattern.
func handleCrudOperation(
	operation BatchOperation,
	createFunc func() (interface{}, error),
	updateFunc func() (interface{}, error),
	deleteFunc func() (interface{}, error),
	getFunc func() (interface{}, error),
) *BatchResult {
	result := &BatchResult{ID: operation.ID}

	switch operation.Type {
	case "create":
		data, err := createFunc()
		result.Success = err == nil
		result.Data = data
		result.Error = err
	case "update":
		data, err := updateFunc()
		result.Success = err == nil
		result.Data = data
		result.Error = err
	case "delete":
		data, err := deleteFunc()
		result.Success = err == nil
		result.Data = data
		result.Error = err
	case "get":
		data, err := getFunc()
		result.Success = err == nil
		result.Data = data
		result.Error = err
	default:
		result.Error = fmt.Errorf("%w: %s", ErrUnsupportedOperationType, operation.Type)
	}

	return result
}

// CRUDOperationConfig holds configuration for CRUD operations.
type CRUDOperationConfig struct {
	InvalidDataTypeErr error
	CreateFunc         func(ctx context.Context, operation BatchOperation) (interface{}, error)
	UpdateFunc         func(ctx context.Context, operation BatchOperation) (interface{}, error)
	DeleteFunc         func(ctx context.Context, operation BatchOperation) (interface{}, error)
	GetFunc            func(ctx context.Context, operation BatchOperation) (interface{}, error)
}

// ResourceClientOps defines the operations available for a resource client.
type ResourceClientOps[TCreateRequest, TUpdateRequest, TResponse any] interface {
	Create(ctx context.Context, request *TCreateRequest) (*TResponse, error)
	Update(ctx context.Context, guid string, request *TUpdateRequest) (*TResponse, error)
	Delete(ctx context.Context, guid string) (*Job, error)
	Get(ctx context.Context, guid string) (*TResponse, error)
}

// createCRUDOperationConfig creates a generic CRUD operation configuration.
func createCRUDOperationConfig[TCreateRequest, TUpdateRequest, TResponse any](
	invalidDataTypeErr error,
	client ResourceClientOps[TCreateRequest, TUpdateRequest, TResponse],
) CRUDOperationConfig {
	return CRUDOperationConfig{
		InvalidDataTypeErr: invalidDataTypeErr,
		CreateFunc: func(ctx context.Context, operation BatchOperation) (interface{}, error) {
			if req, ok := operation.Data.(*TCreateRequest); ok {
				return client.Create(ctx, req)
			}

			return nil, fmt.Errorf("%w create", invalidDataTypeErr)
		},
		UpdateFunc: func(ctx context.Context, operation BatchOperation) (interface{}, error) {
			if data, ok := operation.Data.(*UpdateDataWrapper[TUpdateRequest]); ok {
				return client.Update(ctx, data.GUID, data.Request)
			}

			return nil, fmt.Errorf("%w update", invalidDataTypeErr)
		},
		DeleteFunc: func(ctx context.Context, operation BatchOperation) (interface{}, error) {
			if guid, ok := operation.Data.(string); ok {
				return client.Delete(ctx, guid)
			}

			return nil, fmt.Errorf("%w delete", invalidDataTypeErr)
		},
		GetFunc: func(ctx context.Context, operation BatchOperation) (interface{}, error) {
			if guid, ok := operation.Data.(string); ok {
				return client.Get(ctx, guid)
			}

			return nil, fmt.Errorf("%w get", invalidDataTypeErr)
		},
	}
}

// BatchOperation represents a single operation in a batch.
type BatchOperation struct {
	ID       string
	Type     string // "create", "update", "delete", "get"
	Resource string // "app", "space", "org", etc.
	Data     interface{}
	Callback func(result *BatchResult)
}

// BatchResult represents the result of a batch operation.
type BatchResult struct {
	ID       string
	Success  bool
	Data     interface{}
	Error    error
	Duration time.Duration
}

// BatchExecutor executes batch operations.
type BatchExecutor struct {
	client      Client
	concurrency int
	timeout     time.Duration
}

// NewBatchExecutor creates a new batch executor.
func NewBatchExecutor(client Client, concurrency int) *BatchExecutor {
	if concurrency <= 0 {
		concurrency = 5
	}

	return &BatchExecutor{
		client:      client,
		concurrency: concurrency,
		timeout:     constants.DefaultHTTPTimeout,
	}
}

// SetTimeout sets the timeout for batch operations.
func (b *BatchExecutor) SetTimeout(timeout time.Duration) {
	b.timeout = timeout
}

// Execute runs a batch of operations.
func (b *BatchExecutor) Execute(ctx context.Context, operations []BatchOperation) ([]BatchResult, error) {
	results := make([]BatchResult, len(operations))

	var waitGroup sync.WaitGroup

	semaphore := make(chan struct{}, b.concurrency)

	for index, operation := range operations {
		waitGroup.Add(1)

		go func(index int, operation BatchOperation) {
			defer waitGroup.Done()

			// Acquire semaphore
			semaphore <- struct{}{}

			defer func() { <-semaphore }()

			// Execute operation with timeout
			opCtx, cancel := context.WithTimeout(ctx, b.timeout)
			defer cancel()

			start := time.Now()
			result := b.executeOperation(opCtx, operation)
			result.Duration = time.Since(start)
			results[index] = *result

			// Call callback if provided
			if operation.Callback != nil {
				operation.Callback(result)
			}
		}(index, operation)
	}

	waitGroup.Wait()

	return results, nil
}

// executeGenericCrudOperation handles generic CRUD operations using the provided configuration.
func (b *BatchExecutor) executeGenericCrudOperation(ctx context.Context, operation BatchOperation, config CRUDOperationConfig) *BatchResult {
	return handleCrudOperation(operation,
		func() (interface{}, error) { return config.CreateFunc(ctx, operation) },
		func() (interface{}, error) { return config.UpdateFunc(ctx, operation) },
		func() (interface{}, error) { return config.DeleteFunc(ctx, operation) },
		func() (interface{}, error) { return config.GetFunc(ctx, operation) },
	)
}

// createSpaceOperationConfig creates CRUD operation configuration for spaces.
func (b *BatchExecutor) createSpaceOperationConfig() CRUDOperationConfig {
	return createCRUDOperationConfig(ErrInvalidDataTypeSpace, b.client.Spaces())
}

// createOrgOperationConfig creates CRUD operation configuration for organizations.
func (b *BatchExecutor) createOrgOperationConfig() CRUDOperationConfig {
	return createCRUDOperationConfig(ErrInvalidDataTypeOrg, b.client.Organizations())
}

// createRouteOperationConfig creates CRUD operation configuration for routes.
func (b *BatchExecutor) createRouteOperationConfig() CRUDOperationConfig {
	return createCRUDOperationConfig(ErrInvalidDataTypeRoute, b.client.Routes())
}

// executeOperation executes a single operation.
func (b *BatchExecutor) executeOperation(ctx context.Context, operation BatchOperation) *BatchResult {
	result := &BatchResult{
		ID: operation.ID,
	}

	switch operation.Resource {
	case "app":
		result = b.executeAppOperation(ctx, operation)
	case "space":
		result = b.executeSpaceOperation(ctx, operation)
	case "organization":
		result = b.executeOrgOperation(ctx, operation)
	case "route":
		result = b.executeRouteOperation(ctx, operation)
	case "service_instance":
		result = b.executeServiceInstanceOperation(ctx, operation)
	default:
		result.Success = false
		result.Error = fmt.Errorf("%w: %s", ErrUnsupportedResourceType, operation.Resource)
	}

	return result
}

// executeAppOperation handles app operations using the common CRUD helper.
func (b *BatchExecutor) executeAppOperation(ctx context.Context, operation BatchOperation) *BatchResult {
	return handleCrudOperation(operation,
		func() (interface{}, error) {
			if req, ok := operation.Data.(*AppCreateRequest); ok {
				return b.client.Apps().Create(ctx, req)
			}

			return nil, fmt.Errorf("%w create", ErrInvalidDataTypeApp)
		},
		func() (interface{}, error) {
			if data, ok := operation.Data.(*UpdateDataWrapper[AppUpdateRequest]); ok {
				return b.client.Apps().Update(ctx, data.GUID, data.Request)
			}

			return nil, fmt.Errorf("%w update", ErrInvalidDataTypeApp)
		},
		func() (interface{}, error) {
			if guid, ok := operation.Data.(string); ok {
				err := b.client.Apps().Delete(ctx, guid)
				if err != nil {
					return nil, fmt.Errorf("failed to delete app: %w", err)
				}

				return nil, nil
			}

			return nil, fmt.Errorf("%w delete", ErrInvalidDataTypeApp)
		},
		func() (interface{}, error) {
			if guid, ok := operation.Data.(string); ok {
				return b.client.Apps().Get(ctx, guid)
			}

			return nil, fmt.Errorf("%w get", ErrInvalidDataTypeApp)
		},
	)
}

// executeSpaceOperation handles space operations using the common CRUD helper.
func (b *BatchExecutor) executeSpaceOperation(ctx context.Context, operation BatchOperation) *BatchResult {
	config := b.createSpaceOperationConfig()

	return b.executeGenericCrudOperation(ctx, operation, config)
}

// executeOrgOperation handles organization operations using the common CRUD helper.
func (b *BatchExecutor) executeOrgOperation(ctx context.Context, operation BatchOperation) *BatchResult {
	config := b.createOrgOperationConfig()

	return b.executeGenericCrudOperation(ctx, operation, config)
}

// executeRouteOperation handles route operations using the common CRUD helper.
func (b *BatchExecutor) executeRouteOperation(ctx context.Context, operation BatchOperation) *BatchResult {
	config := b.createRouteOperationConfig()

	return b.executeGenericCrudOperation(ctx, operation, config)
}

// executeServiceInstanceOperation handles service instance operations with special return handling.
func (b *BatchExecutor) executeServiceInstanceOperation(ctx context.Context, operation BatchOperation) *BatchResult {
	return handleCrudOperation(operation,
		func() (interface{}, error) {
			if req, ok := operation.Data.(*ServiceInstanceCreateRequest); ok {
				return b.client.ServiceInstances().Create(ctx, req)
			}

			return nil, fmt.Errorf("%w create", ErrInvalidDataTypeServiceInstance)
		},
		func() (interface{}, error) {
			if data, ok := operation.Data.(*UpdateDataWrapper[ServiceInstanceUpdateRequest]); ok {
				return b.client.ServiceInstances().Update(ctx, data.GUID, data.Request)
			}

			return nil, fmt.Errorf("%w update", ErrInvalidDataTypeServiceInstance)
		},
		func() (interface{}, error) {
			if guid, ok := operation.Data.(string); ok {
				return b.client.ServiceInstances().Delete(ctx, guid)
			}

			return nil, fmt.Errorf("%w delete", ErrInvalidDataTypeServiceInstance)
		},
		func() (interface{}, error) {
			if guid, ok := operation.Data.(string); ok {
				return b.client.ServiceInstances().Get(ctx, guid)
			}

			return nil, fmt.Errorf("%w get", ErrInvalidDataTypeServiceInstance)
		},
	)
}

// BatchBuilder helps build batch operations.
type BatchBuilder struct {
	operations []BatchOperation
}

// NewBatchBuilder creates a new batch builder.
func NewBatchBuilder() *BatchBuilder {
	return &BatchBuilder{
		operations: make([]BatchOperation, 0),
	}
}

// AddCreateApp adds an app creation operation.
func (b *BatchBuilder) AddCreateApp(id string, request *AppCreateRequest) *BatchBuilder {
	b.operations = append(b.operations, BatchOperation{
		ID:       id,
		Type:     "create",
		Resource: "app",
		Data:     request,
	})

	return b
}

// AddUpdateApp adds an app update operation.
func (b *BatchBuilder) AddUpdateApp(id, guid string, request *AppUpdateRequest) *BatchBuilder {
	b.operations = append(b.operations, BatchOperation{
		ID:       id,
		Type:     "update",
		Resource: "app",
		Data: &UpdateDataWrapper[AppUpdateRequest]{
			GUID:    guid,
			Request: request,
		},
	})

	return b
}

// AddDeleteApp adds an app deletion operation.
func (b *BatchBuilder) AddDeleteApp(id, guid string) *BatchBuilder {
	b.operations = append(b.operations, BatchOperation{
		ID:       id,
		Type:     "delete",
		Resource: "app",
		Data:     guid,
	})

	return b
}

// AddGetApp adds an app get operation.
func (b *BatchBuilder) AddGetApp(id, guid string) *BatchBuilder {
	b.operations = append(b.operations, BatchOperation{
		ID:       id,
		Type:     "get",
		Resource: "app",
		Data:     guid,
	})

	return b
}

// AddCreateSpace adds a space creation operation.
func (b *BatchBuilder) AddCreateSpace(id string, request *SpaceCreateRequest) *BatchBuilder {
	b.operations = append(b.operations, BatchOperation{
		ID:       id,
		Type:     "create",
		Resource: "space",
		Data:     request,
	})

	return b
}

// AddUpdateSpace adds a space update operation.
func (b *BatchBuilder) AddUpdateSpace(id, guid string, request *SpaceUpdateRequest) *BatchBuilder {
	b.operations = append(b.operations, BatchOperation{
		ID:       id,
		Type:     "update",
		Resource: "space",
		Data: &UpdateDataWrapper[SpaceUpdateRequest]{
			GUID:    guid,
			Request: request,
		},
	})

	return b
}

// AddDeleteSpace adds a space deletion operation.
func (b *BatchBuilder) AddDeleteSpace(id, guid string) *BatchBuilder {
	b.operations = append(b.operations, BatchOperation{
		ID:       id,
		Type:     "delete",
		Resource: "space",
		Data:     guid,
	})

	return b
}

// AddCreateOrganization adds an organization creation operation.
func (b *BatchBuilder) AddCreateOrganization(id string, request *OrganizationCreateRequest) *BatchBuilder {
	b.operations = append(b.operations, BatchOperation{
		ID:       id,
		Type:     "create",
		Resource: "organization",
		Data:     request,
	})

	return b
}

// AddOperation adds a custom operation.
func (b *BatchBuilder) AddOperation(operation BatchOperation) *BatchBuilder {
	b.operations = append(b.operations, operation)

	return b
}

// Build returns the built operations.
func (b *BatchBuilder) Build() []BatchOperation {
	return b.operations
}

// BatchTransaction represents a transactional batch of operations.
type BatchTransaction struct {
	operations []BatchOperation
	results    []BatchResult
	executor   *BatchExecutor
	rollback   bool
}

// NewBatchTransaction creates a new batch transaction.
func NewBatchTransaction(executor *BatchExecutor) *BatchTransaction {
	return &BatchTransaction{
		executor:   executor,
		operations: make([]BatchOperation, 0),
		rollback:   true,
	}
}

// Add adds an operation to the transaction.
func (t *BatchTransaction) Add(operation BatchOperation) *BatchTransaction {
	t.operations = append(t.operations, operation)

	return t
}

// SetRollback sets whether to rollback on failure.
func (t *BatchTransaction) SetRollback(rollback bool) *BatchTransaction {
	t.rollback = rollback

	return t
}

// Execute executes the transaction.
func (t *BatchTransaction) Execute(ctx context.Context) ([]BatchResult, error) {
	results, err := t.executor.Execute(ctx, t.operations)
	t.results = results

	// Check for failures
	var failedOps []string

	for _, result := range results {
		if !result.Success {
			failedOps = append(failedOps, result.ID)
		}
	}

	// If there were failures and rollback is enabled
	if len(failedOps) > 0 && t.rollback {
		// Attempt to rollback successful operations
		t.performRollback(ctx)

		return results, fmt.Errorf("%w, %d operations failed: %v", ErrTransactionFailed, len(failedOps), failedOps)
	}

	return results, err
}

// performRollback attempts to rollback successful operations.
func (t *BatchTransaction) performRollback(ctx context.Context) {
	// This is a simplified rollback - in practice, this would need to be more sophisticated
	var rollbackOps []BatchOperation

	for i, result := range t.results {
		if result.Success {
			original := t.operations[i]
			// Create inverse operation
			switch original.Type {
			case "create":
				// Delete what was created
				if original.Resource == "app" || original.Resource == "space" ||
					original.Resource == "organization" || original.Resource == "route" {
					// Extract GUID from result data if possible
					// This would need proper type assertions based on resource type
					rollbackOps = append(rollbackOps, BatchOperation{
						ID:       "rollback_" + original.ID,
						Type:     "delete",
						Resource: original.Resource,
						Data:     result.Data, // This would need proper GUID extraction
					})
				}
			case "delete":
				// Can't easily recreate deleted resources
				// Log this for manual intervention
			case "update":
				// Would need to store original state to rollback updates
				// Log this for manual intervention
			}
		}
	}

	// Execute rollback operations if any
	if len(rollbackOps) > 0 {
		_, _ = t.executor.Execute(ctx, rollbackOps)
	}
}
