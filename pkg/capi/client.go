package capi

import (
	"context"
	"fmt"
	"time"
)

// Client is the main interface for interacting with the CF API
type Client interface {
	// Resource accessors
	Apps() AppsClient
	Organizations() OrganizationsClient
	Spaces() SpacesClient
	Domains() DomainsClient
	Routes() RoutesClient
	ServiceBrokers() ServiceBrokersClient
	ServiceOfferings() ServiceOfferingsClient
	ServicePlans() ServicePlansClient
	ServiceInstances() ServiceInstancesClient
	Builds() BuildsClient
	Buildpacks() BuildpacksClient
	Deployments() DeploymentsClient
	Droplets() DropletsClient
	Packages() PackagesClient
	Processes() ProcessesClient
	Tasks() TasksClient
	Stacks() StacksClient
	Users() UsersClient
	Roles() RolesClient
	SecurityGroups() SecurityGroupsClient
	IsolationSegments() IsolationSegmentsClient
	FeatureFlags() FeatureFlagsClient
	Jobs() JobsClient

	// Info endpoints
	GetInfo(ctx context.Context) (*Info, error)
	GetRootInfo(ctx context.Context) (*RootInfo, error)
	GetUsageSummary(ctx context.Context) (*UsageSummary, error)

	// Admin operations
	ClearBuildpackCache(ctx context.Context) (*Job, error)
}

// Logger interface for logging
type Logger interface {
	Debug(msg string, fields map[string]interface{})
	Info(msg string, fields map[string]interface{})
	Warn(msg string, fields map[string]interface{})
	Error(msg string, fields map[string]interface{})
}

// Config represents client configuration
type Config struct {
	// Required fields
	APIEndpoint string

	// Authentication options (provide one)
	ClientID     string
	ClientSecret string
	Username     string
	Password     string
	RefreshToken string
	AccessToken  string

	// Optional configurations
	HTTPTimeout   time.Duration
	RetryMax      int
	RetryWaitMin  time.Duration
	RetryWaitMax  time.Duration
	Debug         bool
	Logger        Logger
	SkipTLSVerify bool
	UserAgent     string
}

// NewClient creates a new CF API client
func NewClient(config *Config) (Client, error) {
	// We'll implement a factory in a separate package to avoid circular dependencies
	// For now, return an error indicating the client needs to be created using the factory
	return nil, fmt.Errorf("use github.com/fivetwenty-io/capi-client-go/pkg/cfclient.New to create a client")
}

// Info represents the /v3/info response
type Info struct {
	Build       string                 `json:"build"`
	CLIVersion  CLIVersion             `json:"cli_version"`
	Custom      map[string]interface{} `json:"custom"`
	Description string                 `json:"description"`
	Name        string                 `json:"name"`
	Version     int                    `json:"version"`
	Links       Links                  `json:"links"`
	CFOnK8s     bool                   `json:"cf_on_k8s"`
}

// CLIVersion represents CLI version information
type CLIVersion struct {
	Minimum     string `json:"minimum"`
	Recommended string `json:"recommended"`
}

// RootInfo represents the root / response
type RootInfo struct {
	Links Links `json:"links"`
}

// UsageSummary represents platform usage summary
type UsageSummary struct {
	UsageSummary UsageSummaryData `json:"usage_summary"`
	Links        Links            `json:"links"`
}

// UsageSummaryData contains the actual usage data
type UsageSummaryData struct {
	StartedInstances int `json:"started_instances"`
	MemoryInMB       int `json:"memory_in_mb"`
}

// Job represents an asynchronous job
type Job struct {
	Resource
	Operation string     `json:"operation"`
	State     string     `json:"state"`
	Errors    []APIError `json:"errors,omitempty"`
	Warnings  []Warning  `json:"warnings,omitempty"`
}

// Warning represents a warning in API responses
type Warning struct {
	Detail string `json:"detail"`
}
