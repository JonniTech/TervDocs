package cmd

import (
	"github.com/spf13/cobra"

	"tervdocs/internal/cli"
	"tervdocs/internal/config"
)

func newProvidersCmd() *cobra.Command {
	providers := &cobra.Command{
		Use:   "providers",
		Short: "Provider utilities",
	}
	providers.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List available providers",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := config.Default()
			cli.PrintTable(cmd.OutOrStdout(), "Providers", []string{"Provider", "Default Model", "Quality", "Auth"}, [][]string{
				{"free", cfg.Providers.Free.Model, "shared / variable", "built-in shared key"},
				{"openai", cfg.Providers.OpenAI.Model, "high", "OPENAI_API_KEY"},
				{"gemini", cfg.Providers.Gemini.Model, "high", "GEMINI_API_KEY"},
				{"claude", cfg.Providers.Claude.Model, "high", "ANTHROPIC_API_KEY"},
			})
			cli.PrintTable(cmd.OutOrStdout(), "Recommendation", []string{"State", "Detail"}, [][]string{
				{"WARN", "The free provider is convenient, but it can be rate-limited or unstable because it is shared."},
				{"INFO", "For the best README quality and reliability, prefer Claude, Gemini, or OpenAI with your own API key."},
			})
		},
	})
	return providers
}
