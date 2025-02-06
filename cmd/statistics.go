package cmd

import (
	"fmt"
	"log"
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
		Run: func(cmd *cobra.Command, args []string) {
			showStatistics(year, month)
		},
	}
)

func init() {
	statisticsCmd.Flags().IntVarP(&year, "year", "y", 0, "Year for statistics")
	statisticsCmd.Flags().IntVarP(&month, "month", "m", 0, "Month for statistics (1-12)")

	RootCmd.AddCommand(statisticsCmd)
}

func showStatistics(year, month int) {
	statistics, err := obsidian.GetStatistics(&year, &month)
	if err != nil {
		log.Fatal(err)
	}

	var header string
	if month > 0 {
		header = fmt.Sprintf("\n### %s %d ###", time.Month(month), year)
	} else {
		header = fmt.Sprintf("\n### %d ###", year)
	}
	fmt.Println(header)

	for _, p := range statistics {
		if p.Value == 0 {
			continue
		}
		fmt.Printf("%s: %d %s\n", p.Name, p.Value, p.Unit)
	}
}
