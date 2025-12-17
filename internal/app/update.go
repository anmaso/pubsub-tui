package app

import (
	"context"
	"fmt"

	"github.com/anmaso/pubsub-tui/internal/components/common"
	"github.com/anmaso/pubsub-tui/internal/components/publisher"
	"github.com/anmaso/pubsub-tui/internal/components/subscriber"
	"github.com/anmaso/pubsub-tui/internal/components/subscriptions"
	"github.com/anmaso/pubsub-tui/internal/components/topics"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles messages for the application
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle help popup first
		if m.showHelp {
			// Close help on any key
			m.showHelp = false
			return m, nil
		}

		// Check if any panel has an active input field
		inputActive := m.topics.IsInputActive() ||
			m.subscriptions.IsInputActive() ||
			m.publisher.IsInputActive() ||
			m.subscriber.IsInputActive()

		// Global key handling
		switch {
		case key.Matches(msg, keys.Quit):
			m.stopSubscription()
			return m, tea.Quit

		case key.Matches(msg, keys.Help):
			m.showHelp = true
			return m, nil

		case key.Matches(msg, keys.Tab):
			// Cycle focus forward
			m.cycleFocus()
			return m, nil

		case key.Matches(msg, keys.ShiftTab):
			// Cycle focus backward
			m.cycleFocusReverse()
			return m, nil

		case key.Matches(msg, keys.Panel1) && !inputActive:
			m.focus = FocusTopics
			m.updateFocus()
			return m, nil

		case key.Matches(msg, keys.Panel2) && !inputActive:
			m.focus = FocusSubscriptions
			m.updateFocus()
			return m, nil

		case key.Matches(msg, keys.Panel3) && !inputActive:
			m.focus = FocusPublisher
			m.updateFocus()
			return m, nil

		case key.Matches(msg, keys.Panel4) && !inputActive:
			m.focus = FocusSubscriber
			m.updateFocus()
			return m, nil

		default:
			// Route to focused component
			cmd := m.routeKeyToFocused(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.updateComponentSizes()
		return m, nil

	case common.TopicsLoadedMsg:
		var cmd tea.Cmd
		m.topics, cmd = m.topics.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

		if msg.Err != nil {
			cmds = append(cmds, func() tea.Msg {
				return common.Error(fmt.Sprintf("Failed to load topics: %v", msg.Err))
			})
		} else {
			cmds = append(cmds, func() tea.Msg {
				return common.Success(fmt.Sprintf("Loaded %d topics", len(msg.Topics)))
			})
		}

	case common.SubscriptionsLoadedMsg:
		var cmd tea.Cmd
		m.subscriptions, cmd = m.subscriptions.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

		if msg.Err != nil {
			cmds = append(cmds, func() tea.Msg {
				return common.Error(fmt.Sprintf("Failed to load subscriptions: %v", msg.Err))
			})
		} else {
			cmds = append(cmds, func() tea.Msg {
				return common.Success(fmt.Sprintf("Loaded %d subscriptions", len(msg.Subscriptions)))
			})
		}

	case common.TopicSelectedMsg:
		m.selectedTopic = msg.TopicName

		// Update topics panel with selected indicator
		m.topics.SetSelectedTopic(msg.TopicName)

		// Update subscriptions filter
		m.subscriptions.SetTopicFilter(msg.TopicName)

		// Update publisher target
		m.publisher.SetTargetTopic(msg.TopicName)

		cmds = append(cmds, func() tea.Msg {
			return common.Info(fmt.Sprintf("Selected topic: %s", msg.TopicName))
		})

	case common.SubscriptionSelectedMsg:
		// Stop any existing subscription first
		if m.selectedSubscription != "" && m.selectedSubscription != msg.SubscriptionName {
			m.stopSubscription()
			cmds = append(cmds, func() tea.Msg {
				return common.Info(fmt.Sprintf("Stopped previous subscription: %s", m.selectedSubscription))
			})
		}

		m.selectedSubscription = msg.SubscriptionName

		// Update subscriptions panel with active subscription
		m.subscriptions.SetActiveSubscription(msg.SubscriptionName)

		// Update subscriber - pass message through Update to start spinner
		var cmd tea.Cmd
		m.subscriber, cmd = m.subscriber.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

		// Start subscription stream
		cmd = m.startSubscription(msg.SubscriptionName, msg.TopicName)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

		cmds = append(cmds, func() tea.Msg {
			return common.Network(fmt.Sprintf("Started subscription: %s", msg.SubscriptionName))
		})

	case common.StopSubscriptionMsg:
		// Stop the active subscription
		subName := m.selectedSubscription
		m.stopSubscription()
		m.selectedSubscription = ""

		// Notify both panels
		m.subscriptions.SetActiveSubscription("")
		m.subscriber.ClearSubscription()

		if subName != "" {
			cmds = append(cmds, func() tea.Msg {
				return common.Info(fmt.Sprintf("Stopped subscription: %s", subName))
			})
		}

	// Topic CRUD messages
	case topics.CreateTopicMsg:
		cmds = append(cmds, m.createTopic(msg.TopicName))
		cmds = append(cmds, func() tea.Msg {
			return common.Network(fmt.Sprintf("Creating topic: %s", msg.TopicName))
		})

	case topics.DeleteTopicMsg:
		cmds = append(cmds, m.deleteTopic(msg.TopicName))
		cmds = append(cmds, func() tea.Msg {
			return common.Network(fmt.Sprintf("Deleting topic: %s", msg.TopicName))
		})

	case common.TopicCreatedMsg:
		var cmd tea.Cmd
		m.topics, cmd = m.topics.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		if msg.Err == nil {
			cmds = append(cmds, func() tea.Msg {
				return common.Success(fmt.Sprintf("Created topic: %s", msg.TopicName))
			})
		} else {
			cmds = append(cmds, func() tea.Msg {
				return common.Error(fmt.Sprintf("Failed to create topic: %v", msg.Err))
			})
		}

	case common.TopicDeletedMsg:
		var cmd tea.Cmd
		m.topics, cmd = m.topics.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		if msg.Err == nil {
			cmds = append(cmds, func() tea.Msg {
				return common.Success(fmt.Sprintf("Deleted topic: %s", msg.TopicName))
			})
		} else {
			cmds = append(cmds, func() tea.Msg {
				return common.Error(fmt.Sprintf("Failed to delete topic: %v", msg.Err))
			})
		}

	// Subscription CRUD messages
	case subscriptions.CreateSubscriptionMsg:
		cmds = append(cmds, m.createSubscription(msg.SubscriptionName, msg.TopicName))
		cmds = append(cmds, func() tea.Msg {
			return common.Network(fmt.Sprintf("Creating subscription: %s", msg.SubscriptionName))
		})

	case subscriptions.DeleteSubscriptionMsg:
		cmds = append(cmds, m.deleteSubscription(msg.SubscriptionName))
		cmds = append(cmds, func() tea.Msg {
			return common.Network(fmt.Sprintf("Deleting subscription: %s", msg.SubscriptionName))
		})

	case common.SubscriptionCreatedMsg:
		var cmd tea.Cmd
		m.subscriptions, cmd = m.subscriptions.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		if msg.Err == nil {
			cmds = append(cmds, func() tea.Msg {
				return common.Success(fmt.Sprintf("Created subscription: %s", msg.SubscriptionName))
			})
		} else {
			cmds = append(cmds, func() tea.Msg {
				return common.Error(fmt.Sprintf("Failed to create subscription: %v", msg.Err))
			})
		}

	case common.SubscriptionDeletedMsg:
		var cmd tea.Cmd
		m.subscriptions, cmd = m.subscriptions.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		if msg.Err == nil {
			cmds = append(cmds, func() tea.Msg {
				return common.Success(fmt.Sprintf("Deleted subscription: %s", msg.SubscriptionName))
			})
		} else {
			cmds = append(cmds, func() tea.Msg {
				return common.Error(fmt.Sprintf("Failed to delete subscription: %v", msg.Err))
			})
		}

	// Refresh messages
	case common.RefreshTopicsMsg:
		cmds = append(cmds, m.loadTopics())

	case common.RefreshSubscriptionsMsg:
		cmds = append(cmds, m.loadSubscriptions())

	case publisher.FilesLoadedMsg:
		var cmd tea.Cmd
		m.publisher, cmd = m.publisher.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case publisher.PublishRequestMsg:
		// Execute publish
		cmd := m.publishMessage(msg.Topic, msg.Content)
		cmds = append(cmds, cmd)

	case publisher.PublishResultMsg:
		var cmd tea.Cmd
		m.publisher, cmd = m.publisher.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case subscriber.MessageReceivedMsg:
		var cmd tea.Cmd
		m.subscriber, cmd = m.subscriber.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

		// Continue polling for messages
		if m.activeSubscription != nil {
			cmds = append(cmds, m.pollMessages())
		}

	case subscriber.SubscriptionErrorMsg:
		var cmd tea.Cmd
		m.subscriber, cmd = m.subscriber.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case common.LogMsg:
		var cmd tea.Cmd
		m.activity, cmd = m.activity.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	default:
		// Always update subscriber if connected (for spinner animation)
		// even when not focused
		if m.subscriber.IsConnected() && m.focus != FocusSubscriber {
			var cmd tea.Cmd
			m.subscriber, cmd = m.subscriber.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

		// Always update topics if loading (for spinner animation)
		// even when not focused
		if m.topics.IsLoading() && m.focus != FocusTopics {
			var cmd tea.Cmd
			m.topics, cmd = m.topics.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

		// Always update subscriptions if loading (for spinner animation)
		// even when not focused
		if m.subscriptions.IsLoading() && m.focus != FocusSubscriptions {
			var cmd tea.Cmd
			m.subscriptions, cmd = m.subscriptions.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

		// Route to focused component
		cmd := m.routeToFocused(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// createTopic creates a new topic
func (m *Model) createTopic(topicName string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		err := m.client.CreateTopic(ctx, topicName)
		return common.TopicCreatedMsg{
			TopicName: topicName,
			Err:       err,
		}
	}
}

// deleteTopic deletes a topic
func (m *Model) deleteTopic(topicName string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		err := m.client.DeleteTopic(ctx, topicName)
		return common.TopicDeletedMsg{
			TopicName: topicName,
			Err:       err,
		}
	}
}

// createSubscription creates a new subscription
func (m *Model) createSubscription(subName, topicName string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		err := m.client.CreateSubscription(ctx, subName, topicName)
		return common.SubscriptionCreatedMsg{
			SubscriptionName: subName,
			TopicName:        topicName,
			Err:              err,
		}
	}
}

// deleteSubscription deletes a subscription
func (m *Model) deleteSubscription(subName string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		err := m.client.DeleteSubscription(ctx, subName)
		return common.SubscriptionDeletedMsg{
			SubscriptionName: subName,
			Err:              err,
		}
	}
}

// cycleFocus moves focus to the next panel
func (m *Model) cycleFocus() {
	switch m.focus {
	case FocusTopics:
		m.focus = FocusSubscriptions
	case FocusSubscriptions:
		m.focus = FocusPublisher
	case FocusPublisher:
		m.focus = FocusSubscriber
	case FocusSubscriber:
		m.focus = FocusTopics
	}
	m.updateFocus()
}

// cycleFocusReverse moves focus to the previous panel
func (m *Model) cycleFocusReverse() {
	switch m.focus {
	case FocusTopics:
		m.focus = FocusSubscriber
	case FocusSubscriptions:
		m.focus = FocusTopics
	case FocusPublisher:
		m.focus = FocusSubscriptions
	case FocusSubscriber:
		m.focus = FocusPublisher
	}
	m.updateFocus()
}

// updateFocus updates the focused state of child components
func (m *Model) updateFocus() {
	m.topics.SetFocused(m.focus == FocusTopics)
	m.subscriptions.SetFocused(m.focus == FocusSubscriptions)
	m.publisher.SetFocused(m.focus == FocusPublisher)
	m.subscriber.SetFocused(m.focus == FocusSubscriber)
}

// routeKeyToFocused routes keyboard input to the focused component
func (m *Model) routeKeyToFocused(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd

	switch m.focus {
	case FocusTopics:
		m.topics, cmd = m.topics.Update(msg)
	case FocusSubscriptions:
		m.subscriptions, cmd = m.subscriptions.Update(msg)
	case FocusPublisher:
		m.publisher, cmd = m.publisher.Update(msg)
	case FocusSubscriber:
		m.subscriber, cmd = m.subscriber.Update(msg)
	}

	return cmd
}

// routeToFocused routes any message to the focused component
func (m *Model) routeToFocused(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	switch m.focus {
	case FocusTopics:
		m.topics, cmd = m.topics.Update(msg)
	case FocusSubscriptions:
		m.subscriptions, cmd = m.subscriptions.Update(msg)
	case FocusPublisher:
		m.publisher, cmd = m.publisher.Update(msg)
	case FocusSubscriber:
		m.subscriber, cmd = m.subscriber.Update(msg)
	}

	return cmd
}

// updateComponentSizes recalculates and sets component sizes
func (m *Model) updateComponentSizes() {
	// Left panel: 1/3 width
	// Right panel: 2/3 width
	leftWidth := m.width / 3
	if leftWidth < 25 {
		leftWidth = 25
	}
	rightWidth := m.width - leftWidth
	if rightWidth < 30 {
		rightWidth = 30
	}

	// Available height (minus footer: 2 lines)
	availableHeight := m.height - 2
	if availableHeight < 15 {
		availableHeight = 15
	}

	// Left panel heights: Topics 33%, Subscriptions 33%, Activity 34%
	// Right panel heights: Publisher 33%, Subscriber 67%
	// Topics and Publisher are aligned at 33%
	topicsHeight := availableHeight * 33 / 100
	publisherHeight := topicsHeight // Same height as topics

	subsHeight := availableHeight * 33 / 100
	activityHeight := availableHeight - topicsHeight - subsHeight
	subscriberHeight := availableHeight - publisherHeight

	// Set component sizes
	m.topics.SetSize(leftWidth, topicsHeight)
	m.subscriptions.SetSize(leftWidth, subsHeight)
	m.activity.SetSize(leftWidth, activityHeight)
	m.publisher.SetSize(rightWidth, publisherHeight)
	m.subscriber.SetSize(rightWidth, subscriberHeight)

	// Update focus state
	m.updateFocus()
}

// Key bindings
type keyMap struct {
	Quit     key.Binding
	Tab      key.Binding
	ShiftTab key.Binding
	Panel1   key.Binding
	Panel2   key.Binding
	Panel3   key.Binding
	Panel4   key.Binding
	Help     key.Binding
}

var keys = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next panel"),
	),
	ShiftTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev panel"),
	),
	Panel1: key.NewBinding(
		key.WithKeys("1"),
		key.WithHelp("1", "topics"),
	),
	Panel2: key.NewBinding(
		key.WithKeys("2"),
		key.WithHelp("2", "subscriptions"),
	),
	Panel3: key.NewBinding(
		key.WithKeys("3"),
		key.WithHelp("3", "publisher"),
	),
	Panel4: key.NewBinding(
		key.WithKeys("4"),
		key.WithHelp("4", "subscriber"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
}
