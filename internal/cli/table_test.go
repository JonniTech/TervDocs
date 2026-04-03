package cli

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestRenderTableRespectsTerminalWidth(t *testing.T) {
	t.Setenv("COLUMNS", "60")
	out := RenderTable("Providers", []string{"Provider", "Default Model", "Quality", "Auth"}, [][]string{
		{"free", "glm-4.7-flash", "shared / variable", "built-in shared key"},
		{"openai", "gpt-4o-mini", "high", "OPENAI_API_KEY"},
	})
	for _, line := range strings.Split(out, "\n") {
		if lipgloss.Width(line) > 60 {
			t.Fatalf("line exceeds terminal width: %d > 60\n%s", lipgloss.Width(line), line)
		}
	}
}

func TestBannerCompactsOnSmallTerminal(t *testing.T) {
	t.Setenv("COLUMNS", "50")
	out := Banner("0.1.0")
	for _, line := range strings.Split(out, "\n") {
		if lipgloss.Width(line) > 50 {
			t.Fatalf("banner line exceeds terminal width: %d > 50\n%s", lipgloss.Width(line), line)
		}
	}
}

func TestBannerWideLogoStaysReadable(t *testing.T) {
	t.Setenv("COLUMNS", "120")
	out := Banner("0.1.0")
	if !strings.Contains(out, "\n\n    T E R V U X") {
		t.Fatalf("expected readable wordmark block in wide banner:\n%s", out)
	}
	if !strings.Contains(out, "• • • • •") {
		t.Fatalf("expected dot logo in wide banner:\n%s", out)
	}
	for _, line := range strings.Split(out, "\n") {
		if lipgloss.Width(line) > 120 {
			t.Fatalf("wide banner line exceeds terminal width: %d > 120\n%s", lipgloss.Width(line), line)
		}
	}
}

func TestBannerUsesLargeLogoOnExpandedTerminal(t *testing.T) {
	t.Setenv("COLUMNS", "80")
	out := Banner("0.1.0")
	if !strings.Contains(out, "\n\n    T E R V U X") {
		t.Fatalf("expected large wordmark block on expanded terminal:\n%s", out)
	}
}

func TestRenderTableHidesOverflowOnVeryNarrowTerminal(t *testing.T) {
	t.Setenv("COLUMNS", "20")
	out := RenderTable("Providers", []string{"Provider", "Default Model", "Quality", "Auth"}, [][]string{
		{"free", "glm-4.7-flash", "shared / variable", "built-in shared key"},
		{"openai", "gpt-4o-mini", "high", "OPENAI_API_KEY"},
	})
	lines := strings.Split(out, "\n")
	if len(lines) != 7 {
		t.Fatalf("expected clipped table to stay single-line per row, got %d lines:\n%s", len(lines), out)
	}
	for _, line := range lines {
		if lipgloss.Width(line) > 20 {
			t.Fatalf("line exceeds terminal width: %d > 20\n%s", lipgloss.Width(line), line)
		}
	}
}

func TestRenderTableShowsExpandedContentOnMediumTerminal(t *testing.T) {
	t.Setenv("COLUMNS", "80")
	out := RenderTable("Providers", []string{"Provider", "Default Model", "Quality", "Auth"}, [][]string{
		{"free", "glm-4.7-flash", "shared / variable", "built-in shared key"},
		{"openai", "gpt-4o-mini", "high", "OPENAI_API_KEY"},
	})
	if !strings.Contains(out, "built-in shared key") {
		t.Fatalf("expected medium terminal to reveal more content after expand:\n%s", out)
	}
	for _, line := range strings.Split(out, "\n") {
		if lipgloss.Width(line) > 80 {
			t.Fatalf("line exceeds terminal width: %d > 80\n%s", lipgloss.Width(line), line)
		}
	}
}

func TestBannerHidesOverflowOnVeryNarrowTerminal(t *testing.T) {
	t.Setenv("COLUMNS", "18")
	out := Banner("0.1.0")
	for _, line := range strings.Split(out, "\n") {
		if lipgloss.Width(line) > 18 {
			t.Fatalf("banner line exceeds terminal width: %d > 18\n%s", lipgloss.Width(line), line)
		}
	}
}
