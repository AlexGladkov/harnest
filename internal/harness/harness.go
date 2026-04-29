package harness

import (
	"fmt"
	"os"
	"path/filepath"
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
	// GlobalConfigFile is the filename for this harness's global config (e.g. "CLAUDE.md", ".cursorrules").
	GlobalConfigFile string
}

var registry = map[string]HarnessInfo{
	"claude-code": {Generator: &ClaudeCodeGenerator{}, AgentDir: ".claude/agents", GlobalConfigFile: "CLAUDE.md"},
	"cursor":      {Generator: &CursorGenerator{}, AgentDir: ".cursor/agents", GlobalConfigFile: ".cursorrules"},
	"windsurf":    {Generator: &WindsurfGenerator{}, AgentDir: ".windsurf/agents", GlobalConfigFile: ".windsurfrules"},
	"codex":       {Generator: &CodexGenerator{}, AgentDir: ".codex/agents", GlobalConfigFile: "AGENTS.md"},
	"opencode":    {Generator: &OpenCodeGenerator{}, AgentDir: ".config/opencode/agents", GlobalConfigFile: "AGENTS.md"},
	"qwen-code":   {Generator: &QwenCodeGenerator{}, AgentDir: ".qwen/agents", GlobalConfigFile: "QWEN.md"},
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

// GlobalDir returns the absolute path to a harness's home directory.
// Derived from AgentDir's parent joined with $HOME.
func GlobalDir(name string) (string, error) {
	h, ok := registry[name]
	if !ok {
		return "", fmt.Errorf("unknown harness: %s (available: %s)", name, strings.Join(Names(), ", "))
	}
	if h.AgentDir == "" {
		return "", fmt.Errorf("harness %s has no agent dir configured", name)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	// AgentDir is like ".cursor/agents" — parent is ".cursor"
	parent := filepath.Dir(h.AgentDir)
	return filepath.Join(home, parent), nil
}

// GlobalConfigPath returns the absolute path to a harness's global config file.
func GlobalConfigPath(name string) (string, error) {
	h, ok := registry[name]
	if !ok {
		return "", fmt.Errorf("unknown harness: %s (available: %s)", name, strings.Join(Names(), ", "))
	}
	dir, err := GlobalDir(name)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, h.GlobalConfigFile), nil
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
