package cmd

import (
	"errors"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"tervdocs/internal/cli"
	"tervdocs/internal/config"
	"tervdocs/internal/templates"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize .tervdocs.toml with interactive setup",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Default()
			path := config.ResolvePath(configPath)
			if existing, err := config.Load(path); err == nil {
				cfg = existing
			} else if !errors.Is(err, config.ErrConfigNotFound) {
				return err
			}

			qs := []*survey.Question{
				{
					Name: "provider",
					Prompt: &survey.Select{
						Message: "Choose provider:",
						Options: []string{"free", "openai", "gemini", "claude"},
						Default: cfg.Provider,
					},
				},
				{
					Name: "template",
					Prompt: &survey.Select{
						Message: "Choose README template:",
						Options: templates.List(),
						Default: cfg.Template,
					},
				},
				{
					Name: "developer_name",
					Prompt: &survey.Input{
						Message: "Developer name for README footer:",
						Default: cfg.DeveloperName,
					},
				},
				{
					Name: "output",
					Prompt: &survey.Input{
						Message: "Output file path:",
						Default: cfg.Output.File,
					},
				},
				{
					Name: "include",
					Prompt: &survey.Input{
						Message: "Include paths (comma-separated, blank for all):",
					},
				},
				{
					Name: "exclude",
					Prompt: &survey.Input{
						Message: "Extra excludes (comma-separated):",
					},
				},
			}
			ans := struct {
				Provider      string
				Template      string
				DeveloperName string `survey:"developer_name"`
				Output        string
				Include       string
				Exclude       string
			}{}
			if err := survey.Ask(qs, &ans); err != nil {
				return err
			}

			cfg.Provider = ans.Provider
			cfg.Template = ans.Template
			cfg.DeveloperName = ans.DeveloperName
			cfg.Output.File = ans.Output
			cfg.Scan.Include = csv(ans.Include)
			cfg.Scan.Exclude = append(cfg.Scan.Exclude, csv(ans.Exclude)...)

			if err := config.Save(path, cfg); err != nil {
				return err
			}
			cli.PrintTable(cmd.OutOrStdout(), "Initialization Complete", []string{"Field", "Value"}, [][]string{
				{"Config File", path},
				{"Provider", cfg.Provider},
				{"Template", cfg.Template},
				{"Developer", emptyFallback(cfg.DeveloperName, "not set")},
				{"Output", cfg.Output.File},
			})
			printProviderAdvisory(cmd, cfg.Provider)
			return nil
		},
	}
}

func csv(s string) []string {
	out := []string{}
	cur := ""
	for _, r := range s {
		if r == ',' {
			if cur != "" {
				out = append(out, trimSpace(cur))
			}
			cur = ""
			continue
		}
		cur += string(r)
	}
	if trimSpace(cur) != "" {
		out = append(out, trimSpace(cur))
	}
	return out
}

func trimSpace(s string) string {
	i := 0
	j := len(s) - 1
	for i <= j && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	for j >= i && (s[j] == ' ' || s[j] == '\t') {
		j--
	}
	if i > j {
		return ""
	}
	return s[i : j+1]
}

func emptyFallback(v, fallback string) string {
	if trimSpace(v) == "" {
		return fallback
	}
	return trimSpace(v)
}
