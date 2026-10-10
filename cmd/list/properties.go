package list

import (
	"fmt"

	"github.com/danielsrojo/obser/pkg/obsidian"
	"github.com/spf13/cobra"
)

func init() {
	ListCmd.AddCommand(propertiesCmd)
}

var propertiesCmd = &cobra.Command{
	Use:   "properties",
	Short: "properties",
	Long:  "properties",
	RunE: func(cmd *cobra.Command, args []string) error {
		return listProperties()
	},
}

func listProperties() error {
	properties, err := obsidian.GetPropertiesNames()
	if err != nil {
		return err
	}

	for _, p := range properties {
		fmt.Println(p)
	}
	return nil
}
