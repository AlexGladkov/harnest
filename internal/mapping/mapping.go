package mapping

import "github.com/AlexGladkov/harnest/internal/detector"

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
	defaultUI      = "voltagent-core-dev:ui-designer"
	defaultDevops  = "devops-orchestrator"
	defaultAPI     = "voltagent-core-dev:api-designer"
	defaultSecurity = "voltagent-infra:security-engineer"
	defaultDiag    = "kotlin-diagnostics"
	defaultTest    = "test-spring"
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

	if frontendName != "" {
		config.Consilium = append(config.Consilium, ConsiliumRole{
			Role:  "frontend",
			Agent: lookupOrDefault(frontendMap, frontendName, "voltagent-lang:vue-expert"),
		})
	}

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

	// Exec agents from detected stacks
	for _, s := range stacks {
		if ea, ok := execMap[s.Name]; ok {
			// Adjust scope path based on detected path
			scope := ea.Scope
			if s.Path != "." && s.Path != "./" {
				// scope already has the right prefix for known structures
			}
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

	// Always include all 8 roles
	s.Roles = []string{"architect", "frontend", "ui", "security", "devops", "api", "diagnostics", "test"}

	// Exec scopes from detected stacks
	for _, st := range stacks {
		if ea, ok := execMap[st.Name]; ok {
			s.ExecScopes = append(s.ExecScopes, ExecScope{
				StackName: st.Name,
				Scope:     ea.Scope,
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

	// Consilium suggestions
	if v, ok := architectMap[primaryLang]; ok {
		sug.Consilium["architect"] = v
	}
	if frontendName != "" {
		if v, ok := frontendMap[frontendName]; ok {
			sug.Consilium["frontend"] = v
		}
	}
	sug.Consilium["ui"] = defaultUI
	if v, ok := securityMap[primaryLang]; ok {
		sug.Consilium["security"] = v
	} else {
		sug.Consilium["security"] = defaultSecurity
	}
	sug.Consilium["devops"] = defaultDevops
	sug.Consilium["api"] = defaultAPI
	if v, ok := diagnosticsMap[primaryLang]; ok {
		sug.Consilium["diagnostics"] = v
	}
	if v, ok := testMap[primaryLang]; ok {
		sug.Consilium["test"] = v
	}

	// Exec suggestions
	for _, st := range stacks {
		if ea, ok := execMap[st.Name]; ok {
			sug.Exec[st.Name] = ea.Agent
		}
	}

	return sug
}

func lookupOrDefault(m map[string]string, key, fallback string) string {
	if v, ok := m[key]; ok {
		return v
	}
	return fallback
}
