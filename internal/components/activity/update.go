package activity

import (
	"github.com/anmaso/pubsub-tui/internal/components/common"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles messages for the activity log panel
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case common.LogMsg:
		m.AddLog(msg)
		return m, nil
	}

	// Pass other messages to viewport
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)

	return m, cmd
}
