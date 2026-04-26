package wizard

import (
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

const maxVisible = 5

// Pick opens an interactive dropdown picker.
func Pick(items []string) (string, bool) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return "", false
	}
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return "", false
	}
	defer func() {
		term.Restore(fd, oldState)
		fmt.Print("\r\n") // newline after picker closes
	}()

	query := ""
	cursor := 0
	linesDrawn := 0

	redraw := func() {
		// Move to start of our drawing area
		if linesDrawn > 0 {
			// Move up to the input line
			if linesDrawn > 1 {
				write("\x1b[%dA", linesDrawn-1)
			}
		}
		// Go to column 0 and clear from here to end of screen
		write("\r\x1b[J")

		// Filter
		matches := filterItems(items, query)
		if cursor >= len(matches) {
			cursor = len(matches) - 1
		}
		if cursor < 0 {
			cursor = 0
		}

		show := matches
		if len(show) > maxVisible {
			show = show[:maxVisible]
		}

		// Draw input line
		write("  > %s", query)

		// Draw items
		for i, item := range show {
			if i == cursor {
				write("\r\n  \x1b[7m %s \x1b[0m", item)
			} else {
				write("\r\n    %s", item)
			}
		}

		if len(matches) > maxVisible {
			write("\r\n    ... +%d more", len(matches)-maxVisible)
			linesDrawn = 1 + len(show) + 1
		} else {
			linesDrawn = 1 + len(show)
		}

		// Move cursor back to input line
		up := linesDrawn - 1
		if up > 0 {
			write("\x1b[%dA", up)
		}
		// Position at end of query text
		write("\r\x1b[%dC", len(query)+4) // "  > " = 4 chars
	}

	redraw()

	for {
		b, esc, ok := readKey(fd)
		if !ok {
			clearToEnd(linesDrawn)
			return "", false
		}

		if esc != 0 {
			matches := filterItems(items, query)
			visible := len(matches)
			if visible > maxVisible {
				visible = maxVisible
			}
			switch esc {
			case 'A': // Up
				if cursor > 0 {
					cursor--
					redraw()
				}
			case 'B': // Down
				if cursor < visible-1 {
					cursor++
					redraw()
				}
			}
			continue
		}

		switch b {
		case 27: // Escape (plain, no sequence)
			clearToEnd(linesDrawn)
			return "", false

		case 3: // Ctrl+C
			clearToEnd(linesDrawn)
			return "", false

		case 13, 10: // Enter
			matches := filterItems(items, query)
			clearToEnd(linesDrawn)
			if len(matches) > 0 && cursor < len(matches) {
				return matches[cursor], true
			}
			return "", false

		case 9: // Tab
			matches := filterItems(items, query)
			clearToEnd(linesDrawn)
			if len(matches) > 0 && cursor < len(matches) {
				return matches[cursor], true
			}
			return "", false

		case 127, 8: // Backspace
			if len(query) > 0 {
				query = query[:len(query)-1]
				cursor = 0
				redraw()
			}

		default:
			if b >= 32 && b < 127 {
				query += string(rune(b))
				cursor = 0
				redraw()
			}
		}
	}
}

// readKey reads a single key press. Returns (byte, escapeChar, ok).
// For arrow keys: escapeChar = 'A'(up), 'B'(down), 'C'(right), 'D'(left).
// For normal keys: byte = the character, escapeChar = 0.
func readKey(fd int) (byte, byte, bool) {
	buf := make([]byte, 1)
	n, err := os.Stdin.Read(buf)
	if err != nil || n == 0 {
		return 0, 0, false
	}

	if buf[0] != 27 {
		return buf[0], 0, true
	}

	// Got ESC — check if it's an escape sequence or plain Escape
	// Set a short deadline to distinguish
	os.Stdin.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
	seq := make([]byte, 2)
	n, err = os.Stdin.Read(seq)
	os.Stdin.SetReadDeadline(time.Time{}) // clear deadline

	if err != nil || n < 2 {
		// Plain Escape key
		return 27, 0, true
	}

	if seq[0] == '[' {
		return 0, seq[1], true
	}

	return 27, 0, true
}

func filterItems(items []string, query string) []string {
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

func clearToEnd(lines int) {
	if lines > 1 {
		write("\x1b[%dA", lines-1)
	}
	write("\r\x1b[J")
}

func write(format string, args ...any) {
	fmt.Fprintf(os.Stdout, format, args...)
}
