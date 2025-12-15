package publisher

import (
	"pubsub-tui/internal/components/common"
	"pubsub-tui/internal/utils"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
)

// FileItem implements list.Item for displaying JSON files
type FileItem struct {
	name string
	path string
	size int64
}

func (f FileItem) Title() string       { return f.name }
func (f FileItem) Description() string { return "" }
func (f FileItem) FilterValue() string { return f.name }

// FocusArea represents which area of the publisher is focused
type FocusArea int

const (
	FocusFileList FocusArea = iota
	FocusVariables
)

// Model represents the state of the publisher panel
type Model struct {
	fileList       list.Model
	variablesInput textinput.Model
	preview        viewport.Model

	allFiles       []utils.JSONFile
	selectedFile   *utils.JSONFile
	fileContent    string // Raw file content
	previewContent string // Content with substitutions applied

	width     int
	height    int
	focused   bool
	focusArea FocusArea

	targetTopic string // Topic to publish to
	status      string // Status message
	statusError bool   // Whether status is an error

	publishing bool // Whether a publish is in progress
}

// New creates a new publisher panel model
func New() Model {
	// Create file list with compact style
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.SetSpacing(0) // No spacing between items
	delegate.Styles.SelectedTitle = common.SelectedItem
	delegate.Styles.NormalTitle = common.NormalText

	fl := list.New([]list.Item{}, delegate, 0, 0)
	fl.Title = "JSON Files"
	fl.SetShowTitle(false)
	fl.SetShowStatusBar(false)
	fl.SetShowHelp(false)
	fl.SetFilteringEnabled(false)
	fl.DisableQuitKeybindings()

	// Create variables input
	vi := textinput.New()
	vi.Placeholder = "key=value key2=value2..."
	vi.Prompt = "Vars: "
	vi.PromptStyle = common.FilterPromptStyle
	vi.TextStyle = common.FilterInputStyle
	vi.CharLimit = 512

	// Create preview viewport
	pv := viewport.New(0, 0)

	return Model{
		fileList:       fl,
		variablesInput: vi,
		preview:        pv,
		focusArea:      FocusFileList,
	}
}

// SetFocused sets whether the panel is focused
func (m *Model) SetFocused(focused bool) {
	m.focused = focused
	if focused && m.focusArea == FocusVariables {
		m.variablesInput.Focus()
	} else {
		m.variablesInput.Blur()
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

	// Split: left side for file list and vars, right side for preview
	// Left: 40%, Right: 60%
	contentHeight := height - 4 // borders and status
	if contentHeight < 4 {
		contentHeight = 4
	}

	leftWidth := (width - 4) * 40 / 100
	if leftWidth < 15 {
		leftWidth = 15
	}
	rightWidth := (width - 4) - leftWidth - 1 // -1 for separator
	if rightWidth < 15 {
		rightWidth = 15
	}

	// File list takes most of left side, vars at bottom
	fileListHeight := contentHeight - 3 // Leave room for vars input and topic
	if fileListHeight < 2 {
		fileListHeight = 2
	}

	m.fileList.SetSize(leftWidth, fileListHeight)
	m.preview.Width = rightWidth
	m.preview.Height = contentHeight
}

// SetTargetTopic sets the topic to publish to
func (m *Model) SetTargetTopic(topic string) {
	m.targetTopic = topic
}

// TargetTopic returns the current target topic
func (m Model) TargetTopic() string {
	return m.targetTopic
}

// SetFiles updates the list of JSON files
func (m *Model) SetFiles(files []utils.JSONFile) {
	m.allFiles = files

	var items []list.Item
	for _, f := range files {
		items = append(items, FileItem{
			name: f.Name,
			path: f.Path,
			size: f.Size,
		})
	}

	m.fileList.SetItems(items)

	// Auto-select first file
	if len(files) > 0 && m.selectedFile == nil {
		m.selectFile(&files[0])
	}
}

// selectFile selects a file and loads its content
func (m *Model) selectFile(file *utils.JSONFile) {
	m.selectedFile = file

	// Load file content
	content, err := utils.ReadFile(file.Path)
	if err != nil {
		m.fileContent = ""
		m.previewContent = "Error loading file: " + err.Error()
		return
	}

	m.fileContent = string(content)
	m.updatePreview()
}

// updatePreview updates the preview with variable substitutions
func (m *Model) updatePreview() {
	if m.fileContent == "" {
		m.previewContent = ""
		m.preview.SetContent("")
		return
	}

	// Parse variables and substitute
	vars := ParseVariables(m.variablesInput.Value())
	substituted := SubstituteVariables(m.fileContent, vars)

	// Try to format as JSON
	formatted, _ := utils.FormatJSON([]byte(substituted))
	m.previewContent = formatted
	m.preview.SetContent(formatted)
}

// SelectedFile returns the currently selected file
func (m Model) SelectedFile() *utils.JSONFile {
	return m.selectedFile
}

// GetMessageContent returns the message content with substitutions applied
func (m Model) GetMessageContent() string {
	if m.fileContent == "" {
		return ""
	}

	vars := ParseVariables(m.variablesInput.Value())
	return SubstituteVariables(m.fileContent, vars)
}

// SetStatus sets the status message
func (m *Model) SetStatus(msg string, isError bool) {
	m.status = msg
	m.statusError = isError
}

// ClearStatus clears the status message
func (m *Model) ClearStatus() {
	m.status = ""
	m.statusError = false
}

// IsPublishing returns whether a publish is in progress
func (m Model) IsPublishing() bool {
	return m.publishing
}

// SetPublishing sets the publishing state
func (m *Model) SetPublishing(publishing bool) {
	m.publishing = publishing
}

// FocusArea returns the current focus area
func (m Model) GetFocusArea() FocusArea {
	return m.focusArea
}

// IsInputActive returns whether an input field is active
func (m Model) IsInputActive() bool {
	return m.focusArea == FocusVariables
}
