package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	infoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	okStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	warnStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	errStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	mutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
)

func Banner(version string) string {
	lines := []string{
		titleStyle.Render("tervdocs"),
		mutedStyle.Render("AI-powered README generator for modern codebases"),
	}
	if version != "" {
		lines = append(lines, mutedStyle.Render("version "+version))
	}
	return strings.Join(lines, "\n")
}

func Info(format string, args ...any) string {
	return infoStyle.Render("INFO") + " " + fmt.Sprintf(format, args...)
}

func Success(format string, args ...any) string {
	return okStyle.Render("OK") + " " + fmt.Sprintf(format, args...)
}

func Warn(format string, args ...any) string {
	return warnStyle.Render("WARN") + " " + fmt.Sprintf(format, args...)
}

func Error(format string, args ...any) string {
	return errStyle.Render("ERROR") + " " + fmt.Sprintf(format, args...)
}
