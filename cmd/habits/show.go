package habits

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/danielsrojo/obser/pkg/obsidian"
	"github.com/danielsrojo/obser/pkg/render"
	"github.com/spf13/cobra"
)

const barWidth = 20

var (
	showYear int

	showCmd = &cobra.Command{
		Use:   "show <habit>",
		Short: "Show a habit's calendar, monthly and weekday breakdown for a year",
		Long: `Show a single habit for a year: its days done, rate, streaks and amounts,
a calendar heatmap of every day, a bar per month, and its completion rate on
each weekday.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return showHabit(args[0], showYear)
		},
	}
)

func init() {
	showCmd.Flags().IntVarP(&showYear, "year", "y", time.Now().Year(), "year to show")
	HabitsCmd.AddCommand(showCmd)
}

func showHabit(name string, year int) error {
	j, err := loadJournal()
	if err != nil {
		return err
	}
	h, ok := j.Habits[name]
	if !ok {
		return fmt.Errorf("unknown habit %q, known habits are: %s", name, strings.Join(j.HabitNames(), ", "))
	}
	r, err := yearRange(j, year)
	if err != nil {
		return err
	}
	style := render.DetectStyle()
	s := j.Stats(name, r)

	total, avg := "-", "-"
	if s.HasTotal && s.DaysDone > 0 {
		total = render.Amount(s.Total, h.Unit)
		avg = render.AmountF(s.Avg, h.Unit)
	}
	fmt.Printf("%s · %d\n", style.Bold(name), year)
	fmt.Printf("Days %d · Rate %s · Streak %d · Best %d · Total %s · Avg %s\n\n",
		s.DaysDone, percent(s.Rate), s.CurrentStreak, s.LongestStreak, total, avg)

	yearLevels := j.Levels(name, obsidian.Range{From: r.From, To: time.Date(year, 12, 31, 0, 0, 0, 0, time.Local)})
	fmt.Print(render.Heatmap(year, func(day time.Time) int {
		if level, ok := yearLevels[day]; ok {
			return level
		}
		return -1
	}, style))
	legend := make([]string, 5)
	for i := range legend {
		legend[i] = style.Level(i)
	}
	fmt.Printf("    %s %s %s\n\n", style.Faint("Less"), strings.Join(legend, ""), style.Faint("More"))

	months := render.Table{Right: map[int]bool{2: true, 3: true}}
	for _, month := range obsidian.Periods(obsidian.Month, r) {
		ms := j.Stats(name, month)
		row := []string{month.From.Format("Jan"), render.Bar(ms.Rate, barWidth), strconv.Itoa(ms.DaysDone)}
		if h.HasAmounts {
			row = append(row, render.Amount(ms.Total, h.Unit))
		}
		months.Rows = append(months.Rows, row)
	}
	fmt.Print(months.String())
	fmt.Println()

	weekdays := render.Table{Right: map[int]bool{2: true}}
	rates := j.WeekdayRates(name, r)
	for i, wd := range []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"} {
		weekdays.Rows = append(weekdays.Rows, []string{wd, render.Bar(rates[i], barWidth), percent(rates[i])})
	}
	fmt.Print(weekdays.String())
	return nil
}
