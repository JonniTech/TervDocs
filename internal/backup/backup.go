package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func Create(path string) (string, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	ts := time.Now().Format("20060102-150405")
	dst := fmt.Sprintf("%s.bak.%s", path, ts)
	if err := os.WriteFile(dst, src, 0o644); err != nil {
		return "", err
	}
	return filepath.Clean(dst), nil
}
