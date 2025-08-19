// Package cfclient provides the main entry point for creating Cloud Foundry API clients
package cfclient

import (
	"github.com/fivetwenty-io/capi-client-go/internal/client"
	"github.com/fivetwenty-io/capi-client-go/pkg/capi"
)

// New creates a new Cloud Foundry API client
func New(config *capi.Config) (capi.Client, error) {
	return client.New(config)
}

// NewWithEndpoint creates a new client with just an API endpoint (no auth)
func NewWithEndpoint(endpoint string) (capi.Client, error) {
	return New(&capi.Config{
		APIEndpoint: endpoint,
	})
}

// NewWithToken creates a new client with an API endpoint and access token
func NewWithToken(endpoint, token string) (capi.Client, error) {
	return New(&capi.Config{
		APIEndpoint: endpoint,
		AccessToken: token,
	})
}

// NewWithClientCredentials creates a new client using OAuth2 client credentials
func NewWithClientCredentials(endpoint, clientID, clientSecret string) (capi.Client, error) {
	return New(&capi.Config{
		APIEndpoint:  endpoint,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	})
}

// NewWithPassword creates a new client using username/password authentication
func NewWithPassword(endpoint, username, password string) (capi.Client, error) {
	return New(&capi.Config{
		APIEndpoint: endpoint,
		Username:    username,
		Password:    password,
	})
}
