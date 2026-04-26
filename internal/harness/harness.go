package harness

import (
	"fmt"

	"github.com/AlexGladkov/harnest/internal/detector"
	"github.com/AlexGladkov/harnest/internal/mapping"
)

type Generator interface {
	Generate(projectDir string, stacks []detector.Stack, agents mapping.AgentConfig) (string, error)
}

var registry = map[string]Generator{
	"claude-code": &ClaudeCodeGenerator{},
	"cursor":      &CursorGenerator{},
	"windsurf":    &WindsurfGenerator{},
}

func Get(name string) (Generator, error) {
	g, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown harness: %s (available: claude-code, cursor, windsurf)", name)
	}
	return g, nil
}
