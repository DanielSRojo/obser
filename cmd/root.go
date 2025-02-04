package cmd

import (
	"fmt"
	"os"

	"github.com/danielsrojo/obser/cmd/list"
	"github.com/spf13/cobra"
)

var (
	cfgFile string

	RootCmd = &cobra.Command{
		Use:   "obser",
		Short: "obser",
		Long:  "obser",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("RootCmd")
			cmd.Help()
		},
	}
)

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.obser.toml)")

	RootCmd.AddCommand(list.ListCmd)
}
