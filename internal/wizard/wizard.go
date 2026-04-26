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
	fmt.Println("Enter = accept suggestion, s = skip, ? = search agents\n")

	// Consilium roles
	for _, role := range structure.Roles {
		suggestion := suggestions.Consilium[role]
		agent := promptRole(scanner, role, suggestion, available)
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
		agent := promptExec(scanner, es, suggestion, available)
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

func promptRole(scanner *bufio.Scanner, role, suggestion string, available []string) string {
	for {
		if suggestion != "" {
			fmt.Printf("[Consilium: %s]\n  Suggestion: %s\n  (Enter=accept, s=skip, ?=search): ", role, suggestion)
		} else {
			fmt.Printf("[Consilium: %s]\n  (type to search, s=skip, ?=list all): ", role)
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
		case input == "?" || (input == "" && suggestion == ""):
			picked := searchAndPick(scanner, available, "")
			if picked != "" {
				return picked
			}
			continue
		default:
			// Check if input matches an agent or is a search query
			for _, a := range available {
				if a == input {
					return input
				}
			}
			// Try as search
			results := agents.Search(available, input)
			if len(results) == 1 {
				fmt.Printf("  → %s\n", results[0])
				return results[0]
			}
			if len(results) > 1 {
				picked := searchAndPick(scanner, available, input)
				if picked != "" {
					return picked
				}
				continue
			}
			// No match — use as literal agent name
			fmt.Printf("  Agent '%s' not found locally. Use anyway? (y/n): ", input)
			if scanner.Scan() && strings.TrimSpace(scanner.Text()) == "y" {
				return input
			}
			continue
		}
	}
}

func promptExec(scanner *bufio.Scanner, es mapping.ExecScope, suggestion string, available []string) string {
	for {
		if suggestion != "" {
			fmt.Printf("[Exec: %s → %s]\n  Suggestion: %s\n  (Enter=accept, s=skip, ?=search): ", es.StackName, es.Scope, suggestion)
		} else {
			fmt.Printf("[Exec: %s → %s]\n  (type to search, s=skip, ?=list all): ", es.StackName, es.Scope)
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
		case input == "?" || (input == "" && suggestion == ""):
			picked := searchAndPick(scanner, available, "")
			if picked != "" {
				return picked
			}
			continue
		default:
			for _, a := range available {
				if a == input {
					return input
				}
			}
			results := agents.Search(available, input)
			if len(results) == 1 {
				fmt.Printf("  → %s\n", results[0])
				return results[0]
			}
			if len(results) > 1 {
				picked := searchAndPick(scanner, available, input)
				if picked != "" {
					return picked
				}
				continue
			}
			fmt.Printf("  Agent '%s' not found locally. Use anyway? (y/n): ", input)
			if scanner.Scan() && strings.TrimSpace(scanner.Text()) == "y" {
				return input
			}
			continue
		}
	}
}

// searchAndPick shows filtered list, user picks by number or refines search.
func searchAndPick(scanner *bufio.Scanner, available []string, query string) string {
	for {
		results := agents.Search(available, query)
		if len(results) == 0 {
			fmt.Printf("  No agents matching '%s'. Try again (or Enter to go back): ", query)
			if !scanner.Scan() {
				return ""
			}
			query = strings.TrimSpace(scanner.Text())
			if query == "" {
				return ""
			}
			continue
		}

		fmt.Printf("  Agents")
		if query != "" {
			fmt.Printf(" matching '%s'", query)
		}
		fmt.Printf(" (%d):\n", len(results))
		for i, a := range results {
			fmt.Printf("  %3d) %s\n", i+1, a)
		}

		fmt.Print("  Pick number, refine search, or Enter to go back: ")
		if !scanner.Scan() {
			return ""
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			return ""
		}

		// Try as number
		idx := 0
		if _, err := fmt.Sscanf(input, "%d", &idx); err == nil && idx >= 1 && idx <= len(results) {
			picked := results[idx-1]
			fmt.Printf("  → %s\n", picked)
			return picked
		}

		// Refine search
		query = input
	}
}
