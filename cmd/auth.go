package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

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
			fmt.Fprintln(cmd.OutOrStdout(), "auth configuration looks valid")
			return nil
		},
	})
	return auth
}
