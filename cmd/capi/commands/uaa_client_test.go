package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUAAClientWrapper_IsAuthenticated(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   bool
	}{
		{
			name: "with UAA token",
			config: &Config{
				UAAToken: "test-token",
			},
			want: true,
		},
		{
			name: "with CF token",
			config: &Config{
				Token: "test-token",
			},
			want: true,
		},
		{
			name: "with both tokens",
			config: &Config{
				UAAToken: "uaa-token",
				Token:    "cf-token",
			},
			want: true,
		},
		{
			name:   "no authentication",
			config: &Config{},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapper := &UAAClientWrapper{
				config: tt.config,
			}
			assert.Equal(t, tt.want, wrapper.IsAuthenticated())
		})
	}
}

func TestUAAClientWrapper_GetToken(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name: "UAA token takes precedence",
			config: &Config{
				UAAToken: "uaa-token",
				Token:    "cf-token",
			},
			expected: "uaa-token",
		},
		{
			name: "fallback to CF token",
			config: &Config{
				Token: "cf-token",
			},
			expected: "cf-token",
		},
		{
			name:     "no tokens",
			config:   &Config{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapper := &UAAClientWrapper{
				config: tt.config,
			}
			assert.Equal(t, tt.expected, wrapper.GetToken())
		})
	}
}

func TestUAAClientWrapper_SetToken(t *testing.T) {
	config := &Config{}
	wrapper := &UAAClientWrapper{
		config: config,
	}

	wrapper.SetToken("new-token")
	assert.Equal(t, "new-token", config.UAAToken)
}

func TestUAAClientWrapper_InferUAAEndpoint(t *testing.T) {
	tests := []struct {
		name        string
		cfAPIURL    string
		expectedURL string
	}{
		{
			name:        "standard CF API URL",
			cfAPIURL:    "https://api.cf.example.com",
			expectedURL: "https://uaa.cf.example.com",
		},
		{
			name:        "API with subdomain",
			cfAPIURL:    "https://api.system.cf.example.com",
			expectedURL: "https://uaa.system.cf.example.com",
		},
		{
			name:        "localhost development",
			cfAPIURL:    "https://api.bosh-lite.com",
			expectedURL: "https://uaa.bosh-lite.com",
		},
		{
			name:        "empty URL",
			cfAPIURL:    "",
			expectedURL: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapper := &UAAClientWrapper{
				config: &Config{
					API: tt.cfAPIURL,
				},
			}
			url := wrapper.inferUAAEndpoint()
			assert.Equal(t, tt.expectedURL, url)
		})
	}
}

func TestUAAClientWrapper_Endpoint(t *testing.T) {
	wrapper := &UAAClientWrapper{
		endpoint: "https://uaa.example.com",
	}
	assert.Equal(t, "https://uaa.example.com", wrapper.Endpoint())
}

func TestNewUAAClient_NoEndpoint(t *testing.T) {
	config := &Config{
		// No UAA endpoint or CF API endpoint
	}

	client, err := NewUAAClient(config)
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "no UAA endpoint configured")
}
