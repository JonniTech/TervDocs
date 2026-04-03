package scan

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tervdocs/internal/config"
)

func TestScanDetectsSignals(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "go.mod"), "module github.com/acme/demo\n")
	mustWrite(t, filepath.Join(root, "main.go"), "// Demo CLI generates docs\npackage main\nfunc main(){}\n")
	mustWrite(t, filepath.Join(root, "README.md"), "# Old readme\n")
	mustWrite(t, filepath.Join(root, "README.md.bak.1"), "# backup\n")
	mustWrite(t, filepath.Join(root, "Dockerfile"), "FROM golang:1.26\n")
	mustWrite(t, filepath.Join(root, ".github/workflows/ci.yml"), "name: ci\n")
	mustWrite(t, filepath.Join(root, ".env.example"), "PORT=8080\n")

	sc := New(config.Default().Scan)
	sum, err := sc.Scan(context.Background(), root)
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}
	if !sum.UsesDocker {
		t.Fatalf("expected docker detection")
	}
	if sum.ProjectName != "github.com/acme/demo" {
		t.Fatalf("expected module-derived project name, got %s", sum.ProjectName)
	}
	if len(sum.CIConfigs) == 0 {
		t.Fatalf("expected ci configs")
	}
	if !sum.ReadmeExists {
		t.Fatalf("expected README existence to be detected")
	}
	for _, file := range sum.ImportantFiles {
		if strings.Contains(strings.ToLower(file.Path), "readme") {
			t.Fatalf("readme artifacts should not be used as key context files: %s", file.Path)
		}
	}
	if len(sum.PurposeHints) == 0 {
		t.Fatalf("expected purpose hints from code comments")
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
}
