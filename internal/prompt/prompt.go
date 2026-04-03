package prompt

import (
	"fmt"

	"tervdocs/internal/summarize"
	"tervdocs/internal/templates"
)

func Build(t templates.Template, ctx summarize.Context) (system string, user string) {
	templateRules := templates.RenderInstruction(t)
	system = fmt.Sprintf(
		"You are tervdocs, a premium documentation assistant. Create a polished README.md.\n%s\nDo not invent unknown technical details.",
		templateRules,
	)
	user = fmt.Sprintf(
		"Generate README markdown for this repository context JSON:\n%s\nReturn markdown only.",
		summarize.CompactJSON(ctx),
	)
	return system, user
}
