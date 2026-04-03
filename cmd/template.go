package cmd

import (
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"tervdocs/internal/cli"
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
			rows := [][]string{}
			for _, name := range templates.List() {
				tpl := templates.MustGet(name)
				rows = append(rows, []string{name, strconv.Itoa(len(tpl.Sections)), tpl.Tone})
			}
			cli.PrintTable(cmd.OutOrStdout(), "Templates", []string{"Template", "Sections", "Tone"}, rows)
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
			tpl := templates.MustGet(args[0])
			cli.PrintTable(cmd.OutOrStdout(), "Template Updated", []string{"Field", "Value"}, [][]string{
				{"OK", "Template updated"},
				{"Template", tpl.Name},
				{"Tone", tpl.Tone},
				{"Sections", strings.Join(tpl.Sections[:minInt(len(tpl.Sections), 6)], ", ")},
			})
			return nil
		},
	})
	return templateCmd
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
