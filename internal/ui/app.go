package ui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/kincoy/cc9s/internal/claudefs"
	"github.com/kincoy/cc9s/internal/ui/styles"
)

// Screen screen type enum
type Screen int

const (
	ScreenProjects Screen = iota // project list
	ScreenSessions               // session list (unified, supports context filtering)
	ScreenSkills                 // skills list
)

// ResourceType resource type
type ResourceType int

const (
	ResourceProjects ResourceType = iota // projects resource
	ResourceSessions                     // sessions resource
	ResourceSkills                       // skills resource
)

// InputMode input mode
type InputMode int

const (
	InputNormal  InputMode = iota // normal mode
	InputSearch                   // search mode (triggered by /)
	InputCommand                  // command mode (triggered by :)
)

// AppModel root application Model, responsible for routing and layout
type AppModel struct {
	width    int
	height   int
	ready    bool
	showHelp bool
	quitting bool

	currentScreen     Screen
	currentResource   ResourceType
	inputMode         InputMode
	projectList       *ProjectListModel
	sessionList       *SessionListModel
	skillList         *SkillListModel
	lastProjectCursor int // saves the cursor position of the project list

	searchInput  textinput.Model
	commandInput textinput.Model

	tabCompletionIndex int

	showingDialog bool
	confirmDialog *ConfirmDialogModel

	flashMsg     string
	flashUntil   time.Time
	flashIsError bool

	showingDetail      bool
	detailView         *DetailViewModel
	showingSkillDetail bool
	skillDetailView    *SkillDetailViewModel
	showingLog         bool
	logView            *LogViewModel
}

// NewAppModel creates a new application Model
func NewAppModel() *AppModel {
	si := textinput.New()
	si.Prompt = "/"
	si.Placeholder = "search..."
	si.CharLimit = 256

	ci := textinput.New()
	ci.Prompt = ":"
	ci.Placeholder = "command..."
	ci.CharLimit = 256

	return &AppModel{
		currentScreen:   ScreenProjects,
		currentResource: ResourceProjects,
		projectList:     NewProjectListModel(),
		searchInput:     si,
		commandInput:    ci,
	}
}

func (a *AppModel) Init() tea.Cmd {
	return a.projectList.Init()
}

func (a *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Dialog has highest priority
	if a.showingDialog && a.confirmDialog != nil {
		cmd := a.confirmDialog.Update(msg)
		if cmd != nil {
			// Close dialog and execute command
			a.showingDialog = false
			a.confirmDialog = nil
			return a, cmd
		}
		// Dialog consumed the message, do not pass down
		return a, nil
	}

	// Detail view handles message
	if a.showingDetail && a.detailView != nil {
		cmd := a.detailView.Update(msg)
		if cmd != nil {
			resultMsg := cmd()
			if _, ok := resultMsg.(CloseDetailMsg); ok {
				a.showingDetail = false
				a.detailView = nil
				return a, nil
			}
		}
		return a, cmd
	}

	// Skill detail view handles message
	if a.showingSkillDetail && a.skillDetailView != nil {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok && (keyMsg.String() == "e" || keyMsg.String() == "E") {
			return a, editSkillCmd(a.skillDetailView.skill)
		}
		cmd := a.skillDetailView.Update(msg)
		if cmd != nil {
			resultMsg := cmd()
			if _, ok := resultMsg.(CloseSkillDetailMsg); ok {
				a.showingSkillDetail = false
				a.skillDetailView = nil
				return a, nil
			}
		}
		return a, cmd
	}

	// Log view handles message
	if a.showingLog && a.logView != nil {
		cmd := a.logView.Update(msg)
		if cmd != nil {
			resultMsg := cmd()
			if _, ok := resultMsg.(CloseLogMsg); ok {
				a.showingLog = false
				a.logView = nil
				return a, nil
			}
		}
		return a, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.ready = true

	case ShowConfirmDialogMsg:
		a.showingDialog = true
		a.confirmDialog = msg.Dialog
		return a, nil

	case ShowDetailMsg:
		a.showingDetail = true
		a.detailView = NewDetailViewModel(msg.Session)
		return a, a.detailView.Init()

	case CloseDetailMsg:
		a.showingDetail = false
		a.detailView = nil
		return a, nil

	case ShowSkillDetailMsg:
		a.showingSkillDetail = true
		a.skillDetailView = NewSkillDetailViewModel(msg.Skill)
		return a, a.skillDetailView.Init()

	case CloseSkillDetailMsg:
		a.showingSkillDetail = false
		a.skillDetailView = nil
		return a, nil

	case EditSkillMsg:
		return a, editSkillCmd(msg.Skill)

	case SkillEditorFinishedMsg:
		if msg.Err != nil && !isSignalError(msg.Err) {
			a.SetFlash(fmt.Sprintf("Editor exited with error: %v", msg.Err), true, 5*time.Second)
		} else if msg.Err == nil {
			a.SetFlash(fmt.Sprintf("Edited skill: %s", msg.Skill.Name), false, 2*time.Second)
		}
		if a.skillList != nil {
			return a, a.skillList.Reload()
		}
		return a, nil

	case ShowLogMsg:
		a.showingLog = true
		a.logView = NewLogViewModel(msg.Session)
		return a, a.logView.Init()

	case CloseLogMsg:
		a.showingLog = false
		a.logView = nil
		return a, nil

	case SwitchResourceMsg:
		a.currentResource = msg.Resource
		switch msg.Resource {
		case ResourceProjects:
			a.currentScreen = ScreenProjects
		case ResourceSessions:
			targetCtx := a.currentSessionTargetContext()
			a.currentScreen = ScreenSessions
			if a.sessionList == nil {
				a.sessionList = NewSessionListModel()
				cmd := a.sessionList.Init()
				if ctxCmd := a.sessionList.SetContext(targetCtx); ctxCmd != nil {
					return a, tea.Batch(cmd, ctxCmd)
				}
				return a, cmd
			}
			cmd := a.sessionList.SetContext(targetCtx)
			if cmd != nil {
				return a, cmd
			}
		case ResourceSkills:
			targetCtx := a.currentSkillTargetContext()
			a.currentScreen = ScreenSkills
			if a.skillList == nil {
				a.skillList = NewSkillListModel()
				cmd := a.skillList.Init()
				if ctxCmd := a.skillList.SetContext(targetCtx); ctxCmd != nil {
					return a, tea.Batch(cmd, ctxCmd)
				}
				return a, cmd
			}
			cmd := a.skillList.SetContext(targetCtx)
			if cmd != nil {
				return a, cmd
			}
		}

	case SwitchContextMsg:
		if a.currentScreen == ScreenSessions && a.sessionList != nil {
			a.currentScreen = ScreenSessions
			return a, a.sessionList.SetContext(msg.Context)
		}
		if a.currentScreen == ScreenSkills && a.skillList != nil {
			a.currentScreen = ScreenSkills
			return a, a.skillList.SetContext(msg.Context)
		}

	case DeleteSessionsMsg:
		// Execute deletion asynchronously
		return a, deleteSessionsCmd(msg.Targets)

	case SessionsDeletedMsg:
		if len(msg.Errs) > 0 {
			a.SetFlash(
				fmt.Sprintf("Deleted %d sessions, %d errors", msg.Deleted, len(msg.Errs)),
				true,
				5*time.Second,
			)
		} else {
			a.SetFlash(
				fmt.Sprintf("Deleted %d sessions", msg.Deleted),
				false,
				2*time.Second,
			)
		}
		// Refresh current list + project list
		var cmds []tea.Cmd
		if a.sessionList != nil {
			a.sessionList.ClearSelection()
			cmds = append(cmds, a.sessionList.Reload())
		}
		// Also refresh project list (SessionCount may have changed)
		cmds = append(cmds, scanProjectsCmd)
		return a, tea.Batch(cmds...)

	case sessionResumedMsg:
		// Set flash message
		if msg.err != nil && !isSignalError(msg.err) {
			a.SetFlash(
				fmt.Sprintf("❌ Resume failed: %v", msg.err),
				true,
				5*time.Second,
			)
		} else if msg.err == nil {
			a.SetFlash(
				fmt.Sprintf("✓ Returned from session %s", msg.sessionID),
				false,
				2*time.Second,
			)
		}

		// Refresh session list
		if a.sessionList != nil {
			return a, a.sessionList.Reload()
		}
		return a, nil

	case tea.KeyPressMsg:
		// Search/command mode has highest priority (after dialog)
		switch a.inputMode {
		case InputSearch:
			return a.handleSearchInput(msg)
		case InputCommand:
			return a.handleCommandInput(msg)
		}

		// Global shortcuts
		switch msg.String() {
		case "q", "ctrl+c":
			a.quitting = true
			return a, tea.Quit
		case "?":
			a.showHelp = !a.showHelp
		case "esc":
			if a.clearActiveSearch() {
				return a, nil
			}
		case "/":
			a.inputMode = InputSearch
			a.searchInput.SetValue("")
			a.searchInput.Focus()
			return a, nil
		case ":":
			a.inputMode = InputCommand
			a.commandInput.SetValue("")
			a.commandInput.Focus()
			return a, nil
		case "0":
			if a.currentScreen == ScreenSessions || a.currentScreen == ScreenSkills {
				// Shortcut to switch to context=all
				return a, func() tea.Msg {
					return SwitchContextMsg{Context: Context{Type: ContextAll}}
				}
			}
		}
	}

	// Always forward project data updates (project stats need refresh after session deletion)
	if _, ok := msg.(projectsLoadedMsg); ok {
		a.projectList.Update(msg)
	}

	// Dispatch message to current view
	var cmd tea.Cmd
	switch a.currentScreen {
	case ScreenProjects:
		cmd = a.projectList.Update(msg)

		// Listen for EnterProjectMsg
		if cmd != nil {
			resultMsg := cmd()
			if enterMsg, ok := resultMsg.(EnterProjectMsg); ok {
				// Save current cursor
				a.lastProjectCursor = a.projectList.cursor
				// Create session list Model (with specific project context)
				a.sessionList = NewSessionListModelForProject(enterMsg.Project.Name)
				a.currentScreen = ScreenSessions
				return a, a.sessionList.Init()
			}
		}

	case ScreenSessions:
		if a.sessionList != nil {
			cmd = a.sessionList.Update(msg)
		}

		// Listen for BackToProjectsMsg
		if cmd != nil {
			resultMsg := cmd()
			if _, ok := resultMsg.(BackToProjectsMsg); ok {
				// Restore cursor position
				a.projectList.cursor = a.lastProjectCursor
				a.currentScreen = ScreenProjects
				return a, nil
			}
		}

	case ScreenSkills:
		if a.skillList != nil {
			cmd = a.skillList.Update(msg)
		}
	}

	if _, ok := msg.(skillsLoadedMsg); ok && a.showingSkillDetail && a.skillDetailView != nil && a.skillList != nil {
		if skill, found := a.skillList.GetSelectedSkill(); found {
			a.skillDetailView.skill = skill
		}
	}

	return a, cmd
}

func (a *AppModel) View() tea.View {
	if !a.ready {
		v := tea.NewView("Initializing...")
		v.AltScreen = true
		return v
	}
	if a.quitting {
		return tea.NewView("")
	}

	var content string

	// Render different layouts based on screen state
	if a.height < 6 {
		// Small screen: show body + footer only
		bodyHeight := a.height - 1
		if bodyHeight < 1 {
			bodyHeight = 1
		}
		bodyContent := a.renderBody(a.width, bodyHeight)
		body := lipgloss.NewStyle().
			Height(bodyHeight).
			Width(a.width).
			Render(bodyContent)

		footer := a.renderFooter(a.width)
		content = lipgloss.JoinVertical(lipgloss.Left, body, footer)
	} else {
		// Normal screen: header + cmdline + body + footer
		header := a.renderHeader()

		// Command line area (between header and body, shown only in search/command mode)
		cmdLine := ""
		cmdLineHeight := 0
		if a.inputMode != InputNormal {
			var cmdContent string
			var cmdBorder lipgloss.Style
			switch a.inputMode {
			case InputSearch:
				cmdContent = a.searchInput.View()
				cmdBorder = styles.SearchBorderStyle
			case InputCommand:
				cmdContent = a.renderCommandLine()
				cmdBorder = styles.CommandBorderStyle
			}
			// Wrap command line content with ThickBorder (outputs 3 lines: top border + content + bottom border)
			cmdLine = cmdBorder.Width(a.width).Render(cmdContent)
			cmdLineHeight = 3
		}

		// Height calculation: header(3, ThickBorder) + cmdline(0|3) + footer(1) = 4|7
		bodyHeight := a.height - 4 - cmdLineHeight
		if bodyHeight < 1 {
			bodyHeight = 1
		}
		bodyContent := a.renderBody(a.width, bodyHeight)
		body := lipgloss.NewStyle().
			Height(bodyHeight).
			Width(a.width).
			Render(bodyContent)

		var footer string
		if a.HasFlash() {
			footer = renderFlashFooter(a.width, a.flashMsg, a.flashIsError)
		} else {
			footer = a.renderFooter(a.width)
		}

		if cmdLineHeight > 0 {
			content = lipgloss.JoinVertical(lipgloss.Left, header, cmdLine, body, footer)
		} else {
			content = lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
		}
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// getSessionContext gets the current session context (for footer etc.)
func (a *AppModel) getSessionContext() Context {
	if a.sessionList != nil {
		return a.sessionList.GetContext()
	}
	return Context{Type: ContextAll}
}

func (a *AppModel) getSkillContext() Context {
	if a.skillList != nil {
		return a.skillList.GetContext()
	}
	return Context{Type: ContextAll}
}

func (a *AppModel) currentSessionTargetContext() Context {
	switch a.currentScreen {
	case ScreenProjects:
		if a.projectList != nil && len(a.projectList.projects) > 0 && a.projectList.cursor < len(a.projectList.projects) {
			return Context{Type: ContextProject, Value: a.projectList.projects[a.projectList.cursor].Name}
		}
	case ScreenSessions:
		return a.getSessionContext()
	case ScreenSkills:
		return a.getSkillContext()
	}
	return Context{Type: ContextAll}
}

func (a *AppModel) currentSkillTargetContext() Context {
	switch a.currentScreen {
	case ScreenProjects:
		return Context{Type: ContextAll}
	case ScreenSessions:
		return a.getSessionContext()
	case ScreenSkills:
		return a.getSkillContext()
	}
	return Context{Type: ContextAll}
}

// renderHeader renders the header (context-aware)
func (a *AppModel) renderHeader() string {
	if a.currentScreen == ScreenProjects {
		projectCount, totalSessions, activeCount := a.projectList.GetStats()
		contextLabel := fmt.Sprintf("%d projects", projectCount)
		statsLabel := fmt.Sprintf("%d sessions / %d active", totalSessions, activeCount)
		if a.inputMode == InputSearch {
			filtered, total := a.projectList.GetFilterStats()
			return renderHeaderWithFilter(a.width, contextLabel, statsLabel, filtered, total)
		}
		return renderHeader(a.width, contextLabel, statsLabel)
	}

	if a.currentScreen == ScreenSkills && a.skillList != nil {
		total, ready, invalid := a.skillList.GetStats()
		ctx := a.skillList.GetContext()
		contextLabel := "All Skills"
		if ctx.Type == ContextProject {
			contextLabel = ctx.Value
		}
		statsLabel := formatSkillSummary(a.width, total, ready, invalid)
		if a.inputMode == InputSearch {
			filtered, total := a.skillList.GetFilterStats()
			return renderHeaderWithFilter(a.width, contextLabel, statsLabel, filtered, total)
		}
		return renderHeader(a.width, contextLabel, statsLabel)
	}

	if a.sessionList != nil {
		ctx := a.sessionList.GetContext()
		summary := summarizeGlobalSessions(a.sessionList.sessions)
		var contextLabel string
		if ctx.Type == ContextAll {
			contextLabel = "All Projects"
		} else {
			contextLabel = ctx.Value
		}
		statsLabel := formatLifecycleSummary(a.width, summary)
		if a.inputMode == InputSearch {
			filtered, total := a.sessionList.GetFilterStats()
			return renderHeaderWithFilter(a.width, contextLabel, statsLabel, filtered, total)
		}
		return renderHeader(a.width, contextLabel, statsLabel)
	}

	return renderHeader(a.width, "0 projects", "0 sessions / 0 active")
}

// renderBody renders the body area content
func (a *AppModel) renderBody(width, height int) string {
	// Priority: dialog > detail > log > help > session/project

	// Render background content
	var background string
	switch a.currentScreen {
	case ScreenProjects:
		background = a.projectList.View(width, height)
	case ScreenSessions:
		if a.sessionList != nil {
			background = a.sessionList.View(width, height)
		} else {
			background = renderCenteredText("Session list not initialized", width, height)
		}
	case ScreenSkills:
		if a.skillList != nil {
			background = a.skillList.View(width, height)
		} else {
			background = renderCenteredText("Skill list not initialized", width, height)
		}
	default:
		background = renderCenteredText("Unknown screen", width, height)
	}

	// Dialog overlay (highest priority)
	if a.showingDialog && a.confirmDialog != nil {
		return overlayDialog(background, a.confirmDialog.ViewBox(width), width, height)
	}

	// Detail view overlay
	if a.showingDetail && a.detailView != nil {
		cmd := a.detailView.Update(tea.WindowSizeMsg{Width: width, Height: height})
		if cmd != nil {
			// If CloseDetailMsg is returned, handle in Update
		}
		// Get panel content with ViewBox, then overlay on background
		return overlayDialog(background, a.detailView.ViewBox(width), width, height)
	}

	if a.showingSkillDetail && a.skillDetailView != nil {
		cmd := a.skillDetailView.Update(tea.WindowSizeMsg{Width: width, Height: height})
		if cmd != nil {
			// If CloseSkillDetailMsg is returned, handle in Update
		}
		return overlayDialog(background, a.skillDetailView.ViewBox(width), width, height)
	}

	// Log view (fullscreen, does not preserve background)
	if a.showingLog && a.logView != nil {
		cmd := a.logView.Update(tea.WindowSizeMsg{Width: width, Height: height})
		if cmd != nil {
			// If CloseLogMsg is returned, handle in Update
		}
		// Log view fullscreen
		return a.logView.View(width, height)
	}

	// Help panel
	if a.showHelp {
		return renderHelp(width, height)
	}

	// Normal view
	return background
}

// overlayDialog overlays the dialog on background content (centered, background dimmed)
func overlayDialog(background, dialogBox string, width, height int) string {
	bgLines := strings.Split(background, "\n")
	dialogLines := strings.Split(dialogBox, "\n")

	dialogHeight := len(dialogLines)
	startRow := (height - dialogHeight) / 2
	if startRow < 0 {
		startRow = 0
	}

	// Ensure enough background rows
	for len(bgLines) < height {
		bgLines = append(bgLines, strings.Repeat(" ", width))
	}
	if len(bgLines) > height {
		bgLines = bgLines[:height]
	}

	// Dim uncovered background rows (ANSI SGR 2 = faint/dim, independent of color attributes)
	const dimOn = "\x1b[2m"
	const dimOff = "\x1b[22m"
	for i := range bgLines {
		bgLines[i] = dimOn + bgLines[i] + dimOff
	}

	// Overlay dialog rows on corresponding background positions (normal brightness, no dim)
	for i, line := range dialogLines {
		row := startRow + i
		if row >= 0 && row < len(bgLines) {
			dialogWidth := lipgloss.Width(line)
			padding := (width - dialogWidth) / 2
			if padding < 0 {
				padding = 0
			}
			bgLines[row] = strings.Repeat(" ", padding) + line
		}
	}

	return strings.Join(bgLines, "\n")
}

// renderFooter renders the footer (with multi-select count + context info)
// resolveFooterContext resolves the current footer state machine context
func (a *AppModel) resolveFooterContext() FooterContext {
	ctx := FooterContext{
		Screen: a.currentScreen,
		Mode:   FooterModeNormal,
	}

	// Multi-select state
	if a.sessionList != nil {
		ctx.HasMulti = a.sessionList.HasSelection()
		ctx.SelCount = a.sessionList.SelectedCount()
	}

	// Overlay priority: Dialog > Help > Detail > Log
	if a.showingDialog {
		ctx.Overlay = OverlayDialog
		if a.confirmDialog != nil {
			ctx.DialogIsAlert = a.confirmDialog.alert
		}
		return ctx
	}

	if a.showHelp {
		ctx.Overlay = OverlayHelp
		return ctx
	}

	if a.showingDetail || a.showingSkillDetail {
		ctx.Overlay = OverlayDetail
		return ctx
	}

	if a.showingLog {
		ctx.Overlay = OverlayLog
		return ctx
	}

	// No overlay, check InputMode
	if a.inputMode == InputSearch {
		ctx.Mode = FooterModeSearch
		return ctx
	}

	if a.inputMode == InputCommand {
		ctx.Mode = FooterModeCommand
		return ctx
	}

	// Normal mode
	return ctx
}

// renderFooter renders the footer
func (a *AppModel) renderFooter(width int) string {
	footerCtx := a.resolveFooterContext()
	hints := hintsForContext(footerCtx)
	return renderFooterWithHints(width, hints)
}

// SetFlash sets a flash message
func (a *AppModel) SetFlash(msg string, isError bool, duration time.Duration) {
	a.flashMsg = msg
	a.flashIsError = isError
	a.flashUntil = time.Now().Add(duration)
}

// HasFlash checks if there is an active flash message
func (a *AppModel) HasFlash() bool {
	return time.Now().Before(a.flashUntil)
}

// deleteSessionsCmd asynchronously executes session deletion
func deleteSessionsCmd(targets []claudefs.DeleteTarget) tea.Cmd {
	return func() tea.Msg {
		deleted, errs := claudefs.DeleteSessions(targets)
		return SessionsDeletedMsg{Deleted: deleted, Errs: errs}
	}
}

// isSignalError checks if the error is a signal error (user Ctrl+C)
func isSignalError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "signal") ||
		strings.Contains(errStr, "interrupt")
}

// handleSearchInput handles key presses in search input mode
func (a *AppModel) handleSearchInput(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		a.inputMode = InputNormal
		a.searchInput.SetValue("")
		a.searchInput.Blur()
		a.applySearchToCurrentView("")
		return a, nil
	case "enter":
		// Search filters in real-time, Enter exits search mode
		a.inputMode = InputNormal
		a.searchInput.Blur()
		return a, nil
	}

	var cmd tea.Cmd
	a.searchInput, cmd = a.searchInput.Update(msg)

	// Real-time filter: pass search query to current view
	query := strings.TrimSpace(a.searchInput.Value())
	a.applySearchToCurrentView(query)

	return a, cmd
}

// handleCommandInput handles key presses in command input mode
func (a *AppModel) handleCommandInput(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		a.inputMode = InputNormal
		a.commandInput.Blur()
		a.tabCompletionIndex = 0
		return a, nil
	case "enter":
		cmd := a.executeCommand(a.commandInput.Value())
		a.inputMode = InputNormal
		a.commandInput.Blur()
		a.commandInput.SetValue("")
		a.tabCompletionIndex = 0
		return a, cmd
	case "tab":
		a.handleTabCompletion()
		return a, nil
	}

	// Non-Tab key resets completion index
	a.tabCompletionIndex = 0

	var cmd tea.Cmd
	a.commandInput, cmd = a.commandInput.Update(msg)
	return a, cmd
}

// handleTabCompletion handles Tab completion
func (a *AppModel) handleTabCompletion() {
	input := a.commandInput.Value()
	segments := strings.Fields(input)

	var candidates []string
	var prefix string

	if len(segments) == 0 {
		return
	}

	if len(segments) == 1 {
		// First segment: command name completion
		prefix = segments[0]
		commands := []string{"sessions", "projects", "skills", "context"}
		for _, cmd := range commands {
			if strings.HasPrefix(cmd, prefix) {
				candidates = append(candidates, cmd)
			}
		}
	} else if len(segments) >= 2 && segments[0] == "context" {
		// context command: project name completion
		prefix = segments[len(segments)-1]
		if a.projectList != nil {
			for _, proj := range a.projectList.projects {
				if strings.HasPrefix(proj.Name, prefix) {
					candidates = append(candidates, proj.Name)
				}
			}
		}
	}

	if len(candidates) == 0 {
		return
	}

	// Cycle through candidates
	if a.tabCompletionIndex >= len(candidates) {
		a.tabCompletionIndex = 0
	}
	selected := candidates[a.tabCompletionIndex]
	a.tabCompletionIndex++

	// Build new input value
	if len(segments) == 1 {
		// Add space after command name completion
		a.commandInput.SetValue(selected + " ")
	} else {
		// Project name completion: replace last segment
		prefix := strings.Join(segments[:len(segments)-1], " ")
		a.commandInput.SetValue(prefix + " " + selected + " ")
	}

	// Move cursor to end
	a.commandInput.CursorEnd()
}

// getCompletionSuggestion gets the completion suggestion text for the current input
func (a *AppModel) getCompletionSuggestion() string {
	input := a.commandInput.Value()
	segments := strings.Fields(input)

	if len(segments) == 0 {
		return ""
	}

	var candidates []string
	var prefix string

	if len(segments) == 1 {
		prefix = segments[0]
		commands := []string{"sessions", "projects", "skills", "context"}
		for _, cmd := range commands {
			if strings.HasPrefix(cmd, prefix) && cmd != prefix {
				candidates = append(candidates, cmd)
			}
		}
	} else if len(segments) >= 2 && segments[0] == "context" {
		prefix = segments[len(segments)-1]
		if a.projectList != nil {
			for _, proj := range a.projectList.projects {
				if strings.HasPrefix(proj.Name, prefix) && proj.Name != prefix {
					candidates = append(candidates, proj.Name)
				}
			}
		}
	}

	if len(candidates) == 0 {
		return ""
	}

	// Take the first matching completion
	return candidates[0][len(prefix):]
}

// renderCommandLine custom renders the command line (with inline suggestion)
func (a *AppModel) renderCommandLine() string {
	input := a.commandInput.Value()

	// Get completion suggestion
	suggestion := ""
	if input != "" {
		suggestion = a.getCompletionSuggestion()
	}

	if suggestion == "" {
		// No suggestion: fall back to textinput native rendering (preserves cursor blink)
		return a.commandInput.View()
	}

	// Custom render: prompt + user input (normal color) + gray suggestion
	promptStyle := lipgloss.NewStyle().Foreground(styles.ColorHighlight).Bold(true)
	suggestionStyle := lipgloss.NewStyle().Foreground(styles.ColorDim)
	userInputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)

	prompt := promptStyle.Render(":")
	userText := userInputStyle.Render(input)
	suggestText := suggestionStyle.Render(suggestion)

	return prompt + userText + suggestText
}

// executeCommand parses and executes a command string
func (a *AppModel) executeCommand(cmdStr string) tea.Cmd {
	cmdStr = strings.TrimSpace(cmdStr)
	if cmdStr == "" {
		return nil
	}

	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return nil
	}

	switch parts[0] {
	case "sessions":
		return func() tea.Msg { return SwitchResourceMsg{Resource: ResourceSessions} }
	case "projects":
		return func() tea.Msg { return SwitchResourceMsg{Resource: ResourceProjects} }
	case "skills":
		return func() tea.Msg { return SwitchResourceMsg{Resource: ResourceSkills} }
	case "context":
		// :context all | :context <project-name>
		if len(parts) < 2 {
			ctx := a.getSessionContext()
			if a.currentScreen == ScreenSkills {
				ctx = a.getSkillContext()
			}
			if ctx.Type == ContextAll {
				a.SetFlash("Current context: all", false, 2*time.Second)
			} else {
				a.SetFlash(fmt.Sprintf("Current context: %s", ctx.Value), false, 2*time.Second)
			}
			return nil
		}
		if parts[1] == "all" {
			return func() tea.Msg {
				return SwitchContextMsg{Context: Context{Type: ContextAll}}
			}
		}
		// Specific project name
		return func() tea.Msg {
			return SwitchContextMsg{Context: Context{Type: ContextProject, Value: parts[1]}}
		}
	default:
		a.SetFlash(fmt.Sprintf("Unknown command: %s", parts[0]), true, 3*time.Second)
		return nil
	}
}

// applySearchToCurrentView applies the search query to the current view
func (a *AppModel) applySearchToCurrentView(query string) {
	switch a.currentScreen {
	case ScreenProjects:
		a.projectList.ApplyFilter(query)
	case ScreenSessions:
		if a.sessionList != nil {
			a.sessionList.ApplyFilter(query)
		}
	case ScreenSkills:
		if a.skillList != nil {
			a.skillList.ApplyFilter(query)
		}
	}
}

func (a *AppModel) clearActiveSearch() bool {
	switch a.currentScreen {
	case ScreenProjects:
		if a.projectList != nil && a.projectList.HasActiveFilter() {
			a.searchInput.SetValue("")
			a.projectList.ApplyFilter("")
			return true
		}
	case ScreenSessions:
		if a.sessionList != nil && a.sessionList.HasActiveFilter() {
			a.searchInput.SetValue("")
			a.sessionList.ApplyFilter("")
			return true
		}
	case ScreenSkills:
		if a.skillList != nil && a.skillList.HasActiveFilter() {
			a.searchInput.SetValue("")
			a.skillList.ApplyFilter("")
			return true
		}
	}
	return false
}

func summarizeGlobalSessions(global []claudefs.GlobalSession) claudefs.LifecycleSummary {
	sessions := make([]claudefs.Session, 0, len(global))
	for _, gs := range global {
		sessions = append(sessions, gs.Session)
	}
	return claudefs.SummarizeLifecycleSessions(sessions)
}

func formatLifecycleSummary(width int, summary claudefs.LifecycleSummary) string {
	if width < 120 {
		return fmt.Sprintf("%d sessions / A:%d I:%d C:%d S:%d", summary.Total, summary.Active, summary.Idle, summary.Completed, summary.Stale)
	}
	return fmt.Sprintf("%d sessions / %d active / %d idle / %d completed / %d stale", summary.Total, summary.Active, summary.Idle, summary.Completed, summary.Stale)
}

func formatSkillSummary(width, total, ready, invalid int) string {
	if width < 120 {
		return fmt.Sprintf("%d skills / R:%d I:%d", total, ready, invalid)
	}
	return fmt.Sprintf("%d skills / %d ready / %d invalid", total, ready, invalid)
}
