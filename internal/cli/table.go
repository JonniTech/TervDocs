package cli

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"golang.org/x/term"
)

var (
	tableTitleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color(TervuxBlack)).Background(lipgloss.Color(TervuxPurple)).Bold(true).Padding(0, 1)
	tableHeaderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(TervuxBlack)).Background(lipgloss.Color(TervuxPink)).Bold(true)
	tableCellStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(TervuxSoft))
	tableKeyStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(TervuxPink)).Bold(true)
	borderStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color(TervuxPurple))
)

const (
	minRenderableWidth    = 8
	overflowHideThreshold = 72
	clippedLayoutMinWidth = 76
)

func PrintTable(w io.Writer, title string, headers []string, rows [][]string) {
	fmt.Fprintln(w, RenderTable(title, headers, rows))
}

func RenderTable(title string, headers []string, rows [][]string) string {
	colCount := len(headers)
	if colCount == 0 {
		return ""
	}

	maxTableWidth := terminalWidth() - 2
	if maxTableWidth < minRenderableWidth {
		maxTableWidth = minRenderableWidth
	}
	if shouldHideOverflow() && maxTableWidth < clippedLayoutMinWidth {
		maxTableWidth = clippedLayoutMinWidth
	}
	widths := fitColumnWidths(headers, rows, maxTableWidth)

	var b strings.Builder
	if title != "" {
		b.WriteString(renderTitle(title, maxTableWidth))
		b.WriteString("\n")
	}
	b.WriteString(borderLine("╭", "┬", "╮", widths))
	b.WriteString("\n")
	b.WriteString(renderRow(headers, widths, true))
	b.WriteString("\n")
	b.WriteString(borderLine("├", "┼", "┤", widths))
	for _, row := range rows {
		b.WriteString("\n")
		b.WriteString(renderRow(row, widths, false))
	}
	b.WriteString("\n")
	b.WriteString(borderLine("╰", "┴", "╯", widths))
	return clipLinesToTerminalWidth(b.String())
}

func KVRows(items map[string]string) [][]string {
	keys := make([]string, 0, len(items))
	for k := range items {
		keys = append(keys, k)
	}
	sortStrings(keys)
	rows := make([][]string, 0, len(keys))
	for _, k := range keys {
		rows = append(rows, []string{k, items[k]})
	}
	return rows
}

func StatusTableRows(checks, warnings, errs []string) [][]string {
	rows := [][]string{}
	for _, item := range checks {
		rows = append(rows, []string{"OK", item})
	}
	for _, item := range warnings {
		rows = append(rows, []string{"WARN", item})
	}
	for _, item := range errs {
		rows = append(rows, []string{"ERROR", item})
	}
	return rows
}

func renderRow(cols []string, widths []int, header bool) string {
	wrappedCols := make([][]string, len(widths))
	maxLines := 1
	hideOverflow := shouldHideOverflow()
	for i := range widths {
		cell := ""
		if i < len(cols) {
			cell = cols[i]
		}
		if hideOverflow {
			wrappedCols[i] = clipCellText(cell, widths[i])
		} else {
			wrappedCols[i] = wrapText(cell, widths[i])
		}
		if len(wrappedCols[i]) > maxLines {
			maxLines = len(wrappedCols[i])
		}
	}

	lines := make([]string, 0, maxLines)
	for lineIdx := 0; lineIdx < maxLines; lineIdx++ {
		out := make([]string, len(widths))
		for i := range widths {
			cellLine := ""
			if lineIdx < len(wrappedCols[i]) {
				cellLine = wrappedCols[i][lineIdx]
			}
			padded := pad(cellLine, widths[i])
			switch {
			case header:
				out[i] = tableHeaderStyle.Width(widths[i] + 2).Render(" " + padded + " ")
			case i == 0:
				out[i] = renderFirstCell(cellLine, padded, widths[i]+2)
			default:
				out[i] = tableCellStyle.Width(widths[i] + 2).Render(" " + padded + " ")
			}
		}
		lines = append(lines, borderStyle.Render("│")+strings.Join(out, borderStyle.Render("│"))+borderStyle.Render("│"))
	}
	return strings.Join(lines, "\n")
}

func renderFirstCell(raw, padded string, width int) string {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "OK":
		return lipgloss.NewStyle().Foreground(lipgloss.Color(TervuxBlack)).Background(lipgloss.Color("#7CFFB2")).Bold(true).Width(width).Render(" " + padded + " ")
	case "WARN":
		return lipgloss.NewStyle().Foreground(lipgloss.Color(TervuxBlack)).Background(lipgloss.Color("#FFD166")).Bold(true).Width(width).Render(" " + padded + " ")
	case "ERROR":
		return lipgloss.NewStyle().Foreground(lipgloss.Color(TervuxBlack)).Background(lipgloss.Color("#FF6B6B")).Bold(true).Width(width).Render(" " + padded + " ")
	case "INFO":
		return lipgloss.NewStyle().Foreground(lipgloss.Color(TervuxBlack)).Background(lipgloss.Color(TervuxPink)).Bold(true).Width(width).Render(" " + padded + " ")
	default:
		return tableKeyStyle.Width(width).Render(" " + padded + " ")
	}
}

func borderLine(left, mid, right string, widths []int) string {
	parts := make([]string, len(widths))
	for i, width := range widths {
		parts[i] = strings.Repeat("─", width+2)
	}
	return borderStyle.Render(left + strings.Join(parts, mid) + right)
}

func fitColumnWidths(headers []string, rows [][]string, maxTableWidth int) []int {
	colCount := len(headers)
	natural := make([]int, colCount)
	mins := make([]int, colCount)
	widths := make([]int, colCount)

	for i, h := range headers {
		natural[i] = max(natural[i], lipgloss.Width(h))
		mins[i] = minColumnWidth(i, colCount, lipgloss.Width(h))
	}
	for _, row := range rows {
		for i := 0; i < colCount && i < len(row); i++ {
			for _, part := range strings.Split(row[i], "\n") {
				natural[i] = max(natural[i], lipgloss.Width(strings.TrimSpace(part)))
			}
		}
	}

	preferredCaps := preferredColumnCaps(colCount, maxTableWidth)
	for i := range widths {
		widths[i] = natural[i]
		if preferredCaps[i] > 0 && widths[i] > preferredCaps[i] {
			widths[i] = preferredCaps[i]
		}
		if widths[i] < mins[i] {
			widths[i] = mins[i]
		}
	}

	for totalTableWidth(widths) > maxTableWidth {
		idx := widestShrinkable(widths, mins)
		if idx == -1 {
			break
		}
		widths[idx]--
	}
	return widths
}

func preferredColumnCaps(colCount, maxTableWidth int) []int {
	caps := make([]int, colCount)
	if colCount == 1 {
		caps[0] = maxTableWidth - 4
		return caps
	}
	if colCount == 2 {
		caps[0] = max(10, min(20, maxTableWidth/3))
		caps[1] = maxTableWidth
		return caps
	}
	if colCount == 3 {
		caps[0] = max(8, maxTableWidth/5)
		caps[1] = max(8, maxTableWidth/6)
		caps[2] = maxTableWidth
		return caps
	}
	for i := range caps {
		if i == colCount-1 {
			caps[i] = maxTableWidth
		} else {
			caps[i] = max(7, maxTableWidth/(colCount+1))
		}
	}
	return caps
}

func minColumnWidth(index, colCount, headerWidth int) int {
	base := 3
	if colCount == 2 && index == 0 {
		base = 4
	}
	if headerWidth > 0 && headerWidth < 4 {
		base = max(base, headerWidth)
	}
	return base
}

func widestShrinkable(widths, mins []int) int {
	idx := -1
	best := -1
	for i := range widths {
		if widths[i] <= mins[i] {
			continue
		}
		if widths[i] > best {
			best = widths[i]
			idx = i
		}
	}
	return idx
}

func totalTableWidth(widths []int) int {
	total := 1
	for _, width := range widths {
		total += width + 3
	}
	return total
}

func wrapText(text string, width int) []string {
	if width <= 1 {
		return []string{ansi.Truncate(text, max(width, 0), "")}
	}
	paragraphs := strings.Split(strings.ReplaceAll(text, "\t", " "), "\n")
	out := make([]string, 0, len(paragraphs))
	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			out = append(out, "")
			continue
		}
		words := strings.Fields(paragraph)
		if len(words) == 0 {
			out = append(out, "")
			continue
		}
		line := ""
		for _, word := range words {
			wordParts := splitLongWord(word, width)
			for _, part := range wordParts {
				if line == "" {
					line = part
					continue
				}
				if lipgloss.Width(line)+1+lipgloss.Width(part) <= width {
					line += " " + part
					continue
				}
				out = append(out, line)
				line = part
			}
		}
		if line != "" {
			out = append(out, line)
		}
	}
	if len(out) == 0 {
		return []string{""}
	}
	return out
}

func clipCellText(text string, width int) []string {
	if width <= 0 {
		return []string{""}
	}
	clean := strings.Join(strings.Fields(strings.ReplaceAll(strings.ReplaceAll(text, "\t", " "), "\n", " ")), " ")
	return []string{ansi.Truncate(clean, width, "")}
}

func splitLongWord(word string, width int) []string {
	if lipgloss.Width(word) <= width {
		return []string{word}
	}
	runes := []rune(word)
	out := []string{}
	chunk := make([]rune, 0, width)
	for _, r := range runes {
		chunk = append(chunk, r)
		if lipgloss.Width(string(chunk)) >= width {
			out = append(out, string(chunk))
			chunk = chunk[:0]
		}
	}
	if len(chunk) > 0 {
		out = append(out, string(chunk))
	}
	return out
}

func renderTitle(title string, maxWidth int) string {
	rendered := tableTitleStyle.Render(" " + title + " ")
	if lipgloss.Width(rendered) <= maxWidth {
		return rendered
	}
	return tableTitleStyle.Width(maxWidth).Render(" " + title + " ")
}

func terminalWidth() int {
	if raw := os.Getenv("COLUMNS"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			return n
		}
	}
	if width, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && width > 0 {
		return width
	}
	return 100
}

func shouldHideOverflow() bool {
	return terminalWidth() < overflowHideThreshold
}

func clipLinesToTerminalWidth(s string) string {
	width := terminalWidth()
	if width <= 0 {
		return s
	}

	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = ansi.Truncate(line, width, "")
	}
	return strings.Join(lines, "\n")
}

func pad(s string, width int) string {
	if lipgloss.Width(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-lipgloss.Width(s))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func sortStrings(items []string) {
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j] < items[i] {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}
