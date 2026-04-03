package cmd

import (
	"github.com/spf13/cobra"

	"tervdocs/internal/cli"
)

func printProviderAdvisory(cmd *cobra.Command, provider string) {
	switch provider {
	case "free":
		cli.PrintTable(cmd.OutOrStdout(), "Provider Advisory", []string{"State", "Detail"}, [][]string{
			{"WARN", "The shared free provider can be rate-limited, slow, or temporarily unavailable."},
			{"WARN", "This can lead to fallback generation or, in some environments, incomplete README quality."},
			{"INFO", "If this happens, it is a provider-side limitation rather than negligence by tervdocs."},
			{"INFO", "For higher-quality and more reliable output, use Claude, Gemini, or OpenAI with your own API key."},
			{"INFO", "Suggested API keys: ANTHROPIC_API_KEY, GEMINI_API_KEY, OPENAI_API_KEY."},
		})
	case "claude", "gemini", "openai":
		cli.PrintTable(cmd.OutOrStdout(), "Provider Advisory", []string{"State", "Detail"}, [][]string{
			{"OK", "Using a dedicated provider generally gives more reliable and higher-quality README output."},
			{"INFO", "Make sure the correct API key is set before running generation."},
			{"INFO", "Expected environment variable: " + providerEnvVar(provider)},
		})
	}
}

func providerEnvVar(provider string) string {
	switch provider {
	case "openai":
		return "OPENAI_API_KEY"
	case "gemini":
		return "GEMINI_API_KEY"
	case "claude":
		return "ANTHROPIC_API_KEY"
	default:
		return "ZAI_API_KEY"
	}
}
