package agents

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestDiscover(t *testing.T) {
	all := Discover("")
	t.Logf("Found %d agents", len(all))
	for _, a := range all {
		t.Logf("  %s", a)
	}
	if len(all) == 0 {
		t.Log("No agents found (may be expected in CI)")
	}
}

func TestDiscoverProject(t *testing.T) {
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".claude", "agents")
	mustMkdir(t, agentsDir)

	// Agent 1: with frontmatter name (different from filename)
	writeFile(t, filepath.Join(agentsDir, "backend.md"), `---
name: backend-csharp
description: Backend specialist for C#
---
Some agent instructions.
`)

	// Agent 2: with frontmatter name matching filename
	writeFile(t, filepath.Join(agentsDir, "vue-expert.md"), `---
name: vue-expert
description: Vue.js frontend specialist
---
Vue expert instructions.
`)

	// Agent 3: no frontmatter — falls back to filename
	writeFile(t, filepath.Join(agentsDir, "no-fm.md"), "Just markdown content.\nNo frontmatter here.")

	// Agent 4: README.md — should be skipped
	writeFile(t, filepath.Join(agentsDir, "README.md"), `---
name: readme-agent
---
This is a README, should be skipped.
`)

	// Agent 5: frontmatter has no name field — fallback to filename
	writeFile(t, filepath.Join(agentsDir, "fallback.md"), `---
description: No name field here
---
Uses filename as agent name.
`)

	// Agent 6: non-.md file — should be skipped
	writeFile(t, filepath.Join(agentsDir, "plugin.json"), `{"name": "test"}`)

	agents := DiscoverProject(tmpDir)
	sort.Strings(agents)

	expected := []string{"backend-csharp", "fallback", "no-fm", "vue-expert"}
	if len(agents) != len(expected) {
		t.Fatalf("expected %d agents, got %d: %v", len(expected), len(agents), agents)
	}
	for i, a := range agents {
		if a != expected[i] {
			t.Errorf("agent[%d]: expected %q, got %q", i, expected[i], a)
		}
	}
}

func TestDiscoverProject_EmptyDir(t *testing.T) {
	emptyDir := filepath.Join(t.TempDir(), ".claude", "agents")
	mustMkdir(t, emptyDir)
	agents := DiscoverProject(filepath.Dir(emptyDir))
	if len(agents) != 0 {
		t.Errorf("expected 0 agents from empty dir, got %d", len(agents))
	}
}

func TestDiscoverProject_NonExistent(t *testing.T) {
	agents := DiscoverProject(filepath.Join(t.TempDir(), "nonexistent"))
	if len(agents) != 0 {
		t.Errorf("expected 0 agents from nonexistent dir, got %d", len(agents))
	}
}

func TestDiscoverProject_MultiHarness(t *testing.T) {
	// Agents in different harness dirs should all be discovered
	tmpDir := t.TempDir()

	// claude-code
	claudeDir := filepath.Join(tmpDir, ".claude", "agents")
	mustMkdir(t, claudeDir)
	writeFile(t, filepath.Join(claudeDir, "claude-agent.md"), "---\nname: claude-expert\n---\nbody")

	// cursor
	cursorDir := filepath.Join(tmpDir, ".cursor", "agents")
	mustMkdir(t, cursorDir)
	writeFile(t, filepath.Join(cursorDir, "cursor-agent.md"), "---\nname: cursor-expert\n---\nbody")

	// windsurf
	windsurfDir := filepath.Join(tmpDir, ".windsurf", "agents")
	mustMkdir(t, windsurfDir)
	writeFile(t, filepath.Join(windsurfDir, "windsurf-agent.md"), "---\nname: windsurf-expert\n---\nbody")

	// codex
	codexDir := filepath.Join(tmpDir, ".codex", "agents")
	mustMkdir(t, codexDir)
	writeFile(t, filepath.Join(codexDir, "codex-agent.md"), "---\nname: codex-expert\n---\nbody")

	// opencode
	opencodeDir := filepath.Join(tmpDir, ".config", "opencode", "agents")
	mustMkdir(t, opencodeDir)
	writeFile(t, filepath.Join(opencodeDir, "opencode-agent.md"), "---\nname: opencode-expert\n---\nbody")

	// qwen-code
	qwenDir := filepath.Join(tmpDir, ".qwen", "agents")
	mustMkdir(t, qwenDir)
	writeFile(t, filepath.Join(qwenDir, "qwen-agent.md"), "---\nname: qwen-expert\n---\nbody")

	agents := DiscoverProject(tmpDir)

	if len(agents) != 6 {
		t.Fatalf("expected 6 agents from all harness dirs, got %d: %v", len(agents), agents)
	}

	expected := []string{"claude-expert", "codex-expert", "cursor-expert", "opencode-expert", "qwen-expert", "windsurf-expert"}
	for i, a := range agents {
		if a != expected[i] {
			t.Errorf("agent[%d]: expected %q, got %q", i, expected[i], a)
		}
	}
}

func TestParseAgentName(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		filename string
		content  string
		expected string
	}{
		{
			filename: "my-agent.md",
			content:  "---\nname: my-agent\ndescription: test\n---\nbody",
			expected: "my-agent",
		},
		{
			filename: "crlf.md",
			content:  "---\r\nname: crlf-agent\r\ndescription: test\r\n---\r\nbody",
			expected: "crlf-agent",
		},
		{
			filename: "no-fm.md",
			content:  "Just some markdown content.\nNo frontmatter here.",
			expected: "no-fm", // no frontmatter → falls back to filename
		},
		{
			filename: "fallback.md",
			content:  "---\ndescription: only description\n---\nbody",
			expected: "fallback", // falls back to filename
		},
		{
			filename: "unnamed.md",
			content:  "---\nname: \"\"\ndescription: test\n---\nbody",
			expected: "unnamed", // empty name → falls back to filename
		},
		{
			filename: "broken.md",
			content:  "---\nname: broken\n",
			expected: "broken", // no closing delimiter → falls back to filename
		},
		{
			filename: "deep-nested.md",
			content:  "---\nname: ok\ndep1:\n  dep2:\n    dep3:\n      dep4: deep\n---\nbody",
			expected: "ok", // deeply nested but valid YAML
		},
		{
			filename: "unicode.md",
			content:  "---\nname: агент-тест\ndescription: тестирование\n---\nbody",
			expected: "агент-тест", // unicode name
		},
		{
			filename: "binary-null.md",
			content:  "---\nname: bad\x00agent\n---\nbody",
			expected: "", // null byte → rejected
		},
		{
			filename: "huge-frontmatter.md",
			content:  "---\nname: ok\n" + strings.Repeat("data: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\n", 100) + "---\nbody",
			expected: "huge-frontmatter", // frontmatter > 4KB → falls back to filename
		},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			filePath := filepath.Join(tmpDir, tt.filename)
			writeFile(t, filePath, tt.content)

			got := parseAgentName(filePath)
			if got != tt.expected {
				t.Errorf("parseAgentName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestIndexClosingDelim(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want int
	}{
		{"clean", "name: x\n---\nbody", 7},
		{"eof", "name: x\n---", 7},
		{"crlf", "name: x\n---\r\nbody", 7},
		{"dashes-skipped", "name: x\n----\n---\nbody", 12},
		{"text-after-skipped", "name: x\n---bad\n---\nbody", 14},
		{"none", "name: x\nno delim", -1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := indexClosingDelim(c.in); got != c.want {
				t.Errorf("indexClosingDelim(%q) = %d, want %d", c.in, got, c.want)
			}
		})
	}
}

func TestScanPlugins_AgentsDirectory(t *testing.T) {
	// Simulate plugin cache: <root>/marketplace/plugin/1.0.0/.claude-plugin/plugin.json
	// with agents/ at version level
	root := t.TempDir()

	pluginDir := filepath.Join(root, "marketplace", "my-plugin", "1.0.0", ".claude-plugin")
	mustMkdir(t, pluginDir)

	// plugin.json — no "agents" field (modern style)
	writeFile(t, filepath.Join(pluginDir, "plugin.json"), `{
  "name": "my-plugin",
  "version": "1.0.0"
}`)

	// agents/ directory at version level
	versionDir := filepath.Dir(pluginDir)
	agentsDir := filepath.Join(versionDir, "agents")
	mustMkdir(t, agentsDir)
	writeFile(t, filepath.Join(agentsDir, "alpha.md"), "---\nname: alpha\n---\nbody")
	writeFile(t, filepath.Join(agentsDir, "beta.md"), "---\nname: beta\n---\nbody")
	writeFile(t, filepath.Join(agentsDir, "README.md"), "# README")

	agents := scanPlugins(root)

	// scanFlatNames uses filename (not frontmatter), namespaced as plugin:agent
	if !contains(agents, "my-plugin:alpha") {
		t.Error("expected my-plugin:alpha to be discovered from agents/ dir")
	}
	if !contains(agents, "my-plugin:beta") {
		t.Error("expected my-plugin:beta to be discovered from agents/ dir")
	}
	if contains(agents, "my-plugin:README") {
		t.Error("README.md should be skipped")
	}
}

func TestScanPlugins_ExplicitAgentsField(t *testing.T) {
	// Backward compat: plugin.json with explicit "agents" field
	root := t.TempDir()

	pluginDir := filepath.Join(root, "marketplace", "old-plugin", "2.0.0", ".claude-plugin")
	mustMkdir(t, pluginDir)

	writeFile(t, filepath.Join(pluginDir, "plugin.json"), `{
  "name": "old-plugin",
  "agents": ["./custom-agent.md", "./specialist.md"]
}`)

	// Create agent files relative to version dir (parent of .claude-plugin/)
	versionDir := filepath.Dir(pluginDir)
	writeFile(t, filepath.Join(versionDir, "custom-agent.md"), "# Custom Agent")
	writeFile(t, filepath.Join(versionDir, "specialist.md"), "# Specialist")

	agents := scanPlugins(root)

	if !contains(agents, "old-plugin:custom-agent") {
		t.Error("expected old-plugin:custom-agent from explicit agents field")
	}
	if !contains(agents, "old-plugin:specialist") {
		t.Error("expected old-plugin:specialist from explicit agents field")
	}
}

func TestScanPlugins_NoPluginName(t *testing.T) {
	// plugin.json without "name" field — should be skipped entirely
	root := t.TempDir()

	pluginDir := filepath.Join(root, "marketplace", "anon", "1.0.0", ".claude-plugin")
	mustMkdir(t, pluginDir)
	writeFile(t, filepath.Join(pluginDir, "plugin.json"), `{
  "version": "1.0.0"
}`)

	// Create agents/ directory — should NOT be scanned (no plugin name)
	versionDir := filepath.Dir(pluginDir)
	agentsDir := filepath.Join(versionDir, "agents")
	mustMkdir(t, agentsDir)
	writeFile(t, filepath.Join(agentsDir, "ghost.md"), "---\nname: ghost\n---\nbody")

	agents := scanPlugins(root)

	if len(agents) != 0 {
		t.Errorf("expected 0 agents from nameless plugin, got %d: %v", len(agents), agents)
	}
}

func TestScanPlugins_EmptyAgentsField(t *testing.T) {
	// plugin.json with empty agents array but agents/ dir → dir scan finds them
	root := t.TempDir()

	pluginDir := filepath.Join(root, "marketplace", "mixed", "1.0.0", ".claude-plugin")
	mustMkdir(t, pluginDir)
	writeFile(t, filepath.Join(pluginDir, "plugin.json"), `{
  "name": "mixed-plugin",
  "agents": []
}`)

	versionDir := filepath.Dir(pluginDir)
	agentsDir := filepath.Join(versionDir, "agents")
	mustMkdir(t, agentsDir)
	writeFile(t, filepath.Join(agentsDir, "from-dir.md"), "---\nname: from-dir\n---\nbody")

	agents := scanPlugins(root)

	if !contains(agents, "mixed-plugin:from-dir") {
		t.Error("expected mixed-plugin:from-dir from agents/ dir despite empty agents field")
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatal(err)
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
