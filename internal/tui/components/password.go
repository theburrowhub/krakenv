package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// PasswordModel wraps textinput for password entry with hidden display.
type PasswordModel struct {
	textInput textinput.Model
	err       error
	validator func(string) error
}

// NewPasswordModel creates a new password input model.
func NewPasswordModel(placeholder string) PasswordModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'

	return PasswordModel{
		textInput: ti,
	}
}

// WithValidator sets a validation function.
func (m PasswordModel) WithValidator(v func(string) error) PasswordModel {
	m.validator = v
	return m
}

// WithWidth sets the input width.
func (m PasswordModel) WithWidth(w int) PasswordModel {
	m.textInput.Width = w
	return m
}

// WithCharLimit sets the character limit.
func (m PasswordModel) WithCharLimit(limit int) PasswordModel {
	m.textInput.CharLimit = limit
	return m
}

// Init implements tea.Model.
func (m PasswordModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model.
func (m PasswordModel) Update(msg tea.Msg) (PasswordModel, tea.Cmd) {
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)

	// Validate on change
	if m.validator != nil {
		m.err = m.validator(m.textInput.Value())
	}

	return m, cmd
}

// View implements tea.Model.
func (m PasswordModel) View() string {
	return m.textInput.View()
}

// Value returns the current input value.
func (m PasswordModel) Value() string {
	return m.textInput.Value()
}

// SetValue sets the input value.
func (m *PasswordModel) SetValue(s string) {
	m.textInput.SetValue(s)
}

// Error returns the current validation error.
func (m PasswordModel) Error() error {
	return m.err
}

// IsValid returns true if there's no validation error.
func (m PasswordModel) IsValid() bool {
	return m.err == nil
}

// Focus focuses the input.
func (m *PasswordModel) Focus() tea.Cmd {
	return m.textInput.Focus()
}

// Blur removes focus from the input.
func (m *PasswordModel) Blur() {
	m.textInput.Blur()
}

// Focused returns whether the input is focused.
func (m PasswordModel) Focused() bool {
	return m.textInput.Focused()
}

// Reset clears the input.
func (m *PasswordModel) Reset() {
	m.textInput.Reset()
	m.err = nil
}

// Validate runs the validator and returns the error.
func (m *PasswordModel) Validate() error {
	if m.validator != nil {
		m.err = m.validator(m.textInput.Value())
	}
	return m.err
}

// MaskedValue returns a masked version of the value for display.
func (m PasswordModel) MaskedValue() string {
	if m.textInput.Value() == "" {
		return ""
	}
	return strings.Repeat("•", len(m.textInput.Value()))
}
