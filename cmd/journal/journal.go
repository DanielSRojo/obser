package journal

import (
	"github.com/spf13/cobra"
)

var JournalCmd = &cobra.Command{
	Use:   "journal",
	Short: "Various journal management tools",
	Long:  "", //TODO
	// Runnable with no arguments, so that an unknown subcommand is an error
	// instead of printing the help and exiting 0.
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}
