package templates

import (
	"fmt"
	"slices"
	"strings"
)

type Template struct {
	Name      string
	Sections  []string
	Tone      string
	Format    []string
	SystemAdd string
}

var builtIn = map[string]Template{
	"default": {
		Name: "default",
		Sections: []string{
			"Title", "Overview", "Features", "Tech Stack", "Installation", "Usage", "Scripts/Commands",
			"Environment Variables", "Project Structure", "API or Architecture Notes",
			"Development Workflow", "Contributing", "License",
		},
		Tone:   "professional and practical",
		Format: []string{"Use clean markdown headings", "Prefer concise examples and commands"},
	},
	"minimal": {
		Name:     "minimal",
		Sections: []string{"Title", "Overview", "Tech Stack", "Quick Start", "Usage", "License"},
		Tone:     "concise and direct",
		Format:   []string{"Keep sections short", "Avoid long prose"},
	},
	"detailed": {
		Name: "detailed",
		Sections: []string{
			"Title", "Overview", "Problem Statement", "Features", "Tech Stack", "Architecture",
			"Installation", "Configuration", "Environment Variables", "Usage", "Scripts/Commands",
			"Project Structure", "Testing", "Deployment", "Troubleshooting", "Contributing", "License",
		},
		Tone:      "deeply explanatory but developer-friendly",
		Format:    []string{"Include command snippets", "Call out assumptions"},
		SystemAdd: "When context is uncertain, mark assumptions clearly.",
	},
	"tervux": {
		Name: "tervux",
		Sections: []string{
			"Title", "Overview", "Why This Project", "Feature Highlights", "Tech Stack",
			"Getting Started", "Usage", "Developer Commands", "Environment Variables",
			"Architecture Notes", "Project Structure", "Contributing", "License",
		},
		Tone:   "confident, modern, and practical under the Tervux brand",
		Format: []string{"Use polished markdown and strong section flow", "Keep DX-first language"},
	},
}

func List() []string {
	out := make([]string, 0, len(builtIn))
	for k := range builtIn {
		out = append(out, k)
	}
	slices.Sort(out)
	return out
}

func Get(name string) (Template, error) {
	t, ok := builtIn[name]
	if !ok {
		return Template{}, fmt.Errorf("unknown template: %s", name)
	}
	return t, nil
}

func MustGet(name string) Template {
	t, err := Get(name)
	if err != nil {
		return builtIn["default"]
	}
	return t
}

func RenderInstruction(t Template) string {
	return fmt.Sprintf(
		"Template: %s\nTone: %s\nSection Order: %s\nFormatting Rules: %s\n%s",
		t.Name,
		t.Tone,
		strings.Join(t.Sections, ", "),
		strings.Join(t.Format, "; "),
		t.SystemAdd,
	)
}
