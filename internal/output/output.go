package output

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"tervdocs/internal/backup"
)

type Result struct {
	Wrote      bool
	Path       string
	BackupPath string
}

func ValidateMarkdown(content string) error {
	if strings.TrimSpace(content) == "" {
		return errors.New("generated output is empty")
	}
	if !strings.Contains(content, "#") && !strings.Contains(content, "\n") {
		return errors.New("generated output does not look like markdown")
	}
	return nil
}

func Write(path, content string, doBackup, dryRun bool) (Result, error) {
	if err := ValidateMarkdown(content); err != nil {
		return Result{}, err
	}
	res := Result{Path: path}
	if dryRun {
		return res, nil
	}
	if _, err := os.Stat(path); err == nil && doBackup {
		bak, err := backup.Create(path)
		if err != nil {
			return res, fmt.Errorf("backup failed: %w", err)
		}
		res.BackupPath = bak
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return res, err
	}
	res.Wrote = true
	return res, nil
}
