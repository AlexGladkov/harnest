package wizard

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

const maxVisible = 5

// Pick opens an interactive dropdown picker.
// User types to filter, arrows to navigate, Enter to select, Escape to cancel.
func Pick(items []string) (string, bool) {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return "", false
	}
	defer term.Restore(fd, oldState)

	query := ""
	cursor := 0
	prevLines := 0

	draw := func() {
		// Clear previous output
		if prevLines > 0 {
			fmt.Printf("\033[%dA", prevLines) // move up
		}
		for i := 0; i <= prevLines; i++ {
			fmt.Print("\r\033[K") // clear line
			if i < prevLines {
				fmt.Print("\n")
			}
		}
		if prevLines > 0 {
			fmt.Printf("\033[%dA", prevLines)
		}

		// Filter
		matches := filter(items, query)

		// Draw input line
		fmt.Printf("\r\033[K  > %s", query)

		// Draw matches
		show := matches
		if len(show) > maxVisible {
			show = show[:maxVisible]
		}
		if cursor >= len(show) {
			cursor = len(show) - 1
		}
		if cursor < 0 {
			cursor = 0
		}

		for i, item := range show {
			fmt.Print("\n\033[K")
			if i == cursor {
				fmt.Printf("  \033[7m %s \033[0m", item) // inverted = selected
			} else {
				fmt.Printf("    %s", item)
			}
		}
		if len(matches) > maxVisible {
			fmt.Printf("\n\033[K    ... +%d more", len(matches)-maxVisible)
			prevLines = len(show) + 1
		} else {
			prevLines = len(show)
		}

		// Move cursor back to input line
		if prevLines > 0 {
			fmt.Printf("\033[%dA", prevLines)
		}
		fmt.Printf("\r\033[%dC", len(query)+4) // position after "> query"
	}

	draw()

	buf := make([]byte, 3)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			cleanup(prevLines)
			return "", false
		}

		switch {
		// Escape
		case n == 1 && buf[0] == 27:
			cleanup(prevLines)
			return "", false

		// Ctrl+C
		case n == 1 && buf[0] == 3:
			cleanup(prevLines)
			return "", false

		// Enter
		case n == 1 && (buf[0] == 13 || buf[0] == 10):
			matches := filter(items, query)
			cleanup(prevLines)
			if len(matches) > 0 && cursor < len(matches) {
				return matches[cursor], true
			}
			return "", false

		// Backspace
		case n == 1 && (buf[0] == 127 || buf[0] == 8):
			if len(query) > 0 {
				query = query[:len(query)-1]
				cursor = 0
			}
			draw()

		// Arrow keys: ESC [ A/B
		case n == 3 && buf[0] == 27 && buf[1] == 91:
			matches := filter(items, query)
			show := len(matches)
			if show > maxVisible {
				show = maxVisible
			}
			switch buf[2] {
			case 65: // Up
				if cursor > 0 {
					cursor--
				}
			case 66: // Down
				if cursor < show-1 {
					cursor++
				}
			}
			draw()

		// Tab — autocomplete from selected
		case n == 1 && buf[0] == 9:
			matches := filter(items, query)
			if len(matches) > 0 && cursor < len(matches) {
				cleanup(prevLines)
				return matches[cursor], true
			}

		// Printable character
		case n == 1 && buf[0] >= 32 && buf[0] < 127:
			query += string(buf[0])
			cursor = 0
			draw()
		}
	}
}

func filter(items []string, query string) []string {
	if query == "" {
		return items
	}
	q := strings.ToLower(query)
	var results []string
	for _, item := range items {
		if strings.Contains(strings.ToLower(item), q) {
			results = append(results, item)
		}
	}
	return results
}

func cleanup(lines int) {
	// Clear dropdown lines
	for i := 0; i <= lines; i++ {
		fmt.Print("\r\033[K")
		if i < lines {
			fmt.Print("\n")
		}
	}
	if lines > 0 {
		fmt.Printf("\033[%dA", lines)
	}
	fmt.Print("\r\033[K")
}
