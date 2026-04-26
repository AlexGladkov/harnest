package wizard

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/AlexGladkov/harnest/internal/agents"
	"github.com/AlexGladkov/harnest/internal/mapping"
)

const maxSuggestions = 5

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
		return searchMode(scanner, available)
	default:
		// Exact match check
		for _, a := range available {
			if a == input {
				return input
			}
		}
		// Not found — use as literal
		fmt.Printf("  '%s' not found locally. Use anyway? (y/n): ", input)
		if scanner.Scan() && strings.TrimSpace(scanner.Text()) == "y" {
			return input
		}
		return ""
	}
}

// searchMode: interactive search loop. Type to filter, see top 5, refine or pick.
func searchMode(scanner *bufio.Scanner, available []string) string {
	// Show initial top 5 so user sees what's available
	show := available
	if len(show) > maxSuggestions {
		show = show[:maxSuggestions]
	}
	for i, a := range show {
		fmt.Printf("    %d) %s\n", i+1, a)
	}
	if len(available) > maxSuggestions {
		fmt.Printf("    ... and %d more (type to filter)\n", len(available)-maxSuggestions)
	}
	fmt.Print("  Search or pick number: ")
	for {
		if !scanner.Scan() {
			return ""
		}
		query := strings.TrimSpace(scanner.Text())
		if query == "" {
			return ""
		}

		results := agents.Search(available, query)

		if len(results) == 0 {
			fmt.Printf("  No match for '%s'. Try again or Enter to cancel: ", query)
			continue
		}

		if len(results) == 1 {
			fmt.Printf("  → %s\n", results[0])
			return results[0]
		}

		// Show top 5
		show := results
		if len(show) > maxSuggestions {
			show = show[:maxSuggestions]
		}
		for i, a := range show {
			fmt.Printf("    %d) %s\n", i+1, a)
		}
		if len(results) > maxSuggestions {
			fmt.Printf("    ... and %d more\n", len(results)-maxSuggestions)
		}

		fmt.Print("  Pick number, refine, or Enter to cancel: ")
		if !scanner.Scan() {
			return ""
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			return ""
		}

		// Try number
		idx := 0
		if _, err := fmt.Sscanf(input, "%d", &idx); err == nil && idx >= 1 && idx <= len(show) {
			fmt.Printf("  → %s\n", show[idx-1])
			return show[idx-1]
		}

		// Exact match
		for _, a := range available {
			if a == input {
				fmt.Printf("  → %s\n", a)
				return a
			}
		}

		// Refine — loop again with new query
		results = agents.Search(available, input)
		if len(results) == 1 {
			fmt.Printf("  → %s\n", results[0])
			return results[0]
		}
		if len(results) == 0 {
			fmt.Printf("  No match for '%s'. Try again or Enter to cancel: ", input)
			continue
		}

		show = results
		if len(show) > maxSuggestions {
			show = show[:maxSuggestions]
		}
		for i, a := range show {
			fmt.Printf("    %d) %s\n", i+1, a)
		}
		if len(results) > maxSuggestions {
			fmt.Printf("    ... and %d more\n", len(results)-maxSuggestions)
		}
		fmt.Print("  Pick number, refine, or Enter to cancel: ")
	}
}
