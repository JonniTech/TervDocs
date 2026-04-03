package doctor

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tervdocs/internal/config"
)

func TestRunWarnsForFreeProvider(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module demo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := config.Default()
	cfg.Provider = "free"
	report := Run(context.Background(), cfg, root)
	joined := strings.Join(report.Warnings, " | ")
	if !strings.Contains(joined, "shared free provider") {
		t.Fatalf("expected free provider warning, got: %s", joined)
	}
	if !strings.Contains(joined, "claude, gemini, or openai") {
		t.Fatalf("expected recommendation warning, got: %s", joined)
	}
}
