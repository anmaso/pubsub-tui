package subscriber

import (
	"fmt"
	"time"

	"github.com/anmaso/pubsub-tui/internal/components/common"
	"github.com/anmaso/pubsub-tui/internal/pubsub"
	"github.com/anmaso/pubsub-tui/internal/utils"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
)

// MessageItem implements list.Item for displaying messages
type MessageItem struct {
	message *pubsub.ReceivedMessage
}

func (m MessageItem) Title() string {
	ackMark := "○"
	if m.message.IsAcked() {
		ackMark = "✓"
	}
	// Show first 8 chars of ID
	shortID := m.message.ID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}
	timeStr := m.message.PublishTime.Format("15:04:05")
	return fmt.Sprintf("[%s] %s %s", ackMark, shortID, timeStr)
}

func (m MessageItem) Description() string {
	// Show first 40 chars of data
	data := string(m.message.Data)
	if len(data) > 40 {
		data = data[:40] + "..."
	}
	return data
}

func (m MessageItem) FilterValue() string {
	return m.message.ID + string(m.message.Data)
}

// Model represents the state of the subscriber panel
type Model struct {
	messageList list.Model
	filterInput textinput.Model
	detailView  viewport.Model

	messages        []*pubsub.ReceivedMessage
	selectedMessage *pubsub.ReceivedMessage

	width   int
	height  int
	focused bool

	filtering  bool
	filterText string
	filterError error
	autoAck    bool

	subscriptionName string
	topicName        string
	connected        bool
}

// New creates a new subscriber panel model
func New() Model {
	// Create message list with compact style
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.SetSpacing(0) // Compact spacing between items
	delegate.Styles.SelectedTitle = common.SelectedItem
	delegate.Styles.NormalTitle = common.NormalText
	delegate.Styles.NormalDesc = common.MutedText
	delegate.Styles.SelectedDesc = common.MutedText

	ml := list.New([]list.Item{}, delegate, 0, 0)
	ml.Title = "Messages"
	ml.SetShowTitle(false)
	ml.SetShowStatusBar(true) // Show pagination info
	ml.SetShowHelp(false)
	ml.SetFilteringEnabled(false)
	ml.DisableQuitKeybindings()
	
	// Customize status bar styles
	ml.Styles.StatusBar = common.MutedText
	ml.Styles.StatusEmpty = common.MutedText

	// Create filter input
	fi := textinput.New()
	fi.Placeholder = "regex filter..."
	fi.Prompt = "/ "
	fi.PromptStyle = common.FilterPromptStyle
	fi.TextStyle = common.FilterInputStyle

	// Create detail viewport
	dv := viewport.New(0, 0)

	return Model{
		messageList: ml,
		filterInput: fi,
		detailView:  dv,
		messages:    make([]*pubsub.ReceivedMessage, 0, 100),
	}
}

// SetFocused sets whether the panel is focused
func (m *Model) SetFocused(focused bool) {
	m.focused = focused
}

// IsFocused returns whether the panel is focused
func (m Model) IsFocused() bool {
	return m.focused
}

// SetSize sets the panel dimensions
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Split: left side for message list, right side for detail
	// Left: 40%, Right: 60% (matches Publisher panel)
	contentHeight := height - 5 // borders, header, filter
	if contentHeight < 4 {
		contentHeight = 4
	}

	contentWidth := width - 4
	leftWidth := contentWidth * 40 / 100
	if leftWidth < 15 {
		leftWidth = 15
	}
	rightWidth := contentWidth - leftWidth - 1 // separator
	if rightWidth < 15 {
		rightWidth = 15
	}

	m.messageList.SetSize(leftWidth, contentHeight)
	m.detailView.Width = rightWidth
	m.detailView.Height = contentHeight
}

// SetSubscription sets the active subscription
func (m *Model) SetSubscription(name, topic string) {
	m.subscriptionName = name
	m.topicName = topic
	m.connected = true
	m.messages = make([]*pubsub.ReceivedMessage, 0, 100)
	m.selectedMessage = nil
	m.applyFilter()
	m.updateDetailView()
}

// ClearSubscription clears the active subscription
func (m *Model) ClearSubscription() {
	m.subscriptionName = ""
	m.topicName = ""
	m.connected = false
	m.messages = make([]*pubsub.ReceivedMessage, 0, 100)
	m.selectedMessage = nil
	m.messageList.SetItems([]list.Item{})
	m.updateDetailView()
}

// AddMessage adds a new message to the list
func (m *Model) AddMessage(msg *pubsub.ReceivedMessage) {
	// Auto-ack if enabled
	if m.autoAck {
		msg.Ack()
	}

	// Append to list (newest last)
	m.messages = append(m.messages, msg)

	// Cap at 100 messages
	if len(m.messages) > 100 {
		m.messages = m.messages[1:]
	}

	m.applyFilter()

	// Auto-select newest message
	m.selectedMessage = msg
	m.updateDetailView()

	// Move cursor to bottom
	m.messageList.Select(len(m.messageList.Items()) - 1)
}

// SelectedMessage returns the currently selected message
func (m Model) SelectedMessage() *pubsub.ReceivedMessage {
	if m.messageList.SelectedItem() == nil {
		return nil
	}

	item, ok := m.messageList.SelectedItem().(MessageItem)
	if !ok {
		return nil
	}

	return item.message
}

// ToggleAutoAck toggles auto-acknowledgment
func (m *Model) ToggleAutoAck() {
	m.autoAck = !m.autoAck
}

// IsAutoAck returns whether auto-ack is enabled
func (m Model) IsAutoAck() bool {
	return m.autoAck
}

// IsConnected returns whether connected to a subscription
func (m Model) IsConnected() bool {
	return m.connected
}

// SubscriptionName returns the current subscription name
func (m Model) SubscriptionName() string {
	return m.subscriptionName
}

// TopicName returns the current topic name
func (m Model) TopicName() string {
	return m.topicName
}

// IsFiltering returns whether filter mode is active
func (m Model) IsFiltering() bool {
	return m.filtering
}

// MessageCount returns the total message count
func (m Model) MessageCount() int {
	return len(m.messages)
}

// DisplayedCount returns the count of displayed messages
func (m Model) DisplayedCount() int {
	return len(m.messageList.Items())
}

// applyFilter filters messages based on current filter text
func (m *Model) applyFilter() {
	var items []list.Item

	for _, msg := range m.messages {
		if m.filterText == "" {
			items = append(items, MessageItem{message: msg})
			continue
		}

		// Search in ID and data
		searchText := msg.ID + string(msg.Data)
		result := utils.MatchesFilter(searchText, m.filterText)
		if result.Error != nil {
			m.filterError = result.Error
			items = append(items, MessageItem{message: msg})
		} else if result.Matches {
			m.filterError = nil
			items = append(items, MessageItem{message: msg})
		}
	}

	m.messageList.SetItems(items)
}

// updateDetailView updates the detail view content
func (m *Model) updateDetailView() {
	msg := m.SelectedMessage()
	if msg == nil {
		m.detailView.SetContent(common.MutedText.Render("No message selected"))
		return
	}

	var content string

	// Message ID
	content += common.FilterPromptStyle.Render("ID: ") + msg.ID + "\n"
	content += common.FilterPromptStyle.Render("Time: ") + msg.PublishTime.Format(time.RFC3339) + "\n"

	// Ack status
	status := "Pending"
	statusStyle := common.LogWarningStyle
	if msg.IsAcked() {
		status = "Acknowledged"
		statusStyle = common.LogSuccessStyle
	}
	content += common.FilterPromptStyle.Render("Status: ") + statusStyle.Render(status) + "\n"

	// Attributes
	if len(msg.Attributes) > 0 {
		content += "\n" + common.FilterPromptStyle.Render("Attributes:") + "\n"
		for k, v := range msg.Attributes {
			content += fmt.Sprintf("  %s: %s\n", k, v)
		}
	}

	// Data
	content += "\n" + common.FilterPromptStyle.Render("Data:") + "\n"
	formatted, _ := utils.FormatJSON(msg.Data)
	content += formatted

	m.detailView.SetContent(content)
	m.detailView.GotoTop()
}

// AckSelected acknowledges the selected message
func (m *Model) AckSelected() bool {
	msg := m.SelectedMessage()
	if msg != nil && !msg.IsAcked() {
		msg.Ack()
		m.applyFilter() // Refresh display
		m.updateDetailView()
		return true
	}
	return false
}

// UpdateSelection updates the detail view when selection changes
func (m *Model) UpdateSelection() {
	m.selectedMessage = m.SelectedMessage()
	m.updateDetailView()
}

// IsInputActive returns whether an input field is active
func (m Model) IsInputActive() bool {
	return m.filtering
}
