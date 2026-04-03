package render

import (
	"sort"
	"strings"

	"tervdocs/internal/config"
	"tervdocs/internal/scan"
	"tervdocs/internal/summarize"
)

func FallbackMarkdown(repo scan.RepoSummary, ctx summarize.Context, cfg config.Config) string {
	sections := []string{
		"# " + fallbackProjectName(repo.ProjectName),
		"",
		ctx.PurposeGuess,
		"",
		"## Overview",
		"",
		ctx.SolutionSummary,
		"",
		"## Aim",
		"",
		ctx.PurposeGuess,
		"",
		"## Problem",
		"",
		ctx.ProblemStatement,
		"",
		"## Solution",
		"",
		ctx.SolutionSummary,
		"",
		"## Features",
		"",
		bulletList(featureBullets(ctx)),
		"",
		"## Tech Stack",
		"",
		techStackSection(repo, ctx),
		"",
		"## Installation",
		"",
		fallbackInstallation(repo),
		"",
		"## Usage",
		"",
		fallbackUsage(repo),
		"",
	}

	if len(ctx.Scripts) > 0 {
		sections = append(sections, "## Scripts and Commands", "", commandSection(ctx.Scripts), "")
	}
	if len(ctx.EnvVars) > 0 {
		sections = append(sections, "## Environment Variables", "", envSection(ctx.EnvVars), "")
	}

	sections = append(sections,
		"## Flow Diagram",
		"",
		mermaidDiagram(repo),
		"",
		"## Architecture Notes",
		"",
		bulletList(architectureBullets(ctx)),
		"",
		"## Project Structure",
		"",
		projectStructureSection(repo),
		"",
		"## Contributing",
		"",
		"Contributions should preserve the existing repository structure, developer workflow, and documentation quality.",
		"",
		"## License",
		"",
		"Set the appropriate project license for distribution.",
	)

	return Enhance(strings.Join(sections, "\n"), repo, ctx, cfg)
}

func fallbackProjectName(name string) string {
	if strings.TrimSpace(name) == "" {
		return "Project"
	}
	return name
}

func fallbackInstallation(repo scan.RepoSummary) string {
	lines := []string{}
	if repo.PackageManager == "go modules" {
		lines = append(lines, "```bash", "go mod tidy", "```")
	} else if repo.PackageManager != "" && repo.PackageManager != "unknown" {
		lines = append(lines, "```bash", repo.PackageManager+" install", "```")
	}
	if len(repo.EntryPoints) > 0 {
		lines = append(lines, "Primary entrypoints: "+strings.Join(trimStringSlice(repo.EntryPoints, 4), ", "))
	}
	if len(lines) == 0 {
		lines = append(lines, "Follow the repository package-manager and entrypoint conventions detected during scanning.")
	}
	return strings.Join(lines, "\n")
}

func fallbackUsage(repo scan.RepoSummary) string {
	switch repo.PackageManager {
	case "go modules":
		return "```bash\ngo run .\n```"
	case "npm", "pnpm", "yarn":
		return "```bash\n" + repo.PackageManager + " run dev\n```"
	default:
		return "Run the primary application command or entrypoint detected in the repository."
	}
}

func commandSection(scripts map[string]string) string {
	keys := make([]string, 0, len(scripts))
	for k := range scripts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	lines := []string{}
	for _, key := range keys {
		lines = append(lines, "- `"+key+"`: "+scripts[key])
	}
	return strings.Join(lines, "\n")
}

func envSection(files []string) string {
	lines := []string{"Environment-related files detected:"}
	for _, file := range files {
		lines = append(lines, "- `"+file+"`")
	}
	return strings.Join(lines, "\n")
}
