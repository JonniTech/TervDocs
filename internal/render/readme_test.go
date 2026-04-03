package render

import (
	"strings"
	"testing"

	"tervdocs/internal/config"
	"tervdocs/internal/scan"
	"tervdocs/internal/summarize"
)

func TestEnhanceAddsSectionsAndFooter(t *testing.T) {
	repo := scan.RepoSummary{
		ProjectName:      "demo",
		PrimaryLanguage:  "Go",
		MainLanguages:    []string{"Go"},
		Frameworks:       []string{"Cobra"},
		PackageManager:   "go modules",
		TopFolders:       []string{"cmd", "internal"},
		FolderFileCounts: map[string]int{"cmd": 3, "internal": 8},
		EntryPoints:      []string{"main.go"},
	}
	ctx := summarize.Context{
		PurposeGuess:     "Build clear documentation from code.",
		ProblemStatement: "Projects are hard to understand quickly.",
		SolutionSummary:  "The CLI scans the repo and generates a structured README.",
		Languages:        []string{"Go"},
		Frameworks:       []string{"Cobra"},
		Dependencies:     []string{"cobra", "viper"},
		BrandColor:       "#00ADD8",
	}
	cfg := config.Config{DeveloperName: "Someone"}

	got := Enhance("# Demo\n\nShort intro.\n", repo, ctx, cfg)
	if !strings.Contains(got, "## Aim") {
		t.Fatalf("expected aim section")
	}
	if !strings.Contains(got, "```mermaid") {
		t.Fatalf("expected mermaid diagram")
	}
	if !strings.Contains(got, "Programmed by Someone") {
		t.Fatalf("expected developer footer")
	}
	if !strings.Contains(got, "data-tervdocs-divider") {
		t.Fatalf("expected SVG divider")
	}
}
