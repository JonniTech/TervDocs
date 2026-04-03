package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".tervdocs.toml")
	cfg := Default()
	cfg.Provider = "free"
	cfg.Template = "tervux"
	cfg.Scan.MaxFiles = 123

	if err := Save(path, cfg); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if loaded.Template != "tervux" {
		t.Fatalf("expected template tervux, got %s", loaded.Template)
	}
	if loaded.Scan.MaxFiles != 123 {
		t.Fatalf("expected max files 123, got %d", loaded.Scan.MaxFiles)
	}
}

func TestEnvKeyFallback(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "abc")
	dir := t.TempDir()
	path := filepath.Join(dir, ".tervdocs.toml")
	cfg := Default()
	cfg.Provider = "openai"
	cfg.Providers.OpenAI.APIKey = ""
	if err := Save(path, cfg); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if loaded.Providers.OpenAI.APIKey != "abc" {
		t.Fatalf("expected OPENAI_API_KEY fallback")
	}
	_ = os.Unsetenv("OPENAI_API_KEY")
}
