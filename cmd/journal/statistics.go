package journal

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/danielsrojo/obser/pkg/obsidian"
	"github.com/spf13/cobra"
)

var (
	year  int
	month int

	statisticsCmd = &cobra.Command{
		Use:   "statistics",
		Short: "Show journal properties based statistics",
		Long:  "Show statisctics based on journal notes properties values",
		RunE: func(cmd *cobra.Command, args []string) error {
			return showStatistics(year, month)
		},
	}
)

func init() {
	statisticsCmd.Flags().IntVarP(&year, "year", "y", 0, "Year for statistics")
	statisticsCmd.Flags().IntVarP(&month, "month", "m", 0, "Month for statistics (1-12)")
	JournalCmd.AddCommand(statisticsCmd)
}

func showStatistics(year, month int) error {
	setDefaultDateValues(&year, &month)

	statistics, err := obsidian.GetStatistics(year, month)
	if err != nil {
		return err
	}
	if len(statistics) == 0 {
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "%s\t%s\t%s\n", "NAME", "VALUE", "UNIT")
	for _, p := range statistics {
		value, unit := formatValue(p)
		fmt.Fprintf(w, "%s\t%s\t%s\n", p.Name, value, unit)
	}
	return w.Flush()
}

func setDefaultDateValues(year, month *int) {
	if *year != 0 {
		return
	}
	now := time.Now()
	*year = now.Year()
	if *month == 0 {
		*month = int(now.Month())
	}
}

// formatValue shows durations of 200 seconds or more in minutes, and of 100
// minutes or more in hours, rounded to whole minutes or tenths of an hour.
func formatValue(p obsidian.Property) (value, unit string) {
	if p.Unit != "seconds" || p.Value < 200 {
		return strconv.Itoa(p.Value), p.Unit
	}
	minutes := float64(p.Value) / 60
	if math.Round(minutes) < 100 {
		return strconv.Itoa(int(math.Round(minutes))), "minutes"
	}
	return strconv.FormatFloat(math.Round(minutes/6)/10, 'f', -1, 64), "hours"
}
