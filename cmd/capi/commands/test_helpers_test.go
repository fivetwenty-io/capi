package commands_test

import "github.com/spf13/cobra"

// findSubcommand finds a subcommand by name within a cobra command.
func findSubcommand(cmd *cobra.Command, name string) *cobra.Command {
	for _, c := range cmd.Commands() {
		if c.Name() == name {
			return c
		}
	}

	return nil
}
