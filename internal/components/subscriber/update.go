package subscriber

import (
	"github.com/anmaso/pubsub-tui/internal/components/common"
	"github.com/anmaso/pubsub-tui/internal/pubsub"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// MessageReceivedMsg is sent when a new message is received
type MessageReceivedMsg struct {
	Message *pubsub.ReceivedMessage
}

// SubscriptionErrorMsg is sent when a subscription error occurs
type SubscriptionErrorMsg struct {
	Error error
}

// StartSubscriptionMsg requests to start a subscription
type StartSubscriptionMsg struct {
	SubscriptionName string
	TopicName        string
}

// StopSubscriptionMsg requests to stop the current subscription
type StopSubscriptionMsg struct{}

// Update handles messages for the subscriber panel
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.filtering {
			return m.handleFilterInput(msg)
		}
		return m.handleNavigation(msg)

	case MessageReceivedMsg:
		m.AddMessage(msg.Message)
		return m, nil

	case SubscriptionErrorMsg:
		return m, func() tea.Msg {
			return common.Error("Subscription error: " + msg.Error.Error())
		}

	case common.SubscriptionSelectedMsg:
		m.SetSubscription(msg.SubscriptionName, msg.TopicName)
		// Start the spinner
		return m, m.spinner.Tick

	case common.SubscriptionStoppedMsg:
		m.ClearSubscription()
		return m, nil
	}

	// Pass other messages to sub-components
	var cmd tea.Cmd

	// Update spinner if connected
	if m.connected {
		m.spinner, cmd = m.spinner.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	m.detailView, cmd = m.detailView.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	m.messageList, cmd = m.messageList.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// handleFilterInput handles keyboard input in filter mode
func (m Model) handleFilterInput(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.filtering = false
		m.filterText = ""
		m.filterInput.SetValue("")
		m.filterError = nil
		m.applyFilter()
		return m, nil

	case tea.KeyEnter:
		m.filtering = false
		return m, nil

	default:
		var cmd tea.Cmd
		m.filterInput, cmd = m.filterInput.Update(msg)
		m.filterText = m.filterInput.Value()
		m.applyFilter()
		return m, cmd
	}
}

// handleNavigation handles keyboard input in normal mode
func (m Model) handleNavigation(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Filter):
		m.filtering = true
		m.filterInput.Focus()
		return m, nil

	case key.Matches(msg, keys.Ack):
		if m.AckSelected() {
			msg := m.SelectedMessage()
			if msg != nil {
				// Move to next message after acknowledging
				m.messageList.CursorDown()
				m.UpdateSelection()
				msgID := msg.ID
				return m, func() tea.Msg {
					return common.Info("Acknowledged message: " + truncateID(msgID))
				}
			}
		}
		return m, nil

	case key.Matches(msg, keys.AutoAck):
		m.ToggleAutoAck()
		status := "disabled"
		if m.autoAck {
			status = "enabled"
		}
		return m, func() tea.Msg {
			return common.Info("Auto-ack " + status)
		}

	case key.Matches(msg, keys.Up):
		m.messageList.CursorUp()
		m.UpdateSelection()
		return m, nil

	case key.Matches(msg, keys.Down):
		m.messageList.CursorDown()
		m.UpdateSelection()
		return m, nil

	case key.Matches(msg, keys.ScrollUp):
		m.detailView.LineUp(3)
		return m, nil

	case key.Matches(msg, keys.ScrollDown):
		m.detailView.LineDown(3)
		return m, nil

	default:
		var cmd tea.Cmd
		m.messageList, cmd = m.messageList.Update(msg)
		m.UpdateSelection()
		return m, cmd
	}
}

// Key bindings
type keyMap struct {
	Filter     key.Binding
	Ack        key.Binding
	AutoAck    key.Binding
	Up         key.Binding
	Down       key.Binding
	ScrollUp   key.Binding
	ScrollDown key.Binding
}

var keys = keyMap{
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	Ack: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "ack"),
	),
	AutoAck: key.NewBinding(
		key.WithKeys("A"),
		key.WithHelp("A", "toggle auto-ack"),
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
		key.WithHelp("ctrl+u", "scroll detail up"),
	),
	ScrollDown: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "scroll detail down"),
	),
}

// truncateID safely truncates a message ID for display
func truncateID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8] + "..."
}
