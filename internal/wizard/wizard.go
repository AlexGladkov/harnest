package wizard

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/AlexGladkov/harnest/internal/mapping"
)

func Run(r io.Reader, structure mapping.AgentStructure, suggestions mapping.Suggestions) mapping.AgentConfig {
	scanner := bufio.NewScanner(r)
	config := mapping.AgentConfig{}

	fmt.Println("\n── Agent Wizard ──")
	fmt.Println("Enter = accept suggestion, s = skip, or type agent name\n")

	// Consilium roles
	for _, role := range structure.Roles {
		suggestion := suggestions.Consilium[role]
		agent := promptRole(scanner, role, suggestion)
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
		agent := promptExec(scanner, es, suggestion)
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

func promptRole(scanner *bufio.Scanner, role, suggestion string) string {
	if suggestion != "" {
		fmt.Printf("[Consilium: %s]\n  Suggestion: %s\n  Enter agent (Enter=suggestion, s=skip): ", role, suggestion)
	} else {
		fmt.Printf("[Consilium: %s]\n  No suggestion\n  Enter agent (s=skip): ", role)
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
	case input == "" && suggestion == "":
		return ""
	default:
		return input
	}
}

func promptExec(scanner *bufio.Scanner, es mapping.ExecScope, suggestion string) string {
	if suggestion != "" {
		fmt.Printf("[Exec: %s → %s]\n  Suggestion: %s\n  Enter agent (Enter=suggestion, s=skip): ", es.StackName, es.Scope, suggestion)
	} else {
		fmt.Printf("[Exec: %s → %s]\n  No suggestion\n  Enter agent (s=skip): ", es.StackName, es.Scope)
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
	case input == "" && suggestion == "":
		return ""
	default:
		return input
	}
}
