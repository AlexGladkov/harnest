package mapping

import (
	"testing"

	"github.com/AlexGladkov/harnest/internal/detector"
)

func TestMatchAgent_ExactSubstring(t *testing.T) {
	discovered := []string{"voltagent-lang:java-architect", "voltagent-lang:python-pro", "test-spring"}
	got := MatchAgent(discovered, []string{"java", "architect"})
	if got != "voltagent-lang:java-architect" {
		t.Errorf("expected voltagent-lang:java-architect, got %q", got)
	}
}

func TestMatchAgent_MultiKeywordScoring(t *testing.T) {
	discovered := []string{"my-kotlin-agent", "kotlin-multiplatform-developer"}
	// "kotlin" matches both, "multiplatform" matches only second → second wins
	got := MatchAgent(discovered, []string{"kotlin", "multiplatform"})
	if got != "kotlin-multiplatform-developer" {
		t.Errorf("expected kotlin-multiplatform-developer, got %q", got)
	}
}

func TestMatchAgent_SingleKeyword(t *testing.T) {
	discovered := []string{"security-kotlin", "voltagent-lang:python-pro"}
	got := MatchAgent(discovered, []string{"security"})
	if got != "security-kotlin" {
		t.Errorf("expected security-kotlin, got %q", got)
	}
}

func TestMatchAgent_NoMatch(t *testing.T) {
	discovered := []string{"voltagent-lang:python-pro"}
	got := MatchAgent(discovered, []string{"rust", "engineer"})
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestMatchAgent_EmptyDiscovered(t *testing.T) {
	got := MatchAgent(nil, []string{"java"})
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestMatchAgent_EmptyKeywords(t *testing.T) {
	got := MatchAgent([]string{"some-agent"}, nil)
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestMatchWithFallback_ClaudeCodeFallback(t *testing.T) {
	// No match → claude-code gets "general-purpose"
	got := matchWithFallback([]string{"unrelated-agent"}, []string{"nonexistent"}, "claude-code")
	if got != "general-purpose" {
		t.Errorf("expected general-purpose, got %q", got)
	}
}

func TestMatchWithFallback_CursorNoFallback(t *testing.T) {
	// No match → cursor gets empty
	got := matchWithFallback([]string{"unrelated-agent"}, []string{"nonexistent"}, "cursor")
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestMatchWithFallback_MatchBeforeFallback(t *testing.T) {
	// Match found → no fallback needed
	got := matchWithFallback([]string{"my-rust-engineer"}, []string{"rust"}, "claude-code")
	if got != "my-rust-engineer" {
		t.Errorf("expected my-rust-engineer, got %q", got)
	}
}

func TestMatchAgent_CaseInsensitive(t *testing.T) {
	discovered := []string{"MyJavaArchitect"}
	got := MatchAgent(discovered, []string{"java", "architect"})
	if got != "MyJavaArchitect" {
		t.Errorf("expected MyJavaArchitect, got %q", got)
	}
}

func TestMatchAgent_TieBreaker(t *testing.T) {
	// When scores are equal, first in list wins (project agents come first)
	discovered := []string{"qa-engineer", "ai-sdlc:test-spring"}
	// qa-engineer matches "qa" (1), ai-sdlc:test-spring matches "test" (1) → tie
	got := MatchAgent(discovered, []string{"qa", "test"})
	if got != "qa-engineer" {
		t.Errorf("expected qa-engineer (first in tie), got %q", got)
	}
}

func TestMatchAgent_KeywordQa(t *testing.T) {
	// "qa" keyword should match qa-engineer
	discovered := []string{"qa-engineer", "general-purpose"}
	got := MatchAgent(discovered, []string{"test", "qa"})
	if got != "qa-engineer" {
		t.Errorf("expected qa-engineer, got %q", got)
	}
}

func TestMatchAgent_KeywordSre(t *testing.T) {
	// "sre" keyword should match sre-platform-agent
	discovered := []string{"sre-platform-agent", "general-purpose"}
	got := MatchAgent(discovered, []string{"devops", "sre"})
	if got != "sre-platform-agent" {
		t.Errorf("expected sre-platform-agent, got %q", got)
	}
}

// --- Language keyword map coverage tests ---

func TestMatchRole_Security_Go(t *testing.T) {
	discovered := []string{"golang-security-auditor", "general-purpose"}
	got := matchRole(discovered, securityKeywords, "go", defaultRoleKeywords["security"], "claude-code")
	if got != "golang-security-auditor" {
		t.Errorf("expected golang-security-auditor, got %q", got)
	}
}

func TestMatchRole_Security_CSharp(t *testing.T) {
	discovered := []string{"owasp-dotnet-auditor", "general-purpose"}
	// "dotnet" keyword matches "owasp-dotnet-auditor"
	got := matchRole(discovered, securityKeywords, "csharp", defaultRoleKeywords["security"], "claude-code")
	if got != "owasp-dotnet-auditor" {
		t.Errorf("expected owasp-dotnet-auditor, got %q", got)
	}
}

func TestMatchRole_Security_Python(t *testing.T) {
	discovered := []string{"python-security-scanner", "general-purpose"}
	got := matchRole(discovered, securityKeywords, "python", defaultRoleKeywords["security"], "claude-code")
	if got != "python-security-scanner" {
		t.Errorf("expected python-security-scanner, got %q", got)
	}
}

func TestMatchRole_Diagnostics_Rust(t *testing.T) {
	discovered := []string{"rust-diagnostics-bot", "general-purpose"}
	got := matchRole(discovered, diagnosticsKeywords, "rust", defaultRoleKeywords["diagnostics"], "claude-code")
	if got != "rust-diagnostics-bot" {
		t.Errorf("expected rust-diagnostics-bot, got %q", got)
	}
}

func TestMatchRole_Test_TypeScript(t *testing.T) {
	discovered := []string{"typescript-jest-tester", "general-purpose"}
	// "jest" keyword matches "typescript-jest-tester"
	got := matchRole(discovered, testKeywords, "typescript", defaultRoleKeywords["test"], "claude-code")
	if got != "typescript-jest-tester" {
		t.Errorf("expected typescript-jest-tester, got %q", got)
	}
}

func TestMatchRole_Test_Java(t *testing.T) {
	discovered := []string{"spring-junit-tester", "general-purpose"}
	// "junit" keyword matches
	got := matchRole(discovered, testKeywords, "java", defaultRoleKeywords["test"], "claude-code")
	if got != "spring-junit-tester" {
		t.Errorf("expected spring-junit-tester, got %q", got)
	}
}

func TestMatchRole_Test_Swift(t *testing.T) {
	discovered := []string{"xctest-runner", "general-purpose"}
	// "xctest" keyword matches
	got := matchRole(discovered, testKeywords, "swift", defaultRoleKeywords["test"], "claude-code")
	if got != "xctest-runner" {
		t.Errorf("expected xctest-runner, got %q", got)
	}
}

func TestMatchRole_FallbackToDefault(t *testing.T) {
	// Language not in specific map → fallback to defaultRoleKeywords
	discovered := []string{"generic-security-bot", "other-agent"}
	got := matchRole(discovered, securityKeywords, "elixir", defaultRoleKeywords["security"], "claude-code")
	// elixir not in securityKeywords → uses defaultRoleKeywords["security"] = ["security"]
	if got != "generic-security-bot" {
		t.Errorf("expected generic-security-bot from fallback, got %q", got)
	}
}

func TestMatchRole_RoleKeywordBonus(t *testing.T) {
	// Role-name agent (e.g. "security") should beat language-only agent (e.g. "backend-csharp")
	// because default role keywords are now appended to language keywords.
	discovered := []string{"backend-csharp", "security"}
	got := matchRole(discovered, securityKeywords, "csharp", defaultRoleKeywords["security"], "claude-code")
	// kw = {"security"} + {"security","dotnet","csharp"} = {"security","security","dotnet","csharp"}
	// "backend-csharp": matches "csharp" → score 1
	// "security": matches "security" twice → score 2
	if got != "security" {
		t.Errorf("expected security (role keyword bonus), got %q", got)
	}
}

func TestExecKeywords_DockerMatchesDevops(t *testing.T) {
	// docker exec keywords include "devops" → devops-k8s should match
	ek, ok := execKeywords["docker"]
	if !ok {
		t.Fatal("execKeywords[docker] not found")
	}
	discovered := []string{"devops-k8s", "general-purpose"}
	got := MatchAgent(discovered, ek.Keywords)
	if got != "devops-k8s" {
		t.Errorf("expected devops-k8s, got %q (keywords: %v)", got, ek.Keywords)
	}
}

func TestExecKeywords_DotnetMatchesCsharp(t *testing.T) {
	// dotnet exec keywords include "csharp" → backend-csharp should match
	ek, ok := execKeywords["dotnet"]
	if !ok {
		t.Fatal("execKeywords[dotnet] not found")
	}
	discovered := []string{"backend-csharp", "general-purpose"}
	got := MatchAgent(discovered, ek.Keywords)
	if got != "backend-csharp" {
		t.Errorf("expected backend-csharp, got %q (keywords: %v)", got, ek.Keywords)
	}
}

func TestMatchRole_TestQaOverBackend(t *testing.T) {
	// "qa-engineer" should beat "backend-csharp" for test role
	discovered := []string{"backend-csharp", "qa-engineer"}
	got := matchRole(discovered, testKeywords, "csharp", defaultRoleKeywords["test"], "claude-code")
	// kw = {"test","qa","tester"} + {"test","dotnet","csharp","xunit","nunit","qa"}
	// "backend-csharp": matches "csharp","test" → score 2
	// "qa-engineer": matches "qa"×2, "test"×2 → score 4
	if got != "qa-engineer" {
		t.Errorf("expected qa-engineer (qa keyword bonus), got %q", got)
	}
}

// --- buildScopeFromKeywords tests ---

func TestBuildScope_WildcardRecursive(t *testing.T) {
	// Pattern **/*.py with non-root path → must keep **/ recursive wildcard
	s := detector.Stack{Name: "fastapi", Lang: "python", Path: "backend/"}
	ek := ExecKeywords{Keywords: []string{"fastapi"}, Scope: "**/*.py"}
	got := buildScopeFromKeywords(s, ek)
	expected := "backend/**/*.py"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestBuildScope_DirPrefix(t *testing.T) {
	// Pattern like "src/**/*.vue" — first component is dir, not wildcard
	s := detector.Stack{Name: "vue", Lang: "typescript", Path: "frontend/"}
	ek := ExecKeywords{Keywords: []string{"vue"}, Scope: "src/**/*.vue"}
	got := buildScopeFromKeywords(s, ek)
	expected := "frontend/**/*.vue"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestBuildScope_RootPath(t *testing.T) {
	// Root path "." → return scope unchanged
	s := detector.Stack{Name: "fastapi", Lang: "python", Path: "."}
	ek := ExecKeywords{Keywords: []string{"fastapi"}, Scope: "**/*.py"}
	got := buildScopeFromKeywords(s, ek)
	expected := "**/*.py"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestBuildScope_KMP_Unchanged(t *testing.T) {
	// KMP pattern "composeApp/**/*.kt" → must keep structure intact
	s := detector.Stack{Name: "compose-multiplatform", Lang: "kotlin", Path: "shared/"}
	ek := ExecKeywords{Keywords: []string{"kotlin", "multiplatform"}, Scope: "composeApp/**/*.kt"}
	got := buildScopeFromKeywords(s, ek)
	expected := "shared/**/*.kt"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestBuildScope_DotnetVsMaui(t *testing.T) {
	// Same language, different paths → scopes must differ
	dotnetStack := detector.Stack{Name: "dotnet", Lang: "csharp", Path: "backend/"}
	mauiStack := detector.Stack{Name: "maui", Lang: "csharp", Path: "maui/"}

	dotnetScope := buildScopeFromKeywords(dotnetStack, ExecKeywords{Scope: "**/*.cs"})
	mauiScope := buildScopeFromKeywords(mauiStack, ExecKeywords{Scope: "**/*.cs"})

	if dotnetScope != "backend/**/*.cs" {
		t.Errorf("dotnet: expected backend/**/*.cs, got %q", dotnetScope)
	}
	if mauiScope != "maui/**/*.cs" {
		t.Errorf("maui: expected maui/**/*.cs, got %q", mauiScope)
	}
	if dotnetScope == mauiScope {
		t.Error("dotnet and maui scopes must differ")
	}
}
