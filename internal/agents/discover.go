package agents

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Discover scans installed agents from known harness locations.
// Returns sorted list of agent names found on disk.
func Discover() []string {
	var agents []string
	seen := map[string]bool{}

	for _, dir := range agentDirs() {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			name := strings.TrimSuffix(e.Name(), ".md")
			if name != "" && !seen[name] {
				agents = append(agents, name)
				seen[name] = true
			}
		}
	}

	sort.Strings(agents)
	return agents
}

// Search filters agents by substring match (case-insensitive).
func Search(agents []string, query string) []string {
	if query == "" {
		return agents
	}
	q := strings.ToLower(query)
	var results []string
	for _, a := range agents {
		if strings.Contains(strings.ToLower(a), q) {
			results = append(results, a)
		}
	}
	return results
}

func agentDirs() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	return []string{
		// Claude Code
		filepath.Join(home, ".claude", "agents"),
		// Cursor
		filepath.Join(home, ".cursor", "agents"),
		// Windsurf
		filepath.Join(home, ".windsurf", "agents"),
	}
}
