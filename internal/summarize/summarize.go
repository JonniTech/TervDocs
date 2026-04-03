package summarize

import (
	"encoding/json"
	"path/filepath"
	"slices"
	"strings"

	"tervdocs/internal/scan"
)

type KeyFile struct {
	Path    string `json:"path"`
	Snippet string `json:"snippet"`
}

type Context struct {
	ProjectMetadata   map[string]any    `json:"project_metadata"`
	DeveloperName     string            `json:"developer_name"`
	PrimaryLanguage   string            `json:"primary_language"`
	BrandColor        string            `json:"brand_color"`
	PurposeGuess      string            `json:"project_purpose_guess"`
	ProblemStatement  string            `json:"problem_statement"`
	SolutionSummary   string            `json:"solution_summary"`
	Languages         []string          `json:"languages"`
	Frameworks        []string          `json:"frameworks"`
	Dependencies      []string          `json:"dependencies"`
	FolderMap         []string          `json:"folder_map"`
	FolderFileCounts  map[string]int    `json:"folder_file_counts"`
	Scripts           map[string]string `json:"scripts"`
	EnvVars           []string          `json:"env_vars"`
	ConfigFiles       []string          `json:"config_files"`
	SetupHints        []string          `json:"setup_hints"`
	APIHints          []string          `json:"api_hints"`
	NotableFeatures   []string          `json:"notable_features"`
	TestingInfo       []string          `json:"testing_info"`
	DeploymentHints   []string          `json:"deployment_hints"`
	PurposeEvidence   []string          `json:"purpose_evidence"`
	ArchitectureHints []string          `json:"architecture_hints"`
	KeyFiles          []KeyFile         `json:"key_files"`
}

func Build(repo scan.RepoSummary, developerName string) Context {
	return Context{
		ProjectMetadata: map[string]any{
			"name":             repo.ProjectName,
			"files_scanned":    repo.FilesScanned,
			"bytes_scanned":    repo.BytesScanned,
			"uses_docker":      repo.UsesDocker,
			"package_manager":  repo.PackageManager,
			"entry_points":     repo.EntryPoints,
			"top_folders":      repo.TopFolders,
			"readme_preexists": repo.ReadmeExists,
		},
		DeveloperName:     developerName,
		PrimaryLanguage:   repo.PrimaryLanguage,
		BrandColor:        brandColor(repo.PrimaryLanguage),
		PurposeGuess:      guessPurpose(repo),
		ProblemStatement:  guessProblem(repo),
		SolutionSummary:   guessSolution(repo),
		Languages:         trimList(repo.MainLanguages, 6),
		Frameworks:        trimList(repo.Frameworks, 6),
		Dependencies:      trimList(repo.Dependencies, 16),
		FolderMap:         trimList(repo.TopFolders, 12),
		FolderFileCounts:  repo.FolderFileCounts,
		Scripts:           repo.Scripts,
		EnvVars:           repo.EnvFiles,
		ConfigFiles:       trimList(repo.ConfigFiles, 12),
		SetupHints:        setupHints(repo),
		APIHints:          trimList(repo.RouteHints, 12),
		NotableFeatures:   notableFeatures(repo),
		TestingInfo:       repo.TestSetup,
		DeploymentHints:   deploymentHints(repo),
		PurposeEvidence:   trimList(repo.PurposeHints, 8),
		ArchitectureHints: trimList(repo.ArchitectureHints, 10),
		KeyFiles:          pickKeyFiles(repo.ImportantFiles),
	}
}

func CompactJSON(ctx Context) string {
	b, _ := json.MarshalIndent(ctx, "", "  ")
	return string(b)
}

func guessPurpose(repo scan.RepoSummary) string {
	if len(repo.PurposeHints) > 0 {
		return repo.PurposeHints[0]
	}
	if len(repo.Frameworks) > 0 && strings.EqualFold(repo.PackageManager, "go modules") {
		return "A Go-based application with framework integrations and a structured developer workflow."
	}
	if len(repo.RouteHints) > 0 {
		return "A project that exposes API routes and service logic."
	}
	if len(repo.EntryPoints) > 0 {
		return "A software project with runnable application entrypoints and documented workflows."
	}
	return "A software project requiring clear setup, architecture, and usage documentation."
}

func guessProblem(repo scan.RepoSummary) string {
	if len(repo.RouteHints) > 0 {
		return "Developers and collaborators need a clear way to understand how requests move through the codebase and how the main runtime pieces fit together."
	}
	if len(repo.EntryPoints) > 0 {
		return "New contributors need faster onboarding into the project's purpose, setup steps, and runtime flow without reading the entire codebase first."
	}
	return "The project needs documentation that explains its purpose, structure, and development workflow in a grounded way."
}

func guessSolution(repo scan.RepoSummary) string {
	parts := []string{}
	if len(repo.EntryPoints) > 0 {
		parts = append(parts, "The codebase exposes clear entrypoints such as "+strings.Join(trimList(repo.EntryPoints, 3), ", "))
	}
	if len(repo.TopFolders) > 0 {
		parts = append(parts, "major concerns are split across "+strings.Join(trimList(repo.TopFolders, 6), ", "))
	}
	if repo.UsesDocker {
		parts = append(parts, "container tooling supports consistent environments")
	}
	if len(parts) == 0 {
		return "The solution is implemented through the repository's source layout, scripts, and supporting configuration."
	}
	return strings.Join(parts, "; ") + "."
}

func setupHints(repo scan.RepoSummary) []string {
	out := []string{}
	if repo.PackageManager != "unknown" {
		out = append(out, "Install dependencies using "+repo.PackageManager)
	}
	if len(repo.EntryPoints) > 0 {
		out = append(out, "Primary entry files: "+strings.Join(trimList(repo.EntryPoints, 4), ", "))
	}
	if repo.UsesDocker {
		out = append(out, "Docker is available for local or deployment workflows")
	}
	if len(repo.ConfigFiles) > 0 {
		out = append(out, "Important configuration files: "+strings.Join(trimList(repo.ConfigFiles, 4), ", "))
	}
	return out
}

func notableFeatures(repo scan.RepoSummary) []string {
	out := []string{}
	if len(repo.CIConfigs) > 0 {
		out = append(out, "CI workflows present")
	}
	if len(repo.MonorepoHints) > 0 {
		out = append(out, "Monorepo signals detected")
	}
	if len(repo.RouteHints) > 0 {
		out = append(out, "API routes or routing files detected")
	}
	if len(repo.Scripts) > 0 {
		out = append(out, "Runnable scripts and commands are defined in the repository")
	}
	if len(repo.Dependencies) > 0 {
		out = append(out, "The stack includes explicit third-party dependencies and tooling")
	}
	return out
}

func deploymentHints(repo scan.RepoSummary) []string {
	out := []string{}
	if repo.UsesDocker {
		out = append(out, "Container-based deployment likely supported")
	}
	if len(repo.CIConfigs) > 0 {
		out = append(out, "Automated workflows exist under .github/workflows")
	}
	return out
}

func pickKeyFiles(files []scan.FileSummary) []KeyFile {
	candidates := make([]scan.FileSummary, len(files))
	copy(candidates, files)
	slices.SortFunc(candidates, func(a, b scan.FileSummary) int {
		return filePriority(b.Path) - filePriority(a.Path)
	})
	out := []KeyFile{}
	for _, f := range candidates {
		if len(out) >= 12 {
			break
		}
		if strings.TrimSpace(f.Snippet) == "" {
			continue
		}
		out = append(out, KeyFile{Path: f.Path, Snippet: f.Snippet})
	}
	return out
}

func filePriority(path string) int {
	base := strings.ToLower(filepath.Base(path))
	switch base {
	case "main.go", "package.json", "go.mod", "dockerfile", "makefile":
		return 10
	}
	if strings.HasPrefix(path, "cmd/") {
		return 9
	}
	if strings.HasPrefix(path, "internal/") {
		return 8
	}
	if strings.Contains(strings.ToLower(path), "route") || strings.Contains(strings.ToLower(path), "/api/") {
		return 7
	}
	return 1
}

func brandColor(language string) string {
	switch strings.ToLower(language) {
	case "go":
		return "#00ADD8"
	case "typescript":
		return "#3178C6"
	case "javascript":
		return "#F7DF1E"
	case "python":
		return "#3776AB"
	case "rust":
		return "#CE422B"
	case "java":
		return "#007396"
	case "kotlin":
		return "#7F52FF"
	case "swift":
		return "#FA7343"
	case "php":
		return "#777BB4"
	case "ruby":
		return "#CC342D"
	case "c#":
		return "#512BD4"
	default:
		return "#4F46E5"
	}
}

func trimList[T any](in []T, max int) []T {
	if len(in) <= max {
		return in
	}
	return in[:max]
}
