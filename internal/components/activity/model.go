package activity

import (
	"github.com/anmaso/pubsub-tui/internal/components/common"

	"github.com/charmbracelet/bubbles/viewport"
)

// LogEntry represents a single log entry
type LogEntry struct {
	Level   common.LogLevel
	Message string
	Time    string // Formatted time string
}

// Model represents the state of the activity log panel
type Model struct {
	viewport viewport.Model
	entries  []LogEntry
	width    int
	height   int
}

// New creates a new activity log panel model
func New() Model {
	vp := viewport.New(0, 0)
	vp.Style = common.NormalText

	return Model{
		viewport: vp,
		entries:  make([]LogEntry, 0),
	}
}

// SetSize sets the panel dimensions
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Reserve space for title (1 line) and borders
	vpHeight := height - 3
	if vpHeight < 1 {
		vpHeight = 1
	}
	vpWidth := width - 4
	if vpWidth < 1 {
		vpWidth = 1
	}

	m.viewport.Width = vpWidth
	m.viewport.Height = vpHeight
	m.updateContent()
}

// AddLog adds a new log entry
func (m *Model) AddLog(msg common.LogMsg) {
	entry := LogEntry{
		Level:   msg.Level,
		Message: msg.Message,
		Time:    msg.Time.Format("15:04:05"),
	}

	m.entries = append(m.entries, entry)
	m.updateContent()

	// Auto-scroll to bottom
	m.viewport.GotoBottom()
}

// updateContent rebuilds the viewport content from entries
func (m *Model) updateContent() {
	content := renderEntries(m.entries, m.viewport.Width)
	m.viewport.SetContent(content)
}

// EntryCount returns the number of log entries
func (m Model) EntryCount() int {
	return len(m.entries)
}


