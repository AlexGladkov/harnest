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

	add := func(name string) {
		if name != "" && !seen[name] {
			agents = append(agents, name)
			seen[name] = true
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	// 1. Custom agents: ~/.claude/agents/*.md, ~/.cursor/agents/*.md, etc.
	for _, dir := range []string{
		filepath.Join(home, ".claude", "agents"),
		filepath.Join(home, ".cursor", "agents"),
		filepath.Join(home, ".windsurf", "agents"),
	} {
		scanFlat(dir, "", add)
	}

	// 2. VoltAgent plugins: ~/.claude/plugins/cache/voltagent-subagents/<group>/<version>/*.md
	voltDir := filepath.Join(home, ".claude", "plugins", "cache", "voltagent-subagents")
	groups, err := os.ReadDir(voltDir)
	if err == nil {
		for _, g := range groups {
			if !g.IsDir() {
				continue
			}
			groupName := g.Name() // e.g. "voltagent-lang"
			// Find latest version dir
			versionDir := latestVersionDir(filepath.Join(voltDir, groupName))
			if versionDir != "" {
				scanFlat(versionDir, groupName+":", add)
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

// scanFlat reads *.md from dir, adds prefix+basename to callback.
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

// latestVersionDir returns path to the latest semver dir inside parent.
func latestVersionDir(parent string) string {
	entries, err := os.ReadDir(parent)
	if err != nil {
		return ""
	}
	// Pick last dir alphabetically (semver sorts correctly for simple cases)
	var latest string
	for _, e := range entries {
		if e.IsDir() {
			latest = filepath.Join(parent, e.Name())
		}
	}
	return latest
}
