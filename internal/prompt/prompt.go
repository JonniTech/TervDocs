package prompt

import (
	"fmt"
	"strings"

	"tervdocs/internal/summarize"
	"tervdocs/internal/templates"
)

func Build(t templates.Template, ctx summarize.Context) (system string, user string) {
	templateRules := templates.RenderInstruction(t)
	visualRules := []string{
		"Use no emoji anywhere in the README.",
		"Prefer icons, badges, inline SVG, and mermaid diagrams instead of emoji.",
		"Separate each major section with a very thin animated SVG divider using the detected brand color " + ctx.BrandColor + ".",
		"Use the primary language brand color " + ctx.BrandColor + " for visual accents.",
	}
	contentRules := []string{
		"Stay grounded in the provided repository context only.",
		"Do not mention tools, services, flows, or frameworks that are not evidenced in the context.",
		"Write a substantial README with enough detail for onboarding, setup, architecture understanding, and daily development.",
		"Include concrete sections covering aim, problem, solution, features, tech stack, installation, usage, project structure, development workflow, and contribution guidance.",
		"Include at least one mermaid flow diagram and one architecture-oriented visual section when the context allows it.",
		"Prefer precise bullets and short paragraphs over generic marketing filler.",
		"If any detail is uncertain, state it as an assumption instead of inventing facts.",
	}
	if ctx.DeveloperName != "" {
		contentRules = append(contentRules, "End with a footer that says exactly: Programmed by "+ctx.DeveloperName)
	}
	system = fmt.Sprintf(
		"You are tervdocs, a premium documentation assistant.\n%s\nVisual Rules: %s\nContent Rules: %s",
		templateRules,
		strings.Join(visualRules, " "),
		strings.Join(contentRules, " "),
	)
	user = fmt.Sprintf(
		"Generate a high-quality README.md from this repository context JSON.\nReturn markdown only.\nRepository context:\n%s",
		summarize.CompactJSON(ctx),
	)
	return system, user
}
