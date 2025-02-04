package list

import (
	"fmt"
	"log"

	"github.com/danielsrojo/obser/pkg/obsidian"
	"github.com/spf13/cobra"
)

func init() {
	ListCmd.AddCommand(notesCmd)
}

var notesCmd = &cobra.Command{
	Use: "notes",
	Run: func(cmd *cobra.Command, args []string) {
		listNotes()
	},
}

func listNotes() {
	entries, err := obsidian.GetNotesNames()
	if err != nil {
		log.Fatal("error: couldn't list vault's notes")
	}
	for _, e := range entries {
		fmt.Println(e)
	}
}
