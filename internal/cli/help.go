package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func Help(cmd *cobra.Command, version string) string {
	sections := []string{Banner(version)}

	overview := [][]string{
		{"Command", cmd.CommandPath()},
		{"Summary", cmd.Short},
		{"Usage", strings.TrimSpace(cmd.UseLine())},
	}
	if cmd.Long != "" && cmd.Long != cmd.Short && !looksLikeBanner(cmd.Long) {
		overview = append(overview, []string{"Details", singleLine(cmd.Long)})
	}
	if cmd.Example != "" {
		overview = append(overview, []string{"Examples", singleLine(cmd.Example)})
	}
	sections = append(sections, RenderTable("Overview", []string{"Field", "Value"}, overview))

	if rows := commandRows(cmd); len(rows) > 0 {
		sections = append(sections, RenderTable("Commands", []string{"Command", "Description"}, rows))
	}
	if rows := flagRows(cmd.InheritedFlags()); len(rows) > 0 {
		sections = append(sections, RenderTable("Global Flags", []string{"Flag", "Default", "Description"}, rows))
	}
	if rows := flagRows(cmd.NonInheritedFlags()); len(rows) > 0 {
		sections = append(sections, RenderTable("Flags", []string{"Flag", "Default", "Description"}, rows))
	}
	return clipLinesToTerminalWidth(strings.Join(sections, "\n\n"))
}

func commandRows(cmd *cobra.Command) [][]string {
	rows := [][]string{}
	for _, child := range cmd.Commands() {
		if child.Hidden || !child.IsAvailableCommand() {
			continue
		}
		rows = append(rows, []string{child.Name(), child.Short})
	}
	return rows
}

func flagRows(flags *pflag.FlagSet) [][]string {
	rows := [][]string{}
	flags.VisitAll(func(flag *pflag.Flag) {
		name := "--" + flag.Name
		if flag.Shorthand != "" {
			name = fmt.Sprintf("-%s, --%s", flag.Shorthand, flag.Name)
		}
		rows = append(rows, []string{name, flag.DefValue, flag.Usage})
	})
	return rows
}

func singleLine(s string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
}

func looksLikeBanner(s string) bool {
	return strings.Contains(s, "TERVUX DOCS") || strings.Contains(s, "•")
}
