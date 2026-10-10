package habits

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/danielsrojo/obser/pkg/obsidian"
	"github.com/danielsrojo/obser/pkg/render"
	"github.com/spf13/cobra"
)

var (
	trendBy     string
	trendMetric string
	trendYear   int

	metrics = []string{"rate", "days", "total", "avg"}

	trendCmd = &cobra.Command{
		Use:   "trend",
		Short: "Compare habits across months or years",
		Long: `Show a grid with one row per habit and one column per month of a year
(--by month, the default) or per year since the first journal entry
(--by year). The last column shows whether the last complete period went up
(↑), down (↓) or stayed within 10% (→) of the period before it.`,
		Args: cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			switch trendBy {
			case "month":
			case "year":
				if cmd.Flags().Changed("year") {
					return fmt.Errorf("--year only applies to --by month")
				}
			default:
				return fmt.Errorf("invalid --by %q, must be month or year", trendBy)
			}
			for _, m := range metrics {
				if trendMetric == m {
					return nil
				}
			}
			return fmt.Errorf("invalid metric %q, must be one of: %s", trendMetric, strings.Join(metrics, ", "))
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return showTrend(trendBy, trendMetric, trendYear)
		},
	}
)

func init() {
	trendCmd.Flags().StringVar(&trendBy, "by", "month", "period of each column: month or year")
	trendCmd.Flags().StringVar(&trendMetric, "metric", "rate", "value of each cell: "+strings.Join(metrics, ", "))
	trendCmd.Flags().IntVarP(&trendYear, "year", "y", time.Now().Year(), "year to show with --by month")
	HabitsCmd.AddCommand(trendCmd)
}

// cell is one value of the grid. ok is false for blank cells and for amount
// metrics of habits without amounts.
type cell struct {
	text  string
	value float64
	ok    bool
}

func showTrend(by, metric string, year int) error {
	j, err := loadJournal()
	if err != nil {
		return err
	}

	var periods []obsidian.Range
	if by == "year" {
		if len(j.Days) == 0 {
			return fmt.Errorf("no journal entries")
		}
		from := time.Date(j.Days[0].Date.Year(), 1, 1, 0, 0, 0, 0, time.Local)
		to := time.Date(j.Today.Year(), 12, 31, 0, 0, 0, 0, time.Local)
		periods = obsidian.Periods(obsidian.Year, obsidian.Range{From: from, To: to})
	} else {
		if _, err := yearRange(j, year); err != nil {
			return err
		}
		from := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
		periods = obsidian.Periods(obsidian.Month, obsidian.Range{From: from, To: from.AddDate(1, 0, -1)})
	}
	end := periods[len(periods)-1].To
	if end.After(j.Today) {
		end = j.Today
	}

	header := []string{"HABIT"}
	for _, p := range periods {
		if by == "year" {
			header = append(header, strconv.Itoa(p.From.Year()))
		} else {
			header = append(header, p.From.Format("Jan"))
		}
	}
	header = append(header, "Δ")

	habits := habitsIn(j, obsidian.Range{From: periods[0].From, To: end})
	grid := make([][]cell, len(habits))
	for i, h := range habits {
		grid[i] = make([]cell, len(periods))
		for k, p := range periods {
			if p.From.After(j.Today) || p.To.Before(h.Start) {
				continue
			}
			grid[i][k] = metricCell(j.Stats(h.Name, p), h, metric)
		}
	}

	// The last complete period is the last one that ends before today.
	last := -1
	for k, p := range periods {
		if p.To.Before(j.Today) {
			last = k
		}
	}

	widths := make([]int, len(periods))
	for k := range periods {
		widths[k] = len(header[k+1])
		for i := range grid {
			widths[k] = max(widths[k], len([]rune(grid[i][k].text)))
		}
	}

	style := render.DetectStyle()
	for i := range header {
		header[i] = style.Bold(header[i])
	}
	table := render.Table{Header: header, Right: map[int]bool{}}
	for k := range periods {
		table.Right[k+1] = true
	}
	for i, h := range habits {
		var rowMax float64
		for _, c := range grid[i] {
			if c.ok {
				rowMax = math.Max(rowMax, c.value)
			}
		}
		row := []string{h.Name}
		for k, c := range grid[i] {
			text := strings.Repeat(" ", widths[k]-len([]rune(c.text))) + c.text
			if c.ok {
				text = style.Shade(text, shadeLevel(c.value, rowMax))
			}
			row = append(row, text)
		}
		var d string
		if last >= 1 && grid[i][last-1].ok && grid[i][last].ok {
			d = delta(grid[i][last-1].value, grid[i][last].value)
		}
		table.Rows = append(table.Rows, append(row, d))
	}
	fmt.Print(table.String())
	return nil
}

func metricCell(s obsidian.Stats, h obsidian.Habit, metric string) cell {
	switch metric {
	case "days":
		return cell{strconv.Itoa(s.DaysDone), float64(s.DaysDone), true}
	case "total":
		if !h.HasAmounts {
			return cell{text: "-"}
		}
		return cell{render.Amount(s.Total, h.Unit), float64(s.Total), true}
	case "avg":
		if !h.HasAmounts {
			return cell{text: "-"}
		}
		return cell{render.AmountF(s.Avg, h.Unit), s.Avg, true}
	}
	return cell{percent(s.Rate), s.Rate, true}
}

// shadeLevel maps a value to a level 0–4 relative to the highest value.
func shadeLevel(value, highest float64) int {
	if value <= 0 || highest <= 0 {
		return 0
	}
	return min(4, max(1, int(math.Ceil(value/highest*4))))
}

// delta compares a period's value with the one before it: ↑ or ↓ when it
// changed by more than 10%, → otherwise. From 0, any increase is ↑.
func delta(before, after float64) string {
	if before == 0 {
		if after > 0 {
			return "↑"
		}
		return "→"
	}
	switch change := (after - before) / before; {
	case change > 0.1:
		return "↑"
	case change < -0.1:
		return "↓"
	}
	return "→"
}
