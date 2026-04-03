package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"tervdocs/internal/config"
)

func newConfigCmd() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage tervdocs configuration",
	}
	configCmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Show current config",
		RunE: func(cmd *cobra.Command, args []string) error {
			raw, err := config.Show(configPath)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), raw)
			return nil
		},
	})
	configCmd.AddCommand(&cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a config value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.Set(configPath, args[0], args[1]); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "updated")
			return nil
		},
	})
	return configCmd
}
