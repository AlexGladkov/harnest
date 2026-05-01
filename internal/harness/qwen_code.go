package harness

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlexGladkov/harnest/internal/detector"
	"github.com/AlexGladkov/harnest/internal/mapping"
)

type QwenCodeGenerator struct{}

func (g *QwenCodeGenerator) Generate(projectDir string, stacks []detector.Stack, agents mapping.AgentConfig) (string, error) {
	var b strings.Builder

	projectName := filepath.Base(projectDir)

	b.WriteString(fmt.Sprintf("# %s\n\n", projectName))

	// Stack section
	b.WriteString("## Stack\n")
	for _, s := range stacks {
		b.WriteString(fmt.Sprintf("- %s (%s) at %s\n", s.Name, s.Lang, s.Path))
	}
	b.WriteString("\n")

	// Expert roles
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
		b.WriteString(fmt.Sprintf("- `%s` → %s patterns\n", e.Scope, agentToStyle(e.Agent)))
	}
	b.WriteString("\n")

	// Model recommendations (Qwen-specific models)
	if len(agents.Models) > 0 {
		b.WriteString("## Model Recommendations\n")
		b.WriteString("For best results with Qwen models:\n\n")
		var high, standard, low []string
		for _, c := range agents.Consilium {
			if c.Agent == "" {
				continue
			}
			tier := agents.Models[c.Role]
			switch tier {
			case "high":
				high = append(high, c.Role)
			case "low":
				low = append(low, c.Role)
			default:
				standard = append(standard, c.Role)
			}
		}
		if len(high) > 0 {
			b.WriteString(fmt.Sprintf("- %s — use qwen-max\n", strings.Join(high, ", ")))
		}
		if len(standard) > 0 {
			b.WriteString(fmt.Sprintf("- %s — use qwen-plus\n", strings.Join(standard, ", ")))
		}
		if len(low) > 0 {
			b.WriteString(fmt.Sprintf("- %s — use qwen-turbo\n", strings.Join(low, ", ")))
		}
		b.WriteString("\n")
	}

	outPath := filepath.Join(projectDir, "QWEN.md")
	if _, err := os.Stat(outPath); err == nil {
		outPath = filepath.Join(projectDir, "QWEN.generated.md")
	}

	err := os.WriteFile(outPath, []byte(b.String()), 0644)
	if err != nil {
		return "", fmt.Errorf("writing %s: %w", outPath, err)
	}

	return outPath, nil
}
