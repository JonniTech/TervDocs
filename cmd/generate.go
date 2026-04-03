package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"tervdocs/internal/cli"
	"tervdocs/internal/config"
	"tervdocs/internal/generate"
)

func newGenerateCmd() *cobra.Command {
	var (
		dryRun   bool
		provider string
		model    string
		template string
		output   string
	)

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Scan repository and generate README.md",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configPath)
			if err != nil {
				if errors.Is(err, config.ErrConfigNotFound) {
					return fmt.Errorf("missing config file; run `tervdocs init` first")
				}
				return err
			}
			selectedProvider := valueOr(cfg.Provider, provider)
			printProviderAdvisory(cmd, selectedProvider)

			spin := cli.NewSpinner("Scanning and generating README")
			spin.Start()

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.TimeoutSec)*time.Second)
			defer cancel()
			res, err := container.Generator.Run(ctx, cfg, generate.Options{
				RootDir:  rootDir,
				DryRun:   dryRun,
				Provider: provider,
				Model:    model,
				Template: template,
				Output:   output,
			})
			spin.Stop()
			if err != nil {
				return err
			}
			rows := [][]string{
				{"OK", "README generated successfully"},
				{"Output", res.OutputPath},
				{"Provider", res.Provider},
				{"Model", res.Model},
				{"Template", valueOr(cfg.Template, template)},
				{"Files", fmt.Sprintf("%d", res.Scan.FilesScanned)},
				{"Language", emptyValue(res.Scan.PrimaryLanguage)},
				{"Frameworks", joinOrNone(res.Scan.Frameworks)},
			}
			if res.BackupPath != "" {
				rows = append(rows, []string{"Backup", res.BackupPath})
			}
			if dryRun {
				rows = append(rows, []string{"WARN", "Dry-run enabled, file not written"})
			}
			for _, warning := range res.Warnings {
				rows = append(rows, []string{"WARN", warning})
			}
			cli.PrintTable(cmd.OutOrStdout(), "Generation Summary", []string{"Field", "Value"}, rows)
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Generate without writing output file")
	cmd.Flags().StringVar(&provider, "provider", "", "Override provider for this run")
	cmd.Flags().StringVar(&model, "model", "", "Override model for this run")
	cmd.Flags().StringVar(&template, "template", "", "Override template for this run")
	cmd.Flags().StringVar(&output, "output", "", "Override output file path")
	return cmd
}

func joinOrNone(items []string) string {
	if len(items) == 0 {
		return "none detected"
	}
	return strings.Join(items, ", ")
}

func emptyValue(v string) string {
	if strings.TrimSpace(v) == "" {
		return "unknown"
	}
	return v
}

func valueOr(current, override string) string {
	if strings.TrimSpace(override) != "" {
		return override
	}
	return current
}
