package converter

import (
	"fmt"

	"github.com/AlexGladkov/harnest/internal/agents"
	"github.com/AlexGladkov/harnest/internal/config"
	"github.com/AlexGladkov/harnest/internal/detector"
	"github.com/AlexGladkov/harnest/internal/harness"
	"github.com/AlexGladkov/harnest/internal/mapping"
)

// Convert reads existing config from one harness format and generates another.
func Convert(dir, from, to string) (string, error) {
	// Try to read existing project config
	cfg, err := config.ReadProject(dir)

	var agentsCfg mapping.AgentConfig
	if err != nil {
		// No existing config — detect and generate fresh
		fmt.Printf("No existing %s config found, detecting stack...\n", from)
		stacks := detector.Detect(dir)
		discovered := agents.Discover()
		agentsCfg = mapping.Resolve(stacks, discovered, to)
	} else {
		// Use existing config
		agentsCfg = mapping.AgentConfig{
			Consilium: cfg.Consilium,
			Exec:      cfg.Exec,
			Models:    cfg.Models,
		}
		// Fill default tiers for roles without explicit model
		if agentsCfg.Models == nil {
			agentsCfg.Models = mapping.DefaultModelTiers()
		} else {
			defaults := mapping.DefaultModelTiers()
			for role, tier := range defaults {
				if _, ok := agentsCfg.Models[role]; !ok {
					agentsCfg.Models[role] = tier
				}
			}
		}
	}

	stacks := detector.Detect(dir)

	gen, err := harness.Get(to)
	if err != nil {
		return "", fmt.Errorf("target harness: %w", err)
	}

	outPath, err := gen.Generate(dir, stacks, agentsCfg)
	if err != nil {
		return "", fmt.Errorf("generating %s config: %w", to, err)
	}

	return outPath, nil
}
