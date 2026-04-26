package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/AlexGladkov/harnest/internal/mapping"
)

type ProjectConfig struct {
	Consilium []mapping.ConsiliumRole
	Exec      []mapping.ExecAgent
}

// ReadProject parses ## Agents section from CLAUDE.md
func ReadProject(dir string) (*ProjectConfig, error) {
	path := findConfigFile(dir)
	if path == "" {
		return nil, fmt.Errorf("no config file found in %s", dir)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := string(data)
	cfg := &ProjectConfig{}

	// Parse Consilium table
	consiliumSection := extractSection(content, "Consilium", "Executing")
	if consiliumSection != "" {
		cfg.Consilium = parseConsiliumTable(consiliumSection)
	}

	// Parse Executing table
	execSection := extractSection(content, "Executing", "")
	if execSection != "" {
		cfg.Exec = parseExecTable(execSection)
	}

	if len(cfg.Consilium) == 0 && len(cfg.Exec) == 0 {
		return nil, fmt.Errorf("no agent config found in %s", path)
	}

	return cfg, nil
}

// SetAgent modifies a consilium role in the project config
func SetAgent(dir, role, agent string) error {
	path := findConfigFile(dir)
	if path == "" {
		return fmt.Errorf("no config file found in %s (run 'harnest init' first)", dir)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	content := string(data)

	// Try to find and replace the role in consilium table
	// Pattern: | role | old-agent |
	re := regexp.MustCompile(`(?m)^\|\s*` + regexp.QuoteMeta(role) + `\s*\|[^|]+\|`)
	if re.MatchString(content) {
		content = re.ReplaceAllString(content, fmt.Sprintf("| %s | %s |", role, agent))
		return os.WriteFile(path, []byte(content), 0644)
	}

	return fmt.Errorf("role '%s' not found in config. Available roles are in the Consilium table", role)
}

func findConfigFile(dir string) string {
	candidates := []string{
		"CLAUDE.md",
		".cursorrules",
		".windsurfrules",
	}
	for _, c := range candidates {
		p := filepath.Join(dir, c)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func extractSection(content, startHeader, endHeader string) string {
	// Support both English and Russian headers
	aliases := map[string][]string{
		"Consilium": {"Consilium", "Консилиум"},
		"Executing": {"Executing"},
	}

	startHeaders := aliases[startHeader]
	if startHeaders == nil {
		startHeaders = []string{startHeader}
	}

	startIdx := -1
	headerLen := 0
	for _, h := range startHeaders {
		idx := strings.Index(content, "### "+h)
		if idx != -1 {
			startIdx = idx
			headerLen = len("### " + h)
			break
		}
	}
	if startIdx == -1 {
		return ""
	}
	startIdx += headerLen

	endIdx := len(content)
	if endHeader != "" {
		endHeaders := aliases[endHeader]
		if endHeaders == nil {
			endHeaders = []string{endHeader}
		}
		for _, h := range endHeaders {
			idx := strings.Index(content[startIdx:], "### "+h)
			if idx != -1 {
				endIdx = startIdx + idx
				break
			}
		}
	}

	return content[startIdx:endIdx]
}

func parseConsiliumTable(section string) []mapping.ConsiliumRole {
	var roles []mapping.ConsiliumRole
	for _, line := range strings.Split(section, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") || strings.Contains(line, "---") || strings.Contains(line, "Role") {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 3 {
			role := strings.TrimSpace(parts[1])
			agent := strings.TrimSpace(parts[2])
			if role != "" && agent != "" {
				roles = append(roles, mapping.ConsiliumRole{Role: role, Agent: agent})
			}
		}
	}
	return roles
}

func parseExecTable(section string) []mapping.ExecAgent {
	var agents []mapping.ExecAgent
	for _, line := range strings.Split(section, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") || strings.Contains(line, "---") || strings.Contains(line, "Agent") {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 3 {
			agent := strings.TrimSpace(parts[1])
			scope := strings.TrimSpace(parts[2])
			if agent != "" && scope != "" {
				agents = append(agents, mapping.ExecAgent{Agent: agent, Scope: scope})
			}
		}
	}
	return agents
}
