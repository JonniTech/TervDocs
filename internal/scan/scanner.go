package scan

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"tervdocs/internal/config"
)

var defaultIgnores = map[string]struct{}{
	".git": {}, "node_modules": {}, "dist": {}, "build": {}, "vendor": {}, "coverage": {},
	".next": {}, ".turbo": {}, "bin": {}, ".cache": {}, ".pnpm-store": {}, ".yarn": {},
	".idea": {}, ".vscode": {}, ".qoder": {}, ".codex": {},
}

var binaryLikeExt = map[string]struct{}{
	".png": {}, ".jpg": {}, ".jpeg": {}, ".gif": {}, ".webp": {}, ".ico": {}, ".pdf": {},
	".zip": {}, ".gz": {}, ".tar": {}, ".mp4": {}, ".mov": {}, ".avi": {}, ".exe": {},
	".dll": {}, ".so": {}, ".dylib": {}, ".woff": {}, ".woff2": {},
}

var languageByExt = map[string]string{
	".go": "Go", ".ts": "TypeScript", ".tsx": "TypeScript", ".js": "JavaScript", ".jsx": "JavaScript",
	".py": "Python", ".rs": "Rust", ".java": "Java", ".kt": "Kotlin", ".swift": "Swift",
	".php": "PHP", ".rb": "Ruby", ".cs": "C#", ".c": "C", ".cpp": "C++", ".h": "C/C++",
}

type Scanner struct {
	cfg config.ScanConfig
}

func New(cfg config.ScanConfig) *Scanner {
	return &Scanner{cfg: cfg}
}

func (s *Scanner) Scan(ctx context.Context, root string) (RepoSummary, error) {
	out := RepoSummary{
		Root:              root,
		Scripts:           map[string]string{},
		FolderFileCounts:  map[string]int{},
		ExcludedByDefault: mapKeys(defaultIgnores),
	}

	if root == "" {
		root = "."
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return out, err
	}
	out.Root = absRoot
	out.ProjectName = filepath.Base(absRoot)

	langCount := map[string]int{}
	frameworkSet := map[string]struct{}{}
	testSet := map[string]struct{}{}
	routeSet := map[string]struct{}{}
	topFolders := map[string]struct{}{}
	monorepoIndicators := map[string]int{}

	includePrefixes := normalizePaths(s.cfg.Include)
	excludes := buildExcludes(s.cfg.Exclude)

	err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			out.Warnings = append(out.Warnings, walkErr.Error())
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		rel, err := filepath.Rel(absRoot, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}

		if d.IsDir() {
			base := d.Name()
			if shouldIgnoreDir(base, rel, excludes, includePrefixes) {
				return filepath.SkipDir
			}
			if parts := strings.Split(rel, "/"); len(parts) > 0 {
				topFolders[parts[0]] = struct{}{}
			}
			return nil
		}

		baseLower := strings.ToLower(filepath.Base(rel))
		if isReadmeName(baseLower) {
			out.ReadmeExists = true
			return nil
		}
		if isGeneratedArtifact(rel, baseLower) {
			return nil
		}
		if !s.shouldIncludeFile(rel, includePrefixes) && !isAlwaysRelevantFile(rel) {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(d.Name()))
		if _, skip := binaryLikeExt[ext]; skip {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			out.Warnings = append(out.Warnings, "cannot stat "+rel)
			return nil
		}
		if out.FilesScanned >= s.cfg.MaxFiles {
			return errors.New("repo too large: file limit reached")
		}
		if int(info.Size()) > s.cfg.MaxBytesPerFile && !isAlwaysRelevantFile(rel) {
			return nil
		}

		contents, err := os.ReadFile(path)
		if err != nil {
			out.Warnings = append(out.Warnings, "cannot read "+rel)
			return nil
		}

		body := string(contents)
		out.FilesScanned++
		out.BytesScanned += int64(len(contents))
		incrementFolderCount(out.FolderFileCounts, rel)

		if lang, ok := languageByExt[ext]; ok {
			langCount[lang]++
		}
		if shouldCaptureSnippet(rel) {
			out.ImportantFiles = append(out.ImportantFiles, FileSummary{
				Path:    rel,
				Size:    int64(len(contents)),
				Snippet: compactSnippet(body, snippetLimitFor(rel)),
			})
		}

		s.detectByNameAndContent(rel, body, frameworkSet, testSet, routeSet, monorepoIndicators, &out)
		return nil
	})
	if err != nil {
		if strings.Contains(err.Error(), "file limit reached") {
			return out, err
		}
		if errors.Is(err, context.Canceled) {
			return out, err
		}
	}

	out.MainLanguages = sortedByCount(langCount)
	out.Frameworks = setToSorted(frameworkSet)
	out.TestSetup = setToSorted(testSet)
	out.RouteHints = setToSorted(routeSet)
	out.TopFolders = setToSorted(topFolders)
	out.MonorepoHints = monorepoHints(monorepoIndicators)
	out.Dependencies = trimList(setToSorted(stringSliceToSet(out.Dependencies)), 20)
	out.ConfigFiles = trimList(setToSorted(stringSliceToSet(out.ConfigFiles)), 20)
	out.PurposeHints = trimList(uniqueStrings(out.PurposeHints), 10)
	out.ArchitectureHints = trimList(uniqueStrings(append(out.ArchitectureHints, deriveArchitectureHints(out)...)), 12)

	if len(out.MainLanguages) > 0 {
		out.PrimaryLanguage = out.MainLanguages[0]
	}
	if len(out.EntryPoints) == 0 {
		out.EntryPoints = guessEntryPoints(out.ImportantFiles)
	}
	if out.PackageManager == "" {
		out.PackageManager = "unknown"
	}
	return out, nil
}

func (s *Scanner) shouldIncludeFile(rel string, include []string) bool {
	if len(include) == 0 {
		return true
	}
	for _, p := range include {
		if strings.HasPrefix(rel, p+"/") || rel == p {
			return true
		}
	}
	return false
}

func (s *Scanner) detectByNameAndContent(rel, body string, frameworkSet, testSet, routeSet map[string]struct{}, monorepoIndicators map[string]int, out *RepoSummary) {
	name := strings.ToLower(filepath.Base(rel))
	switch name {
	case "dockerfile", "docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml":
		out.UsesDocker = true
		out.ConfigFiles = appendUnique(out.ConfigFiles, rel)
	case "pnpm-workspace.yaml", "lerna.json", "turbo.json":
		monorepoIndicators[name]++
		out.ConfigFiles = appendUnique(out.ConfigFiles, rel)
	case "go.mod", "cargo.toml", "pyproject.toml", "requirements.txt", "package.json", "makefile":
		out.EntryPoints = appendUnique(out.EntryPoints, rel)
	}

	if isConfigFile(rel) {
		out.ConfigFiles = appendUnique(out.ConfigFiles, rel)
	}
	if strings.HasPrefix(rel, ".github/workflows/") {
		out.CIConfigs = appendUnique(out.CIConfigs, rel)
	}
	if strings.HasPrefix(name, ".env") {
		out.EnvFiles = appendUnique(out.EnvFiles, rel)
	}
	if strings.Contains(name, "test") || strings.Contains(rel, "/test") || strings.Contains(rel, "__tests__") {
		testSet["tests-detected"] = struct{}{}
	}
	if strings.HasSuffix(name, "_test.go") {
		testSet["go test"] = struct{}{}
	}
	if strings.Contains(strings.ToLower(rel), "route") || strings.Contains(strings.ToLower(rel), "router") || strings.Contains(strings.ToLower(rel), "/api/") {
		routeSet[rel] = struct{}{}
	}

	lower := strings.ToLower(body)
	if strings.Contains(lower, "next") && strings.Contains(rel, "package.json") {
		frameworkSet["Next.js"] = struct{}{}
	}
	if strings.Contains(lower, "\"react\"") && strings.Contains(rel, "package.json") {
		frameworkSet["React"] = struct{}{}
	}
	if strings.Contains(lower, "\"vue\"") && strings.Contains(rel, "package.json") {
		frameworkSet["Vue"] = struct{}{}
	}
	if strings.Contains(lower, "express") {
		frameworkSet["Express"] = struct{}{}
	}
	if strings.Contains(lower, "gin-gonic/gin") {
		frameworkSet["Gin"] = struct{}{}
	}
	if strings.Contains(lower, "fiber") {
		frameworkSet["Fiber"] = struct{}{}
	}
	if strings.Contains(lower, "django") {
		frameworkSet["Django"] = struct{}{}
	}
	if strings.Contains(lower, "flask") {
		frameworkSet["Flask"] = struct{}{}
	}
	if strings.Contains(lower, "cobra") {
		out.Dependencies = appendUnique(out.Dependencies, "cobra")
	}
	if strings.Contains(lower, "viper") {
		out.Dependencies = appendUnique(out.Dependencies, "viper")
	}

	if strings.HasSuffix(rel, "package.json") {
		type pkgJSON struct {
			Name            string            `json:"name"`
			Description     string            `json:"description"`
			Scripts         map[string]string `json:"scripts"`
			Dependencies    map[string]string `json:"dependencies"`
			DevDependencies map[string]string `json:"devDependencies"`
		}
		var p pkgJSON
		if json.Unmarshal([]byte(body), &p) == nil {
			for k, v := range p.Scripts {
				out.Scripts[k] = v
			}
			if p.Description != "" {
				out.PurposeHints = append(out.PurposeHints, p.Description)
			}
			if p.Name != "" && out.ProjectName == filepath.Base(out.Root) {
				out.ProjectName = p.Name
			}
			for dep := range p.Dependencies {
				out.Dependencies = appendUnique(out.Dependencies, dep)
			}
			for dep := range p.DevDependencies {
				out.Dependencies = appendUnique(out.Dependencies, dep)
			}
		}
		if strings.Contains(lower, "\"pnpm\"") {
			out.PackageManager = "pnpm"
		} else if strings.Contains(lower, "\"yarn\"") {
			out.PackageManager = "yarn"
		} else if out.PackageManager == "" {
			out.PackageManager = "npm"
		}
		monorepoIndicators["package.json"]++
	}
	if strings.HasSuffix(rel, "go.mod") {
		if out.PackageManager == "" {
			out.PackageManager = "go modules"
		}
		out.Dependencies = append(out.Dependencies, parseGoModDependencies(body)...)
		sc := bufio.NewScanner(strings.NewReader(body))
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if strings.HasPrefix(line, "module ") {
				out.ProjectName = strings.TrimSpace(strings.TrimPrefix(line, "module "))
				break
			}
		}
	}

	if comment := extractLeadingComment(rel, body); comment != "" {
		out.PurposeHints = append(out.PurposeHints, comment)
	}

	httpRE := regexp.MustCompile(`\b(GET|POST|PUT|PATCH|DELETE)\s+(/[^\s"']+)`)
	for _, m := range httpRE.FindAllStringSubmatch(body, 8) {
		routeSet[m[2]] = struct{}{}
	}
	routerRE := regexp.MustCompile(`(?i)(?:handlefunc|get|post|put|patch|delete)\(\s*"([^"]+)"`)
	for _, m := range routerRE.FindAllStringSubmatch(body, 8) {
		routeSet[m[1]] = struct{}{}
	}

	if strings.HasPrefix(rel, "cmd/") {
		out.ArchitectureHints = append(out.ArchitectureHints, "CLI entry commands are organized under cmd/")
	}
	if strings.HasPrefix(rel, "internal/") {
		out.ArchitectureHints = append(out.ArchitectureHints, "Core application logic is separated under internal/")
	}
	if strings.HasPrefix(rel, "pkg/") {
		out.ArchitectureHints = append(out.ArchitectureHints, "Reusable packages exist under pkg/")
	}
}

func buildExcludes(extra []string) map[string]struct{} {
	out := map[string]struct{}{}
	for k := range defaultIgnores {
		out[k] = struct{}{}
	}
	for _, e := range extra {
		out[strings.TrimSpace(e)] = struct{}{}
	}
	return out
}

func shouldIgnoreDir(base, rel string, excludes map[string]struct{}, includes []string) bool {
	if len(includes) > 0 {
		for _, i := range includes {
			if strings.HasPrefix(i, rel+"/") || rel == i || strings.HasPrefix(rel, i+"/") {
				return false
			}
		}
	}
	_, okBase := excludes[base]
	_, okRel := excludes[rel]
	return okBase || okRel
}

func normalizePaths(in []string) []string {
	out := make([]string, 0, len(in))
	for _, p := range in {
		p = filepath.ToSlash(strings.TrimSpace(p))
		if p != "" {
			out = append(out, strings.Trim(p, "/"))
		}
	}
	return out
}

func compactSnippet(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func mapKeys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	slices.Sort(out)
	return out
}

func setToSorted(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	slices.Sort(out)
	return out
}

func sortedByCount(m map[string]int) []string {
	type kv struct {
		k string
		v int
	}
	arr := make([]kv, 0, len(m))
	for k, v := range m {
		arr = append(arr, kv{k: k, v: v})
	}
	slices.SortFunc(arr, func(a, b kv) int {
		if a.v == b.v {
			if a.k < b.k {
				return -1
			}
			if a.k > b.k {
				return 1
			}
			return 0
		}
		if a.v > b.v {
			return -1
		}
		return 1
	})
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		out = append(out, item.k)
	}
	return out
}

func appendUnique(s []string, v string) []string {
	for _, item := range s {
		if item == v {
			return s
		}
	}
	return append(s, v)
}

func guessEntryPoints(files []FileSummary) []string {
	picks := []string{}
	for _, f := range files {
		name := strings.ToLower(filepath.Base(f.Path))
		switch name {
		case "main.go", "index.ts", "index.js", "app.py", "manage.py":
			picks = append(picks, f.Path)
		}
	}
	return picks
}

func monorepoHints(indicators map[string]int) []string {
	out := []string{}
	if indicators["pnpm-workspace.yaml"] > 0 {
		out = append(out, "pnpm workspace detected")
	}
	if indicators["turbo.json"] > 0 {
		out = append(out, "turbo repository detected")
	}
	if indicators["package.json"] > 1 {
		out = append(out, "multiple package.json files detected")
	}
	return out
}

func isAlwaysRelevantFile(rel string) bool {
	rel = filepath.ToSlash(rel)
	base := strings.ToLower(filepath.Base(rel))
	switch base {
	case "package.json", "go.mod", "go.sum", "cargo.toml", "pyproject.toml", "requirements.txt",
		"dockerfile", "docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml",
		"makefile", ".env.example", ".env", "tsconfig.json":
		return true
	}
	if strings.HasPrefix(rel, ".github/workflows/") {
		return true
	}
	return strings.HasPrefix(rel, "cmd/") || strings.HasPrefix(rel, "internal/") || strings.HasPrefix(rel, "src/")
}

func shouldCaptureSnippet(rel string) bool {
	base := strings.ToLower(filepath.Base(rel))
	if strings.HasPrefix(base, ".env") {
		return false
	}
	return isAlwaysRelevantFile(rel) || strings.Contains(strings.ToLower(rel), "route") || strings.Contains(strings.ToLower(rel), "/api/")
}

func snippetLimitFor(rel string) int {
	if isAlwaysRelevantFile(rel) {
		return 1200
	}
	return 600
}

func isGeneratedArtifact(rel, baseLower string) bool {
	if strings.Contains(baseLower, ".bak.") {
		return true
	}
	if baseLower == ".tervdocs.toml" {
		return true
	}
	if strings.HasPrefix(baseLower, "readme.") {
		return true
	}
	return strings.Contains(strings.ToLower(rel), "/dist/")
}

func isReadmeName(baseLower string) bool {
	return baseLower == "readme.md" || baseLower == "readme"
}

func isConfigFile(rel string) bool {
	base := strings.ToLower(filepath.Base(rel))
	ext := strings.ToLower(filepath.Ext(base))
	if strings.HasPrefix(base, ".env") {
		return true
	}
	if base == "makefile" || strings.HasPrefix(rel, ".github/workflows/") {
		return true
	}
	return strings.Contains(base, "config") || ext == ".yaml" || ext == ".yml" || ext == ".toml"
}

func parseGoModDependencies(body string) []string {
	out := []string{}
	sc := bufio.NewScanner(strings.NewReader(body))
	inRequireBlock := false
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "require (" {
			inRequireBlock = true
			continue
		}
		if inRequireBlock && line == ")" {
			break
		}
		if strings.HasPrefix(line, "require ") {
			parts := strings.Fields(strings.TrimPrefix(line, "require "))
			if len(parts) > 0 {
				out = append(out, parts[0])
			}
			continue
		}
		if inRequireBlock {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				out = append(out, parts[0])
			}
		}
	}
	return out
}

func extractLeadingComment(rel, body string) string {
	ext := strings.ToLower(filepath.Ext(rel))
	commentPrefix := ""
	switch ext {
	case ".go", ".js", ".ts", ".tsx", ".jsx", ".java", ".rs", ".c", ".cpp", ".h", ".cs":
		commentPrefix = "//"
	case ".py", ".rb", ".sh":
		commentPrefix = "#"
	}
	if commentPrefix == "" {
		return ""
	}
	lines := strings.Split(body, "\n")
	parts := []string{}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if len(parts) > 0 {
				break
			}
			continue
		}
		if !strings.HasPrefix(trimmed, commentPrefix) {
			break
		}
		text := strings.TrimSpace(strings.TrimPrefix(trimmed, commentPrefix))
		if text != "" {
			parts = append(parts, text)
		}
		if len(strings.Join(parts, " ")) > 240 {
			break
		}
	}
	return strings.Join(parts, " ")
}

func incrementFolderCount(counts map[string]int, rel string) {
	parts := strings.Split(rel, "/")
	if len(parts) > 0 {
		counts[parts[0]]++
	}
}

func deriveArchitectureHints(repo RepoSummary) []string {
	out := []string{}
	if repo.FolderFileCounts["cmd"] > 0 && repo.FolderFileCounts["internal"] > 0 {
		out = append(out, "The repository separates command entrypoints from internal application logic")
	}
	if repo.FolderFileCounts["pkg"] > 0 {
		out = append(out, "Shared reusable packages are exposed from pkg/")
	}
	if repo.UsesDocker {
		out = append(out, "Container workflows are part of the developer or deployment flow")
	}
	if len(repo.CIConfigs) > 0 {
		out = append(out, "Continuous integration is configured through GitHub workflows")
	}
	return out
}

func trimList(in []string, max int) []string {
	if len(in) <= max {
		return in
	}
	return in[:max]
}

func uniqueStrings(in []string) []string {
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

func stringSliceToSet(in []string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, item := range in {
		item = strings.TrimSpace(item)
		if item != "" {
			out[item] = struct{}{}
		}
	}
	return out
}
