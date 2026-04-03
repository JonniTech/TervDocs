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
	if !strings.Contains(got, "# **Demo**") {
		t.Fatalf("expected bolded title")
	}
	if !strings.Contains(got, "## **Aim**") {
		t.Fatalf("expected bolded aim section")
	}
	if !strings.Contains(got, "```mermaid") {
		t.Fatalf("expected mermaid diagram")
	}
	if !strings.Contains(got, "Programmed by Someone") {
		t.Fatalf("expected developer footer")
	}
	if !strings.Contains(got, DividerAssetRelativePath()) {
		t.Fatalf("expected GitHub-safe divider asset path")
	}
}

func TestEnhanceUsesBrandedProjectTitleForTervDocs(t *testing.T) {
	repo := scan.RepoSummary{
		ProjectName:      "tervdocs",
		PrimaryLanguage:  "Go",
		MainLanguages:    []string{"Go"},
		PackageManager:   "go modules",
		FolderFileCounts: map[string]int{"cmd": 1},
	}
	ctx := summarize.Context{
		PurposeGuess:     "Build clear documentation from code.",
		ProblemStatement: "Projects are hard to understand quickly.",
		SolutionSummary:  "The CLI scans the repo and generates a structured README.",
		BrandColor:       "#00ADD8",
	}

	got := Enhance("Plain intro.\n", repo, ctx, config.Config{})
	if !strings.Contains(got, "# **TervDocs**") {
		t.Fatalf("expected branded TervDocs heading, got:\n%s", got)
	}
}
