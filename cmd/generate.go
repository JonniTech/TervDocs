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

			spin := cli.NewSpinner("Scanning and generating README")
			spin.Start()
			defer spin.Stop()

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
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), cli.Success("README generated at %s", res.OutputPath))
			if res.BackupPath != "" {
				fmt.Fprintln(cmd.OutOrStdout(), cli.Info("Backup created at %s", res.BackupPath))
			}
			if dryRun {
				fmt.Fprintln(cmd.OutOrStdout(), cli.Warn("Dry-run enabled, file not written"))
			}
			fmt.Fprintln(cmd.OutOrStdout(), cli.Info("Scanned files: %d", res.Scan.FilesScanned))
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
