package topics

import (
	"pubsub-tui/internal/components/common"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// CreateTopicMsg requests topic creation
type CreateTopicMsg struct {
	TopicName string
}

// DeleteTopicMsg requests topic deletion
type DeleteTopicMsg struct {
	TopicName string
}

// Update handles messages for the topics panel
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case ModeFilter:
			return m.handleFilterInput(msg)
		case ModeCreate:
			return m.handleCreateInput(msg)
		case ModeConfirmDelete:
			return m.handleConfirmDelete(msg)
		default:
			return m.handleNavigation(msg)
		}

	case common.TopicsLoadedMsg:
		if msg.Err != nil {
			m.SetError(msg.Err)
		} else {
			m.SetTopics(msg.Topics)
		}
		return m, nil

	case common.TopicCreatedMsg:
		if msg.Err != nil {
			m.SetStatus("Create failed: "+msg.Err.Error(), true)
		} else {
			m.SetStatus("Created topic: "+msg.TopicName, false)
			// Request refresh
			cmds = append(cmds, func() tea.Msg {
				return common.RefreshTopicsMsg{}
			})
		}
		return m, tea.Batch(cmds...)

	case common.TopicDeletedMsg:
		if msg.Err != nil {
			m.SetStatus("Delete failed: "+msg.Err.Error(), true)
		} else {
			m.SetStatus("Deleted topic: "+msg.TopicName, false)
			// Request refresh
			cmds = append(cmds, func() tea.Msg {
				return common.RefreshTopicsMsg{}
			})
		}
		return m, tea.Batch(cmds...)
	}

	// Pass other messages to the list
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// handleFilterInput handles keyboard input in filter mode
func (m Model) handleFilterInput(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		// Exit filter mode and clear filter
		m.mode = ModeNormal
		m.filterText = ""
		m.filterInput.SetValue("")
		m.filterError = nil
		m.filterInput.Blur()
		m.applyFilter()
		return m, nil

	case tea.KeyEnter:
		// Exit filter mode but keep filter
		m.mode = ModeNormal
		m.filterInput.Blur()
		return m, nil

	default:
		// Update filter input
		var cmd tea.Cmd
		m.filterInput, cmd = m.filterInput.Update(msg)
		m.filterText = m.filterInput.Value()
		m.applyFilter()
		return m, cmd
	}
}

// handleCreateInput handles keyboard input in create mode
func (m Model) handleCreateInput(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		// Cancel creation
		m.mode = ModeNormal
		m.createInput.SetValue("")
		m.createInput.Blur()
		return m, nil

	case tea.KeyEnter:
		// Submit creation
		topicName := m.createInput.Value()
		if topicName == "" {
			return m, nil
		}

		m.mode = ModeNormal
		m.createInput.SetValue("")
		m.createInput.Blur()

		return m, func() tea.Msg {
			return CreateTopicMsg{TopicName: topicName}
		}

	default:
		// Update create input
		var cmd tea.Cmd
		m.createInput, cmd = m.createInput.Update(msg)
		return m, cmd
	}
}

// handleConfirmDelete handles keyboard input in delete confirmation mode
func (m Model) handleConfirmDelete(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		// Confirm deletion
		topic := m.SelectedTopic()
		if topic == nil {
			m.mode = ModeNormal
			return m, nil
		}

		topicName := topic.Name
		m.mode = ModeNormal

		return m, func() tea.Msg {
			return DeleteTopicMsg{TopicName: topicName}
		}

	case "n", "N", "esc":
		// Cancel deletion
		m.mode = ModeNormal
		return m, nil
	}

	return m, nil
}

// handleNavigation handles keyboard input in normal navigation mode
func (m Model) handleNavigation(msg tea.KeyMsg) (Model, tea.Cmd) {
	// Clear status on any key
	m.ClearStatus()

	switch {
	case key.Matches(msg, keys.Filter):
		// Enter filter mode
		m.mode = ModeFilter
		m.filterInput.Focus()
		return m, nil

	case key.Matches(msg, keys.Create):
		// Enter create mode
		m.mode = ModeCreate
		m.createInput.Focus()
		return m, nil

	case key.Matches(msg, keys.Delete):
		// Enter delete confirmation mode
		if m.SelectedTopic() != nil {
			m.mode = ModeConfirmDelete
		}
		return m, nil

	case key.Matches(msg, keys.Select):
		// Select current topic
		if topic := m.SelectedTopic(); topic != nil {
			return m, func() tea.Msg {
				return common.TopicSelectedMsg{
					TopicName: topic.Name,
					TopicFull: topic.FullName,
				}
			}
		}
		return m, nil

	case key.Matches(msg, keys.Up):
		m.list.CursorUp()
		return m, nil

	case key.Matches(msg, keys.Down):
		m.list.CursorDown()
		return m, nil

	default:
		// Pass to list for handling
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}
}

// Key bindings
type keyMap struct {
	Filter key.Binding
	Create key.Binding
	Delete key.Binding
	Select key.Binding
	Up     key.Binding
	Down   key.Binding
}

var keys = keyMap{
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	Create: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "new"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
}
