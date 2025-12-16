package publisher

import (
	"fmt"
	"strings"

	"github.com/anmaso/pubsub-tui/internal/components/common"

	"github.com/charmbracelet/lipgloss"
)

// View renders the publisher panel
func (m Model) View() string {
	// Build title
	title := "3 Publisher"
	if m.targetTopic != "" {
		title = fmt.Sprintf("3 Publisher → %s", m.targetTopic)
	}

	// Calculate dimensions for split view
	contentWidth := m.width - 4   // borders
	contentHeight := m.height - 4 // borders + status

	leftWidth := contentWidth * 40 / 100
	if leftWidth < 15 {
		leftWidth = 15
	}
	rightWidth := contentWidth - leftWidth - 1 // separator
	if rightWidth < 15 {
		rightWidth = 15
	}

	// Build left panel (files + variables)
	leftContent := m.buildLeftPanel(leftWidth, contentHeight)

	// Build right panel (preview)
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

	// Add status line
	var status string
	if m.status != "" {
		style := common.LogSuccessStyle
		if m.statusError {
			style = common.LogErrorStyle
		}
		status = style.Render(m.status)
	} else if m.focusArea == FocusVariables {
		status = common.MutedText.Render("Enter: publish  Esc: exit  Tab: files")
	} else {
		status = common.MutedText.Render("Enter: publish  v: variables  j/k: navigate")
	}

	fullContent := mainContent + "\n" + status

	return common.BorderedPanel(title, fullContent, m.focused, m.width, m.height)
}

// buildLeftPanel builds the left side with files and variables
func (m Model) buildLeftPanel(width, height int) string {
	var content strings.Builder

	// Files section header
	filesHeader := common.MutedText.Render("Files")
	if m.focusArea == FocusFileList && m.focused {
		filesHeader = common.FilterPromptStyle.Render("Files")
	}
	if len(m.allFiles) > 0 {
		filesHeader += common.MutedText.Render(fmt.Sprintf(" (%d)", len(m.allFiles)))
	}
	content.WriteString(filesHeader)
	content.WriteString("\n")

	// File list
	if len(m.allFiles) == 0 {
		content.WriteString(common.MutedText.Render("No JSON files"))
	} else {
		content.WriteString(m.fileList.View())
	}

	// Variables section
	content.WriteString("\n")
	varsHeader := common.MutedText.Render("Variables (v)")
	if m.focusArea == FocusVariables && m.focused {
		varsHeader = common.FilterPromptStyle.Render("Variables")
	}
	content.WriteString(varsHeader)
	content.WriteString("\n")
	content.WriteString(m.variablesInput.View())

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

// buildRightPanel builds the right side with preview
func (m Model) buildRightPanel(width, height int) string {
	var content strings.Builder

	// Preview header
	previewHeader := common.MutedText.Render("Preview")
	if m.selectedFile != nil {
		previewHeader += common.MutedText.Render(fmt.Sprintf(" - %s", m.selectedFile.Name))
	}
	content.WriteString(previewHeader)
	content.WriteString("\n")

	// Preview content
	if m.previewContent != "" {
		content.WriteString(m.preview.View())
	} else if m.selectedFile != nil {
		content.WriteString(common.MutedText.Render("Loading..."))
	} else {
		content.WriteString(common.MutedText.Render("Select a file"))
	}

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
	if m.focusArea == FocusVariables {
		return []string{"esc: back", "tab: files"}
	}
	return []string{"enter: publish", "v: variables", "j/k: navigate"}
}
