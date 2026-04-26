package profile

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var builtinProfiles = map[string]string{
	"business-feature": businessFeature,
	"bug-hunting":      bugHunting,
	"research":         research,
	"refactoring":      refactoring,
	"e2e-testing":      e2eTesting,
	"e2e-authoring":    e2eAuthoring,
}

var validName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$`)

func ValidateName(name string) error {
	if !validName.MatchString(name) {
		return fmt.Errorf("invalid profile name %q: must match [a-zA-Z0-9][a-zA-Z0-9_-]{0,63}", name)
	}
	return nil
}

func profilesDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".claude", "profiles"), nil
}

func safePath(name string) (string, error) {
	if err := ValidateName(name); err != nil {
		return "", err
	}
	dir, err := profilesDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, name+".md")
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(absPath, absDir+string(filepath.Separator)) {
		return "", fmt.Errorf("path traversal detected: %s", name)
	}
	return absPath, nil
}

// BuiltinNames returns sorted list of builtin profile names.
func BuiltinNames() []string {
	names := make([]string, 0, len(builtinProfiles))
	for k := range builtinProfiles {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// IsBuiltin checks if name is a builtin profile.
func IsBuiltin(name string) bool {
	_, ok := builtinProfiles[name]
	return ok
}

// BuiltinContent returns builtin profile content.
func BuiltinContent(name string) (string, bool) {
	content, ok := builtinProfiles[name]
	return content, ok
}

func List() ([]string, error) {
	dir, err := profilesDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			names = append(names, strings.TrimSuffix(e.Name(), ".md"))
		}
	}
	return names, nil
}

// Install writes a builtin profile to disk.
func Install(name string) error {
	content, ok := builtinProfiles[name]
	if !ok {
		return fmt.Errorf("unknown builtin profile: %s", name)
	}

	dir, err := profilesDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating profiles dir: %w", err)
	}

	path := filepath.Join(dir, name+".md")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0600); err != nil {
		return fmt.Errorf("writing profile: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("renaming profile: %w", err)
	}

	fmt.Printf("  → %s\n", path)
	return nil
}

// IsModified checks if an installed builtin profile differs from its template.
func IsModified(name string) (bool, error) {
	builtin, ok := builtinProfiles[name]
	if !ok {
		return false, nil
	}
	path, err := safePath(name)
	if err != nil {
		return false, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return string(data) != builtin, nil
}

// Remove deletes a profile from disk.
func Remove(name string) error {
	path, err := safePath(name)
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("profile not found: %s", name)
	}
	return os.Remove(path)
}

// Edit opens a profile in $EDITOR.
func Edit(name string) error {
	path, err := safePath(name)
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("profile not found: %s", name)
	}
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	parts := strings.Fields(editor)
	bin := parts[0]
	args := append(parts[1:], path)

	cmd := exec.Command(bin, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Create runs an interactive wizard to create a custom profile.
func Create(name string, r *bufio.Reader) error {
	if err := ValidateName(name); err != nil {
		return err
	}
	path, err := safePath(name)
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("profile already exists: %s\nUse 'harnest profiles edit %s' to modify", name, name)
	}

	fmt.Printf("\nCreating profile: %s\n", name)

	allRoles := []string{"architect", "frontend", "ui", "security", "devops", "api", "diagnostics", "test"}

	var stages []stage

	for i := 1; ; i++ {
		fmt.Printf("\n--- Stage %d ---\n", i)

		sName := prompt(r, "Stage name")
		if sName == "" {
			fmt.Println("Stage name is required.")
			continue
		}

		agentType := promptChoice(r, "Agent type", []string{"single", "consilium", "bash", "none"})

		s := stage{Name: sName, AgentType: agentType}

		switch agentType {
		case "consilium":
			fmt.Printf("Available roles: %s\n", strings.Join(allRoles, ", "))
			rolesStr := prompt(r, "Roles (comma-separated)")
			for _, role := range strings.Split(rolesStr, ",") {
				role = strings.TrimSpace(role)
				if role != "" {
					s.Roles = append(s.Roles, role)
				}
			}
		case "single":
			fmt.Printf("Available roles: %s\n", strings.Join(allRoles, ", "))
			s.Role = prompt(r, "Role (or Enter for 'general-purpose')")
			if s.Role == "" {
				s.Role = "general-purpose"
			}
		}

		if len(stages) > 0 {
			existing := make([]string, len(stages))
			for j, st := range stages {
				existing[j] = st.Name
			}
			fmt.Printf("Existing stages: %s\n", strings.Join(existing, ", "))
		}
		transStr := prompt(r, fmt.Sprintf("After %s, allowed transitions (comma-separated stage names)", sName))
		for _, t := range strings.Split(transStr, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				s.Transitions = append(s.Transitions, t)
			}
		}

		stages = append(stages, s)

		more := prompt(r, "Add another stage? (y/n)")
		if strings.ToLower(more) != "y" {
			break
		}
	}

	keywords := prompt(r, "\nKeywords (comma-separated)")
	description := prompt(r, "Description")

	content := renderProfile(name, keywords, description, stages)

	dir, err := profilesDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating profiles dir: %w", err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0600); err != nil {
		return fmt.Errorf("writing profile: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("renaming profile: %w", err)
	}

	fmt.Printf("\nProfile '%s' created.\n", name)
	fmt.Printf("  → %s\n", path)
	return nil
}

type stage struct {
	Name        string
	AgentType   string
	Role        string   // for single
	Roles       []string // for consilium
	Transitions []string
}

func renderProfile(name, keywords, description string, stages []stage) string {
	title := strings.ReplaceAll(name, "-", " ")
	title = strings.Title(title)

	var b strings.Builder
	b.WriteString(fmt.Sprintf("# Profile: %s\n\n", title))

	b.WriteString("## Meta\n")
	b.WriteString(fmt.Sprintf("- **Keywords:** %s\n", keywords))
	b.WriteString(fmt.Sprintf("- **Description:** %s\n", description))

	b.WriteString("\n## Workflow (STRICT)\n\n")
	b.WriteString("### Stages\n")
	for i, s := range stages {
		desc := stageDescription(s)
		b.WriteString(fmt.Sprintf("%d. **%s** — %s\n", i+1, s.Name, desc))
	}

	b.WriteString("\n### Allowed transitions\n```\n")
	for _, s := range stages {
		for _, t := range s.Transitions {
			b.WriteString(fmt.Sprintf("%-15s -> %s\n", s.Name, t))
		}
	}
	b.WriteString("```\n")

	b.WriteString("\n### Agents per stage\n\n")
	b.WriteString("| Stage | Agents | Model |\n")
	b.WriteString("|-------|--------|-------|\n")
	for _, s := range stages {
		agents, model := stageAgentInfo(s)
		b.WriteString(fmt.Sprintf("| %s | %s | %s |\n", s.Name, agents, model))
	}

	for _, s := range stages {
		if s.AgentType == "consilium" && len(s.Roles) > 0 {
			b.WriteString(fmt.Sprintf("\n### %s — Agent consilium\n\n", s.Name))
			b.WriteString("| Role | Responsibility |\n")
			b.WriteString("|------|----------------|\n")
			for _, role := range s.Roles {
				b.WriteString(fmt.Sprintf("| `%s` | |\n", role))
			}
		}
	}

	return b.String()
}

func stageDescription(s stage) string {
	switch s.AgentType {
	case "consilium":
		return "consilium analyzes task"
	case "bash":
		return "bash execution"
	case "single":
		return s.Role
	case "none":
		return "terminal stage"
	default:
		return s.AgentType
	}
}

func stageAgentInfo(s stage) (string, string) {
	switch s.AgentType {
	case "consilium":
		return "CONSILIUM (see below)", "opus"
	case "bash":
		return "Bash", "sonnet"
	case "single":
		return s.Role, "opus"
	case "none":
		return "—", "—"
	default:
		return s.AgentType, "opus"
	}
}

func prompt(r *bufio.Reader, label string) string {
	fmt.Printf("%s: ", label)
	input, _ := r.ReadString('\n')
	return strings.TrimSpace(input)
}

func promptChoice(r *bufio.Reader, label string, options []string) string {
	for {
		fmt.Printf("%s (%s): ", label, strings.Join(options, "/"))
		input, _ := r.ReadString('\n')
		input = strings.TrimSpace(input)
		for _, opt := range options {
			if input == opt {
				return input
			}
		}
		fmt.Printf("Invalid choice. Options: %s\n", strings.Join(options, ", "))
	}
}
