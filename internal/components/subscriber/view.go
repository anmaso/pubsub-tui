package subscriber

import (
	"fmt"
	"strings"

	"github.com/anmaso/pubsub-tui/internal/components/common"

	"github.com/charmbracelet/lipgloss"
)

// View renders the subscriber panel
func (m Model) View() string {
	// Build title
	title := "4 Subscriber"
	if m.subscriptionName != "" {
		title = fmt.Sprintf("4 Subscriber ← %s", m.subscriptionName)
	}
	if m.MessageCount() > 0 {
		if m.filterText != "" {
			title += fmt.Sprintf(" (%d/%d)", m.DisplayedCount(), m.MessageCount())
		} else {
			title += fmt.Sprintf(" (%d)", m.MessageCount())
		}
	}

	// Calculate dimensions for split view
	// Left: 40%, Right: 60% (matches Publisher panel)
	contentWidth := m.width - 4   // borders
	contentHeight := m.height - 5 // borders + header + filter

	leftWidth := contentWidth * 40 / 100
	if leftWidth < 15 {
		leftWidth = 15
	}
	rightWidth := contentWidth - leftWidth - 1 // separator
	if rightWidth < 15 {
		rightWidth = 15
	}

	// Build header line with auto-ack status and spinner
	var header strings.Builder
	autoAckStatus := "[ ] auto-ack"
	if m.autoAck {
		autoAckStatus = "[✓] auto-ack"
	}
	header.WriteString(common.MutedText.Render(autoAckStatus + " (A)"))

	// Add spinner when connected
	if m.connected {
		header.WriteString("  ")
		header.WriteString(m.spinner.View())
		header.WriteString(" ")
		header.WriteString(common.LogNetworkStyle.Render("listening"))
	}

	// Build left panel (message list)
	leftContent := m.buildLeftPanel(leftWidth, contentHeight)

	// Build right panel (message detail)
	rightContent := m.buildRightPanel(rightWidth, contentHeight)

	// Join panels with separator
	separator := strings.Repeat("│\n", contentHeight)
	separator = strings.TrimSuffix(separator, "\n")

	separatorStyle := lipgloss.NewStyle().Foreground(common.ColorTextMuted)

	mainContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftContent,
		separatorStyle.Render(separator),
		rightContent,
	)

	// Add filter/status line
	var footer string
	if m.filtering {
		footer = m.filterInput.View()
		if m.filterError != nil {
			footer += " " + common.FilterErrorStyle.Render("(invalid regex)")
		}
	} else if m.filterText != "" {
		footer = common.FilterPromptStyle.Render("/ ") + common.FilterInputStyle.Render(m.filterText)
	} else if !m.connected {
		footer = common.MutedText.Render("Select a subscription to start")
	}

	fullContent := header.String() + "\n" + mainContent + "\n" + footer

	return common.BorderedPanel(title, fullContent, m.focused, m.width, m.height)
}

// buildLeftPanel builds the left side with message list
func (m Model) buildLeftPanel(width, height int) string {
	var content strings.Builder

	// Messages header
	messagesHeader := common.MutedText.Render("Messages")
	content.WriteString(messagesHeader)
	content.WriteString("\n")

	// Message list or placeholder
	if !m.connected {
		placeholder := common.MutedText.Render("Not subscribed")
		content.WriteString(placeholder)
		// Pad placeholder
		for i := 0; i < height-2; i++ {
			content.WriteString("\n")
		}
	} else if m.MessageCount() == 0 {
		placeholder := m.spinner.View() + " " + common.MutedText.Render("Waiting for messages...")
		content.WriteString(placeholder)
		// Pad placeholder
		for i := 0; i < height-2; i++ {
			content.WriteString("\n")
		}
	} else {
		// Render the list (includes status bar with pagination)
		content.WriteString(m.messageList.View())
	}

	result := content.String()

	// Pad lines to width only
	lines := strings.Split(result, "\n")
	var paddedLines []string
	for _, line := range lines {
		lineWidth := lipgloss.Width(line)
		if lineWidth < width {
			line = line + strings.Repeat(" ", width-lineWidth)
		}
		paddedLines = append(paddedLines, line)
	}

	return strings.Join(paddedLines, "\n")
}

// buildRightPanel builds the right side with message detail
func (m Model) buildRightPanel(width, height int) string {
	var content strings.Builder

	// Detail header
	detailHeader := common.MutedText.Render("Detail")
	content.WriteString(detailHeader)
	content.WriteString("\n")

	// Detail content
	content.WriteString(m.detailView.View())

	result := content.String()

	// Pad to width
	lines := strings.Split(result, "\n")
	var paddedLines []string
	for _, line := range lines {
		lineWidth := lipgloss.Width(line)
		if lineWidth < width {
			line = line + strings.Repeat(" ", width-lineWidth)
		}
		paddedLines = append(paddedLines, line)
	}

	// Pad to height
	for len(paddedLines) < height {
		paddedLines = append(paddedLines, strings.Repeat(" ", width))
	}

	return strings.Join(paddedLines[:height], "\n")
}

// ShortHelp returns key bindings for the help display
func (m Model) ShortHelp() []string {
	if m.filtering {
		return []string{"esc: clear", "enter: apply"}
	}
	return []string{"/: filter", "a: ack", "A: auto-ack", "j/k: navigate"}
}
