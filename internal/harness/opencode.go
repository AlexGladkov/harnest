package harness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlexGladkov/harnest/internal/detector"
	"github.com/AlexGladkov/harnest/internal/mapping"
)

type OpenCodeGenerator struct{}

type openCodeConfig struct {
	Agent map[string]openCodeAgent `json:"agent,omitempty"`
}

type openCodeAgent struct {
	Mode        string `json:"mode"`
	Description string `json:"description"`
	Model       string `json:"model,omitempty"`
}

func (g *OpenCodeGenerator) Generate(projectDir string, stacks []detector.Stack, agents mapping.AgentConfig) (string, error) {
	// 1. Generate opencode.json with agent declarations
	cfg := openCodeConfig{
		Agent: make(map[string]openCodeAgent),
	}

	for _, c := range agents.Consilium {
		if c.Agent == "" {
			continue
		}
		agent := openCodeAgent{
			Mode:        "subagent",
			Description: fmt.Sprintf("%s — %s", describeRole(c.Role), c.Agent),
		}
		if tier, ok := agents.Models[c.Role]; ok {
			agent.Model = ResolveTier("opencode", tier)
		}
		cfg.Agent[c.Role] = agent
	}

	for _, e := range agents.Exec {
		if e.Agent == "" {
			continue
		}
		name := agentToStyle(e.Agent)
		cfg.Agent[name] = openCodeAgent{
			Mode:        "subagent",
			Description: fmt.Sprintf("Handles %s files", e.Scope),
		}
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling opencode config: %w", err)
	}

	outPath := filepath.Join(projectDir, "opencode.json")
	if _, err := os.Stat(outPath); err == nil {
		outPath = filepath.Join(projectDir, "opencode.generated.json")
	}

	if err := os.WriteFile(outPath, data, 0644); err != nil {
		return "", fmt.Errorf("writing %s: %w", outPath, err)
	}

	// 2. Generate .opencode/agents/ markdown files with instructions
	agentsDir := filepath.Join(projectDir, ".opencode", "agents")
	os.MkdirAll(agentsDir, 0755)

	// Stack context file
	var stackInfo strings.Builder
	for _, s := range stacks {
		stackInfo.WriteString(fmt.Sprintf("- %s (%s) at %s\n", s.Name, s.Lang, s.Path))
	}

	for _, c := range agents.Consilium {
		if c.Agent == "" {
			continue
		}
		content := fmt.Sprintf("---\nmode: subagent\ndescription: \"%s\"\n---\n\n# %s\n\nRole: %s\nAgent: %s\n\n## Stack\n%s",
			describeRole(c.Role), c.Role, describeRole(c.Role), c.Agent, stackInfo.String())
		agentFile := filepath.Join(agentsDir, c.Role+".md")
		os.WriteFile(agentFile, []byte(content), 0644)
	}

	return outPath, nil
}
