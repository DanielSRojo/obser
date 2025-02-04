package list

import (
	"fmt"

	"github.com/danielsrojo/obser/pkg/obsidian"
	"github.com/spf13/cobra"
)

func init() {
	ListCmd.AddCommand(journalCmd)
}

var journalCmd = &cobra.Command{
	Use:   "journal",
	Short: "journal",
	Long:  "journal",
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
