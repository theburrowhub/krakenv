// Package wizard provides the TUI wizard for configuring environment variables.
package wizard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/theburrowhub/krakenv/internal/parser"
	"github.com/theburrowhub/krakenv/internal/tui/components"
	"github.com/theburrowhub/krakenv/internal/validator"
)

// State represents the wizard state.
type State int

const (
	// StatePrompting is asking for user input.
	StatePrompting State = iota
	// StateValidating is validating user input.
	StateValidating
	// StateComplete is done with all variables.
	StateComplete
	// StateAborted is user cancelled.
	StateAborted
	// StateError is an error occurred.
	StateError
)

// Model is the wizard model.
type Model struct {
	Variables    []parser.Variable // Variables to prompt for
	CurrentIndex int               // Current variable index
	Values       map[string]string // Collected values
	State        State
	Error        error

	// Input components
	textInput   textinput.Model
	selectModel components.SelectModel
	useSelect   bool // True if current variable is enum

	// Display
	Width  int
	Height int

	// Interruption handling
	hasUnsavedChanges bool
	showExitPrompt    bool
	exitChoice        int // 0 = discard, 1 = save
}

// New creates a new wizard model.
func New(variables []parser.Variable) Model {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	return Model{
		Variables:    variables,
		Values:       make(map[string]string),
		State:        StatePrompting,
		textInput:    ti,
		selectModel:  components.NewSelectModel(nil),
		CurrentIndex: 0,
	}
}

// CurrentVariable returns the current variable being prompted.
func (m Model) CurrentVariable() *parser.Variable {
	if m.CurrentIndex >= len(m.Variables) {
		return nil
	}
	return &m.Variables[m.CurrentIndex]
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	m.setupCurrentInput()
	return textinput.Blink
}

// setupCurrentInput configures the input for the current variable.
func (m *Model) setupCurrentInput() {
	v := m.CurrentVariable()
	if v == nil {
		return
	}

	// Reset input
	m.textInput.Reset()
	m.textInput.Placeholder = ""

	// Check if it's an enum (use select) or other type (use text input)
	if v.Annotation != nil && v.Annotation.Type == parser.TypeEnum {
		options := strings.Split(v.Annotation.GetConstraint("options"), ",")
		for i := range options {
			options[i] = strings.TrimSpace(options[i])
		}
		m.selectModel.SetOptions(options)
		m.useSelect = true

		// Pre-select default if available
		if v.Value != "" {
			m.selectModel.Select(v.Value)
		}
	} else {
		m.useSelect = false

		// Set default value
		if v.Value != "" {
			m.textInput.SetValue(v.Value)
		}

		// Configure for secret input
		if v.Annotation != nil && v.Annotation.IsSecret {
			m.textInput.EchoMode = textinput.EchoPassword
			m.textInput.EchoCharacter = '•'
		} else {
			m.textInput.EchoMode = textinput.EchoNormal
		}
	}
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Handle exit prompt
		if m.showExitPrompt {
			return m.handleExitPrompt(msg)
		}

		switch msg.String() {
		case "ctrl+c":
			if m.hasUnsavedChanges {
				m.showExitPrompt = true
				return m, nil
			}
			m.State = StateAborted
			return m, tea.Quit

		case "ctrl+d":
			// Skip optional variable
			v := m.CurrentVariable()
			if v != nil && v.Annotation != nil && v.Annotation.IsOptional {
				m.Values[v.Name] = ""
				m.hasUnsavedChanges = true
				return m.nextVariable()
			}

		case "enter":
			return m.submitInput()

		case "tab":
			// Auto-complete with default if available
			v := m.CurrentVariable()
			if v != nil && v.Value != "" && m.textInput.Value() == "" {
				m.textInput.SetValue(v.Value)
			}
		}
	}

	// Update input components
	var cmd tea.Cmd
	if m.useSelect {
		m.selectModel, cmd = m.selectModel.Update(msg)
	} else {
		m.textInput, cmd = m.textInput.Update(msg)
	}

	return m, cmd
}

func (m Model) handleExitPrompt(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "h":
		m.exitChoice = 0
	case "right", "l":
		m.exitChoice = 1
	case "enter":
		if m.exitChoice == 0 {
			// Discard
			m.State = StateAborted
			m.Values = make(map[string]string) // Clear values
		} else {
			// Save partial
			m.State = StateComplete
		}
		return m, tea.Quit
	case "escape":
		m.showExitPrompt = false
	}
	return m, nil
}

func (m Model) submitInput() (tea.Model, tea.Cmd) {
	v := m.CurrentVariable()
	if v == nil {
		m.State = StateComplete
		return m, tea.Quit
	}

	// Get value
	var value string
	if m.useSelect {
		if !m.selectModel.HasSelection() {
			// Must select an option
			return m, nil
		}
		value = m.selectModel.Value()
	} else {
		value = m.textInput.Value()
	}

	// Validate
	if v.Annotation != nil {
		if err := validator.ValidateValue(value, v.Annotation); err != nil {
			m.Error = err
			m.State = StateValidating
			return m, nil
		}
	}

	// Store value
	m.Values[v.Name] = value
	m.hasUnsavedChanges = true
	m.Error = nil
	m.State = StatePrompting

	return m.nextVariable()
}

func (m Model) nextVariable() (tea.Model, tea.Cmd) {
	m.CurrentIndex++
	if m.CurrentIndex >= len(m.Variables) {
		m.State = StateComplete
		return m, tea.Quit
	}

	m.setupCurrentInput()
	return m, textinput.Blink
}

// View implements tea.Model.
func (m Model) View() string {
	if m.showExitPrompt {
		return m.viewExitPrompt()
	}

	var b strings.Builder

	// Header
	b.WriteString(components.RenderHeader("Krakenv Configuration Wizard"))
	b.WriteString("\n\n")

	// Progress
	progress := fmt.Sprintf("Variable %d of %d", m.CurrentIndex+1, len(m.Variables))
	b.WriteString(components.MutedStyle.Render(progress))
	b.WriteString("\n\n")

	v := m.CurrentVariable()
	if v == nil {
		b.WriteString(components.RenderSuccess("All variables configured!"))
		return b.String()
	}

	// Variable name
	b.WriteString(components.BoldStyle.Render(v.Name))
	b.WriteString("\n")

	// Prompt
	if v.Annotation != nil {
		prompt := components.RenderPrompt(v.Annotation.PromptText, v.Annotation.IsOptional, v.Annotation.IsSecret)
		b.WriteString(prompt)
		b.WriteString("\n")

		// Type constraint
		constraint := formatConstraint(v.Annotation)
		if constraint != "" {
			b.WriteString(components.RenderConstraint(constraint))
			b.WriteString("\n")
		}
	}

	// Default value hint
	if v.Value != "" && !m.useSelect {
		b.WriteString(components.RenderDefault(v.Value))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Input
	if m.useSelect {
		b.WriteString(m.selectModel.View())
	} else {
		b.WriteString(m.textInput.View())
	}
	b.WriteString("\n")

	// Error message
	if m.Error != nil {
		b.WriteString("\n")
		b.WriteString(components.RenderError(m.Error.Error()))
	}

	// Help
	b.WriteString("\n\n")
	help := "Enter: submit • Tab: use default • Ctrl+C: exit"
	if v.Annotation != nil && v.Annotation.IsOptional {
		help += " • Ctrl+D: skip"
	}
	b.WriteString(components.RenderFooter(help))

	return b.String()
}

func (m Model) viewExitPrompt() string {
	var b strings.Builder

	b.WriteString(components.RenderHeader("Krakenv Configuration Wizard"))
	b.WriteString("\n\n")

	b.WriteString(components.WarningStyle.Render("You have unsaved changes!"))
	b.WriteString("\n\n")
	b.WriteString("What would you like to do?\n\n")

	discardStyle := components.MutedStyle
	saveStyle := components.MutedStyle

	if m.exitChoice == 0 {
		discardStyle = components.ErrorStyle.Bold(true)
	} else {
		saveStyle = components.SuccessStyle.Bold(true)
	}

	b.WriteString("  ")
	b.WriteString(discardStyle.Render("[ Discard changes ]"))
	b.WriteString("    ")
	b.WriteString(saveStyle.Render("[ Save progress ]"))
	b.WriteString("\n\n")

	b.WriteString(components.RenderFooter("←/→: select • Enter: confirm • Esc: cancel"))

	return b.String()
}

func formatConstraint(ann *parser.Annotation) string {
	if ann == nil {
		return ""
	}

	parts := []string{ann.Type.String()}

	switch ann.Type {
	case parser.TypeInt, parser.TypeNumeric:
		min := ann.GetConstraint("min")
		max := ann.GetConstraint("max")
		if min != "" && max != "" {
			parts = append(parts, fmt.Sprintf("%s-%s", min, max))
		} else if min != "" {
			parts = append(parts, fmt.Sprintf(">=%s", min))
		} else if max != "" {
			parts = append(parts, fmt.Sprintf("<=%s", max))
		}
	case parser.TypeString:
		minlen := ann.GetConstraint("minlen")
		maxlen := ann.GetConstraint("maxlen")
		if minlen != "" || maxlen != "" {
			parts = append(parts, fmt.Sprintf("len:%s-%s", minlen, maxlen))
		}
		if pattern := ann.GetConstraint("pattern"); pattern != "" {
			parts = append(parts, "pattern")
		}
	case parser.TypeEnum:
		options := ann.GetConstraint("options")
		parts = append(parts, strings.ReplaceAll(options, ",", "|"))
	}

	return strings.Join(parts, ": ")
}

// GetValues returns the collected values.
func (m Model) GetValues() map[string]string {
	return m.Values
}

// IsComplete returns true if the wizard completed successfully.
func (m Model) IsComplete() bool {
	return m.State == StateComplete
}

// IsAborted returns true if the wizard was aborted.
func (m Model) IsAborted() bool {
	return m.State == StateAborted
}
