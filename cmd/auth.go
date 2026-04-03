package cmd

import (
	"github.com/spf13/cobra"

	"tervdocs/internal/cli"
	"tervdocs/internal/config"
	"tervdocs/internal/providers"
)

func newAuthCmd() *cobra.Command {
	auth := &cobra.Command{
		Use:   "auth",
		Short: "Authentication and credential checks",
	}
	auth.AddCommand(&cobra.Command{
		Use:   "test",
		Short: "Verify selected provider credentials/config",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configPath)
			if err != nil {
				return err
			}
			p, err := providers.New(cfg.Provider, cfg)
			if err != nil {
				return err
			}
			if err := p.Validate(); err != nil {
				return err
			}
			cli.PrintTable(cmd.OutOrStdout(), "Auth Test", []string{"Field", "Value"}, [][]string{
				{"OK", "Authentication configuration looks valid"},
				{"Provider", cfg.Provider},
				{"Model", cfg.Model},
			})
			return nil
		},
	})
	return auth
}
