package harness

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlexGladkov/harnest/internal/detector"
	"github.com/AlexGladkov/harnest/internal/mapping"
)

type CursorGenerator struct{}

func (g *CursorGenerator) Generate(projectDir string, stacks []detector.Stack, agents mapping.AgentConfig) (string, error) {
	var b strings.Builder

	b.WriteString("# Project Rules\n\n")

	// Stack context
	b.WriteString("## Tech Stack\n")
	for _, s := range stacks {
		b.WriteString(fmt.Sprintf("- %s (%s) at %s\n", s.Name, s.Lang, s.Path))
	}
	b.WriteString("\n")

	// Simplified agent guidance (Cursor doesn't have Task tool / consilium)
	b.WriteString("## Expert Roles\n")
	b.WriteString("When analyzing code, consider these perspectives:\n\n")
	for _, c := range agents.Consilium {
		b.WriteString(fmt.Sprintf("- **%s**: %s\n", c.Role, describeRole(c.Role)))
	}
	b.WriteString("\n")

	// File-scope guidance
	b.WriteString("## File Ownership\n")
	b.WriteString("Match code style and patterns for each area:\n\n")
	for _, e := range agents.Exec {
		b.WriteString(fmt.Sprintf("- `%s` → %s patterns\n", e.Scope, e.Agent))
	}
	b.WriteString("\n")

	outPath := filepath.Join(projectDir, ".cursorrules")
	if _, err := os.Stat(outPath); err == nil {
		outPath = filepath.Join(projectDir, ".cursorrules.generated")
	}

	err := os.WriteFile(outPath, []byte(b.String()), 0644)
	if err != nil {
		return "", fmt.Errorf("writing %s: %w", outPath, err)
	}

	return outPath, nil
}

func describeRole(role string) string {
	descriptions := map[string]string{
		"architect":   "Architecture, modules, dependencies, SOLID",
		"frontend":    "UI/UX review, frontend patterns",
		"ui":          "Visual design, UX, components",
		"security":    "OWASP, vulnerabilities, auth",
		"devops":      "Infrastructure, CI/CD, deployment",
		"api":         "API contracts, REST/GraphQL",
		"diagnostics": "Logs, stacktraces, debugging",
		"test":        "Test coverage, quality",
	}
	if d, ok := descriptions[role]; ok {
		return d
	}
	return role
}
