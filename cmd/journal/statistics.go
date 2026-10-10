package journal

import (
	"fmt"
	"os"
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

	if month != 0 {
		return showMonthStatistics(year, month)
	}

	return showYearStatistics(year)
}

func showMonthStatistics(year, month int) error {
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
		if p.Value == 0 {
			continue
		}
		formatProperty(&p)
		fmt.Fprintf(w, "%s\t%d\t%s\n", p.Name, p.Value, p.Unit)
	}
	return w.Flush()
}

func showYearStatistics(year int) error {
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
		if p.Value == 0 {
			continue
		}
		formatProperty(&p)
		fmt.Fprintf(w, "%s\t%d\t%s\n", p.Name, p.Value, p.Unit)
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

func formatProperty(p *obsidian.Property) {
	if p.Unit == "seconds" && p.Value >= 200 {
		p.Value = p.Value / 60
		p.Unit = "minutes"
	}
	if p.Unit == "minutes" && p.Value >= 100 {
		p.Value = p.Value / 60
		p.Unit = "hours"
	}
}
