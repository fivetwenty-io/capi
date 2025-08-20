package capi

// App represents a Cloud Foundry application
type App struct {
	Resource
	Name                 string                 `json:"name"`
	State                string                 `json:"state"`
	Lifecycle            Lifecycle              `json:"lifecycle"`
	Metadata             *Metadata              `json:"metadata,omitempty"`
	Relationships        AppRelationships       `json:"relationships"`
	EnvironmentVariables map[string]interface{} `json:"environment_variables,omitempty"`
}

// AppCreateRequest represents a request to create an app
type AppCreateRequest struct {
	Name                 string                 `json:"name"`
	Relationships        AppRelationships       `json:"relationships"`
	Lifecycle            *Lifecycle             `json:"lifecycle,omitempty"`
	EnvironmentVariables map[string]interface{} `json:"environment_variables,omitempty"`
	Metadata             *Metadata              `json:"metadata,omitempty"`
}

// AppUpdateRequest represents a request to update an app
type AppUpdateRequest struct {
	Name      *string    `json:"name,omitempty"`
	Lifecycle *Lifecycle `json:"lifecycle,omitempty"`
	Metadata  *Metadata  `json:"metadata,omitempty"`
}

// AppRelationships represents app relationships
type AppRelationships struct {
	Space Relationship `json:"space"`
}

// Lifecycle represents app lifecycle configuration
type Lifecycle struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

// AppEnvironment represents app environment information
type AppEnvironment struct {
	StagingEnvJSON       map[string]interface{} `json:"staging_env_json"`
	RunningEnvJSON       map[string]interface{} `json:"running_env_json"`
	EnvironmentVariables map[string]interface{} `json:"environment_variables"`
	SystemEnvJSON        map[string]interface{} `json:"system_env_json"`
	ApplicationEnvJSON   map[string]interface{} `json:"application_env_json"`
}

// AppSSHEnabled represents SSH enablement status
type AppSSHEnabled struct {
	Enabled bool   `json:"enabled"`
	Reason  string `json:"reason,omitempty"`
}

// AppPermissions represents app permissions
type AppPermissions struct {
	ReadBasicData     bool `json:"read_basic_data"`
	ReadSensitiveData bool `json:"read_sensitive_data"`
}

// Organization represents a Cloud Foundry organization
type Organization struct {
	Resource
	Name          string            `json:"name"`
	Suspended     bool              `json:"suspended"`
	Metadata      *Metadata         `json:"metadata,omitempty"`
	Relationships *OrgRelationships `json:"relationships,omitempty"`
}

// OrganizationCreateRequest represents a request to create an organization
type OrganizationCreateRequest struct {
	Name     string    `json:"name"`
	Metadata *Metadata `json:"metadata,omitempty"`
}

// OrganizationUpdateRequest represents a request to update an organization
type OrganizationUpdateRequest struct {
	Name      *string   `json:"name,omitempty"`
	Suspended *bool     `json:"suspended,omitempty"`
	Metadata  *Metadata `json:"metadata,omitempty"`
}

// OrgRelationships represents organization relationships
type OrgRelationships struct {
	Quota Relationship `json:"quota,omitempty"`
}

// OrganizationUsageSummary represents organization usage summary
type OrganizationUsageSummary struct {
	UsageSummary struct {
		StartedInstances int `json:"started_instances"`
		MemoryInMB       int `json:"memory_in_mb"`
	} `json:"usage_summary"`
}

// Space represents a Cloud Foundry space
type Space struct {
	Resource
	Name          string             `json:"name"`
	Metadata      *Metadata          `json:"metadata,omitempty"`
	Relationships SpaceRelationships `json:"relationships"`
}

// SpaceCreateRequest represents a request to create a space
type SpaceCreateRequest struct {
	Name          string             `json:"name"`
	Relationships SpaceRelationships `json:"relationships"`
	Metadata      *Metadata          `json:"metadata,omitempty"`
}

// SpaceUpdateRequest represents a request to update a space
type SpaceUpdateRequest struct {
	Name     *string   `json:"name,omitempty"`
	Metadata *Metadata `json:"metadata,omitempty"`
}

// SpaceRelationships represents space relationships
type SpaceRelationships struct {
	Organization Relationship  `json:"organization"`
	Quota        *Relationship `json:"quota,omitempty"`
}

// SpaceFeatures represents space features
type SpaceFeatures struct {
	SSHEnabled bool `json:"ssh_enabled"`
}

// SpaceFeature represents a single space feature
type SpaceFeature struct {
	Name        string `json:"name"`
	Enabled     bool   `json:"enabled"`
	Description string `json:"description,omitempty"`
}

// SpaceUsageSummary represents space usage summary
type SpaceUsageSummary struct {
	UsageSummary struct {
		StartedInstances int `json:"started_instances"`
		MemoryInMB       int `json:"memory_in_mb"`
	} `json:"usage_summary"`
}

// Domain represents a domain
type Domain struct {
	Resource
	Name               string              `json:"name"`
	Internal           bool                `json:"internal"`
	RouterGroup        *string             `json:"router_group,omitempty"`
	SupportedProtocols []string            `json:"supported_protocols"`
	Metadata           *Metadata           `json:"metadata,omitempty"`
	Relationships      DomainRelationships `json:"relationships"`
}

// DomainCreateRequest represents a request to create a domain
type DomainCreateRequest struct {
	Name          string               `json:"name"`
	Internal      *bool                `json:"internal,omitempty"`
	RouterGroup   *string              `json:"router_group,omitempty"`
	Relationships *DomainRelationships `json:"relationships,omitempty"`
	Metadata      *Metadata            `json:"metadata,omitempty"`
}

// DomainUpdateRequest represents a request to update a domain
type DomainUpdateRequest struct {
	Metadata *Metadata `json:"metadata,omitempty"`
}

// DomainRelationships represents domain relationships
type DomainRelationships struct {
	Organization        *Relationship       `json:"organization,omitempty"`
	SharedOrganizations *ToManyRelationship `json:"shared_organizations,omitempty"`
}

// Route represents a route
type Route struct {
	Resource
	Protocol      string             `json:"protocol"`
	Host          string             `json:"host"`
	Path          string             `json:"path"`
	Port          *int               `json:"port,omitempty"`
	URL           string             `json:"url"`
	Destinations  []RouteDestination `json:"destinations"`
	Metadata      *Metadata          `json:"metadata,omitempty"`
	Relationships RouteRelationships `json:"relationships"`
}

// RouteCreateRequest represents a request to create a route
type RouteCreateRequest struct {
	Host          *string            `json:"host,omitempty"`
	Path          *string            `json:"path,omitempty"`
	Port          *int               `json:"port,omitempty"`
	Relationships RouteRelationships `json:"relationships"`
	Metadata      *Metadata          `json:"metadata,omitempty"`
}

// RouteUpdateRequest represents a request to update a route
type RouteUpdateRequest struct {
	Metadata *Metadata `json:"metadata,omitempty"`
}

// RouteRelationships represents route relationships
type RouteRelationships struct {
	Space  Relationship `json:"space"`
	Domain Relationship `json:"domain"`
}

// RouteDestination represents a route destination
type RouteDestination struct {
	GUID     string              `json:"guid"`
	App      RouteDestinationApp `json:"app"`
	Port     *int                `json:"port,omitempty"`
	Protocol *string             `json:"protocol,omitempty"`
	Weight   *int                `json:"weight,omitempty"`
}

// RouteDestinationApp represents the app in a route destination
type RouteDestinationApp struct {
	GUID    string   `json:"guid"`
	Process *Process `json:"process,omitempty"`
}

// RouteDestinations represents a list of route destinations
type RouteDestinations struct {
	Destinations []RouteDestination `json:"destinations"`
	Links        Links              `json:"links"`
}

// RouteReservation represents a route reservation check
type RouteReservation struct {
	MatchingRoute *Route `json:"matching_route"`
}

// RouteReservationRequest represents a request to check route reservation
type RouteReservationRequest struct {
	Host string `json:"host,omitempty"`
	Path string `json:"path,omitempty"`
	Port *int   `json:"port,omitempty"`
}

// ManifestDiff represents a manifest diff
type ManifestDiff struct {
	Diff string `json:"diff"`
}

// User represents a user
type User struct {
	Resource
	Username         string    `json:"username"`
	PresentationName string    `json:"presentation_name"`
	Origin           string    `json:"origin"`
	Metadata         *Metadata `json:"metadata,omitempty"`
}

// UserCreateRequest represents a request to create a user
type UserCreateRequest struct {
	GUID     string    `json:"guid"`
	Metadata *Metadata `json:"metadata,omitempty"`
}

// UserUpdateRequest represents a request to update a user
type UserUpdateRequest struct {
	Metadata *Metadata `json:"metadata,omitempty"`
}

// Role represents a role
type Role struct {
	Resource
	Type          string            `json:"type"`
	Relationships RoleRelationships `json:"relationships"`
}

// RoleCreateRequest represents a request to create a role
type RoleCreateRequest struct {
	Type          string            `json:"type"`
	Relationships RoleRelationships `json:"relationships"`
}

// RoleRelationships represents role relationships
type RoleRelationships struct {
	User         Relationship  `json:"user"`
	Organization *Relationship `json:"organization,omitempty"`
	Space        *Relationship `json:"space,omitempty"`
}

// Package represents a Cloud Foundry package
type Package struct {
	Resource
	Type          string                `json:"type"`
	Data          *PackageData          `json:"data"`
	State         string                `json:"state"`
	Metadata      *Metadata             `json:"metadata,omitempty"`
	Relationships *PackageRelationships `json:"relationships,omitempty"`
}

// PackageData represents package-specific data
type PackageData struct {
	Checksum *PackageChecksum `json:"checksum,omitempty"`
	Error    *string          `json:"error,omitempty"`
	Image    *string          `json:"image,omitempty"`    // For Docker packages
	Username *string          `json:"username,omitempty"` // For Docker packages
	Password *string          `json:"password,omitempty"` // For Docker packages
}

// PackageChecksum represents package checksum information
type PackageChecksum struct {
	Type  string  `json:"type"` // e.g., "sha256"
	Value *string `json:"value"`
}

// PackageRelationships represents the relationships for a package
type PackageRelationships struct {
	App *Relationship `json:"app,omitempty"`
}

// PackageCreateRequest represents a request to create a package
type PackageCreateRequest struct {
	Type          string               `json:"type"`
	Relationships PackageRelationships `json:"relationships"`
	Data          *PackageCreateData   `json:"data,omitempty"`
	Metadata      *Metadata            `json:"metadata,omitempty"`
}

// PackageCreateData represents data for creating a package
type PackageCreateData struct {
	Image    *string `json:"image,omitempty"`    // For Docker packages
	Username *string `json:"username,omitempty"` // For Docker packages
	Password *string `json:"password,omitempty"` // For Docker packages
}

// PackageUpdateRequest represents a request to update a package
type PackageUpdateRequest struct {
	Metadata *Metadata `json:"metadata,omitempty"`
}

// PackageUploadRequest represents a request to upload package bits
type PackageUploadRequest struct {
	Bits      []byte            `json:"-"` // The actual file bits
	Resources []PackageResource `json:"resources,omitempty"`
}

// PackageResource represents a resource in a package upload
type PackageResource struct {
	SHA1 string `json:"sha1"`
	Size int64  `json:"size"`
	Path string `json:"path"`
	Mode string `json:"mode"`
}

// PackageCopyRequest represents a request to copy a package
type PackageCopyRequest struct {
	Relationships PackageRelationships `json:"relationships"`
}

// Droplet represents a Cloud Foundry droplet
type Droplet struct {
	Resource
	State             string                `json:"state"`
	Error             *string               `json:"error"`
	Lifecycle         Lifecycle             `json:"lifecycle"`
	ExecutionMetadata string                `json:"execution_metadata"`
	ProcessTypes      map[string]string     `json:"process_types"`
	Checksum          *DropletChecksum      `json:"checksum,omitempty"`
	Buildpacks        []DetectedBuildpack   `json:"buildpacks,omitempty"`
	Stack             *string               `json:"stack,omitempty"`
	Image             *string               `json:"image,omitempty"`
	Metadata          *Metadata             `json:"metadata,omitempty"`
	Relationships     *DropletRelationships `json:"relationships,omitempty"`
}

// DropletChecksum represents droplet checksum information
type DropletChecksum struct {
	Type  string `json:"type"` // e.g., "sha256" or "sha1"
	Value string `json:"value"`
}

// DetectedBuildpack represents a buildpack detected during staging
type DetectedBuildpack struct {
	Name          string  `json:"name"`
	DetectOutput  string  `json:"detect_output"`
	Version       *string `json:"version,omitempty"`
	BuildpackName *string `json:"buildpack_name,omitempty"`
}

// DropletRelationships represents the relationships for a droplet
type DropletRelationships struct {
	App *Relationship `json:"app,omitempty"`
}

// DropletCreateRequest represents a request to create a droplet
type DropletCreateRequest struct {
	Relationships DropletRelationships `json:"relationships"`
	ProcessTypes  map[string]string    `json:"process_types,omitempty"`
}

// DropletUpdateRequest represents a request to update a droplet
type DropletUpdateRequest struct {
	Metadata     *Metadata         `json:"metadata,omitempty"`
	Image        *string           `json:"image,omitempty"`
	ProcessTypes map[string]string `json:"process_types,omitempty"`
}

// DropletCopyRequest represents a request to copy a droplet
type DropletCopyRequest struct {
	Relationships DropletRelationships `json:"relationships"`
}

// Build represents a Cloud Foundry build
type Build struct{ Resource }
type BuildCreateRequest struct{}
type BuildUpdateRequest struct{}
type Buildpack struct{ Resource }
type BuildpackCreateRequest struct{}
type BuildpackUpdateRequest struct{}
type Deployment struct{ Resource }
type DeploymentCreateRequest struct{}
type DeploymentUpdateRequest struct{}

// Process represents a Cloud Foundry process
type Process struct {
	Resource
	Type                         string                `json:"type"`
	Command                      *string               `json:"command"`
	User                         string                `json:"user,omitempty"`
	Instances                    int                   `json:"instances"`
	MemoryInMB                   int                   `json:"memory_in_mb"`
	DiskInMB                     int                   `json:"disk_in_mb"`
	LogRateLimitInBytesPerSecond *int                  `json:"log_rate_limit_in_bytes_per_second"`
	HealthCheck                  *HealthCheck          `json:"health_check"`
	ReadinessHealthCheck         *ReadinessHealthCheck `json:"readiness_health_check"`
	Version                      string                `json:"version,omitempty"`
	Metadata                     *Metadata             `json:"metadata,omitempty"`
	Relationships                *ProcessRelationships `json:"relationships,omitempty"`
}

// ProcessRelationships represents the relationships for a process
type ProcessRelationships struct {
	App      *Relationship `json:"app,omitempty"`
	Revision *Relationship `json:"revision,omitempty"`
}

// HealthCheck represents a process health check
type HealthCheck struct {
	Type string           `json:"type"` // "port", "process", or "http"
	Data *HealthCheckData `json:"data,omitempty"`
}

// HealthCheckData represents health check configuration data
type HealthCheckData struct {
	Timeout           *int    `json:"timeout,omitempty"`
	InvocationTimeout *int    `json:"invocation_timeout,omitempty"`
	Interval          *int    `json:"interval,omitempty"`
	Endpoint          *string `json:"endpoint,omitempty"` // For HTTP health checks
}

// ReadinessHealthCheck represents a process readiness health check
type ReadinessHealthCheck struct {
	Type string                    `json:"type"` // "process", "port", or "http"
	Data *ReadinessHealthCheckData `json:"data,omitempty"`
}

// ReadinessHealthCheckData represents readiness health check configuration data
type ReadinessHealthCheckData struct {
	InvocationTimeout *int    `json:"invocation_timeout,omitempty"`
	Interval          *int    `json:"interval,omitempty"`
	Endpoint          *string `json:"endpoint,omitempty"` // For HTTP readiness checks
}

// ProcessUpdateRequest represents a request to update a process
type ProcessUpdateRequest struct {
	Command  *string   `json:"command,omitempty"`
	Metadata *Metadata `json:"metadata,omitempty"`
}

// ProcessScaleRequest represents a request to scale a process
type ProcessScaleRequest struct {
	Instances                    *int `json:"instances,omitempty"`
	MemoryInMB                   *int `json:"memory_in_mb,omitempty"`
	DiskInMB                     *int `json:"disk_in_mb,omitempty"`
	LogRateLimitInBytesPerSecond *int `json:"log_rate_limit_in_bytes_per_second,omitempty"`
}

// ProcessStats represents statistics for a process
type ProcessStats struct {
	Pagination *Pagination          `json:"pagination"`
	Resources  []ProcessStatsDetail `json:"resources"`
}

// ProcessStatsDetail represents detailed statistics for a process instance
type ProcessStatsDetail struct {
	Type             string                `json:"type"`
	Index            int                   `json:"index"`
	State            string                `json:"state"`
	Usage            *ProcessUsage         `json:"usage,omitempty"`
	Host             string                `json:"host,omitempty"`
	InstancePorts    []ProcessInstancePort `json:"instance_ports,omitempty"`
	Uptime           int                   `json:"uptime,omitempty"`
	MemQuota         int64                 `json:"mem_quota,omitempty"`
	DiskQuota        int64                 `json:"disk_quota,omitempty"`
	FdsQuota         int                   `json:"fds_quota,omitempty"`
	IsolationSegment *string               `json:"isolation_segment,omitempty"`
	Details          *string               `json:"details,omitempty"`
}

// ProcessUsage represents CPU and memory usage for a process instance
type ProcessUsage struct {
	Time           string  `json:"time"`
	CPU            float64 `json:"cpu"`
	CPUEntitlement float64 `json:"cpu_entitlement,omitempty"`
	Mem            int64   `json:"mem"`
	Disk           int64   `json:"disk"`
	LogRate        int     `json:"log_rate"`
}

// ProcessInstancePort represents port mappings for a process instance
type ProcessInstancePort struct {
	External             int `json:"external"`
	Internal             int `json:"internal"`
	ExternalTLSProxyPort int `json:"external_tls_proxy_port,omitempty"`
	InternalTLSProxyPort int `json:"internal_tls_proxy_port,omitempty"`
}

// Task represents a Cloud Foundry task
type Task struct {
	Resource
	SequenceID                   int                `json:"sequence_id"`
	Name                         string             `json:"name"`
	Command                      string             `json:"command,omitempty"`
	User                         *string            `json:"user"`
	State                        string             `json:"state"`
	MemoryInMB                   int                `json:"memory_in_mb"`
	DiskInMB                     int                `json:"disk_in_mb"`
	LogRateLimitInBytesPerSecond *int               `json:"log_rate_limit_in_bytes_per_second"`
	Result                       *TaskResult        `json:"result,omitempty"`
	DropletGUID                  string             `json:"droplet_guid"`
	Metadata                     *Metadata          `json:"metadata,omitempty"`
	Relationships                *TaskRelationships `json:"relationships,omitempty"`
}

// TaskResult represents the result of a task execution
type TaskResult struct {
	FailureReason *string `json:"failure_reason"`
}

// TaskRelationships represents the relationships for a task
type TaskRelationships struct {
	App *Relationship `json:"app,omitempty"`
}

// TaskCreateRequest represents a request to create a task
type TaskCreateRequest struct {
	Command                      *string       `json:"command,omitempty"`
	Name                         *string       `json:"name,omitempty"`
	MemoryInMB                   *int          `json:"memory_in_mb,omitempty"`
	DiskInMB                     *int          `json:"disk_in_mb,omitempty"`
	LogRateLimitInBytesPerSecond *int          `json:"log_rate_limit_in_bytes_per_second,omitempty"`
	Template                     *TaskTemplate `json:"template,omitempty"`
	Metadata                     *Metadata     `json:"metadata,omitempty"`
	DropletGUID                  *string       `json:"droplet_guid,omitempty"`
}

// TaskTemplate represents a template for creating a task from a process
type TaskTemplate struct {
	Process *TaskTemplateProcess `json:"process,omitempty"`
}

// TaskTemplateProcess represents a process reference in a task template
type TaskTemplateProcess struct {
	GUID string `json:"guid"`
}

// TaskUpdateRequest represents a request to update a task
type TaskUpdateRequest struct {
	Metadata *Metadata `json:"metadata,omitempty"`
}
type Stack struct{ Resource }
type StackCreateRequest struct{}
type StackUpdateRequest struct{}
type SecurityGroup struct{ Resource }
type SecurityGroupCreateRequest struct{}
type SecurityGroupUpdateRequest struct{}
type IsolationSegment struct{ Resource }
type IsolationSegmentCreateRequest struct{}
type IsolationSegmentUpdateRequest struct{}
type FeatureFlag struct {
	Name    string
	Enabled bool
}
type ServiceBroker struct{ Resource }
type ServiceBrokerCreateRequest struct{}
type ServiceBrokerUpdateRequest struct{}
type ServiceOffering struct{ Resource }
type ServiceOfferingUpdateRequest struct{}
type ServicePlan struct{ Resource }
type ServicePlanUpdateRequest struct{}
type ServicePlanVisibility struct{}
type ServicePlanVisibilityUpdateRequest struct{}
type ServicePlanVisibilityApplyRequest struct{}
type ServiceInstance struct{ Resource }
type ServiceInstanceCreateRequest struct{}
type ServiceInstanceUpdateRequest struct{}
type ServiceInstancePermissions struct{}
type ServiceInstanceUsageSummary struct{}
