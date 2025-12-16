package common

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Color palette - a cohesive dark theme with teal accents
var (
	// Primary colors
	ColorPrimary    = lipgloss.Color("#00D4AA") // Teal accent
	ColorSecondary  = lipgloss.Color("#7C3AED") // Purple accent
	ColorBackground = lipgloss.Color("#0F172A") // Dark slate
	ColorSurface    = lipgloss.Color("#1E293B") // Lighter slate

	// Text colors
	ColorText       = lipgloss.Color("#E2E8F0") // Light gray text
	ColorTextMuted  = lipgloss.Color("#64748B") // Muted text
	ColorTextBright = lipgloss.Color("#F8FAFC") // Bright white

	// Status colors
	ColorSuccess = lipgloss.Color("#22C55E") // Green
	ColorWarning = lipgloss.Color("#EAB308") // Yellow
	ColorError   = lipgloss.Color("#EF4444") // Red
	ColorInfo    = lipgloss.Color("#3B82F6") // Blue
	ColorNetwork = lipgloss.Color("#06B6D4") // Cyan
)

// Border styles
var (
	// FocusedBorder is used for the currently focused panel
	FocusedBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary)

	// UnfocusedBorder is used for panels that are not focused
	UnfocusedBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorTextMuted)
)

// BorderedPanel creates a panel with title embedded in the top border
func BorderedPanel(title string, content string, focused bool, width, height int) string {
	borderColor := ColorTextMuted
	titleColor := ColorTextMuted

	if focused {
		borderColor = ColorPrimary
		titleColor = ColorPrimary
	}

	borderStyle := lipgloss.NewStyle().Foreground(borderColor)
	titleStyle := lipgloss.NewStyle().Foreground(titleColor).Bold(true)

	// Border characters (rounded)
	topLeft := borderStyle.Render("╭")
	topRight := borderStyle.Render("╮")
	bottomLeft := borderStyle.Render("╰")
	bottomRight := borderStyle.Render("╯")
	horizontal := borderStyle.Render("─")
	vertical := borderStyle.Render("│")

	// Calculate inner width (subtract 2 for left/right borders)
	innerWidth := width - 2
	if innerWidth < 1 {
		innerWidth = 1
	}

	// Build top border with title
	styledTitle := titleStyle.Render(" " + title + " ")
	titleLen := lipgloss.Width(styledTitle)

	remainingWidth := innerWidth - titleLen
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	leftPad := 1
	rightPad := remainingWidth - leftPad
	if rightPad < 0 {
		rightPad = 0
	}

	topBorder := topLeft +
		repeatString(horizontal, leftPad) +
		styledTitle +
		repeatString(horizontal, rightPad) +
		topRight

	// Build bottom border
	bottomBorder := bottomLeft + repeatString(horizontal, innerWidth) + bottomRight

	// Process content lines
	contentLines := strings.Split(content, "\n")

	// Calculate inner height (subtract 2 for top/bottom borders)
	innerHeight := height - 2
	if innerHeight < 1 {
		innerHeight = 1
	}

	// Pad or truncate content to fit
	var lines []string
	for i := 0; i < innerHeight; i++ {
		var line string
		if i < len(contentLines) {
			line = contentLines[i]
		}
		// Pad line to inner width
		lineWidth := lipgloss.Width(line)
		if lineWidth < innerWidth {
			line = line + strings.Repeat(" ", innerWidth-lineWidth)
		} else if lineWidth > innerWidth {
			// Truncate - this is simplified, proper truncation would need rune handling
			line = truncateString(line, innerWidth)
		}
		lines = append(lines, vertical+line+vertical)
	}

	// Combine all parts
	result := topBorder + "\n" + strings.Join(lines, "\n") + "\n" + bottomBorder
	return result
}

// repeatString repeats a styled string n times
func repeatString(s string, n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat(s, n)
}

// truncateString truncates a string to fit within maxWidth
func truncateString(s string, maxWidth int) string {
	if lipgloss.Width(s) <= maxWidth {
		return s
	}
	// Simple truncation - could be improved with rune handling
	runes := []rune(s)
	for len(runes) > 0 && lipgloss.Width(string(runes)) > maxWidth {
		runes = runes[:len(runes)-1]
	}
	return string(runes)
}

// Panel title styles
var (
	// TitleStyle for panel headers
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Padding(0, 1)

	// TitleStyleMuted for unfocused panel headers
	TitleStyleMuted = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorTextMuted).
			Padding(0, 1)
)

// Text styles
var (
	// NormalText for regular content
	NormalText = lipgloss.NewStyle().
			Foreground(ColorText)

	// MutedText for secondary content
	MutedText = lipgloss.NewStyle().
			Foreground(ColorTextMuted)

	// BrightText for highlighted content
	BrightText = lipgloss.NewStyle().
			Foreground(ColorTextBright)

	// SelectedItem for currently selected list items
	SelectedItem = lipgloss.NewStyle().
			Foreground(ColorTextBright).
			Background(ColorPrimary).
			Bold(true)
)

// Status styles for log messages
var (
	LogInfoStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	LogSuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	LogWarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning)

	LogErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError)

	LogNetworkStyle = lipgloss.NewStyle().
			Foreground(ColorNetwork)

	LogTimestampStyle = lipgloss.NewStyle().
				Foreground(ColorTextMuted)
)

// Footer styles
var (
	FooterStyle = lipgloss.NewStyle().
			Foreground(ColorTextMuted).
			Padding(0, 1)

	FooterKeyStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	FooterDescStyle = lipgloss.NewStyle().
			Foreground(ColorTextMuted)

	FooterProjectStyle = lipgloss.NewStyle().
				Foreground(ColorSecondary)
)

// Filter styles
var (
	FilterPromptStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary)

	FilterInputStyle = lipgloss.NewStyle().
				Foreground(ColorText)

	FilterErrorStyle = lipgloss.NewStyle().
				Foreground(ColorError)
)

// Helper functions for creating panel styles with dimensions
func PanelStyle(focused bool, width, height int) lipgloss.Style {
	var style lipgloss.Style
	if focused {
		style = FocusedBorder
	} else {
		style = UnfocusedBorder
	}

	return style.
		Width(width - 2).  // Account for border
		Height(height - 2) // Account for border
}

// GetLogStyle returns the appropriate style for a log level
func GetLogStyle(level LogLevel) lipgloss.Style {
	switch level {
	case LogSuccess:
		return LogSuccessStyle
	case LogWarning:
		return LogWarningStyle
	case LogError:
		return LogErrorStyle
	case LogNetwork:
		return LogNetworkStyle
	default:
		return LogInfoStyle
	}
}
