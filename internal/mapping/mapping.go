package mapping

import (
	"strings"

	"github.com/AlexGladkov/harnest/internal/detector"
)

type AgentConfig struct {
	Consilium []ConsiliumRole
	Exec      []ExecAgent
	Models    map[string]string // role → tier (high/medium/low)
}

// defaultModelTiers defines the default capability tier per consilium role.
var defaultModelTiers = map[string]string{
	"architect":   "high",
	"security":    "high",
	"api":         "medium",
	"frontend":    "medium",
	"ui":          "medium",
	"devops":      "medium",
	"diagnostics": "medium",
	"test":        "medium",
	"mobile":      "medium",
}

// DefaultModelTiers returns a copy of the default role→tier mapping.
func DefaultModelTiers() map[string]string {
	m := make(map[string]string, len(defaultModelTiers))
	for k, v := range defaultModelTiers {
		m[k] = v
	}
	return m
}

type ConsiliumRole struct {
	Role  string
	Agent string
}

type ExecAgent struct {
	Agent string
	Scope string
}

// ExecKeywords defines keywords to match an agent + its file scope.
type ExecKeywords struct {
	Keywords []string
	Scope    string
}

// --- Keyword maps: lang -> search keywords for matching discovered agents ---

var architectKeywords = map[string][]string{
	"kotlin":     {"java", "architect"},
	"java":       {"java", "architect"},
	"scala":      {"java", "architect"},
	"groovy":     {"java", "architect"},
	"clojure":    {"java", "architect"},
	"swift":      {"swift"},
	"python":     {"python"},
	"typescript": {"typescript"},
	"go":         {"golang", "go"},
	"rust":       {"rust"},
	"dart":       {"flutter", "dart"},
	"ruby":       {"ruby", "rails"},
	"php":        {"php"},
	"csharp":     {"dotnet", "csharp"},
	"elixir":     {"elixir"},
	"erlang":     {"elixir", "erlang"},
	"gleam":      {"elixir", "gleam"},
	"haskell":    {"elixir", "haskell"},
	"ocaml":      {"elixir", "ocaml"},
	"c":          {"cpp", "c"},
	"cpp":        {"cpp"},
	"zig":        {"cpp", "zig"},
	"nim":        {"python", "nim"},
	"vlang":      {"golang", "go"},
	"crystal":    {"ruby", "crystal"},
	"julia":      {"python", "julia"},
	"r":          {"python", "r"},
	"lua":        {"javascript", "lua"},
	"perl":       {"python", "perl"},
	"hcl":        {"terraform"},
	"yaml":       {"devops"},
}

var frontendKeywords = map[string][]string{
	"vue":       {"vue"},
	"nuxt":      {"vue", "nuxt"},
	"react":     {"react"},
	"nextjs":    {"nextjs", "next"},
	"gatsby":    {"react", "gatsby"},
	"remix":     {"react", "remix"},
	"angular":   {"angular"},
	"svelte":    {"javascript", "svelte"},
	"sveltekit": {"javascript", "svelte"},
	"solid":     {"javascript", "solid"},
	"qwik":      {"javascript", "qwik"},
	"astro":     {"javascript", "astro"},
	"ember":     {"javascript", "ember"},
	"eleventy":  {"javascript", "eleventy"},
	"hugo":      {"golang", "hugo"},
	"jekyll":    {"rails", "jekyll"},
	"flutter":   {"flutter"},
	"swiftui":   {"swift"},
}

var mobileKeywords = map[string][]string{
	"kotlin": {"kotlin", "multiplatform"},
	"swift":  {"swift"},
	"dart":   {"flutter", "dart"},
}

var securityKeywords = map[string][]string{
	"kotlin": {"security", "kotlin"},
}

var diagnosticsKeywords = map[string][]string{
	"kotlin": {"diagnostics", "kotlin"},
}

var testKeywords = map[string][]string{
	"kotlin": {"test", "spring"},
}

// defaultRoleKeywords — fallback keywords per role (used when lang not in map)
var defaultRoleKeywords = map[string][]string{
	"architect":   {"architect"},
	"frontend":    {"frontend", "vue"},
	"ui":          {"ui", "designer"},
	"security":    {"security"},
	"devops":      {"devops"},
	"api":         {"api", "designer"},
	"diagnostics": {"diagnostics"},
	"test":        {"test"},
	"mobile":      {"mobile", "multiplatform"},
}

// execKeywords — stack name -> keywords + scope
var execKeywords = map[string]ExecKeywords{
	// --- Kotlin ---
	"spring-boot":           {Keywords: []string{"spring", "feature", "builder"}, Scope: "backend/**/*.kt"},
	"ktor":                  {Keywords: []string{"kotlin", "ktor"}, Scope: "backend/**/*.kt"},
	"quarkus":               {Keywords: []string{"spring", "boot", "kotlin"}, Scope: "src/**/*.kt"},
	"micronaut":             {Keywords: []string{"spring", "boot", "kotlin"}, Scope: "src/**/*.kt"},
	"compose-multiplatform": {Keywords: []string{"kotlin", "multiplatform", "compose"}, Scope: "composeApp/**/*.kt"},
	"android":               {Keywords: []string{"kotlin", "multiplatform", "android"}, Scope: "app/**/*.kt"},

	// --- Java ---
	"spring-boot-java": {Keywords: []string{"spring", "boot", "java"}, Scope: "src/**/*.java"},
	"java":             {Keywords: []string{"java", "architect"}, Scope: "src/**/*.java"},

	// --- Swift ---
	"ios-native":    {Keywords: []string{"swift"}, Scope: "iosApp/**/*.swift"},
	"swift-package": {Keywords: []string{"swift"}, Scope: "**/*.swift"},
	"vapor":         {Keywords: []string{"swift", "vapor"}, Scope: "Sources/**/*.swift"},

	// --- JS/TS Frontend ---
	"vue":       {Keywords: []string{"vue"}, Scope: "src/**/*.vue"},
	"nuxt":      {Keywords: []string{"vue", "nuxt"}, Scope: "src/**/*.vue"},
	"react":     {Keywords: []string{"react"}, Scope: "src/**/*.tsx"},
	"nextjs":    {Keywords: []string{"nextjs", "next"}, Scope: "src/**/*.tsx"},
	"gatsby":    {Keywords: []string{"react", "gatsby"}, Scope: "src/**/*.tsx"},
	"remix":     {Keywords: []string{"react", "remix"}, Scope: "app/**/*.tsx"},
	"angular":   {Keywords: []string{"angular"}, Scope: "src/**/*.ts"},
	"svelte":    {Keywords: []string{"javascript", "svelte"}, Scope: "src/**/*.svelte"},
	"sveltekit": {Keywords: []string{"javascript", "svelte"}, Scope: "src/**/*.svelte"},
	"solid":     {Keywords: []string{"javascript", "solid"}, Scope: "src/**/*.tsx"},
	"qwik":      {Keywords: []string{"javascript", "qwik"}, Scope: "src/**/*.tsx"},
	"astro":     {Keywords: []string{"javascript", "astro"}, Scope: "src/**/*.astro"},
	"ember":     {Keywords: []string{"javascript", "ember"}, Scope: "app/**/*.js"},
	"eleventy":  {Keywords: []string{"javascript", "eleventy"}, Scope: "src/**/*.njk"},

	// --- JS/TS Backend ---
	"node":   {Keywords: []string{"node"}, Scope: "src/**/*.ts"},
	"deno":   {Keywords: []string{"typescript", "deno"}, Scope: "**/*.ts"},
	"bun":    {Keywords: []string{"typescript", "bun"}, Scope: "**/*.ts"},
	"strapi": {Keywords: []string{"node", "strapi"}, Scope: "src/**/*.js"},

	// --- JS/TS Mobile ---
	"expo":         {Keywords: []string{"expo", "react-native"}, Scope: "src/**/*.tsx"},
	"react-native": {Keywords: []string{"expo", "react-native"}, Scope: "src/**/*.tsx"},
	"ionic":        {Keywords: []string{"react", "ionic"}, Scope: "src/**/*.tsx"},
	"capacitor":    {Keywords: []string{"react", "capacitor"}, Scope: "src/**/*.tsx"},

	// --- JS/TS Desktop ---
	"electron": {Keywords: []string{"electron"}, Scope: "src/**/*.ts"},
	"tauri":    {Keywords: []string{"rust", "tauri"}, Scope: "src-tauri/**/*.rs"},

	// --- Python ---
	"fastapi":   {Keywords: []string{"fastapi"}, Scope: "**/*.py"},
	"django":    {Keywords: []string{"django"}, Scope: "**/*.py"},
	"flask":     {Keywords: []string{"python", "flask"}, Scope: "**/*.py"},
	"starlette": {Keywords: []string{"fastapi", "starlette"}, Scope: "**/*.py"},
	"pyramid":   {Keywords: []string{"python", "pyramid"}, Scope: "**/*.py"},
	"litestar":  {Keywords: []string{"python", "litestar"}, Scope: "**/*.py"},
	"streamlit": {Keywords: []string{"python", "streamlit"}, Scope: "**/*.py"},
	"gradio":    {Keywords: []string{"python", "gradio"}, Scope: "**/*.py"},
	"jupyter":   {Keywords: []string{"python", "jupyter"}, Scope: "**/*.ipynb"},

	// --- Go ---
	"go":      {Keywords: []string{"golang", "go"}, Scope: "**/*.go"},
	"gin":     {Keywords: []string{"golang", "gin"}, Scope: "**/*.go"},
	"fiber":   {Keywords: []string{"golang", "fiber"}, Scope: "**/*.go"},
	"echo":    {Keywords: []string{"golang", "echo"}, Scope: "**/*.go"},
	"chi":     {Keywords: []string{"golang", "chi"}, Scope: "**/*.go"},
	"buffalo": {Keywords: []string{"golang", "buffalo"}, Scope: "**/*.go"},

	// --- Rust ---
	"rust":   {Keywords: []string{"rust"}, Scope: "src/**/*.rs"},
	"axum":   {Keywords: []string{"rust", "axum"}, Scope: "src/**/*.rs"},
	"actix":  {Keywords: []string{"rust", "actix"}, Scope: "src/**/*.rs"},
	"rocket": {Keywords: []string{"rust", "rocket"}, Scope: "src/**/*.rs"},
	"warp":   {Keywords: []string{"rust", "warp"}, Scope: "src/**/*.rs"},

	// --- Dart ---
	"flutter": {Keywords: []string{"flutter"}, Scope: "lib/**/*.dart"},

	// --- Ruby ---
	"rails":   {Keywords: []string{"rails"}, Scope: "app/**/*.rb"},
	"sinatra": {Keywords: []string{"rails", "ruby", "sinatra"}, Scope: "**/*.rb"},
	"jekyll":  {Keywords: []string{"rails", "ruby", "jekyll"}, Scope: "**/*.rb"},

	// --- PHP ---
	"laravel":   {Keywords: []string{"laravel"}, Scope: "app/**/*.php"},
	"symfony":   {Keywords: []string{"symfony"}, Scope: "src/**/*.php"},
	"wordpress": {Keywords: []string{"php", "wordpress"}, Scope: "**/*.php"},

	// --- C# / .NET ---
	"dotnet": {Keywords: []string{"dotnet"}, Scope: "**/*.cs"},
	"maui":   {Keywords: []string{"dotnet", "maui"}, Scope: "**/*.cs"},

	// --- Elixir / Erlang / BEAM ---
	"phoenix": {Keywords: []string{"elixir", "phoenix"}, Scope: "lib/**/*.ex"},
	"elixir":  {Keywords: []string{"elixir"}, Scope: "lib/**/*.ex"},
	"erlang":  {Keywords: []string{"elixir", "erlang"}, Scope: "src/**/*.erl"},
	"gleam":   {Keywords: []string{"elixir", "gleam"}, Scope: "src/**/*.gleam"},

	// --- JVM (non-Java/Kotlin) ---
	"scala":   {Keywords: []string{"java", "scala"}, Scope: "src/**/*.scala"},
	"play":    {Keywords: []string{"java", "scala", "play"}, Scope: "app/**/*.scala"},
	"akka":    {Keywords: []string{"java", "scala", "akka"}, Scope: "src/**/*.scala"},
	"clojure": {Keywords: []string{"java", "clojure"}, Scope: "src/**/*.clj"},
	"grails":  {Keywords: []string{"java", "groovy", "grails"}, Scope: "grails-app/**/*.groovy"},

	// --- C / C++ ---
	"c":   {Keywords: []string{"cpp", "c"}, Scope: "src/**/*.c"},
	"cpp": {Keywords: []string{"cpp"}, Scope: "src/**/*.cpp"},

	// --- Systems / Emerging ---
	"zig":     {Keywords: []string{"cpp", "zig"}, Scope: "src/**/*.zig"},
	"nim":     {Keywords: []string{"python", "nim"}, Scope: "src/**/*.nim"},
	"vlang":   {Keywords: []string{"golang", "vlang"}, Scope: "src/**/*.v"},
	"crystal": {Keywords: []string{"rails", "ruby", "crystal"}, Scope: "src/**/*.cr"},

	// --- Functional ---
	"haskell": {Keywords: []string{"elixir", "haskell"}, Scope: "src/**/*.hs"},
	"ocaml":   {Keywords: []string{"elixir", "ocaml"}, Scope: "lib/**/*.ml"},

	// --- Scientific / Data ---
	"julia": {Keywords: []string{"python", "julia"}, Scope: "src/**/*.jl"},
	"r":     {Keywords: []string{"python", "r"}, Scope: "R/**/*.R"},

	// --- Scripting ---
	"lua":  {Keywords: []string{"javascript", "lua"}, Scope: "**/*.lua"},
	"perl": {Keywords: []string{"python", "perl"}, Scope: "lib/**/*.pm"},

	// --- Static Site Generators ---
	"hugo": {Keywords: []string{"golang", "hugo"}, Scope: "content/**/*.md"},

	// --- Infra ---
	"docker":         {Keywords: []string{"docker"}, Scope: "**/Dockerfile"},
	"terraform":      {Keywords: []string{"terraform"}, Scope: "**/*.tf"},
	"helm":           {Keywords: []string{"kubernetes", "helm"}, Scope: "**/*.yaml"},
	"pulumi":         {Keywords: []string{"cloud", "architect", "pulumi"}, Scope: "**/*"},
	"ansible":        {Keywords: []string{"devops", "ansible"}, Scope: "**/*.yml"},
	"github-actions": {Keywords: []string{"deployment", "github"}, Scope: ".github/workflows/**/*.yml"},
}

// MatchAgent finds best agent from discovered list by keyword scoring.
// Returns empty string if no match found.
func MatchAgent(discovered []string, keywords []string) string {
	if len(discovered) == 0 || len(keywords) == 0 {
		return ""
	}

	bestAgent := ""
	bestScore := 0

	for _, agent := range discovered {
		lower := strings.ToLower(agent)
		score := 0
		for _, kw := range keywords {
			if strings.Contains(lower, strings.ToLower(kw)) {
				score++
			}
		}
		if score > bestScore {
			bestScore = score
			bestAgent = agent
		}
	}

	return bestAgent
}

// matchWithFallback matches agent from discovered list, with harness-specific fallback.
// For claude-code: fallback to "general-purpose" (built-in).
// For other harnesses: no fallback (empty string).
func matchWithFallback(discovered []string, keywords []string, harnessName string) string {
	agent := MatchAgent(discovered, keywords)
	if agent != "" {
		return agent
	}
	if harnessName == "claude-code" {
		return "general-purpose"
	}
	return ""
}

// matchRole resolves a consilium role agent from discovered agents.
func matchRole(discovered []string, langMap map[string][]string, langKey string, defaultKW []string, harnessName string) string {
	kw := defaultKW
	if langKW, ok := langMap[langKey]; ok {
		kw = langKW
	}
	return matchWithFallback(discovered, kw, harnessName)
}

func Resolve(stacks []detector.Stack, discovered []string, harnessName string) AgentConfig {
	config := AgentConfig{
		Models: DefaultModelTiers(),
	}

	primaryLang, frontendName := extractLangAndFrontend(stacks)

	// Consilium roles
	config.Consilium = append(config.Consilium, ConsiliumRole{
		Role:  "architect",
		Agent: matchRole(discovered, architectKeywords, primaryLang, defaultRoleKeywords["architect"], harnessName),
	})
	config.Consilium = append(config.Consilium, ConsiliumRole{
		Role:  "frontend",
		Agent: matchRole(discovered, frontendKeywords, frontendName, defaultRoleKeywords["frontend"], harnessName),
	})
	config.Consilium = append(config.Consilium, ConsiliumRole{
		Role:  "ui",
		Agent: matchWithFallback(discovered, defaultRoleKeywords["ui"], harnessName),
	})
	config.Consilium = append(config.Consilium, ConsiliumRole{
		Role:  "security",
		Agent: matchRole(discovered, securityKeywords, primaryLang, defaultRoleKeywords["security"], harnessName),
	})
	config.Consilium = append(config.Consilium, ConsiliumRole{
		Role:  "devops",
		Agent: matchWithFallback(discovered, defaultRoleKeywords["devops"], harnessName),
	})
	config.Consilium = append(config.Consilium, ConsiliumRole{
		Role:  "api",
		Agent: matchWithFallback(discovered, defaultRoleKeywords["api"], harnessName),
	})
	config.Consilium = append(config.Consilium, ConsiliumRole{
		Role:  "diagnostics",
		Agent: matchRole(discovered, diagnosticsKeywords, primaryLang, defaultRoleKeywords["diagnostics"], harnessName),
	})
	config.Consilium = append(config.Consilium, ConsiliumRole{
		Role:  "test",
		Agent: matchRole(discovered, testKeywords, primaryLang, defaultRoleKeywords["test"], harnessName),
	})
	config.Consilium = append(config.Consilium, ConsiliumRole{
		Role:  "mobile",
		Agent: matchRole(discovered, mobileKeywords, primaryLang, defaultRoleKeywords["mobile"], harnessName),
	})

	// Exec agents from detected stacks
	for _, s := range stacks {
		if ek, ok := execKeywords[s.Name]; ok {
			agent := matchWithFallback(discovered, ek.Keywords, harnessName)
			if agent != "" {
				scope := buildScopeFromKeywords(s, ek)
				config.Exec = append(config.Exec, ExecAgent{
					Agent: agent,
					Scope: scope,
				})
			}
		}
	}

	return config
}

// --- Structure + Suggestions API ---

type AgentStructure struct {
	Roles      []string
	ExecScopes []ExecScope
}

type ExecScope struct {
	StackName string // "spring-boot" — display in wizard
	Scope     string // "backend/**/*.kt"
}

type Suggestions struct {
	Consilium  map[string]string // role -> suggested agent
	Exec       map[string]string // stackName -> suggested agent
	ModelTiers map[string]string // role -> suggested tier (high/medium/low)
}

func ResolveStructure(stacks []detector.Stack) AgentStructure {
	s := AgentStructure{}

	// Always include all 9 roles
	s.Roles = []string{"architect", "frontend", "ui", "security", "devops", "api", "diagnostics", "test", "mobile"}

	// Exec scopes from detected stacks
	for _, st := range stacks {
		if ek, ok := execKeywords[st.Name]; ok {
			scope := buildScopeFromKeywords(st, ek)
			s.ExecScopes = append(s.ExecScopes, ExecScope{
				StackName: st.Name,
				Scope:     scope,
			})
		}
	}

	return s
}

func GetSuggestions(stacks []detector.Stack, discovered []string, harnessName string) Suggestions {
	sug := Suggestions{
		Consilium:  make(map[string]string),
		Exec:       make(map[string]string),
		ModelTiers: DefaultModelTiers(),
	}

	primaryLang, frontendName := extractLangAndFrontend(stacks)

	// Consilium suggestions
	sug.Consilium["architect"] = matchRole(discovered, architectKeywords, primaryLang, defaultRoleKeywords["architect"], harnessName)
	sug.Consilium["frontend"] = matchRole(discovered, frontendKeywords, frontendName, defaultRoleKeywords["frontend"], harnessName)
	sug.Consilium["ui"] = matchWithFallback(discovered, defaultRoleKeywords["ui"], harnessName)
	sug.Consilium["security"] = matchRole(discovered, securityKeywords, primaryLang, defaultRoleKeywords["security"], harnessName)
	sug.Consilium["devops"] = matchWithFallback(discovered, defaultRoleKeywords["devops"], harnessName)
	sug.Consilium["api"] = matchWithFallback(discovered, defaultRoleKeywords["api"], harnessName)
	sug.Consilium["diagnostics"] = matchRole(discovered, diagnosticsKeywords, primaryLang, defaultRoleKeywords["diagnostics"], harnessName)
	sug.Consilium["test"] = matchRole(discovered, testKeywords, primaryLang, defaultRoleKeywords["test"], harnessName)
	sug.Consilium["mobile"] = matchRole(discovered, mobileKeywords, primaryLang, defaultRoleKeywords["mobile"], harnessName)

	// Exec suggestions
	for _, st := range stacks {
		if ek, ok := execKeywords[st.Name]; ok {
			agent := matchWithFallback(discovered, ek.Keywords, harnessName)
			if agent != "" {
				sug.Exec[st.Name] = agent
			}
		}
	}

	return sug
}

// buildScopeFromKeywords generates the correct glob scope using detected path.
func buildScopeFromKeywords(s detector.Stack, ek ExecKeywords) string {
	if s.Path == "." || s.Path == "./" {
		return ek.Scope
	}

	detectedDir := strings.TrimSuffix(s.Path, "/")

	parts := strings.SplitN(ek.Scope, "/", 2)
	if len(parts) == 2 {
		return detectedDir + "/" + parts[1]
	}
	return detectedDir + "/**"
}

// extractLangAndFrontend extracts primary language and frontend name from stacks.
func extractLangAndFrontend(stacks []detector.Stack) (string, string) {
	primaryLang := ""
	frontendName := ""
	for _, s := range stacks {
		if s.Category == "backend" && primaryLang == "" {
			primaryLang = s.Lang
		}
		if s.Category == "frontend" || s.Category == "shared" {
			frontendName = s.Name
		}
	}
	if primaryLang == "" && len(stacks) > 0 {
		primaryLang = stacks[0].Lang
	}
	return primaryLang, frontendName
}
