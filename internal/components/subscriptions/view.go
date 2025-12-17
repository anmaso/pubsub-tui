package subscriptions

import (
	"fmt"
	"strings"

	"github.com/anmaso/pubsub-tui/internal/components/common"
)

// View renders the subscriptions panel
func (m Model) View() string {
	var content strings.Builder

	// Build title with count
	title := "2 Subscriptions"
	if len(m.allSubscriptions) > 0 {
		displayed := m.DisplayCount()
		total := m.TotalCount()

		if m.selectedTopic != "" || m.filterText != "" {
			title = fmt.Sprintf("2 Subscriptions (%d/%d)", displayed, total)
		} else {
			title = fmt.Sprintf("2 Subscriptions (%d)", total)
		}
	}

	// Topic filter indicator
	if m.selectedTopic != "" {
		topicIndicator := common.FilterPromptStyle.Render("Topic: ") +
			common.BrightText.Render(m.selectedTopic) +
			common.MutedText.Render(" (c to clear)")
		content.WriteString(topicIndicator)
		content.WriteString("\n")
	} else {
		content.WriteString(common.MutedText.Render("All topics"))
		content.WriteString("\n")
	}

	// Main content area
	if m.loading {
		content.WriteString(m.spinner.View())
		content.WriteString(" ")
		content.WriteString(common.LogNetworkStyle.Render("Loading subscriptions..."))
	} else if m.loadError != nil {
		content.WriteString(common.LogErrorStyle.Render(fmt.Sprintf("Error: %v", m.loadError)))
	} else if len(m.allSubscriptions) == 0 {
		content.WriteString(common.MutedText.Render("No subscriptions found"))
	} else if m.DisplayCount() == 0 {
		content.WriteString(common.MutedText.Render("No matching subscriptions"))
	} else {
		content.WriteString(m.list.View())
	}

	// Bottom area based on mode
	content.WriteString("\n")

	switch m.mode {
	case ModeFilter:
		content.WriteString(m.filterInput.View())
		if m.filterError != nil {
			content.WriteString(" ")
			content.WriteString(common.FilterErrorStyle.Render("(invalid regex)"))
		}

	case ModeCreate:
		content.WriteString(m.createInput.View())
		content.WriteString("\n")
		content.WriteString(common.MutedText.Render(fmt.Sprintf("Creating for topic: %s", m.selectedTopic)))

	case ModeConfirmDelete:
		if sub := m.SelectedSubscription(); sub != nil {
			content.WriteString(common.LogWarningStyle.Render(fmt.Sprintf("Delete '%s'? (y/n)", sub.Name)))
		}

	default:
		// Show status or active filter
		if m.statusMsg != "" {
			style := common.LogSuccessStyle
			if m.statusError {
				style = common.LogErrorStyle
			}
			content.WriteString(style.Render(m.statusMsg))
		} else if m.filterText != "" {
			filterDisplay := common.FilterPromptStyle.Render("/ ") +
				common.FilterInputStyle.Render(m.filterText)
			content.WriteString(filterDisplay)
		}
	}

	return common.BorderedPanel(title, content.String(), m.focused, m.width, m.height)
}

// ShortHelp returns key bindings for the help display
func (m Model) ShortHelp() []string {
	switch m.mode {
	case ModeFilter:
		return []string{"esc: clear", "enter: apply"}
	case ModeCreate:
		return []string{"enter: create", "esc: cancel"}
	case ModeConfirmDelete:
		return []string{"y: yes", "n: no"}
	default:
		help := []string{"/: filter", "n: new", "d: delete", "enter: select"}
		if m.selectedTopic != "" {
			help = append(help, "c: clear topic")
		}
		return help
	}
}
