package capi

import "time"

// App represents a Cloud Foundry application.
type App struct {
	Resource

	Name                 string                 `json:"name"                            yaml:"name"`
	State                string                 `json:"state"                           yaml:"state"`
	Lifecycle            Lifecycle              `json:"lifecycle"                       yaml:"lifecycle"`
	Metadata             *Metadata              `json:"metadata,omitempty"              yaml:"metadata,omitempty"`
	Relationships        AppRelationships       `json:"relationships"                   yaml:"relationships"`
	EnvironmentVariables map[string]interface{} `json:"environment_variables,omitempty" yaml:"environment_variables,omitempty"`
}

// AppCreateRequest represents a request to create an app.
type AppCreateRequest struct {
	// Name is the app name (unique within a space).
	Name string `json:"name" yaml:"name"`
	// Relationships must include a Space relationship.
	Relationships AppRelationships `json:"relationships" yaml:"relationships"`
	// Lifecycle optionally specifies staging type/config; if nil, the platform default is used.
	Lifecycle *Lifecycle `json:"lifecycle,omitempty" yaml:"lifecycle,omitempty"`
	// EnvironmentVariables sets initial app-level env vars.
	EnvironmentVariables map[string]interface{} `json:"environment_variables,omitempty" yaml:"environment_variables,omitempty"`
	// Metadata sets labels/annotations on the app.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// AppUpdateRequest represents a request to update an app.
type AppUpdateRequest struct {
	// Name updates the app name; nil leaves it unchanged.
	Name *string `json:"name,omitempty" yaml:"name,omitempty"`
	// Lifecycle updates staging config; nil leaves it unchanged.
	Lifecycle *Lifecycle `json:"lifecycle,omitempty" yaml:"lifecycle,omitempty"`
	// Metadata updates labels/annotations; nil leaves it unchanged.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// AppRelationships represents app relationships.
type AppRelationships struct {
	Space Relationship `json:"space" yaml:"space"`
}

// Lifecycle represents app lifecycle configuration.
type Lifecycle struct {
	Type string                 `json:"type" yaml:"type"`
	Data map[string]interface{} `json:"data" yaml:"data"`
}

// AppEnvironment represents app environment information.
type AppEnvironment struct {
	StagingEnvJSON       map[string]interface{} `json:"staging_env_json"      yaml:"staging_env_json"`
	RunningEnvJSON       map[string]interface{} `json:"running_env_json"      yaml:"running_env_json"`
	EnvironmentVariables map[string]interface{} `json:"environment_variables" yaml:"environment_variables"`
	SystemEnvJSON        map[string]interface{} `json:"system_env_json"       yaml:"system_env_json"`
	ApplicationEnvJSON   map[string]interface{} `json:"application_env_json"  yaml:"application_env_json"`
}

// AppSSHEnabled represents SSH enablement status.
type AppSSHEnabled struct {
	Enabled bool   `json:"enabled"          yaml:"enabled"`
	Reason  string `json:"reason,omitempty" yaml:"reason,omitempty"`
}

// AppPermissions represents app permissions.
type AppPermissions struct {
	ReadBasicData     bool `json:"read_basic_data"     yaml:"read_basic_data"`
	ReadSensitiveData bool `json:"read_sensitive_data" yaml:"read_sensitive_data"`
}

// AppFeature represents a single app feature.
type AppFeature struct {
	Name        string `json:"name"        yaml:"name"`
	Description string `json:"description" yaml:"description"`
	Enabled     bool   `json:"enabled"     yaml:"enabled"`
}

// AppFeatures represents a collection of app features.
type AppFeatures struct {
	Resources []AppFeature `json:"resources" yaml:"resources"`
}

// AppFeatureUpdateRequest represents a request to update an app feature.
type AppFeatureUpdateRequest struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
}

// Organization represents a Cloud Foundry organization.
type Organization struct {
	Resource

	Name          string            `json:"name"                    yaml:"name"`
	Suspended     bool              `json:"suspended"               yaml:"suspended"`
	Metadata      *Metadata         `json:"metadata,omitempty"      yaml:"metadata,omitempty"`
	Relationships *OrgRelationships `json:"relationships,omitempty" yaml:"relationships,omitempty"`
}

// OrganizationCreateRequest represents a request to create an organization.
type OrganizationCreateRequest struct {
	Name     string    `json:"name"               yaml:"name"`
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// OrganizationUpdateRequest represents a request to update an organization.
type OrganizationUpdateRequest struct {
	Name      *string   `json:"name,omitempty"      yaml:"name,omitempty"`
	Suspended *bool     `json:"suspended,omitempty" yaml:"suspended,omitempty"`
	Metadata  *Metadata `json:"metadata,omitempty"  yaml:"metadata,omitempty"`
}

// OrgRelationships represents organization relationships.
type OrgRelationships struct {
	Quota Relationship `json:"quota,omitempty" yaml:"quota,omitempty"`
}

// OrganizationUsageSummary represents organization usage summary.
type OrganizationUsageSummary struct {
	UsageSummary struct {
		StartedInstances int `json:"started_instances" yaml:"started_instances"`
		MemoryInMB       int `json:"memory_in_mb"      yaml:"memory_in_mb"`
	} `json:"usage_summary"`
}

// Space represents a Cloud Foundry space.
type Space struct {
	Resource

	Name          string             `json:"name"               yaml:"name"`
	Metadata      *Metadata          `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Relationships SpaceRelationships `json:"relationships"      yaml:"relationships"`
}

// SpaceCreateRequest represents a request to create a space.
type SpaceCreateRequest struct {
	// Name is the space name (unique within an organization).
	Name string `json:"name" yaml:"name"`
	// Relationships must include an Organization relationship.
	Relationships SpaceRelationships `json:"relationships" yaml:"relationships"`
	// Metadata sets labels/annotations on the space.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// SpaceUpdateRequest represents a request to update a space.
type SpaceUpdateRequest struct {
	// Name updates the space name; nil leaves it unchanged.
	Name *string `json:"name,omitempty" yaml:"name,omitempty"`
	// Metadata updates labels/annotations; nil leaves it unchanged.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// SpaceRelationships represents space relationships.
type SpaceRelationships struct {
	Organization Relationship  `json:"organization"    yaml:"organization"`
	Quota        *Relationship `json:"quota,omitempty" yaml:"quota,omitempty"`
}

// SpaceFeatures represents space features.
type SpaceFeatures struct {
	SSHEnabled bool `json:"ssh_enabled" yaml:"ssh_enabled"`
}

// SpaceFeature represents a single space feature.
type SpaceFeature struct {
	Name        string `json:"name"                  yaml:"name"`
	Enabled     bool   `json:"enabled"               yaml:"enabled"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// SpaceUsageSummary represents space usage summary.
type SpaceUsageSummary struct {
	UsageSummary struct {
		StartedInstances int `json:"started_instances" yaml:"started_instances"`
		MemoryInMB       int `json:"memory_in_mb"      yaml:"memory_in_mb"`
	} `json:"usage_summary"`
}

// Domain represents a domain.
type Domain struct {
	Resource

	Name               string              `json:"name"                   yaml:"name"`
	Internal           bool                `json:"internal"               yaml:"internal"`
	RouterGroup        *string             `json:"router_group,omitempty" yaml:"router_group,omitempty"`
	SupportedProtocols []string            `json:"supported_protocols"    yaml:"supported_protocols"`
	Metadata           *Metadata           `json:"metadata,omitempty"     yaml:"metadata,omitempty"`
	Relationships      DomainRelationships `json:"relationships"          yaml:"relationships"`
}

// DomainCreateRequest represents a request to create a domain.
type DomainCreateRequest struct {
	// Name is the domain name (e.g., example.com).
	Name string `json:"name" yaml:"name"`
	// Internal marks a private domain for internal routing.
	Internal *bool `json:"internal,omitempty" yaml:"internal,omitempty"`
	// RouterGroup associates a TCP router group when creating TCP domains.
	RouterGroup *string `json:"router_group,omitempty" yaml:"router_group,omitempty"`
	// Relationships optionally set the owning organization or shared orgs.
	Relationships *DomainRelationships `json:"relationships,omitempty" yaml:"relationships,omitempty"`
	// Metadata sets labels/annotations on the domain.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// DomainUpdateRequest represents a request to update a domain.
type DomainUpdateRequest struct {
	// Metadata updates labels/annotations; nil leaves it unchanged.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// DomainRelationships represents domain relationships.
type DomainRelationships struct {
	Organization        *Relationship       `json:"organization,omitempty"         yaml:"organization,omitempty"`
	SharedOrganizations *ToManyRelationship `json:"shared_organizations,omitempty" yaml:"shared_organizations,omitempty"`
}

// Route represents a route.
type Route struct {
	Resource

	Protocol      string             `json:"protocol"           yaml:"protocol"`
	Host          string             `json:"host"               yaml:"host"`
	Path          string             `json:"path"               yaml:"path"`
	Port          *int               `json:"port,omitempty"     yaml:"port,omitempty"`
	URL           string             `json:"url"                yaml:"url"`
	Destinations  []RouteDestination `json:"destinations"       yaml:"destinations"`
	Metadata      *Metadata          `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Relationships RouteRelationships `json:"relationships"      yaml:"relationships"`
}

// RouteCreateRequest represents a request to create a route.
type RouteCreateRequest struct {
	// Host is required for HTTP routes; omit for TCP routes.
	Host *string `json:"host,omitempty" yaml:"host,omitempty"`
	// Path is optional and must begin with "/" when set.
	Path *string `json:"path,omitempty" yaml:"path,omitempty"`
	// Port is required for TCP routes and must be unique per domain.
	Port *int `json:"port,omitempty" yaml:"port,omitempty"`
	// Relationships must include Space and Domain.
	Relationships RouteRelationships `json:"relationships" yaml:"relationships"`
	// Metadata sets labels/annotations on the route.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// RouteUpdateRequest represents a request to update a route.
type RouteUpdateRequest struct {
	// Metadata updates labels/annotations; nil leaves it unchanged.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// RouteRelationships represents route relationships.
type RouteRelationships struct {
	Space  Relationship `json:"space"  yaml:"space"`
	Domain Relationship `json:"domain" yaml:"domain"`
}

// RouteDestination represents a route destination.
type RouteDestination struct {
	GUID     string              `json:"guid"               yaml:"guid"`
	App      RouteDestinationApp `json:"app"                yaml:"app"`
	Port     *int                `json:"port,omitempty"     yaml:"port,omitempty"`
	Protocol *string             `json:"protocol,omitempty" yaml:"protocol,omitempty"`
	Weight   *int                `json:"weight,omitempty"   yaml:"weight,omitempty"`
}

// RouteDestinationApp represents the app in a route destination.
type RouteDestinationApp struct {
	GUID    string   `json:"guid"              yaml:"guid"`
	Process *Process `json:"process,omitempty" yaml:"process,omitempty"`
}

// RouteDestinations represents a list of route destinations.
type RouteDestinations struct {
	Destinations []RouteDestination `json:"destinations" yaml:"destinations"`
	Links        Links              `json:"links"        yaml:"links"`
}

// RouteReservation represents a route reservation check.
type RouteReservation struct {
	MatchingRoute *Route `json:"matching_route" yaml:"matching_route"`
}

// RouteReservationRequest represents a request to check route reservation.
type RouteReservationRequest struct {
	// Host to check (HTTP routes). Optional for TCP routes.
	Host string `json:"host,omitempty" yaml:"host,omitempty"`
	// Path to check; must start with "/" when set.
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
	// Port to check (TCP routes).
	Port *int `json:"port,omitempty" yaml:"port,omitempty"`
}

// User represents a user.
type User struct {
	Resource

	Username         string    `json:"username"           yaml:"username"`
	PresentationName string    `json:"presentation_name"  yaml:"presentation_name"`
	Origin           string    `json:"origin"             yaml:"origin"`
	Metadata         *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// UserCreateRequest represents a request to create a user.
type UserCreateRequest struct {
	// GUID references an existing identity in UAA; alternative to Username/Origin.
	GUID string `json:"guid,omitempty" yaml:"guid,omitempty"`
	// Username creates a user by username when paired with Origin.
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
	// Origin identifies the identity provider for Username (e.g., "uaa", "ldap").
	Origin string `json:"origin,omitempty" yaml:"origin,omitempty"`
	// Metadata sets labels/annotations on the user record.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// UserUpdateRequest represents a request to update a user.
type UserUpdateRequest struct {
	// Metadata updates labels/annotations; nil leaves it unchanged.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// Role represents a role.
type Role struct {
	Resource

	Type          string            `json:"type"          yaml:"type"`
	Relationships RoleRelationships `json:"relationships" yaml:"relationships"`
}

// RoleCreateRequest represents a request to create a role.
type RoleCreateRequest struct {
	// Type is the role name (e.g., "organization_manager", "space_developer").
	Type string `json:"type" yaml:"type"`
	// Relationships specify the target scope (organization or space) and user.
	Relationships RoleRelationships `json:"relationships" yaml:"relationships"`
}

// RoleRelationships represents role relationships.
type RoleRelationships struct {
	User         Relationship  `json:"user"                   yaml:"user"`
	Organization *Relationship `json:"organization,omitempty" yaml:"organization,omitempty"`
	Space        *Relationship `json:"space,omitempty"        yaml:"space,omitempty"`
}

// SecurityGroup represents a security group.
type SecurityGroup struct {
	Resource

	Name            string                       `json:"name"             yaml:"name"`
	GloballyEnabled SecurityGroupGloballyEnabled `json:"globally_enabled" yaml:"globally_enabled"`
	Rules           []SecurityGroupRule          `json:"rules"            yaml:"rules"`
	Relationships   SecurityGroupRelationships   `json:"relationships"    yaml:"relationships"`
}

// SecurityGroupGloballyEnabled represents globally enabled settings for a security group.
type SecurityGroupGloballyEnabled struct {
	Running bool `json:"running" yaml:"running"`
	Staging bool `json:"staging" yaml:"staging"`
}

// SecurityGroupRule defines a network rule for a security group.
// IPv6 security groups can be configured if cc.enable_ipv6 is set to true.
// For 'icmp' protocol, only IPv4 addresses are allowed in Destination.
// For 'icmpv6' protocol, only IPv6 addresses are allowed in Destination.
type SecurityGroupRule struct {
	Protocol string `json:"protocol" yaml:"protocol"`
	// Destination where the rule applies. Must be a singular valid CIDR, IP address,
	// or IP address range unless cc.security_groups.enable_comma_delimited_destinations
	// is enabled. Then, the destination can be a comma-delimited string of CIDRs,
	// IP addresses, or IP address ranges. Octets within IPv4 destinations cannot
	// contain leading zeros; eg. 10.0.0.0/24 is valid, but 010.00.000.0/24 is not.
	Destination string  `json:"destination"           yaml:"destination"`
	Ports       *string `json:"ports,omitempty"       yaml:"ports,omitempty"`
	Type        *int    `json:"type,omitempty"        yaml:"type,omitempty"`
	Code        *int    `json:"code,omitempty"        yaml:"code,omitempty"`
	Description *string `json:"description,omitempty" yaml:"description,omitempty"`
	Log         *bool   `json:"log,omitempty"         yaml:"log,omitempty"`
}

// SecurityGroupRelationships represents security group relationships.
type SecurityGroupRelationships struct {
	RunningSpaces ToManyRelationship `json:"running_spaces" yaml:"running_spaces"`
	StagingSpaces ToManyRelationship `json:"staging_spaces" yaml:"staging_spaces"`
}

// SecurityGroupCreateRequest represents a request to create a security group.
type SecurityGroupCreateRequest struct {
	// Name is the security group name.
	Name string `json:"name" yaml:"name"`
	// GloballyEnabled toggles default binding to running/staging.
	GloballyEnabled *SecurityGroupGloballyEnabled `json:"globally_enabled,omitempty" yaml:"globally_enabled,omitempty"`
	// Rules is the set of egress/ingress rules.
	Rules []SecurityGroupRule `json:"rules,omitempty" yaml:"rules,omitempty"`
	// Relationships optionally bind to running/staging spaces.
	Relationships *SecurityGroupRelationships `json:"relationships,omitempty" yaml:"relationships,omitempty"`
}

// SecurityGroupUpdateRequest represents a request to update a security group.
type SecurityGroupUpdateRequest struct {
	// Name updates the group name; nil leaves it unchanged.
	Name *string `json:"name,omitempty" yaml:"name,omitempty"`
	// GloballyEnabled updates default running/staging enablement.
	GloballyEnabled *SecurityGroupGloballyEnabled `json:"globally_enabled,omitempty" yaml:"globally_enabled,omitempty"`
	// Rules replaces the ruleset when provided.
	Rules []SecurityGroupRule `json:"rules,omitempty" yaml:"rules,omitempty"`
}

// SecurityGroupBindRequest represents a request to bind a security group to spaces.
type SecurityGroupBindRequest struct {
	// Data contains space relationships to bind the group to.
	Data []RelationshipData `json:"data" yaml:"data"`
}

// Package represents a Cloud Foundry package.
type Package struct {
	Resource

	Type          string                `json:"type"                    yaml:"type"`
	Data          *PackageData          `json:"data"                    yaml:"data"`
	State         string                `json:"state"                   yaml:"state"`
	Metadata      *Metadata             `json:"metadata,omitempty"      yaml:"metadata,omitempty"`
	Relationships *PackageRelationships `json:"relationships,omitempty" yaml:"relationships,omitempty"`
}

// PackageData represents package-specific data.
type PackageData struct {
	Checksum *PackageChecksum `json:"checksum,omitempty" yaml:"checksum,omitempty"`
	Error    *string          `json:"error,omitempty"    yaml:"error,omitempty"`
	Image    *string          `json:"image,omitempty"    yaml:"image,omitempty"`    // For Docker packages
	Username *string          `json:"username,omitempty" yaml:"username,omitempty"` // For Docker packages
	Password *string          `json:"password,omitempty" yaml:"password,omitempty"` // For Docker packages
}

// PackageChecksum represents package checksum information.
type PackageChecksum struct {
	Type  string  `json:"type"  yaml:"type"` // e.g., "sha256"
	Value *string `json:"value" yaml:"value"`
}

// PackageRelationships represents the relationships for a package.
type PackageRelationships struct {
	App *Relationship `json:"app,omitempty" yaml:"app,omitempty"`
}

// PackageCreateRequest represents a request to create a package.
type PackageCreateRequest struct {
	// Type is the package type ("bits" or "docker").
	Type string `json:"type" yaml:"type"`
	// Relationships must include App.
	Relationships PackageRelationships `json:"relationships" yaml:"relationships"`
	// Data supplies docker image credentials for docker packages.
	Data *PackageCreateData `json:"data,omitempty" yaml:"data,omitempty"`
	// Metadata sets labels/annotations on the package.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// PackageCreateData represents data for creating a package.
type PackageCreateData struct {
	Image    *string `json:"image,omitempty"    yaml:"image,omitempty"`    // For Docker packages
	Username *string `json:"username,omitempty" yaml:"username,omitempty"` // For Docker packages
	Password *string `json:"password,omitempty" yaml:"password,omitempty"` // For Docker packages
}

// PackageUpdateRequest represents a request to update a package.
type PackageUpdateRequest struct {
	// Metadata updates labels/annotations; nil leaves it unchanged.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// PackageUploadRequest represents a request to upload package bits.
type PackageUploadRequest struct {
	// Bits is the ZIP file contents for bits packages. Not JSON-encoded.
	Bits []byte `json:"-" yaml:"-"`
	// Resources lists matched resources to enable resource-only uploads.
	Resources []PackageResource `json:"resources,omitempty" yaml:"resources,omitempty"`
}

// PackageResource represents a resource in a package upload.
type PackageResource struct {
	SHA1 string `json:"sha1" yaml:"sha1"`
	Size int64  `json:"size" yaml:"size"`
	Path string `json:"path" yaml:"path"`
	Mode string `json:"mode" yaml:"mode"`
}

// PackageCopyRequest represents a request to copy a package.
type PackageCopyRequest struct {
	// Relationships identifies the target App for the copied package.
	Relationships PackageRelationships `json:"relationships" yaml:"relationships"`
}

// Droplet represents a Cloud Foundry droplet.
type Droplet struct {
	Resource

	State             string                `json:"state"                   yaml:"state"`
	Error             *string               `json:"error"                   yaml:"error"`
	Lifecycle         Lifecycle             `json:"lifecycle"               yaml:"lifecycle"`
	ExecutionMetadata string                `json:"execution_metadata"      yaml:"execution_metadata"`
	ProcessTypes      map[string]string     `json:"process_types"           yaml:"process_types"`
	Checksum          *DropletChecksum      `json:"checksum,omitempty"      yaml:"checksum,omitempty"`
	Buildpacks        []DetectedBuildpack   `json:"buildpacks,omitempty"    yaml:"buildpacks,omitempty"`
	Stack             *string               `json:"stack,omitempty"         yaml:"stack,omitempty"`
	Image             *string               `json:"image,omitempty"         yaml:"image,omitempty"`
	Metadata          *Metadata             `json:"metadata,omitempty"      yaml:"metadata,omitempty"`
	Relationships     *DropletRelationships `json:"relationships,omitempty" yaml:"relationships,omitempty"`
}

// DropletChecksum represents droplet checksum information.
type DropletChecksum struct {
	Type  string `json:"type"  yaml:"type"` // e.g., "sha256" or "sha1"
	Value string `json:"value" yaml:"value"`
}

// DetectedBuildpack represents a buildpack detected during staging.
type DetectedBuildpack struct {
	Name          string  `json:"name"                     yaml:"name"`
	DetectOutput  string  `json:"detect_output"            yaml:"detect_output"`
	Version       *string `json:"version,omitempty"        yaml:"version,omitempty"`
	BuildpackName *string `json:"buildpack_name,omitempty" yaml:"buildpack_name,omitempty"`
}

// DropletRelationships represents the relationships for a droplet.
type DropletRelationships struct {
	App *Relationship `json:"app,omitempty" yaml:"app,omitempty"`
}

// DropletCreateRequest represents a request to create a droplet.
type DropletCreateRequest struct {
	// Relationships must include App.
	Relationships DropletRelationships `json:"relationships" yaml:"relationships"`
	// ProcessTypes optionally sets default process types for the droplet.
	ProcessTypes map[string]string `json:"process_types,omitempty" yaml:"process_types,omitempty"`
}

// DropletUpdateRequest represents a request to update a droplet.
type DropletUpdateRequest struct {
	// Metadata updates labels/annotations; nil leaves it unchanged.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	// Image sets a pre-built OCI image reference (platform-dependent).
	Image *string `json:"image,omitempty" yaml:"image,omitempty"`
	// ProcessTypes replaces the droplet's process types.
	ProcessTypes map[string]string `json:"process_types,omitempty" yaml:"process_types,omitempty"`
}

// DropletCopyRequest represents a request to copy a droplet.
type DropletCopyRequest struct {
	// Relationships identifies the target App for the copied droplet.
	Relationships DropletRelationships `json:"relationships" yaml:"relationships"`
}

// Build represents a Cloud Foundry build.
type Build struct {
	Resource

	State                             string              `json:"state"                                   yaml:"state"`
	StagingMemoryInMB                 int                 `json:"staging_memory_in_mb"                    yaml:"staging_memory_in_mb"`
	StagingDiskInMB                   int                 `json:"staging_disk_in_mb"                      yaml:"staging_disk_in_mb"`
	StagingLogRateLimitBytesPerSecond *int                `json:"staging_log_rate_limit_bytes_per_second" yaml:"staging_log_rate_limit_bytes_per_second"`
	Error                             *string             `json:"error"                                   yaml:"error"`
	Lifecycle                         *Lifecycle          `json:"lifecycle,omitempty"                     yaml:"lifecycle,omitempty"`
	Package                           *BuildPackageRef    `json:"package"                                 yaml:"package"`
	Droplet                           *BuildDropletRef    `json:"droplet"                                 yaml:"droplet"`
	CreatedBy                         *UserRef            `json:"created_by"                              yaml:"created_by"`
	Relationships                     *BuildRelationships `json:"relationships,omitempty"                 yaml:"relationships,omitempty"`
	Metadata                          *Metadata           `json:"metadata,omitempty"                      yaml:"metadata,omitempty"`
}

// BuildPackageRef represents a package reference in a build.
type BuildPackageRef struct {
	GUID string `json:"guid" yaml:"guid"`
}

// BuildDropletRef represents a droplet reference in a build.
type BuildDropletRef struct {
	GUID string `json:"guid" yaml:"guid"`
}

// UserRef represents a user reference.
type UserRef struct {
	GUID  string `json:"guid"  yaml:"guid"`
	Name  string `json:"name"  yaml:"name"`
	Email string `json:"email" yaml:"email"`
}

// BuildRelationships represents the relationships for a build.
type BuildRelationships struct {
	App *Relationship `json:"app,omitempty" yaml:"app,omitempty"`
}

// BuildCreateRequest represents a request to create a build.
type BuildCreateRequest struct {
	// Package references the package to stage.
	Package *BuildPackageRef `json:"package" yaml:"package"`
	// Lifecycle optionally overrides staging lifecycle configuration.
	Lifecycle *Lifecycle `json:"lifecycle,omitempty" yaml:"lifecycle,omitempty"`
	// StagingMemoryInMB optionally sets memory for staging.
	StagingMemoryInMB *int `json:"staging_memory_in_mb,omitempty" yaml:"staging_memory_in_mb,omitempty"`
	// StagingDiskInMB optionally sets disk for staging.
	StagingDiskInMB *int `json:"staging_disk_in_mb,omitempty" yaml:"staging_disk_in_mb,omitempty"`
	// StagingLogRateLimitBytesPerSecond optionally limits staging log rate.
	StagingLogRateLimitBytesPerSecond *int `json:"staging_log_rate_limit_bytes_per_second,omitempty" yaml:"staging_log_rate_limit_bytes_per_second,omitempty"`
	// Metadata sets labels/annotations on the build.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// BuildUpdateRequest represents a request to update a build.
type BuildUpdateRequest struct {
	// Metadata updates labels/annotations; nil leaves it unchanged.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	// State may be set to cancel builds in certain states.
	State *string `json:"state,omitempty" yaml:"state,omitempty"`
}

// Buildpack represents a Cloud Foundry buildpack.
type Buildpack struct {
	Resource

	Name      string    `json:"name"               yaml:"name"`
	State     string    `json:"state"              yaml:"state"`
	Filename  *string   `json:"filename"           yaml:"filename"`
	Stack     *string   `json:"stack"              yaml:"stack"`
	Position  int       `json:"position"           yaml:"position"`
	Lifecycle string    `json:"lifecycle"          yaml:"lifecycle"`
	Enabled   bool      `json:"enabled"            yaml:"enabled"`
	Locked    bool      `json:"locked"             yaml:"locked"`
	Metadata  *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Links     Links     `json:"links,omitempty"    yaml:"links,omitempty"`
}

// BuildpackCreateRequest represents a request to create a buildpack.
type BuildpackCreateRequest struct {
	// Name is the buildpack name.
	Name string `json:"name" yaml:"name"`
	// Stack limits the buildpack to a specific stack when set.
	Stack *string `json:"stack,omitempty" yaml:"stack,omitempty"`
	// Position controls buildpack order.
	Position *int `json:"position,omitempty" yaml:"position,omitempty"`
	// Lifecycle targets a specific lifecycle when set.
	Lifecycle *string `json:"lifecycle,omitempty" yaml:"lifecycle,omitempty"`
	// Enabled toggles whether the buildpack is active.
	Enabled *bool `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	// Locked prevents changes or re-upload when true.
	Locked *bool `json:"locked,omitempty" yaml:"locked,omitempty"`
	// Metadata sets labels/annotations on the buildpack.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// BuildpackUpdateRequest represents a request to update a buildpack.
type BuildpackUpdateRequest struct {
	// Name updates the buildpack name.
	Name *string `json:"name,omitempty" yaml:"name,omitempty"`
	// Stack updates the targeted stack.
	Stack *string `json:"stack,omitempty" yaml:"stack,omitempty"`
	// Position updates the order.
	Position *int `json:"position,omitempty" yaml:"position,omitempty"`
	// Enabled toggles activation.
	Enabled *bool `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	// Locked toggles immutability.
	Locked *bool `json:"locked,omitempty" yaml:"locked,omitempty"`
	// Metadata updates labels/annotations; nil leaves it unchanged.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// Deployment represents a Cloud Foundry deployment.
type Deployment struct {
	Resource

	State           string                   `json:"state"                   yaml:"state"`
	Status          DeploymentStatus         `json:"status"                  yaml:"status"`
	Strategy        string                   `json:"strategy"                yaml:"strategy"`
	Options         *DeploymentOptions       `json:"options,omitempty"       yaml:"options,omitempty"`
	Droplet         *DeploymentDropletRef    `json:"droplet"                 yaml:"droplet"`
	PreviousDroplet *DeploymentDropletRef    `json:"previous_droplet"        yaml:"previous_droplet"`
	NewProcesses    []DeploymentProcess      `json:"new_processes"           yaml:"new_processes"`
	Revision        *DeploymentRevisionRef   `json:"revision,omitempty"      yaml:"revision,omitempty"`
	Metadata        *Metadata                `json:"metadata,omitempty"      yaml:"metadata,omitempty"`
	Relationships   *DeploymentRelationships `json:"relationships,omitempty" yaml:"relationships,omitempty"`
}

// DeploymentStatus represents the status of a deployment.
type DeploymentStatus struct {
	Value   string                   `json:"value"             yaml:"value"`
	Reason  string                   `json:"reason"            yaml:"reason"`
	Details *DeploymentStatusDetails `json:"details,omitempty" yaml:"details,omitempty"`
	Canary  *DeploymentCanaryStatus  `json:"canary,omitempty"  yaml:"canary,omitempty"`
}

// DeploymentStatusDetails provides details about deployment status.
type DeploymentStatusDetails struct {
	LastHealthyAt    *time.Time `json:"last_healthy_at,omitempty"    yaml:"last_healthy_at,omitempty"`
	LastStatusChange *time.Time `json:"last_status_change,omitempty" yaml:"last_status_change,omitempty"`
	Error            *string    `json:"error,omitempty"              yaml:"error,omitempty"`
}

// DeploymentCanaryStatus represents canary deployment status.
type DeploymentCanaryStatus struct {
	Steps DeploymentCanarySteps `json:"steps" yaml:"steps"`
}

// DeploymentCanarySteps represents canary deployment step info.
type DeploymentCanarySteps struct {
	Current int `json:"current" yaml:"current"`
	Total   int `json:"total"   yaml:"total"`
}

// DeploymentOptions represents deployment options.
type DeploymentOptions struct {
	MaxInFlight                  *int                     `json:"max_in_flight,omitempty"                      yaml:"max_in_flight,omitempty"`
	WebInstances                 *int                     `json:"web_instances,omitempty"                      yaml:"web_instances,omitempty"`
	MemoryInMB                   *int                     `json:"memory_in_mb,omitempty"                       yaml:"memory_in_mb,omitempty"`
	DiskInMB                     *int                     `json:"disk_in_mb,omitempty"                         yaml:"disk_in_mb,omitempty"`
	LogRateLimitInBytesPerSecond *int                     `json:"log_rate_limit_in_bytes_per_second,omitempty" yaml:"log_rate_limit_in_bytes_per_second,omitempty"`
	Canary                       *DeploymentCanaryOptions `json:"canary,omitempty"                             yaml:"canary,omitempty"`
}

// DeploymentCanaryOptions represents canary deployment options.
type DeploymentCanaryOptions struct {
	Steps []DeploymentCanaryStep `json:"steps" yaml:"steps"`
}

// DeploymentCanaryStep represents a canary deployment step.
type DeploymentCanaryStep struct {
	Instances int `json:"instances"           yaml:"instances"`
	WaitTime  int `json:"wait_time,omitempty" yaml:"wait_time,omitempty"`
}

// DeploymentDropletRef represents a droplet reference in a deployment.
type DeploymentDropletRef struct {
	GUID string `json:"guid" yaml:"guid"`
}

// DeploymentRevisionRef represents a revision reference in a deployment.
type DeploymentRevisionRef struct {
	GUID    string `json:"guid"    yaml:"guid"`
	Version int    `json:"version" yaml:"version"`
}

// DeploymentProcess represents a process created during deployment.
type DeploymentProcess struct {
	GUID string `json:"guid" yaml:"guid"`
	Type string `json:"type" yaml:"type"`
}

// DeploymentRelationships represents the relationships for a deployment.
type DeploymentRelationships struct {
	App *Relationship `json:"app,omitempty" yaml:"app,omitempty"`
}

// DeploymentCreateRequest represents a request to create a deployment.
type DeploymentCreateRequest struct {
	// Droplet specifies the droplet to deploy. Exactly one of Droplet or Revision should be set.
	Droplet *DeploymentDropletRef `json:"droplet,omitempty" yaml:"droplet,omitempty"`
	// Revision specifies a revision to deploy. Exactly one of Droplet or Revision should be set.
	Revision *DeploymentRevisionRef `json:"revision,omitempty" yaml:"revision,omitempty"`
	// Strategy optionally sets the deployment strategy (e.g., "rolling").
	Strategy *string `json:"strategy,omitempty" yaml:"strategy,omitempty"`
	// Options configures rollout settings (in-flight, canary, etc.).
	Options *DeploymentOptions `json:"options,omitempty" yaml:"options,omitempty"`
	// Relationships must include App.
	Relationships DeploymentRelationships `json:"relationships" yaml:"relationships"`
	// Metadata sets labels/annotations on the deployment.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// DeploymentUpdateRequest represents a request to update a deployment.
type DeploymentUpdateRequest struct {
	// Metadata updates labels/annotations; nil leaves it unchanged.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// Process represents a Cloud Foundry process.
type Process struct {
	Resource

	Type                         string                `json:"type"                               yaml:"type"`
	Command                      *string               `json:"command"                            yaml:"command"`
	User                         string                `json:"user,omitempty"                     yaml:"user,omitempty"`
	Instances                    int                   `json:"instances"                          yaml:"instances"`
	MemoryInMB                   int                   `json:"memory_in_mb"                       yaml:"memory_in_mb"`
	DiskInMB                     int                   `json:"disk_in_mb"                         yaml:"disk_in_mb"`
	LogRateLimitInBytesPerSecond *int                  `json:"log_rate_limit_in_bytes_per_second" yaml:"log_rate_limit_in_bytes_per_second"`
	HealthCheck                  *HealthCheck          `json:"health_check"                       yaml:"health_check"`
	ReadinessHealthCheck         *ReadinessHealthCheck `json:"readiness_health_check"             yaml:"readiness_health_check"`
	Version                      string                `json:"version,omitempty"                  yaml:"version,omitempty"`
	Metadata                     *Metadata             `json:"metadata,omitempty"                 yaml:"metadata,omitempty"`
	Relationships                *ProcessRelationships `json:"relationships,omitempty"            yaml:"relationships,omitempty"`
}

// ProcessRelationships represents the relationships for a process.
type ProcessRelationships struct {
	App      *Relationship `json:"app,omitempty"      yaml:"app,omitempty"`
	Revision *Relationship `json:"revision,omitempty" yaml:"revision,omitempty"`
}

// HealthCheck represents a process health check.
type HealthCheck struct {
	Type string           `json:"type"           yaml:"type"` // "port", "process", or "http"
	Data *HealthCheckData `json:"data,omitempty" yaml:"data,omitempty"`
}

// HealthCheckData represents health check configuration data.
type HealthCheckData struct {
	Timeout           *int    `json:"timeout,omitempty"            yaml:"timeout,omitempty"`
	InvocationTimeout *int    `json:"invocation_timeout,omitempty" yaml:"invocation_timeout,omitempty"`
	Interval          *int    `json:"interval,omitempty"           yaml:"interval,omitempty"`
	Endpoint          *string `json:"endpoint,omitempty"           yaml:"endpoint,omitempty"` // For HTTP health checks
}

// ReadinessHealthCheck represents a process readiness health check.
type ReadinessHealthCheck struct {
	Type string                    `json:"type"           yaml:"type"` // "process", "port", or "http"
	Data *ReadinessHealthCheckData `json:"data,omitempty" yaml:"data,omitempty"`
}

// ReadinessHealthCheckData represents readiness health check configuration data.
type ReadinessHealthCheckData struct {
	InvocationTimeout *int    `json:"invocation_timeout,omitempty" yaml:"invocation_timeout,omitempty"`
	Interval          *int    `json:"interval,omitempty"           yaml:"interval,omitempty"`
	Endpoint          *string `json:"endpoint,omitempty"           yaml:"endpoint,omitempty"` // For HTTP readiness checks
}

// ProcessUpdateRequest represents a request to update a process.
type ProcessUpdateRequest struct {
	// Command overrides the start command; nil leaves unchanged.
	Command *string `json:"command,omitempty" yaml:"command,omitempty"`
	// HealthCheck updates liveness health check configuration.
	HealthCheck *HealthCheck `json:"health_check,omitempty" yaml:"health_check,omitempty"`
	// ReadinessHealthCheck updates readiness check configuration.
	ReadinessHealthCheck *ReadinessHealthCheck `json:"readiness_health_check,omitempty" yaml:"readiness_health_check,omitempty"`
	// Metadata updates labels/annotations; nil leaves it unchanged.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// ProcessScaleRequest represents a request to scale a process.
type ProcessScaleRequest struct {
	// Instances sets the desired instance count.
	Instances *int `json:"instances,omitempty" yaml:"instances,omitempty"`
	// MemoryInMB sets memory per instance.
	MemoryInMB *int `json:"memory_in_mb,omitempty" yaml:"memory_in_mb,omitempty"`
	// DiskInMB sets disk per instance.
	DiskInMB *int `json:"disk_in_mb,omitempty" yaml:"disk_in_mb,omitempty"`
	// LogRateLimitInBytesPerSecond sets per-instance log rate limit.
	LogRateLimitInBytesPerSecond *int `json:"log_rate_limit_in_bytes_per_second,omitempty" yaml:"log_rate_limit_in_bytes_per_second,omitempty"`
}

// ProcessStats represents statistics for a process.
type ProcessStats struct {
	Pagination *Pagination          `json:"pagination" yaml:"pagination"`
	Resources  []ProcessStatsDetail `json:"resources"  yaml:"resources"`
}

// ProcessStatsDetail represents detailed statistics for a process instance.
type ProcessStatsDetail struct {
	Type             string                `json:"type"                        yaml:"type"`
	Index            int                   `json:"index"                       yaml:"index"`
	State            string                `json:"state"                       yaml:"state"`
	Usage            *ProcessUsage         `json:"usage,omitempty"             yaml:"usage,omitempty"`
	Host             string                `json:"host,omitempty"              yaml:"host,omitempty"`
	InstancePorts    []ProcessInstancePort `json:"instance_ports,omitempty"    yaml:"instance_ports,omitempty"`
	Uptime           int                   `json:"uptime,omitempty"            yaml:"uptime,omitempty"`
	MemQuota         int64                 `json:"mem_quota,omitempty"         yaml:"mem_quota,omitempty"`
	DiskQuota        int64                 `json:"disk_quota,omitempty"        yaml:"disk_quota,omitempty"`
	FdsQuota         int                   `json:"fds_quota,omitempty"         yaml:"fds_quota,omitempty"`
	IsolationSegment *string               `json:"isolation_segment,omitempty" yaml:"isolation_segment,omitempty"`
	Details          *string               `json:"details,omitempty"           yaml:"details,omitempty"`
}

// ProcessUsage represents CPU and memory usage for a process instance.
type ProcessUsage struct {
	Time           string  `json:"time"                      yaml:"time"`
	CPU            float64 `json:"cpu"                       yaml:"cpu"`
	CPUEntitlement float64 `json:"cpu_entitlement,omitempty" yaml:"cpu_entitlement,omitempty"`
	Mem            int64   `json:"mem"                       yaml:"mem"`
	Disk           int64   `json:"disk"                      yaml:"disk"`
	LogRate        int     `json:"log_rate"                  yaml:"log_rate"`
}

// ProcessInstancePort represents port mappings for a process instance.
type ProcessInstancePort struct {
	External             int `json:"external"                          yaml:"external"`
	Internal             int `json:"internal"                          yaml:"internal"`
	ExternalTLSProxyPort int `json:"external_tls_proxy_port,omitempty" yaml:"external_tls_proxy_port,omitempty"`
	InternalTLSProxyPort int `json:"internal_tls_proxy_port,omitempty" yaml:"internal_tls_proxy_port,omitempty"`
}

// Task represents a Cloud Foundry task.
type Task struct {
	Resource

	SequenceID                   int                `json:"sequence_id"                        yaml:"sequence_id"`
	Name                         string             `json:"name"                               yaml:"name"`
	Command                      string             `json:"command,omitempty"                  yaml:"command,omitempty"`
	User                         *string            `json:"user"                               yaml:"user"`
	State                        string             `json:"state"                              yaml:"state"`
	MemoryInMB                   int                `json:"memory_in_mb"                       yaml:"memory_in_mb"`
	DiskInMB                     int                `json:"disk_in_mb"                         yaml:"disk_in_mb"`
	LogRateLimitInBytesPerSecond *int               `json:"log_rate_limit_in_bytes_per_second" yaml:"log_rate_limit_in_bytes_per_second"`
	Result                       *TaskResult        `json:"result,omitempty"                   yaml:"result,omitempty"`
	DropletGUID                  string             `json:"droplet_guid"                       yaml:"droplet_guid"`
	Metadata                     *Metadata          `json:"metadata,omitempty"                 yaml:"metadata,omitempty"`
	Relationships                *TaskRelationships `json:"relationships,omitempty"            yaml:"relationships,omitempty"`
}

// TaskResult represents the result of a task execution.
type TaskResult struct {
	FailureReason *string `json:"failure_reason" yaml:"failure_reason"`
}

// TaskRelationships represents the relationships for a task.
type TaskRelationships struct {
	App *Relationship `json:"app,omitempty" yaml:"app,omitempty"`
}

// TaskCreateRequest represents a request to create a task.
type TaskCreateRequest struct {
	// Command is the task command when not using a Template.
	Command *string `json:"command,omitempty" yaml:"command,omitempty"`
	// Name optionally sets a friendly task name.
	Name *string `json:"name,omitempty" yaml:"name,omitempty"`
	// MemoryInMB and DiskInMB set task resources.
	MemoryInMB *int `json:"memory_in_mb,omitempty" yaml:"memory_in_mb,omitempty"`
	DiskInMB   *int `json:"disk_in_mb,omitempty"   yaml:"disk_in_mb,omitempty"`
	// LogRateLimitInBytesPerSecond limits task log rate.
	LogRateLimitInBytesPerSecond *int `json:"log_rate_limit_in_bytes_per_second,omitempty" yaml:"log_rate_limit_in_bytes_per_second,omitempty"`
	// Template references a process to derive command/env/runtime from.
	Template *TaskTemplate `json:"template,omitempty" yaml:"template,omitempty"`
	// Metadata sets labels/annotations on the task.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	// DropletGUID optionally pins the task to a specific droplet.
	DropletGUID *string `json:"droplet_guid,omitempty" yaml:"droplet_guid,omitempty"`
}

// TaskTemplate represents a template for creating a task from a process.
type TaskTemplate struct {
	Process *TaskTemplateProcess `json:"process,omitempty" yaml:"process,omitempty"`
}

// TaskTemplateProcess represents a process reference in a task template.
type TaskTemplateProcess struct {
	GUID string `json:"guid" yaml:"guid"`
}

// TaskUpdateRequest represents a request to update a task.
type TaskUpdateRequest struct {
	// Metadata updates labels/annotations; nil leaves it unchanged.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// Stack represents a Cloud Foundry stack (a pre-built rootfs and associated executables).
type Stack struct {
	Resource

	Name             string    `json:"name"               yaml:"name"`
	Description      string    `json:"description"        yaml:"description"`
	BuildRootfsImage string    `json:"build_rootfs_image" yaml:"build_rootfs_image"`
	RunRootfsImage   string    `json:"run_rootfs_image"   yaml:"run_rootfs_image"`
	Default          bool      `json:"default"            yaml:"default"`
	Metadata         *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Links            Links     `json:"links,omitempty"    yaml:"links,omitempty"`
}

// StackCreateRequest is the request for creating a stack.
type StackCreateRequest struct {
	// Name is the stack name.
	Name string `json:"name" yaml:"name"`
	// Description describes the stack.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// BuildRootfsImage and RunRootfsImage point to OCI images used for staging and running.
	BuildRootfsImage string `json:"build_rootfs_image,omitempty" yaml:"build_rootfs_image,omitempty"`
	RunRootfsImage   string `json:"run_rootfs_image,omitempty"   yaml:"run_rootfs_image,omitempty"`
	// Metadata sets labels/annotations on the stack.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// StackUpdateRequest is the request for updating a stack.
type StackUpdateRequest struct {
	// Metadata updates labels/annotations; nil leaves it unchanged.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// IsolationSegment represents an isolation segment.
type IsolationSegment struct {
	Resource

	Name     string    `json:"name"               yaml:"name"`
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// IsolationSegmentCreateRequest represents a request to create an isolation segment.
type IsolationSegmentCreateRequest struct {
	// Name is the isolation segment name.
	Name string `json:"name" yaml:"name"`
	// Metadata sets labels/annotations on the isolation segment.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// IsolationSegmentUpdateRequest represents a request to update an isolation segment.
type IsolationSegmentUpdateRequest struct {
	// Name updates the isolation segment name; nil leaves it unchanged.
	Name *string `json:"name,omitempty" yaml:"name,omitempty"`
	// Metadata updates labels/annotations; nil leaves it unchanged.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// IsolationSegmentEntitleOrganizationsRequest represents a request to entitle organizations.
type IsolationSegmentEntitleOrganizationsRequest = ToManyRelationship

// FeatureFlag represents a feature flag.
type FeatureFlag struct {
	Name               string     `json:"name"                 yaml:"name"`
	Enabled            bool       `json:"enabled"              yaml:"enabled"`
	UpdatedAt          *time.Time `json:"updated_at"           yaml:"updated_at"`
	CustomErrorMessage *string    `json:"custom_error_message" yaml:"custom_error_message"`
	Links              Links      `json:"links,omitempty"      yaml:"links,omitempty"`
}

// FeatureFlagUpdateRequest represents a request to update a feature flag.
type FeatureFlagUpdateRequest struct {
	// Enabled toggles the feature flag.
	Enabled bool `json:"enabled" yaml:"enabled"`
	// CustomErrorMessage optionally overrides the error shown when disabled.
	CustomErrorMessage *string `json:"custom_error_message,omitempty" yaml:"custom_error_message,omitempty"`
}

// ServiceBroker represents a service broker.
type ServiceBroker struct {
	Resource

	Name          string                     `json:"name"               yaml:"name"`
	URL           string                     `json:"url"                yaml:"url"`
	Relationships ServiceBrokerRelationships `json:"relationships"      yaml:"relationships"`
	Metadata      *Metadata                  `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// ServiceBrokerRelationships represents service broker relationships.
type ServiceBrokerRelationships struct {
	Space *Relationship `json:"space,omitempty" yaml:"space,omitempty"`
}

// ServiceBrokerAuthentication represents authentication for a service broker.
type ServiceBrokerAuthentication struct {
	Type        string                                 `json:"type"        yaml:"type"`
	Credentials ServiceBrokerAuthenticationCredentials `json:"credentials" yaml:"credentials"`
}

// ServiceBrokerAuthenticationCredentials represents authentication credentials.
type ServiceBrokerAuthenticationCredentials struct {
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

// ServiceBrokerCreateRequest represents a request to create a service broker.
type ServiceBrokerCreateRequest struct {
	// Name is the broker name; URL is the broker endpoint.
	Name string `json:"name" yaml:"name"`
	URL  string `json:"url"  yaml:"url"`
	// Authentication supplies basic credentials or other supported types.
	Authentication ServiceBrokerAuthentication `json:"authentication" yaml:"authentication"`
	// Relationships optionally scope the broker to a space.
	Relationships *ServiceBrokerRelationships `json:"relationships,omitempty" yaml:"relationships,omitempty"`
	// Metadata sets labels/annotations on the broker.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// ServiceBrokerUpdateRequest represents a request to update a service broker.
type ServiceBrokerUpdateRequest struct {
	// Name/URL update broker identification; nil leaves unchanged.
	Name *string `json:"name,omitempty" yaml:"name,omitempty"`
	URL  *string `json:"url,omitempty"  yaml:"url,omitempty"`
	// Authentication updates credentials.
	Authentication *ServiceBrokerAuthentication `json:"authentication,omitempty" yaml:"authentication,omitempty"`
	// Metadata updates labels/annotations; nil leaves it unchanged.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// ServiceOffering represents a service offering.
type ServiceOffering struct {
	Resource

	Name             string                       `json:"name"                        yaml:"name"`
	Description      string                       `json:"description"                 yaml:"description"`
	Available        bool                         `json:"available"                   yaml:"available"`
	Tags             []string                     `json:"tags"                        yaml:"tags"`
	Requires         []string                     `json:"requires"                    yaml:"requires"`
	Shareable        bool                         `json:"shareable"                   yaml:"shareable"`
	DocumentationURL *string                      `json:"documentation_url,omitempty" yaml:"documentation_url,omitempty"`
	BrokerCatalog    ServiceOfferingCatalog       `json:"broker_catalog"              yaml:"broker_catalog"`
	Relationships    ServiceOfferingRelationships `json:"relationships"               yaml:"relationships"`
	Metadata         *Metadata                    `json:"metadata,omitempty"          yaml:"metadata,omitempty"`
}

// ServiceOfferingCatalog represents catalog information for a service offering.
type ServiceOfferingCatalog struct {
	ID       string                         `json:"id"                 yaml:"id"`
	Metadata map[string]interface{}         `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Features ServiceOfferingCatalogFeatures `json:"features"           yaml:"features"`
}

// ServiceOfferingCatalogFeatures represents features of a service offering catalog.
type ServiceOfferingCatalogFeatures struct {
	PlanUpdateable       bool `json:"plan_updateable"       yaml:"plan_updateable"`
	Bindable             bool `json:"bindable"              yaml:"bindable"`
	InstancesRetrievable bool `json:"instances_retrievable" yaml:"instances_retrievable"`
	BindingsRetrievable  bool `json:"bindings_retrievable"  yaml:"bindings_retrievable"`
	AllowContextUpdates  bool `json:"allow_context_updates" yaml:"allow_context_updates"`
}

// ServiceOfferingRelationships represents service offering relationships.
type ServiceOfferingRelationships struct {
	ServiceBroker Relationship `json:"service_broker" yaml:"service_broker"`
}

// ServiceOfferingUpdateRequest represents a request to update a service offering.
type ServiceOfferingUpdateRequest struct {
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// ServicePlan represents a service plan.
type ServicePlan struct {
	Resource

	Name            string                   `json:"name"                       yaml:"name"`
	Description     string                   `json:"description"                yaml:"description"`
	Available       bool                     `json:"available"                  yaml:"available"`
	VisibilityType  string                   `json:"visibility_type"            yaml:"visibility_type"`
	Free            bool                     `json:"free"                       yaml:"free"`
	Costs           []ServicePlanCost        `json:"costs"                      yaml:"costs"`
	MaintenanceInfo *ServicePlanMaintenance  `json:"maintenance_info,omitempty" yaml:"maintenance_info,omitempty"`
	BrokerCatalog   ServicePlanCatalog       `json:"broker_catalog"             yaml:"broker_catalog"`
	Schemas         ServicePlanSchemas       `json:"schemas"                    yaml:"schemas"`
	Relationships   ServicePlanRelationships `json:"relationships"              yaml:"relationships"`
	Metadata        *Metadata                `json:"metadata,omitempty"         yaml:"metadata,omitempty"`
}

// ServicePlanCost represents the cost information for a service plan.
type ServicePlanCost struct {
	Amount   float64 `json:"amount"   yaml:"amount"`
	Currency string  `json:"currency" yaml:"currency"`
	Unit     string  `json:"unit"     yaml:"unit"`
}

// ServicePlanMaintenance represents maintenance information for a service plan.
type ServicePlanMaintenance struct {
	Version     string `json:"version"     yaml:"version"`
	Description string `json:"description" yaml:"description"`
}

// ServicePlanCatalog represents catalog information for a service plan.
type ServicePlanCatalog struct {
	ID                     string                     `json:"id"                                 yaml:"id"`
	Metadata               map[string]interface{}     `json:"metadata,omitempty"                 yaml:"metadata,omitempty"`
	MaximumPollingDuration *int                       `json:"maximum_polling_duration,omitempty" yaml:"maximum_polling_duration,omitempty"`
	Features               ServicePlanCatalogFeatures `json:"features"                           yaml:"features"`
}

// ServicePlanCatalogFeatures represents features of a service plan catalog.
type ServicePlanCatalogFeatures struct {
	PlanUpdateable bool `json:"plan_updateable" yaml:"plan_updateable"`
	Bindable       bool `json:"bindable"        yaml:"bindable"`
}

// ServicePlanSchemas represents the schemas for a service plan.
type ServicePlanSchemas struct {
	ServiceInstance ServiceInstanceSchema `json:"service_instance" yaml:"service_instance"`
	ServiceBinding  ServiceBindingSchema  `json:"service_binding"  yaml:"service_binding"`
}

// ServiceInstanceSchema represents the schema for service instance operations.
type ServiceInstanceSchema struct {
	Create SchemaDefinition `json:"create" yaml:"create"`
	Update SchemaDefinition `json:"update" yaml:"update"`
}

// ServiceBindingSchema represents the schema for service binding operations.
type ServiceBindingSchema struct {
	Create SchemaDefinition `json:"create" yaml:"create"`
}

// SchemaDefinition represents a schema definition.
type SchemaDefinition struct {
	Parameters map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

// ServicePlanRelationships represents service plan relationships.
type ServicePlanRelationships struct {
	ServiceOffering Relationship  `json:"service_offering" yaml:"service_offering"`
	Space           *Relationship `json:"space,omitempty"  yaml:"space,omitempty"`
}

// ServicePlanUpdateRequest represents a request to update a service plan.
type ServicePlanUpdateRequest struct {
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// ServicePlanVisibility represents service plan visibility.
type ServicePlanVisibility struct {
	Type          string                      `json:"type"                    yaml:"type"`
	Organizations []ServicePlanVisibilityOrg  `json:"organizations,omitempty" yaml:"organizations,omitempty"`
	Space         *ServicePlanVisibilitySpace `json:"space,omitempty"         yaml:"space,omitempty"`
}

// ServicePlanVisibilityOrg represents an organization in service plan visibility.
type ServicePlanVisibilityOrg struct {
	GUID string `json:"guid"           yaml:"guid"`
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
}

// ServicePlanVisibilitySpace represents a space in service plan visibility.
type ServicePlanVisibilitySpace struct {
	GUID string `json:"guid"           yaml:"guid"`
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
}

// ServicePlanVisibilityUpdateRequest represents a request to update service plan visibility.
type ServicePlanVisibilityUpdateRequest struct {
	// Type sets visibility (e.g., "public", "organization").
	Type string `json:"type" yaml:"type"`
	// Organizations lists org GUIDs when type is organization-scoped.
	Organizations []string `json:"organizations,omitempty" yaml:"organizations,omitempty"`
}

// ServicePlanVisibilityApplyRequest represents a request to apply service plan visibility.
type ServicePlanVisibilityApplyRequest struct {
	// Type sets visibility (e.g., "public", "organization").
	Type string `json:"type" yaml:"type"`
	// Organizations lists org GUIDs when type is organization-scoped.
	Organizations []string `json:"organizations,omitempty" yaml:"organizations,omitempty"`
}

// ServiceInstance represents a service instance.
type ServiceInstance struct {
	Resource

	Name             string                        `json:"name"                        yaml:"name"`
	Type             string                        `json:"type"                        yaml:"type"` // "managed" or "user-provided"
	Tags             []string                      `json:"tags"                        yaml:"tags"`
	MaintenanceInfo  *ServiceInstanceMaintenance   `json:"maintenance_info,omitempty"  yaml:"maintenance_info,omitempty"`
	UpgradeAvailable bool                          `json:"upgrade_available"           yaml:"upgrade_available"`
	DashboardURL     *string                       `json:"dashboard_url,omitempty"     yaml:"dashboard_url,omitempty"`
	LastOperation    *ServiceInstanceLastOperation `json:"last_operation"              yaml:"last_operation"`
	SyslogDrainURL   *string                       `json:"syslog_drain_url,omitempty"  yaml:"syslog_drain_url,omitempty"`  // For user-provided
	RouteServiceURL  *string                       `json:"route_service_url,omitempty" yaml:"route_service_url,omitempty"` // For user-provided
	Relationships    ServiceInstanceRelationships  `json:"relationships"               yaml:"relationships"`
	Metadata         *Metadata                     `json:"metadata,omitempty"          yaml:"metadata,omitempty"`
}

// ServiceInstanceMaintenance represents maintenance information for a service instance.
type ServiceInstanceMaintenance struct {
	Version     string `json:"version"               yaml:"version"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// ServiceInstanceLastOperation represents the last operation performed on a service instance.
type ServiceInstanceLastOperation struct {
	Type        string     `json:"type"        yaml:"type"`  // "create", "update", "delete"
	State       string     `json:"state"       yaml:"state"` // "initial", "in progress", "succeeded", "failed"
	Description string     `json:"description" yaml:"description"`
	CreatedAt   *time.Time `json:"created_at"  yaml:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"  yaml:"updated_at"`
}

// ServiceInstanceRelationships represents service instance relationships.
type ServiceInstanceRelationships struct {
	Space       Relationship  `json:"space"                  yaml:"space"`
	ServicePlan *Relationship `json:"service_plan,omitempty" yaml:"service_plan,omitempty"` // For managed instances
}

// ServiceInstanceCreateRequest represents a request to create a service instance.
type ServiceInstanceCreateRequest struct {
	// Type chooses "managed" (brokered) or "user-provided".
	Type string `json:"type" yaml:"type"`
	// Name is the service instance name.
	Name string `json:"name" yaml:"name"`
	// Tags are arbitrary labels.
	Tags []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	// Parameters are passed to the broker for managed instances.
	Parameters map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	// Credentials, SyslogDrainURL, and RouteServiceURL apply to user-provided instances.
	Credentials     map[string]interface{} `json:"credentials,omitempty"       yaml:"credentials,omitempty"`
	SyslogDrainURL  *string                `json:"syslog_drain_url,omitempty"  yaml:"syslog_drain_url,omitempty"`
	RouteServiceURL *string                `json:"route_service_url,omitempty" yaml:"route_service_url,omitempty"`
	// Relationships must include Space; ServicePlan is required for managed.
	Relationships ServiceInstanceRelationships `json:"relationships" yaml:"relationships"`
	// Metadata sets labels/annotations on the instance.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// ServiceInstanceUpdateRequest represents a request to update a service instance.
type ServiceInstanceUpdateRequest struct {
	// Name/Tags update identification; nil zero value leaves unchanged.
	Name *string  `json:"name,omitempty" yaml:"name,omitempty"`
	Tags []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	// Parameters are passed to the broker for managed instances.
	Parameters map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	// Credentials, SyslogDrainURL, and RouteServiceURL apply to user-provided instances.
	Credentials     map[string]interface{} `json:"credentials,omitempty"       yaml:"credentials,omitempty"`
	SyslogDrainURL  *string                `json:"syslog_drain_url,omitempty"  yaml:"syslog_drain_url,omitempty"`
	RouteServiceURL *string                `json:"route_service_url,omitempty" yaml:"route_service_url,omitempty"`
	// MaintenanceInfo supplies upgrade target for brokered instances.
	MaintenanceInfo *ServiceInstanceMaintenance `json:"maintenance_info,omitempty" yaml:"maintenance_info,omitempty"`
	// Metadata updates labels/annotations; nil leaves it unchanged.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	// Relationships may be updated for targeted operations (rare).
	Relationships *ServiceInstanceRelationships `json:"relationships,omitempty" yaml:"relationships,omitempty"`
}

// ServiceInstanceParameters represents parameters for a managed service instance.
type ServiceInstanceParameters struct {
	Parameters map[string]interface{} `json:"parameters" yaml:"parameters"`
}

// ServiceInstanceCredentials represents credentials for a user-provided service instance.
type ServiceInstanceCredentials struct {
	Credentials map[string]interface{} `json:"credentials" yaml:"credentials"`
}

// ServiceInstanceSharedSpacesRelationships represents shared spaces relationships.
type ServiceInstanceSharedSpacesRelationships struct {
	Data  []Relationship `json:"data"            yaml:"data"`
	Links Links          `json:"links,omitempty" yaml:"links,omitempty"`
}

// ServiceInstanceShareRequest represents a request to share a service instance.
type ServiceInstanceShareRequest struct {
	// Data contains target Space relationships to share with.
	Data []Relationship `json:"data" yaml:"data"`
}

// ServiceInstanceUsageSummary represents usage summary for service instances.
type ServiceInstanceUsageSummary struct {
	UsageSummary ServiceInstanceUsageData `json:"usage_summary"   yaml:"usage_summary"`
	Links        Links                    `json:"links,omitempty" yaml:"links,omitempty"`
}

// ServiceInstanceUsageData represents usage data for service instances.
type ServiceInstanceUsageData struct {
	StartedInstances int `json:"started_instances" yaml:"started_instances"`
	MemoryInMB       int `json:"memory_in_mb"      yaml:"memory_in_mb"`
}

// ServiceInstancePermissions represents permissions for a service instance.
type ServiceInstancePermissions struct {
	Read   bool `json:"read"   yaml:"read"`
	Manage bool `json:"manage" yaml:"manage"`
}

// ServiceCredentialBinding represents a service credential binding (new name for service binding).
type ServiceCredentialBinding struct {
	Resource

	Name          string                                 `json:"name"                     yaml:"name"`
	Type          string                                 `json:"type"                     yaml:"type"` // "app" or "key"
	LastOperation *ServiceCredentialBindingLastOperation `json:"last_operation,omitempty" yaml:"last_operation,omitempty"`
	Metadata      *Metadata                              `json:"metadata,omitempty"       yaml:"metadata,omitempty"`
	Relationships ServiceCredentialBindingRelationships  `json:"relationships"            yaml:"relationships"`
	Links         Links                                  `json:"links,omitempty"          yaml:"links,omitempty"`
}

// ServiceCredentialBindingLastOperation represents the last operation for a service credential binding.
type ServiceCredentialBindingLastOperation struct {
	Type        string     `json:"type"                  yaml:"type"` // "create", "update", "delete"
	State       string     `json:"state"                 yaml:"state"`
	Description *string    `json:"description,omitempty" yaml:"description,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"  yaml:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"  yaml:"updated_at,omitempty"`
}

// ServiceCredentialBindingRelationships represents relationships for a service credential binding.
type ServiceCredentialBindingRelationships struct {
	App             *Relationship `json:"app,omitempty"    yaml:"app,omitempty"` // Only for type="app"
	ServiceInstance Relationship  `json:"service_instance" yaml:"service_instance"`
}

// ServiceCredentialBindingCreateRequest represents a request to create a service credential binding.
type ServiceCredentialBindingCreateRequest struct {
	// Type chooses binding type: "app" or "key".
	Type string `json:"type" yaml:"type"`
	// Name optionally names the binding (keys commonly use this).
	Name *string `json:"name,omitempty" yaml:"name,omitempty"`
	// Parameters are broker-specific binding parameters.
	Parameters map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	// Metadata sets labels/annotations on the binding.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	// Relationships must include ServiceInstance; App is required for type="app".
	Relationships ServiceCredentialBindingRelationships `json:"relationships" yaml:"relationships"`
}

// ServiceCredentialBindingUpdateRequest represents a request to update a service credential binding.
type ServiceCredentialBindingUpdateRequest struct {
	// Metadata updates labels/annotations; nil leaves it unchanged.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// ServiceCredentialBindingDetails represents the details of a service credential binding.
type ServiceCredentialBindingDetails struct {
	Credentials    map[string]interface{} `json:"credentials"                yaml:"credentials"`
	SyslogDrainURL *string                `json:"syslog_drain_url,omitempty" yaml:"syslog_drain_url,omitempty"`
	VolumeMounts   []interface{}          `json:"volume_mounts,omitempty"    yaml:"volume_mounts,omitempty"`
}

// ServiceCredentialBindingParameters represents the parameters of a service credential binding.
type ServiceCredentialBindingParameters struct {
	Parameters map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

// ServiceBinding is an alias for ServiceCredentialBinding for backward compatibility.
type ServiceBinding = ServiceCredentialBinding

// ServiceRouteBinding represents a service route binding.
type ServiceRouteBinding struct {
	Resource

	RouteServiceURL *string                           `json:"route_service_url,omitempty" yaml:"route_service_url,omitempty"`
	LastOperation   *ServiceRouteBindingLastOperation `json:"last_operation,omitempty"    yaml:"last_operation,omitempty"`
	Metadata        *Metadata                         `json:"metadata,omitempty"          yaml:"metadata,omitempty"`
	Relationships   ServiceRouteBindingRelationships  `json:"relationships"               yaml:"relationships"`
	Links           Links                             `json:"links,omitempty"             yaml:"links,omitempty"`
}

// ServiceRouteBindingLastOperation represents the last operation for a service route binding.
type ServiceRouteBindingLastOperation struct {
	Type        string     `json:"type"                  yaml:"type"`  // "create", "update", "delete"
	State       string     `json:"state"                 yaml:"state"` // "initial", "in_progress", "succeeded", "failed"
	Description *string    `json:"description,omitempty" yaml:"description,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"  yaml:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"  yaml:"updated_at,omitempty"`
}

// ServiceRouteBindingRelationships represents the relationships for a service route binding.
type ServiceRouteBindingRelationships struct {
	ServiceInstance Relationship `json:"service_instance" yaml:"service_instance"`
	Route           Relationship `json:"route"            yaml:"route"`
}

// ServiceRouteBindingCreateRequest represents a request to create a service route binding.
type ServiceRouteBindingCreateRequest struct {
	// Parameters are broker-specific route binding parameters.
	Parameters map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	// Metadata sets labels/annotations on the route binding.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	// Relationships must include ServiceInstance and Route.
	Relationships ServiceRouteBindingRelationships `json:"relationships" yaml:"relationships"`
}

// ServiceRouteBindingUpdateRequest represents a request to update a service route binding.
type ServiceRouteBindingUpdateRequest struct {
	// Metadata updates labels/annotations; nil leaves it unchanged.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// ServiceRouteBindingParameters represents parameters for a service route binding.
type ServiceRouteBindingParameters struct {
	Parameters map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

// LogMessage represents a single log message.
type LogMessage struct {
	Message     string    `json:"message"      yaml:"message"`
	MessageType string    `json:"message_type" yaml:"message_type"`
	Timestamp   time.Time `json:"timestamp"    yaml:"timestamp"`
	AppID       string    `json:"app_id"       yaml:"app_id"`
	SourceType  string    `json:"source_type"  yaml:"source_type"`
	SourceID    string    `json:"source_id"    yaml:"source_id"`
}

// AppLogs represents a collection of log messages for an app.
type AppLogs struct {
	Messages []LogMessage `json:"messages" yaml:"messages"`
}

// LogCacheEnvelope represents a log cache response envelope.
type LogCacheEnvelope struct {
	Timestamp  string               `json:"timestamp"     yaml:"timestamp"`
	SourceID   string               `json:"source_id"     yaml:"source_id"`
	InstanceID string               `json:"instance_id"   yaml:"instance_id"`
	Tags       map[string]string    `json:"tags"          yaml:"tags"`
	Log        *LogCacheLogEnvelope `json:"log,omitempty" yaml:"log,omitempty"`
}

// LogCacheLogEnvelope represents the log content within a log cache envelope.
type LogCacheLogEnvelope struct {
	Payload []byte `json:"payload" yaml:"payload"`
	Type    string `json:"type"    yaml:"type"`
}

// LogCacheResponse represents the response from log cache API.
type LogCacheResponse struct {
	Envelopes LogCacheEnvelopesWrapper `json:"envelopes" yaml:"envelopes"`
}

// LogCacheEnvelopesWrapper wraps the batch array in the log cache response.
type LogCacheEnvelopesWrapper struct {
	Batch []LogCacheEnvelope `json:"batch" yaml:"batch"`
}

// OrganizationQuota represents an organization quota.
type OrganizationQuota struct {
	Resource

	Name          string                          `json:"name"                    yaml:"name"`
	Apps          *OrganizationQuotaApps          `json:"apps,omitempty"          yaml:"apps,omitempty"`
	Services      *OrganizationQuotaServices      `json:"services,omitempty"      yaml:"services,omitempty"`
	Routes        *OrganizationQuotaRoutes        `json:"routes,omitempty"        yaml:"routes,omitempty"`
	Domains       *OrganizationQuotaDomains       `json:"domains,omitempty"       yaml:"domains,omitempty"`
	Relationships *OrganizationQuotaRelationships `json:"relationships,omitempty" yaml:"relationships,omitempty"`
	Metadata      *Metadata                       `json:"metadata,omitempty"      yaml:"metadata,omitempty"`
}

// OrganizationQuotaApps represents app limits in an organization quota.
type OrganizationQuotaApps struct {
	TotalMemoryInMB              *int `json:"total_memory_in_mb,omitempty"                 yaml:"total_memory_in_mb,omitempty"`
	TotalInstanceMemoryInMB      *int `json:"total_instance_memory_in_mb,omitempty"        yaml:"total_instance_memory_in_mb,omitempty"`
	LogRateLimitInBytesPerSecond *int `json:"log_rate_limit_in_bytes_per_second,omitempty" yaml:"log_rate_limit_in_bytes_per_second,omitempty"`
	TotalInstances               *int `json:"total_instances,omitempty"                    yaml:"total_instances,omitempty"`
	TotalAppTasks                *int `json:"total_app_tasks,omitempty"                    yaml:"total_app_tasks,omitempty"`
}

// OrganizationQuotaServices represents service limits in an organization quota.
type OrganizationQuotaServices struct {
	PaidServicesAllowed   *bool `json:"paid_services_allowed,omitempty"   yaml:"paid_services_allowed,omitempty"`
	TotalServiceInstances *int  `json:"total_service_instances,omitempty" yaml:"total_service_instances,omitempty"`
	TotalServiceKeys      *int  `json:"total_service_keys,omitempty"      yaml:"total_service_keys,omitempty"`
}

// OrganizationQuotaRoutes represents route limits in an organization quota.
type OrganizationQuotaRoutes struct {
	TotalRoutes        *int `json:"total_routes,omitempty"         yaml:"total_routes,omitempty"`
	TotalReservedPorts *int `json:"total_reserved_ports,omitempty" yaml:"total_reserved_ports,omitempty"`
}

// OrganizationQuotaDomains represents domain limits in an organization quota.
type OrganizationQuotaDomains struct {
	TotalDomains *int `json:"total_domains,omitempty" yaml:"total_domains,omitempty"`
}

// OrganizationQuotaRelationships represents organization quota relationships.
type OrganizationQuotaRelationships struct {
	Organizations ToManyRelationship `json:"organizations" yaml:"organizations"`
}

// OrganizationQuotaCreateRequest represents a request to create an organization quota.
type OrganizationQuotaCreateRequest struct {
	// Name is the quota name.
	Name string `json:"name" yaml:"name"`
	// Apps/Services/Routes/Domains set resource limits; omit to use defaults.
	Apps     *OrganizationQuotaApps     `json:"apps,omitempty"     yaml:"apps,omitempty"`
	Services *OrganizationQuotaServices `json:"services,omitempty" yaml:"services,omitempty"`
	Routes   *OrganizationQuotaRoutes   `json:"routes,omitempty"   yaml:"routes,omitempty"`
	Domains  *OrganizationQuotaDomains  `json:"domains,omitempty"  yaml:"domains,omitempty"`
	// Metadata sets labels/annotations on the quota.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// OrganizationQuotaUpdateRequest represents a request to update an organization quota.
type OrganizationQuotaUpdateRequest struct {
	// Name updates the quota name; nil leaves it unchanged.
	Name *string `json:"name,omitempty" yaml:"name,omitempty"`
	// Apps/Services/Routes/Domains update resource limits.
	Apps     *OrganizationQuotaApps     `json:"apps,omitempty"     yaml:"apps,omitempty"`
	Services *OrganizationQuotaServices `json:"services,omitempty" yaml:"services,omitempty"`
	Routes   *OrganizationQuotaRoutes   `json:"routes,omitempty"   yaml:"routes,omitempty"`
	Domains  *OrganizationQuotaDomains  `json:"domains,omitempty"  yaml:"domains,omitempty"`
	// Metadata updates labels/annotations; nil leaves it unchanged.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// SpaceQuotaV3 represents a space quota (v3 API).
type SpaceQuotaV3 struct {
	Resource

	Name          string                   `json:"name"                    yaml:"name"`
	Apps          *SpaceQuotaApps          `json:"apps,omitempty"          yaml:"apps,omitempty"`
	Services      *SpaceQuotaServices      `json:"services,omitempty"      yaml:"services,omitempty"`
	Routes        *SpaceQuotaRoutes        `json:"routes,omitempty"        yaml:"routes,omitempty"`
	Relationships *SpaceQuotaRelationships `json:"relationships,omitempty" yaml:"relationships,omitempty"`
	Metadata      *Metadata                `json:"metadata,omitempty"      yaml:"metadata,omitempty"`
}

// SpaceQuotaApps represents app limits in a space quota.
type SpaceQuotaApps struct {
	TotalMemoryInMB              *int `json:"total_memory_in_mb,omitempty"                 yaml:"total_memory_in_mb,omitempty"`
	TotalInstanceMemoryInMB      *int `json:"total_instance_memory_in_mb,omitempty"        yaml:"total_instance_memory_in_mb,omitempty"`
	LogRateLimitInBytesPerSecond *int `json:"log_rate_limit_in_bytes_per_second,omitempty" yaml:"log_rate_limit_in_bytes_per_second,omitempty"`
	TotalInstances               *int `json:"total_instances,omitempty"                    yaml:"total_instances,omitempty"`
	TotalAppTasks                *int `json:"total_app_tasks,omitempty"                    yaml:"total_app_tasks,omitempty"`
}

// SpaceQuotaServices represents service limits in a space quota.
type SpaceQuotaServices struct {
	PaidServicesAllowed   *bool `json:"paid_services_allowed,omitempty"   yaml:"paid_services_allowed,omitempty"`
	TotalServiceInstances *int  `json:"total_service_instances,omitempty" yaml:"total_service_instances,omitempty"`
	TotalServiceKeys      *int  `json:"total_service_keys,omitempty"      yaml:"total_service_keys,omitempty"`
}

// SpaceQuotaRoutes represents route limits in a space quota.
type SpaceQuotaRoutes struct {
	TotalRoutes        *int `json:"total_routes,omitempty"         yaml:"total_routes,omitempty"`
	TotalReservedPorts *int `json:"total_reserved_ports,omitempty" yaml:"total_reserved_ports,omitempty"`
}

// SpaceQuotaRelationships represents space quota relationships.
type SpaceQuotaRelationships struct {
	Organization Relationship       `json:"organization" yaml:"organization"`
	Spaces       ToManyRelationship `json:"spaces"       yaml:"spaces"`
}

// SpaceQuotaV3CreateRequest represents a request to create a space quota.
type SpaceQuotaV3CreateRequest struct {
	// Name is the space quota name.
	Name string `json:"name" yaml:"name"`
	// Apps/Services/Routes set resource limits; omit to use defaults.
	Apps     *SpaceQuotaApps     `json:"apps,omitempty"     yaml:"apps,omitempty"`
	Services *SpaceQuotaServices `json:"services,omitempty" yaml:"services,omitempty"`
	Routes   *SpaceQuotaRoutes   `json:"routes,omitempty"   yaml:"routes,omitempty"`
	// Relationships must include Organization and target Spaces.
	Relationships SpaceQuotaRelationships `json:"relationships" yaml:"relationships"`
	// Metadata sets labels/annotations on the space quota.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// SpaceQuotaV3UpdateRequest represents a request to update a space quota.
type SpaceQuotaV3UpdateRequest struct {
	// Name updates the space quota name; nil leaves it unchanged.
	Name *string `json:"name,omitempty" yaml:"name,omitempty"`
	// Apps/Services/Routes update resource limits.
	Apps     *SpaceQuotaApps     `json:"apps,omitempty"     yaml:"apps,omitempty"`
	Services *SpaceQuotaServices `json:"services,omitempty" yaml:"services,omitempty"`
	Routes   *SpaceQuotaRoutes   `json:"routes,omitempty"   yaml:"routes,omitempty"`
	// Metadata updates labels/annotations; nil leaves it unchanged.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// Sidecar represents a sidecar.
type Sidecar struct {
	Resource

	Name          string               `json:"name"                   yaml:"name"`
	Command       string               `json:"command"                yaml:"command"`
	ProcessTypes  []string             `json:"process_types"          yaml:"process_types"`
	MemoryInMB    *int                 `json:"memory_in_mb,omitempty" yaml:"memory_in_mb,omitempty"`
	Origin        string               `json:"origin"                 yaml:"origin"`
	Relationships SidecarRelationships `json:"relationships"          yaml:"relationships"`
}

// SidecarRelationships represents sidecar relationships.
type SidecarRelationships struct {
	App Relationship `json:"app" yaml:"app"`
}

// SidecarCreateRequest represents a request to create a sidecar.
type SidecarCreateRequest struct {
	// Name and Command define the sidecar; ProcessTypes lists bound process types.
	Name         string   `json:"name"          yaml:"name"`
	Command      string   `json:"command"       yaml:"command"`
	ProcessTypes []string `json:"process_types" yaml:"process_types"`
	// MemoryInMB optionally sets sidecar memory.
	MemoryInMB *int `json:"memory_in_mb,omitempty" yaml:"memory_in_mb,omitempty"`
}

// SidecarUpdateRequest represents a request to update a sidecar.
type SidecarUpdateRequest struct {
	// Fields update sidecar configuration; nil/empty leaves unchanged.
	Name         *string  `json:"name,omitempty"          yaml:"name,omitempty"`
	Command      *string  `json:"command,omitempty"       yaml:"command,omitempty"`
	ProcessTypes []string `json:"process_types,omitempty" yaml:"process_types,omitempty"`
	MemoryInMB   *int     `json:"memory_in_mb,omitempty"  yaml:"memory_in_mb,omitempty"`
}

// Revision represents a revision.
type Revision struct {
	Resource

	Version       int                   `json:"version"               yaml:"version"`
	Droplet       RevisionDropletRef    `json:"droplet"               yaml:"droplet"`
	Processes     map[string]Process    `json:"processes"             yaml:"processes"`
	Sidecars      []Sidecar             `json:"sidecars"              yaml:"sidecars"`
	Relationships RevisionRelationships `json:"relationships"         yaml:"relationships"`
	Metadata      *Metadata             `json:"metadata,omitempty"    yaml:"metadata,omitempty"`
	Description   *string               `json:"description,omitempty" yaml:"description,omitempty"`
	Deployable    bool                  `json:"deployable"            yaml:"deployable"`
}

// RevisionDropletRef represents a droplet reference in a revision.
type RevisionDropletRef struct {
	GUID string `json:"guid" yaml:"guid"`
}

// RevisionRelationships represents revision relationships.
type RevisionRelationships struct {
	App Relationship `json:"app" yaml:"app"`
}

// RevisionUpdateRequest represents a request to update a revision.
type RevisionUpdateRequest struct {
	// Metadata updates labels/annotations; nil leaves it unchanged.
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// EnvironmentVariableGroup represents an environment variable group.
type EnvironmentVariableGroup struct {
	Name      string                 `json:"name"                 yaml:"name"`
	Var       map[string]interface{} `json:"var"                  yaml:"var"`
	UpdatedAt *time.Time             `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
	Links     Links                  `json:"links,omitempty"      yaml:"links,omitempty"`
}

// AppUsageEvent represents an app usage event.
type AppUsageEvent struct {
	Resource

	State                         string               `json:"state"                                        yaml:"state"`
	PreviousState                 *string              `json:"previous_state,omitempty"                     yaml:"previous_state,omitempty"`
	MemoryInMBPerInstance         int                  `json:"memory_in_mb_per_instance"                    yaml:"memory_in_mb_per_instance"`
	PreviousMemoryInMBPerInstance *int                 `json:"previous_memory_in_mb_per_instance,omitempty" yaml:"previous_memory_in_mb_per_instance,omitempty"`
	InstanceCount                 int                  `json:"instance_count"                               yaml:"instance_count"`
	PreviousInstanceCount         *int                 `json:"previous_instance_count,omitempty"            yaml:"previous_instance_count,omitempty"`
	AppName                       string               `json:"app_name"                                     yaml:"app_name"`
	AppGUID                       string               `json:"app_guid"                                     yaml:"app_guid"`
	SpaceName                     string               `json:"space_name"                                   yaml:"space_name"`
	SpaceGUID                     string               `json:"space_guid"                                   yaml:"space_guid"`
	OrganizationName              string               `json:"organization_name"                            yaml:"organization_name"`
	OrganizationGUID              string               `json:"organization_guid"                            yaml:"organization_guid"`
	BuildpackName                 *string              `json:"buildpack_name,omitempty"                     yaml:"buildpack_name,omitempty"`
	BuildpackGUID                 *string              `json:"buildpack_guid,omitempty"                     yaml:"buildpack_guid,omitempty"`
	Package                       AppUsageEventPackage `json:"package"                                      yaml:"package"`
	ParentAppName                 *string              `json:"parent_app_name,omitempty"                    yaml:"parent_app_name,omitempty"`
	ParentAppGUID                 *string              `json:"parent_app_guid,omitempty"                    yaml:"parent_app_guid,omitempty"`
	ProcessType                   string               `json:"process_type"                                 yaml:"process_type"`
	TaskName                      *string              `json:"task_name,omitempty"                          yaml:"task_name,omitempty"`
	TaskGUID                      *string              `json:"task_guid,omitempty"                          yaml:"task_guid,omitempty"`
}

// AppUsageEventPackage represents package information in an app usage event.
type AppUsageEventPackage struct {
	State string `json:"state" yaml:"state"`
}

// ServiceUsageEvent represents a service usage event.
type ServiceUsageEvent struct {
	Resource

	State               string  `json:"state"                    yaml:"state"`
	PreviousState       *string `json:"previous_state,omitempty" yaml:"previous_state,omitempty"`
	ServiceInstanceName string  `json:"service_instance_name"    yaml:"service_instance_name"`
	ServiceInstanceGUID string  `json:"service_instance_guid"    yaml:"service_instance_guid"`
	ServiceInstanceType string  `json:"service_instance_type"    yaml:"service_instance_type"`
	ServicePlanName     string  `json:"service_plan_name"        yaml:"service_plan_name"`
	ServicePlanGUID     string  `json:"service_plan_guid"        yaml:"service_plan_guid"`
	ServiceOfferingName string  `json:"service_offering_name"    yaml:"service_offering_name"`
	ServiceOfferingGUID string  `json:"service_offering_guid"    yaml:"service_offering_guid"`
	ServiceBrokerName   string  `json:"service_broker_name"      yaml:"service_broker_name"`
	ServiceBrokerGUID   string  `json:"service_broker_guid"      yaml:"service_broker_guid"`
	SpaceName           string  `json:"space_name"               yaml:"space_name"`
	SpaceGUID           string  `json:"space_guid"               yaml:"space_guid"`
	OrganizationName    string  `json:"organization_name"        yaml:"organization_name"`
	OrganizationGUID    string  `json:"organization_guid"        yaml:"organization_guid"`
}

// AuditEvent represents an audit event.
type AuditEvent struct {
	Resource

	Type         string                  `json:"type"                   yaml:"type"`
	Actor        AuditEventActor         `json:"actor"                  yaml:"actor"`
	Target       AuditEventTarget        `json:"target"                 yaml:"target"`
	Data         map[string]interface{}  `json:"data"                   yaml:"data"`
	Space        *AuditEventSpace        `json:"space,omitempty"        yaml:"space,omitempty"`
	Organization *AuditEventOrganization `json:"organization,omitempty" yaml:"organization,omitempty"`
}

// AuditEventActor represents the actor in an audit event.
type AuditEventActor struct {
	GUID string `json:"guid" yaml:"guid"`
	Type string `json:"type" yaml:"type"`
	Name string `json:"name" yaml:"name"`
}

// AuditEventTarget represents the target in an audit event.
type AuditEventTarget struct {
	GUID string `json:"guid" yaml:"guid"`
	Type string `json:"type" yaml:"type"`
	Name string `json:"name" yaml:"name"`
}

// AuditEventSpace represents space information in an audit event.
type AuditEventSpace struct {
	GUID string `json:"guid" yaml:"guid"`
	Name string `json:"name" yaml:"name"`
}

// AuditEventOrganization represents organization information in an audit event.
type AuditEventOrganization struct {
	GUID string `json:"guid" yaml:"guid"`
	Name string `json:"name" yaml:"name"`
}

// ResourceMatches represents resource matches.
type ResourceMatches struct {
	Resources []ResourceMatch `json:"resources" yaml:"resources"`
}

// ResourceMatch represents a single resource match.
type ResourceMatch struct {
	SHA1 string `json:"sha1" yaml:"sha1"`
	Size int64  `json:"size" yaml:"size"`
	Path string `json:"path" yaml:"path"`
	Mode string `json:"mode" yaml:"mode"`
}

// ResourceMatchesRequest represents a request to create resource matches.
type ResourceMatchesRequest struct {
	// Resources lists files (sha1, size, path, mode) to check for blob reuse.
	Resources []ResourceMatch `json:"resources" yaml:"resources"`
}
