package list

import (
	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List various Obsidian elements",
	Long:  "List properties, journal entries, or other elements from the Obsidian vault",
	// Runnable with no arguments, so that an unknown subcommand is an error
	// instead of printing the help and exiting 0.
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}
