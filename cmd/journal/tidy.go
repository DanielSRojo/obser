package journal

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/danielsrojo/obser/pkg/obsidian"
	"github.com/spf13/cobra"
)

var (
	tidyBefore string
	tidyDryRun bool

	tidyCmd = &cobra.Command{
		Use:   "tidy",
		Short: "Remove empty properties from past journal entries",
		Long: `Remove the frontmatter properties that have no value (key:, null, ~, "",
[] or {}) from the journal entries dated before today.

Only the lines of the empty properties are removed; the rest of each note,
including its modification time, stays the same. Run it with --dry-run first.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return tidyJournal(tidyBefore, tidyDryRun)
		},
	}
)

func init() {
	tidyCmd.Flags().StringVar(&tidyBefore, "before", "", "only tidy entries dated before this day, as YYYY-MM-DD (default is today)")
	tidyCmd.Flags().BoolVarP(&tidyDryRun, "dry-run", "n", false, "print the properties that would be removed without changing any file")
	JournalCmd.AddCommand(tidyCmd)
}

func tidyJournal(beforeArg string, dryRun bool) error {
	today := obsidian.Midnight(time.Now())
	before := today
	if beforeArg != "" {
		var err error
		if before, err = time.ParseInLocation(time.DateOnly, beforeArg, time.Local); err != nil {
			return fmt.Errorf("invalid --before date: %w", err)
		}
		if before.After(today) {
			return fmt.Errorf("invalid --before date: %s is later than today", beforeArg)
		}
	}

	results, skipped, err := obsidian.TidyJournal(before, dryRun)
	if err != nil {
		return err
	}

	var properties int
	for _, r := range results {
		fmt.Printf("%s: %s\n", r.Path, strings.Join(r.Removed, ", "))
		properties += len(r.Removed)
	}
	verb := "removed"
	if dryRun {
		verb = "would remove"
	}
	fmt.Printf("%s %d properties from %d files\n", verb, properties, len(results))

	for _, s := range skipped {
		fmt.Fprintf(os.Stderr, "warning: skipped %s: %v\n", s.Path, s.Err)
	}
	if len(skipped) > 0 {
		return fmt.Errorf("%d journal entries skipped", len(skipped))
	}
	return nil
}
