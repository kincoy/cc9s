package ui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/spinner"
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
	ScreenAgents                 // agents list
)

// ResourceType resource type
type ResourceType int

const (
	ResourceProjects ResourceType = iota // projects resource
	ResourceSessions                     // sessions resource
	ResourceSkills                       // skills resource
	ResourceAgents                       // agents resource
)

// InputMode input mode
type InputMode int

const (
	InputNormal  InputMode = iota // normal mode
	InputSearch                   // search mode (triggered by /)
	InputCommand                  // command mode (triggered by :)
)

// Layout constants
const (
	headerHeight = 3 // header block (title + status + breadcrumb)
	footerHeight = 1 // footer hint bar
)

// AppModel root application Model, responsible for routing and layout
type AppModel struct {
	width    int
	height   int
	ready    bool
	showHelp bool
	helpScroll int
	quitting bool
	currentResource      ResourceType
	globalProjectContext Context // shared project context across Sessions/Skills/Agents
	inputMode            InputMode
	projectList       *ProjectListModel
	sessionList       *SessionListModel
	skillList         *SkillListModel
	agentList         *AgentListModel
	resourceRegistry  *ResourceRegistry
	lastProjectCursor int // saves the cursor position of the project list

	searchInput  textinput.Model
	commandInput textinput.Model

	tabCompletionIndex int

	showingDialog bool
	confirmDialog *ConfirmDialogModel

	flashMsg     string
	flashUntil   time.Time
	flashIsError bool

	// Toast notification system
	toastMsg     string
	toastUntil   time.Time
	toastIsError bool

	showingDetail        bool
	detailView           *DetailViewModel
	showingProjectDetail bool
	projectDetailView    *ProjectDetailViewModel
	showingSkillDetail   bool
	skillDetailView      *SkillDetailViewModel
	showingAgentDetail   bool
	agentDetailView      *AgentDetailViewModel
	showingLog           bool
	logView              *LogViewModel

	tabs *TabsModel

	// Animation system
	spinner        spinner.Model
	isLoading      bool
	loadingText    string
	loadingResource ResourceType // which resource triggered the loading
	currentTime    time.Time
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

	s := spinner.New()
	s.Spinner = spinner.Line
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

	return &AppModel{
		currentResource:  ResourceProjects,
		projectList:      NewProjectListModel(),
		resourceRegistry: newResourceRegistry(),
		searchInput:      si,
		commandInput:     ci,
		spinner:          s,
		currentTime:      time.Now(),
		tabs:             NewTabsModel(),
	}
}

func (a *AppModel) Init() tea.Cmd {
	return tea.Batch(
		a.projectList.Init(),
		clockTicker(),
	)
}

func clockTicker() tea.Cmd {
	return tea.Every(time.Second, func(t time.Time) tea.Msg {
		return ClockTickMsg{Time: t}
	})
}

func computeHealthCmd() tea.Cmd {
	return func() tea.Msg {
		health, err := claudefs.ComputeHealthMetrics()
		if err != nil {
			return StopLoadingMsg{}
		}
		result := make(map[string]int)
		for _, ps := range health.ProjectScores {
			result[ps.ProjectName] = ps.HealthScore
		}
		return HealthComputedMsg{Health: result}
	}
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

	// Check if this is a data-loaded message that needs special handling.
	// These messages must be processed even when overlays are active, to:
	// 1. Stop the loading spinner for the correct resource
	// 2. Sync detail overlays with updated data
	isDataLoadedMsg := false
	switch msg.(type) {
	case projectsLoadedMsg, skillsLoadedMsg, agentsLoadedMsg, sessionsLoadedMsg:
		isDataLoadedMsg = true
	}

	// Detail views and overlays handle most messages, but data-loaded messages
	// are processed first to ensure spinner state and overlay sync.
	if !isDataLoadedMsg {
		// Detail view handles message
		if a.showingDetail && a.detailView != nil {
			if _, ok := msg.(CloseDetailMsg); ok {
				a.showingDetail = false
				a.detailView = nil
				return a, nil
			}
			return a, a.detailView.Update(msg)
		}

		if a.showingProjectDetail && a.projectDetailView != nil {
			if _, ok := msg.(CloseProjectDetailMsg); ok {
				a.showingProjectDetail = false
				a.projectDetailView = nil
				return a, nil
			}
			return a, a.projectDetailView.Update(msg)
		}

		// Skill detail view handles message
		if a.showingSkillDetail && a.skillDetailView != nil {
			if _, ok := msg.(CloseSkillDetailMsg); ok {
				a.showingSkillDetail = false
				a.skillDetailView = nil
				return a, nil
			}
			if keyMsg, ok := msg.(tea.KeyPressMsg); ok && (keyMsg.String() == "e" || keyMsg.String() == "E") {
				return a, editSkillCmd(a.skillDetailView.skill)
			}
			return a, a.skillDetailView.Update(msg)
		}

		if a.showingAgentDetail && a.agentDetailView != nil {
			if _, ok := msg.(CloseAgentDetailMsg); ok {
				a.showingAgentDetail = false
				a.agentDetailView = nil
				return a, nil
			}
			if keyMsg, ok := msg.(tea.KeyPressMsg); ok && (keyMsg.String() == "e" || keyMsg.String() == "E") {
				return a, editAgentCmd(a.agentDetailView.agent)
			}
			return a, a.agentDetailView.Update(msg)
		}

		// Log view handles message
		if a.showingLog && a.logView != nil {
			if _, ok := msg.(CloseLogMsg); ok {
				a.showingLog = false
				a.logView = nil
				return a, nil
			}
			return a, a.logView.Update(msg)
		}
	}

	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(msg)
		return a, cmd

	case ClockTickMsg:
		a.currentTime = msg.Time
		return a, clockTicker()

	case StartLoadingMsg:
		a.isLoading = true
		a.loadingText = msg.Text
		return a, a.spinner.Tick

	case StopLoadingMsg:
		// Only stop loading if the message targets the current loading resource.
		// Zero Resource means unconditional stop (e.g. health computation failure).
		if msg.Resource == 0 || msg.Resource == a.loadingResource {
			a.isLoading = false
		}
		return a, nil

	case ShowToastMsg:
		a.toastMsg = msg.Message
		a.toastIsError = msg.IsError
		a.toastUntil = time.Now().Add(2 * time.Second)
		return a, nil

	case HealthComputedMsg:
		a.projectList.projectHealth = msg.Health
		a.projectList.updateViewportContent()
		a.isLoading = false
		a.toastMsg = "Health metrics computed"
		a.toastIsError = false
		a.toastUntil = time.Now().Add(2 * time.Second)
		return a, nil

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.ready = true

		// Forward resize to all list models so their viewports get correct dimensions
		a.projectList.Update(msg)
		if a.sessionList != nil {
			a.sessionList.Update(msg)
		}
		if a.skillList != nil {
			a.skillList.Update(msg)
		}
		if a.agentList != nil {
			a.agentList.Update(msg)
		}

	case ShowConfirmDialogMsg:
		a.showingDialog = true
		a.confirmDialog = msg.Dialog
		return a, nil

	case BackToProjectsMsg:
		a.globalProjectContext = Context{Type: ContextAll}
		a.projectList.cursor = a.lastProjectCursor
		if ensureActive := a.resourceRegistry.MustGet(ResourceProjects).EnsureActive; ensureActive != nil {
			return a, ensureActive(a, Context{Type: ContextAll})
		}
		return a, nil

	case ShowDetailMsg:
		a.showingDetail = true
		a.detailView = NewDetailViewModel(msg.Session)
		return a, a.detailView.Init()

	case CloseDetailMsg:
		a.showingDetail = false
		a.detailView = nil
		return a, nil

	case ShowProjectDetailMsg:
		a.showingProjectDetail = true
		a.projectDetailView = NewProjectDetailViewModel(msg.Project)
		return a, a.projectDetailView.Init()

	case CloseProjectDetailMsg:
		a.showingProjectDetail = false
		a.projectDetailView = nil
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

	case ShowAgentDetailMsg:
		a.showingAgentDetail = true
		a.agentDetailView = NewAgentDetailViewModel(msg.Agent)
		return a, a.agentDetailView.Init()

	case CloseAgentDetailMsg:
		a.showingAgentDetail = false
		a.agentDetailView = nil
		return a, nil

	case EditAgentMsg:
		return a, editAgentCmd(msg.Agent)

	case AgentEditorFinishedMsg:
		if msg.Err != nil && !isSignalError(msg.Err) {
			a.SetFlash(fmt.Sprintf("Editor exited with error: %v", msg.Err), true, 5*time.Second)
		} else if msg.Err == nil {
			a.SetFlash(fmt.Sprintf("Edited agent: %s", msg.Agent.Name), false, 2*time.Second)
		}
		if a.agentList != nil {
			return a, a.agentList.Reload()
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
		descriptor := a.resourceRegistry.MustGet(msg.Resource)
		targetCtx := Context{Type: ContextAll}
		if descriptor.ResolveTargetContext != nil {
			targetCtx = descriptor.ResolveTargetContext(a)
		}
		if descriptor.EnsureActive != nil {
			return a, descriptor.EnsureActive(a, targetCtx)
		}

	case EnterProjectMsg:
		a.lastProjectCursor = a.projectList.cursor
		// Set global project context so all resources (Sessions/Skills/Agents) share it
		a.globalProjectContext = Context{Type: ContextProject, Value: msg.Project.Name}
		descriptor := a.resourceRegistry.MustGet(ResourceSessions)
		if descriptor.EnsureActive != nil {
			return a, descriptor.EnsureActive(a, a.globalProjectContext)
		}
		return a, nil

	case SwitchContextMsg:
		a.globalProjectContext = msg.Context
		if setContext := a.currentResourceDescriptor().SetContext; setContext != nil {
			return a, setContext(a, msg.Context)
		}

	case ToggleCleanupHintsMsg:
		if a.sessionList != nil {
			a.sessionList.showCleanupHints = !a.sessionList.showCleanupHints
			a.sessionList.updateViewportContent()
		}
		return a, nil

	case DeleteSessionsMsg:
		// Execute deletion asynchronously
		return a, deleteSessionsCmd(msg.Targets)

	case SessionsDeletedMsg:
		if len(msg.Errs) > 0 {
			a.toastMsg = fmt.Sprintf("Deleted %d, %d errors", msg.Deleted, len(msg.Errs))
			a.toastIsError = true
			a.toastUntil = time.Now().Add(3 * time.Second)
		} else {
			a.toastMsg = fmt.Sprintf("Deleted %d sessions", msg.Deleted)
			a.toastIsError = false
			a.toastUntil = time.Now().Add(2 * time.Second)
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
			if a.showHelp {
				a.helpScroll = 0
			}
			return a, nil
		case "1":
			return a, func() tea.Msg { return SwitchResourceMsg{Resource: ResourceProjects} }
		case "2":
			return a, func() tea.Msg { return SwitchResourceMsg{Resource: ResourceSessions} }
		case "3":
			return a, func() tea.Msg { return SwitchResourceMsg{Resource: ResourceSkills} }
		case "4":
			return a, func() tea.Msg { return SwitchResourceMsg{Resource: ResourceAgents} }
		case "esc":
			if a.showHelp {
				a.showHelp = false
				a.helpScroll = 0
				return a, nil
			}
			if a.clearActiveSearch() {
				return a, nil
			}
		case "j", "down":
			if a.showHelp {
				maxScroll := a.maxHelpScroll()
				if a.helpScroll < maxScroll {
					a.helpScroll++
				}
				return a, nil
			}
		case "k", "up":
			if a.showHelp {
				if a.helpScroll > 0 {
					a.helpScroll--
				}
				return a, nil
			}
		case "g":
			if a.showHelp {
				a.helpScroll = 0
				return a, nil
			}
		case "G":
			if a.showHelp {
				a.helpScroll = a.maxHelpScroll()
				return a, nil
			}
		case "/":
			if !a.currentResourceDescriptor().Capabilities.SupportsSearch {
				return a, nil
			}
			if canStartSearch := a.currentResourceDescriptor().CanStartSearch; canStartSearch != nil && !canStartSearch(a) {
				return a, nil
			}
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
			if a.currentResourceDescriptor().Capabilities.SupportsAllContextShortcut {
				// Shortcut to switch to context=all
				return a, func() tea.Msg {
					return SwitchContextMsg{Context: Context{Type: ContextAll}}
				}
			}
		case "r":
			if a.isLoading {
				return a, nil
			}
			return a, a.refreshCurrentResource()
		case "left":
			if a.showHelp {
				return a, nil
			}
			next := a.tabs.PrevResource()
			return a, func() tea.Msg { return SwitchResourceMsg{Resource: next} }
		case "right":
			if a.showHelp {
				return a, nil
			}
			next := a.tabs.NextResource()
			return a, func() tea.Msg { return SwitchResourceMsg{Resource: next} }
		}
	}

	// Always forward data-loaded messages to their respective list models,
	// even when the user has switched away during loading. This ensures
	// StopLoadingMsg gets emitted and the loading spinner doesn't get stuck.
	var forwardedCmd tea.Cmd
	var isLoadedMsg bool
	if _, ok := msg.(projectsLoadedMsg); ok {
		forwardedCmd = a.projectList.Update(msg)
		isLoadedMsg = true
	}
	if _, ok := msg.(skillsLoadedMsg); ok && a.skillList != nil {
		forwardedCmd = a.skillList.Update(msg)
		isLoadedMsg = true
	}
	if _, ok := msg.(agentsLoadedMsg); ok && a.agentList != nil {
		forwardedCmd = a.agentList.Update(msg)
		isLoadedMsg = true
	}
	if _, ok := msg.(sessionsLoadedMsg); ok && a.sessionList != nil {
		forwardedCmd = a.sessionList.Update(msg)
		isLoadedMsg = true
	}

	// Sync detail overlays after processing loaded messages
	if isLoadedMsg {
		a.syncDetailOverlaysAfterLoad(msg)
		// Return the StopLoadingMsg command without calling updateCurrentView again
		// (the loaded message was already processed by the list model above)
		return a, forwardedCmd
	}

	// For non-loaded messages, delegate to current view
	cmd := a.updateCurrentView(msg)
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
	if a.height < 8 {
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
		// Normal screen: header + tabs + cmdline + body + footer
		header := a.renderHeader()
		tabs := a.tabs.Render(a.width)

		// Command line area (between tabs and body, shown only in search/command mode)
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

		// Height calculation: header(3) + tabs(2) + cmdline(0|3) + footer(1)
		bodyHeight := a.height - headerHeight - TabsHeight - footerHeight - cmdLineHeight
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
			content = lipgloss.JoinVertical(lipgloss.Left, header, tabs, cmdLine, body, footer)
		} else {
			content = lipgloss.JoinVertical(lipgloss.Left, header, tabs, body, footer)
		}
	}

	// Apply forced background if theme requires it
	if styles.ForceBackground && styles.ColorBackground != nil {
		content = lipgloss.NewStyle().
			Background(styles.ColorBackground).
			Width(a.width).
			Height(a.height).
			Render(content)
	}

	// Overlay toast in top-right if active
	if toast := a.renderToast(); toast != "" {
		toastWidth := lipgloss.Width(toast)
		padding := a.width - toastWidth
		if padding < 1 {
			padding = 1
		}
		rightAligned := strings.Repeat(" ", padding) + toast
		contentLines := strings.Split(content, "\n")
		toastLines := strings.Split(rightAligned, "\n")
		for i := 0; i < len(toastLines) && i < len(contentLines); i++ {
			contentLines[i] = toastLines[i]
		}
		content = strings.Join(contentLines, "\n")
	}

	v := tea.NewView(content)
	v.AltScreen = true

	// Dynamic window title
	resourceName := a.currentResourceDescriptor().DisplayName
	currentCount := a.resourceCount(a.currentResource)
	title := fmt.Sprintf("cc9s - %s (%d)", resourceName, currentCount)
	if a.currentResource == ResourceProjects && len(a.projectList.projectHealth) > 0 {
		// Compute average health across projects
		total := 0
		for _, score := range a.projectList.projectHealth {
			total += score
		}
		avgHealth := total / len(a.projectList.projectHealth)
		title += fmt.Sprintf(" | Health: %d", avgHealth)
	}
	v.WindowTitle = title

	return v
}

func (a *AppModel) contextForResource(resource ResourceType) Context {
	switch resource {
	case ResourceSessions:
		if a.sessionList != nil {
			return a.sessionList.GetContext()
		}
	case ResourceSkills:
		if a.skillList != nil {
			return a.skillList.GetContext()
		}
	case ResourceAgents:
		if a.agentList != nil {
			return a.agentList.GetContext()
		}
	}
	return Context{Type: ContextAll}
}

func (a *AppModel) selectedProjectContext() (Context, bool) {
	if a.projectList == nil || len(a.projectList.projects) == 0 {
		return Context{}, false
	}
	if a.projectList.cursor < 0 || a.projectList.cursor >= len(a.projectList.projects) {
		return Context{}, false
	}
	return Context{Type: ContextProject, Value: a.projectList.projects[a.projectList.cursor].Name}, true
}

func (a *AppModel) currentTargetContext(includeSelectedProject bool) Context {
	activeResource := a.currentResource
	if includeSelectedProject && activeResource == ResourceProjects {
		if ctx, ok := a.selectedProjectContext(); ok {
			return ctx
		}
		return Context{Type: ContextAll}
	}
	return a.contextForResource(activeResource)
}

// renderHeader renders the header (context-aware)
func (a *AppModel) renderHeader() string {
	state := a.currentHeaderState()
	resourceLabel := a.currentResourceDescriptor().DisplayName
	if state.HasFilteredState {
		return renderHeaderWithFilter(a.width, resourceLabel, state.ContextLabel, state.StatsLabel, state.FilteredCount, state.TotalCount, a.currentTime)
	}
	return renderHeader(a.width, resourceLabel, state.ContextLabel, state.StatsLabel, a.currentTime)
}

func (a *AppModel) currentHeaderState() ResourceHeaderState {
	if headerState := a.currentResourceDescriptor().HeaderState; headerState != nil {
		state := headerState(a)
		if state.ContextLabel != "" || state.StatsLabel != "" {
			return state
		}
	}

	return ResourceHeaderState{
		ContextLabel: "",
		StatsLabel:   "0 projects / 0 sessions / 0 active",
	}
}

// renderBody renders the body area content
func (a *AppModel) renderBody(width, height int) string {
	// Priority: dialog > detail > log > help > session/project

	// Render background content
	background := a.renderCurrentView(width, height)

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

	if a.showingProjectDetail && a.projectDetailView != nil {
		cmd := a.projectDetailView.Update(tea.WindowSizeMsg{Width: width, Height: height})
		if cmd != nil {
			// If CloseProjectDetailMsg is returned, handle in Update
		}
		return overlayDialog(background, a.projectDetailView.ViewBox(width), width, height)
	}

	if a.showingSkillDetail && a.skillDetailView != nil {
		cmd := a.skillDetailView.Update(tea.WindowSizeMsg{Width: width, Height: height})
		if cmd != nil {
			// If CloseSkillDetailMsg is returned, handle in Update
		}
		return overlayDialog(background, a.skillDetailView.ViewBox(width), width, height)
	}

	if a.showingAgentDetail && a.agentDetailView != nil {
		cmd := a.agentDetailView.Update(tea.WindowSizeMsg{Width: width, Height: height})
		if cmd != nil {
			// If CloseAgentDetailMsg is returned, handle in Update
		}
		return overlayDialog(background, a.agentDetailView.ViewBox(width), width, height)
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
		return renderHelp(width, height, a.resourceRegistry, a.currentResourceDescriptor(), a.helpScroll)
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
		Resource:   a.currentResource,
		Descriptor: a.currentResourceDescriptor(),
		Mode:       FooterModeNormal,
	}

	// Multi-select state
	if a.sessionList != nil {
		ctx.HasMulti = a.sessionList.HasSelection()
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

	if a.showingDetail || a.showingProjectDetail || a.showingSkillDetail || a.showingAgentDetail {
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

func (a *AppModel) updateCurrentView(msg tea.Msg) tea.Cmd {
	switch a.currentResource {
	case ResourceProjects:
		return a.projectList.Update(msg)
	case ResourceSessions:
		if a.sessionList != nil {
			return a.sessionList.Update(msg)
		}
	case ResourceSkills:
		if a.skillList != nil {
			return a.skillList.Update(msg)
		}
	case ResourceAgents:
		if a.agentList != nil {
			return a.agentList.Update(msg)
		}
	}
	return nil
}

func (a *AppModel) refreshCurrentResource() tea.Cmd {
	a.isLoading = true
	a.loadingResource = a.currentResource
	a.loadingText = "Refreshing..."

	var reloadCmd tea.Cmd
	switch a.currentResource {
	case ResourceProjects:
		reloadCmd = a.projectList.Reload()
	case ResourceSessions:
		if a.sessionList != nil {
			reloadCmd = a.sessionList.Reload()
		}
	case ResourceSkills:
		if a.skillList != nil {
			reloadCmd = a.skillList.Reload()
		}
	case ResourceAgents:
		if a.agentList != nil {
			reloadCmd = a.agentList.Reload()
		}
	}
	if reloadCmd == nil {
		a.isLoading = false
		return nil
	}
	return tea.Batch(reloadCmd, a.spinner.Tick)
}

func (a *AppModel) syncDetailOverlaysAfterLoad(msg tea.Msg) {
	if _, ok := msg.(skillsLoadedMsg); ok && a.showingSkillDetail && a.skillDetailView != nil && a.skillList != nil {
		if skill, found := a.skillList.GetSelectedSkill(); found {
			a.skillDetailView.skill = skill
		}
	}
	if _, ok := msg.(agentsLoadedMsg); ok && a.showingAgentDetail && a.agentDetailView != nil && a.agentList != nil {
		if agent, found := a.agentList.GetSelectedAgent(); found {
			a.agentDetailView.agent = agent
		}
	}
}



func (a *AppModel) renderCurrentView(width, height int) string {
	// Show inline spinner in body area only when loading for the *current* resource
	if a.isLoading && a.loadingResource == a.currentResource {
		return a.renderBodySpinner(width, height)
	}

	switch a.currentResource {
	case ResourceProjects:
		return a.projectList.View(width, height)
	case ResourceSessions:
		if a.sessionList != nil {
			return a.sessionList.View(width, height)
		}
		return renderCenteredText("Session list not initialized", width, height)
	case ResourceSkills:
		if a.skillList != nil {
			return a.skillList.View(width, height)
		}
		return renderCenteredText("Skill list not initialized", width, height)
	case ResourceAgents:
		if a.agentList != nil {
			return a.agentList.View(width, height)
		}
		return renderCenteredText("Agent list not initialized", width, height)
	default:
		return renderCenteredText("Unknown screen", width, height)
	}
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
	candidates, _, replaceAll := a.commandCompletionCandidates(input, false)
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
	if replaceAll {
		// Add space after command name completion
		a.commandInput.SetValue(selected + " ")
	} else {
		segments := strings.Fields(input)
		// Project name completion: replace last segment
		prefix := strings.Join(segments[:len(segments)-1], " ")
		a.commandInput.SetValue(prefix + " " + selected + " ")
	}

	// Move cursor to end
	a.commandInput.CursorEnd()
}

// getCompletionSuggestion gets the completion suggestion text for the current input
func (a *AppModel) getCompletionSuggestion() string {
	candidates, prefix, _ := a.commandCompletionCandidates(a.commandInput.Value(), true)
	if len(candidates) == 0 {
		return ""
	}

	// Take the first matching completion
	return candidates[0][len(prefix):]
}

func (a *AppModel) commandCompletionCandidates(input string, excludeExact bool) ([]string, string, bool) {
	segments := strings.Fields(input)
	if len(segments) == 0 {
		return nil, "", false
	}

	if len(segments) == 1 {
		prefix := segments[0]
		commands := append(a.resourceRegistry.CompletionCandidates(prefix), "context")
		if a.currentResource == ResourceProjects {
			commands = append(commands, "health")
		}
		if a.currentResource == ResourceSessions {
			commands = append(commands, "cleanup")
		}
		candidates := make([]string, 0, len(commands))
		for _, cmd := range commands {
			if !strings.HasPrefix(cmd, prefix) {
				continue
			}
			if excludeExact && cmd == prefix {
				continue
			}
			candidates = append(candidates, cmd)
		}
		return candidates, prefix, true
	}

	if len(segments) >= 2 && segments[0] == "context" {
		prefix := segments[len(segments)-1]
		candidates := make([]string, 0)
		if a.projectList != nil {
			for _, proj := range a.projectList.projects {
				if !strings.HasPrefix(proj.Name, prefix) {
					continue
				}
				if excludeExact && proj.Name == prefix {
					continue
				}
				candidates = append(candidates, proj.Name)
			}
		}
		return candidates, prefix, false
	}

	return nil, "", false
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
	case "context":
		// :context all | :context <project-name>
		if len(parts) < 2 {
			ctx := Context{Type: ContextAll}
			if getContext := a.currentResourceDescriptor().CurrentContext; getContext != nil {
				ctx = getContext(a)
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
	case "cleanup":
		if a.currentResource == ResourceSessions {
			return func() tea.Msg { return ToggleCleanupHintsMsg{} }
		}
		return nil
	case "health":
		if a.currentResource == ResourceProjects && a.projectList != nil {
			a.projectList.showHealthColumn = !a.projectList.showHealthColumn
			if a.projectList.showHealthColumn && len(a.projectList.projectHealth) == 0 {
				return tea.Batch(
					func() tea.Msg { return StartLoadingMsg{Text: "Computing health metrics..."} },
					computeHealthCmd(),
				)
			}
		}
		return nil
	default:
		if descriptor, ok := a.resourceRegistry.FindByCommand(parts[0]); ok {
			return func() tea.Msg { return SwitchResourceMsg{Resource: descriptor.Resource} }
		}
		a.SetFlash(fmt.Sprintf("Unknown command: %s", parts[0]), true, 3*time.Second)
		return nil
	}
}

func (a *AppModel) currentResourceDescriptor() ResourceDescriptor {
	return a.resourceRegistry.MustGet(a.currentResource)
}

// resourceCount returns the item count for the given resource type
func (a *AppModel) resourceCount(r ResourceType) int {
	switch r {
	case ResourceProjects:
		return len(a.projectList.projects)
	case ResourceSessions:
		if a.sessionList != nil {
			return len(a.sessionList.allSessions)
		}
		return 0
	case ResourceSkills:
		if a.skillList != nil {
			return len(a.skillList.state.AllItems)
		}
		return 0
	case ResourceAgents:
		if a.agentList != nil {
			return len(a.agentList.state.AllItems)
		}
		return 0
	default:
		return 0
	}
}

func (a *AppModel) setActiveResource(resource ResourceType) {
	a.currentResource = resource
	if a.tabs != nil {
		a.tabs.SetCurrent(resource)
	}
}

// applySearchToCurrentView applies the search query to the current view
func (a *AppModel) applySearchToCurrentView(query string) {
	if applyFilter := a.currentResourceDescriptor().ApplyFilter; applyFilter != nil {
		applyFilter(a, query)
	}
}

func (a *AppModel) clearActiveSearch() bool {
	descriptor := a.currentResourceDescriptor()
	if descriptor.HasActiveFilter != nil && descriptor.HasActiveFilter(a) {
		a.searchInput.SetValue("")
		if descriptor.ApplyFilter != nil {
			descriptor.ApplyFilter(a, "")
		}
		return true
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

func formatProjectSummary(projectCount, totalSessions, activeCount int) string {
	return fmt.Sprintf("%d projects / %d sessions / %d active", projectCount, totalSessions, activeCount)
}

func formatResourceContextLabel(allLabel string, ctx Context) string {
	if ctx.Type == ContextProject {
		return ctx.Value
	}
	return allLabel
}

func formatSkillSummary(width, total, ready, invalid int) string {
	if width < 120 {
		return fmt.Sprintf("%d skills / R:%d I:%d", total, ready, invalid)
	}
	return fmt.Sprintf("%d skills / %d ready / %d invalid", total, ready, invalid)
}

func formatAgentSummary(width, total, ready, invalid int) string {
	if width < 120 {
		return fmt.Sprintf("%d agents / R:%d I:%d", total, ready, invalid)
	}
	return fmt.Sprintf("%d agents / %d ready / %d invalid", total, ready, invalid)
}

func (a *AppModel) helpViewportHeight() int {
	if a.height < 8 {
		bodyHeight := a.height - 1
		if bodyHeight < 1 {
			return 1
		}
		return bodyHeight
	}

	bodyHeight := a.height - headerHeight - TabsHeight - footerHeight
	if bodyHeight < 1 {
		return 1
	}
	return bodyHeight
}

func (a *AppModel) maxHelpScroll() int {
	lines := buildHelpLines(a.resourceRegistry, a.currentResourceDescriptor())
	maxScroll := len(lines) - a.helpViewportHeight()
	if maxScroll < 0 {
		return 0
	}
	return maxScroll
}

func (a *AppModel) renderBodySpinner(width, height int) string {
	if !a.isLoading {
		return ""
	}

	content := fmt.Sprintf("  %s %s", a.spinner.View(), a.loadingText)
	style := lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(lipgloss.Color("252"))
	box := style.Render(content)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}

// renderToast renders the toast notification box (returns empty if expired)
func (a *AppModel) renderToast() string {
	if time.Now().After(a.toastUntil) {
		return ""
	}

	icon := "✓"
	color := lipgloss.Color("42") // Green
	if a.toastIsError {
		icon = "✗"
		color = lipgloss.Color("196") // Red
	}

	content := fmt.Sprintf("%s %s", icon, a.toastMsg)
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Foreground(color).
		BorderForeground(color)

	return style.Render(content)
}
