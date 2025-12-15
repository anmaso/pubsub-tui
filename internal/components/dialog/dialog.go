package dialog

import (
	"strings"

	"github.com/anmaso/pubsub-tui/internal/components/common"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// Type represents the type of dialog
type Type int

const (
	TypeConfirm Type = iota // Yes/No confirmation
	TypeInput               // Text input
)

// Result represents the result of a dialog
type Result struct {
	Confirmed bool
	Value     string
	Context   interface{} // Optional context data
}

// ResultMsg is sent when a dialog is completed
type ResultMsg struct {
	ID     string // Dialog identifier
	Result Result
}

// Model represents a dialog box
type Model struct {
	id          string
	dialogType  Type
	title       string
	message     string
	input       textinput.Model
	visible     bool
	context     interface{}
	placeholder string
}

// New creates a new dialog model
func New() Model {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 255
	ti.Width = 40

	return Model{
		input: ti,
	}
}

// ShowConfirm shows a confirmation dialog
func (m *Model) ShowConfirm(id, title, message string, context interface{}) {
	m.id = id
	m.dialogType = TypeConfirm
	m.title = title
	m.message = message
	m.visible = true
	m.context = context
}

// ShowInput shows an input dialog
func (m *Model) ShowInput(id, title, message, placeholder string, context interface{}) {
	m.id = id
	m.dialogType = TypeInput
	m.title = title
	m.message = message
	m.placeholder = placeholder
	m.visible = true
	m.context = context
	m.input.SetValue("")
	m.input.Placeholder = placeholder
	m.input.Focus()
}

// Hide hides the dialog
func (m *Model) Hide() {
	m.visible = false
}

// IsVisible returns whether the dialog is visible
func (m Model) IsVisible() bool {
	return m.visible
}

// ID returns the dialog ID
func (m Model) ID() string {
	return m.id
}

// View renders the dialog
func (m Model) View(width, height int) string {
	if !m.visible {
		return ""
	}

	var content strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(common.ColorPrimary).
		Bold(true)
	content.WriteString(titleStyle.Render(m.title))
	content.WriteString("\n\n")

	// Message
	content.WriteString(common.NormalText.Render(m.message))
	content.WriteString("\n\n")

	// Input or buttons
	if m.dialogType == TypeInput {
		content.WriteString(m.input.View())
		content.WriteString("\n\n")
		content.WriteString(common.MutedText.Render("Enter: confirm  Esc: cancel"))
	} else {
		content.WriteString(common.MutedText.Render("y: yes  n/Esc: no"))
	}

	dialogContent := content.String()

	// Dialog box style
	dialogWidth := 50

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(common.ColorPrimary).
		Padding(1, 2).
		Width(dialogWidth)

	dialog := boxStyle.Render(dialogContent)

	// Center the dialog
	dialogRendered := lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		dialog,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("0")),
	)

	// Add semi-transparent overlay effect by dimming
	return dialogRendered
}

// Update handles input for the dialog
func (m Model) Update(keyStr string, keyType string) (Model, *ResultMsg) {
	if !m.visible {
		return m, nil
	}

	if m.dialogType == TypeConfirm {
		switch keyStr {
		case "y", "Y":
			m.visible = false
			return m, &ResultMsg{
				ID: m.id,
				Result: Result{
					Confirmed: true,
					Context:   m.context,
				},
			}
		case "n", "N", "esc":
			m.visible = false
			return m, &ResultMsg{
				ID: m.id,
				Result: Result{
					Confirmed: false,
					Context:   m.context,
				},
			}
		}
	} else {
		switch keyType {
		case "enter":
			value := m.input.Value()
			if value != "" {
				m.visible = false
				return m, &ResultMsg{
					ID: m.id,
					Result: Result{
						Confirmed: true,
						Value:     value,
						Context:   m.context,
					},
				}
			}
		case "esc":
			m.visible = false
			return m, &ResultMsg{
				ID: m.id,
				Result: Result{
					Confirmed: false,
					Context:   m.context,
				},
			}
		}
	}

	return m, nil
}

// UpdateInput updates the text input
func (m *Model) UpdateInput(msg interface{}) {
	if m.dialogType == TypeInput && m.visible {
		m.input, _ = m.input.Update(msg)
	}
}

