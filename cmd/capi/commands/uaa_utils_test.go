//nolint:testpackage // Need access to internal types
package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateUsersCurlCommand(t *testing.T) {
	t.Parallel()

	cmd := createUsersCurlCommand()
	assert.Equal(t, "curl <path>", cmd.Use)
	assert.Equal(t, "Direct UAA API access", cmd.Short)
	assert.NotNil(t, cmd.RunE)
	assert.Contains(t, cmd.Long, "Make direct HTTP requests")

	// Check flags
	assert.NotNil(t, cmd.Flags().Lookup("method"))
	assert.NotNil(t, cmd.Flags().Lookup("data"))
	assert.NotNil(t, cmd.Flags().Lookup("header"))
	assert.NotNil(t, cmd.Flags().Lookup("output"))
}

func TestCreateUsersUserinfoCommand(t *testing.T) {
	t.Parallel()

	cmd := createUsersUserinfoCommand()
	assert.Equal(t, "userinfo", cmd.Use)
	assert.Equal(t, "Display current user claims", cmd.Short)
	assert.NotNil(t, cmd.RunE)
	assert.Contains(t, cmd.Long, "claims about the currently authenticated user")
}

// Note: Display functions would require more complex setup to test properly
// as they depend on actual UAA API responses and internal data structures.
// These tests focus on command structure validation instead.
