package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOAuth2TokenManager_GetToken(t *testing.T) {
	t.Run("returns existing valid token", func(t *testing.T) {
		manager := NewOAuth2TokenManager(&OAuth2Config{
			AccessToken: "existing-token",
		})

		token, err := manager.GetToken(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "existing-token", token)
	})

	t.Run("refreshes expired token using refresh token", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/oauth/token", r.URL.Path)
			assert.Equal(t, "POST", r.Method)

			err := r.ParseForm()
			require.NoError(t, err)
			assert.Equal(t, "refresh_token", r.Form.Get("grant_type"))
			assert.Equal(t, "old-refresh-token", r.Form.Get("refresh_token"))

			response := Token{
				AccessToken:  "new-access-token",
				RefreshToken: "new-refresh-token",
				ExpiresIn:    3600,
				TokenType:    "bearer",
			}
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		manager := NewOAuth2TokenManager(&OAuth2Config{
			TokenURL:     server.URL + "/oauth/token",
			RefreshToken: "old-refresh-token",
		})

		// Set expired token
		manager.store.Set(&Token{
			AccessToken:  "expired-token",
			RefreshToken: "old-refresh-token",
			ExpiresAt:    time.Now().Add(-1 * time.Hour),
		})

		token, err := manager.GetToken(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "new-access-token", token)
	})

	t.Run("uses client credentials when no refresh token", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/oauth/token", r.URL.Path)
			assert.Equal(t, "POST", r.Method)

			// Check basic auth
			username, password, ok := r.BasicAuth()
			assert.True(t, ok)
			assert.Equal(t, "client-id", username)
			assert.Equal(t, "client-secret", password)

			err := r.ParseForm()
			require.NoError(t, err)
			assert.Equal(t, "client_credentials", r.Form.Get("grant_type"))

			response := Token{
				AccessToken: "client-token",
				ExpiresIn:   3600,
				TokenType:   "bearer",
			}
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		manager := NewOAuth2TokenManager(&OAuth2Config{
			TokenURL:     server.URL + "/oauth/token",
			ClientID:     "client-id",
			ClientSecret: "client-secret",
		})

		token, err := manager.GetToken(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "client-token", token)
	})

	t.Run("uses password grant", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/oauth/token", r.URL.Path)
			assert.Equal(t, "POST", r.Method)

			err := r.ParseForm()
			require.NoError(t, err)
			assert.Equal(t, "password", r.Form.Get("grant_type"))
			assert.Equal(t, "testuser", r.Form.Get("username"))
			assert.Equal(t, "testpass", r.Form.Get("password"))

			response := Token{
				AccessToken:  "password-token",
				RefreshToken: "refresh-token",
				ExpiresIn:    3600,
				TokenType:    "bearer",
			}
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		manager := NewOAuth2TokenManager(&OAuth2Config{
			TokenURL: server.URL + "/oauth/token",
			Username: "testuser",
			Password: "testpass",
		})

		token, err := manager.GetToken(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "password-token", token)
	})

	t.Run("handles token request error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			response := map[string]string{
				"error":             "invalid_client",
				"error_description": "Client authentication failed",
			}
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		manager := NewOAuth2TokenManager(&OAuth2Config{
			TokenURL:     server.URL + "/oauth/token",
			ClientID:     "bad-client",
			ClientSecret: "bad-secret",
		})

		token, err := manager.GetToken(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid_client")
		assert.Contains(t, err.Error(), "Client authentication failed")
		assert.Equal(t, "", token)
	})

	t.Run("no credentials available", func(t *testing.T) {
		manager := NewOAuth2TokenManager(&OAuth2Config{
			TokenURL: "http://example.com/oauth/token",
		})

		token, err := manager.GetToken(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no valid credentials available")
		assert.Equal(t, "", token)
	})
}

func TestOAuth2TokenManager_SetToken(t *testing.T) {
	manager := NewOAuth2TokenManager(&OAuth2Config{})

	expiresAt := time.Now().Add(1 * time.Hour)
	manager.SetToken("manual-token", expiresAt)

	token, err := manager.GetToken(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "manual-token", token)

	storedToken := manager.store.Get()
	assert.Equal(t, "manual-token", storedToken.AccessToken)
	assert.Equal(t, "bearer", storedToken.TokenType)
	assert.Equal(t, expiresAt.Unix(), storedToken.ExpiresAt.Unix())
}

func TestOAuth2TokenManager_RefreshToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := Token{
			AccessToken: "refreshed-token",
			ExpiresIn:   3600,
			TokenType:   "bearer",
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	manager := NewOAuth2TokenManager(&OAuth2Config{
		TokenURL:     server.URL + "/oauth/token",
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	})

	// Set a valid token
	manager.SetToken("current-token", time.Now().Add(1*time.Hour))

	// Force refresh
	err := manager.RefreshToken(context.Background())
	require.NoError(t, err)

	// Should have new token
	token, err := manager.GetToken(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "refreshed-token", token)
}

func TestNewUAATokenManager(t *testing.T) {
	t.Run("creates manager with correct token URL", func(t *testing.T) {
		manager := NewUAATokenManager("https://uaa.example.com", "client-id", "client-secret")
		assert.NotNil(t, manager)
		assert.Equal(t, "https://uaa.example.com/oauth/token", manager.config.TokenURL)
		assert.Equal(t, "client-id", manager.config.ClientID)
		assert.Equal(t, "client-secret", manager.config.ClientSecret)
		assert.Contains(t, manager.config.Scopes, "cloud_controller.read")
		assert.Contains(t, manager.config.Scopes, "cloud_controller.write")
	})

	t.Run("handles trailing slash in UAA URL", func(t *testing.T) {
		manager := NewUAATokenManager("https://uaa.example.com/", "client-id", "client-secret")
		assert.Equal(t, "https://uaa.example.com/oauth/token", manager.config.TokenURL)
	})
}

func TestNewUAATokenManagerWithPassword(t *testing.T) {
	manager := NewUAATokenManagerWithPassword(
		"https://uaa.example.com",
		"client-id",
		"client-secret",
		"username",
		"password",
	)

	assert.NotNil(t, manager)
	assert.Equal(t, "https://uaa.example.com/oauth/token", manager.config.TokenURL)
	assert.Equal(t, "client-id", manager.config.ClientID)
	assert.Equal(t, "client-secret", manager.config.ClientSecret)
	assert.Equal(t, "username", manager.config.Username)
	assert.Equal(t, "password", manager.config.Password)
	assert.Contains(t, manager.config.Scopes, "cloud_controller.read")
	assert.Contains(t, manager.config.Scopes, "cloud_controller.write")
}
