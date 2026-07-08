// Package drift detects configuration drift between a project's recorded
// agent assignments (CLAUDE.md / .cursorrules / etc.) and the current
// state of the repository and installed agents.
package drift

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	agents_pkg "github.com/AlexGladkov/harnest/internal/agents"
	"github.com/AlexGladkov/harnest/internal/config"
	"github.com/AlexGladkov/harnest/internal/detector"
	"github.com/AlexGladkov/harnest/internal/mapping"
)

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

// DriftType identifies the category of a drift item.
type DriftType string

const (
	// DriftStack is raised when a newly detected stack has no exec agent covering it.
	DriftStack DriftType = "stack"

	// DriftScope is raised when an exec agent's glob scope matches zero files.
	DriftScope DriftType = "scope"

	// DriftAgent is raised when a configured agent name cannot be found in the
	// set of discovered (installed) agents.
	DriftAgent DriftType = "agent_availability"
)

// Severity communicates urgency to the end user.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// DriftItem describes a single detected drift.
type DriftItem struct {
	// Type classifies the drift.
	Type DriftType `json:"type"`

	// Severity indicates how critical the drift is.
	Severity Severity `json:"severity"`

	// Current is the value recorded in the config file.
	Current string `json:"current,omitempty"`

	// Expected is the value suggested by the current environment.
	Expected string `json:"expected,omitempty"`

	// Role is populated for consilium-role drift items.
	Role string `json:"role,omitempty"`

	// Scope is populated for exec-scope drift items.
	Scope string `json:"scope,omitempty"`

	// Stack is the stack name associated with this item.
	Stack string `json:"stack,omitempty"`

	// AutoFixable indicates whether 'harnest sync' can resolve this automatically.
	AutoFixable bool `json:"auto_fixable"`

	// Message is a short human-readable summary.
	Message string `json:"message"`

	// Hint offers a suggested corrective action.
	Hint string `json:"hint,omitempty"`
}

// DriftResult is the top-level output of a drift check.
type DriftResult struct {
	// ConfigFile is the absolute path of the project config that was analysed.
	ConfigFile string `json:"config_file"`

	// Harness is the detected harness name (e.g. "claude-code", "cursor").
	Harness string `json:"harness"`

	// Items contains all detected drift items, ordered by severity then type.
	Items []DriftItem `json:"items"`
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

// Check analyses the project rooted at dir and returns a DriftResult.
//
// It reads the existing agent config, re-detects stacks from the working tree,
// discovers installed agents, and then compares all three to produce DriftItems.
//
// An error is returned only when the environment is so broken that no
// meaningful analysis is possible (e.g. no config file found).  A non-nil
// result with an empty Items slice means no drift was detected.
func Check(dir string) (*DriftResult, error) {
	// 1. Read existing project config.
	cfg, err := config.ReadProject(dir)
	if err != nil {
		return nil, fmt.Errorf("reading project config: %w", err)
	}

	// Resolve config-file path and harness name from the directory.
	configFile, harness := resolveConfigMeta(dir)

	result := &DriftResult{
		ConfigFile: configFile,
		Harness:    harness,
	}

	// 2. Re-detect stacks from the working tree.
	stacks := detector.Detect(dir)

	// 3. Discover installed agents and build a lookup set.
	discovered := agents_pkg.Discover(dir)
	agentSet := make(map[string]bool, len(discovered))
	for _, a := range discovered {
		agentSet[a] = true
	}

	// 4. Agent-availability checks for consilium roles.
	result.Items = append(result.Items, checkConsiliumAgents(cfg, agentSet)...)

	// 5. Agent-availability checks for exec agents.
	result.Items = append(result.Items, checkExecAgents(cfg, agentSet)...)

	// 6. Scope-file checks: does each exec scope still match any files?
	result.Items = append(result.Items, checkScopeFiles(dir, cfg)...)

	// 7. Stack-coverage checks: does every detected stack have an exec agent?
	result.Items = append(result.Items, checkStackCoverage(stacks, cfg)...)

	return result, nil
}

// ---------------------------------------------------------------------------
// Output helpers
// ---------------------------------------------------------------------------

// ANSI colour codes used in terminal output.
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)

// FormatTerminal returns a human-readable, ANSI-coloured string suitable for
// printing directly to a terminal.
func FormatTerminal(result *DriftResult) string {
	if result == nil || len(result.Items) == 0 {
		return colorBold + "No drift detected." + colorReset + "\n"
	}

	var sb strings.Builder

	fmt.Fprintf(&sb, "%sDrift detected (%d issue(s)):%s\n\n",
		colorBold, len(result.Items), colorReset)

	for _, item := range result.Items {
		// Severity prefix (fixed-width for alignment).
		switch item.Severity {
		case SeverityError:
			fmt.Fprintf(&sb, "  %s%-7s%s", colorRed, "ERROR", colorReset)
		case SeverityWarning:
			fmt.Fprintf(&sb, "  %s%-7s%s", colorYellow, "WARNING", colorReset)
		default:
			fmt.Fprintf(&sb, "  %s%-7s%s", colorCyan, "INFO", colorReset)
		}

		fmt.Fprintf(&sb, " [%s]", item.Type)

		// Type-specific detail lines.
		switch item.Type {
		case DriftAgent:
			if item.Role != "" {
				fmt.Fprintf(&sb, " Role: %s", item.Role)
			} else if item.Scope != "" {
				fmt.Fprintf(&sb, " Exec scope: %s", item.Scope)
			}
			fmt.Fprintln(&sb)
			if item.Current != "" {
				fmt.Fprintf(&sb, "          Config: %s\n", item.Current)
			}
			fmt.Fprintf(&sb, "          Status: Agent not found\n")

		case DriftScope:
			fmt.Fprintf(&sb, " Agent: %s\n", item.Current)
			fmt.Fprintf(&sb, "          Scope:  %s\n", item.Scope)
			fmt.Fprintf(&sb, "          Status: Matches no files\n")

		case DriftStack:
			if item.Stack != "" {
				fmt.Fprintf(&sb, " New stack: %s\n", item.Stack)
			} else {
				fmt.Fprintln(&sb)
			}
			if item.Message != "" {
				fmt.Fprintf(&sb, "          %s\n", item.Message)
			}

		default:
			fmt.Fprintln(&sb)
			if item.Message != "" {
				fmt.Fprintf(&sb, "          %s\n", item.Message)
			}
		}

		if item.Hint != "" {
			fmt.Fprintf(&sb, "          Hint: %s\n", item.Hint)
		}

		fmt.Fprintln(&sb)
	}

	return sb.String()
}

// FormatJSON serialises result as indented JSON.
func FormatJSON(result *DriftResult) ([]byte, error) {
	return json.MarshalIndent(result, "", "  ")
}

// ---------------------------------------------------------------------------
// Internal check functions
// ---------------------------------------------------------------------------

// checkConsiliumAgents verifies that every consilium role's configured agent
// exists among the installed agents.
func checkConsiliumAgents(cfg *config.ProjectConfig, agentSet map[string]bool) []DriftItem {
	var items []DriftItem
	for _, role := range cfg.Consilium {
		if role.Agent == "" {
			continue
		}
		if !agentKnown(role.Agent, agentSet) {
			items = append(items, DriftItem{
				Type:        DriftAgent,
				Severity:    SeverityError,
				Current:     role.Agent,
				Role:        role.Role,
				AutoFixable: false,
				Message: fmt.Sprintf(
					"Consilium role '%s': agent '%s' not found in installed agents",
					role.Role, role.Agent,
				),
				Hint: fmt.Sprintf(
					"Run 'harnest install %s' or update the role with 'harnest set %s <agent>'",
					role.Agent, role.Role,
				),
			})
		}
	}
	return items
}

// checkExecAgents verifies that every exec agent exists among the installed agents.
func checkExecAgents(cfg *config.ProjectConfig, agentSet map[string]bool) []DriftItem {
	var items []DriftItem
	for _, exec := range cfg.Exec {
		if exec.Agent == "" {
			continue
		}
		if !agentKnown(exec.Agent, agentSet) {
			items = append(items, DriftItem{
				Type:        DriftAgent,
				Severity:    SeverityError,
				Current:     exec.Agent,
				Scope:       exec.Scope,
				AutoFixable: false,
				Message: fmt.Sprintf(
					"Exec agent '%s' (scope: %s) not found in installed agents",
					exec.Agent, exec.Scope,
				),
				Hint: fmt.Sprintf(
					"Run 'harnest install %s' to install the missing agent",
					exec.Agent,
				),
			})
		}
	}
	return items
}

// checkScopeFiles verifies that each exec agent's glob scope matches at least
// one file under dir. A scope that matches nothing suggests the project
// structure has changed since the config was generated.
func checkScopeFiles(dir string, cfg *config.ProjectConfig) []DriftItem {
	var items []DriftItem
	for _, exec := range cfg.Exec {
		if exec.Scope == "" {
			continue
		}
		pattern := filepath.Join(dir, exec.Scope)
		matches, err := filepath.Glob(pattern)
		if err != nil {
			// Malformed glob pattern.
			items = append(items, DriftItem{
				Type:        DriftScope,
				Severity:    SeverityError,
				Current:     exec.Agent,
				Scope:       exec.Scope,
				AutoFixable: false,
				Message: fmt.Sprintf(
					"Exec agent '%s': scope pattern '%s' is invalid: %v",
					exec.Agent, exec.Scope, err,
				),
				Hint: "Fix the scope pattern in your config file",
			})
			continue
		}
		if len(matches) == 0 {
			items = append(items, DriftItem{
				Type:        DriftScope,
				Severity:    SeverityError,
				Current:     exec.Agent,
				Scope:       exec.Scope,
				AutoFixable: true,
				Message: fmt.Sprintf(
					"Exec agent '%s': scope '%s' matches no files",
					exec.Agent, exec.Scope,
				),
				Hint: "Run 'harnest drift --fix' to remove the dead scope, or update it manually",
			})
		}
	}
	return items
}

// checkStackCoverage looks for stacks detected in the repository that are not
// covered by any exec agent scope. This surfaces stacks added to the project
// after the last 'harnest init' / 'harnest sync'.
func checkStackCoverage(stacks []detector.Stack, cfg *config.ProjectConfig) []DriftItem { //nolint:unparam
	var items []DriftItem
	for _, stack := range stacks {
		if stackCoveredByExec(stack, cfg.Exec) {
			continue
		}

		// Build a short description for the message.
		stackDesc := fmt.Sprintf("%s (%s)", stack.Name, stack.Lang)
		if stack.Path != "" && stack.Path != "." && stack.Path != "./" {
			stackDesc += fmt.Sprintf(" [%s]", stack.Path)
		}

		items = append(items, DriftItem{
			Type:        DriftStack,
			Severity:    SeverityWarning,
			Stack:       stack.Name,
			Expected:    stack.Path,
			AutoFixable: true,
			Message:     fmt.Sprintf("New stack detected: %s — no exec agent covers it", stackDesc),
			Hint:        "Run 'harnest drift --fix' to add a default exec agent, or update manually",
		})
	}
	return items
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// agentKnown reports whether agentName is present in agentSet.
//
// Matching logic:
//  1. Exact match (e.g. "voltagent-lang:vue-expert" == "voltagent-lang:vue-expert").
//  2. Built-in agent names that are always valid (e.g. "general-purpose").
//  3. Bare name suffix match: "vue-expert" matches "voltagent-lang:vue-expert".
//     This handles configs written before the namespace prefix was introduced.
func agentKnown(agentName string, agentSet map[string]bool) bool {
	// Exact match.
	if agentSet[agentName] {
		return true
	}

	// Well-known harness built-in agents that are never installed on disk.
	switch agentName {
	case "general-purpose":
		return true
	}

	// For namespaced references that weren't found exactly, they are absent.
	if strings.Contains(agentName, ":") {
		return false
	}

	// Bare name: check if any discovered agent ends with ":agentName".
	suffix := ":" + agentName
	for known := range agentSet {
		if strings.HasSuffix(known, suffix) {
			return true
		}
	}
	return false
}

// stackCoveredByExec reports whether at least one exec agent's scope
// encompasses the given stack's path.
//
// The check is path-based (not file-existence-based) — it answers the
// question "does any scope reference a glob that could include files from
// this stack's directory?"
func stackCoveredByExec(stack detector.Stack, execs []mapping.ExecAgent) bool {
	stackDir := strings.TrimSuffix(stack.Path, "/")

	for _, exec := range execs {
		if scopeCoversDir(exec.Scope, stackDir) {
			return true
		}
	}
	return false
}

// scopeCoversDir reports whether a glob scope pattern is intended to cover
// files within stackDir.
func scopeCoversDir(scope, stackDir string) bool {
	if scope == "" {
		return false
	}

	// Normalise root indicators.
	if stackDir == "" || stackDir == "." || stackDir == "./" {
		// Root-level stack: covered if the scope has no directory prefix or
		// its prefix is "." / "**".
		prefix := leadingDir(scope)
		return prefix == "" || prefix == "." || prefix == "**"
	}

	prefix := leadingDir(scope)

	// Wildcard prefix covers all directories.
	if prefix == "**" || prefix == "." || prefix == "" {
		return true
	}

	// Direct match or ancestor match.
	if prefix == stackDir || strings.HasPrefix(stackDir+"/", prefix+"/") {
		return true
	}

	return false
}

// leadingDir returns the first path component of a glob pattern.
//
// Examples:
//
//	"backend/**/*.kt"  → "backend"
//	"**/*.go"          → "**"
//	"src/**/*.ts"      → "src"
func leadingDir(scope string) string {
	parts := strings.SplitN(scope, "/", 2)
	return parts[0]
}

// resolveConfigMeta finds the project config file in dir and infers the
// harness name from the filename.
func resolveConfigMeta(dir string) (configFile, harness string) {
	type candidate struct {
		file    string
		harness string
	}
	candidates := []candidate{
		{"CLAUDE.md", "claude-code"},
		{".cursorrules", "cursor"},
		{".windsurfrules", "windsurf"},
		{"AGENTS.md", "codex"},
		{"QWEN.md", "qwen-code"},
	}
	for _, c := range candidates {
		p := filepath.Join(dir, c.file)
		if fileExistsOnDisk(p) {
			return p, c.harness
		}
	}
	return "", ""
}

// fileExistsOnDisk reports whether path refers to an existing regular file.
func fileExistsOnDisk(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
