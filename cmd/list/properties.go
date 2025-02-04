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
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("properties command")
		listProperties()
	},
}

func listProperties() {
	properties, err := obsidian.GetPropertiesNames()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("# Properties")
	for _, p := range properties {
		fmt.Println(p)
	}
}
