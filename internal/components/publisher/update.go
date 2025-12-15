package publisher

import (
	"github.com/anmaso/pubsub-tui/internal/components/common"
	"github.com/anmaso/pubsub-tui/internal/utils"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// FilesLoadedMsg is sent when JSON files are loaded
type FilesLoadedMsg struct {
	Files []utils.JSONFile
	Err   error
}

// PublishRequestMsg requests a publish operation
type PublishRequestMsg struct {
	Topic   string
	Content []byte
}

// PublishResultMsg is sent when a publish operation completes
type PublishResultMsg struct {
	MessageID string
	Err       error
}

// Update handles messages for the publisher panel
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.focusArea == FocusVariables {
			return m.handleVariablesInput(msg)
		}
		return m.handleNavigation(msg)

	case FilesLoadedMsg:
		if msg.Err != nil {
			m.SetStatus("Error loading files: "+msg.Err.Error(), true)
		} else {
			m.SetFiles(msg.Files)
			if len(msg.Files) == 0 {
				m.SetStatus("No JSON files found in current directory", false)
			}
		}
		return m, nil

	case PublishResultMsg:
		m.SetPublishing(false)
		if msg.Err != nil {
			m.SetStatus("Publish failed: "+msg.Err.Error(), true)
			return m, func() tea.Msg {
				return common.Error("Publish failed: " + msg.Err.Error())
			}
		}
		m.SetStatus("Published: "+msg.MessageID, false)
		return m, func() tea.Msg {
			return common.Success("Published message: " + msg.MessageID)
		}

	case common.TopicSelectedMsg:
		m.SetTargetTopic(msg.TopicName)
		return m, nil
	}

	// Pass other messages to sub-components
	switch m.focusArea {
	case FocusFileList:
		var cmd tea.Cmd
		m.fileList, cmd = m.fileList.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// handleVariablesInput handles keyboard input when editing variables
func (m Model) handleVariablesInput(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		// Publish message on Enter
		return m.triggerPublish()

	case tea.KeyEsc:
		// Exit variables input mode
		m.focusArea = FocusFileList
		m.variablesInput.Blur()
		return m, nil

	case tea.KeyTab:
		// Move back to file list
		m.focusArea = FocusFileList
		m.variablesInput.Blur()
		return m, nil

	default:
		// Update variables input
		var cmd tea.Cmd
		m.variablesInput, cmd = m.variablesInput.Update(msg)
		m.updatePreview()
		return m, cmd
	}
}

// handleNavigation handles keyboard input in normal mode
func (m Model) handleNavigation(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Variables):
		// Focus variables input
		m.focusArea = FocusVariables
		m.variablesInput.Focus()
		return m, nil

	case key.Matches(msg, keys.Publish):
		return m.triggerPublish()

	case key.Matches(msg, keys.Select):
		// Select current file from list
		if item := m.fileList.SelectedItem(); item != nil {
			if fileItem, ok := item.(FileItem); ok {
				for i := range m.allFiles {
					if m.allFiles[i].Path == fileItem.path {
						m.selectFile(&m.allFiles[i])
						break
					}
				}
			}
		}
		return m, nil

	case key.Matches(msg, keys.Up):
		m.fileList.CursorUp()
		// Auto-select file on navigation
		if item := m.fileList.SelectedItem(); item != nil {
			if fileItem, ok := item.(FileItem); ok {
				for i := range m.allFiles {
					if m.allFiles[i].Path == fileItem.path {
						m.selectFile(&m.allFiles[i])
						break
					}
				}
			}
		}
		return m, nil

	case key.Matches(msg, keys.Down):
		m.fileList.CursorDown()
		// Auto-select file on navigation
		if item := m.fileList.SelectedItem(); item != nil {
			if fileItem, ok := item.(FileItem); ok {
				for i := range m.allFiles {
					if m.allFiles[i].Path == fileItem.path {
						m.selectFile(&m.allFiles[i])
						break
					}
				}
			}
		}
		return m, nil

	case key.Matches(msg, keys.ScrollUp):
		m.preview.LineUp(1)
		return m, nil

	case key.Matches(msg, keys.ScrollDown):
		m.preview.LineDown(1)
		return m, nil

	default:
		// Pass to focused component
		var cmd tea.Cmd
		m.fileList, cmd = m.fileList.Update(msg)
		return m, cmd
	}
}

// triggerPublish initiates a publish operation
func (m Model) triggerPublish() (Model, tea.Cmd) {
	if m.targetTopic == "" {
		m.SetStatus("No topic selected", true)
		return m, nil
	}

	if m.selectedFile == nil {
		m.SetStatus("No file selected", true)
		return m, nil
	}

	if m.publishing {
		return m, nil
	}

	content := m.GetMessageContent()
	if content == "" {
		m.SetStatus("No content to publish", true)
		return m, nil
	}

	m.SetPublishing(true)
	m.SetStatus("Publishing...", false)

	return m, func() tea.Msg {
		return PublishRequestMsg{
			Topic:   m.targetTopic,
			Content: []byte(content),
		}
	}
}

// LoadFiles creates a command to load JSON files
func LoadFiles() tea.Cmd {
	return func() tea.Msg {
		files, err := utils.ListJSONFiles("")
		return FilesLoadedMsg{Files: files, Err: err}
	}
}

// Key bindings
type keyMap struct {
	Variables  key.Binding
	Publish    key.Binding
	Select     key.Binding
	Up         key.Binding
	Down       key.Binding
	ScrollUp   key.Binding
	ScrollDown key.Binding
}

var keys = keyMap{
	Variables: key.NewBinding(
		key.WithKeys("v"),
		key.WithHelp("v", "variables"),
	),
	Publish: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "publish"),
	),
	Select: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "select file"),
	),
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	ScrollUp: key.NewBinding(
		key.WithKeys("ctrl+u"),
		key.WithHelp("ctrl+u", "scroll preview up"),
	),
	ScrollDown: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "scroll preview down"),
	),
}
