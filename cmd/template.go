package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"tervdocs/internal/config"
	"tervdocs/internal/templates"
)

func newTemplateCmd() *cobra.Command {
	templateCmd := &cobra.Command{
		Use:   "template",
		Short: "Template management",
	}
	templateCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List built-in templates",
		Run: func(cmd *cobra.Command, args []string) {
			for _, name := range templates.List() {
				fmt.Fprintln(cmd.OutOrStdout(), name)
			}
		},
	})
	templateCmd.AddCommand(&cobra.Command{
		Use:   "use <name>",
		Short: "Set template in config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := templates.Get(args[0]); err != nil {
				return err
			}
			if err := config.Set(configPath, "template", args[0]); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "template updated")
			return nil
		},
	})
	return templateCmd
}
