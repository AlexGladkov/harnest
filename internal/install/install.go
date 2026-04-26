package install

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AlexGladkov/harnest/internal/profile"
)

func InstallAll() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Installing profiles...")
	for _, name := range profile.BuiltinNames() {
		modified, err := profile.IsModified(name)
		if err != nil {
			return fmt.Errorf("checking profile %s: %w", name, err)
		}
		if modified {
			action := promptConflict(reader, name)
			if action == "skip" {
				fmt.Printf("  → skipped %s.md\n", name)
				continue
			}
		}
		if err := profile.Install(name); err != nil {
			return fmt.Errorf("installing profile %s: %w", name, err)
		}
	}

	fmt.Println("\nInstalling global CLAUDE.md...")
	if err := installGlobalConfig(); err != nil {
		return fmt.Errorf("installing global config: %w", err)
	}

	return nil
}

func promptConflict(reader *bufio.Reader, name string) string {
	for {
		fmt.Printf("  → %s.md (modified locally)\n", name)
		fmt.Print("    [o]verwrite  [s]kip  [d]iff: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "o", "overwrite":
			return "overwrite"
		case "s", "skip":
			return "skip"
		case "d", "diff":
			showDiff(name)
			// After diff, ask again without diff option
			for {
				fmt.Print("    [o]verwrite  [s]kip: ")
				input2, _ := reader.ReadString('\n')
				input2 = strings.TrimSpace(strings.ToLower(input2))
				switch input2 {
				case "o", "overwrite":
					return "overwrite"
				case "s", "skip":
					return "skip"
				default:
					fmt.Println("    Invalid choice.")
				}
			}
		default:
			fmt.Println("    Invalid choice.")
		}
	}
}

func showDiff(name string) {
	builtin, ok := profile.BuiltinContent(name)
	if !ok {
		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("    Cannot determine home directory")
		return
	}

	existingPath := filepath.Join(home, ".claude", "profiles", name+".md")

	tmp, err := os.CreateTemp("", "harnest-builtin-*.md")
	if err != nil {
		fmt.Printf("    Cannot create temp file: %v\n", err)
		return
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.WriteString(builtin); err != nil {
		tmp.Close()
		return
	}
	tmp.Close()

	cmd := exec.Command("diff", "-u", "--label", "builtin/"+name+".md", "--label", "installed/"+name+".md", tmp.Name(), existingPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run() // diff returns exit code 1 when files differ — ignore
	fmt.Println()
}

func installGlobalConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	path := filepath.Join(dir, "CLAUDE.md")
	managed := managedStart + "\n" + globalTemplate + "\n" + managedEnd

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.WriteFile(path, []byte(managed+"\n"), 0600); err != nil {
				return err
			}
			fmt.Printf("  → created %s\n", path)
			return nil
		}
		return err
	}

	content := string(data)

	startIdx := strings.Index(content, managedStart)
	endIdx := strings.Index(content, managedEnd)

	if startIdx != -1 && endIdx != -1 {
		content = content[:startIdx] + managed + content[endIdx+len(managedEnd):]
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			return err
		}
		fmt.Printf("  → updated managed block in %s\n", path)
		return nil
	}

	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	content += "\n" + managed + "\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return err
	}
	fmt.Printf("  → appended managed block to %s\n", path)
	return nil
}
