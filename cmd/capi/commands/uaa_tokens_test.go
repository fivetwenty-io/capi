//nolint:testpackage // Need access to internal types
package commands

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func TestCreateUsersGetPasswordTokenCommand(t *testing.T) {
	t.Parallel()

	cmd := createUsersGetPasswordTokenCommand()
	assert.Equal(t, "get-password-token", cmd.Use)
	assert.Equal(t, "Obtain access token using password grant", cmd.Short)
	assert.NotNil(t, cmd.RunE)
	assert.Contains(t, cmd.Long, "password grant")

	// Check flags
	assert.NotNil(t, cmd.Flags().Lookup("username"))
	assert.NotNil(t, cmd.Flags().Lookup("password"))
	assert.NotNil(t, cmd.Flags().Lookup("client-id"))
	assert.NotNil(t, cmd.Flags().Lookup("client-secret"))
}

func TestCreateUsersGetClientCredentialsTokenCommand(t *testing.T) {
	t.Parallel()

	cmd := createUsersGetClientCredentialsTokenCommand()
	assert.Equal(t, "get-client-credentials-token", cmd.Use)
	assert.Equal(t, "Obtain access token using client credentials grant", cmd.Short)
	assert.NotNil(t, cmd.RunE)
	assert.Contains(t, cmd.Long, "client_credentials")

	// Check flags
	assert.NotNil(t, cmd.Flags().Lookup("client-id"))
	assert.NotNil(t, cmd.Flags().Lookup("client-secret"))
	assert.NotNil(t, cmd.Flags().Lookup("token-format"))
}

func TestCreateUsersGetAuthcodeTokenCommand(t *testing.T) {
	t.Parallel()

	cmd := createUsersGetAuthcodeTokenCommand()
	assert.Equal(t, "get-authcode-token", cmd.Use)
	assert.Equal(t, "Obtain access token using authorization code grant", cmd.Short)
	assert.NotNil(t, cmd.RunE)
	assert.Contains(t, cmd.Long, "authorization code")

	// Check flags
	assert.NotNil(t, cmd.Flags().Lookup("client-id"))
	assert.NotNil(t, cmd.Flags().Lookup("client-secret"))
	assert.NotNil(t, cmd.Flags().Lookup("auth-code"))
	assert.NotNil(t, cmd.Flags().Lookup("redirect-uri"))
}

func TestCreateUsersRefreshTokenCommand(t *testing.T) {
	t.Parallel()

	cmd := createUsersRefreshTokenCommand()
	assert.Equal(t, "refresh-token", cmd.Use)
	assert.Equal(t, "Refresh access token", cmd.Short)
	assert.NotNil(t, cmd.RunE)
	assert.Contains(t, cmd.Long, "refresh")
}

func TestCreateUsersGetTokenKeyCommand(t *testing.T) {
	t.Parallel()

	cmd := createUsersGetTokenKeyCommand()
	assert.Equal(t, "get-token-key", cmd.Use)
	assert.Equal(t, "View JWT signing key", cmd.Short)
	assert.NotNil(t, cmd.RunE)
	assert.Contains(t, cmd.Long, "JWT")
}

func TestCreateUsersGetTokenKeysCommand(t *testing.T) {
	t.Parallel()

	cmd := createUsersGetTokenKeysCommand()
	assert.Equal(t, "get-token-keys", cmd.Use)
	assert.Equal(t, "View all JWT signing keys", cmd.Short)
	assert.NotNil(t, cmd.RunE)
	assert.Contains(t, cmd.Long, "JWT")
}

func TestCreateUsersGetImplicitTokenCommand(t *testing.T) {
	t.Parallel()

	cmd := createUsersGetImplicitTokenCommand()
	assert.Equal(t, "get-implicit-token", cmd.Use)
	assert.Equal(t, "Obtain access token using implicit grant", cmd.Short)
	assert.NotNil(t, cmd.RunE)
	assert.Contains(t, cmd.Long, "implicit grant")

	// No flags for implicit grant as it requires manual implementation
}

func TestDisplayTokenInfo(t *testing.T) {
	t.Parallel()

	token := &oauth2.Token{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "bearer",
		Expiry:       time.Now().Add(time.Hour),
	}

	// Test that function doesn't error
	err := displayTokenInfo(token, "Test Grant")
	assert.NoError(t, err)
}
