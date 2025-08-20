package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewVersionCommand creates the version command
func NewVersionCommand(version, commit, date string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Display version information",
		Long:  "Display detailed version information about the CAPI CLI",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("CAPI CLI\n")
			fmt.Printf("  Version:  %s\n", version)
			fmt.Printf("  Commit:   %s\n", commit)
			fmt.Printf("  Built:    %s\n", date)
		},
	}
}
