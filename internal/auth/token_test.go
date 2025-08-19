package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToken_Valid(t *testing.T) {
	tests := []struct {
		name     string
		token    *Token
		expected bool
	}{
		{
			name:     "nil token",
			token:    nil,
			expected: false,
		},
		{
			name: "empty access token",
			token: &Token{
				AccessToken: "",
			},
			expected: false,
		},
		{
			name: "valid token without expiry",
			token: &Token{
				AccessToken: "test-token",
			},
			expected: true,
		},
		{
			name: "valid token with future expiry",
			token: &Token{
				AccessToken: "test-token",
				ExpiresAt:   time.Now().Add(1 * time.Hour),
			},
			expected: true,
		},
		{
			name: "expired token",
			token: &Token{
				AccessToken: "test-token",
				ExpiresAt:   time.Now().Add(-1 * time.Hour),
			},
			expected: false,
		},
		{
			name: "token expiring within buffer",
			token: &Token{
				AccessToken: "test-token",
				ExpiresAt:   time.Now().Add(15 * time.Second),
			},
			expected: false, // Should be false due to 30 second buffer
		},
		{
			name: "token expiring just outside buffer",
			token: &Token{
				AccessToken: "test-token",
				ExpiresAt:   time.Now().Add(35 * time.Second),
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.token.Valid())
		})
	}
}

func TestTokenStore(t *testing.T) {
	t.Run("new store is empty", func(t *testing.T) {
		store := NewTokenStore()
		assert.Nil(t, store.Get())
	})

	t.Run("set and get token", func(t *testing.T) {
		store := NewTokenStore()
		token := &Token{
			AccessToken: "test-token",
			TokenType:   "bearer",
		}

		store.Set(token)
		retrieved := store.Get()
		assert.NotNil(t, retrieved)
		assert.Equal(t, token.AccessToken, retrieved.AccessToken)
		assert.Equal(t, token.TokenType, retrieved.TokenType)
	})

	t.Run("clear token", func(t *testing.T) {
		store := NewTokenStore()
		token := &Token{
			AccessToken: "test-token",
		}

		store.Set(token)
		assert.NotNil(t, store.Get())

		store.Clear()
		assert.Nil(t, store.Get())
	})

	t.Run("concurrent access", func(t *testing.T) {
		store := NewTokenStore()
		done := make(chan bool)

		// Multiple goroutines setting tokens
		go func() {
			for i := 0; i < 100; i++ {
				store.Set(&Token{
					AccessToken: "token-1",
				})
			}
			done <- true
		}()

		go func() {
			for i := 0; i < 100; i++ {
				store.Set(&Token{
					AccessToken: "token-2",
				})
			}
			done <- true
		}()

		// Multiple goroutines getting tokens
		go func() {
			for i := 0; i < 100; i++ {
				_ = store.Get()
			}
			done <- true
		}()

		go func() {
			for i := 0; i < 100; i++ {
				_ = store.Get()
			}
			done <- true
		}()

		// Wait for all goroutines
		for i := 0; i < 4; i++ {
			<-done
		}

		// Should not panic and should have a token
		finalToken := store.Get()
		assert.NotNil(t, finalToken)
		assert.Contains(t, []string{"token-1", "token-2"}, finalToken.AccessToken)
	})
}
