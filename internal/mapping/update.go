package mapping

import "fmt"

// Update fetches latest agent mappings from remote.
// For now — placeholder. Future: fetch from GitHub releases or registry.
func Update() error {
	// TODO: fetch latest mappings from github.com/AlexGladkov/harnest-registry
	// For now, mappings are compiled into the binary.
	fmt.Println("  Mappings are compiled into binary (v0.1.0)")
	fmt.Println("  To update: brew upgrade harnest")
	return nil
}
