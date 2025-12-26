package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SelectModel provides a selectable list of options.
type SelectModel struct {
	Options  []string
	Cursor   int
	Selected int
	focused  bool
}

// NewSelectModel creates a new select model with options.
func NewSelectModel(options []string) SelectModel {
	return SelectModel{
		Options:  options,
		Cursor:   0,
		Selected: -1,
		focused:  true,
	}
}

// Init implements tea.Model.
func (m SelectModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m SelectModel) Update(msg tea.Msg) (SelectModel, tea.Cmd) {
	if !m.focused {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Options)-1 {
				m.Cursor++
			}
		case "enter", " ":
			m.Selected = m.Cursor
		case "home":
			m.Cursor = 0
		case "end":
			m.Cursor = len(m.Options) - 1
		}
	}

	return m, nil
}

// View implements tea.Model.
func (m SelectModel) View() string {
	if len(m.Options) == 0 {
		return MutedStyle.Render("No options available")
	}

	var b strings.Builder

	cursorStyle := lipgloss.NewStyle().Foreground(ColorAccent)
	selectedStyle := lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(ColorText)

	for i, opt := range m.Options {
		cursor := "  "
		style := normalStyle

		if i == m.Cursor && m.focused {
			cursor = cursorStyle.Render("> ")
			style = cursorStyle
		}

		if i == m.Selected {
			style = selectedStyle
		}

		b.WriteString(cursor)
		b.WriteString(style.Render(opt))
		if i < len(m.Options)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// Value returns the selected option value.
func (m SelectModel) Value() string {
	if m.Selected >= 0 && m.Selected < len(m.Options) {
		return m.Options[m.Selected]
	}
	return ""
}

// HasSelection returns true if an option is selected.
func (m SelectModel) HasSelection() bool {
	return m.Selected >= 0
}

// Focus focuses the select.
func (m *SelectModel) Focus() {
	m.focused = true
}

// Blur removes focus.
func (m *SelectModel) Blur() {
	m.focused = false
}

// Focused returns whether the select is focused.
func (m SelectModel) Focused() bool {
	return m.focused
}

// SetOptions updates the available options.
func (m *SelectModel) SetOptions(options []string) {
	m.Options = options
	m.Cursor = 0
	m.Selected = -1
}

// Reset clears the selection.
func (m *SelectModel) Reset() {
	m.Cursor = 0
	m.Selected = -1
}

// Select programmatically selects an option by value.
func (m *SelectModel) Select(value string) bool {
	for i, opt := range m.Options {
		if opt == value {
			m.Selected = i
			m.Cursor = i
			return true
		}
	}
	return false
}
