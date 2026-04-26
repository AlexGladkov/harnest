package harness

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlexGladkov/harnest/internal/detector"
	"github.com/AlexGladkov/harnest/internal/mapping"
)

type WindsurfGenerator struct{}

func (g *WindsurfGenerator) Generate(projectDir string, stacks []detector.Stack, agents mapping.AgentConfig) (string, error) {
	var b strings.Builder

	b.WriteString("# Project Configuration\n\n")

	b.WriteString("## Stack\n")
	for _, s := range stacks {
		b.WriteString(fmt.Sprintf("- %s (%s) at %s\n", s.Name, s.Lang, s.Path))
	}
	b.WriteString("\n")

	b.WriteString("## Code Areas\n")
	for _, e := range agents.Exec {
		b.WriteString(fmt.Sprintf("- `%s`: follow %s conventions\n", e.Scope, agentToStyle(e.Agent)))
	}
	b.WriteString("\n")

	outPath := filepath.Join(projectDir, ".windsurfrules")
	if _, err := os.Stat(outPath); err == nil {
		outPath = filepath.Join(projectDir, ".windsurfrules.generated")
	}

	err := os.WriteFile(outPath, []byte(b.String()), 0644)
	if err != nil {
		return "", fmt.Errorf("writing %s: %w", outPath, err)
	}

	return outPath, nil
}

func agentToStyle(agent string) string {
	parts := strings.Split(agent, ":")
	if len(parts) > 1 {
		return parts[1]
	}
	return agent
}
