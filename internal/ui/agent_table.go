package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/kincoy/cc9s/internal/claudefs"
	"github.com/kincoy/cc9s/internal/ui/styles"
)

func renderAgentTable(agents []claudefs.AgentResource, cursor, width, height int, showProjectColumn bool, sortBy AgentSortField, sortAsc bool, contextLabel string) string {
	if len(agents) == 0 {
		return ""
	}

	var sb strings.Builder

	resourceType := styles.HighlightStyle.Render("Agents")
	countStr := lipgloss.NewStyle().
		Foreground(styles.ColorWarning).
		Bold(true).
		Render(fmt.Sprintf("(%d)", len(agents)))

	contextPart := ""
	if contextLabel != "" {
		contextPart = lipgloss.NewStyle().
			Foreground(styles.ColorPurple).
			Bold(true).
			Render(fmt.Sprintf("(%s)", contextLabel))
	}

	title := resourceType + contextPart + countStr
	sb.WriteString(renderTopBorder(width, title))

	contentWidth := width - 4
	showSummary := width >= 80

	projectWidth := 0
	if showProjectColumn {
		projectWidth = 18
	}
	scopeWidth := 10
	nameWidth := 22
	statusWidth := 11
	summaryWidth := 28

	colCount := 3
	if showProjectColumn {
		colCount++
	}
	if showSummary {
		colCount++
	}
	sepCount := colCount - 1
	fixedWidth := projectWidth + scopeWidth + nameWidth + statusWidth
	if showSummary {
		candidate := contentWidth - fixedWidth - (sepCount * 2)
		if candidate >= 20 {
			summaryWidth = candidate
		} else {
			showSummary = false
			colCount--
			sepCount = colCount - 1
		}
	}

	arrow := sortArrow(sortAsc)
	type headerCol struct {
		text    string
		width   int
		align   lipgloss.Position
		sort    AgentSortField
		visible bool
	}
	cols := []headerCol{
		{text: "PROJECT", width: projectWidth, align: lipgloss.Left, sort: -1, visible: showProjectColumn},
		{text: "SCOPE", width: scopeWidth, align: lipgloss.Left, sort: SortByAgentScope, visible: true},
		{text: "NAME", width: nameWidth, align: lipgloss.Left, sort: SortByAgentName, visible: true},
		{text: "SUMMARY", width: summaryWidth, align: lipgloss.Left, sort: -1, visible: showSummary},
		{text: "STATUS", width: statusWidth, align: lipgloss.Left, sort: SortByAgentStatus, visible: true},
	}

	var headerParts []string
	for _, col := range cols {
		if !col.visible {
			continue
		}
		headerParts = append(headerParts, headerSep())
		isSorted := sortBy == col.sort
		text := col.text
		if isSorted {
			text += arrow
		}
		headerParts = append(headerParts, headerStyle(isSorted).Width(col.width).Align(col.align).Render(text))
	}
	headerParts = headerParts[1:]
	sb.WriteString(renderRowBorder(lipgloss.JoinHorizontal(lipgloss.Top, headerParts...)))

	visibleHeight := height - 3
	if visibleHeight < 1 {
		visibleHeight = 1
	}
	startIdx, endIdx := calculateScrollWindow(len(agents), cursor, visibleHeight)

	for i := startIdx; i < endIdx; i++ {
		agent := agents[i]
		rowStyle := styles.TableCellStyle.Faint(true)
		if i == cursor {
			rowStyle = styles.SelectedRowStyle
		}
		rowSep := rowStyle.Render("  ")
		statusStyle := styles.AgentStatusStyle(agent.Status)
		summaryText := agent.Summary
		if summaryText == "" {
			summaryText = "-"
		}

		rowParts := []string{}

		if showProjectColumn {
			rowParts = append(rowParts,
				rowStyle.Width(projectWidth).Render(clampDisplayText(agentProjectLabel(agent), projectWidth)),
				rowSep,
			)
		}

		rowParts = append(rowParts,
			styles.AgentScopeStyle(agent.Scope).Inherit(rowStyle).Width(scopeWidth).Render(styles.AgentScopeText(agent.Scope)),
			rowSep,
			rowStyle.Width(nameWidth).Render(clampDisplayText(agent.Name, nameWidth)),
		)

		if showSummary {
			rowParts = append(rowParts,
				rowSep,
				rowStyle.Width(summaryWidth).Render(truncateSummary(summaryText, summaryWidth)),
			)
		}

		rowParts = append(rowParts,
			rowSep,
			statusStyle.Inherit(rowStyle).Width(statusWidth).Render(styles.AgentStatusText(agent.Status)),
		)

		row := lipgloss.JoinHorizontal(lipgloss.Top, rowParts...)
		row = rowStyle.Width(contentWidth).Render(row)
		sb.WriteString(renderRowBorder(row))
	}

	fillEmptyRows(&sb, contentWidth, endIdx-startIdx, visibleHeight)
	sb.WriteString(renderBottomBorder(width))
	return sb.String()
}

func agentProjectLabel(agent claudefs.AgentResource) string {
	if strings.TrimSpace(agent.ProjectName) != "" {
		return agent.ProjectName
	}
	return "Global"
}
