package cmd

import (
	"fmt"
	"os"

	"github.com/danielsrojo/obser/cmd/habits"
	"github.com/danielsrojo/obser/cmd/journal"
	"github.com/danielsrojo/obser/cmd/list"
	"github.com/danielsrojo/obser/pkg/obsidian"
	"github.com/spf13/cobra"
)

var (
	vaultDir string

	RootCmd = &cobra.Command{
		Use:   "obser",
		Short: "obser",
		Long:  "obser",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return setVaultDir()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
)

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVar(&vaultDir, "vault", "", "path to the Obsidian vault (default is $OBSER_VAULT or $HOME/.obsidian)")

	RootCmd.AddCommand(list.ListCmd)
	RootCmd.AddCommand(journal.JournalCmd)
	RootCmd.AddCommand(habits.HabitsCmd)
}

// setVaultDir picks the vault from --vault, then $OBSER_VAULT, then the default.
func setVaultDir() error {
	dir := vaultDir
	if dir == "" {
		dir = os.Getenv("OBSER_VAULT")
	}
	if dir == "" {
		var err error
		dir, err = obsidian.DefaultVaultDir()
		if err != nil {
			return err
		}
	}

	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("vault %s: %w", dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("vault %s: not a directory", dir)
	}

	obsidian.SetVaultDir(dir)
	return nil
}
