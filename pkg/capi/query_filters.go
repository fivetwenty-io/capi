package capi

import (
	"strconv"
	"strings"
)

// This file defines typed List filter options for the CF v3 collection
// endpoints that accept resource-specific filters but no include parameter.
// Each resource exposes a sealed XListOption interface so only options
// defined in this package may be passed to that resource's List method.
//
// Cross-cutting parameters (order_by, label_selector, created_ats,
// updated_ats, pagination) are intentionally NOT duplicated here: they are
// expressed through the *QueryParams argument that every List method also
// accepts (QueryParams.WithOrderBy, WithLabelSelector, WithFilter, and the
// package-level WithTimestampFilter helper). The options below cover the
// resource-specific entity and enumerated-value filters where typing
// prevents the most mistakes.

// joinKind comma-joins a slice of string-kinded values (typed enums) into a
// single CF filter value.
func joinKind[T ~string](vals []T) string {
	parts := make([]string, len(vals))
	for i, v := range vals {
		parts[i] = string(v)
	}

	return strings.Join(parts, ",")
}

// ---- builds ----

// BuildListOption configures GET /v3/builds.
type BuildListOption interface {
	QueryOption
	buildList()
}

type buildListScalar struct{ scalarOption }

func (buildListScalar) buildList() {}

// BuildState is a CF v3 build lifecycle state.
type BuildState string

// Valid build states (CF v3).
const (
	BuildStateStaging BuildState = "STAGING"
	BuildStateStaged  BuildState = "STAGED"
	BuildStateFailed  BuildState = "FAILED"
)

// WithBuildGUIDs filters builds by GUID.
func WithBuildGUIDs(guids ...string) BuildListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return buildListScalar{scalarOption{"guids", strings.Join(guids, ",")}}
}

// WithBuildAppGUIDs filters builds by app GUID.
func WithBuildAppGUIDs(guids ...string) BuildListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return buildListScalar{scalarOption{"app_guids", strings.Join(guids, ",")}}
}

// WithBuildPackageGUIDs filters builds by package GUID.
func WithBuildPackageGUIDs(guids ...string) BuildListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return buildListScalar{scalarOption{"package_guids", strings.Join(guids, ",")}}
}

// WithBuildStates filters builds by lifecycle state.
func WithBuildStates(states ...BuildState) BuildListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return buildListScalar{scalarOption{"states", joinKind(states)}}
}

// ---- droplets ----

// DropletListOption configures GET /v3/droplets.
type DropletListOption interface {
	QueryOption
	dropletList()
}

type dropletListScalar struct{ scalarOption }

func (dropletListScalar) dropletList() {}

// DropletState is a CF v3 droplet lifecycle state.
type DropletState string

// Valid droplet states (CF v3).
const (
	DropletStateAwaitingUpload   DropletState = "AWAITING_UPLOAD"
	DropletStateProcessingUpload DropletState = "PROCESSING_UPLOAD"
	DropletStateCopying          DropletState = "COPYING"
	DropletStateStaging          DropletState = "STAGING"
	DropletStateStaged           DropletState = "STAGED"
	DropletStateFailed           DropletState = "FAILED"
	DropletStateExpired          DropletState = "EXPIRED"
)

// WithDropletGUIDs filters droplets by GUID.
func WithDropletGUIDs(guids ...string) DropletListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return dropletListScalar{scalarOption{"guids", strings.Join(guids, ",")}}
}

// WithDropletAppGUIDs filters droplets by app GUID.
func WithDropletAppGUIDs(guids ...string) DropletListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return dropletListScalar{scalarOption{"app_guids", strings.Join(guids, ",")}}
}

// WithDropletPackageGUIDs filters droplets by package GUID.
func WithDropletPackageGUIDs(guids ...string) DropletListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return dropletListScalar{scalarOption{"package_guids", strings.Join(guids, ",")}}
}

// WithDropletSpaceGUIDs filters droplets by space GUID.
func WithDropletSpaceGUIDs(guids ...string) DropletListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return dropletListScalar{scalarOption{"space_guids", strings.Join(guids, ",")}}
}

// WithDropletOrganizationGUIDs filters droplets by organization GUID.
func WithDropletOrganizationGUIDs(guids ...string) DropletListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return dropletListScalar{scalarOption{"organization_guids", strings.Join(guids, ",")}}
}

// WithDropletStates filters droplets by lifecycle state.
func WithDropletStates(states ...DropletState) DropletListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return dropletListScalar{scalarOption{"states", joinKind(states)}}
}

// ---- packages ----

// PackageListOption configures GET /v3/packages.
type PackageListOption interface {
	QueryOption
	packageList()
}

type packageListScalar struct{ scalarOption }

func (packageListScalar) packageList() {}

// PackageState is a CF v3 package lifecycle state.
type PackageState string

// Valid package states (CF v3).
const (
	PackageStateAwaitingUpload   PackageState = "AWAITING_UPLOAD"
	PackageStateProcessingUpload PackageState = "PROCESSING_UPLOAD"
	PackageStateCopying          PackageState = "COPYING"
	PackageStateReady            PackageState = "READY"
	PackageStateFailed           PackageState = "FAILED"
	PackageStateExpired          PackageState = "EXPIRED"
)

// PackageType is a CF v3 package type. CF uses lowercase values here.
type PackageType string

// Valid package types (CF v3).
const (
	PackageTypeBits   PackageType = "bits"
	PackageTypeDocker PackageType = "docker"
)

// WithPackageGUIDs filters packages by GUID.
func WithPackageGUIDs(guids ...string) PackageListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return packageListScalar{scalarOption{"guids", strings.Join(guids, ",")}}
}

// WithPackageAppGUIDs filters packages by app GUID.
func WithPackageAppGUIDs(guids ...string) PackageListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return packageListScalar{scalarOption{"app_guids", strings.Join(guids, ",")}}
}

// WithPackageSpaceGUIDs filters packages by space GUID.
func WithPackageSpaceGUIDs(guids ...string) PackageListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return packageListScalar{scalarOption{"space_guids", strings.Join(guids, ",")}}
}

// WithPackageOrganizationGUIDs filters packages by organization GUID.
func WithPackageOrganizationGUIDs(guids ...string) PackageListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return packageListScalar{scalarOption{"organization_guids", strings.Join(guids, ",")}}
}

// WithPackageStates filters packages by lifecycle state.
func WithPackageStates(states ...PackageState) PackageListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return packageListScalar{scalarOption{"states", joinKind(states)}}
}

// WithPackageTypes filters packages by type (bits or docker).
func WithPackageTypes(types ...PackageType) PackageListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return packageListScalar{scalarOption{"types", joinKind(types)}}
}

// ---- tasks ----

// TaskListOption configures GET /v3/tasks.
type TaskListOption interface {
	QueryOption
	taskList()
}

type taskListScalar struct{ scalarOption }

func (taskListScalar) taskList() {}

// TaskState is a CF v3 task state. Note CF has no CANCELED state, only
// CANCELING.
type TaskState string

// Valid task states (CF v3).
const (
	TaskStatePending   TaskState = "PENDING"
	TaskStateRunning   TaskState = "RUNNING"
	TaskStateCanceling TaskState = "CANCELING"
	TaskStateSucceeded TaskState = "SUCCEEDED"
	TaskStateFailed    TaskState = "FAILED"
)

// WithTaskGUIDs filters tasks by GUID.
func WithTaskGUIDs(guids ...string) TaskListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return taskListScalar{scalarOption{"guids", strings.Join(guids, ",")}}
}

// WithTaskAppGUIDs filters tasks by app GUID.
func WithTaskAppGUIDs(guids ...string) TaskListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return taskListScalar{scalarOption{"app_guids", strings.Join(guids, ",")}}
}

// WithTaskSpaceGUIDs filters tasks by space GUID.
func WithTaskSpaceGUIDs(guids ...string) TaskListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return taskListScalar{scalarOption{"space_guids", strings.Join(guids, ",")}}
}

// WithTaskOrganizationGUIDs filters tasks by organization GUID.
func WithTaskOrganizationGUIDs(guids ...string) TaskListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return taskListScalar{scalarOption{"organization_guids", strings.Join(guids, ",")}}
}

// WithTaskNames filters tasks by name.
func WithTaskNames(names ...string) TaskListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return taskListScalar{scalarOption{"names", strings.Join(names, ",")}}
}

// WithTaskStates filters tasks by state.
func WithTaskStates(states ...TaskState) TaskListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return taskListScalar{scalarOption{"states", joinKind(states)}}
}

// ---- deployments ----

// DeploymentListOption configures GET /v3/deployments.
type DeploymentListOption interface {
	QueryOption
	deploymentList()
}

type deploymentListScalar struct{ scalarOption }

func (deploymentListScalar) deploymentList() {}

// DeploymentState is a CF v3 deployment state.
type DeploymentState string

// Valid deployment states (CF v3).
const (
	DeploymentStateDeploying DeploymentState = "DEPLOYING"
	DeploymentStatePrepaused DeploymentState = "PREPAUSED"
	DeploymentStatePaused    DeploymentState = "PAUSED"
	DeploymentStateDeployed  DeploymentState = "DEPLOYED"
	DeploymentStateCanceling DeploymentState = "CANCELING"
	DeploymentStateCanceled  DeploymentState = "CANCELED"
)

// DeploymentStatusValue is a CF v3 deployment status.value.
type DeploymentStatusValue string

// Valid deployment status values (CF v3).
const (
	DeploymentStatusValueActive    DeploymentStatusValue = "ACTIVE"
	DeploymentStatusValueFinalized DeploymentStatusValue = "FINALIZED"
)

// DeploymentStatusReason is a CF v3 deployment status.reason.
type DeploymentStatusReason string

// Valid deployment status reasons (CF v3).
const (
	DeploymentStatusReasonDeploying  DeploymentStatusReason = "DEPLOYING"
	DeploymentStatusReasonPaused     DeploymentStatusReason = "PAUSED"
	DeploymentStatusReasonDeployed   DeploymentStatusReason = "DEPLOYED"
	DeploymentStatusReasonCanceled   DeploymentStatusReason = "CANCELED"
	DeploymentStatusReasonCanceling  DeploymentStatusReason = "CANCELING"
	DeploymentStatusReasonSuperseded DeploymentStatusReason = "SUPERSEDED"
)

// WithDeploymentAppGUIDs filters deployments by app GUID.
func WithDeploymentAppGUIDs(guids ...string) DeploymentListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return deploymentListScalar{scalarOption{"app_guids", strings.Join(guids, ",")}}
}

// WithDeploymentStates filters deployments by state.
func WithDeploymentStates(states ...DeploymentState) DeploymentListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return deploymentListScalar{scalarOption{"states", joinKind(states)}}
}

// WithDeploymentStatusValues filters deployments by status value.
func WithDeploymentStatusValues(values ...DeploymentStatusValue) DeploymentListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return deploymentListScalar{scalarOption{"status_values", joinKind(values)}}
}

// WithDeploymentStatusReasons filters deployments by status reason.
func WithDeploymentStatusReasons(reasons ...DeploymentStatusReason) DeploymentListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return deploymentListScalar{scalarOption{"status_reasons", joinKind(reasons)}}
}

// ---- organizations ----

// OrganizationListOption configures GET /v3/organizations.
type OrganizationListOption interface {
	QueryOption
	organizationList()
}

type organizationListScalar struct{ scalarOption }

func (organizationListScalar) organizationList() {}

// WithOrganizationNames filters organizations by name.
func WithOrganizationNames(names ...string) OrganizationListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return organizationListScalar{scalarOption{"names", strings.Join(names, ",")}}
}

// WithOrganizationGUIDs filters organizations by GUID.
func WithOrganizationGUIDs(guids ...string) OrganizationListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return organizationListScalar{scalarOption{"guids", strings.Join(guids, ",")}}
}

// ---- domains ----

// DomainListOption configures GET /v3/domains.
type DomainListOption interface {
	QueryOption
	domainList()
}

type domainListScalar struct{ scalarOption }

func (domainListScalar) domainList() {}

// WithDomainNames filters domains by name.
func WithDomainNames(names ...string) DomainListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return domainListScalar{scalarOption{"names", strings.Join(names, ",")}}
}

// WithDomainGUIDs filters domains by GUID.
func WithDomainGUIDs(guids ...string) DomainListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return domainListScalar{scalarOption{"guids", strings.Join(guids, ",")}}
}

// WithDomainOrganizationGUIDs filters domains by owning organization GUID.
func WithDomainOrganizationGUIDs(guids ...string) DomainListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return domainListScalar{scalarOption{"organization_guids", strings.Join(guids, ",")}}
}

// ---- organization quotas ----

// OrganizationQuotaListOption configures GET /v3/organization_quotas.
type OrganizationQuotaListOption interface {
	QueryOption
	organizationQuotaList()
}

type organizationQuotaListScalar struct{ scalarOption }

func (organizationQuotaListScalar) organizationQuotaList() {}

// WithOrganizationQuotaNames filters organization quotas by name.
func WithOrganizationQuotaNames(names ...string) OrganizationQuotaListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return organizationQuotaListScalar{scalarOption{"names", strings.Join(names, ",")}}
}

// WithOrganizationQuotaGUIDs filters organization quotas by GUID.
func WithOrganizationQuotaGUIDs(guids ...string) OrganizationQuotaListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return organizationQuotaListScalar{scalarOption{"guids", strings.Join(guids, ",")}}
}

// WithOrganizationQuotaOrganizationGUIDs filters organization quotas by
// associated organization GUID.
func WithOrganizationQuotaOrganizationGUIDs(guids ...string) OrganizationQuotaListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return organizationQuotaListScalar{scalarOption{"organization_guids", strings.Join(guids, ",")}}
}

// ---- space quotas ----

// SpaceQuotaListOption configures GET /v3/space_quotas.
type SpaceQuotaListOption interface {
	QueryOption
	spaceQuotaList()
}

type spaceQuotaListScalar struct{ scalarOption }

func (spaceQuotaListScalar) spaceQuotaList() {}

// WithSpaceQuotaNames filters space quotas by name.
func WithSpaceQuotaNames(names ...string) SpaceQuotaListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return spaceQuotaListScalar{scalarOption{"names", strings.Join(names, ",")}}
}

// WithSpaceQuotaGUIDs filters space quotas by GUID.
func WithSpaceQuotaGUIDs(guids ...string) SpaceQuotaListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return spaceQuotaListScalar{scalarOption{"guids", strings.Join(guids, ",")}}
}

// WithSpaceQuotaOrganizationGUIDs filters space quotas by owning
// organization GUID.
func WithSpaceQuotaOrganizationGUIDs(guids ...string) SpaceQuotaListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return spaceQuotaListScalar{scalarOption{"organization_guids", strings.Join(guids, ",")}}
}

// WithSpaceQuotaSpaceGUIDs filters space quotas by associated space GUID.
func WithSpaceQuotaSpaceGUIDs(guids ...string) SpaceQuotaListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return spaceQuotaListScalar{scalarOption{"space_guids", strings.Join(guids, ",")}}
}

// ---- security groups ----

// SecurityGroupListOption configures GET /v3/security_groups.
type SecurityGroupListOption interface {
	QueryOption
	securityGroupList()
}

type securityGroupListScalar struct{ scalarOption }

func (securityGroupListScalar) securityGroupList() {}

// WithSecurityGroupGUIDs filters security groups by GUID.
func WithSecurityGroupGUIDs(guids ...string) SecurityGroupListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return securityGroupListScalar{scalarOption{"guids", strings.Join(guids, ",")}}
}

// WithSecurityGroupNames filters security groups by name.
func WithSecurityGroupNames(names ...string) SecurityGroupListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return securityGroupListScalar{scalarOption{"names", strings.Join(names, ",")}}
}

// WithSecurityGroupRunningSpaceGUIDs filters security groups by the spaces
// where they are bound to the running lifecycle.
func WithSecurityGroupRunningSpaceGUIDs(guids ...string) SecurityGroupListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return securityGroupListScalar{scalarOption{"running_space_guids", strings.Join(guids, ",")}}
}

// WithSecurityGroupStagingSpaceGUIDs filters security groups by the spaces
// where they are bound to the staging lifecycle.
func WithSecurityGroupStagingSpaceGUIDs(guids ...string) SecurityGroupListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return securityGroupListScalar{scalarOption{"staging_space_guids", strings.Join(guids, ",")}}
}

// WithSecurityGroupGloballyEnabledRunning filters security groups by whether
// they apply globally to the running lifecycle.
func WithSecurityGroupGloballyEnabledRunning(enabled bool) SecurityGroupListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return securityGroupListScalar{scalarOption{"globally_enabled_running", strconv.FormatBool(enabled)}}
}

// WithSecurityGroupGloballyEnabledStaging filters security groups by whether
// they apply globally to the staging lifecycle.
func WithSecurityGroupGloballyEnabledStaging(enabled bool) SecurityGroupListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return securityGroupListScalar{scalarOption{"globally_enabled_staging", strconv.FormatBool(enabled)}}
}

// ---- isolation segments ----

// IsolationSegmentListOption configures GET /v3/isolation_segments.
type IsolationSegmentListOption interface {
	QueryOption
	isolationSegmentList()
}

type isolationSegmentListScalar struct{ scalarOption }

func (isolationSegmentListScalar) isolationSegmentList() {}

// WithIsolationSegmentGUIDs filters isolation segments by GUID.
func WithIsolationSegmentGUIDs(guids ...string) IsolationSegmentListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return isolationSegmentListScalar{scalarOption{"guids", strings.Join(guids, ",")}}
}

// WithIsolationSegmentNames filters isolation segments by name.
func WithIsolationSegmentNames(names ...string) IsolationSegmentListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return isolationSegmentListScalar{scalarOption{"names", strings.Join(names, ",")}}
}

// WithIsolationSegmentOrganizationGUIDs filters isolation segments by the
// organizations entitled to them.
func WithIsolationSegmentOrganizationGUIDs(guids ...string) IsolationSegmentListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return isolationSegmentListScalar{scalarOption{"organization_guids", strings.Join(guids, ",")}}
}

// ---- service brokers ----

// ServiceBrokerListOption configures GET /v3/service_brokers.
type ServiceBrokerListOption interface {
	QueryOption
	serviceBrokerList()
}

type serviceBrokerListScalar struct{ scalarOption }

func (serviceBrokerListScalar) serviceBrokerList() {}

// WithServiceBrokerGUIDs filters service brokers by GUID.
func WithServiceBrokerGUIDs(guids ...string) ServiceBrokerListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return serviceBrokerListScalar{scalarOption{"guids", strings.Join(guids, ",")}}
}

// WithServiceBrokerNames filters service brokers by name.
func WithServiceBrokerNames(names ...string) ServiceBrokerListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return serviceBrokerListScalar{scalarOption{"names", strings.Join(names, ",")}}
}

// WithServiceBrokerSpaceGUIDs filters service brokers by the space they are
// scoped to (space-scoped brokers).
func WithServiceBrokerSpaceGUIDs(guids ...string) ServiceBrokerListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return serviceBrokerListScalar{scalarOption{"space_guids", strings.Join(guids, ",")}}
}

// ---- buildpacks ----

// BuildpackListOption configures GET /v3/buildpacks.
type BuildpackListOption interface {
	QueryOption
	buildpackList()
}

type buildpackListScalar struct{ scalarOption }

func (buildpackListScalar) buildpackList() {}

// BuildpackLifecycle is a CF v3 buildpack lifecycle.
type BuildpackLifecycle string

// Valid buildpack lifecycles (CF v3).
const (
	BuildpackLifecycleBuildpack BuildpackLifecycle = "buildpack"
	BuildpackLifecycleCNB       BuildpackLifecycle = "cnb"
)

// WithBuildpackGUIDs filters buildpacks by GUID.
func WithBuildpackGUIDs(guids ...string) BuildpackListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return buildpackListScalar{scalarOption{"guids", strings.Join(guids, ",")}}
}

// WithBuildpackNames filters buildpacks by name.
func WithBuildpackNames(names ...string) BuildpackListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return buildpackListScalar{scalarOption{"names", strings.Join(names, ",")}}
}

// WithBuildpackStacks filters buildpacks by stack. An empty string matches
// buildpacks with no stack.
func WithBuildpackStacks(stacks ...string) BuildpackListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return buildpackListScalar{scalarOption{"stacks", strings.Join(stacks, ",")}}
}

// WithBuildpackLifecycle filters buildpacks by lifecycle (buildpack or cnb).
func WithBuildpackLifecycle(lifecycle BuildpackLifecycle) BuildpackListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return buildpackListScalar{scalarOption{"lifecycle", string(lifecycle)}}
}

// ---- stacks ----

// StackListOption configures GET /v3/stacks.
type StackListOption interface {
	QueryOption
	stackList()
}

type stackListScalar struct{ scalarOption }

func (stackListScalar) stackList() {}

// WithStackGUIDs filters stacks by GUID.
func WithStackGUIDs(guids ...string) StackListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return stackListScalar{scalarOption{"guids", strings.Join(guids, ",")}}
}

// WithStackNames filters stacks by name.
func WithStackNames(names ...string) StackListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return stackListScalar{scalarOption{"names", strings.Join(names, ",")}}
}

// WithStackDefault filters stacks by whether they are the default stack.
func WithStackDefault(isDefault bool) StackListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return stackListScalar{scalarOption{"default", strconv.FormatBool(isDefault)}}
}

// ---- users ----

// UserListOption configures GET /v3/users.
type UserListOption interface {
	QueryOption
	userList()
}

type userListScalar struct{ scalarOption }

func (userListScalar) userList() {}

// WithUserGUIDs filters users by GUID.
func WithUserGUIDs(guids ...string) UserListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return userListScalar{scalarOption{"guids", strings.Join(guids, ",")}}
}

// WithUserUsernames filters users by exact username. Mutually exclusive with
// WithUserPartialUsernames per CF.
func WithUserUsernames(usernames ...string) UserListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return userListScalar{scalarOption{"usernames", strings.Join(usernames, ",")}}
}

// WithUserPartialUsernames filters users by partial (substring) username.
// Mutually exclusive with WithUserUsernames per CF.
func WithUserPartialUsernames(partials ...string) UserListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return userListScalar{scalarOption{"partial_usernames", strings.Join(partials, ",")}}
}

// WithUserOrigins filters users by identity-provider origin. CF requires this
// alongside WithUserUsernames or WithUserPartialUsernames.
func WithUserOrigins(origins ...string) UserListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return userListScalar{scalarOption{"origins", strings.Join(origins, ",")}}
}

// ---- audit events ----

// AuditEventListOption configures GET /v3/audit_events.
type AuditEventListOption interface {
	QueryOption
	auditEventList()
}

type auditEventListScalar struct{ scalarOption }

func (auditEventListScalar) auditEventList() {}

// WithAuditEventTypes filters audit events by event type (e.g.
// "audit.app.create").
func WithAuditEventTypes(types ...string) AuditEventListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return auditEventListScalar{scalarOption{"types", strings.Join(types, ",")}}
}

// WithAuditEventTargetGUIDs filters audit events by target GUID.
func WithAuditEventTargetGUIDs(guids ...string) AuditEventListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return auditEventListScalar{scalarOption{"target_guids", strings.Join(guids, ",")}}
}

// WithAuditEventSpaceGUIDs filters audit events by space GUID.
func WithAuditEventSpaceGUIDs(guids ...string) AuditEventListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return auditEventListScalar{scalarOption{"space_guids", strings.Join(guids, ",")}}
}

// WithAuditEventOrganizationGUIDs filters audit events by organization GUID.
func WithAuditEventOrganizationGUIDs(guids ...string) AuditEventListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return auditEventListScalar{scalarOption{"organization_guids", strings.Join(guids, ",")}}
}

// ---- app usage events ----

// AppUsageEventListOption configures GET /v3/app_usage_events.
type AppUsageEventListOption interface {
	QueryOption
	appUsageEventList()
}

type appUsageEventListScalar struct{ scalarOption }

func (appUsageEventListScalar) appUsageEventList() {}

// WithAppUsageEventGUIDs filters app usage events by GUID.
func WithAppUsageEventGUIDs(guids ...string) AppUsageEventListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return appUsageEventListScalar{scalarOption{"guids", strings.Join(guids, ",")}}
}

// WithAppUsageEventAfterGUID returns only events recorded after the event
// with the given GUID. CF accepts a single value here.
func WithAppUsageEventAfterGUID(guid string) AppUsageEventListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return appUsageEventListScalar{scalarOption{"after_guid", guid}}
}

// ---- service usage events ----

// ServiceUsageEventListOption configures GET /v3/service_usage_events.
type ServiceUsageEventListOption interface {
	QueryOption
	serviceUsageEventList()
}

type serviceUsageEventListScalar struct{ scalarOption }

func (serviceUsageEventListScalar) serviceUsageEventList() {}

// ServiceInstanceType is a CF v3 service instance type used to filter service
// usage events.
type ServiceInstanceType string

// Valid service instance types (CF v3).
const (
	ServiceInstanceTypeManaged      ServiceInstanceType = "managed_service_instance"
	ServiceInstanceTypeUserProvided ServiceInstanceType = "user_provided_service_instance"
)

// WithServiceUsageEventGUIDs filters service usage events by GUID.
func WithServiceUsageEventGUIDs(guids ...string) ServiceUsageEventListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return serviceUsageEventListScalar{scalarOption{"guids", strings.Join(guids, ",")}}
}

// WithServiceUsageEventAfterGUID returns only events recorded after the event
// with the given GUID. CF accepts a single value here.
func WithServiceUsageEventAfterGUID(guid string) ServiceUsageEventListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return serviceUsageEventListScalar{scalarOption{"after_guid", guid}}
}

// WithServiceUsageEventServiceInstanceTypes filters service usage events by
// service instance type.
func WithServiceUsageEventServiceInstanceTypes(types ...ServiceInstanceType) ServiceUsageEventListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return serviceUsageEventListScalar{scalarOption{"service_instance_types", joinKind(types)}}
}

// WithServiceUsageEventServiceOfferingGUIDs filters service usage events by
// service offering GUID.
func WithServiceUsageEventServiceOfferingGUIDs(guids ...string) ServiceUsageEventListOption { //nolint:ireturn // sealed-option pattern: typed option composed by callers
	return serviceUsageEventListScalar{scalarOption{"service_offering_guids", strings.Join(guids, ",")}}
}
