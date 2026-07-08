package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	agents_pkg "github.com/AlexGladkov/harnest/internal/agents"
	"github.com/AlexGladkov/harnest/internal/config"
	"github.com/AlexGladkov/harnest/internal/converter"
	"github.com/AlexGladkov/harnest/internal/detector"
	"github.com/AlexGladkov/harnest/internal/drift"
	"github.com/AlexGladkov/harnest/internal/harness"
	"github.com/AlexGladkov/harnest/internal/install"
	"github.com/AlexGladkov/harnest/internal/mapping"
	"github.com/AlexGladkov/harnest/internal/profile"
	"github.com/AlexGladkov/harnest/internal/wizard"
	harnestYaml "github.com/AlexGladkov/harnest/internal/yaml"
	goyaml "gopkg.in/yaml.v3"
)

const version = "0.12.0"

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
	case "drift":
		runDrift()
	case "generate":
		runGenerate()
	case "export":
		runExport()
	case "convert":
		runConvert()
	case "update":
		runUpdate()
	case "local":
		runLocal()
	case "config":
		runConfig()
	case "version", "--version", "-v":
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
	harnessName := parseFlag("--harness", "claude-code")

	fmt.Printf("Installing Harnest framework for %s...\n", harnessName)
	if err := install.InstallAll(harnessName); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	globalDir, _ := harness.GlobalDir(harnessName)
	configPath, _ := harness.GlobalConfigPath(harnessName)

	fmt.Println("\nDone. Installed:")
	fmt.Printf("  - %d workflow profiles → %s/profiles/\n", len(profile.BuiltinNames()), globalDir)
	fmt.Printf("  - global config        → %s\n", configPath)
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
	discovered := agents_pkg.Discover(dir)

	var agentsCfg mapping.AgentConfig
	if nonInteractive {
		agentsCfg = mapping.Resolve(stacks, discovered, harnessName)
	} else {
		structure := mapping.ResolveStructure(stacks)
		suggestions := mapping.GetSuggestions(stacks, discovered, harnessName)
		agentsCfg = wizard.Run(os.Stdin, structure, suggestions, discovered)
	}

	gen, err := harness.Get(harnessName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	outPath, err := gen.Generate(dir, stacks, agentsCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error generating config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nGenerated: %s\n", outPath)
	fmt.Printf("  Consilium roles: %d\n", len(agentsCfg.Consilium))
	fmt.Printf("  Exec agents: %d\n", len(agentsCfg.Exec))
}

func selectHarness() string {
	fmt.Println("\nTarget harness:")
	options := harness.Names()
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

	if input == "" {
		return options[0]
	}
	if idx, err := strconv.Atoi(input); err == nil && idx >= 1 && idx <= len(options) {
		return options[idx-1]
	}
	return input
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
		fmt.Fprintln(os.Stderr, "usage: harnest agents <list|set|set-model> [role] [agent|tier]")
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
			disc := agents_pkg.Discover(dir)
			resolved := mapping.Resolve(stacks, disc, "claude-code")
			printAgentConfig(resolved)
			return
		}
		fmt.Println("Project agent config:")
		fmt.Println("\nConsilium:")
		for _, c := range cfg.Consilium {
			tier := ""
			if cfg.Models != nil {
				if t, ok := cfg.Models[c.Role]; ok {
					tier = fmt.Sprintf(" [%s]", t)
				}
			}
			fmt.Printf("  %-15s → %s%s\n", c.Role, c.Agent, tier)
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

	case "set-model":
		if len(os.Args) < 5 {
			fmt.Fprintln(os.Stderr, "usage: harnest agents set-model <role> <tier>")
			fmt.Fprintln(os.Stderr, "  tier: high, medium, low")
			os.Exit(1)
		}
		role := os.Args[3]
		tier := os.Args[4]
		dir, _ := os.Getwd()
		if d := parseFlag("--dir", ""); d != "" {
			dir = d
		}
		err := config.SetModel(dir, role, tier)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Set model for %s → %s\n", role, tier)

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
	subs := []string{"list", "add", "edit", "remove", "set", "set-model", "unset", "show", "diff"}
	for _, sub := range subs {
		if s == sub {
			return true
		}
	}
	return false
}

// --- drift ---

func runDrift() {
	dir := "."
	if len(os.Args) > 2 && !strings.HasPrefix(os.Args[2], "-") {
		dir = os.Args[2]
	}

	jsonOutput := hasFlag("--json") || hasFlag("--ci")
	ciMode := hasFlag("--ci")

	result, err := drift.Check(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	if jsonOutput {
		data, _ := drift.FormatJSON(result)
		fmt.Println(string(data))
	} else {
		fmt.Print(drift.FormatTerminal(result))
	}

	// --fix: auto-resolve fixable drift items before applying CI exit codes.
	if hasFlag("--fix") {
		fixResult, err := drift.Fix(dir, result)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("\nFixed %d issue(s).\n", len(fixResult.Fixed))
		if len(fixResult.Skipped) > 0 {
			fmt.Printf("Skipped %d issue(s) (require manual decision).\n", len(fixResult.Skipped))
		}
		for _, fixErr := range fixResult.Errors {
			fmt.Fprintf(os.Stderr, "fix error: %v\n", fixErr)
		}
	}

	if ciMode && len(result.Items) > 0 {
		// Determine exit code based on --fail-on level
		failOn := parseFlag("--fail-on", "error")
		for _, item := range result.Items {
			if string(item.Severity) == failOn || (failOn == "warning" && item.Severity == drift.SeverityError) {
				os.Exit(1)
			}
		}
	}
}

// --- generate ---

func runGenerate() {
	dir := "."
	if len(os.Args) > 2 && !strings.HasPrefix(os.Args[2], "-") {
		dir = os.Args[2]
	}

	if !harnestYaml.Exists(dir) {
		fmt.Fprintln(os.Stderr, "error: no harnest.yaml found")
		fmt.Fprintln(os.Stderr, "Run 'harnest init --yaml' to create one, or 'harnest export' to generate from existing config.")
		os.Exit(1)
	}

	cfg, err := harnestYaml.Load(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if hasFlag("--dry-run") {
		contents, err := harnestYaml.GenerateDryRun(dir, cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Would generate:")
		for name, content := range contents {
			lines := strings.Count(content, "\n")
			fmt.Printf("  %s (%d lines)\n", name, lines)
		}
		fmt.Println("\nNo files written. Remove --dry-run to generate.")
		return
	}

	files, err := harnestYaml.Generate(dir, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Generated:")
	for _, f := range files {
		fmt.Printf("  %s\n", f)
	}

	if err := harnestYaml.UpdateGitignore(dir, files); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not update .gitignore: %v\n", err)
	} else {
		fmt.Println("\nUpdated .gitignore")
	}
}

// --- export ---

func runExport() {
	dir := "."
	if len(os.Args) > 2 && !strings.HasPrefix(os.Args[2], "-") {
		dir = os.Args[2]
	}

	if harnestYaml.Exists(dir) {
		fmt.Fprintln(os.Stderr, "error: harnest.yaml already exists in this directory")
		fmt.Fprintln(os.Stderr, "Delete it first if you want to re-export.")
		os.Exit(1)
	}

	// Read existing project config
	projectCfg, err := config.ReadProject(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Detect stacks
	stacks := detector.Detect(dir)

	// Build harnest.yaml config
	cfg := &harnestYaml.HarnestConfig{
		Version:   1,
		Harnesses: []string{"claude-code"}, // default, inferred from found config file
		Agents: harnestYaml.AgentsBlock{
			Consilium: make(map[string]string),
			Models:    make(map[string]string),
		},
		Settings: harnestYaml.SettingsBlock{
			AutoDetect:    true,
			StackStrategy: "merge",
		},
	}

	// Convert stacks
	for _, s := range stacks {
		cfg.Stacks = append(cfg.Stacks, harnestYaml.StackEntry{
			Name:     s.Name,
			Lang:     s.Lang,
			Category: s.Category,
			Path:     s.Path,
		})
	}

	// Convert consilium
	for _, c := range projectCfg.Consilium {
		cfg.Agents.Consilium[c.Role] = c.Agent
	}

	// Convert exec
	for _, e := range projectCfg.Exec {
		cfg.Agents.Executing = append(cfg.Agents.Executing, harnestYaml.ExecEntry{
			Agent: e.Agent,
			Scope: e.Scope,
		})
	}

	// Convert models
	for role, tier := range projectCfg.Models {
		cfg.Agents.Models[role] = tier
	}

	if err := harnestYaml.Save(dir, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Exported to harnest.yaml")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Review harnest.yaml")
	fmt.Println("  2. Run 'harnest generate' to verify output")
	fmt.Println("  3. Add generated config files to .gitignore")
	fmt.Println("  4. Commit harnest.yaml")
}

// --- local ---

func runLocal() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: harnest local <set|unset|show>")
		os.Exit(1)
	}

	dir, _ := os.Getwd()
	if d := parseFlag("--dir", ""); d != "" {
		dir = d
	}

	switch os.Args[2] {
	case "set":
		runLocalSet(dir)
	case "unset":
		runLocalUnset(dir)
	case "show":
		runLocalShow(dir)
	default:
		fmt.Fprintf(os.Stderr, "unknown local subcommand: %s\n", os.Args[2])
		fmt.Fprintln(os.Stderr, "usage: harnest local <set|unset|show>")
		os.Exit(1)
	}
}

// runLocalSet handles: harnest local set <key> <value>
//
// Supported key paths:
//   - agents.consilium.<role>   — override a consilium agent
//   - agents.models.<role>      — override a model tier
//   - harnesses                 — add a harness to the list
//   - design_system             — override the design system
func runLocalSet(dir string) {
	if len(os.Args) < 5 {
		fmt.Fprintln(os.Stderr, "usage: harnest local set <key> <value>")
		fmt.Fprintln(os.Stderr, "  keys: agents.consilium.<role>, agents.models.<role>, harnesses, design_system")
		os.Exit(1)
	}

	key := os.Args[3]
	value := os.Args[4]

	local, err := loadOrNewLocal(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	parts := strings.SplitN(key, ".", 3)

	switch {
	case len(parts) == 3 && parts[0] == "agents" && parts[1] == "consilium":
		role := parts[2]
		if local.Agents.Consilium == nil {
			local.Agents.Consilium = make(map[string]string)
		}
		local.Agents.Consilium[role] = value
		fmt.Printf("Set agents.consilium.%s = %s\n", role, value)

	case len(parts) == 3 && parts[0] == "agents" && parts[1] == "models":
		role := parts[2]
		if local.Agents.Models == nil {
			local.Agents.Models = make(map[string]string)
		}
		local.Agents.Models[role] = value
		fmt.Printf("Set agents.models.%s = %s\n", role, value)

	case key == "harnesses":
		for _, h := range local.Harnesses {
			if h == value {
				fmt.Printf("Harness %q already present in local config.\n", value)
				return
			}
		}
		local.Harnesses = append(local.Harnesses, value)
		fmt.Printf("Added harness: %s\n", value)

	case key == "design_system":
		local.DesignSystem = value
		fmt.Printf("Set design_system = %s\n", value)

	default:
		fmt.Fprintf(os.Stderr, "unknown key %q\n", key)
		fmt.Fprintln(os.Stderr, "  supported: agents.consilium.<role>, agents.models.<role>, harnesses, design_system")
		os.Exit(1)
	}

	if err := harnestYaml.SaveLocal(dir, local); err != nil {
		fmt.Fprintf(os.Stderr, "error saving local config: %v\n", err)
		os.Exit(1)
	}
}

// runLocalUnset handles: harnest local unset <key>
func runLocalUnset(dir string) {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "usage: harnest local unset <key>")
		os.Exit(1)
	}

	key := os.Args[3]

	local, err := loadOrNewLocal(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	parts := strings.SplitN(key, ".", 3)

	switch {
	case len(parts) == 3 && parts[0] == "agents" && parts[1] == "consilium":
		role := parts[2]
		delete(local.Agents.Consilium, role)
		if len(local.Agents.Consilium) == 0 {
			local.Agents.Consilium = nil
		}
		fmt.Printf("Unset agents.consilium.%s\n", role)

	case len(parts) == 3 && parts[0] == "agents" && parts[1] == "models":
		role := parts[2]
		delete(local.Agents.Models, role)
		if len(local.Agents.Models) == 0 {
			local.Agents.Models = nil
		}
		fmt.Printf("Unset agents.models.%s\n", role)

	case key == "harnesses":
		local.Harnesses = nil
		fmt.Println("Cleared harnesses override.")

	case key == "design_system":
		local.DesignSystem = ""
		fmt.Println("Unset design_system.")

	default:
		fmt.Fprintf(os.Stderr, "unknown key %q\n", key)
		os.Exit(1)
	}

	if err := harnestYaml.SaveLocal(dir, local); err != nil {
		fmt.Fprintf(os.Stderr, "error saving local config: %v\n", err)
		os.Exit(1)
	}
}

// runLocalShow handles: harnest local show
func runLocalShow(dir string) {
	if !harnestYaml.LocalExists(dir) {
		fmt.Println("No .harnest-local.yaml found. Nothing to show.")
		return
	}

	local, err := harnestYaml.LoadLocal(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if local == nil {
		fmt.Println("Empty local config.")
		return
	}

	data, err := goyaml.Marshal(local)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshaling: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(string(data))
}

// loadOrNewLocal loads the existing local config or returns a blank one.
func loadOrNewLocal(dir string) (*harnestYaml.LocalConfig, error) {
	if harnestYaml.LocalExists(dir) {
		return harnestYaml.LoadLocal(dir)
	}
	return &harnestYaml.LocalConfig{}, nil
}

// --- config ---

func runConfig() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: harnest config <show|diff> [dir]")
		os.Exit(1)
	}

	switch os.Args[2] {
	case "show":
		runConfigShow()
	case "diff":
		runConfigDiff()
	default:
		fmt.Fprintf(os.Stderr, "unknown config subcommand: %s\n", os.Args[2])
		fmt.Fprintln(os.Stderr, "usage: harnest config <show|diff> [dir]")
		os.Exit(1)
	}
}

// runConfigShow prints the fully merged (team + local) effective configuration.
func runConfigShow() {
	dir := parseDirArg(3)

	if !harnestYaml.Exists(dir) {
		fmt.Fprintln(os.Stderr, "error: no harnest.yaml found")
		os.Exit(1)
	}

	team, err := harnestYaml.Load(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading harnest.yaml: %v\n", err)
		os.Exit(1)
	}

	var effective *harnestYaml.HarnestConfig
	if harnestYaml.LocalExists(dir) {
		local, err := harnestYaml.LoadLocal(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading .harnest-local.yaml: %v\n", err)
			os.Exit(1)
		}
		effective = harnestYaml.Merge(team, local)
		fmt.Println("# Effective config (team + local overrides)")
	} else {
		effective = team
		fmt.Println("# Effective config (team only — no local overrides)")
	}

	data, err := goyaml.Marshal(effective)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshaling: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(string(data))
}

// runConfigDiff prints only the local overrides, showing what differs from the
// team config.
func runConfigDiff() {
	dir := parseDirArg(3)

	if !harnestYaml.LocalExists(dir) {
		fmt.Println("No .harnest-local.yaml found — no local overrides.")
		return
	}

	local, err := harnestYaml.LoadLocal(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading .harnest-local.yaml: %v\n", err)
		os.Exit(1)
	}
	if local == nil {
		fmt.Println("Empty local config — no overrides.")
		return
	}

	fmt.Println("# Local overrides (.harnest-local.yaml)")
	data, err := goyaml.Marshal(local)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshaling: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(string(data))
}

func printUsage() {
	hlist := strings.Join(harness.Names(), "|")
	fmt.Printf(`harnest - AI coding assistant configurator

Usage:
  harnest install [--harness %s]
  harnest init [dir] [--harness %s] [--non-interactive]
  harnest detect [dir]
  harnest profiles list
  harnest profiles add <name>
  harnest profiles edit <name>
  harnest profiles remove <name>
  harnest agents list [dir]
  harnest agents set <role> <agent>
  harnest agents set-model <role> <tier>
  harnest drift [dir] [--json] [--ci] [--fail-on error|warning] [--fix]
  harnest generate [dir] [--dry-run]
  harnest export [dir]
  harnest convert --from <harness> --to <harness> [dir]
  harnest local set <key> <value>
  harnest local unset <key>
  harnest local show
  harnest config show [dir]
  harnest config diff [dir]
  harnest update
  harnest version

Commands:
  install    Install Harnest framework (profiles + global config) for a harness
  init       Detect stack and generate project config with agent wizard
  detect     Show detected stack without generating
  drift      Detect config drift (stale agents, missing scopes, new stacks)
  generate   Generate config files from harnest.yaml
  export     Export existing config to harnest.yaml
  profiles   Manage workflow profiles (create custom, edit, list, remove)
  agents     View/modify agent role mappings
  local      Manage personal config overrides (.harnest-local.yaml)
  config     View effective (merged) configuration
  convert    Convert config between AI assistants
  update     Update agent mappings and profiles

Local key paths (harnest local set/unset):
  agents.consilium.<role>  Override consilium agent for a role
  agents.models.<role>     Override model tier for a role (high|medium|low)
  harnesses                Add a harness to the local list
  design_system            Override the project design system

Flags:
  --harness          Target harness (%s)
  --non-interactive  Use suggested agents without wizard (for CI/scripts)
`, hlist, hlist, hlist)
}
