package cli

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	TervuxPink   = "#d946ef"
	TervuxPurple = "#5E429C"
	TervuxWhite  = "#FFFFFF"
	TervuxBlack  = "#000000"
	TervuxInk    = "#b58ad4"
	TervuxSoft   = "#e9d5ff"
)

var (
	bannerTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(TervuxPink))
	bannerBodyStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color(TervuxSoft))
	mutedStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color(TervuxInk))
	titleBadgeStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color(TervuxBlack)).Background(lipgloss.Color(TervuxPink)).Bold(true).Padding(0, 1)
	wordmarkStyle    = lipgloss.NewStyle().Bold(true)
)

const (
	largeBannerThreshold = 80
	largeLogoInset       = "    "
	largeLogoGap         = "    "
)

func Banner(version string) string {
	if terminalWidth() < largeBannerThreshold {
		lines := []string{
			renderCompactLogo(),
			renderWordmark("TERVUX") + "  " + titleBadgeStyle.Render("DOCS"),
			bannerBodyStyle.Render("Premium README generation"),
		}
		if version != "" {
			lines = append(lines, mutedStyle.Render("version "+version))
		}
		return clipLinesToTerminalWidth(strings.Join(lines, "\n"))
	}

	lines := []string{
		renderDotLogo(),
		"",
		largeLogoInset + renderWordmark("TERVUX") + "  " + titleBadgeStyle.Render("DOCS CLI"),
		largeLogoInset + bannerBodyStyle.Render("High-context README generation with premium Tervux DX"),
	}
	if version != "" {
		lines = append(lines, largeLogoInset+mutedStyle.Render("version "+version))
	}
	return clipLinesToTerminalWidth(strings.Join(lines, "\n"))
}

func renderCompactLogo() string {
	return strings.Join([]string{
		renderDotLine(" 101 ", 0),
		renderDotLine("01010", 1),
		renderDotLine(" 101 ", 2),
	}, "\n")
}

func renderDotLogo() string {
	patterns := map[rune][]string{
		'T': {"11111", "00100", "00100", "00100", "00100", "00100", "00100"},
		'E': {"11111", "10000", "10000", "11110", "10000", "10000", "11111"},
		'R': {"11110", "10001", "10001", "11110", "10100", "10010", "10001"},
		'V': {"10001", "10001", "10001", "10001", "01010", "01010", "00100"},
		'U': {"10001", "10001", "10001", "10001", "10001", "10001", "11111"},
		'X': {"10001", "10001", "01010", "00100", "01010", "10001", "10001"},
	}
	word := "TERVUX"
	lines := make([]string, len(patterns['T']))
	for row := range lines {
		parts := make([]string, 0, len(word))
		for idx, ch := range word {
			parts = append(parts, renderDotGlyphLine(patterns[ch][row], idx))
		}
		lines[row] = largeLogoInset + strings.Join(parts, largeLogoGap)
	}
	return strings.Join(lines, "\n")
}

func renderDotLine(pattern string, seed int) string {
	var b strings.Builder
	for i, ch := range pattern {
		if ch == '1' {
			color := TervuxPink
			if (seed+i)%2 == 1 {
				color = TervuxPurple
			}
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render("•"))
		} else {
			b.WriteString(" ")
		}
	}
	return b.String()
}

func renderDotGlyphLine(pattern string, seed int) string {
	var b strings.Builder
	color := TervuxPink
	if seed%2 == 1 {
		color = TervuxPurple
	}
	for i, ch := range pattern {
		if ch == '1' {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render("•"))
		} else {
			b.WriteString(" ")
		}
		if i < len(pattern)-1 {
			b.WriteString(" ")
		}
	}
	return b.String()
}

func renderWordmark(text string) string {
	runes := []rune(text)
	var b strings.Builder
	for i, r := range runes {
		style := wordmarkStyle.Foreground(lipgloss.Color(TervuxPink))
		if i%2 == 1 {
			style = style.Foreground(lipgloss.Color(TervuxSoft))
		}
		b.WriteString(style.Render(string(r)))
		if i < len(runes)-1 {
			b.WriteString(" ")
		}
	}
	return b.String()
}
