package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/kincoy/cc9s/internal/claudefs"
	"github.com/kincoy/cc9s/internal/ui/styles"
)

func renderSkillTable(skills []claudefs.SkillResource, cursor, width, height int, sortBy SkillSortField, sortAsc bool, showProjectColumn bool) string {
	if len(skills) == 0 {
		return ""
	}

	var sb strings.Builder

	resourceType := styles.HighlightStyle.Render("Skills")
	countStr := lipgloss.NewStyle().
		Foreground(styles.ColorWarning).
		Bold(true).
		Render(fmt.Sprintf("(%d)", len(skills)))
	title := resourceType + countStr
	sb.WriteString(renderTopBorder(width, title))

	contentWidth := width - 4
	showSummary := width >= 100

	nameWidth := 18
	scopeWidth := 10
	typeWidth := 10
	summaryWidth := 24
	statusWidth := 11

	colCount := 4
	if showSummary {
		colCount++
	}
	sepCount := colCount - 1

	fixedWidth := nameWidth + scopeWidth + typeWidth + statusWidth
	if showSummary {
		candidate := contentWidth - fixedWidth - (sepCount * 2)
		if candidate >= 22 {
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
		sort    SkillSortField
		visible bool
	}
	cols := []headerCol{
		{text: "SCOPE", width: scopeWidth, align: lipgloss.Left, sort: SortBySkillScope, visible: true},
		{text: "NAME", width: nameWidth, align: lipgloss.Left, sort: SortBySkillName, visible: true},
		{text: "TYPE", width: typeWidth, align: lipgloss.Left, sort: SortBySkillType, visible: true},
		{text: "SUMMARY", width: summaryWidth, align: lipgloss.Left, sort: -1, visible: showSummary},
		{text: "STATUS", width: statusWidth, align: lipgloss.Left, sort: SortBySkillStatus, visible: true},
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
	startIdx, endIdx := calculateScrollWindow(len(skills), cursor, visibleHeight)

	for i := startIdx; i < endIdx; i++ {
		skill := skills[i]
		rowStyle := styles.TableCellStyle
		if i == cursor {
			rowStyle = styles.SelectedRowStyle
		}
		rowSep := rowStyle.Render("  ")
		statusStyle := styles.SkillStatusStyle(skill.Status)
		summaryText := skill.Summary
		if summaryText == "" {
			summaryText = "-"
		}

		rowParts := []string{}

		rowParts = append(rowParts,
			styles.SkillScopeStyle(skill.Scope).Inherit(rowStyle).Width(scopeWidth).Render(styles.SkillScopeText(skill.Scope)),
			rowSep,
			rowStyle.Width(nameWidth).Render(clampDisplayText(skill.Name, nameWidth)),
			rowSep,
			styles.SkillKindStyle(skill.Kind).Inherit(rowStyle).Width(typeWidth).Render(styles.SkillKindText(skill.Kind)),
		)

		if showSummary {
			rowParts = append(rowParts,
				rowSep,
				rowStyle.Width(summaryWidth).Render(truncateSummary(summaryText, summaryWidth)),
			)
		}

		rowParts = append(rowParts,
			rowSep,
			statusStyle.Inherit(rowStyle).Width(statusWidth).Render(styles.SkillStatusText(skill.Status)),
		)

		sb.WriteString(renderRowBorder(lipgloss.JoinHorizontal(lipgloss.Top, rowParts...)))
	}

	fillEmptyRows(&sb, contentWidth, endIdx-startIdx, visibleHeight)
	sb.WriteString(renderBottomBorder(width))
	return sb.String()
}

func clampDisplayText(s string, maxWidth int) string {
	if lipgloss.Width(s) <= maxWidth {
		return s
	}
	return truncateSummary(s, maxWidth)
}
