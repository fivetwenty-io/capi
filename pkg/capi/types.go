package capi

import (
	"time"
)

// Resource represents the base structure for all CF API resources.
type Resource struct {
	GUID      string    `json:"guid"       yaml:"guid"`
	CreatedAt time.Time `json:"created_at" yaml:"created_at"`
	UpdatedAt time.Time `json:"updated_at" yaml:"updated_at"`
	Links     Links     `json:"links"      yaml:"links"`
}

// Links represents resource links.
type Links map[string]Link

// Link represents a single link.
type Link struct {
	Href   string `json:"href"             yaml:"href"`
	Method string `json:"method,omitempty" yaml:"method,omitempty"`
}

// Metadata represents labels and annotations.
type Metadata struct {
	Labels      map[string]string `json:"labels,omitempty"      yaml:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

// Relationship represents a to-one relationship.
type Relationship struct {
	Data *RelationshipData `json:"data,omitempty" yaml:"data,omitempty"`
}

// RelationshipData contains the GUID of the related resource.
type RelationshipData struct {
	GUID string `json:"guid" yaml:"guid"`
}

// ToManyRelationship represents a to-many relationship.
type ToManyRelationship struct {
	Data []RelationshipData `json:"data" yaml:"data"`
}

// Pagination represents pagination information.
type Pagination struct {
	TotalResults int   `json:"total_results"      yaml:"total_results"`
	TotalPages   int   `json:"total_pages"        yaml:"total_pages"`
	First        Link  `json:"first"              yaml:"first"`
	Last         Link  `json:"last"               yaml:"last"`
	Next         *Link `json:"next,omitempty"     yaml:"next,omitempty"`
	Previous     *Link `json:"previous,omitempty" yaml:"previous,omitempty"`
}

// ListResponse represents a paginated list response.
type ListResponse[T any] struct {
	Pagination Pagination `json:"pagination" yaml:"pagination"`
	Resources  []T        `json:"resources"  yaml:"resources"`
}

// AppEnv is an alias for AppEnvironment to maintain backward compatibility.
type AppEnv = AppEnvironment

// ProcessList represents a paginated list of Process resources.
type ProcessList = ListResponse[Process]

// AuditEventsList represents a paginated list of AuditEvent resources.
type AuditEventsList = ListResponse[AuditEvent]

// ProcessStat is an alias for ProcessStatsDetail to maintain backward compatibility.
type ProcessStat = ProcessStatsDetail

// InstancePort is an alias for ProcessInstancePort to maintain backward compatibility.
type InstancePort = ProcessInstancePort

// Actor is an alias for AuditEventActor to maintain backward compatibility.
type Actor = AuditEventActor

// Target is an alias for AuditEventTarget to maintain backward compatibility.
type Target = AuditEventTarget

// BuildpacksList represents a paginated list of Buildpack resources.
type BuildpacksList = ListResponse[Buildpack]

// DomainsList represents a paginated list of Domain resources.
type DomainsList = ListResponse[Domain]

// Include represents include parameters for API requests.
type Include []string

// Fields represents field selection parameters.
type Fields map[string][]string
