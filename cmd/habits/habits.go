package habits

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/danielsrojo/obser/pkg/obsidian"
	"github.com/danielsrojo/obser/pkg/render"
	"github.com/spf13/cobra"
)

var (
	summaryYear  int
	summaryMonth int

	HabitsCmd = &cobra.Command{
		Use:   "habits",
		Short: "Show how consistently each habit is done",
		Long: `Show one row per habit with the days it was done, its completion rate,
current and longest streak, total and average amount, and a sparkline of the
monthly completion rate.

By default it covers the last 365 days; use --year for a calendar year and
--month for a calendar month of --year or of the current year.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return showSummary(cmd.Flags().Changed("year"), summaryYear, cmd.Flags().Changed("month"), summaryMonth)
		},
	}
)

func init() {
	HabitsCmd.Flags().IntVarP(&summaryYear, "year", "y", 0, "calendar year to summarize (default is the last 365 days)")
	HabitsCmd.Flags().IntVarP(&summaryMonth, "month", "m", 0, "calendar month to summarize, 1-12, of --year or the current year")
}

// loadJournal loads the journal up to today and prints its warnings to stderr.
func loadJournal() (*obsidian.Journal, error) {
	j, err := obsidian.LoadJournal(time.Now())
	if err != nil {
		return nil, err
	}
	for _, w := range j.Warnings {
		fmt.Fprintln(os.Stderr, "warning:", w)
	}
	return j, nil
}

// yearRange returns the given calendar year, ending today for the current
// year, after checking that the journal covers it.
func yearRange(j *obsidian.Journal, year int) (obsidian.Range, error) {
	if len(j.Days) == 0 || year < j.Days[0].Date.Year() || year > j.Today.Year() {
		return obsidian.Range{}, fmt.Errorf("no journal entries for %d", year)
	}
	from := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	to := from.AddDate(1, 0, -1)
	if to.After(j.Today) {
		to = j.Today
	}
	return obsidian.Range{From: from, To: to}, nil
}

// monthRange returns the given calendar month, ending today for the current
// month, after checking that the journal covers it.
func monthRange(j *obsidian.Journal, year, month int) (obsidian.Range, error) {
	if month < 1 || month > 12 {
		return obsidian.Range{}, fmt.Errorf("invalid month %d, must be 1-12", month)
	}
	from := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	to := from.AddDate(0, 1, -1)
	if len(j.Days) == 0 || to.Before(j.Days[0].Date) || from.After(j.Today) {
		return obsidian.Range{}, fmt.Errorf("no journal entries for %s", from.Format("January 2006"))
	}
	if to.After(j.Today) {
		to = j.Today
	}
	return obsidian.Range{From: from, To: to}, nil
}

// habitsIn returns the habits whose tracking starts on or before the end of
// the range, sorted by name.
func habitsIn(j *obsidian.Journal, r obsidian.Range) []obsidian.Habit {
	var habits []obsidian.Habit
	for _, name := range j.HabitNames() {
		if h := j.Habits[name]; !h.Start.After(r.To) {
			habits = append(habits, h)
		}
	}
	return habits
}

func percent(rate float64) string {
	return strconv.Itoa(int(math.Round(rate*100))) + "%"
}

func showSummary(hasYear bool, year int, hasMonth bool, month int) error {
	j, err := loadJournal()
	if err != nil {
		return err
	}
	r := obsidian.Range{From: j.Today.AddDate(0, 0, -364), To: j.Today}
	switch {
	case hasMonth:
		if !hasYear {
			year = j.Today.Year()
		}
		r, err = monthRange(j, year, month)
	case hasYear:
		r, err = yearRange(j, year)
	}
	if err != nil {
		return err
	}

	style := render.DetectStyle()
	header := []string{"HABIT", "DAYS", "RATE", "STREAK", "BEST", "TOTAL", "AVG", "TREND"}
	for i := range header {
		header[i] = style.Bold(header[i])
	}
	table := render.Table{
		Header: header,
		Right:  map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true, 6: true},
	}
	for _, h := range habitsIn(j, r) {
		s := j.Stats(h.Name, r)
		total, avg := "-", "-"
		if s.HasTotal && s.DaysDone > 0 {
			total = render.Amount(s.Total, h.Unit)
			avg = render.AmountF(s.Avg, h.Unit)
		}
		table.Rows = append(table.Rows, []string{
			h.Name,
			strconv.Itoa(s.DaysDone),
			percent(s.Rate),
			strconv.Itoa(s.CurrentStreak),
			strconv.Itoa(s.LongestStreak),
			total,
			avg,
			render.Sparkline(monthlyRates(j, h, r.To)),
		})
	}
	fmt.Print(table.String())
	return nil
}

// monthlyRates returns the habit's completion rate in each of the 12
// calendar months ending with the month of last, NaN before tracking start.
func monthlyRates(j *obsidian.Journal, h obsidian.Habit, last time.Time) []float64 {
	first := time.Date(last.Year(), last.Month()-11, 1, 0, 0, 0, 0, time.Local)
	rates := make([]float64, 12)
	for i := range rates {
		from := first.AddDate(0, i, 0)
		month := obsidian.Range{From: from, To: from.AddDate(0, 1, -1)}
		if month.To.Before(h.Start) {
			rates[i] = math.NaN()
			continue
		}
		rates[i] = j.Stats(h.Name, month).Rate
	}
	return rates
}
