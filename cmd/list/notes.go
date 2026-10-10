package list

import (
	"fmt"

	"github.com/danielsrojo/obser/pkg/obsidian"
	"github.com/spf13/cobra"
)

func init() {
	ListCmd.AddCommand(notesCmd)
}

var notesCmd = &cobra.Command{
	Use: "notes",
	RunE: func(cmd *cobra.Command, args []string) error {
		return listNotes()
	},
}

func listNotes() error {
	entries, err := obsidian.GetNotesNames()
	if err != nil {
		return fmt.Errorf("couldn't list vault's notes: %w", err)
	}
	for _, e := range entries {
		fmt.Println(e)
	}
	return nil
}
