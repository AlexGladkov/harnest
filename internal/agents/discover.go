package agents

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Discover scans all installed agents from known locations.
// Returns sorted list of agent names found on disk.
func Discover() []string {
	seen := map[string]bool{}
	add := func(name string) {
		if name != "" && !seen[name] {
			seen[name] = true
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	// 1. Custom agents: ~/.claude/agents/*.md
	scanFlat(filepath.Join(home, ".claude", "agents"), "", add)
	scanFlat(filepath.Join(home, ".cursor", "agents"), "", add)
	scanFlat(filepath.Join(home, ".windsurf", "agents"), "", add)

	// 2. All plugins: walk ~/.claude/plugins/cache/ for plugin.json with "agents" field
	pluginsDir := filepath.Join(home, ".claude", "plugins", "cache")
	scanPlugins(pluginsDir, add)

	agents := make([]string, 0, len(seen))
	for name := range seen {
		agents = append(agents, name)
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

type pluginJSON struct {
	Name   string   `json:"name"`
	Agents []string `json:"agents"`
}

// scanPlugins walks plugins dir, finds all plugin.json, extracts agents.
func scanPlugins(root string, add func(string)) {
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() || filepath.Base(path) != "plugin.json" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		var p pluginJSON
		if err := json.Unmarshal(data, &p); err != nil || len(p.Agents) == 0 {
			return nil
		}

		for _, agentPath := range p.Agents {
			// agentPath is like "./vue-expert.md"
			base := filepath.Base(agentPath)
			name := strings.TrimSuffix(base, ".md")
			if name == "" || name == "README" {
				continue
			}
			// Verify file exists relative to plugin.json dir
			pluginDir := filepath.Dir(filepath.Dir(path)) // up from .claude-plugin/
			// Actually plugin.json is in .claude-plugin/ dir, agents are relative to parent
			agentFile := filepath.Join(filepath.Dir(path), "..", agentPath)
			if _, err := os.Stat(agentFile); err != nil {
				continue
			}
			if p.Name != "" {
				add(p.Name + ":" + name)
			} else {
				add(name)
			}
			_ = pluginDir
		}
		return nil
	})
}

// scanFlat reads *.md from dir, adds prefix+basename.
func scanFlat(dir, prefix string, add func(string)) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".md")
		if name == "README" || name == "" {
			continue
		}
		add(prefix + name)
	}
}
