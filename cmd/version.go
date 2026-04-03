package cmd

import (
	"github.com/spf13/cobra"

	"tervdocs/internal/cli"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			cli.PrintTable(cmd.OutOrStdout(), "Version", []string{"Field", "Value"}, [][]string{
				{"CLI", "tervdocs"},
				{"Version", version},
				{"Brand", "Tervux"},
			})
		},
	}
}
