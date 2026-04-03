package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"tervdocs/internal/cli"
	"tervdocs/internal/config"
	"tervdocs/internal/doctor"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Validate configuration, credentials, and scan assumptions",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configPath)
			if err != nil {
				if errors.Is(err, config.ErrConfigNotFound) {
					return fmt.Errorf("missing config file; run `tervdocs init` first")
				}
				return err
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.TimeoutSec)*time.Second)
			defer cancel()
			rep := doctor.Run(ctx, cfg, rootDir)

			for _, c := range rep.Checks {
				fmt.Fprintln(cmd.OutOrStdout(), cli.Success("%s", c))
			}
			for _, w := range rep.Warnings {
				fmt.Fprintln(cmd.OutOrStdout(), cli.Warn("%s", w))
			}
			for _, e := range rep.Errors {
				fmt.Fprintln(cmd.OutOrStdout(), cli.Error("%s", e))
			}
			if len(rep.Errors) > 0 {
				return fmt.Errorf("doctor found %d issue(s)", len(rep.Errors))
			}
			return nil
		},
	}
}
