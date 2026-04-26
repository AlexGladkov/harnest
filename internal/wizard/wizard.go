package wizard

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/AlexGladkov/harnest/internal/agents"
	"github.com/AlexGladkov/harnest/internal/mapping"
)

func Run(r io.Reader, structure mapping.AgentStructure, suggestions mapping.Suggestions) mapping.AgentConfig {
	scanner := bufio.NewScanner(r)
	config := mapping.AgentConfig{}

	available := agents.Discover()

	fmt.Println("\n── Agent Wizard ──")
	fmt.Printf("Found %d agents on this machine\n", len(available))
	fmt.Println("Enter = accept suggestion, s = skip, ? = search\n")

	// Consilium roles
	for _, role := range structure.Roles {
		suggestion := suggestions.Consilium[role]
		agent := pickAgent(scanner, fmt.Sprintf("Consilium: %s", role), suggestion, available)
		if agent != "" {
			config.Consilium = append(config.Consilium, mapping.ConsiliumRole{
				Role:  role,
				Agent: agent,
			})
		}
	}

	// Exec scopes
	if len(structure.ExecScopes) > 0 {
		fmt.Println()
	}
	for _, es := range structure.ExecScopes {
		suggestion := suggestions.Exec[es.StackName]
		agent := pickAgent(scanner, fmt.Sprintf("Exec: %s → %s", es.StackName, es.Scope), suggestion, available)
		if agent != "" {
			config.Exec = append(config.Exec, mapping.ExecAgent{
				Agent: agent,
				Scope: es.Scope,
			})
		}
	}

	fmt.Println()
	return config
}

func pickAgent(scanner *bufio.Scanner, label, suggestion string, available []string) string {
	if suggestion != "" {
		fmt.Printf("[%s]\n  Suggestion: %s\n  (Enter=accept, s=skip, ?=search): ", label, suggestion)
	} else {
		fmt.Printf("[%s]\n  (s=skip, ?=search): ", label)
	}

	if !scanner.Scan() {
		return suggestion
	}
	input := strings.TrimSpace(scanner.Text())

	switch {
	case input == "s":
		return ""
	case input == "" && suggestion != "":
		return suggestion
	case input == "?":
		picked, ok := Pick(available)
		if ok {
			fmt.Printf("  → %s\n", picked)
			return picked
		}
		return ""
	default:
		// Exact match
		for _, a := range available {
			if a == input {
				return input
			}
		}
		fmt.Printf("  '%s' not found locally. Use anyway? (y/n): ", input)
		if scanner.Scan() && strings.TrimSpace(scanner.Text()) == "y" {
			return input
		}
		return ""
	}
}
