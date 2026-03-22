package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/kincoy/cc9s/internal/ui/styles"
)

// ConfirmDialogModel confirm dialog Model (also supports alert mode)
type ConfirmDialogModel struct {
	title     string
	message   []string
	yesLabel  string
	noLabel   string
	onConfirm tea.Cmd // command executed on y key
	onCancel  tea.Cmd // command executed on n/Esc/any key
	alert     bool    // when true, acts as alert (any key closes, no y button shown)
}

// NewConfirmDialogModel creates a confirm dialog
func NewConfirmDialogModel(title string, message []string, onConfirm, onCancel tea.Cmd) *ConfirmDialogModel {
	return &ConfirmDialogModel{
		title:     title,
		message:   message,
		yesLabel:  "y",
		noLabel:   "n",
		onConfirm: onConfirm,
		onCancel:  onCancel,
	}
}

// NewAlertDialogModel creates an alert dialog (any key or Esc to close)
func NewAlertDialogModel(title string, message []string) *ConfirmDialogModel {
	return &ConfirmDialogModel{
		title:    title,
		message:  message,
		alert:    true,
		onCancel: closeDialogCmd(),
	}
}

// Update handles key presses within the dialog
func (m *ConfirmDialogModel) Update(msg tea.Msg) tea.Cmd {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		if m.alert {
			// alert mode: any key or Esc closes
			return m.onCancel
		}
		switch keyMsg.String() {
		case "y":
			return m.onConfirm
		case "n", "esc":
			return m.onCancel
		}
	}
	return nil
}

// View renders the dialog (fullscreen centered, for standalone rendering)
func (m *ConfirmDialogModel) View(width, height int) string {
	box := m.ViewBox(width)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}

// ViewBox renders the dialog box body (without centering whitespace, for overlaying on background)
func (m *ConfirmDialogModel) ViewBox(width int) string {
	// Build content
	var lines []string

	// Title
	lines = append(lines, styles.DialogTitleStyle.Render(m.title))
	lines = append(lines, "")

	// Message body
	for _, msg := range m.message {
		lines = append(lines, styles.DialogMessageStyle.Render(msg))
	}

	lines = append(lines, "")

	// Buttons
	if m.alert {
		closeButton := styles.FooterStyle.Render("[") +
			styles.DialogButtonStyle.Render("Esc") +
			styles.FooterStyle.Render("] Close")
		lines = append(lines, closeButton)
	} else {
		noButton := styles.FooterStyle.Render("[") +
			styles.DialogButtonStyle.Render(m.noLabel) +
			styles.FooterStyle.Render("] Cancel")

		yesButton := styles.FooterStyle.Render("[") +
			styles.DialogButtonStyle.Render(m.yesLabel) +
			styles.FooterStyle.Render("] Confirm")

		buttons := noButton + "    " + yesButton
		lines = append(lines, buttons)
	}

	content := strings.Join(lines, "\n")

	// Draw border
	boxWidth := 55
	if boxWidth > width {
		boxWidth = width
	}
	borderColor := styles.ColorWarning
	if m.alert {
		borderColor = styles.ColorError
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2).
		Width(boxWidth).
		Render(content)

	return box
}
