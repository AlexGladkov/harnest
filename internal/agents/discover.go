package agents

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/AlexGladkov/harnest/internal/harness"
	"gopkg.in/yaml.v3"
)

// Discover scans all installed agents: project-local (priority), global, and plugins.
// projectDir may be empty — project scan is skipped.
// Returns agents in priority order: project-local first, then global, then plugins.
// Insertion order is preserved — project agents win dedup ties in MatchAgent.
func Discover(projectDir string) []string {
	seen := map[string]bool{}
	var result []string

	add := func(name string) {
		if name != "" && !seen[name] {
			seen[name] = true
			result = append(result, name)
		}
	}

	// 1. Project-local agents (priority — added first, win on dedup)
	if projectDir != "" {
		for _, name := range scanProjectAgents(projectDir) {
			add(name)
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return result
	}

	// 2. Global agents from all registered harness locations
	for _, dir := range harness.AgentDirs() {
		scanFlat(filepath.Join(home, dir), "", add)
	}

	// 3. All plugins: walk ~/.claude/plugins/cache/ for plugin.json
	for _, name := range scanPlugins(filepath.Join(home, ".claude", "plugins", "cache")) {
		add(name)
	}

	return result
}

// scanProjectAgents scans project-local agents across all registered harness dirs
// (e.g. .claude/agents/, .cursor/agents/, .windsurf/agents/), preserving harness-dir
// order and deduplicating. Reads YAML frontmatter: if "name" is present it is used,
// otherwise the filename. Single source of truth shared by Discover (which keeps this
// order so project agents win dedup ties) and DiscoverProject.
func scanProjectAgents(projectDir string) []string {
	seen := map[string]bool{}
	var result []string
	for _, dir := range harness.AgentDirs() {
		agentsDir := filepath.Join(projectDir, dir)
		for _, name := range scanWithFrontmatter(agentsDir, "") {
			if name != "" && !seen[name] {
				seen[name] = true
				result = append(result, name)
			}
		}
	}
	return result
}

// DiscoverProject returns project-local agents (see scanProjectAgents), sorted.
func DiscoverProject(projectDir string) []string {
	agents := scanProjectAgents(projectDir)
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

// --- plugin scanning ---

type pluginJSON struct {
	Name   string   `json:"name"`
	Agents []string `json:"agents"`
}

// scanPlugins walks plugins cache dir, finds plugin.json files, extracts agent names.
// Agents are discovered from:
//  1. <plugin-version>/agents/*.md — auto-discovered by Claude Code (no plugin.json entry needed)
//  2. plugin.json "agents" field — explicit agent list (backward compat)
//
// Agent names are namespaced as "pluginName:agentName".
func scanPlugins(root string) []string {
	seen := map[string]bool{}
	var result []string
	add := func(name string) {
		if name != "" && !seen[name] {
			seen[name] = true
			result = append(result, name)
		}
	}
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
		if err := json.Unmarshal(data, &p); err != nil || p.Name == "" {
			return nil
		}

		// plugin.json is at <version>/.claude-plugin/plugin.json
		// version dir = two levels up from plugin.json
		versionDir := filepath.Dir(filepath.Dir(path))

		// 1. Scan agents/ directory (primary method — Claude Code auto-discovers these)
		agentsDir := filepath.Join(versionDir, "agents")
		scanFlat(agentsDir, p.Name+":", add)

		// 2. Also process explicit "agents" field for backward compatibility
		for _, agentPath := range p.Agents {
			base := filepath.Base(agentPath)
			name := strings.TrimSuffix(base, ".md")
			if name == "" || name == "README" {
				continue
			}
			// Verify file exists relative to plugin.json dir
			agentFile := filepath.Join(filepath.Dir(path), "..", agentPath)
			if _, err := os.Stat(agentFile); err != nil {
				continue
			}
			add(p.Name + ":" + name)
		}
		return nil
	})
	sort.Strings(result)
	return result
}

// --- flat scanning ---

// scanFlat reads *.md from dir, adds prefix+basename (without .md).
// Skips README.md and empty names.
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

// --- frontmatter-aware scanning ---

// frontmatter represents YAML frontmatter in agent .md files.
type frontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// scanWithFrontmatter reads *.md files from dir, parses YAML frontmatter.
// If frontmatter has "name" field → uses it. Otherwise falls back to filename.
// Only unreadable/empty/binary/oversized files are skipped.
// Agent names are prefixed with prefix (e.g. plugin name + ":").
func scanWithFrontmatter(dir, prefix string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var agents []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}

		base := strings.TrimSuffix(e.Name(), ".md")
		if base == "README" || base == "" {
			continue
		}

		filePath := filepath.Join(dir, e.Name())
		name := parseAgentName(filePath)
		if name == "" {
			continue
		}
		agents = append(agents, prefix+name)
	}
	sort.Strings(agents)
	return agents
}

// maxAgentFileSize limits how much of an agent .md file we read for frontmatter.
const maxAgentFileSize = 64 * 1024 // 64 KB

// parseAgentName reads a .md file, extracts "name" from YAML frontmatter.
// Frontmatter is YAML between --- delimiters at the start of the file.
// Falls back to filename (without .md) when:
//   - No frontmatter at all
//   - No closing "---"
//   - Frontmatter YAML is invalid
//   - Frontmatter has no "name" field
//
// Returns empty string only for unreadable files, empty files, binary files, or oversized files.
func parseAgentName(filePath string) string {
	// Stat first so oversized/empty files are rejected without buffering them in RAM.
	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() || info.Size() == 0 || info.Size() > maxAgentFileSize {
		return ""
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return ""
	}

	// Reject binary files (null bytes).
	if bytesContainsNull(data) {
		return ""
	}

	// Fallback to filename (used when frontmatter is absent or unparseable).
	base := filepath.Base(filePath)
	fallback := strings.TrimSuffix(base, ".md")
	if fallback == "README" || fallback == "" {
		return ""
	}

	content := string(data)
	// No frontmatter → fallback to filename.
	if !strings.HasPrefix(content, "---") {
		return fallback
	}

	// Strip opening "---" and following newline(s)
	content = strings.TrimPrefix(content, "---")
	content = strings.TrimLeft(content, "\r\n")

	// Find closing "---" on its own line (delimiter must be followed by EOL or EOF,
	// so a value containing "----" or "\n---text" does not truncate frontmatter early).
	endIdx := indexClosingDelim(content)
	if endIdx == -1 {
		return fallback
	}

	// Limit frontmatter size — anything beyond 4KB is not a real agent definition.
	fm := content[:endIdx]
	if len(fm) > 4096 {
		return fallback
	}

	var fmData frontmatter
	if err := yaml.Unmarshal([]byte(fm), &fmData); err != nil {
		return fallback
	}

	if fmData.Name != "" {
		return fmData.Name
	}

	return fallback
}

// indexClosingDelim returns the index of a "\n---" frontmatter terminator whose
// "---" stands on its own line (next byte is \n, \r, or EOF), or -1 if none.
func indexClosingDelim(content string) int {
	base := 0
	for {
		i := strings.Index(content[base:], "\n---")
		if i == -1 {
			return -1
		}
		abs := base + i
		after := abs + 4 // byte past "\n---"
		if after >= len(content) || content[after] == '\n' || content[after] == '\r' {
			return abs
		}
		base = after
	}
}

// bytesContainsNull checks if data contains a null byte (binary content).
func bytesContainsNull(data []byte) bool {
	for _, b := range data {
		if b == 0 {
			return true
		}
	}
	return false
}
