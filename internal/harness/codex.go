package harness

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlexGladkov/harnest/internal/detector"
	"github.com/AlexGladkov/harnest/internal/mapping"
)

type CodexGenerator struct{}

func (g *CodexGenerator) Generate(projectDir string, stacks []detector.Stack, agents mapping.AgentConfig) (string, error) {
	var b strings.Builder

	b.WriteString("# Project Instructions\n\n")

	// Stack context
	b.WriteString("## Tech Stack\n")
	for _, s := range stacks {
		b.WriteString(fmt.Sprintf("- %s (%s) at %s\n", s.Name, s.Lang, s.Path))
	}
	b.WriteString("\n")

	// Expert roles (consilium equivalent)
	b.WriteString("## Expert Perspectives\n")
	b.WriteString("When analyzing code, consider these perspectives:\n\n")
	for _, c := range agents.Consilium {
		if c.Agent == "" {
			continue
		}
		b.WriteString(fmt.Sprintf("- **%s**: %s (%s)\n", c.Role, describeRole(c.Role), c.Agent))
	}
	b.WriteString("\n")

	// File ownership
	b.WriteString("## File Ownership\n")
	b.WriteString("Match code style and patterns for each area:\n\n")
	for _, e := range agents.Exec {
		if e.Agent == "" {
			continue
		}
		b.WriteString(fmt.Sprintf("- `%s` → %s patterns\n", e.Scope, e.Agent))
	}
	b.WriteString("\n")

	outPath := filepath.Join(projectDir, "AGENTS.md")
	if _, err := os.Stat(outPath); err == nil {
		outPath = filepath.Join(projectDir, "AGENTS.generated.md")
	}

	err := os.WriteFile(outPath, []byte(b.String()), 0644)
	if err != nil {
		return "", fmt.Errorf("writing %s: %w", outPath, err)
	}

	return outPath, nil
}
