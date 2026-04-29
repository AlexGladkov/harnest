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
	"swift":      "voltagent-lang:swift-expert",
	"python":     "voltagent-lang:python-pro",
	"typescript": "voltagent-lang:typescript-pro",
	"go":         "voltagent-lang:golang-pro",
	"rust":       "voltagent-lang:rust-engineer",
	"dart":       "voltagent-lang:flutter-expert",
}

var frontendMap = map[string]string{
	"vue":       "voltagent-lang:vue-expert",
	"react":     "voltagent-lang:react-specialist",
	"nextjs":    "voltagent-lang:nextjs-developer",
	"angular":   "voltagent-lang:angular-architect",
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
	"spring-boot":           {Agent: "builder-spring-feature", Scope: "backend/**/*.kt"},
	"compose-multiplatform": {Agent: "kotlin-multiplatform-developer", Scope: "composeApp/**/*.kt"},
	"android":               {Agent: "kotlin-multiplatform-developer", Scope: "app/**/*.kt"},
	"ios-native":            {Agent: "voltagent-lang:swift-expert", Scope: "iosApp/**/*.swift"},
	"swift-package":         {Agent: "voltagent-lang:swift-expert", Scope: "**/*.swift"},
	"vue":                   {Agent: "voltagent-lang:vue-expert", Scope: "vue-frontend/**"},
	"react":                 {Agent: "voltagent-lang:react-specialist", Scope: "frontend/**/*.tsx"},
	"nextjs":                {Agent: "voltagent-lang:nextjs-developer", Scope: "**/*.tsx"},
	"angular":               {Agent: "voltagent-lang:angular-architect", Scope: "src/**/*.ts"},
	"node":                  {Agent: "voltagent-lang:node-specialist", Scope: "server/**/*.ts"},
	"fastapi":               {Agent: "voltagent-lang:fastapi-developer", Scope: "**/*.py"},
	"django":                {Agent: "voltagent-lang:django-developer", Scope: "**/*.py"},
	"flask":                 {Agent: "voltagent-lang:python-pro", Scope: "**/*.py"},
	"go":                    {Agent: "voltagent-lang:golang-pro", Scope: "**/*.go"},
	"rust":                  {Agent: "voltagent-lang:rust-engineer", Scope: "**/*.rs"},
	"flutter":               {Agent: "voltagent-lang:flutter-expert", Scope: "lib/**/*.dart"},
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

func Resolve(stacks []detector.Stack) AgentConfig {
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

func GetSuggestions(stacks []detector.Stack) Suggestions {
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
