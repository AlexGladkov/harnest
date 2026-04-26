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
	fmt.Println("Enter = accept suggestion, s = skip, or type to search\n")

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

// pickAgent interactive loop: type to search, Enter to accept, s to skip.
func pickAgent(scanner *bufio.Scanner, label, suggestion string, available []string) string {
	for {
		fmt.Printf("[%s]\n", label)
		if suggestion != "" {
			fmt.Printf("  Suggestion: %s\n", suggestion)
			fmt.Print("  Enter=accept, s=skip, or type to search: ")
		} else {
			fmt.Print("  Type to search, s=skip: ")
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
		case input == "":
			continue
		default:
			// Exact match — accept immediately
			for _, a := range available {
				if a == input {
					return input
				}
			}

			// Search and show results
			result := searchLoop(scanner, available, input)
			if result != "" {
				return result
			}
			// User went back — re-show prompt
			continue
		}
	}
}

// searchLoop shows filtered results, user refines or picks.
func searchLoop(scanner *bufio.Scanner, available []string, query string) string {
	for {
		results := agents.Search(available, query)

		if len(results) == 0 {
			fmt.Printf("  No match for '%s'. Refine or Enter to go back: ", query)
		} else if len(results) == 1 {
			// Single match — auto-select
			fmt.Printf("  → %s\n", results[0])
			return results[0]
		} else {
			fmt.Printf("  Matches for '%s':\n", query)
			for _, a := range results {
				fmt.Printf("    %s\n", a)
			}
			fmt.Print("  Refine search or Enter to go back: ")
		}

		if !scanner.Scan() {
			return ""
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			return ""
		}

		// Check exact match
		for _, a := range available {
			if a == input {
				fmt.Printf("  → %s\n", a)
				return a
			}
		}

		// Refine
		query = input
	}
}
