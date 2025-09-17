package commands

import (
	"github.com/fivetwenty-io/capi/v3/internal/constants"
	"github.com/spf13/cobra"
)

// createUsersGetImplicitTokenCommand creates the implicit token command.
func createUsersGetImplicitTokenCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get-implicit-token",
		Short: "Obtain access token using implicit grant",
		Long: `Obtain an access token using the OAuth2 implicit grant type.

NOTE: The implicit grant flow is not directly supported by the go-uaa client library.
For implicit grants, you would typically:

1. Direct users to the UAA authorization endpoint with response_type=token
2. Extract the access token from the URL fragment after redirect
3. Use 'capi uaa set-token' to store the obtained token

Example authorization URL format:
https://uaa.example.com/oauth/authorize?response_type=token&client_id=CLIENT_ID&redirect_uri=REDIRECT_URI

This command is provided for completeness but requires manual token extraction.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return constants.ErrImplicitFlowManual
		},
	}
}
