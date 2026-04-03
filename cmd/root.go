package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"tervdocs/internal/app"
	"tervdocs/internal/cli"
)

var (
	version    = "0.1.0"
	configPath string
	rootDir    string
)

var container = app.New()

var rootCmd = &cobra.Command{
	Use:   "tervdocs",
	Short: "Generate high-quality README files from repository context",
	Long:  cli.Banner(version),
	Example: `  tervdocs init
  tervdocs doctor
  tervdocs generate --provider openai --model gpt-4o-mini
  tervdocs preview --template tervux`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.EnableCommandSorting = true
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Path to .tervdocs.toml")
	rootCmd.PersistentFlags().StringVar(&rootDir, "root", ".", "Project root directory")
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		_ = container
	}
	rootCmd.AddCommand(
		newInitCmd(),
		newGenerateCmd(),
		newPreviewCmd(),
		newDoctorCmd(),
		newProvidersCmd(),
		newAuthCmd(),
		newConfigCmd(),
		newVersionCmd(),
		newTemplateCmd(),
	)
}

func printErr(err error) {
	fmt.Fprintln(os.Stderr, cli.Error("%v", err))
}
