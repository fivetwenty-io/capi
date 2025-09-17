package commands

import (
	"fmt"
	"sync"
	"time"

	"github.com/fivetwenty-io/capi/v3/internal/constants"
)

// ConfigPersister implements the auth.ConfigPersister interface.
type ConfigPersister struct {
	mutex sync.Mutex
}

// NewConfigPersister creates a new config persister.
func NewConfigPersister() *ConfigPersister {
	return &ConfigPersister{}
}

// UpdateAPIToken updates the API token and related metadata in the config.
func (p *ConfigPersister) UpdateAPIToken(apiDomain, token string, expiresAt time.Time, refreshToken string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Load current config
	config := loadConfig()

	// Find or create the API config
	if config.APIs == nil {
		config.APIs = make(map[string]*APIConfig)
	}

	apiConfig, exists := config.APIs[apiDomain]
	if !exists {
		return fmt.Errorf("API configuration for '%s': %w", apiDomain, constants.ErrAPIConfigNotFound)
	}

	// Update token information
	apiConfig.Token = token
	if !expiresAt.IsZero() {
		apiConfig.TokenExpiresAt = &expiresAt
	}

	if refreshToken != "" {
		apiConfig.RefreshToken = refreshToken
	}

	now := time.Now()
	apiConfig.LastRefreshed = &now

	// Save the updated config
	return saveConfigStruct(config)
}
