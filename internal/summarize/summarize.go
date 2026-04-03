package summarize

import (
	"encoding/json"
	"strings"

	"tervdocs/internal/scan"
)

type Context struct {
	ProjectMetadata map[string]any    `json:"project_metadata"`
	PurposeGuess    string            `json:"project_purpose_guess"`
	Languages       []string          `json:"languages"`
	Frameworks      []string          `json:"frameworks"`
	FolderMap       []string          `json:"folder_map"`
	Scripts         map[string]string `json:"scripts"`
	EnvVars         []string          `json:"env_vars"`
	SetupHints      []string          `json:"setup_hints"`
	APIHints        []string          `json:"api_hints"`
	NotableFeatures []string          `json:"notable_features"`
	TestingInfo     []string          `json:"testing_info"`
	DeploymentHints []string          `json:"deployment_hints"`
}

func Build(repo scan.RepoSummary) Context {
	return Context{
		ProjectMetadata: map[string]any{
			"name":            repo.ProjectName,
			"files_scanned":   repo.FilesScanned,
			"bytes_scanned":   repo.BytesScanned,
			"uses_docker":     repo.UsesDocker,
			"package_manager": repo.PackageManager,
		},
		PurposeGuess:    guessPurpose(repo),
		Languages:       trimList(repo.MainLanguages, 6),
		Frameworks:      trimList(repo.Frameworks, 6),
		FolderMap:       trimList(repo.TopFolders, 12),
		Scripts:         repo.Scripts,
		EnvVars:         repo.EnvFiles,
		SetupHints:      setupHints(repo),
		APIHints:        trimList(repo.RouteHints, 12),
		NotableFeatures: notableFeatures(repo),
		TestingInfo:     repo.TestSetup,
		DeploymentHints: deploymentHints(repo),
	}
}

func CompactJSON(ctx Context) string {
	b, _ := json.MarshalIndent(ctx, "", "  ")
	return string(b)
}

func guessPurpose(repo scan.RepoSummary) string {
	if len(repo.Frameworks) > 0 && strings.EqualFold(repo.PackageManager, "go modules") {
		return "A Go-based application with framework integrations."
	}
	if len(repo.RouteHints) > 0 {
		return "A project that exposes API routes and service logic."
	}
	if len(repo.EntryPoints) > 0 {
		return "A software project with runnable application entrypoints."
	}
	return "A software project requiring clear setup and usage documentation."
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

func trimList(in []string, max int) []string {
	if len(in) <= max {
		return in
	}
	return in[:max]
}
