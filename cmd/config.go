package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"tervdocs/internal/cli"
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
			cfg, err := config.Load(configPath)
			if err != nil {
				return err
			}
			cli.PrintTable(cmd.OutOrStdout(), "Current Config", []string{"Key", "Value"}, configRows(cfg))
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
			cli.PrintTable(cmd.OutOrStdout(), "Config Updated", []string{"Field", "Value"}, [][]string{
				{"OK", "Configuration updated"},
				{"Key", args[0]},
				{"Value", args[1]},
				{"Config File", config.ResolvePath(configPath)},
			})
			return nil
		},
	})
	return configCmd
}

func configRows(cfg config.Config) [][]string {
	return [][]string{
		{"provider", cfg.Provider},
		{"model", cfg.Model},
		{"template", cfg.Template},
		{"developer_name", blankAsDash(cfg.DeveloperName)},
		{"output.file", cfg.Output.File},
		{"output.backup", fmt.Sprintf("%t", cfg.Output.Backup)},
		{"scan.include", joinCSV(cfg.Scan.Include)},
		{"scan.exclude", joinCSV(cfg.Scan.Exclude)},
		{"scan.max_files", fmt.Sprintf("%d", cfg.Scan.MaxFiles)},
		{"scan.max_bytes_per_file", fmt.Sprintf("%d", cfg.Scan.MaxBytesPerFile)},
		{"providers.free.model", cfg.Providers.Free.Model},
		{"providers.free.base_url", cfg.Providers.Free.BaseURL},
		{"providers.openai.model", cfg.Providers.OpenAI.Model},
		{"providers.gemini.model", cfg.Providers.Gemini.Model},
		{"providers.claude.model", cfg.Providers.Claude.Model},
	}
}

func joinCSV(items []string) string {
	if len(items) == 0 {
		return "-"
	}
	return strings.Join(items, ", ")
}

func blankAsDash(v string) string {
	if strings.TrimSpace(v) == "" {
		return "-"
	}
	return v
}
