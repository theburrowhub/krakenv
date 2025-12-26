// Package tui provides the terminal user interface for krakenv.
package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/theburrowhub/krakenv/internal/tui/components"
)

// AppState represents the current state of the TUI application.
type AppState int

const (
	// StateInit is the initial loading state.
	StateInit AppState = iota
	// StateWizard is the variable configuration wizard.
	StateWizard
	// StateInspect is the inspection/diff view.
	StateInspect
	// StateComplete is the completion state.
	StateComplete
	// StateError is the error state.
	StateError
)

// AppModel is the main TUI application model.
type AppModel struct {
	State    AppState
	Title    string
	SubTitle string
	Error    error
	Width    int
	Height   int
	quitting bool
}

// NewAppModel creates a new app model.
func NewAppModel(title string) AppModel {
	return AppModel{
		State: StateInit,
		Title: title,
	}
}

// Init implements tea.Model.
func (m AppModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "q":
			if m.State == StateComplete || m.State == StateError {
				m.quitting = true
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

// View implements tea.Model.
func (m AppModel) View() string {
	if m.quitting {
		return ""
	}

	header := components.RenderHeader(m.Title)

	var content string
	switch m.State {
	case StateInit:
		content = components.RenderInfo("Loading...")
	case StateComplete:
		content = components.RenderSuccess("Operation completed successfully!")
	case StateError:
		if m.Error != nil {
			content = components.RenderError(m.Error.Error())
		} else {
			content = components.RenderError("An error occurred")
		}
	default:
		content = ""
	}

	footer := components.RenderFooter("Press Ctrl+C to exit")

	return fmt.Sprintf("%s\n\n%s\n\n%s", header, content, footer)
}

// SetState changes the application state.
func (m *AppModel) SetState(state AppState) {
	m.State = state
}

// SetError sets an error and changes to error state.
func (m *AppModel) SetError(err error) {
	m.Error = err
	m.State = StateError
}

// IsQuitting returns true if the app is quitting.
func (m AppModel) IsQuitting() bool {
	return m.quitting
}

// Run starts the TUI application.
func Run(m tea.Model) error {
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
