package activity

import (
	"fmt"
	"strings"

	"pubsub-tui/internal/components/common"
)

// View renders the activity log panel
func (m Model) View() string {
	var content strings.Builder

	// Build title with count
	title := "Activity"
	if len(m.entries) > 0 {
		title = fmt.Sprintf("Activity (%d)", len(m.entries))
	}

	// Log content
	if len(m.entries) == 0 {
		content.WriteString(common.MutedText.Render("No activity yet"))
	} else {
		content.WriteString(m.viewport.View())
	}

	// Activity log is never focused
	return common.BorderedPanel(title, content.String(), false, m.width, m.height)
}

// renderEntries renders all log entries to a string
func renderEntries(entries []LogEntry, width int) string {
	var sb strings.Builder

	for i, entry := range entries {
		// Timestamp
		timestamp := common.LogTimestampStyle.Render(fmt.Sprintf("[%s]", entry.Time))

		// Message with appropriate color
		style := common.GetLogStyle(entry.Level)
		message := style.Render(entry.Message)

		sb.WriteString(timestamp)
		sb.WriteString(" ")
		sb.WriteString(message)

		if i < len(entries)-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}
