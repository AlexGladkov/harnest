package converter

import (
	"fmt"

	"github.com/AlexGladkov/harnest/internal/config"
	"github.com/AlexGladkov/harnest/internal/detector"
	"github.com/AlexGladkov/harnest/internal/harness"
	"github.com/AlexGladkov/harnest/internal/mapping"
)

// Convert reads existing config from one harness format and generates another.
func Convert(dir, from, to string) (string, error) {
	// Try to read existing project config
	cfg, err := config.ReadProject(dir)

	var agents mapping.AgentConfig
	if err != nil {
		// No existing config — detect and generate fresh
		fmt.Printf("No existing %s config found, detecting stack...\n", from)
		stacks := detector.Detect(dir)
		agents = mapping.Resolve(stacks)
	} else {
		// Use existing config
		agents = mapping.AgentConfig{
			Consilium: cfg.Consilium,
			Exec:      cfg.Exec,
		}
	}

	stacks := detector.Detect(dir)

	gen, err := harness.Get(to)
	if err != nil {
		return "", fmt.Errorf("target harness: %w", err)
	}

	outPath, err := gen.Generate(dir, stacks, agents)
	if err != nil {
		return "", fmt.Errorf("generating %s config: %w", to, err)
	}

	return outPath, nil
}
