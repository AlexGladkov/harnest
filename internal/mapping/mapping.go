package mapping

import (
	"strings"

	"github.com/AlexGladkov/harnest/internal/detector"
)

type AgentConfig struct {
	Consilium []ConsiliumRole
	Exec      []ExecAgent
}

type ConsiliumRole struct {
	Role  string
	Agent string
}

type ExecAgent struct {
	Agent string
	Scope string
}

// agent lookup tables

var architectMap = map[string]string{
	"kotlin":     "voltagent-lang:java-architect",
	"java":       "voltagent-lang:java-architect",
	"scala":      "voltagent-lang:java-architect",
	"groovy":     "voltagent-lang:java-architect",
	"clojure":    "voltagent-lang:java-architect",
	"swift":      "voltagent-lang:swift-expert",
	"python":     "voltagent-lang:python-pro",
	"typescript": "voltagent-lang:typescript-pro",
	"go":         "voltagent-lang:golang-pro",
	"rust":       "voltagent-lang:rust-engineer",
	"dart":       "voltagent-lang:flutter-expert",
	"ruby":       "voltagent-lang:rails-expert",
	"php":        "voltagent-lang:php-pro",
	"csharp":     "voltagent-lang:dotnet-core-expert",
	"elixir":     "voltagent-lang:elixir-expert",
	"erlang":     "voltagent-lang:elixir-expert",
	"gleam":      "voltagent-lang:elixir-expert",
	"haskell":    "voltagent-lang:elixir-expert",
	"ocaml":      "voltagent-lang:elixir-expert",
	"c":          "voltagent-lang:cpp-pro",
	"cpp":        "voltagent-lang:cpp-pro",
	"zig":        "voltagent-lang:cpp-pro",
	"nim":        "voltagent-lang:python-pro",
	"vlang":      "voltagent-lang:golang-pro",
	"crystal":    "voltagent-lang:ruby-pro",
	"julia":      "voltagent-lang:python-pro",
	"r":          "voltagent-lang:python-pro",
	"lua":        "voltagent-lang:javascript-pro",
	"perl":       "voltagent-lang:python-pro",
	"hcl":        "voltagent-infra:terraform-engineer",
	"yaml":       "voltagent-infra:devops-engineer",
}

var frontendMap = map[string]string{
	"vue":       "voltagent-lang:vue-expert",
	"nuxt":      "voltagent-lang:vue-expert",
	"react":     "voltagent-lang:react-specialist",
	"nextjs":    "voltagent-lang:nextjs-developer",
	"gatsby":    "voltagent-lang:react-specialist",
	"remix":     "voltagent-lang:react-specialist",
	"angular":   "voltagent-lang:angular-architect",
	"svelte":    "voltagent-lang:javascript-pro",
	"sveltekit": "voltagent-lang:javascript-pro",
	"solid":     "voltagent-lang:javascript-pro",
	"qwik":      "voltagent-lang:javascript-pro",
	"astro":     "voltagent-lang:javascript-pro",
	"ember":     "voltagent-lang:javascript-pro",
	"eleventy":  "voltagent-lang:javascript-pro",
	"hugo":      "voltagent-lang:golang-pro",
	"jekyll":    "voltagent-lang:rails-expert",
	"flutter":   "voltagent-lang:flutter-expert",
	"swiftui":   "voltagent-lang:swift-expert",
}

var mobileMap = map[string]string{
	"kotlin":  "kotlin-multiplatform-developer",
	"swift":   "voltagent-lang:swift-expert",
	"dart":    "voltagent-lang:flutter-expert",
}

var securityMap = map[string]string{
	"kotlin": "security-kotlin",
}

var diagnosticsMap = map[string]string{
	"kotlin": "kotlin-diagnostics",
}

var devopsMap = map[string]string{
	// default for all
}

var testMap = map[string]string{
	"kotlin": "test-spring",
}

var execMap = map[string]ExecAgent{
	// --- Kotlin ---
	"spring-boot":           {Agent: "builder-spring-feature", Scope: "backend/**/*.kt"},
	"ktor":                  {Agent: "voltagent-lang:kotlin-specialist", Scope: "backend/**/*.kt"},
	"quarkus":               {Agent: "voltagent-lang:spring-boot-engineer", Scope: "src/**/*.kt"},
	"micronaut":             {Agent: "voltagent-lang:spring-boot-engineer", Scope: "src/**/*.kt"},
	"compose-multiplatform": {Agent: "kotlin-multiplatform-developer", Scope: "composeApp/**/*.kt"},
	"android":               {Agent: "kotlin-multiplatform-developer", Scope: "app/**/*.kt"},

	// --- Java ---
	"spring-boot-java": {Agent: "voltagent-lang:spring-boot-engineer", Scope: "src/**/*.java"},
	"java":             {Agent: "voltagent-lang:java-architect", Scope: "src/**/*.java"},

	// --- Swift ---
	"ios-native":    {Agent: "voltagent-lang:swift-expert", Scope: "iosApp/**/*.swift"},
	"swift-package": {Agent: "voltagent-lang:swift-expert", Scope: "**/*.swift"},
	"vapor":         {Agent: "voltagent-lang:swift-expert", Scope: "Sources/**/*.swift"},

	// --- JS/TS Frontend ---
	"vue":       {Agent: "voltagent-lang:vue-expert", Scope: "src/**/*.vue"},
	"nuxt":      {Agent: "voltagent-lang:vue-expert", Scope: "src/**/*.vue"},
	"react":     {Agent: "voltagent-lang:react-specialist", Scope: "src/**/*.tsx"},
	"nextjs":    {Agent: "voltagent-lang:nextjs-developer", Scope: "src/**/*.tsx"},
	"gatsby":    {Agent: "voltagent-lang:react-specialist", Scope: "src/**/*.tsx"},
	"remix":     {Agent: "voltagent-lang:react-specialist", Scope: "app/**/*.tsx"},
	"angular":   {Agent: "voltagent-lang:angular-architect", Scope: "src/**/*.ts"},
	"svelte":    {Agent: "voltagent-lang:javascript-pro", Scope: "src/**/*.svelte"},
	"sveltekit": {Agent: "voltagent-lang:javascript-pro", Scope: "src/**/*.svelte"},
	"solid":     {Agent: "voltagent-lang:javascript-pro", Scope: "src/**/*.tsx"},
	"qwik":      {Agent: "voltagent-lang:javascript-pro", Scope: "src/**/*.tsx"},
	"astro":     {Agent: "voltagent-lang:javascript-pro", Scope: "src/**/*.astro"},
	"ember":     {Agent: "voltagent-lang:javascript-pro", Scope: "app/**/*.js"},
	"eleventy":  {Agent: "voltagent-lang:javascript-pro", Scope: "src/**/*.njk"},

	// --- JS/TS Backend ---
	"node":   {Agent: "voltagent-lang:node-specialist", Scope: "src/**/*.ts"},
	"deno":   {Agent: "voltagent-lang:typescript-pro", Scope: "**/*.ts"},
	"bun":    {Agent: "voltagent-lang:typescript-pro", Scope: "**/*.ts"},
	"strapi": {Agent: "voltagent-lang:node-specialist", Scope: "src/**/*.js"},

	// --- JS/TS Mobile ---
	"expo":         {Agent: "voltagent-lang:expo-react-native-expert", Scope: "src/**/*.tsx"},
	"react-native": {Agent: "voltagent-lang:expo-react-native-expert", Scope: "src/**/*.tsx"},
	"ionic":        {Agent: "voltagent-lang:react-specialist", Scope: "src/**/*.tsx"},
	"capacitor":    {Agent: "voltagent-lang:react-specialist", Scope: "src/**/*.tsx"},

	// --- JS/TS Desktop ---
	"electron": {Agent: "voltagent-core-dev:electron-pro", Scope: "src/**/*.ts"},
	"tauri":    {Agent: "voltagent-lang:rust-engineer", Scope: "src-tauri/**/*.rs"},

	// --- Python ---
	"fastapi":   {Agent: "voltagent-lang:fastapi-developer", Scope: "**/*.py"},
	"django":    {Agent: "voltagent-lang:django-developer", Scope: "**/*.py"},
	"flask":     {Agent: "voltagent-lang:python-pro", Scope: "**/*.py"},
	"starlette": {Agent: "voltagent-lang:fastapi-developer", Scope: "**/*.py"},
	"pyramid":   {Agent: "voltagent-lang:python-pro", Scope: "**/*.py"},
	"litestar":  {Agent: "voltagent-lang:python-pro", Scope: "**/*.py"},
	"streamlit": {Agent: "voltagent-lang:python-pro", Scope: "**/*.py"},
	"gradio":    {Agent: "voltagent-lang:python-pro", Scope: "**/*.py"},
	"jupyter":   {Agent: "voltagent-lang:python-pro", Scope: "**/*.ipynb"},

	// --- Go ---
	"go":      {Agent: "voltagent-lang:golang-pro", Scope: "**/*.go"},
	"gin":     {Agent: "voltagent-lang:golang-pro", Scope: "**/*.go"},
	"fiber":   {Agent: "voltagent-lang:golang-pro", Scope: "**/*.go"},
	"echo":    {Agent: "voltagent-lang:golang-pro", Scope: "**/*.go"},
	"chi":     {Agent: "voltagent-lang:golang-pro", Scope: "**/*.go"},
	"buffalo": {Agent: "voltagent-lang:golang-pro", Scope: "**/*.go"},

	// --- Rust ---
	"rust":   {Agent: "voltagent-lang:rust-engineer", Scope: "src/**/*.rs"},
	"axum":   {Agent: "voltagent-lang:rust-engineer", Scope: "src/**/*.rs"},
	"actix":  {Agent: "voltagent-lang:rust-engineer", Scope: "src/**/*.rs"},
	"rocket": {Agent: "voltagent-lang:rust-engineer", Scope: "src/**/*.rs"},
	"warp":   {Agent: "voltagent-lang:rust-engineer", Scope: "src/**/*.rs"},

	// --- Dart ---
	"flutter": {Agent: "voltagent-lang:flutter-expert", Scope: "lib/**/*.dart"},

	// --- Ruby ---
	"rails":   {Agent: "voltagent-lang:rails-expert", Scope: "app/**/*.rb"},
	"sinatra": {Agent: "voltagent-lang:rails-expert", Scope: "**/*.rb"},
	"jekyll":  {Agent: "voltagent-lang:rails-expert", Scope: "**/*.rb"},

	// --- PHP ---
	"laravel":   {Agent: "voltagent-lang:laravel-specialist", Scope: "app/**/*.php"},
	"symfony":   {Agent: "voltagent-lang:symfony-specialist", Scope: "src/**/*.php"},
	"wordpress": {Agent: "voltagent-lang:php-pro", Scope: "**/*.php"},

	// --- C# / .NET ---
	"dotnet": {Agent: "voltagent-lang:dotnet-core-expert", Scope: "**/*.cs"},
	"maui":   {Agent: "voltagent-lang:dotnet-core-expert", Scope: "**/*.cs"},

	// --- Elixir / Erlang / BEAM ---
	"phoenix": {Agent: "voltagent-lang:elixir-expert", Scope: "lib/**/*.ex"},
	"elixir":  {Agent: "voltagent-lang:elixir-expert", Scope: "lib/**/*.ex"},
	"erlang":  {Agent: "voltagent-lang:elixir-expert", Scope: "src/**/*.erl"},
	"gleam":   {Agent: "voltagent-lang:elixir-expert", Scope: "src/**/*.gleam"},

	// --- JVM (non-Java/Kotlin) ---
	"scala":   {Agent: "voltagent-lang:java-architect", Scope: "src/**/*.scala"},
	"play":    {Agent: "voltagent-lang:java-architect", Scope: "app/**/*.scala"},
	"akka":    {Agent: "voltagent-lang:java-architect", Scope: "src/**/*.scala"},
	"clojure": {Agent: "voltagent-lang:java-architect", Scope: "src/**/*.clj"},
	"grails":  {Agent: "voltagent-lang:java-architect", Scope: "grails-app/**/*.groovy"},

	// --- C / C++ ---
	"c":   {Agent: "voltagent-lang:cpp-pro", Scope: "src/**/*.c"},
	"cpp": {Agent: "voltagent-lang:cpp-pro", Scope: "src/**/*.cpp"},

	// --- Systems / Emerging ---
	"zig":     {Agent: "voltagent-lang:cpp-pro", Scope: "src/**/*.zig"},
	"nim":     {Agent: "voltagent-lang:python-pro", Scope: "src/**/*.nim"},
	"vlang":   {Agent: "voltagent-lang:golang-pro", Scope: "src/**/*.v"},
	"crystal": {Agent: "voltagent-lang:rails-expert", Scope: "src/**/*.cr"},

	// --- Functional ---
	"haskell": {Agent: "voltagent-lang:elixir-expert", Scope: "src/**/*.hs"},
	"ocaml":   {Agent: "voltagent-lang:elixir-expert", Scope: "lib/**/*.ml"},

	// --- Scientific / Data ---
	"julia": {Agent: "voltagent-lang:python-pro", Scope: "src/**/*.jl"},
	"r":     {Agent: "voltagent-lang:python-pro", Scope: "R/**/*.R"},

	// --- Scripting ---
	"lua":  {Agent: "voltagent-lang:javascript-pro", Scope: "**/*.lua"},
	"perl": {Agent: "voltagent-lang:python-pro", Scope: "lib/**/*.pm"},

	// --- Static Site Generators ---
	"hugo": {Agent: "voltagent-lang:golang-pro", Scope: "content/**/*.md"},

	// --- Infra ---
	"docker":         {Agent: "voltagent-infra:docker-expert", Scope: "**/Dockerfile"},
	"terraform":      {Agent: "voltagent-infra:terraform-engineer", Scope: "**/*.tf"},
	"helm":           {Agent: "voltagent-infra:kubernetes-specialist", Scope: "**/*.yaml"},
	"pulumi":         {Agent: "voltagent-infra:cloud-architect", Scope: "**/*"},
	"ansible":        {Agent: "voltagent-infra:devops-engineer", Scope: "**/*.yml"},
	"github-actions": {Agent: "voltagent-infra:deployment-engineer", Scope: ".github/workflows/**/*.yml"},
}

const (
	defaultFrontend = "voltagent-lang:vue-expert"
	defaultUI       = "voltagent-core-dev:ui-designer"
	defaultSecurity = "voltagent-infra:security-engineer"
	defaultDevops   = "devops-orchestrator"
	defaultAPI      = "voltagent-core-dev:api-designer"
	defaultDiag     = "kotlin-diagnostics"
	defaultTest     = "test-spring"
	defaultMobile   = "kotlin-multiplatform-developer"
)

func Resolve(stacks []detector.Stack, _discovered []string, _harnessName string) AgentConfig {
	config := AgentConfig{}

	// Determine primary language from stacks
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

	// Consilium roles
	config.Consilium = append(config.Consilium, ConsiliumRole{
		Role:  "architect",
		Agent: lookupOrDefault(architectMap, primaryLang, "voltagent-lang:java-architect"),
	})

	config.Consilium = append(config.Consilium, ConsiliumRole{
		Role:  "frontend",
		Agent: lookupOrDefault(frontendMap, frontendName, defaultFrontend),
	})
	config.Consilium = append(config.Consilium, ConsiliumRole{Role: "ui", Agent: defaultUI})
	config.Consilium = append(config.Consilium, ConsiliumRole{
		Role:  "security",
		Agent: lookupOrDefault(securityMap, primaryLang, defaultSecurity),
	})
	config.Consilium = append(config.Consilium, ConsiliumRole{Role: "devops", Agent: defaultDevops})
	config.Consilium = append(config.Consilium, ConsiliumRole{Role: "api", Agent: defaultAPI})
	config.Consilium = append(config.Consilium, ConsiliumRole{
		Role:  "diagnostics",
		Agent: lookupOrDefault(diagnosticsMap, primaryLang, defaultDiag),
	})
	config.Consilium = append(config.Consilium, ConsiliumRole{
		Role:  "test",
		Agent: lookupOrDefault(testMap, primaryLang, defaultTest),
	})
	config.Consilium = append(config.Consilium, ConsiliumRole{
		Role:  "mobile",
		Agent: lookupOrDefault(mobileMap, primaryLang, defaultMobile),
	})

	// Exec agents from detected stacks
	for _, s := range stacks {
		if ea, ok := execMap[s.Name]; ok {
			scope := buildScope(s, ea)
			config.Exec = append(config.Exec, ExecAgent{
				Agent: ea.Agent,
				Scope: scope,
			})
		}
	}

	return config
}

// --- Structure + Suggestions API (v0.2.0) ---

type AgentStructure struct {
	Roles      []string
	ExecScopes []ExecScope
}

type ExecScope struct {
	StackName string // "spring-boot" — display in wizard
	Scope     string // "backend/**/*.kt"
}

type Suggestions struct {
	Consilium map[string]string // role -> suggested agent
	Exec      map[string]string // stackName -> suggested agent
}

func ResolveStructure(stacks []detector.Stack) AgentStructure {
	s := AgentStructure{}

	// Always include all 9 roles
	s.Roles = []string{"architect", "frontend", "ui", "security", "devops", "api", "diagnostics", "test", "mobile"}

	// Exec scopes from detected stacks
	for _, st := range stacks {
		if ea, ok := execMap[st.Name]; ok {
			scope := buildScope(st, ea)
			s.ExecScopes = append(s.ExecScopes, ExecScope{
				StackName: st.Name,
				Scope:     scope,
			})
		}
	}

	return s
}

func GetSuggestions(stacks []detector.Stack, _discovered []string, _harnessName string) Suggestions {
	sug := Suggestions{
		Consilium: make(map[string]string),
		Exec:      make(map[string]string),
	}

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

	// Consilium suggestions — every role always gets a default
	sug.Consilium["architect"] = lookupOrDefault(architectMap, primaryLang, "voltagent-lang:java-architect")
	sug.Consilium["frontend"] = lookupOrDefault(frontendMap, frontendName, defaultFrontend)
	sug.Consilium["ui"] = defaultUI
	sug.Consilium["security"] = lookupOrDefault(securityMap, primaryLang, defaultSecurity)
	sug.Consilium["devops"] = defaultDevops
	sug.Consilium["api"] = defaultAPI
	sug.Consilium["diagnostics"] = lookupOrDefault(diagnosticsMap, primaryLang, defaultDiag)
	sug.Consilium["test"] = lookupOrDefault(testMap, primaryLang, defaultTest)
	sug.Consilium["mobile"] = lookupOrDefault(mobileMap, primaryLang, defaultMobile)

	// Exec suggestions
	for _, st := range stacks {
		if ea, ok := execMap[st.Name]; ok {
			sug.Exec[st.Name] = ea.Agent
		}
	}

	return sug
}

// buildScope generates the correct glob scope using the detected path.
// If the detected path differs from the hardcoded exec scope prefix,
// replace the prefix with the actual detected path.
func buildScope(s detector.Stack, ea ExecAgent) string {
	if s.Path == "." || s.Path == "./" {
		return ea.Scope
	}

	detectedDir := strings.TrimSuffix(s.Path, "/")

	// Extract the file extension pattern from the default scope (e.g. "**/*.kt")
	// Default scope format: "prefix/**/*.ext" or "prefix/**"
	parts := strings.SplitN(ea.Scope, "/", 2)
	if len(parts) == 2 {
		return detectedDir + "/" + parts[1]
	}
	return detectedDir + "/**"
}

func lookupOrDefault(m map[string]string, key, fallback string) string {
	if v, ok := m[key]; ok {
		return v
	}
	return fallback
}
