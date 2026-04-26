package wizard

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/AlexGladkov/harnest/internal/agents"
	"github.com/AlexGladkov/harnest/internal/mapping"
)

const maxShow = 5

func Run(r io.Reader, structure mapping.AgentStructure, suggestions mapping.Suggestions) mapping.AgentConfig {
	scanner := bufio.NewScanner(r)
	config := mapping.AgentConfig{}

	available := agents.Discover()

	fmt.Println("\n── Agent Wizard ──")
	fmt.Printf("Found %d agents on this machine\n", len(available))
	fmt.Println("Enter = accept suggestion, s = skip, ? = search\n")

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
	fmt.Printf("[%s]\n", label)
	if suggestion != "" {
		fmt.Printf("  Suggestion: %s\n", suggestion)
		fmt.Print("  Enter=accept, s=skip, ?=search: ")
	} else {
		fmt.Print("  s=skip, ?=search: ")
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
		return searchLoop(scanner, available)
	default:
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

func searchLoop(scanner *bufio.Scanner, available []string) string {
	fmt.Print("  Type to filter: ")

	for {
		if !scanner.Scan() {
			return ""
		}
		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			return "" // cancel
		}

		// Try as number from last shown results
		results := agents.Search(available, "")
		// We need to re-filter with whatever was shown last
		// But since we don't track state, treat input as search query first

		// Check exact match
		for _, a := range available {
			if a == input {
				fmt.Printf("  → %s\n", a)
				return a
			}
		}

		// Filter
		results = agents.Search(available, input)

		if len(results) == 0 {
			fmt.Printf("  No match for '%s'\n", input)
			fmt.Print("  Try again or Enter to cancel: ")
			continue
		}

		// Show filtered results — always, even if 1 match
		showResults(results, input)
		if len(results) == 1 {
			fmt.Print("  Enter=accept, or refine: ")
		} else {
			fmt.Print("  Pick #, refine, or Enter to cancel: ")
		}

		if !scanner.Scan() {
			return ""
		}
		pick := strings.TrimSpace(scanner.Text())

		if pick == "" {
			if len(results) == 1 {
				fmt.Printf("  → %s\n", results[0])
				return results[0]
			}
			return ""
		}

		// Try as number
		idx := 0
		show := results
		if len(show) > maxShow {
			show = show[:maxShow]
		}
		if _, err := fmt.Sscanf(pick, "%d", &idx); err == nil && idx >= 1 && idx <= len(show) {
			fmt.Printf("  → %s\n", show[idx-1])
			return show[idx-1]
		}

		// Exact match
		for _, a := range available {
			if a == pick {
				fmt.Printf("  → %s\n", a)
				return a
			}
		}

		// Treat as refined search
		results = agents.Search(available, pick)
		if len(results) == 0 {
			fmt.Printf("  No match for '%s'\n", pick)
			fmt.Print("  Try again or Enter to cancel: ")
			continue
		}

		showResults(results, pick)
		if len(results) == 1 {
			fmt.Print("  Enter=accept, or refine: ")
		} else {
			fmt.Print("  Pick #, refine, or Enter to cancel: ")
		}
	}
}

func showResults(items []string, query string) {
	show := items
	if len(show) > maxShow {
		show = show[:maxShow]
	}
	if query != "" {
		fmt.Printf("  '%s' → %d matches:\n", query, len(items))
	}
	for i, a := range show {
		fmt.Printf("    %d) %s\n", i+1, a)
	}
	if len(items) > maxShow {
		fmt.Printf("    ... +%d more\n", len(items)-maxShow)
	}
}
