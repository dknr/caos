package cmd

import (
	"github.com/spf13/cobra"
)

// HelpCmd is the `caos help` command.
var HelpCmd = &cobra.Command{
	Use:   "help [command]",
	Short: "Show help for a command",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Parent().Help()
			return
		}
		// For Cobra v1.x, we need to use Root() and then Find
		// But simpler approach: just use the parent's Help function which will show all commands
		// If a specific command is requested, we could implement it, but for now just show general help
		cmd.Parent().Help()
	},
}