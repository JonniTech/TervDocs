package render

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

const (
	dividerAssetDir  = ".tervdocs"
	dividerAssetFile = "readme-divider.svg"
)

var hexColorPattern = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

func DividerAssetRelativePath() string {
	return "./" + dividerAssetDir + "/" + dividerAssetFile
}

func EnsureDividerAsset(outputPath, color string) error {
	assetDir := filepath.Join(filepath.Dir(outputPath), dividerAssetDir)
	if err := os.MkdirAll(assetDir, 0o755); err != nil {
		return err
	}
	assetPath := filepath.Join(assetDir, dividerAssetFile)
	return os.WriteFile(assetPath, []byte(dividerAssetSVG(normalizeDividerColor(color))), 0o644)
}

func normalizeDividerColor(color string) string {
	if hexColorPattern.MatchString(color) {
		return color
	}
	return "#4F46E5"
}

func dividerAssetSVG(color string) string {
	return fmt.Sprintf(`<svg width="1200" height="18" viewBox="0 0 1200 18" fill="none" xmlns="http://www.w3.org/2000/svg" role="img" aria-label="section divider">
  <defs>
    <linearGradient id="tervdocs-divider-gradient" x1="0" y1="9" x2="1200" y2="9" gradientUnits="userSpaceOnUse">
      <stop offset="0" stop-color="%[1]s" stop-opacity="0"/>
      <stop offset="0.22" stop-color="%[1]s" stop-opacity="0.32"/>
      <stop offset="0.5" stop-color="%[1]s" stop-opacity="1"/>
      <stop offset="0.78" stop-color="%[1]s" stop-opacity="0.32"/>
      <stop offset="1" stop-color="%[1]s" stop-opacity="0"/>
    </linearGradient>
  </defs>
  <rect x="0" y="8" width="1200" height="2" rx="1" fill="url(#tervdocs-divider-gradient)"/>
  <rect x="528" y="6" width="144" height="6" rx="3" fill="%[1]s" opacity="0.14"/>
  <circle cx="566" cy="9" r="3" fill="%[1]s" opacity="0.38"/>
  <circle cx="600" cy="9" r="4" fill="%[1]s"/>
  <circle cx="634" cy="9" r="3" fill="%[1]s" opacity="0.38"/>
  <path d="M600 2L607 9L600 16L593 9L600 2Z" fill="%[1]s" opacity="0.18"/>
</svg>
`, color)
}
