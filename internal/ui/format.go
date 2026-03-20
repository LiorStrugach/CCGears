package ui

import (
	"fmt"
	"os"
	"strings"
	"time"
)

var useColor = os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != ""

const version = "0.1.0"

// GetVersion returns the version string.
func GetVersion() string { return version }

// ANSI codes
const (
	colorReset   = "\033[0m"
	colorBold    = "\033[1m"
	colorDim     = "\033[2m"
	colorItalic  = "\033[3m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorWhite   = "\033[37m"
)

func color(codes string, s string) string {
	if !useColor {
		return s
	}
	return codes + s + colorReset
}

func bold(s string) string       { return color(colorBold, s) }
func dim(s string) string        { return color(colorDim, s) }
func italic(s string) string     { return color(colorItalic, s) }
func red(s string) string        { return color(colorRed, s) }
func green(s string) string      { return color(colorGreen, s) }
func yellow(s string) string     { return color(colorYellow, s) }
func blue(s string) string       { return color(colorBlue, s) }
func magenta(s string) string    { return color(colorMagenta, s) }
func cyan(s string) string       { return color(colorCyan, s) }
func redBold(s string) string    { return color(colorRed+colorBold, s) }
func greenBold(s string) string  { return color(colorGreen+colorBold, s) }
func yellowBold(s string) string { return color(colorYellow+colorBold, s) }
func cyanBold(s string) string   { return color(colorCyan+colorBold, s) }
func magBold(s string) string    { return color(colorMagenta+colorBold, s) }
func whiteBold(s string) string  { return color(colorWhite+colorBold, s) }
func cyanItalic(s string) string { return color(colorCyan+colorItalic, s) }

// Banner ASCII art lines (figlet "Small" font)
var bannerLines = []string{
	`   ___  ___  ___                   `,
	`  / __|/ __|/ __|___ __ _ _ _ ___  `,
	` | (__| (__| (_ / -_) _` + "`" + ` | '_(_-< `,
	`  \___|\___|\___|___\__,_|_| /__/  `,
}

// Banner returns the full styled banner with version and tagline.
func Banner() string {
	var sb strings.Builder
	sb.WriteString("\n")
	for _, line := range bannerLines {
		sb.WriteString(cyan(line))
		sb.WriteString("\n")
	}
	sb.WriteString(dim("                              v" + version))
	sb.WriteString("\n")
	sb.WriteString("  " + whiteBold("Claude Code Configuration Manager"))
	sb.WriteString("\n")
	return sb.String()
}

// BannerPlain returns the banner without ANSI codes (for piped output).
func BannerPlain() string {
	var sb strings.Builder
	sb.WriteString("\n")
	for _, line := range bannerLines {
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	sb.WriteString("                              v" + version + "\n")
	sb.WriteString("  Claude Code Configuration Manager\n")
	return sb.String()
}

// Box-drawing characters
const (
	boxTL = "┌"
	boxTR = "┐"
	boxBL = "└"
	boxBR = "┘"
	boxH  = "─"
	boxV  = "│"
	boxHT = "─" // horizontal with title separator
)

// BoxTop returns a top border with an optional title.
func BoxTop(title string, width int) string {
	if title == "" {
		return blue(boxTL + strings.Repeat(boxH, width-2) + boxTR)
	}
	titleStr := " " + title + " "
	remaining := width - 2 - len(titleStr) - 1 // -1 for initial dash
	if remaining < 0 {
		remaining = 0
	}
	return blue(boxTL+boxH) + titleStr + blue(strings.Repeat(boxH, remaining)+boxTR)
}

// BoxBottom returns a bottom border.
func BoxBottom(width int) string {
	return blue(boxBL + strings.Repeat(boxH, width-2) + boxBR)
}

// BoxLine returns content padded inside box borders.
func BoxLine(content string, width int, contentVisualLen int) string {
	padding := width - 2 - contentVisualLen
	if padding < 0 {
		padding = 0
	}
	return blue(boxV) + content + strings.Repeat(" ", padding) + blue(boxV)
}

// BoxEmpty returns an empty line inside box borders.
func BoxEmpty(width int) string {
	return blue(boxV) + strings.Repeat(" ", width-2) + blue(boxV)
}

// FormatSize returns a human-readable file size.
func FormatSize(bytes int64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)
	switch {
	case bytes >= gb:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(gb))
	case bytes >= mb:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(mb))
	case bytes >= kb:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(kb))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// FormatDate formats an RFC3339 string to a friendly display.
func FormatDate(rfc3339 string) string {
	t, err := time.Parse(time.RFC3339, rfc3339)
	if err != nil {
		return rfc3339
	}
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		m := int(diff.Minutes())
		if m == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", m)
	case diff < 24*time.Hour:
		h := int(diff.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	case diff < 7*24*time.Hour:
		d := int(diff.Hours() / 24)
		if d == 1 {
			return "yesterday"
		}
		return fmt.Sprintf("%d days ago", d)
	default:
		return t.Format("Jan 2, 2006")
	}
}

// PrintHeader prints a styled section header.
func PrintHeader(title string) {
	println()
	println(yellowBold(title))
	println(blue(strings.Repeat("─", len(title))))
}

// Success prints a green success message.
func Success(msg string) {
	printf("  %s %s\n", greenBold("✓"), msg)
}

// Error prints a red error message.
func Error(msg string) {
	printf("  %s %s\n", redBold("✗"), msg)
}

// Warn prints a yellow warning message.
func Warn(msg string) {
	printf("  %s %s\n", yellowBold("⚠"), msg)
}

// InfoLine prints a label: value pair with colored label.
func InfoLine(label, value string) {
	printf("    %s %s\n", cyan(label), value)
}
