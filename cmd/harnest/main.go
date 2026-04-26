package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/AlexGladkov/harnest/internal/config"
	"github.com/AlexGladkov/harnest/internal/converter"
	"github.com/AlexGladkov/harnest/internal/detector"
	"github.com/AlexGladkov/harnest/internal/harness"
	"github.com/AlexGladkov/harnest/internal/install"
	"github.com/AlexGladkov/harnest/internal/mapping"
	"github.com/AlexGladkov/harnest/internal/profile"
	"github.com/AlexGladkov/harnest/internal/wizard"
)

const version = "0.3.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "install":
		runInstall()
	case "init":
		runInit()
	case "detect":
		runDetect()
	case "profiles":
		runProfiles()
	case "agents":
		runAgents()
	case "convert":
		runConvert()
	case "update":
		runUpdate()
	case "version":
		fmt.Printf("harnest v%s\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

// --- install ---

func runInstall() {
	fmt.Println("Installing Harnest framework...")
	if err := install.InstallAll(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("\nDone. Installed:")
	fmt.Println("  - 6 workflow profiles → ~/.claude/profiles/")
	fmt.Println("  - global CLAUDE.md    → ~/.claude/CLAUDE.md")
	fmt.Println("\nNext: cd <project> && harnest init")
}

// --- detect ---

func runDetect() {
	dir := parseDirArg(2)
	stacks := detector.Detect(dir)
	if len(stacks) == 0 {
		fmt.Println("No recognized stack detected.")
		return
	}
	fmt.Println("Detected stack:")
	for _, s := range stacks {
		fmt.Printf("  - %s (%s) [%s]\n", s.Name, s.Lang, s.Path)
	}
}

// --- init ---

func runInit() {
	dir := parseDirArg(2)
	harnessName := parseFlag("--harness", "")
	nonInteractive := hasFlag("--non-interactive")

	stacks := detector.Detect(dir)
	if len(stacks) == 0 {
		fmt.Println("No recognized stack detected. Creating minimal config.")
	} else {
		fmt.Println("Detected stack:")
		for _, s := range stacks {
			fmt.Printf("  - %s (%s) [%s]\n", s.Name, s.Lang, s.Path)
		}
	}

	// Harness selection
	if harnessName == "" {
		if nonInteractive {
			harnessName = "claude-code"
		} else {
			harnessName = selectHarness()
		}
	}

	// Agent selection
	var agents mapping.AgentConfig
	if nonInteractive {
		agents = mapping.Resolve(stacks)
	} else {
		structure := mapping.ResolveStructure(stacks)
		suggestions := mapping.GetSuggestions(stacks)
		agents = wizard.Run(os.Stdin, structure, suggestions)
	}

	gen, err := harness.Get(harnessName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	outPath, err := gen.Generate(dir, stacks, agents)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error generating config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nGenerated: %s\n", outPath)
	fmt.Printf("  Consilium roles: %d\n", len(agents.Consilium))
	fmt.Printf("  Exec agents: %d\n", len(agents.Exec))
}

func selectHarness() string {
	fmt.Println("\nTarget harness:")
	options := []string{"claude-code", "cursor", "windsurf"}
	for i, o := range options {
		marker := "  "
		if i == 0 {
			marker = ">"
		}
		fmt.Printf("  %s %d) %s\n", marker, i+1, o)
	}
	fmt.Print("\nSelect [1]: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	switch input {
	case "", "1":
		return "claude-code"
	case "2":
		return "cursor"
	case "3":
		return "windsurf"
	default:
		return input
	}
}

// --- profiles ---

func runProfiles() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: harnest profiles <list|add|edit|remove> [name]")
		os.Exit(1)
	}

	switch os.Args[2] {
	case "list":
		profiles, err := profile.List()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if len(profiles) == 0 {
			fmt.Println("No profiles installed. Run: harnest install")
			return
		}
		fmt.Println("Installed profiles:")
		for _, p := range profiles {
			marker := ""
			if profile.IsBuiltin(p) {
				marker = " (builtin)"
			}
			fmt.Printf("  - %s%s\n", p, marker)
		}

	case "add":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "usage: harnest profiles add <name>")
			os.Exit(1)
		}
		name := os.Args[3]
		reader := bufio.NewReader(os.Stdin)
		if err := profile.Create(name, reader); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "edit":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "usage: harnest profiles edit <name>")
			os.Exit(1)
		}
		name := os.Args[3]
		if err := profile.Edit(name); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "remove":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "usage: harnest profiles remove <name>")
			os.Exit(1)
		}
		name := os.Args[3]
		err := profile.Remove(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Profile '%s' removed.\n", name)

	default:
		fmt.Fprintf(os.Stderr, "unknown profiles subcommand: %s\n", os.Args[2])
		os.Exit(1)
	}
}

// --- agents ---

func runAgents() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: harnest agents <list|set> [role] [agent]")
		os.Exit(1)
	}

	switch os.Args[2] {
	case "list":
		dir := parseDirArg(3)
		cfg, err := config.ReadProject(dir)
		if err != nil {
			// No project config — show what would be generated
			fmt.Println("No project config found. Showing suggestions from detection:")
			stacks := detector.Detect(dir)
			agents := mapping.Resolve(stacks)
			printAgentConfig(agents)
			return
		}
		fmt.Println("Project agent config:")
		fmt.Println("\nConsilium:")
		for _, c := range cfg.Consilium {
			fmt.Printf("  %-15s → %s\n", c.Role, c.Agent)
		}
		fmt.Println("\nExecuting:")
		for _, e := range cfg.Exec {
			fmt.Printf("  %-40s → %s\n", e.Scope, e.Agent)
		}

	case "set":
		if len(os.Args) < 5 {
			fmt.Fprintln(os.Stderr, "usage: harnest agents set <role> <agent>")
			os.Exit(1)
		}
		role := os.Args[3]
		agent := os.Args[4]
		dir, _ := os.Getwd()
		// Optional --dir flag
		if d := parseFlag("--dir", ""); d != "" {
			dir = d
		}
		err := config.SetAgent(dir, role, agent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Set %s → %s\n", role, agent)

	default:
		fmt.Fprintf(os.Stderr, "unknown agents subcommand: %s\n", os.Args[2])
		os.Exit(1)
	}
}

func printAgentConfig(agents mapping.AgentConfig) {
	fmt.Println("\nConsilium:")
	for _, c := range agents.Consilium {
		fmt.Printf("  %-15s → %s\n", c.Role, c.Agent)
	}
	fmt.Println("\nExecuting:")
	for _, e := range agents.Exec {
		fmt.Printf("  %-40s → %s\n", e.Scope, e.Agent)
	}
}

// --- convert ---

func runConvert() {
	from := parseFlag("--from", "")
	to := parseFlag("--to", "")
	dir := parseDirArg(2)

	if from == "" || to == "" {
		fmt.Fprintln(os.Stderr, "usage: harnest convert --from <harness> --to <harness> [dir]")
		os.Exit(1)
	}

	outPath, err := converter.Convert(dir, from, to)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Converted %s → %s: %s\n", from, to, outPath)
}

// --- update ---

func runUpdate() {
	fmt.Println("Checking for updates...")
	err := mapping.Update()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Agent mappings and profiles are up to date.")
}

// --- helpers ---

func parseDirArg(startIdx int) string {
	dir, _ := os.Getwd()
	for i := startIdx; i < len(os.Args); i++ {
		arg := os.Args[i]
		if strings.HasPrefix(arg, "-") {
			// skip flag + its value (unless it's a boolean flag)
			if arg == "--non-interactive" {
				continue
			}
			i++
			continue
		}
		// Check if it's a subcommand keyword, skip those
		if isSubcommand(arg) {
			continue
		}
		dir = arg
		break
	}
	return dir
}

func parseFlag(flag, defaultVal string) string {
	for i, arg := range os.Args {
		if arg == flag && i+1 < len(os.Args) {
			return os.Args[i+1]
		}
	}
	return defaultVal
}

func hasFlag(flag string) bool {
	for _, arg := range os.Args {
		if arg == flag {
			return true
		}
	}
	return false
}

func isSubcommand(s string) bool {
	subs := []string{"list", "add", "edit", "remove", "set"}
	for _, sub := range subs {
		if s == sub {
			return true
		}
	}
	return false
}

func printUsage() {
	fmt.Println(`harnest - AI coding assistant configurator

Usage:
  harnest install
  harnest init [dir] [--harness claude-code|cursor|windsurf] [--non-interactive]
  harnest detect [dir]
  harnest profiles list
  harnest profiles add <name>
  harnest profiles edit <name>
  harnest profiles remove <name>
  harnest agents list [dir]
  harnest agents set <role> <agent>
  harnest convert --from <harness> --to <harness> [dir]
  harnest update
  harnest version

Commands:
  install    Install Harnest framework (profiles + global CLAUDE.md)
  init       Detect stack and generate project config with agent wizard
  detect     Show detected stack without generating
  profiles   Manage workflow profiles (create custom, edit, list, remove)
  agents     View/modify agent role mappings
  convert    Convert config between AI assistants
  update     Update agent mappings and profiles

Flags:
  --harness          Target harness (claude-code, cursor, windsurf)
  --non-interactive  Use suggested agents without wizard (for CI/scripts)`)
}
