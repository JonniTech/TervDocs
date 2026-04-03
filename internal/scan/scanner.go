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

		if !s.shouldIncludeFile(rel, includePrefixes) {
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
		if len(out.ImportantFiles) >= s.cfg.MaxFiles {
			return errors.New("repo too large: file limit reached")
		}
		if int(info.Size()) > s.cfg.MaxBytesPerFile {
			return nil
		}

		contents, err := os.ReadFile(path)
		if err != nil {
			out.Warnings = append(out.Warnings, "cannot read "+rel)
			return nil
		}

		out.FilesScanned++
		out.BytesScanned += int64(len(contents))
		snippet := compactSnippet(string(contents), 320)
		out.ImportantFiles = append(out.ImportantFiles, FileSummary{Path: rel, Size: int64(len(contents)), Snippet: snippet})

		if lang, ok := languageByExt[ext]; ok {
			langCount[lang]++
		}

		s.detectByNameAndContent(rel, string(contents), frameworkSet, testSet, routeSet, monorepoIndicators, &out)
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
	case "dockerfile", "docker-compose.yml", "docker-compose.yaml":
		out.UsesDocker = true
	case "pnpm-workspace.yaml", "lerna.json", "turbo.json":
		monorepoIndicators[name]++
	case "go.mod", "cargo.toml", "pyproject.toml", "requirements.txt", "package.json", "makefile":
		out.EntryPoints = appendUnique(out.EntryPoints, rel)
	}

	if strings.HasPrefix(rel, ".github/workflows/") {
		out.CIConfigs = appendUnique(out.CIConfigs, rel)
	}
	if strings.HasPrefix(name, ".env") {
		out.EnvFiles = appendUnique(out.EnvFiles, rel)
	}
	if strings.EqualFold(name, "readme.md") {
		out.ReadmeExists = true
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

	if strings.HasSuffix(rel, "package.json") {
		type pkgJSON struct {
			Scripts map[string]string `json:"scripts"`
		}
		var p pkgJSON
		if json.Unmarshal([]byte(body), &p) == nil {
			for k, v := range p.Scripts {
				out.Scripts[k] = v
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
		sc := bufio.NewScanner(strings.NewReader(body))
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if strings.HasPrefix(line, "module ") {
				out.ProjectName = strings.TrimSpace(strings.TrimPrefix(line, "module "))
				break
			}
		}
	}

	httpRE := regexp.MustCompile(`\b(GET|POST|PUT|PATCH|DELETE)\s+(/[^\s"']+)`)
	for _, m := range httpRE.FindAllStringSubmatch(body, 8) {
		routeSet[m[2]] = struct{}{}
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
