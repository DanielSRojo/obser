package journal

import (
	"fmt"
	"time"

	"github.com/danielsrojo/obser/pkg/obsidian"
	"github.com/spf13/cobra"
)

var (
	fillFrom   string
	fillTo     string
	fillDryRun bool

	fillCmd = &cobra.Command{
		Use:   "fill",
		Short: "Create missing journal entries from the template",
		Long: `Create a journal entry from the daily notes template for every missing day.

By default it fills the days between the last journal entry and today. Days
that already have an entry are left untouched.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fillJournal(fillFrom, fillTo, fillDryRun)
		},
	}
)

func init() {
	fillCmd.Flags().StringVar(&fillFrom, "from", "", "first day to fill, as YYYY-MM-DD (default is the day after the last entry)")
	fillCmd.Flags().StringVar(&fillTo, "to", "", "last day to fill, as YYYY-MM-DD (default is today)")
	fillCmd.Flags().BoolVarP(&fillDryRun, "dry-run", "n", false, "print the entries that would be created without creating them")
	JournalCmd.AddCommand(fillCmd)
}

func fillJournal(fromArg, toArg string, dryRun bool) error {
	cfg, err := obsidian.LoadDailyNotesConfig()
	if err != nil {
		return err
	}

	now := time.Now()
	to := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if toArg != "" {
		if to, err = time.Parse(time.DateOnly, toArg); err != nil {
			return fmt.Errorf("invalid --to date: %w", err)
		}
	}

	var from time.Time
	if fromArg != "" {
		if from, err = time.Parse(time.DateOnly, fromArg); err != nil {
			return fmt.Errorf("invalid --from date: %w", err)
		}
	} else {
		last, err := obsidian.LastJournalDate(cfg, to)
		if err != nil {
			return err
		}
		if last.IsZero() {
			return fmt.Errorf("no journal entries found, use --from to choose the first day")
		}
		from = last.AddDate(0, 0, 1)
	}

	if from.After(to) {
		return nil
	}

	created, err := obsidian.FillJournal(cfg, from, to, dryRun)
	for _, name := range created {
		fmt.Println(name)
	}
	return err
}
