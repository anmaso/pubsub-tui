package subscriptions

import (
	"github.com/anmaso/pubsub-tui/internal/components/common"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// CreateSubscriptionMsg requests subscription creation
type CreateSubscriptionMsg struct {
	SubscriptionName string
	TopicName        string
}

// DeleteSubscriptionMsg requests subscription deletion
type DeleteSubscriptionMsg struct {
	SubscriptionName string
}

// Update handles messages for the subscriptions panel
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

	case common.SubscriptionsLoadedMsg:
		if msg.Err != nil {
			m.SetError(msg.Err)
		} else {
			m.SetSubscriptions(msg.Subscriptions)
		}
		return m, nil

	case common.TopicSelectedMsg:
		// Filter to selected topic
		m.SetTopicFilter(msg.TopicName)
		return m, nil

	case common.SubscriptionCreatedMsg:
		if msg.Err != nil {
			m.SetStatus("Create failed: "+msg.Err.Error(), true)
		} else {
			m.SetStatus("Created subscription: "+msg.SubscriptionName, false)
			// Request refresh
			cmds = append(cmds, func() tea.Msg {
				return common.RefreshSubscriptionsMsg{}
			})
		}
		return m, tea.Batch(cmds...)

	case common.SubscriptionDeletedMsg:
		if msg.Err != nil {
			m.SetStatus("Delete failed: "+msg.Err.Error(), true)
		} else {
			m.SetStatus("Deleted subscription: "+msg.SubscriptionName, false)
			// Request refresh
			cmds = append(cmds, func() tea.Msg {
				return common.RefreshSubscriptionsMsg{}
			})
		}
		return m, tea.Batch(cmds...)

	case common.SubscriptionStoppedMsg:
		m.activeSubscription = ""
		return m, nil
	}

	// Update spinner if loading
	if m.loading {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
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
		subName := m.createInput.Value()
		if subName == "" {
			return m, nil
		}

		if m.selectedTopic == "" {
			m.SetStatus("Select a topic first", true)
			m.mode = ModeNormal
			m.createInput.SetValue("")
			m.createInput.Blur()
			return m, nil
		}

		topicName := m.selectedTopic
		m.mode = ModeNormal
		m.createInput.SetValue("")
		m.createInput.Blur()

		return m, func() tea.Msg {
			return CreateSubscriptionMsg{
				SubscriptionName: subName,
				TopicName:        topicName,
			}
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
		sub := m.SelectedSubscription()
		if sub == nil {
			m.mode = ModeNormal
			return m, nil
		}

		subName := sub.Name
		m.mode = ModeNormal

		return m, func() tea.Msg {
			return DeleteSubscriptionMsg{SubscriptionName: subName}
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
	case key.Matches(msg, keys.Stop):
		// Stop active subscription
		if m.activeSubscription != "" {
			return m, func() tea.Msg {
				return common.StopSubscriptionMsg{}
			}
		}
		return m, nil

	case key.Matches(msg, keys.Filter):
		// Enter filter mode
		m.mode = ModeFilter
		m.filterInput.Focus()
		return m, nil

	case key.Matches(msg, keys.Create):
		// Enter create mode (requires topic selection)
		if m.selectedTopic == "" {
			m.SetStatus("Select a topic first (go to Topics panel)", true)
			return m, nil
		}
		m.mode = ModeCreate
		m.createInput.Focus()
		return m, nil

	case key.Matches(msg, keys.Delete):
		// Enter delete confirmation mode
		if m.SelectedSubscription() != nil {
			m.mode = ModeConfirmDelete
		}
		return m, nil

	case key.Matches(msg, keys.ClearFilter):
		// Clear topic filter
		m.ClearTopicFilter()
		return m, nil

	case key.Matches(msg, keys.Select):
		// Select current subscription or disconnect if already active
		if sub := m.SelectedSubscription(); sub != nil {
			if m.IsActiveSubscription(sub.Name) {
				// Already connected - disconnect
				return m, func() tea.Msg {
					return common.StopSubscriptionMsg{}
				}
			}
			// Connect to this subscription
			return m, func() tea.Msg {
				return common.SubscriptionSelectedMsg{
					SubscriptionName: sub.Name,
					SubscriptionFull: sub.FullName,
					TopicName:        sub.TopicName,
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
	Stop        key.Binding
	Filter      key.Binding
	ClearFilter key.Binding
	Create      key.Binding
	Delete      key.Binding
	Select      key.Binding
	Up          key.Binding
	Down        key.Binding
}

var keys = keyMap{
	Stop: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "stop"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	ClearFilter: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "clear topic filter"),
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
