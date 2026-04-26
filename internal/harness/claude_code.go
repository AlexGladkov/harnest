package harness

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlexGladkov/harnest/internal/detector"
	"github.com/AlexGladkov/harnest/internal/mapping"
)

type ClaudeCodeGenerator struct{}

func (g *ClaudeCodeGenerator) Generate(projectDir string, stacks []detector.Stack, agents mapping.AgentConfig) (string, error) {
	var b strings.Builder

	projectName := filepath.Base(projectDir)

	b.WriteString(fmt.Sprintf("# %s\n\n", projectName))

	// Stack section
	b.WriteString("## Stack\n")
	for _, s := range stacks {
		b.WriteString(fmt.Sprintf("- %s (%s)\n", s.Name, s.Path))
	}
	b.WriteString("\n")

	// Agents section
	b.WriteString("## Agents\n\n")

	// Consilium
	b.WriteString("### Consilium\n")
	b.WriteString("| Role | Agent |\n")
	b.WriteString("|------|-------|\n")
	for _, c := range agents.Consilium {
		if c.Agent == "" {
			continue
		}
		b.WriteString(fmt.Sprintf("| %s | %s |\n", c.Role, c.Agent))
	}
	b.WriteString("\n")

	// Executing
	b.WriteString("### Executing\n")
	b.WriteString("| Agent | Scope |\n")
	b.WriteString("|-------|-------|\n")
	for _, e := range agents.Exec {
		if e.Agent == "" {
			continue
		}
		b.WriteString(fmt.Sprintf("| %s | %s |\n", e.Agent, e.Scope))
	}
	b.WriteString("\n")

	outPath := filepath.Join(projectDir, "CLAUDE.md")

	// Don't overwrite existing
	if _, err := os.Stat(outPath); err == nil {
		outPath = filepath.Join(projectDir, "CLAUDE.generated.md")
	}

	err := os.WriteFile(outPath, []byte(b.String()), 0644)
	if err != nil {
		return "", fmt.Errorf("writing %s: %w", outPath, err)
	}

	return outPath, nil
}
