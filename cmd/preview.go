package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"tervdocs/internal/cli"
	"tervdocs/internal/config"
	"tervdocs/internal/generate"
)

func newPreviewCmd() *cobra.Command {
	var (
		provider string
		model    string
		template string
	)
	cmd := &cobra.Command{
		Use:   "preview",
		Short: "Generate README preview without writing file",
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
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.TimeoutSec)*time.Second)
			defer cancel()
			res, err := container.Generator.Run(ctx, cfg, generate.Options{
				RootDir:  rootDir,
				Preview:  true,
				Provider: provider,
				Model:    model,
				Template: template,
			})
			if err != nil {
				return err
			}
			cli.PrintTable(cmd.OutOrStdout(), "Preview Summary", []string{"Field", "Value"}, [][]string{
				{"Provider", res.Provider},
				{"Model", res.Model},
				{"Template", valueOr(cfg.Template, template)},
				{"Files", fmt.Sprintf("%d", res.Scan.FilesScanned)},
				{"Language", emptyValue(res.Scan.PrimaryLanguage)},
			})
			if len(res.Warnings) > 0 {
				rows := [][]string{}
				for _, warning := range res.Warnings {
					rows = append(rows, []string{"WARN", warning})
				}
				cli.PrintTable(cmd.OutOrStdout(), "Preview Warnings", []string{"State", "Detail"}, rows)
			}
			fmt.Fprintln(cmd.OutOrStdout(), res.Markdown)
			return nil
		},
	}
	cmd.Flags().StringVar(&provider, "provider", "", "Override provider for this run")
	cmd.Flags().StringVar(&model, "model", "", "Override model for this run")
	cmd.Flags().StringVar(&template, "template", "", "Override template for this run")
	return cmd
}
