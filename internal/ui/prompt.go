package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	reader *bufio.Reader
	input  io.Reader = os.Stdin
	output io.Writer = os.Stdout
)

func getReader() *bufio.Reader {
	if reader == nil {
		reader = bufio.NewReader(input)
	}
	return reader
}

// SetIO overrides input/output for testing.
func SetIO(in io.Reader, out io.Writer) {
	input = in
	output = out
	reader = bufio.NewReader(in)
}

// ResetIO restores default stdin/stdout.
func ResetIO() {
	input = os.Stdin
	output = os.Stdout
	reader = nil
}

func printf(format string, a ...interface{}) {
	fmt.Fprintf(output, format, a...)
}

func println(a ...interface{}) {
	fmt.Fprintln(output, a...)
}

// hideCursor hides the terminal cursor.
func hideCursor() {
	if useColor {
		printf("\033[?25l")
	}
}

// showCursor shows the terminal cursor.
func showCursor() {
	if useColor {
		printf("\033[?25h")
	}
}

// ShowCursorPublic is the exported version for signal handlers.
func ShowCursorPublic() { showCursor() }

// moveCursorUp moves the cursor up n lines.
func moveCursorUp(n int) {
	if useColor && n > 0 {
		printf("\033[%dA", n)
	}
}

// clearLine clears the current line.
func clearLine() {
	if useColor {
		printf("\033[2K\r")
	}
}

// SelectMenu displays items and lets the user navigate with arrow keys.
// Returns the 0-based index of the selected item, or -1 if cancelled (Escape/q).
func SelectMenu(items []string, initialIdx int) int {
	if len(items) == 0 {
		return -1
	}

	selected := initialIdx
	if selected < 0 || selected >= len(items) {
		selected = 0
	}

	wasRaw := IsRawMode()
	if !wasRaw {
		EnableRawMode()
	}

	hideCursor()

	// Initial render
	renderMenuItems(items, selected)

	for {
		key, err := ReadKey()
		if err != nil {
			break
		}

		switch key {
		case KeyUp, 'k':
			if selected > 0 {
				selected--
			}
		case KeyDown, 'j':
			if selected < len(items)-1 {
				selected++
			}
		case KeyEnter, KeyNewline:
			// Move below the menu
			showCursor()
			printf("\n")
			if !wasRaw {
				DisableRawMode()
			}
			return selected
		case KeyEscape, 'q':
			showCursor()
			printf("\n")
			if !wasRaw {
				DisableRawMode()
			}
			return -1
		default:
			continue
		}

		// Redraw: move cursor up to first item, redraw all
		moveCursorUp(len(items))
		renderMenuItems(items, selected)
	}

	showCursor()
	if !wasRaw {
		DisableRawMode()
	}
	return selected
}

func renderMenuItems(items []string, selected int) {
	for i, item := range items {
		clearLine()
		if i == selected {
			printf("    %s %s\n", cyan("▸"), whiteBold(item))
		} else {
			printf("      %s\n", item)
		}
	}
}

// SelectMulti displays items with checkboxes and lets the user toggle selections.
// Returns indices of checked items, or nil if cancelled.
func SelectMulti(items []string, checked []bool) []int {
	if len(items) == 0 {
		return nil
	}

	selected := 0

	wasRaw := IsRawMode()
	if !wasRaw {
		EnableRawMode()
	}

	hideCursor()

	// Initial render
	renderMultiItems(items, checked, selected)

	for {
		key, err := ReadKey()
		if err != nil {
			break
		}

		switch key {
		case KeyUp, 'k':
			if selected > 0 {
				selected--
			}
		case KeyDown, 'j':
			if selected < len(items)-1 {
				selected++
			}
		case ' ':
			checked[selected] = !checked[selected]
		case 'a':
			for i := range checked {
				checked[i] = true
			}
		case 'n':
			for i := range checked {
				checked[i] = false
			}
		case KeyEnter, KeyNewline:
			showCursor()
			printf("\n")
			if !wasRaw {
				DisableRawMode()
			}
			var result []int
			for i, c := range checked {
				if c {
					result = append(result, i)
				}
			}
			return result
		case KeyEscape, 'q':
			showCursor()
			printf("\n")
			if !wasRaw {
				DisableRawMode()
			}
			return nil
		default:
			continue
		}

		moveCursorUp(len(items))
		renderMultiItems(items, checked, selected)
	}

	showCursor()
	if !wasRaw {
		DisableRawMode()
	}
	return nil
}

func renderMultiItems(items []string, checked []bool, selected int) {
	for i, item := range items {
		clearLine()
		checkbox := "[ ]"
		if checked[i] {
			checkbox = green("[✓]")
		}
		if i == selected {
			printf("  %s %s %s\n", cyan("▸"), checkbox, whiteBold(item))
		} else {
			printf("    %s %s\n", checkbox, item)
		}
	}
}

// SelectMenuBoxed renders a menu inside a box and lets the user navigate.
// Returns the 0-based index or -1 if cancelled.
func SelectMenuBoxed(items []string, width int) int {
	if len(items) == 0 {
		return -1
	}

	selected := 0

	wasRaw := IsRawMode()
	if !wasRaw {
		EnableRawMode()
	}

	hideCursor()

	// Initial render
	renderBoxedMenu(items, selected, width)

	for {
		key, err := ReadKey()
		if err != nil {
			break
		}

		switch key {
		case KeyUp, 'k':
			if selected > 0 {
				selected--
			}
		case KeyDown, 'j':
			if selected < len(items)-1 {
				selected++
			}
		case KeyEnter, KeyNewline:
			showCursor()
			printf("\n")
			if !wasRaw {
				DisableRawMode()
			}
			return selected
		case KeyEscape, 'q':
			showCursor()
			printf("\n")
			if !wasRaw {
				DisableRawMode()
			}
			return -1
		default:
			continue
		}

		// Move up: items + 2 (top border + empty line) + items + empty line + bottom border
		totalLines := len(items) + 4 // top + empty + items + empty + bottom
		moveCursorUp(totalLines)
		renderBoxedMenu(items, selected, width)
	}

	showCursor()
	if !wasRaw {
		DisableRawMode()
	}
	return selected
}

func renderBoxedMenu(items []string, selected int, width int) {
	clearLine()
	printf("  %s\n", BoxTop("", width))
	clearLine()
	printf("  %s\n", BoxEmpty(width))
	for i, item := range items {
		clearLine()
		var content string
		var visLen int
		if i == selected {
			content = fmt.Sprintf("   %s %s", cyan("▸"), whiteBold(item))
			visLen = 3 + 2 + len(item)
		} else {
			content = "     " + item
			visLen = 5 + len(item)
		}
		printf("  %s\n", BoxLine(content, width, visLen))
	}
	clearLine()
	printf("  %s\n", BoxEmpty(width))
	clearLine()
	printf("  %s\n", BoxBottom(width))
}

// ReadLine temporarily exits raw mode, prompts for input, and returns the trimmed response.
func ReadLine(prompt string) string {
	wasRaw := IsRawMode()
	if wasRaw {
		DisableRawMode()
		// Reset the bufio reader since we changed terminal mode
		reader = nil
	}

	showCursor()
	printf("  %s %s", cyan("❯"), prompt)
	line, _ := getReader().ReadString('\n')

	if wasRaw {
		EnableRawMode()
		reader = nil
	}

	return strings.TrimSpace(line)
}

// Confirm asks a y/N question. Returns true on "y" or "yes".
func Confirm(prompt string) bool {
	response := ReadLine(fmt.Sprintf("%s %s ", prompt, dim("[y/N]")))
	response = strings.ToLower(response)
	return response == "y" || response == "yes"
}

// ConfirmByName asks the user to type a name to confirm. Returns true if it matches.
func ConfirmByName(name string) bool {
	response := ReadLine(fmt.Sprintf("Type %s to confirm: ", magBold(name)))
	return response == name
}

// PromptText asks for text input with an optional default.
func PromptText(label, defaultVal string) string {
	prompt := label
	if defaultVal != "" {
		prompt += fmt.Sprintf(" %s", dim("["+defaultVal+"]"))
	}
	prompt += ": "
	response := ReadLine(prompt)
	if response == "" {
		return defaultVal
	}
	return response
}
