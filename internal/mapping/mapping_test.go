package mapping

import (
	"testing"
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
