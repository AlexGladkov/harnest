package profile

import (
	"fmt"
	"os"
	"path/filepath"
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

func profilesDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "profiles")
}

func List() ([]string, error) {
	dir := profilesDir()
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

func Install(name, destDir string) error {
	content, ok := builtinProfiles[name]
	if !ok {
		return fmt.Errorf("unknown profile: %s\nAvailable: %s", name, availableNames())
	}

	dir := profilesDir()
	if destDir != "" {
		dir = destDir
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating profiles dir: %w", err)
	}

	path := filepath.Join(dir, name+".md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing profile: %w", err)
	}

	fmt.Printf("  → %s\n", path)
	return nil
}

func Remove(name string) error {
	path := filepath.Join(profilesDir(), name+".md")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("profile not found: %s", name)
	}
	return os.Remove(path)
}

func availableNames() string {
	names := make([]string, 0, len(builtinProfiles))
	for k := range builtinProfiles {
		names = append(names, k)
	}
	return strings.Join(names, ", ")
}
