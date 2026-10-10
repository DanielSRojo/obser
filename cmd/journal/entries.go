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
	RunE: func(cmd *cobra.Command, args []string) error {
		return listJournalEntries()
	},
}

func listJournalEntries() error {
	entries, err := obsidian.GetJournalEntries()
	if err != nil {
		return fmt.Errorf("couldn't list journal entries: %w", err)
	}
	for _, e := range entries {
		fmt.Println(e)
	}
	return nil
}
