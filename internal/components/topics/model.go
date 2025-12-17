package topics

import (
	"github.com/anmaso/pubsub-tui/internal/components/common"
	"github.com/anmaso/pubsub-tui/internal/utils"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Mode represents the current mode of the topics panel
type Mode int

const (
	ModeNormal Mode = iota
	ModeFilter
	ModeCreate
	ModeConfirmDelete
)

// TopicItem implements list.Item for displaying topics
type TopicItem struct {
	name     string
	fullName string
	selected bool // Whether this topic is currently selected
}

func (t TopicItem) Title() string {
	prefix := "  "
	if t.selected {
		prefix = "● "
	}
	return prefix + t.name
}
func (t TopicItem) Description() string { return "" }
func (t TopicItem) FilterValue() string { return t.name }

// Model represents the state of the topics panel
type Model struct {
	list          list.Model
	filterInput   textinput.Model
	createInput   textinput.Model
	spinner       spinner.Model
	allTopics     []common.TopicData // All topics from GCP
	width         int
	height        int
	focused       bool
	mode          Mode
	filterText    string
	filterError   error
	loading       bool
	loadError     error
	statusMsg     string
	statusError   bool
	selectedTopic string // Currently selected topic
}

// New creates a new topics panel model
func New() Model {
	// Create list with custom delegate - compact style
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.SetSpacing(0) // No spacing between items
	delegate.Styles.SelectedTitle = common.SelectedItem
	delegate.Styles.NormalTitle = common.NormalText

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Topics"
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

	// Create topic input
	ci := textinput.New()
	ci.Placeholder = "new-topic-name"
	ci.Prompt = "New topic: "
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

	// Reserve space for title (1 line) and filter (1 line) and borders
	listHeight := height - 4
	if listHeight < 1 {
		listHeight = 1
	}

	m.list.SetSize(width-4, listHeight)
}

// SetTopics updates the list with new topics
func (m *Model) SetTopics(topics []common.TopicData) {
	m.allTopics = topics
	m.loading = false
	m.loadError = nil
	m.applyFilter()
}

// SetError sets a loading error
func (m *Model) SetError(err error) {
	m.loading = false
	m.loadError = err
}

// IsLoading returns whether topics are being loaded
func (m Model) IsLoading() bool {
	return m.loading
}

// SelectedTopic returns the currently selected topic, if any
func (m Model) SelectedTopic() *common.TopicData {
	if m.list.SelectedItem() == nil {
		return nil
	}

	item, ok := m.list.SelectedItem().(TopicItem)
	if !ok {
		return nil
	}

	return &common.TopicData{
		Name:     item.name,
		FullName: item.fullName,
	}
}

// IsFiltering returns whether filter mode is active
func (m Model) IsFiltering() bool {
	return m.mode == ModeFilter
}

// Mode returns the current mode
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

// SetSelectedTopic sets the currently selected topic
func (m *Model) SetSelectedTopic(name string) {
	m.selectedTopic = name
	m.applyFilter() // Refresh to update indicators
}

// GetSelectedTopic returns the currently selected topic
func (m Model) GetSelectedTopic() string {
	return m.selectedTopic
}

// IsInputActive returns whether an input field is active
func (m Model) IsInputActive() bool {
	return m.mode == ModeFilter || m.mode == ModeCreate
}

// SpinnerTickCmd returns the spinner tick command
func (m Model) SpinnerTickCmd() tea.Cmd {
	return m.spinner.Tick
}

// applyFilter filters the topics based on current filter text
func (m *Model) applyFilter() {
	var items []list.Item

	for _, topic := range m.allTopics {
		// If no filter, include all
		if m.filterText == "" {
			items = append(items, TopicItem{
				name:     topic.Name,
				fullName: topic.FullName,
				selected: m.selectedTopic == topic.Name,
			})
			continue
		}

		// Apply regex filter
		result := matchFilter(topic.Name, m.filterText)
		if result.err != nil {
			m.filterError = result.err
			// On error, show all topics
			items = append(items, TopicItem{
				name:     topic.Name,
				fullName: topic.FullName,
				selected: m.selectedTopic == topic.Name,
			})
		} else if result.matches {
			m.filterError = nil
			items = append(items, TopicItem{
				name:     topic.Name,
				fullName: topic.FullName,
				selected: m.selectedTopic == topic.Name,
			})
		}
	}

	m.list.SetItems(items)
}

// filterResult holds the result of a filter operation
type filterResult struct {
	matches bool
	err     error
}

// matchFilter checks if text matches the regex pattern
func matchFilter(text, pattern string) filterResult {
	if pattern == "" {
		return filterResult{matches: true}
	}

	// Use our utils package
	result := utils.MatchesFilter(text, pattern)
	return filterResult{matches: result.Matches, err: result.Error}
}
