package list

import (
	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List various Obsidian elements",
	Long:  "List properties, journal entries, or other elements from the Obsidian vault",
	Args:  cobra.MinimumNArgs(1),
}
