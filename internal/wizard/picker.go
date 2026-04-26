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
	drawnLines := 0 // how many lines we drew last time (including input line)

	redraw := func() {
		// Move up to first drawn line and clear everything
		if drawnLines > 1 {
			fmt.Fprintf(os.Stdout, "\x1b[%dA", drawnLines-1)
		}
		for i := 0; i < drawnLines; i++ {
			fmt.Fprintf(os.Stdout, "\r\x1b[2K") // clear entire line
			if i < drawnLines-1 {
				fmt.Fprintf(os.Stdout, "\n")
			}
		}
		// Move back up to start
		if drawnLines > 1 {
			fmt.Fprintf(os.Stdout, "\x1b[%dA", drawnLines-1)
		}

		// Filter items
		matches := filterItems(items, query)
		if cursor >= len(matches) {
			cursor = len(matches) - 1
		}
		if cursor < 0 {
			cursor = 0
		}

		// Determine visible slice
		show := matches
		if len(show) > maxVisible {
			show = show[:maxVisible]
		}

		// Draw input line
		fmt.Fprintf(os.Stdout, "\r\x1b[2K  > %s", query)

		// Draw items below
		for i, item := range show {
			if i == cursor {
				fmt.Fprintf(os.Stdout, "\r\n\x1b[2K  \x1b[7m %s \x1b[0m", item)
			} else {
				fmt.Fprintf(os.Stdout, "\r\n\x1b[2K    %s", item)
			}
		}

		// Show "more" indicator
		extra := 0
		if len(matches) > maxVisible {
			fmt.Fprintf(os.Stdout, "\r\n\x1b[2K    ... +%d more", len(matches)-maxVisible)
			extra = 1
		}

		totalLines := 1 + len(show) + extra // input + items + maybe "more"

		// Move cursor back up to input line
		linesBelow := len(show) + extra
		if linesBelow > 0 {
			fmt.Fprintf(os.Stdout, "\x1b[%dA", linesBelow)
		}
		// Position cursor at end of query
		fmt.Fprintf(os.Stdout, "\r\x1b[%dC", len(query)+4) // "  > " = 4 chars

		drawnLines = totalLines
	}

	redraw()

	buf := make([]byte, 3)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			clearDrawn(drawnLines)
			return "", false
		}

		// Escape sequences (arrows, etc)
		if n >= 1 && buf[0] == 27 {
			if n == 1 {
				// Plain Escape
				clearDrawn(drawnLines)
				return "", false
			}
			if n == 3 && buf[1] == 91 {
				matches := filterItems(items, query)
				visible := len(matches)
				if visible > maxVisible {
					visible = maxVisible
				}
				switch buf[2] {
				case 65: // Up
					if cursor > 0 {
						cursor--
					}
				case 66: // Down
					if cursor < visible-1 {
						cursor++
					}
				}
				redraw()
			}
			continue
		}

		switch {
		// Ctrl+C
		case buf[0] == 3:
			clearDrawn(drawnLines)
			return "", false

		// Enter
		case buf[0] == 13 || buf[0] == 10:
			matches := filterItems(items, query)
			clearDrawn(drawnLines)
			if len(matches) > 0 && cursor < len(matches) {
				return matches[cursor], true
			}
			return "", false

		// Backspace
		case buf[0] == 127 || buf[0] == 8:
			if len(query) > 0 {
				query = query[:len(query)-1]
				cursor = 0
			}
			redraw()

		// Tab
		case buf[0] == 9:
			matches := filterItems(items, query)
			clearDrawn(drawnLines)
			if len(matches) > 0 && cursor < len(matches) {
				return matches[cursor], true
			}

		// Printable ASCII
		case buf[0] >= 32 && buf[0] < 127:
			query += string(buf[0])
			cursor = 0
			redraw()
		}
	}
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

func clearDrawn(lines int) {
	// Move to first line
	if lines > 1 {
		fmt.Fprintf(os.Stdout, "\x1b[%dA", lines-1)
	}
	// Clear all lines
	for i := 0; i < lines; i++ {
		fmt.Fprintf(os.Stdout, "\r\x1b[2K")
		if i < lines-1 {
			fmt.Fprintf(os.Stdout, "\n")
		}
	}
	// Move back to first line
	if lines > 1 {
		fmt.Fprintf(os.Stdout, "\x1b[%dA", lines-1)
	}
	fmt.Fprintf(os.Stdout, "\r\x1b[2K")
}
