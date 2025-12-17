package subscriptions

import (
	"github.com/anmaso/pubsub-tui/internal/components/common"
	"github.com/anmaso/pubsub-tui/internal/utils"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Mode represents the current mode of the subscriptions panel
type Mode int

const (
	ModeNormal Mode = iota
	ModeFilter
	ModeCreate
	ModeConfirmDelete
)

// SubscriptionItem implements list.Item for displaying subscriptions
type SubscriptionItem struct {
	name      string
	fullName  string
	topicName string
	topicFull string
	width     int  // For column formatting
	active    bool // Whether this is the active subscription
}

func (s SubscriptionItem) Title() string {
	// Format as columns: subscription name | topic name
	// Pad subscription name to create alignment
	nameWidth := 20
	if s.width > 0 {
		nameWidth = s.width * 45 / 100 // 45% for name
	}
	if nameWidth < 10 {
		nameWidth = 10
	}

	// Add active indicator
	prefix := "  "
	if s.active {
		prefix = "● "
	}

	name := s.name
	maxNameLen := nameWidth - len(prefix) - 2
	if len(name) > maxNameLen {
		name = name[:maxNameLen-3] + "..."
	}

	// Pad name to fixed width
	fullName := prefix + name
	for len(fullName) < nameWidth {
		fullName += " "
	}

	return fullName + "→ " + s.topicName
}
func (s SubscriptionItem) Description() string { return "" }
func (s SubscriptionItem) FilterValue() string { return s.name }

// Model represents the state of the subscriptions panel
type Model struct {
	list               list.Model
	filterInput        textinput.Model
	createInput        textinput.Model
	spinner            spinner.Model
	allSubscriptions   []common.SubscriptionData // All subscriptions from GCP
	width              int
	height             int
	focused            bool
	mode               Mode
	filterText         string // Current regex filter
	filterError        error
	selectedTopic      string // Topic filter (from topic selection)
	loading            bool
	loadError          error
	statusMsg          string
	statusError        bool
	activeSubscription string // Currently connected subscription
}

// New creates a new subscriptions panel model
func New() Model {
	// Create list with custom delegate - compact style with no description
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false // Topic shown inline in title
	delegate.SetSpacing(0)           // No spacing between items
	delegate.Styles.SelectedTitle = common.SelectedItem
	delegate.Styles.NormalTitle = common.NormalText

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Subscriptions"
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()

	// Create filter input
	fi := textinput.New()
	fi.Placeholder = "regex filter..."
	fi.Prompt = "/ "
	fi.PromptStyle = common.FilterPromptStyle
	fi.TextStyle = common.FilterInputStyle

	// Create subscription input
	ci := textinput.New()
	ci.Placeholder = "new-subscription-name"
	ci.Prompt = "New sub: "
	ci.PromptStyle = common.FilterPromptStyle
	ci.TextStyle = common.FilterInputStyle
	ci.CharLimit = 255

	// Create spinner
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = common.LogNetworkStyle // Blue color for network activity

	return Model{
		list:        l,
		filterInput: fi,
		createInput: ci,
		spinner:     sp,
		loading:     true,
		mode:        ModeNormal,
	}
}

// SetFocused sets whether the panel is focused
func (m *Model) SetFocused(focused bool) {
	m.focused = focused
	if !focused {
		// Reset to normal mode when losing focus
		m.mode = ModeNormal
		m.filterInput.Blur()
		m.createInput.Blur()
	}
}

// IsFocused returns whether the panel is focused
func (m Model) IsFocused() bool {
	return m.focused
}

// SetSize sets the panel dimensions
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Reserve space for title (1 line), topic filter (1 line), filter (1 line), and borders
	listHeight := height - 5
	if listHeight < 1 {
		listHeight = 1
	}

	m.list.SetSize(width-4, listHeight)

	// Refresh items to update column widths
	if len(m.allSubscriptions) > 0 {
		m.applyFilter()
	}
}

// SetSubscriptions updates the list with new subscriptions
func (m *Model) SetSubscriptions(subs []common.SubscriptionData) {
	m.allSubscriptions = subs
	m.loading = false
	m.loadError = nil
	m.applyFilter()
}

// SetError sets a loading error
func (m *Model) SetError(err error) {
	m.loading = false
	m.loadError = err
}

// SetTopicFilter sets the topic filter (from topic selection)
func (m *Model) SetTopicFilter(topicName string) {
	m.selectedTopic = topicName
	m.applyFilter()
}

// ClearTopicFilter clears the topic filter
func (m *Model) ClearTopicFilter() {
	m.selectedTopic = ""
	m.applyFilter()
}

// SelectedTopicFilter returns the current topic filter
func (m Model) SelectedTopicFilter() string {
	return m.selectedTopic
}

// IsLoading returns whether subscriptions are being loaded
func (m Model) IsLoading() bool {
	return m.loading
}

// SelectedSubscription returns the currently selected subscription, if any
func (m Model) SelectedSubscription() *common.SubscriptionData {
	if m.list.SelectedItem() == nil {
		return nil
	}

	item, ok := m.list.SelectedItem().(SubscriptionItem)
	if !ok {
		return nil
	}

	return &common.SubscriptionData{
		Name:      item.name,
		FullName:  item.fullName,
		TopicName: item.topicName,
		TopicFull: item.topicFull,
	}
}

// IsFiltering returns whether filter mode is active
func (m Model) IsFiltering() bool {
	return m.mode == ModeFilter
}

// GetMode returns the current mode
func (m Model) GetMode() Mode {
	return m.mode
}

// SetStatus sets a status message
func (m *Model) SetStatus(msg string, isError bool) {
	m.statusMsg = msg
	m.statusError = isError
}

// ClearStatus clears the status message
func (m *Model) ClearStatus() {
	m.statusMsg = ""
	m.statusError = false
}

// SetActiveSubscription sets the currently active subscription
func (m *Model) SetActiveSubscription(name string) {
	m.activeSubscription = name
	m.applyFilter() // Refresh to update indicators
}

// GetActiveSubscription returns the currently active subscription
func (m Model) GetActiveSubscription() string {
	return m.activeSubscription
}

// IsActiveSubscription checks if the given subscription is the active one
func (m Model) IsActiveSubscription(name string) bool {
	return m.activeSubscription != "" && m.activeSubscription == name
}

// IsInputActive returns whether an input field is active
func (m Model) IsInputActive() bool {
	return m.mode == ModeFilter || m.mode == ModeCreate
}

// SpinnerTickCmd returns the spinner tick command
func (m Model) SpinnerTickCmd() tea.Cmd {
	return m.spinner.Tick
}

// applyFilter filters the subscriptions based on current filters
func (m *Model) applyFilter() {
	var items []list.Item

	for _, sub := range m.allSubscriptions {
		// Apply topic filter first
		if m.selectedTopic != "" && sub.TopicName != m.selectedTopic {
			continue
		}

		// Apply regex filter
		if m.filterText == "" {
			items = append(items, SubscriptionItem{
				name:      sub.Name,
				fullName:  sub.FullName,
				topicName: sub.TopicName,
				topicFull: sub.TopicFull,
				width:     m.width,
				active:    m.activeSubscription == sub.Name,
			})
			continue
		}

		result := utils.MatchesFilter(sub.Name, m.filterText)
		if result.Error != nil {
			m.filterError = result.Error
			// On error, include item
			items = append(items, SubscriptionItem{
				name:      sub.Name,
				fullName:  sub.FullName,
				topicName: sub.TopicName,
				topicFull: sub.TopicFull,
				width:     m.width,
				active:    m.activeSubscription == sub.Name,
			})
		} else if result.Matches {
			m.filterError = nil
			items = append(items, SubscriptionItem{
				name:      sub.Name,
				fullName:  sub.FullName,
				topicName: sub.TopicName,
				topicFull: sub.TopicFull,
				width:     m.width,
				active:    m.activeSubscription == sub.Name,
			})
		}
	}

	m.list.SetItems(items)
}

// TotalCount returns total subscription count
func (m Model) TotalCount() int {
	return len(m.allSubscriptions)
}

// FilteredCount returns count after topic filter (before regex)
func (m Model) FilteredCount() int {
	if m.selectedTopic == "" {
		return len(m.allSubscriptions)
	}

	count := 0
	for _, sub := range m.allSubscriptions {
		if sub.TopicName == m.selectedTopic {
			count++
		}
	}
	return count
}

// DisplayCount returns count of currently displayed items
func (m Model) DisplayCount() int {
	return len(m.list.Items())
}
