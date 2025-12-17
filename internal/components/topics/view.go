package topics

import (
	"fmt"
	"strings"

	"github.com/anmaso/pubsub-tui/internal/components/common"
)

// View renders the topics panel
func (m Model) View() string {
	var content strings.Builder

	// Build title with count
	title := "1 Topics"
	if len(m.allTopics) > 0 {
		if m.filterText != "" {
			title = fmt.Sprintf("1 Topics (%d/%d)", len(m.list.Items()), len(m.allTopics))
		} else {
			title = fmt.Sprintf("1 Topics (%d)", len(m.allTopics))
		}
	}

	// Main content area
	if m.loading {
		content.WriteString(m.spinner.View())
		content.WriteString(" ")
		content.WriteString(common.LogNetworkStyle.Render("Loading topics..."))
	} else if m.loadError != nil {
		content.WriteString(common.LogErrorStyle.Render(fmt.Sprintf("Error: %v", m.loadError)))
	} else if len(m.allTopics) == 0 {
		content.WriteString(common.MutedText.Render("No topics found"))
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
		content.WriteString(common.MutedText.Render("Enter: create  Esc: cancel"))

	case ModeConfirmDelete:
		if topic := m.SelectedTopic(); topic != nil {
			content.WriteString(common.LogWarningStyle.Render(fmt.Sprintf("Delete '%s'? (y/n)", topic.Name)))
		}

	default:
		// Show status or help
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
		} else {
			content.WriteString(common.MutedText.Render("/ filter  n new  d delete"))
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
		return []string{"/: filter", "n: new", "d: delete", "enter: select"}
	}
}
