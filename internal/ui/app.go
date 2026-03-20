package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mryan/ccgears/internal/config"
	"github.com/mryan/ccgears/internal/preset"
	"github.com/mryan/ccgears/internal/preview"
	"github.com/mryan/ccgears/internal/snapshot"
)

// appState represents the current UI state.
type appState int

const (
	stateMainMenu appState = iota
	stateSelectPreset
	statePreview
	stateConfirm
	stateTextInput
	stateScanResults
	stateMessage
	stateConfirmLaunch
)

// flowType identifies which flow is active.
type flowType int

const (
	flowNone flowType = iota
	flowLoad
	flowSave
	flowCreate
	flowScan
	flowDelete
	flowUndo
)

var menuLabels = []string{
	"Load preset",
	"Save preset",
	"Create preset",
	"Scan & import",
	"Delete preset",
	"Undo last load",
	"Exit",
}

// App is the main Bubble Tea model.
type App struct {
	cfg        *config.Config
	projectDir string
	state      appState
	flow       flowType

	// Menu / list cursor
	menuCursor   int
	presetCursor int
	presetList   []*preset.Meta

	// Scan
	scanResults []preset.ScanResult
	scanChecked []bool
	scanCursor  int

	// Text input
	textInput   textinput.Model
	inputTarget string // "presetName", "presetDesc", "scanRoot", "confirmName"
	inputValues map[string]string

	// Confirm
	confirmMsg string

	// Preview / diff content
	previewContent string
	selectedName   string

	// Messages
	message string

	// Post-exit
	shouldLaunch bool

	width, height int
}

// NewApp creates a new App model.
func NewApp(cfg *config.Config, projectDir string) App {
	ti := textinput.New()
	ti.CharLimit = 156
	ti.Width = 50

	return App{
		cfg:         cfg,
		projectDir:  projectDir,
		state:       stateMainMenu,
		textInput:   ti,
		inputValues: make(map[string]string),
	}
}

// ShouldLaunchClaude returns true if claude should be launched after the program exits.
func (m App) ShouldLaunchClaude() bool {
	return m.shouldLaunch
}

func (m App) Init() tea.Cmd {
	return nil
}

// Update handles all input events.
func (m App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		// Global quit on ctrl+c
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	}

	switch m.state {
	case stateMainMenu:
		return m.updateMainMenu(msg)
	case stateSelectPreset:
		return m.updateSelectPreset(msg)
	case statePreview:
		return m.updatePreview(msg)
	case stateConfirm:
		return m.updateConfirm(msg)
	case stateTextInput:
		return m.updateTextInput(msg)
	case stateScanResults:
		return m.updateScanResults(msg)
	case stateMessage:
		return m.updateMessage(msg)
	case stateConfirmLaunch:
		return m.updateConfirmLaunch(msg)
	}

	return m, nil
}

// View renders the current state.
func (m App) View() string {
	switch m.state {
	case stateMainMenu:
		return m.viewMainMenu()
	case stateSelectPreset:
		return m.viewSelectPreset()
	case statePreview:
		return m.viewPreview()
	case stateConfirm:
		return m.viewConfirm()
	case stateTextInput:
		return m.viewTextInput()
	case stateScanResults:
		return m.viewScanResults()
	case stateMessage:
		return m.viewMessage()
	case stateConfirmLaunch:
		return m.viewConfirmLaunch()
	}
	return ""
}

// --- Main Menu ---

func (m App) updateMainMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.menuCursor > 0 {
				m.menuCursor--
			}
		case "down", "j":
			if m.menuCursor < len(menuLabels)-1 {
				m.menuCursor++
			}
		case "enter":
			return m.handleMenuSelect()
		case "q", "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m App) handleMenuSelect() (tea.Model, tea.Cmd) {
	switch m.menuCursor {
	case 0: // Load
		return m.startPresetSelect(flowLoad)
	case 1: // Save
		return m.startPresetSelect(flowSave)
	case 2: // Create
		return m.startCreateFlow()
	case 3: // Scan
		return m.startScanFlow()
	case 4: // Delete
		return m.startPresetSelect(flowDelete)
	case 5: // Undo
		return m.executeUndo()
	case 6: // Exit
		return m, tea.Quit
	}
	return m, nil
}

func (m App) viewMainMenu() string {
	var sb strings.Builder
	sb.WriteString(Banner())
	sb.WriteString("\n")
	sb.WriteString("  " + StyleDim.Render("↑↓ navigate  Enter select  q quit"))
	sb.WriteString("\n\n")
	sb.WriteString(RenderBoxedMenu(menuLabels, m.menuCursor))
	sb.WriteString("\n")
	return sb.String()
}

// --- Preset Selection ---

func (m App) startPresetSelect(f flowType) (tea.Model, tea.Cmd) {
	presets, err := preset.List(m.cfg)
	if err != nil {
		m.message = RenderError(fmt.Sprintf("%v", err))
		m.state = stateMessage
		return m, nil
	}
	if len(presets) == 0 {
		m.message = RenderWarn("No presets found. Create one first.")
		m.state = stateMessage
		return m, nil
	}
	m.presetList = presets
	m.presetCursor = 0
	m.flow = f
	m.state = stateSelectPreset
	return m, nil
}

func (m App) updateSelectPreset(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.presetCursor > 0 {
				m.presetCursor--
			}
		case "down", "j":
			if m.presetCursor < len(m.presetList)-1 {
				m.presetCursor++
			}
		case "enter":
			m.selectedName = m.presetList[m.presetCursor].Name
			return m.handlePresetSelected()
		case "q", "esc":
			m.state = stateMainMenu
			m.flow = flowNone
			return m, nil
		}
	}
	return m, nil
}

func (m App) handlePresetSelected() (tea.Model, tea.Cmd) {
	switch m.flow {
	case flowLoad:
		// Show preview
		presetDir := preset.Dir(m.cfg, m.selectedName)
		p := preview.Build(presetDir)
		m.previewContent = RenderPreview(m.selectedName, p)
		m.state = statePreview
	case flowSave:
		// Show diff
		presetDir := preset.Dir(m.cfg, m.selectedName)
		diff := preset.ComputeDiff(presetDir, m.projectDir)
		if !diff.HasChanges() {
			m.message = RenderWarn("No changes detected. Nothing to save.")
			m.state = stateMessage
			return m, nil
		}
		m.previewContent = RenderDiff(m.selectedName, diff)
		m.confirmMsg = fmt.Sprintf("Save these changes to %s?", StyleMagBold.Render(m.selectedName))
		m.state = stateConfirm
	case flowDelete:
		// Confirm by typing name
		m.inputTarget = "confirmName"
		m.textInput.SetValue("")
		m.textInput.Placeholder = m.selectedName
		m.textInput.Focus()
		m.state = stateTextInput
	default:
		m.state = stateMainMenu
	}
	return m, nil
}

func (m App) viewSelectPreset() string {
	var sb strings.Builder
	sb.WriteString(Banner())
	sb.WriteString("\n")

	action := "Select"
	switch m.flow {
	case flowLoad:
		action = "Load"
	case flowSave:
		action = "Save to"
	case flowDelete:
		action = "Delete"
	default:
		// other flows don't use preset selection
	}
	sb.WriteString("  " + StyleYellowBold.Render(action+" Preset") + "\n")
	sb.WriteString("  " + StyleDim.Render("↑↓ navigate  Enter select  q cancel") + "\n\n")

	items := make([]string, len(m.presetList))
	for i, p := range m.presetList {
		desc := StyleDim.Render("(no description)")
		if p.Description != "" {
			desc = p.Description
		}
		lastUsed := ""
		if p.LastUsed != "" {
			lastUsed = "  " + StyleDim.Render(FormatDate(p.LastUsed))
		}
		items[i] = fmt.Sprintf("%-18s %s%s", StyleMagBold.Render(p.Name), desc, lastUsed)
	}
	sb.WriteString(RenderPresetList(items, m.presetCursor))

	return sb.String()
}

// --- Preview ---

func (m App) updatePreview(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "y":
			return m.executeLoad()
		case "q", "esc", "n":
			m.state = stateMainMenu
			m.flow = flowNone
			return m, nil
		}
	}
	return m, nil
}

func (m App) viewPreview() string {
	var sb strings.Builder
	sb.WriteString(Banner())
	sb.WriteString("\n")
	sb.WriteString(m.previewContent)
	sb.WriteString("\n\n")
	sb.WriteString(fmt.Sprintf("  Load %s? Current state will be backed up.", StyleMagBold.Render(m.selectedName)))
	sb.WriteString("\n")
	sb.WriteString("  " + StyleDim.Render("Enter=load  q=cancel"))
	sb.WriteString("\n")
	return sb.String()
}

// --- Confirm ---

func (m App) updateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "enter":
			return m.executeConfirmedAction()
		case "n", "q", "esc":
			m.message = RenderWarn("Cancelled.")
			m.state = stateMessage
			return m, nil
		}
	}
	return m, nil
}

func (m App) executeConfirmedAction() (tea.Model, tea.Cmd) {
	switch m.flow {
	case flowSave:
		count, err := preset.Save(m.cfg, m.selectedName, m.projectDir)
		if err != nil {
			m.message = RenderError(fmt.Sprintf("%v", err))
		} else {
			m.message = RenderSuccess(fmt.Sprintf("Preset %s saved %s",
				StyleMagBold.Render(m.selectedName),
				StyleDim.Render(fmt.Sprintf("(%d files)", count))))
		}
		m.state = stateMessage
		m.flow = flowNone
	case flowCreate:
		name := m.inputValues["presetName"]
		desc := m.inputValues["presetDesc"]
		meta, count, err := preset.Create(m.cfg, name, desc, m.projectDir)
		if err != nil {
			m.message = RenderError(fmt.Sprintf("%v", err))
		} else {
			m.message = RenderSuccess(fmt.Sprintf("Preset %s created %s",
				StyleMagBold.Render(meta.Name),
				StyleDim.Render(fmt.Sprintf("(%d files)", count))))
		}
		m.state = stateMessage
		m.flow = flowNone
	case flowUndo:
		dir, err := snapshot.RestoreBackup(m.cfg.BackupPath)
		if err != nil {
			m.message = RenderError(fmt.Sprintf("%v", err))
		} else {
			m.message = RenderSuccess(fmt.Sprintf("Previous state restored to %s", dir))
		}
		m.state = stateMessage
		m.flow = flowNone
	default:
		m.state = stateMainMenu
		m.flow = flowNone
	}
	return m, nil
}

func (m App) viewConfirm() string {
	var sb strings.Builder
	sb.WriteString(Banner())
	sb.WriteString("\n")
	if m.previewContent != "" {
		sb.WriteString(m.previewContent)
		sb.WriteString("\n\n")
	}
	sb.WriteString("  " + m.confirmMsg)
	sb.WriteString("\n")
	sb.WriteString("  " + StyleDim.Render("y=yes  n=cancel"))
	sb.WriteString("\n")
	return sb.String()
}

// --- Text Input ---

func (m App) updateTextInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			m.state = stateMainMenu
			m.flow = flowNone
			return m, nil
		case tea.KeyEnter:
			value := m.textInput.Value()
			return m.handleTextInputSubmit(value)
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m App) handleTextInputSubmit(value string) (tea.Model, tea.Cmd) {
	switch m.inputTarget {
	case "presetName":
		if value == "" {
			m.message = RenderWarn("Cancelled.")
			m.state = stateMessage
			m.flow = flowNone
			return m, nil
		}
		m.inputValues["presetName"] = value
		// Now ask for description
		m.inputTarget = "presetDesc"
		m.textInput.SetValue("")
		m.textInput.Placeholder = "(optional)"
		m.textInput.Focus()
		return m, textinput.Blink

	case "presetDesc":
		m.inputValues["presetDesc"] = value
		// Show capture info and confirm
		return m.showCreateConfirm()

	case "scanRoot":
		if value == "" {
			m.state = stateMainMenu
			m.flow = flowNone
			return m, nil
		}
		return m.executeScan(value)

	case "confirmName":
		if value == m.selectedName {
			err := preset.Delete(m.cfg, m.selectedName)
			if err != nil {
				m.message = RenderError(fmt.Sprintf("%v", err))
			} else {
				m.message = RenderSuccess(fmt.Sprintf("Preset %s deleted.", StyleMagBold.Render(m.selectedName)))
			}
		} else {
			m.message = RenderWarn("Name does not match. Delete cancelled.")
		}
		m.state = stateMessage
		m.flow = flowNone
		return m, nil
	}
	return m, nil
}

func (m App) viewTextInput() string {
	var sb strings.Builder
	sb.WriteString(Banner())
	sb.WriteString("\n")

	switch m.inputTarget {
	case "presetName":
		sb.WriteString("  " + StyleYellowBold.Render("Create Preset") + "\n\n")
		sb.WriteString("  Preset name " + StyleDim.Render("(lowercase, hyphens, max 48 chars)") + "\n")
	case "presetDesc":
		sb.WriteString("  " + StyleYellowBold.Render("Create Preset") + "\n\n")
		sb.WriteString("  Name: " + StyleMagBold.Render(m.inputValues["presetName"]) + "\n")
		sb.WriteString("  Description " + StyleDim.Render("(optional, press Enter to skip)") + "\n")
	case "scanRoot":
		sb.WriteString("  " + StyleYellowBold.Render("Scan & Import") + "\n\n")
		sb.WriteString("  Root directory to scan\n")
	case "confirmName":
		sb.WriteString("  " + StyleYellowBold.Render("Delete Preset") + "\n\n")
		meta, _ := preset.Get(m.cfg, m.selectedName)
		if meta != nil {
			sb.WriteString(RenderInfoLine("Name:       ", StyleMagBold.Render(meta.Name)) + "\n")
			if meta.Description != "" {
				sb.WriteString(RenderInfoLine("Description:", meta.Description) + "\n")
			}
			sb.WriteString(RenderInfoLine("Created:    ", StyleDim.Render(FormatDate(meta.Created))) + "\n")
		}
		sb.WriteString("\n")
		sb.WriteString("  " + StyleRedBold.Render("This cannot be undone.") + "\n\n")
		sb.WriteString("  Type the preset name to confirm\n")
	}

	sb.WriteString("\n  " + StyleCyan.Render("❯") + " " + m.textInput.View())
	sb.WriteString("\n\n")
	sb.WriteString("  " + StyleDim.Render("Enter=submit  Esc=cancel"))
	sb.WriteString("\n")
	return sb.String()
}

// --- Create Flow ---

func (m App) startCreateFlow() (tea.Model, tea.Cmd) {
	m.flow = flowCreate
	m.inputTarget = "presetName"
	m.inputValues = make(map[string]string)
	m.textInput.SetValue("")
	m.textInput.Placeholder = "my-preset"
	m.textInput.Focus()
	m.state = stateTextInput
	return m, textinput.Blink
}

func (m App) showCreateConfirm() (tea.Model, tea.Cmd) {
	claudeDir := filepath.Join(m.projectDir, ".claude")
	toolsDir := filepath.Join(m.projectDir, "tools")
	claudeMD := filepath.Join(m.projectDir, "CLAUDE.md")

	var claudeCount int
	var claudeSize int64
	var toolsCount int
	var toolsSize int64
	hasClaudeMD := false

	if fi, err := os.Stat(claudeDir); err == nil && fi.IsDir() {
		claudeCount, claudeSize = snapshot.CountFiles(claudeDir, snapshot.DefaultExcludes)
	}
	if fi, err := os.Stat(toolsDir); err == nil && fi.IsDir() {
		toolsCount, toolsSize = snapshot.CountFiles(toolsDir, snapshot.DefaultExcludes)
	}
	if fi, err := os.Stat(claudeMD); err == nil && !fi.IsDir() {
		hasClaudeMD = true
	}

	m.previewContent = RenderCaptureInfo(claudeCount, claudeSize, toolsCount, toolsSize, hasClaudeMD)
	m.confirmMsg = fmt.Sprintf("Create preset %s?", StyleMagBold.Render(m.inputValues["presetName"]))
	m.state = stateConfirm
	return m, nil
}

// --- Load & Launch ---

func (m App) executeLoad() (tea.Model, tea.Cmd) {
	count, err := preset.Load(m.cfg, m.selectedName, m.projectDir)
	if err != nil {
		m.message = RenderError(fmt.Sprintf("%v", err))
		m.state = stateMessage
		m.flow = flowNone
		return m, nil
	}
	m.message = RenderSuccess(fmt.Sprintf("Loaded %s %s",
		StyleMagBold.Render(m.selectedName),
		StyleDim.Render(fmt.Sprintf("(%d files)", count))))
	m.state = stateConfirmLaunch
	return m, nil
}

func (m App) updateConfirmLaunch(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.shouldLaunch = true
			return m, tea.Quit
		case "q", "esc":
			m.state = stateMainMenu
			m.flow = flowNone
			return m, nil
		}
	}
	return m, nil
}

func (m App) viewConfirmLaunch() string {
	var sb strings.Builder
	sb.WriteString(Banner())
	sb.WriteString("\n")
	sb.WriteString(m.message + "\n\n")
	sb.WriteString("  " + StyleCyanBold.Render("Now opening a new Claude session...") + "\n")
	sb.WriteString("  " + StyleDim.Render("Press Enter to continue, q/Esc to cancel and return to CCGears"))
	sb.WriteString("\n")
	return sb.String()
}

// --- Scan Flow ---

func (m App) startScanFlow() (tea.Model, tea.Cmd) {
	m.flow = flowScan
	m.inputTarget = "scanRoot"

	home, _ := os.UserHomeDir()
	defaultRoot := home
	if defaultRoot == "" {
		defaultRoot = "/"
	}

	m.textInput.SetValue("")
	m.textInput.Placeholder = defaultRoot
	m.textInput.Focus()
	m.state = stateTextInput
	return m, textinput.Blink
}

func (m App) executeScan(rootDir string) (tea.Model, tea.Cmd) {
	home, _ := os.UserHomeDir()
	if strings.HasPrefix(rootDir, "~/") {
		rootDir = filepath.Join(home, rootDir[2:])
	} else if rootDir == "~" {
		rootDir = home
	}

	results, err := preset.ScanForProjects(rootDir, 4)
	if err != nil {
		m.message = RenderError(fmt.Sprintf("Scan failed: %v", err))
		m.state = stateMessage
		m.flow = flowNone
		return m, nil
	}
	if len(results) == 0 {
		m.message = RenderWarn("No projects with Claude Code configs found.")
		m.state = stateMessage
		m.flow = flowNone
		return m, nil
	}

	m.scanResults = results
	m.scanChecked = make([]bool, len(results))
	for i := range m.scanChecked {
		m.scanChecked[i] = true
	}
	m.scanCursor = 0
	m.state = stateScanResults
	return m, nil
}

func (m App) updateScanResults(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.scanCursor > 0 {
				m.scanCursor--
			}
		case "down", "j":
			if m.scanCursor < len(m.scanResults)-1 {
				m.scanCursor++
			}
		case " ":
			m.scanChecked[m.scanCursor] = !m.scanChecked[m.scanCursor]
		case "a":
			for i := range m.scanChecked {
				m.scanChecked[i] = true
			}
		case "n":
			for i := range m.scanChecked {
				m.scanChecked[i] = false
			}
		case "enter":
			return m.executeScanImport()
		case "q", "esc":
			m.state = stateMainMenu
			m.flow = flowNone
			return m, nil
		}
	}
	return m, nil
}

func (m App) executeScanImport() (tea.Model, tea.Cmd) {
	var selected []int
	for i, c := range m.scanChecked {
		if c {
			selected = append(selected, i)
		}
	}
	if len(selected) == 0 {
		m.message = RenderWarn("No projects selected.")
		m.state = stateMessage
		m.flow = flowNone
		return m, nil
	}

	var sb strings.Builder
	imported := 0
	skipped := 0

	for _, idx := range selected {
		r := m.scanResults[idx]
		baseName := preset.AutoPresetName(r.DirName)
		name := preset.UniquePresetName(m.cfg, baseName)

		if name != baseName {
			if _, err := preset.Get(m.cfg, baseName); err == nil {
				sb.WriteString(RenderWarn(fmt.Sprintf("%s already exists, skipping", StyleMagBold.Render(baseName))) + "\n")
				skipped++
				continue
			}
		}

		desc := fmt.Sprintf("Imported from %s", preset.ShortenPath(r.ProjectDir))
		_, count, err := preset.Create(m.cfg, name, desc, r.ProjectDir)
		if err != nil {
			sb.WriteString(RenderError(fmt.Sprintf("%s: %v", r.DirName, err)) + "\n")
			skipped++
			continue
		}
		sb.WriteString(RenderSuccess(fmt.Sprintf("%s %s",
			StyleMagBold.Render(name),
			StyleDim.Render(fmt.Sprintf("(%d files)", count)))) + "\n")
		imported++
	}

	sb.WriteString("\n")
	if imported > 0 {
		sb.WriteString(RenderSuccess(fmt.Sprintf("Imported %d of %d presets.", imported, len(selected))))
	}
	if skipped > 0 {
		sb.WriteString("\n  " + StyleDim.Render(fmt.Sprintf("%d skipped.", skipped)))
	}

	m.message = sb.String()
	m.state = stateMessage
	m.flow = flowNone
	return m, nil
}

func (m App) viewScanResults() string {
	var sb strings.Builder
	sb.WriteString(Banner())
	sb.WriteString("\n")
	sb.WriteString("  " + StyleYellowBold.Render("Scan & Import") + "\n")
	sb.WriteString(fmt.Sprintf("  Found %s projects with Claude Code configs.\n\n",
		StyleCyanBold.Render(fmt.Sprintf("%d", len(m.scanResults)))))
	sb.WriteString("  " + StyleDim.Render("↑↓ navigate  Space toggle  a=all  n=none  Enter confirm  q cancel") + "\n\n")

	items := make([]string, len(m.scanResults))
	for i, r := range m.scanResults {
		shortPath := preset.ShortenPath(r.ProjectDir)
		var parts []string
		if r.SkillCount > 0 {
			parts = append(parts, fmt.Sprintf("%d skills", r.SkillCount))
		}
		if r.ToolCount > 0 {
			parts = append(parts, fmt.Sprintf("%d tools", r.ToolCount))
		}
		if r.HasClaudeMD {
			parts = append(parts, "CLAUDE.md")
		}
		info := ""
		if len(parts) > 0 {
			info = "  " + StyleDim.Render(strings.Join(parts, ", "))
		}
		items[i] = fmt.Sprintf("%-18s %s%s", StyleMagBold.Render(r.DirName), StyleDim.Render(shortPath), info)
	}

	sb.WriteString(RenderMultiSelect(items, m.scanChecked, m.scanCursor))
	return sb.String()
}

// --- Undo ---

func (m App) executeUndo() (tea.Model, tea.Cmd) {
	if !snapshot.HasBackup(m.cfg.BackupPath) {
		m.message = RenderWarn("No backup available. Nothing to undo.")
		m.state = stateMessage
		return m, nil
	}

	info, err := snapshot.GetBackupInfo(m.cfg.BackupPath)
	if err != nil {
		m.message = RenderError(fmt.Sprintf("%v", err))
		m.state = stateMessage
		return m, nil
	}

	// Build info display and set up confirm
	var sb strings.Builder
	sb.WriteString("  " + StyleYellowBold.Render("Undo Last Load") + "\n\n")
	sb.WriteString(RenderInfoLine("Preset:   ", StyleMagBold.Render(info.PresetLoaded)) + "\n")
	sb.WriteString(RenderInfoLine("Project:  ", info.ProjectDir) + "\n")
	sb.WriteString(RenderInfoLine("Timestamp:", StyleDim.Render(FormatDate(info.Timestamp))) + "\n")

	m.previewContent = sb.String()
	m.confirmMsg = "Undo this load and restore previous state?"
	m.flow = flowUndo
	m.state = stateConfirm
	return m, nil
}

// --- Message ---

func (m App) updateMessage(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.state = stateMainMenu
		m.flow = flowNone
		m.previewContent = ""
		_ = msg
		return m, nil
	}
	return m, nil
}

func (m App) viewMessage() string {
	var sb strings.Builder
	sb.WriteString(Banner())
	sb.WriteString("\n")
	sb.WriteString(m.message + "\n\n")
	sb.WriteString("  " + StyleDim.Render("Press any key to continue..."))
	sb.WriteString("\n")
	return sb.String()
}

// --- Non-interactive commands (no Bubble Tea) ---

// RunList outputs presets for non-interactive use.
func RunList(cfg *config.Config, jsonOutput bool) error {
	if err := cfg.EnsureDirs(); err != nil {
		return err
	}

	presets, err := preset.List(cfg)
	if err != nil {
		return err
	}

	if jsonOutput {
		if presets == nil {
			presets = []*preset.Meta{}
		}
		data, err := json.MarshalIndent(presets, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	if len(presets) == 0 {
		fmt.Println("No presets found.")
		return nil
	}

	for _, p := range presets {
		desc := ""
		if p.Description != "" {
			desc = "  " + p.Description
		}
		lastUsed := ""
		if p.LastUsed != "" {
			lastUsed = fmt.Sprintf("  (%s)", FormatDate(p.LastUsed))
		}
		fmt.Printf("%-20s%s%s\n", p.Name, desc, lastUsed)
	}
	return nil
}

// RunLoadNonInteractive loads a preset without interactive prompts.
func RunLoadNonInteractive(cfg *config.Config, name, projectDir string) error {
	if err := cfg.EnsureDirs(); err != nil {
		return err
	}

	count, err := preset.Load(cfg, name, projectDir)
	if err != nil {
		return err
	}
	fmt.Printf("%s Loaded preset %s (%d files). Previous state backed up.\n",
		StyleGreenBold.Render("✓"), StyleMagBold.Render(name), count)
	return nil
}
