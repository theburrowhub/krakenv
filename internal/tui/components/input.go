package components

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// InputModel wraps bubbles textinput with validation support.
type InputModel struct {
	textInput textinput.Model
	err       error
	validator func(string) error
}

// NewInputModel creates a new text input model.
func NewInputModel(placeholder string) InputModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	return InputModel{
		textInput: ti,
	}
}

// WithValidator sets a validation function.
func (m InputModel) WithValidator(v func(string) error) InputModel {
	m.validator = v
	return m
}

// WithDefault sets the initial value.
func (m InputModel) WithDefault(value string) InputModel {
	m.textInput.SetValue(value)
	return m
}

// WithWidth sets the input width.
func (m InputModel) WithWidth(w int) InputModel {
	m.textInput.Width = w
	return m
}

// WithCharLimit sets the character limit.
func (m InputModel) WithCharLimit(limit int) InputModel {
	m.textInput.CharLimit = limit
	return m
}

// Init implements tea.Model.
func (m InputModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model.
func (m InputModel) Update(msg tea.Msg) (InputModel, tea.Cmd) {
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)

	// Validate on change
	if m.validator != nil {
		m.err = m.validator(m.textInput.Value())
	}

	return m, cmd
}

// View implements tea.Model.
func (m InputModel) View() string {
	return m.textInput.View()
}

// Value returns the current input value.
func (m InputModel) Value() string {
	return m.textInput.Value()
}

// SetValue sets the input value.
func (m *InputModel) SetValue(s string) {
	m.textInput.SetValue(s)
}

// Error returns the current validation error.
func (m InputModel) Error() error {
	return m.err
}

// IsValid returns true if there's no validation error.
func (m InputModel) IsValid() bool {
	return m.err == nil
}

// Focus focuses the input.
func (m *InputModel) Focus() tea.Cmd {
	return m.textInput.Focus()
}

// Blur removes focus from the input.
func (m *InputModel) Blur() {
	m.textInput.Blur()
}

// Focused returns whether the input is focused.
func (m InputModel) Focused() bool {
	return m.textInput.Focused()
}

// Reset clears the input.
func (m *InputModel) Reset() {
	m.textInput.Reset()
	m.err = nil
}

// Validate runs the validator and returns the error.
func (m *InputModel) Validate() error {
	if m.validator != nil {
		m.err = m.validator(m.textInput.Value())
	}
	return m.err
}
