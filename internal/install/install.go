package install

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlexGladkov/harnest/internal/profile"
)

var allProfiles = []string{
	"business-feature",
	"bug-hunting",
	"research",
	"refactoring",
	"e2e-testing",
	"e2e-authoring",
}

func InstallAll() error {
	// Install profiles
	fmt.Println("Installing profiles...")
	for _, name := range allProfiles {
		if err := profile.Install(name, ""); err != nil {
			return fmt.Errorf("installing profile %s: %w", name, err)
		}
	}

	// Install global CLAUDE.md
	fmt.Println("\nInstalling global CLAUDE.md...")
	if err := installGlobalConfig(); err != nil {
		return fmt.Errorf("installing global config: %w", err)
	}

	return nil
}

func installGlobalConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	path := filepath.Join(dir, "CLAUDE.md")
	managed := managedStart + "\n" + globalTemplate + "\n" + managedEnd

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist — create with managed block
			if err := os.WriteFile(path, []byte(managed+"\n"), 0644); err != nil {
				return err
			}
			fmt.Printf("  → created %s\n", path)
			return nil
		}
		return err
	}

	content := string(data)

	// File exists — check for markers
	startIdx := strings.Index(content, managedStart)
	endIdx := strings.Index(content, managedEnd)

	if startIdx != -1 && endIdx != -1 {
		// Replace managed block
		content = content[:startIdx] + managed + content[endIdx+len(managedEnd):]
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return err
		}
		fmt.Printf("  → updated managed block in %s\n", path)
		return nil
	}

	// No markers — append managed block at end
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	content += "\n" + managed + "\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return err
	}
	fmt.Printf("  → appended managed block to %s\n", path)
	return nil
}
