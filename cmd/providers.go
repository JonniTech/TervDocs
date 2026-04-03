package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
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
			fmt.Fprintln(cmd.OutOrStdout(), "free")
			fmt.Fprintln(cmd.OutOrStdout(), "openai")
			fmt.Fprintln(cmd.OutOrStdout(), "gemini")
			fmt.Fprintln(cmd.OutOrStdout(), "claude")
		},
	})
	return providers
}
