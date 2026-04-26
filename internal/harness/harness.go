package harness

import (
	"fmt"
	"sort"
	"strings"

	"github.com/AlexGladkov/harnest/internal/detector"
	"github.com/AlexGladkov/harnest/internal/mapping"
)

type Generator interface {
	Generate(projectDir string, stacks []detector.Stack, agents mapping.AgentConfig) (string, error)
}

// HarnessInfo holds metadata about a harness for agent discovery.
type HarnessInfo struct {
	Generator Generator
	// AgentDir is the relative path under $HOME where this harness stores custom agents.
	// Empty means no custom agent dir.
	AgentDir string
}

var registry = map[string]HarnessInfo{
	"claude-code": {Generator: &ClaudeCodeGenerator{}, AgentDir: ".claude/agents"},
	"cursor":      {Generator: &CursorGenerator{}, AgentDir: ".cursor/agents"},
	"windsurf":    {Generator: &WindsurfGenerator{}, AgentDir: ".windsurf/agents"},
	"codex":       {Generator: &CodexGenerator{}, AgentDir: ".codex/agents"},
	"opencode":    {Generator: &OpenCodeGenerator{}, AgentDir: ".config/opencode/agents"},
}

func Get(name string) (Generator, error) {
	h, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown harness: %s (available: %s)", name, strings.Join(Names(), ", "))
	}
	return h.Generator, nil
}

// Names returns sorted list of all registered harness names.
func Names() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// AgentDirs returns all agent directory paths (relative to $HOME) from registered harnesses.
func AgentDirs() []string {
	var dirs []string
	for _, h := range registry {
		if h.AgentDir != "" {
			dirs = append(dirs, h.AgentDir)
		}
	}
	sort.Strings(dirs)
	return dirs
}
