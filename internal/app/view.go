package app

import (
	"strings"

	"github.com/anmaso/pubsub-tui/internal/components/common"

	"github.com/charmbracelet/lipgloss"
)

// View renders the application
func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// Build left panel (Topics, Subscriptions, Activity stacked vertically)
	leftPanel := lipgloss.JoinVertical(
		lipgloss.Left,
		m.topics.View(),
		m.subscriptions.View(),
		m.activity.View(),
	)

	// Build right panel (Publisher, Subscriber stacked vertically)
	rightPanel := lipgloss.JoinVertical(
		lipgloss.Left,
		m.publisher.View(),
		m.subscriber.View(),
	)

	// Combine panels horizontally
	mainContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPanel,
		rightPanel,
	)

	// Build footer
	footer := m.renderFooter()

	// Combine main content and footer
	baseView := lipgloss.JoinVertical(
		lipgloss.Left,
		mainContent,
		footer,
	)

	// Show help popup as overlay if active
	if m.showHelp {
		return m.renderHelpOverlay(baseView)
	}

	return baseView
}

// renderFooter renders the application footer with dynamic shortcuts based on focused panel
func (m Model) renderFooter() string {
	var parts []string

	// Global shortcuts (always shown)
	parts = append(parts, common.FooterKeyStyle.Render("1-4")+common.FooterDescStyle.Render(":panel"))
	parts = append(parts, common.FooterKeyStyle.Render("Tab")+common.FooterDescStyle.Render(":cycle"))
	parts = append(parts, common.FooterKeyStyle.Render("?")+common.FooterDescStyle.Render(":help"))
	parts = append(parts, common.FooterKeyStyle.Render("q")+common.FooterDescStyle.Render(":quit"))

	// Panel-specific shortcuts
	panelShortcuts := m.getPanelShortcuts()
	if len(panelShortcuts) > 0 {
		parts = append(parts, common.FooterDescStyle.Render("│"))
		parts = append(parts, panelShortcuts...)
	}

	// Subscription status
	var statusInfo string
	if m.selectedSubscription != "" {
		statusInfo = common.LogNetworkStyle.Render("● ") +
			common.FooterDescStyle.Render(m.selectedSubscription)
	}

	// Project info
	projectInfo := common.FooterDescStyle.Render("Project: ") +
		common.FooterProjectStyle.Render(m.projectID)

	// Build footer line
	helpText := strings.Join(parts, " ")

	// Calculate spacing
	helpLen := lipgloss.Width(helpText)
	statusLen := lipgloss.Width(statusInfo)
	projectLen := lipgloss.Width(projectInfo)

	totalRight := statusLen + projectLen
	if statusLen > 0 {
		totalRight += 3 // separator
	}

	spacing := m.width - helpLen - totalRight - 4
	if spacing < 1 {
		spacing = 1
	}

	var footer string
	if statusInfo != "" {
		footer = helpText + strings.Repeat(" ", spacing) + statusInfo + " │ " + projectInfo
	} else {
		footer = helpText + strings.Repeat(" ", spacing+3) + projectInfo
	}

	return common.FooterStyle.Render(footer)
}

// getPanelShortcuts returns shortcuts specific to the currently focused panel
func (m Model) getPanelShortcuts() []string {
	var shortcuts []string

	switch m.focus {
	case FocusTopics:
		shortcuts = append(shortcuts,
			common.FooterKeyStyle.Render("↑↓")+common.FooterDescStyle.Render(":nav"),
			common.FooterKeyStyle.Render("Enter")+common.FooterDescStyle.Render(":select"),
			common.FooterKeyStyle.Render("n")+common.FooterDescStyle.Render(":new"),
			common.FooterKeyStyle.Render("d")+common.FooterDescStyle.Render(":delete"),
			common.FooterKeyStyle.Render("/")+common.FooterDescStyle.Render(":filter"),
		)

	case FocusSubscriptions:
		// Show Esc:stop only when there's an active subscription
		if m.subscriptions.GetActiveSubscription() != "" {
			shortcuts = append(shortcuts,
				common.FooterKeyStyle.Render("Esc")+common.FooterDescStyle.Render(":stop"),
			)
		}
		shortcuts = append(shortcuts,
			common.FooterKeyStyle.Render("↑↓")+common.FooterDescStyle.Render(":nav"),
			common.FooterKeyStyle.Render("Enter")+common.FooterDescStyle.Render(":start/stop"),
			common.FooterKeyStyle.Render("n")+common.FooterDescStyle.Render(":new"),
			common.FooterKeyStyle.Render("d")+common.FooterDescStyle.Render(":delete"),
			common.FooterKeyStyle.Render("/")+common.FooterDescStyle.Render(":filter"),
		)

	case FocusPublisher:
		shortcuts = append(shortcuts,
			common.FooterKeyStyle.Render("↑↓")+common.FooterDescStyle.Render(":nav"),
			common.FooterKeyStyle.Render("Enter")+common.FooterDescStyle.Render(":publish"),
			common.FooterKeyStyle.Render("v")+common.FooterDescStyle.Render(":vars"),
		)

	case FocusSubscriber:
		// Show Esc:stop only when connected
		if m.subscriber.IsConnected() {
			shortcuts = append(shortcuts,
				common.FooterKeyStyle.Render("Esc")+common.FooterDescStyle.Render(":stop"),
			)
		}
		shortcuts = append(shortcuts,
			common.FooterKeyStyle.Render("↑↓")+common.FooterDescStyle.Render(":nav"),
			common.FooterKeyStyle.Render("a")+common.FooterDescStyle.Render(":ack"),
			common.FooterKeyStyle.Render("A")+common.FooterDescStyle.Render(":auto-ack"),
			common.FooterKeyStyle.Render("/")+common.FooterDescStyle.Render(":filter"),
			common.FooterKeyStyle.Render("^d/^u")+common.FooterDescStyle.Render(":scroll"),
		)
	}

	return shortcuts
}

// renderHelpOverlay renders the help dialog as an overlay on top of the base view
func (m Model) renderHelpOverlay(baseView string) string {
	// Build help content - each line exactly 66 characters (fits in 70-char box with padding)
	helpLines := []string{
		"",
		"NAVIGATION",
		"",
		"1-4         Jump to panel (Topics/Subscriptions/Publisher/Sub)",
		"Tab         Cycle focus forward",
		"Shift+Tab   Cycle focus backward",
		"q           Quit application",
		"?           Show this help",
		"",
		"TOPICS PANEL (1)",
		"",
		"j/k or ↑↓   Navigate list",
		"Enter       Select topic for publisher",
		"n           Create new topic",
		"d           Delete selected topic",
		"/           Filter topics by regex",
		"",
		"SUBSCRIPTIONS PANEL (2)",
		"",
		"j/k or ↑↓   Navigate list",
		"Enter       Start/stop subscription in subscriber panel",
		"n           Create new subscription",
		"d           Delete selected subscription",
		"/           Filter subscriptions by regex",
		"",
		"PUBLISHER PANEL (3)",
		"",
		"j/k or ↑↓   Navigate message templates",
		"Enter       Publish message to topic",
		"v           Edit variables for substitution",
		"            (use ${varName} in JSON templates)",
		"",
		"SUBSCRIBER PANEL (4)",
		"",
		"j/k or ↑↓   Navigate messages",
		"a           Acknowledge selected message (moves to next)",
		"A           Toggle auto-acknowledge mode",
		"/           Filter messages by regex",
		"Ctrl+d/u    Scroll message detail up/down",
		"",
	}

	// Join content
	contentText := strings.Join(helpLines, "\n")

	// Style for the content
	contentStyle := lipgloss.NewStyle().
		Width(66).
		Foreground(common.ColorPrimary).
		Background(lipgloss.Color("#0a0a0a"))

	// Title
	titleStyle := lipgloss.NewStyle().
		Width(66).
		Align(lipgloss.Center).
		Foreground(common.ColorPrimary).
		Background(lipgloss.Color("#0a0a0a")).
		Bold(true)

	// Footer
	footerStyle := lipgloss.NewStyle().
		Width(66).
		Align(lipgloss.Center).
		Foreground(common.ColorPrimary).
		Background(lipgloss.Color("#0a0a0a"))

	// Build the complete content
	fullContent := titleStyle.Render("PUBSUB-TUI HELP") + "\n" +
		contentStyle.Render(contentText) + "\n" +
		footerStyle.Render("Press any key to close")

	// Apply border around everything
	helpBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(common.ColorPrimary).
		Background(lipgloss.Color("#0a0a0a")).
		Padding(0, 1)

	styledHelpBox := helpBox.Render(fullContent)

	// Use Place to overlay on the base view with dimmed background
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		styledHelpBox,
		lipgloss.WithWhitespaceChars("░"),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("#333333")),
	)
}
