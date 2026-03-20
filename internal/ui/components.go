package ui

import (
	"fmt"
	"strings"
)

// RenderMenuList renders a selectable menu list with a cursor.
func RenderMenuList(items []string, cursor int) string {
	var sb strings.Builder
	for i, item := range items {
		if i == cursor {
			sb.WriteString("   " + StyleCyan.Render("▸") + " " + StyleWhiteBold.Render(item))
		} else {
			sb.WriteString("     " + item)
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// RenderPresetList renders a list of presets with cursor highlighting.
func RenderPresetList(items []string, cursor int) string {
	var sb strings.Builder
	for i, item := range items {
		if i == cursor {
			sb.WriteString("  " + StyleCyan.Render("▸") + " " + item)
		} else {
			sb.WriteString("    " + item)
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// RenderMultiSelect renders a checkbox list with cursor and toggle state.
func RenderMultiSelect(items []string, checked []bool, cursor int) string {
	var sb strings.Builder
	for i, item := range items {
		checkbox := "[ ]"
		if checked[i] {
			checkbox = StyleGreen.Render("[✓]")
		}
		if i == cursor {
			sb.WriteString("  " + StyleCyan.Render("▸") + " " + checkbox + " " + StyleWhiteBold.Render(item))
		} else {
			sb.WriteString("    " + checkbox + " " + item)
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// RenderBoxedMenu renders menu items inside a styled box.
func RenderBoxedMenu(items []string, cursor int) string {
	content := RenderMenuList(items, cursor)
	return StyleBox.Render(content)
}

// RenderCaptureInfo renders the "will capture" display for create flow.
func RenderCaptureInfo(claudeCount int, claudeSize int64, toolsCount int, toolsSize int64, hasClaudeMD bool) string {
	var sb strings.Builder
	sb.WriteString("  " + StyleYellowBold.Render("Will capture:") + "\n")
	if claudeCount > 0 {
		sb.WriteString(fmt.Sprintf("    %s  %s\n",
			StyleCyan.Render(".claude/"),
			StyleDim.Render(fmt.Sprintf("(%d files, %s)", claudeCount, FormatSize(claudeSize)))))
	}
	if toolsCount > 0 {
		sb.WriteString(fmt.Sprintf("    %s  %s\n",
			StyleCyan.Render("tools/  "),
			StyleDim.Render(fmt.Sprintf("(%d files, %s)", toolsCount, FormatSize(toolsSize)))))
	}
	if hasClaudeMD {
		sb.WriteString(fmt.Sprintf("    %s  %s\n",
			StyleCyan.Render("CLAUDE.md"),
			StyleDim.Render("(present)")))
	}
	return sb.String()
}
