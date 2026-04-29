package install

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AlexGladkov/harnest/internal/harness"
	"github.com/AlexGladkov/harnest/internal/profile"
)

// InstallAll installs profiles and global config for the given harness.
// For "claude-code" it uses the default ~/.claude/ paths (backwards compat).
// For other harnesses it derives the target dir from harness.GlobalDir().
func InstallAll(harnessName string) error {
	reader := bufio.NewReader(os.Stdin)

	globalDir, err := harness.GlobalDir(harnessName)
	if err != nil {
		return err
	}

	isDefault := harnessName == "claude-code"

	fmt.Println("Installing profiles...")
	for _, name := range profile.BuiltinNames() {
		var modified bool
		if isDefault {
			modified, err = profile.IsModified(name)
		} else {
			modified, err = profile.IsModifiedIn(name, globalDir)
		}
		if err != nil {
			return fmt.Errorf("checking profile %s: %w", name, err)
		}
		if modified {
			action := promptConflict(reader, name, globalDir)
			if action == "skip" {
				fmt.Printf("  → skipped %s.md\n", name)
				continue
			}
		}
		if isDefault {
			err = profile.Install(name)
		} else {
			err = profile.InstallTo(name, globalDir)
		}
		if err != nil {
			return fmt.Errorf("installing profile %s: %w", name, err)
		}
	}

	configPath, err := harness.GlobalConfigPath(harnessName)
	if err != nil {
		return err
	}

	fmt.Printf("\nInstalling global config → %s ...\n", configPath)
	if err := installGlobalConfig(globalDir, configPath); err != nil {
		return fmt.Errorf("installing global config: %w", err)
	}

	return nil
}

func promptConflict(reader *bufio.Reader, name, baseDir string) string {
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
			showDiff(name, baseDir)
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

func showDiff(name, baseDir string) {
	builtin, ok := profile.BuiltinContent(name)
	if !ok {
		return
	}

	existingPath := filepath.Join(baseDir, "profiles", name+".md")

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
	cmd.Run()
	fmt.Println()
}

func installGlobalConfig(dir, path string) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

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
