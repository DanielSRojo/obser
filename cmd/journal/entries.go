package journal

import (
	"fmt"

	"github.com/danielsrojo/obser/pkg/obsidian"
	"github.com/spf13/cobra"
)

func init() {
	JournalCmd.AddCommand(entriesCmd)
}

var entriesCmd = &cobra.Command{
	Use:   "entries",
	Short: "entries",
	Long:  "entries",
	Run: func(cmd *cobra.Command, args []string) {
		listJournalEntries()
	},
}

func listJournalEntries() {
	entries, err := obsidian.GetJournalEntries()
	if err != nil {
		panic("ERROR journal entries")
	}
	for _, e := range entries {
		fmt.Println(e)
	}
}
