package journal

import (
	"github.com/spf13/cobra"
)

var JournalCmd = &cobra.Command{
	Use:   "journal",
	Short: "Various journal management tools",
	Long:  "", //TODO
	Args:  cobra.MinimumNArgs(1),
}
