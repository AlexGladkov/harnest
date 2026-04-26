package agents

import (
	"fmt"
	"testing"
)

func TestDiscover(t *testing.T) {
	all := Discover()
	fmt.Printf("Found %d agents:\n", len(all))
	for _, a := range all {
		fmt.Printf("  %s\n", a)
	}
	if len(all) == 0 {
		t.Log("No agents found (may be expected in CI)")
	}
}
