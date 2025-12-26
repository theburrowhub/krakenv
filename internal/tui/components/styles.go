// Package components provides reusable TUI components for krakenv.
package components

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette - Kraken theme (deep sea colors).
var (
	// Primary colors
	ColorPrimary   = lipgloss.Color("#6C5CE7") // Deep purple
	ColorSecondary = lipgloss.Color("#00B894") // Sea green
	ColorAccent    = lipgloss.Color("#FDCB6E") // Golden

	// Semantic colors
	ColorSuccess = lipgloss.Color("#00B894") // Green
	ColorError   = lipgloss.Color("#D63031") // Red
	ColorWarning = lipgloss.Color("#FDCB6E") // Yellow
	ColorInfo    = lipgloss.Color("#74B9FF") // Light blue

	// Neutral colors
	ColorText       = lipgloss.Color("#DFE6E9") // Light gray
	ColorMuted      = lipgloss.Color("#636E72") // Dark gray
	ColorBackground = lipgloss.Color("#2D3436") // Dark background
	ColorBorder     = lipgloss.Color("#636E72") // Border gray
)

// Text styles.
var (
	// TitleStyle for main headers.
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginBottom(1)

	// SubtitleStyle for secondary headers.
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			MarginBottom(1)

	// TextStyle for normal text.
	TextStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	// MutedStyle for less important text.
	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	// BoldStyle for emphasized text.
	BoldStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorText)
)

// Status styles.
var (
	// SuccessStyle for success messages.
	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	// ErrorStyle for error messages.
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError)

	// WarningStyle for warning messages.
	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning)

	// InfoStyle for info messages.
	InfoStyle = lipgloss.NewStyle().
			Foreground(ColorInfo)
)

// Component styles.
var (
	// PromptStyle for input prompts.
	PromptStyle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)

	// ConstraintStyle for showing type constraints.
	ConstraintStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Italic(true)

	// ValueStyle for displaying values.
	ValueStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary)

	// DefaultValueStyle for showing default values.
	DefaultValueStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Italic(true)
)

// Box styles.
var (
	// BoxStyle for bordered boxes.
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(1, 2)

	// ActiveBoxStyle for focused/active boxes.
	ActiveBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2)

	// ErrorBoxStyle for error boxes.
	ErrorBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorError).
			Padding(1, 2)
)

// Layout styles.
var (
	// HeaderStyle for the app header.
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Background(ColorBackground).
			Padding(0, 2).
			MarginBottom(1)

	// FooterStyle for the app footer.
	FooterStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			MarginTop(1)

	// HelpStyle for help text.
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Italic(true)
)

// Icons and symbols.
const (
	IconSuccess  = "✓"
	IconError    = "✗"
	IconWarning  = "⚠"
	IconInfo     = "ℹ"
	IconArrow    = "→"
	IconBullet   = "•"
	IconSecret   = "🔒"
	IconOptional = "○"
	IconRequired = "●"
	IconKraken   = "🐙"
)

// RenderSuccess renders a success message.
func RenderSuccess(msg string) string {
	return SuccessStyle.Render(IconSuccess + " " + msg)
}

// RenderError renders an error message.
func RenderError(msg string) string {
	return ErrorStyle.Render(IconError + " " + msg)
}

// RenderWarning renders a warning message.
func RenderWarning(msg string) string {
	return WarningStyle.Render(IconWarning + " " + msg)
}

// RenderInfo renders an info message.
func RenderInfo(msg string) string {
	return InfoStyle.Render(IconInfo + " " + msg)
}

// RenderPrompt renders a prompt with optional indicator.
func RenderPrompt(prompt string, isOptional, isSecret bool) string {
	var prefix string
	if isSecret {
		prefix = IconSecret + " "
	} else if isOptional {
		prefix = IconOptional + " "
	} else {
		prefix = IconRequired + " "
	}
	return PromptStyle.Render(prefix + prompt)
}

// RenderConstraint renders a type constraint hint.
func RenderConstraint(constraint string) string {
	return ConstraintStyle.Render("[" + constraint + "]")
}

// RenderDefault renders a default value hint.
func RenderDefault(value string) string {
	if value == "" {
		return ""
	}
	return DefaultValueStyle.Render("(default: " + value + ")")
}

// RenderHeader renders the application header.
func RenderHeader(title string) string {
	return HeaderStyle.Render(IconKraken + " " + title)
}

// RenderFooter renders help text in the footer.
func RenderFooter(help string) string {
	return FooterStyle.Render(help)
}
