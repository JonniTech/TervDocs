package render

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"tervdocs/internal/config"
	"tervdocs/internal/scan"
	"tervdocs/internal/summarize"
)

func Enhance(content string, repo scan.RepoSummary, ctx summarize.Context, cfg config.Config) string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = stripEmoji(content)
	content = ensureTitle(content, repo.ProjectName)
	content = ensureBadgeRow(content, repo)
	content = ensureSections(content, repo, ctx)
	content = normalizeBoldHeadings(content)
	content = injectSectionDividers(content, ctx.BrandColor)
	content = ensureFooter(content, cfg.DeveloperName, ctx.BrandColor)
	return strings.TrimSpace(content) + "\n"
}

func ensureTitle(content, projectName string) string {
	if strings.HasPrefix(strings.TrimSpace(content), "# ") {
		return content
	}
	if projectName == "" {
		projectName = "Project"
	}
	return "# " + displayProjectTitle(projectName) + "\n\n" + strings.TrimSpace(content)
}

func ensureBadgeRow(content string, repo scan.RepoSummary) string {
	if strings.Contains(content, "img.shields.io") {
		return content
	}
	badges := buildBadges(repo)
	if len(badges) == 0 {
		return content
	}
	lines := strings.Split(content, "\n")
	insertAt := 1
	for insertAt < len(lines) && strings.TrimSpace(lines[insertAt]) == "" {
		insertAt++
	}
	block := []string{"", strings.Join(badges, " "), ""}
	lines = append(lines[:insertAt], append(block, lines[insertAt:]...)...)
	return strings.Join(lines, "\n")
}

func ensureSections(content string, repo scan.RepoSummary, ctx summarize.Context) string {
	sections := []string{}
	if !hasSection(content, "aim") {
		sections = append(sections, "## Aim\n\n"+ctx.PurposeGuess)
	}
	if !hasAnySection(content, "problem", "problem statement") {
		sections = append(sections, "## Problem\n\n"+ctx.ProblemStatement)
	}
	if !hasSection(content, "solution") {
		sections = append(sections, "## Solution\n\n"+ctx.SolutionSummary)
	}
	if !hasAnySection(content, "features", "feature highlights") {
		sections = append(sections, "## Features\n\n"+bulletList(featureBullets(ctx)))
	}
	if !hasAnySection(content, "tech stack", "stack") {
		sections = append(sections, "## Tech Stack\n\n"+techStackSection(repo, ctx))
	}
	if !hasAnySection(content, "flow diagram", "architecture diagram") && !strings.Contains(content, "```mermaid") {
		sections = append(sections, "## Flow Diagram\n\n"+mermaidDiagram(repo))
	}
	if !hasAnySection(content, "architecture notes", "architecture") {
		sections = append(sections, "## Architecture Notes\n\n"+bulletList(architectureBullets(ctx)))
	}
	if !hasAnySection(content, "project structure", "structure") {
		sections = append(sections, "## Project Structure\n\n"+projectStructureSection(repo))
	}
	if len(sections) == 0 {
		return content
	}
	return strings.TrimSpace(content) + "\n\n" + strings.Join(sections, "\n\n") + "\n"
}

func injectSectionDividers(content, color string) string {
	if color == "" {
		color = "#4F46E5"
	}
	if strings.Contains(content, `data-tervdocs-divider="true"`) {
		return content
	}
	lines := strings.Split(content, "\n")
	out := make([]string, 0, len(lines)+16)
	for i, line := range lines {
		if strings.HasPrefix(line, "## ") && i > 0 {
			out = append(out, sectionDivider(color))
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

func ensureFooter(content, developerName, color string) string {
	if strings.TrimSpace(developerName) == "" {
		return content
	}
	footerText := "Programmed by " + developerName
	if strings.Contains(content, footerText) {
		return content
	}
	footer := fmt.Sprintf(`<div align="center"><sub style="color:%s;">%s</sub></div>`, color, footerText)
	return strings.TrimSpace(content) + "\n\n" + sectionDivider(color) + "\n" + footer + "\n"
}

func buildBadges(repo scan.RepoSummary) []string {
	items := []string{}
	added := map[string]struct{}{}
	for _, name := range append(append([]string{}, repo.MainLanguages...), repo.Frameworks...) {
		if len(items) >= 5 {
			break
		}
		key := strings.ToLower(name)
		if _, ok := added[key]; ok {
			continue
		}
		added[key] = struct{}{}
		items = append(items, badgeFor(name))
	}
	if repo.PackageManager != "" && repo.PackageManager != "unknown" && len(items) < 6 {
		items = append(items, badgeFor(repo.PackageManager))
	}
	return items
}

func badgeFor(name string) string {
	type meta struct {
		Label string
		Color string
		Logo  string
	}
	m := map[string]meta{
		"go":         {Label: "Go", Color: "00ADD8", Logo: "go"},
		"typescript": {Label: "TypeScript", Color: "3178C6", Logo: "typescript"},
		"javascript": {Label: "JavaScript", Color: "F7DF1E", Logo: "javascript"},
		"python":     {Label: "Python", Color: "3776AB", Logo: "python"},
		"rust":       {Label: "Rust", Color: "CE422B", Logo: "rust"},
		"react":      {Label: "React", Color: "61DAFB", Logo: "react"},
		"next.js":    {Label: "Next.js", Color: "000000", Logo: "nextdotjs"},
		"express":    {Label: "Express", Color: "000000", Logo: "express"},
		"gin":        {Label: "Gin", Color: "008ECF", Logo: "go"},
		"fiber":      {Label: "Fiber", Color: "00AC47", Logo: "go"},
		"docker":     {Label: "Docker", Color: "2496ED", Logo: "docker"},
		"go modules": {Label: "Go Modules", Color: "00ADD8", Logo: "go"},
		"npm":        {Label: "npm", Color: "CB3837", Logo: "npm"},
		"pnpm":       {Label: "pnpm", Color: "F69220", Logo: "pnpm"},
		"yarn":       {Label: "Yarn", Color: "2C8EBB", Logo: "yarn"},
	}
	key := strings.ToLower(name)
	v, ok := m[key]
	if !ok {
		label := url.QueryEscape(name)
		return fmt.Sprintf("![%s](https://img.shields.io/badge/%s-4F46E5?style=for-the-badge)", name, label)
	}
	return fmt.Sprintf("![%s](https://img.shields.io/badge/%s-%s?style=for-the-badge&logo=%s&logoColor=white)", v.Label, url.QueryEscape(v.Label), v.Color, v.Logo)
}

func featureBullets(ctx summarize.Context) []string {
	out := []string{}
	out = append(out, ctx.NotableFeatures...)
	if len(ctx.APIHints) > 0 {
		out = append(out, "API or routing hints detected in the codebase")
	}
	if len(ctx.Scripts) > 0 {
		out = append(out, "Command scripts are available for common development tasks")
	}
	if len(out) == 0 {
		out = append(out, "The repository provides a clear implementation path grounded in the scanned code and configuration files")
	}
	return unique(out)
}

func architectureBullets(ctx summarize.Context) []string {
	out := append([]string{}, ctx.ArchitectureHints...)
	if len(ctx.KeyFiles) > 0 {
		out = append(out, "Key implementation files include "+strings.Join(firstKeyFilePaths(ctx.KeyFiles, 4), ", "))
	}
	if len(out) == 0 {
		out = append(out, "Architecture details should be inferred from the repository structure and primary entrypoints")
	}
	return unique(out)
}

func techStackSection(repo scan.RepoSummary, ctx summarize.Context) string {
	lines := []string{}
	if len(ctx.Languages) > 0 {
		lines = append(lines, "- Languages: "+strings.Join(ctx.Languages, ", "))
	}
	if len(ctx.Frameworks) > 0 {
		lines = append(lines, "- Frameworks: "+strings.Join(ctx.Frameworks, ", "))
	}
	if repo.PackageManager != "" && repo.PackageManager != "unknown" {
		lines = append(lines, "- Package manager: "+repo.PackageManager)
	}
	if len(ctx.Dependencies) > 0 {
		lines = append(lines, "- Key dependencies: "+strings.Join(ctx.Dependencies[:min(len(ctx.Dependencies), 8)], ", "))
	}
	return strings.Join(lines, "\n")
}

func mermaidDiagram(repo scan.RepoSummary) string {
	project := sanitizeMermaidLabel(repo.ProjectName)
	entry := "Entrypoint"
	if len(repo.EntryPoints) > 0 {
		entry = sanitizeMermaidLabel(repo.EntryPoints[0])
	}
	folders := trimStringSlice(repo.TopFolders, 4)
	if len(folders) == 0 {
		folders = []string{"src"}
	}
	var b strings.Builder
	b.WriteString("```mermaid\n")
	b.WriteString("flowchart TD\n")
	b.WriteString("    User[Developer] --> Entry[" + entry + "]\n")
	b.WriteString("    Entry --> Core[" + project + "]\n")
	for _, folder := range folders {
		b.WriteString("    Core --> " + mermaidID(folder) + "[" + sanitizeMermaidLabel(folder) + "/]\n")
	}
	if repo.UsesDocker {
		b.WriteString("    Core --> Docker[Docker Workflow]\n")
	}
	if len(repo.CIConfigs) > 0 {
		b.WriteString("    Core --> CI[CI Pipeline]\n")
	}
	b.WriteString("```\n")
	return b.String()
}

func projectStructureSection(repo scan.RepoSummary) string {
	keys := make([]string, 0, len(repo.FolderFileCounts))
	for k := range repo.FolderFileCounts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if len(keys) == 0 {
		keys = append(keys, ".")
	}
	var b strings.Builder
	b.WriteString("```text\n")
	for _, key := range trimStringSlice(keys, 8) {
		if key == "." || key == "" {
			continue
		}
		b.WriteString(fmt.Sprintf("%s/  (%d files)\n", key, repo.FolderFileCounts[key]))
	}
	b.WriteString("```\n")
	return b.String()
}

func sectionDivider(color string) string {
	return fmt.Sprintf(`<div align="center" data-tervdocs-divider="true"><img src="%s" alt="section divider" /></div>`, DividerAssetRelativePath())
}

func hasSection(content, name string) bool {
	re := regexp.MustCompile(`(?mi)^##\s+(?:\*\*)?` + regexp.QuoteMeta(name) + `(?:\*\*)?\s*$`)
	return re.MatchString(content)
}

func hasAnySection(content string, names ...string) bool {
	for _, name := range names {
		if hasSection(content, name) {
			return true
		}
	}
	return false
}

func bulletList(items []string) string {
	if len(items) == 0 {
		return "- Additional details should be filled in from the scanned repository context."
	}
	return "- " + strings.Join(items, "\n- ")
}

func stripEmoji(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 0x1F300 && r <= 0x1FAFF:
			continue
		case r >= 0x2600 && r <= 0x27BF:
			continue
		case r == 0xFE0F:
			continue
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func unique(in []string) []string {
	out := []string{}
	seen := map[string]struct{}{}
	for _, item := range in {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func firstKeyFilePaths(in []summarize.KeyFile, max int) []string {
	out := []string{}
	for _, item := range in {
		if len(out) >= max {
			break
		}
		out = append(out, item.Path)
	}
	return out
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func sanitizeMermaidLabel(s string) string {
	s = strings.ReplaceAll(s, `"`, "")
	s = strings.ReplaceAll(s, "`", "")
	if s == "" {
		return "Module"
	}
	return s
}

func mermaidID(s string) string {
	replacer := strings.NewReplacer("/", "_", "-", "_", ".", "_", " ", "_")
	out := replacer.Replace(s)
	if out == "" {
		return "node"
	}
	return out
}

func trimStringSlice(in []string, max int) []string {
	if len(in) <= max {
		return in
	}
	return in[:max]
}

func normalizeBoldHeadings(content string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			title := strings.TrimSpace(strings.TrimPrefix(trimmed, "# "))
			lines[i] = "# " + ensureBoldHeadingText(displayProjectTitle(stripMarkdownBold(title)))
			continue
		}
		if strings.HasPrefix(trimmed, "## ") {
			title := strings.TrimSpace(strings.TrimPrefix(trimmed, "## "))
			lines[i] = "## " + ensureBoldHeadingText(stripMarkdownBold(title))
		}
	}
	return strings.Join(lines, "\n")
}

func ensureBoldHeadingText(text string) string {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "**") && strings.HasSuffix(text, "**") {
		return text
	}
	return "**" + text + "**"
}

func stripMarkdownBold(text string) string {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "**") && strings.HasSuffix(text, "**") && len(text) >= 4 {
		return strings.TrimSuffix(strings.TrimPrefix(text, "**"), "**")
	}
	return text
}

func displayProjectTitle(projectName string) string {
	switch strings.ToLower(strings.TrimSpace(projectName)) {
	case "tervdocs":
		return "TervDocs"
	case "":
		return "Project"
	default:
		return projectName
	}
}
