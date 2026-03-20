package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/mryan/ccgears/internal/preset"
	"github.com/mryan/ccgears/internal/preview"
)

const version = "0.1.0"

// GetVersion returns the version string.
func GetVersion() string { return version }

// Lip Gloss styles
var (
	StyleCyan       = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	StyleMagBold    = lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Bold(true)
	StyleGreenBold  = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	StyleRedBold    = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
	StyleYellowBold = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true)
	StyleDim        = lipgloss.NewStyle().Faint(true)
	StyleBold       = lipgloss.NewStyle().Bold(true)
	StyleWhiteBold  = lipgloss.NewStyle().Bold(true)
	StyleBlue       = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	StyleGreen      = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	StyleYellow     = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	StyleRed        = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	StyleMagenta    = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	StyleCyanItalic = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Italic(true)
	StyleCyanBold   = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)

	StyleBox = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("4")).
			PaddingLeft(2).PaddingRight(2).
			PaddingTop(1).PaddingBottom(1)

	StylePreviewBox = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("4")).
			PaddingLeft(1).PaddingRight(1).
			PaddingTop(0).PaddingBottom(0)
)

// bannerArt is the figlet "Small" font rendered as a single block.
const bannerArt = `   ___  ___  ___
  / __|/ __|/ __|___ __ _ _ _ ___
 | (__| (__| (_ / -_) _` + "`" + ` | '_(_-<
  \___|\___|\___|___\__,_|_| /__/`

// Banner returns the styled banner string.
func Banner() string {
	var sb strings.Builder
	sb.WriteString(StyleCyan.Render(bannerArt))
	sb.WriteString("\n")
	sb.WriteString(StyleDim.Render("                              v" + version))
	sb.WriteString("\n")
	sb.WriteString("  " + StyleWhiteBold.Render("Claude Code Configuration Manager"))
	sb.WriteString("\n")
	return sb.String()
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

// RenderSuccess returns a formatted success message.
func RenderSuccess(msg string) string {
	return "  " + StyleGreenBold.Render("✓") + " " + msg
}

// RenderError returns a formatted error message.
func RenderError(msg string) string {
	return "  " + StyleRedBold.Render("✗") + " " + msg
}

// RenderWarn returns a formatted warning message.
func RenderWarn(msg string) string {
	return "  " + StyleYellowBold.Render("⚠") + " " + msg
}

// RenderInfoLine returns a label: value pair.
func RenderInfoLine(label, value string) string {
	return "    " + StyleCyan.Render(label) + " " + value
}

// RenderPreview returns the formatted preview of a preset.
func RenderPreview(name string, p *preview.PresetPreview) string {
	var sb strings.Builder

	if p.Headline != "" {
		sb.WriteString("  " + StyleCyan.Render("CLAUDE.md") + "  " + StyleCyanItalic.Render("\""+p.Headline+"\"") + "\n")
		sb.WriteString("\n")
	}

	if len(p.Skills) > 0 {
		sb.WriteString("  " + StyleYellowBold.Render(fmt.Sprintf("Skills (%d)", len(p.Skills))) + "\n")
		for _, s := range p.Skills {
			line := "    " + StyleMagenta.Render("●") + " " + StyleBold.Render(s.Name)
			if s.Description != "" {
				short := s.Description
				if len(short) > 40 {
					short = short[:37] + "..."
				}
				line += "  " + StyleDim.Render(short)
			}
			sb.WriteString(line + "\n")
		}
		sb.WriteString("\n")
	}

	hasPerm := p.BashPerms > 0 || len(p.MCPTools) > 0 || len(p.WebPerms) > 0
	if hasPerm {
		sb.WriteString("  " + StyleYellowBold.Render("Permissions") + "\n")
		if len(p.MCPTools) > 0 {
			sb.WriteString("    " + StyleBlue.Render("◆") + " " + StyleCyan.Render("MCP tools:") + fmt.Sprintf(" %d\n", len(p.MCPTools)))
		}
		if p.BashPerms > 0 {
			sb.WriteString("    " + StyleBlue.Render("◆") + " " + StyleCyan.Render("Bash:     ") + fmt.Sprintf(" %d patterns\n", p.BashPerms))
		}
		if len(p.WebPerms) > 0 {
			webStr := strings.Join(p.WebPerms, ", ")
			if len(webStr) > 24 {
				webStr = webStr[:21] + "..."
			}
			sb.WriteString("    " + StyleBlue.Render("◆") + " " + StyleCyan.Render("Web:      ") + " " + webStr + "\n")
		}
		sb.WriteString("\n")
	}

	if len(p.ToolScripts) > 0 {
		sb.WriteString("  " + StyleYellowBold.Render(fmt.Sprintf("Tools (%d)", len(p.ToolScripts))) + "\n")
		scriptStr := strings.Join(p.ToolScripts, ", ")
		if len(scriptStr) > 40 {
			scriptStr = scriptStr[:37] + "..."
		}
		sb.WriteString("    " + scriptStr + "\n")
		sb.WriteString("\n")
	}

	sb.WriteString("  " + StyleDim.Render(fmt.Sprintf("Total: %d files, %s", p.TotalFiles, FormatSize(p.TotalSize))))

	content := sb.String()
	title := StyleMagBold.Render("Preview: " + name)
	box := StylePreviewBox.Render(content)

	// Insert title into the top border
	lines := strings.Split(box, "\n")
	if len(lines) > 0 {
		// Replace part of the top border with the title
		border := lines[0]
		runes := []rune(border)
		if len(runes) > 4 {
			titleStr := " " + title + " "
			lines[0] = string(runes[:3]) + titleStr + string(runes[3+len([]rune(titleStr)):])
		}
		box = strings.Join(lines, "\n")
	}

	return box
}

// RenderDiff returns the formatted diff summary.
func RenderDiff(name string, diff *preset.DiffSummary) string {
	var sb strings.Builder

	sb.WriteString("  " + StyleYellowBold.Render("Changes to") + " " + StyleMagBold.Render(name) + ":\n\n")

	for _, f := range diff.Added {
		sb.WriteString(fmt.Sprintf("    %s %-40s %s\n", StyleGreen.Render("+"), f, StyleDim.Render("(added)")))
	}
	for _, f := range diff.Modified {
		sb.WriteString(fmt.Sprintf("    %s %-40s %s\n", StyleYellow.Render("~"), f, StyleDim.Render("(modified)")))
	}
	for _, f := range diff.Removed {
		sb.WriteString(fmt.Sprintf("    %s %-40s %s\n", StyleRed.Render("-"), f, StyleDim.Render("(removed)")))
	}
	if diff.Unchanged > 0 {
		sb.WriteString("\n    " + StyleDim.Render(fmt.Sprintf("%d unchanged", diff.Unchanged)))
	}

	return sb.String()
}
