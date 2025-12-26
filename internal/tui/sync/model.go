// Package sync provides a TUI for syncing env files interactively.
package sync

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/theburrowhub/krakenv/internal/inspector"
	"github.com/theburrowhub/krakenv/internal/parser"
	"github.com/theburrowhub/krakenv/internal/validator"
)

// Action represents what to do with a discrepancy.
type Action int

const (
	ActionSkip Action = iota
	ActionAdd
	ActionRemove
	ActionAddToDist
)

// Resolution holds the user's decision for a variable.
type Resolution struct {
	Variable   parser.Variable
	Action     Action
	NewValue   string
	Annotation *parser.Annotation // For AddToDist with annotation
}

// State represents the wizard state.
type State int

const (
	StateMissing State = iota
	StateInvalid
	StateExtra
	StateAddToDist // Sub-wizard for adding to distributable
	StateConfirm
	StateDone
	StateAborted
)

// AddToDistStep represents steps in the add-to-dist sub-wizard.
type AddToDistStep int

const (
	StepType AddToDistStep = iota
	StepPrompt
	StepOptional
	StepSecret
	StepConfirmAdd
)

// Styles for the TUI.
var (
	// Colors
	colorPrimary   = lipgloss.Color("#6C5CE7")
	colorSecondary = lipgloss.Color("#00B894")
	colorAccent    = lipgloss.Color("#FDCB6E")
	colorError     = lipgloss.Color("#D63031")
	colorMuted     = lipgloss.Color("#636E72")
	colorText      = lipgloss.Color("#DFE6E9")

	// Header styles
	headerStyle = lipgloss.NewStyle().
			Background(colorPrimary).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Padding(1, 2).
			Width(80)

	taglineStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Italic(true).
			MarginTop(0)

	// Meta block styles
	metaBlockStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorMuted).
			Padding(1, 2).
			Width(76).
			MarginTop(1).
			MarginBottom(1)

	metaLabelStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Width(14)

	metaValueStyle = lipgloss.NewStyle().
			Foreground(colorText).
			Bold(true)

	// Options block styles
	optionsBlockStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorSecondary).
				Padding(1, 2).
				Width(76).
				MarginBottom(1)

	optionStyle = lipgloss.NewStyle().
			Foreground(colorText)

	optionSelectedStyle = lipgloss.NewStyle().
				Foreground(colorSecondary).
				Bold(true)

	optionKeyStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true).
			Width(4)

	// Input styles
	inputBlockStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).
			Padding(1, 2).
			Width(76).
			MarginBottom(1)

	promptStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	hintStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Italic(true)

	errorMsgStyle = lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true)

	// Footer styles
	footerStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			MarginTop(1).
			Width(76)

	footerKeyStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	footerDescStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	// Progress styles
	progressStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Bold(true)

	// Problem type styles
	problemMissingStyle = lipgloss.NewStyle().
				Foreground(colorAccent).
				Bold(true)

	problemInvalidStyle = lipgloss.NewStyle().
				Foreground(colorError).
				Bold(true)

	problemExtraStyle = lipgloss.NewStyle().
				Foreground(colorSecondary).
				Bold(true)
)

// Model is the Bubble Tea model for sync.
type Model struct {
	result      *inspector.InspectionResult
	distFile    *parser.EnvFile
	targetFile  *parser.EnvFile
	state       State
	index       int
	resolutions []Resolution
	textInput   textinput.Model
	menuChoice  int
	width       int
	height      int
	err         error

	// AddToDist sub-wizard state
	addToDistStep AddToDistStep
	addToDistVar  parser.Variable // Variable being added
	inferredType  parser.VariableType
	selectedType  int // Menu choice for type
	promptText    string
	isOptional    bool
	isSecret      bool
}

// New creates a new sync model.
func New(result *inspector.InspectionResult, distFile, targetFile *parser.EnvFile) Model {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 512
	ti.Width = 60
	ti.Prompt = "  "

	initialState := StateMissing
	if len(result.MissingInEnv) == 0 {
		initialState = StateInvalid
		if len(result.InvalidValues) == 0 {
			initialState = StateExtra
			if len(result.ExtraInEnv) == 0 {
				initialState = StateConfirm
			}
		}
	}

	return Model{
		result:      result,
		distFile:    distFile,
		targetFile:  targetFile,
		state:       initialState,
		resolutions: make([]Resolution, 0),
		textInput:   ti,
	}
}

// inferType attempts to infer the variable type from its value.
func inferType(value string) parser.VariableType {
	value = strings.TrimSpace(value)

	// Empty value - default to string
	if value == "" {
		return parser.TypeString
	}

	// Boolean check
	lower := strings.ToLower(value)
	if lower == "true" || lower == "false" ||
		lower == "yes" || lower == "no" ||
		lower == "on" || lower == "off" ||
		value == "1" || value == "0" {
		return parser.TypeBoolean
	}

	// Integer check
	if _, err := strconv.ParseInt(value, 10, 64); err == nil {
		return parser.TypeInt
	}

	// Numeric (float) check
	if _, err := strconv.ParseFloat(value, 64); err == nil {
		return parser.TypeNumeric
	}

	// JSON object check
	if (strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}")) ||
		(strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]")) {
		return parser.TypeObject
	}

	// Default to string
	return parser.TypeString
}

// typeToIndex converts a VariableType to menu index.
func typeToIndex(t parser.VariableType) int {
	switch t {
	case parser.TypeString:
		return 0
	case parser.TypeInt:
		return 1
	case parser.TypeNumeric:
		return 2
	case parser.TypeBoolean:
		return 3
	case parser.TypeEnum:
		return 4
	case parser.TypeObject:
		return 5
	default:
		return 0
	}
}

// indexToType converts menu index to VariableType.
func indexToType(i int) parser.VariableType {
	switch i {
	case 0:
		return parser.TypeString
	case 1:
		return parser.TypeInt
	case 2:
		return parser.TypeNumeric
	case 3:
		return parser.TypeBoolean
	case 4:
		return parser.TypeEnum
	case 5:
		return parser.TypeObject
	default:
		return parser.TypeString
	}
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.state == StateAddToDist {
				// Cancel sub-wizard, go back to extra
				m.state = StateExtra
				return m, nil
			}
			m.state = StateAborted
			return m, tea.Quit

		case "enter":
			return m.handleEnter()

		case "tab", "s":
			if m.state == StateAddToDist {
				return m.handleAddToDistSkip()
			}
			return m.skipCurrent()

		case "up", "k":
			if m.state == StateExtra && m.menuChoice > 0 {
				m.menuChoice--
			}
			if m.state == StateAddToDist && m.addToDistStep == StepType && m.selectedType > 0 {
				m.selectedType--
			}

		case "down", "j":
			if m.state == StateExtra && m.menuChoice < 2 {
				m.menuChoice++
			}
			if m.state == StateAddToDist && m.addToDistStep == StepType && m.selectedType < 5 {
				m.selectedType++
			}

		case "y", "Y":
			if m.state == StateAddToDist {
				if m.addToDistStep == StepOptional {
					m.isOptional = true
					m.addToDistStep = StepSecret
					return m, nil
				}
				if m.addToDistStep == StepSecret {
					m.isSecret = true
					m.addToDistStep = StepConfirmAdd
					return m, nil
				}
			}

		case "n", "N":
			if m.state == StateAddToDist {
				if m.addToDistStep == StepOptional {
					m.isOptional = false
					m.addToDistStep = StepSecret
					return m, nil
				}
				if m.addToDistStep == StepSecret {
					m.isSecret = false
					m.addToDistStep = StepConfirmAdd
					return m, nil
				}
			}

		// Quick keys for extra variables
		case "1":
			if m.state == StateExtra {
				m.menuChoice = 0
				return m.handleEnter()
			}
		case "2":
			if m.state == StateExtra {
				m.menuChoice = 1
				return m.handleEnter()
			}
		case "3":
			if m.state == StateExtra {
				m.menuChoice = 2
				return m.handleEnter()
			}
		}
	}

	if m.state == StateMissing || m.state == StateInvalid ||
		(m.state == StateAddToDist && m.addToDistStep == StepPrompt) {
		m.textInput, cmd = m.textInput.Update(msg)
	}

	return m, cmd
}

func (m Model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.state {
	case StateMissing:
		return m.handleMissingEnter()
	case StateInvalid:
		return m.handleInvalidEnter()
	case StateExtra:
		return m.handleExtraEnter()
	case StateAddToDist:
		return m.handleAddToDistEnter()
	case StateConfirm:
		m.state = StateDone
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleMissingEnter() (tea.Model, tea.Cmd) {
	if m.index >= len(m.result.MissingInEnv) {
		return m.advanceState()
	}

	v := m.result.MissingInEnv[m.index]
	value := m.textInput.Value()

	if v.Annotation != nil && value != "" {
		if err := validator.ValidateValue(value, v.Annotation); err != nil {
			m.err = err
			return m, nil
		}
	}

	if value == "" && v.Value != "" {
		value = v.Value
	}

	if value == "" && v.Annotation != nil && v.Annotation.IsOptional {
		m.resolutions = append(m.resolutions, Resolution{
			Variable: v,
			Action:   ActionSkip,
		})
	} else if value != "" {
		m.resolutions = append(m.resolutions, Resolution{
			Variable: v,
			Action:   ActionAdd,
			NewValue: value,
		})
	} else {
		m.err = fmt.Errorf("value required (not optional)")
		return m, nil
	}

	m.err = nil
	m.index++
	m.textInput.Reset()

	if m.index < len(m.result.MissingInEnv) {
		next := m.result.MissingInEnv[m.index]
		if next.Value != "" {
			m.textInput.SetValue(next.Value)
		}
	}

	if m.index >= len(m.result.MissingInEnv) {
		return m.advanceState()
	}

	return m, nil
}

func (m Model) handleInvalidEnter() (tea.Model, tea.Cmd) {
	if m.index >= len(m.result.InvalidValues) {
		return m.advanceState()
	}

	valErr := m.result.InvalidValues[m.index]
	value := m.textInput.Value()

	distVar := m.distFile.GetVariable(valErr.Variable)
	if distVar != nil && distVar.Annotation != nil {
		if err := validator.ValidateValue(value, distVar.Annotation); err != nil {
			m.err = err
			return m, nil
		}
	}

	if value == "" {
		m.err = fmt.Errorf("value required")
		return m, nil
	}

	m.resolutions = append(m.resolutions, Resolution{
		Variable: parser.Variable{Name: valErr.Variable},
		Action:   ActionAdd,
		NewValue: value,
	})

	m.err = nil
	m.index++
	m.textInput.Reset()

	if m.index >= len(m.result.InvalidValues) {
		return m.advanceState()
	}

	return m, nil
}

func (m Model) handleExtraEnter() (tea.Model, tea.Cmd) {
	if m.index >= len(m.result.ExtraInEnv) {
		return m.advanceState()
	}

	v := m.result.ExtraInEnv[m.index]

	switch m.menuChoice {
	case 0: // Keep (skip)
		m.resolutions = append(m.resolutions, Resolution{
			Variable: v,
			Action:   ActionSkip,
		})
		m.index++
		m.menuChoice = 0
		if m.index >= len(m.result.ExtraInEnv) {
			return m.advanceState()
		}

	case 1: // Remove
		m.resolutions = append(m.resolutions, Resolution{
			Variable: v,
			Action:   ActionRemove,
		})
		m.index++
		m.menuChoice = 0
		if m.index >= len(m.result.ExtraInEnv) {
			return m.advanceState()
		}

	case 2: // Add to distributable - enter sub-wizard
		m.state = StateAddToDist
		m.addToDistStep = StepType
		m.addToDistVar = v
		m.inferredType = inferType(v.Value)
		m.selectedType = typeToIndex(m.inferredType)
		m.promptText = ""
		m.isOptional = false
		m.isSecret = false
		m.textInput.Reset()
		m.textInput.SetValue(fmt.Sprintf("Enter %s", v.Name))
	}

	return m, nil
}

func (m Model) handleAddToDistEnter() (tea.Model, tea.Cmd) {
	switch m.addToDistStep {
	case StepType:
		// Type selected, move to prompt
		m.addToDistStep = StepPrompt
		m.textInput.Focus()

	case StepPrompt:
		// Save prompt and move to optional
		m.promptText = m.textInput.Value()
		if m.promptText == "" {
			m.promptText = fmt.Sprintf("Enter %s", m.addToDistVar.Name)
		}
		m.addToDistStep = StepOptional

	case StepOptional:
		// Handled by y/n keys
		m.addToDistStep = StepSecret

	case StepSecret:
		// Handled by y/n keys
		m.addToDistStep = StepConfirmAdd

	case StepConfirmAdd:
		// Create annotation and add resolution
		ann := &parser.Annotation{
			PromptText: m.promptText,
			Type:       indexToType(m.selectedType),
			IsOptional: m.isOptional,
			IsSecret:   m.isSecret,
		}

		m.resolutions = append(m.resolutions, Resolution{
			Variable:   m.addToDistVar,
			Action:     ActionAddToDist,
			Annotation: ann,
		})

		// Return to extra state for next variable
		m.state = StateExtra
		m.index++
		m.menuChoice = 0

		if m.index >= len(m.result.ExtraInEnv) {
			return m.advanceState()
		}
	}

	return m, nil
}

func (m Model) handleAddToDistSkip() (tea.Model, tea.Cmd) {
	// Skip adding annotation, just add as simple variable
	m.resolutions = append(m.resolutions, Resolution{
		Variable: m.addToDistVar,
		Action:   ActionAddToDist,
	})

	m.state = StateExtra
	m.index++
	m.menuChoice = 0

	if m.index >= len(m.result.ExtraInEnv) {
		return m.advanceState()
	}

	return m, nil
}

func (m Model) skipCurrent() (tea.Model, tea.Cmd) {
	switch m.state {
	case StateMissing:
		if m.index < len(m.result.MissingInEnv) {
			v := m.result.MissingInEnv[m.index]
			m.resolutions = append(m.resolutions, Resolution{
				Variable: v,
				Action:   ActionSkip,
			})
			m.index++
			m.textInput.Reset()
			if m.index >= len(m.result.MissingInEnv) {
				return m.advanceState()
			}
		}
	case StateInvalid:
		if m.index < len(m.result.InvalidValues) {
			m.index++
			m.textInput.Reset()
			if m.index >= len(m.result.InvalidValues) {
				return m.advanceState()
			}
		}
	case StateExtra:
		if m.index < len(m.result.ExtraInEnv) {
			v := m.result.ExtraInEnv[m.index]
			m.resolutions = append(m.resolutions, Resolution{
				Variable: v,
				Action:   ActionSkip,
			})
			m.index++
			if m.index >= len(m.result.ExtraInEnv) {
				return m.advanceState()
			}
		}
	}
	return m, nil
}

func (m Model) advanceState() (tea.Model, tea.Cmd) {
	m.index = 0
	m.textInput.Reset()
	m.err = nil

	switch m.state {
	case StateMissing:
		m.state = StateInvalid
		if len(m.result.InvalidValues) == 0 {
			return m.advanceState()
		}
	case StateInvalid:
		m.state = StateExtra
		if len(m.result.ExtraInEnv) == 0 {
			return m.advanceState()
		}
	case StateExtra:
		m.state = StateConfirm
	}

	return m, nil
}

// View renders the UI.
func (m Model) View() string {
	if m.state == StateDone || m.state == StateAborted {
		return ""
	}

	var content strings.Builder

	// Header
	content.WriteString(m.renderHeader())
	content.WriteString("\n")

	// Main content based on state
	switch m.state {
	case StateMissing:
		content.WriteString(m.renderMissing())
	case StateInvalid:
		content.WriteString(m.renderInvalid())
	case StateExtra:
		content.WriteString(m.renderExtra())
	case StateAddToDist:
		content.WriteString(m.renderAddToDist())
	case StateConfirm:
		content.WriteString(m.renderConfirm())
	}

	// Footer
	content.WriteString(m.renderFooter())

	// Center everything
	return lipgloss.Place(m.width, m.height,
		lipgloss.Left, lipgloss.Top,
		lipgloss.NewStyle().Padding(1, 2).Render(content.String()),
	)
}

func (m Model) renderHeader() string {
	title := headerStyle.Render("🐙 KRAKENV SYNC")
	tagline := taglineStyle.Render("When envs get complex, release the krakenv")
	return title + "\n" + tagline
}

func (m Model) renderMeta(problemType, problemDesc string) string {
	var b strings.Builder

	// File info
	b.WriteString(metaLabelStyle.Render("Distributable"))
	b.WriteString(metaValueStyle.Render(m.result.DistPath))
	b.WriteString("\n")

	b.WriteString(metaLabelStyle.Render("Environment"))
	b.WriteString(metaValueStyle.Render(m.result.TargetPath))
	b.WriteString("\n")

	// Stats
	b.WriteString(metaLabelStyle.Render("Variables"))
	stats := fmt.Sprintf("%d missing · %d invalid · %d extra",
		len(m.result.MissingInEnv),
		len(m.result.InvalidValues),
		len(m.result.ExtraInEnv))
	b.WriteString(metaValueStyle.Render(stats))
	b.WriteString("\n")

	// Current problem
	b.WriteString(metaLabelStyle.Render("Problem"))
	b.WriteString(problemType)
	b.WriteString(" ")
	b.WriteString(problemDesc)

	return metaBlockStyle.Render(b.String())
}

func (m Model) renderProgress() string {
	var total, current int
	var label string

	switch m.state {
	case StateMissing:
		total = len(m.result.MissingInEnv)
		current = m.index + 1
		label = "Missing"
	case StateInvalid:
		total = len(m.result.InvalidValues)
		current = m.index + 1
		label = "Invalid"
	case StateExtra:
		total = len(m.result.ExtraInEnv)
		current = m.index + 1
		label = "Extra"
	default:
		return ""
	}

	return progressStyle.Render(fmt.Sprintf("%s %d/%d", label, current, total))
}

func (m Model) renderMissing() string {
	var b strings.Builder

	if m.index >= len(m.result.MissingInEnv) {
		return ""
	}

	v := m.result.MissingInEnv[m.index]

	// Meta block
	problemType := problemMissingStyle.Render("MISSING")
	problemDesc := fmt.Sprintf("Variable %s not found in environment file", metaValueStyle.Render(v.Name))
	b.WriteString(m.renderMeta(problemType, problemDesc))
	b.WriteString("\n")

	// Input block
	var inputContent strings.Builder

	inputContent.WriteString(m.renderProgress())
	inputContent.WriteString("\n\n")

	// Variable info
	inputContent.WriteString(promptStyle.Render(v.Name))
	inputContent.WriteString("\n")

	if v.Annotation != nil {
		inputContent.WriteString(hintStyle.Render(v.Annotation.PromptText))
		inputContent.WriteString("\n")

		typeInfo := fmt.Sprintf("[%s]", v.Annotation.Type.String())
		if v.Annotation.IsOptional {
			typeInfo += " (optional)"
		}
		inputContent.WriteString(hintStyle.Render(typeInfo))
		inputContent.WriteString("\n")
	}

	if v.Value != "" {
		inputContent.WriteString(hintStyle.Render(fmt.Sprintf("Default: %s", v.Value)))
		inputContent.WriteString("\n")
	}

	inputContent.WriteString("\n")
	inputContent.WriteString(m.textInput.View())

	if m.err != nil {
		inputContent.WriteString("\n")
		inputContent.WriteString(errorMsgStyle.Render("✗ " + m.err.Error()))
	}

	b.WriteString(inputBlockStyle.Render(inputContent.String()))

	return b.String()
}

func (m Model) renderInvalid() string {
	var b strings.Builder

	if m.index >= len(m.result.InvalidValues) {
		return ""
	}

	valErr := m.result.InvalidValues[m.index]

	// Meta block
	problemType := problemInvalidStyle.Render("INVALID")
	problemDesc := fmt.Sprintf("Variable %s has invalid value", metaValueStyle.Render(valErr.Variable))
	b.WriteString(m.renderMeta(problemType, problemDesc))
	b.WriteString("\n")

	// Input block
	var inputContent strings.Builder

	inputContent.WriteString(m.renderProgress())
	inputContent.WriteString("\n\n")

	inputContent.WriteString(promptStyle.Render(valErr.Variable))
	inputContent.WriteString("\n")

	inputContent.WriteString(errorMsgStyle.Render("Error: " + valErr.Message))
	inputContent.WriteString("\n")

	inputContent.WriteString(hintStyle.Render("Suggestion: " + valErr.Suggestion))
	inputContent.WriteString("\n")

	if valErr.Example != "" {
		inputContent.WriteString(hintStyle.Render("Example: " + valErr.Example))
		inputContent.WriteString("\n")
	}

	inputContent.WriteString("\n")
	inputContent.WriteString(m.textInput.View())

	if m.err != nil {
		inputContent.WriteString("\n")
		inputContent.WriteString(errorMsgStyle.Render("✗ " + m.err.Error()))
	}

	b.WriteString(inputBlockStyle.Render(inputContent.String()))

	return b.String()
}

func (m Model) renderExtra() string {
	var b strings.Builder

	if m.index >= len(m.result.ExtraInEnv) {
		return ""
	}

	v := m.result.ExtraInEnv[m.index]

	// Meta block
	problemType := problemExtraStyle.Render("EXTRA")
	problemDesc := fmt.Sprintf("Variable %s exists only in environment file", metaValueStyle.Render(v.Name))
	b.WriteString(m.renderMeta(problemType, problemDesc))
	b.WriteString("\n")

	// Options block
	var optContent strings.Builder

	optContent.WriteString(m.renderProgress())
	optContent.WriteString("\n\n")

	optContent.WriteString(promptStyle.Render(v.Name))
	optContent.WriteString(hintStyle.Render(fmt.Sprintf(" = %q", v.Value)))
	optContent.WriteString("\n\n")

	optContent.WriteString(hintStyle.Render("What would you like to do?"))
	optContent.WriteString("\n\n")

	options := []struct {
		key  string
		text string
	}{
		{"1", "Keep in environment file (ignore)"},
		{"2", "Remove from environment file"},
		{"3", "Add to distributable"},
	}

	for i, opt := range options {
		isSelected := i == m.menuChoice

		// Structure: "X [N]  Text" where X is selector (🐙 or space)
		// Using fixed spacing to ensure alignment
		var line string
		if isSelected {
			line = fmt.Sprintf("🐙 [%s]  %s", opt.key, opt.text)
			optContent.WriteString(optionSelectedStyle.Render(line))
		} else {
			optContent.WriteString("   ")
			optContent.WriteString(optionKeyStyle.Render("[" + opt.key + "]"))
			optContent.WriteString(optionStyle.Render("  " + opt.text))
		}
		optContent.WriteString("\n")
	}

	b.WriteString(optionsBlockStyle.Render(optContent.String()))

	return b.String()
}

func (m Model) renderAddToDist() string {
	var b strings.Builder

	v := m.addToDistVar

	// Meta block
	problemType := problemExtraStyle.Render("ADD TO DIST")
	problemDesc := fmt.Sprintf("Configure annotation for %s", metaValueStyle.Render(v.Name))
	b.WriteString(m.renderMeta(problemType, problemDesc))
	b.WriteString("\n")

	// Content block
	var content strings.Builder

	content.WriteString(promptStyle.Render(v.Name))
	content.WriteString(hintStyle.Render(fmt.Sprintf(" = %q", v.Value)))
	content.WriteString("\n")
	content.WriteString(hintStyle.Render(fmt.Sprintf("Inferred type: %s", m.inferredType.String())))
	content.WriteString("\n\n")

	switch m.addToDistStep {
	case StepType:
		content.WriteString(promptStyle.Render("Select variable type:"))
		content.WriteString("\n\n")

		types := []string{"string", "int", "numeric", "boolean", "enum", "object"}
		for i, t := range types {
			isSelected := i == m.selectedType
			isInferred := i == typeToIndex(m.inferredType)

			var line string
			suffix := ""
			if isInferred {
				suffix = " (inferred)"
			}

			if isSelected {
				line = fmt.Sprintf("🐙 [%d]  %s%s", i+1, t, suffix)
				content.WriteString(optionSelectedStyle.Render(line))
			} else {
				content.WriteString("   ")
				content.WriteString(optionKeyStyle.Render(fmt.Sprintf("[%d]", i+1)))
				content.WriteString(optionStyle.Render(fmt.Sprintf("  %s%s", t, suffix)))
			}
			content.WriteString("\n")
		}

	case StepPrompt:
		content.WriteString(promptStyle.Render("Enter prompt message:"))
		content.WriteString("\n")
		content.WriteString(hintStyle.Render("(This will be shown when generating env files)"))
		content.WriteString("\n\n")
		content.WriteString(m.textInput.View())

	case StepOptional:
		content.WriteString(promptStyle.Render("Is this variable optional?"))
		content.WriteString("\n")
		content.WriteString(hintStyle.Render("(Optional variables can be left empty)"))
		content.WriteString("\n\n")
		content.WriteString(optionKeyStyle.Render("[Y]"))
		content.WriteString(optionStyle.Render(" Yes  "))
		content.WriteString(optionKeyStyle.Render("[N]"))
		content.WriteString(optionStyle.Render(" No"))

	case StepSecret:
		content.WriteString(promptStyle.Render("Is this a secret value?"))
		content.WriteString("\n")
		content.WriteString(hintStyle.Render("(Secret values will be masked during input)"))
		content.WriteString("\n\n")
		content.WriteString(optionKeyStyle.Render("[Y]"))
		content.WriteString(optionStyle.Render(" Yes  "))
		content.WriteString(optionKeyStyle.Render("[N]"))
		content.WriteString(optionStyle.Render(" No"))

	case StepConfirmAdd:
		content.WriteString(promptStyle.Render("Annotation Preview:"))
		content.WriteString("\n\n")

		// Build preview
		preview := fmt.Sprintf("%s=%s #prompt:%s|%s",
			v.Name, v.Value, m.promptText, indexToType(m.selectedType).String())
		if m.isOptional {
			preview += ";optional"
		}
		if m.isSecret {
			preview += ";secret"
		}

		content.WriteString(metaValueStyle.Render(preview))
		content.WriteString("\n\n")
		content.WriteString(hintStyle.Render("Press Enter to confirm"))
	}

	b.WriteString(optionsBlockStyle.Render(content.String()))

	return b.String()
}

func (m Model) renderConfirm() string {
	var b strings.Builder

	// Meta block
	problemType := progressStyle.Render("DONE")
	problemDesc := "Review changes before applying"
	b.WriteString(m.renderMeta(problemType, problemDesc))
	b.WriteString("\n")

	// Summary block
	var sumContent strings.Builder

	sumContent.WriteString(promptStyle.Render("Summary of Changes"))
	sumContent.WriteString("\n\n")

	adds := 0
	removes := 0
	addsToDist := 0
	skips := 0

	for _, r := range m.resolutions {
		switch r.Action {
		case ActionAdd:
			adds++
		case ActionRemove:
			removes++
		case ActionAddToDist:
			addsToDist++
		case ActionSkip:
			skips++
		}
	}

	if adds > 0 {
		sumContent.WriteString(lipgloss.NewStyle().Foreground(colorSecondary).Render(
			fmt.Sprintf("  ✓ Add/update %d variable(s) in env file\n", adds)))
	}
	if removes > 0 {
		sumContent.WriteString(lipgloss.NewStyle().Foreground(colorError).Render(
			fmt.Sprintf("  ✗ Remove %d variable(s) from env file\n", removes)))
	}
	if addsToDist > 0 {
		sumContent.WriteString(lipgloss.NewStyle().Foreground(colorAccent).Render(
			fmt.Sprintf("  + Add %d variable(s) to distributable\n", addsToDist)))
	}
	if skips > 0 {
		sumContent.WriteString(hintStyle.Render(
			fmt.Sprintf("  - Skipped %d item(s)\n", skips)))
	}

	if adds == 0 && removes == 0 && addsToDist == 0 {
		sumContent.WriteString(hintStyle.Render("  No changes to apply\n"))
	}

	sumContent.WriteString("\n")
	sumContent.WriteString(promptStyle.Render("Press Enter to apply changes"))

	b.WriteString(optionsBlockStyle.Render(sumContent.String()))

	return b.String()
}

func (m Model) renderFooter() string {
	var parts []string

	switch m.state {
	case StateMissing, StateInvalid:
		parts = []string{
			footerKeyStyle.Render("Enter") + footerDescStyle.Render(" confirm"),
			footerKeyStyle.Render("Tab/s") + footerDescStyle.Render(" skip"),
			footerKeyStyle.Render("q") + footerDescStyle.Render(" quit"),
		}
	case StateExtra:
		parts = []string{
			footerKeyStyle.Render("↑/↓") + footerDescStyle.Render(" navigate"),
			footerKeyStyle.Render("1-3") + footerDescStyle.Render(" quick select"),
			footerKeyStyle.Render("Enter") + footerDescStyle.Render(" confirm"),
			footerKeyStyle.Render("Tab/s") + footerDescStyle.Render(" skip"),
			footerKeyStyle.Render("q") + footerDescStyle.Render(" quit"),
		}
	case StateAddToDist:
		switch m.addToDistStep {
		case StepType:
			parts = []string{
				footerKeyStyle.Render("↑/↓") + footerDescStyle.Render(" navigate"),
				footerKeyStyle.Render("Enter") + footerDescStyle.Render(" select"),
				footerKeyStyle.Render("Tab") + footerDescStyle.Render(" skip annotation"),
				footerKeyStyle.Render("q") + footerDescStyle.Render(" cancel"),
			}
		case StepPrompt:
			parts = []string{
				footerKeyStyle.Render("Enter") + footerDescStyle.Render(" confirm"),
				footerKeyStyle.Render("Tab") + footerDescStyle.Render(" skip annotation"),
				footerKeyStyle.Render("q") + footerDescStyle.Render(" cancel"),
			}
		case StepOptional, StepSecret:
			parts = []string{
				footerKeyStyle.Render("Y") + footerDescStyle.Render(" yes"),
				footerKeyStyle.Render("N") + footerDescStyle.Render(" no"),
				footerKeyStyle.Render("Tab") + footerDescStyle.Render(" skip annotation"),
				footerKeyStyle.Render("q") + footerDescStyle.Render(" cancel"),
			}
		case StepConfirmAdd:
			parts = []string{
				footerKeyStyle.Render("Enter") + footerDescStyle.Render(" add to dist"),
				footerKeyStyle.Render("q") + footerDescStyle.Render(" cancel"),
			}
		}
	case StateConfirm:
		parts = []string{
			footerKeyStyle.Render("Enter") + footerDescStyle.Render(" apply"),
			footerKeyStyle.Render("q") + footerDescStyle.Render(" cancel"),
		}
	}

	return footerStyle.Render(strings.Join(parts, "  │  "))
}

// GetResolutions returns the resolutions after the wizard completes.
func (m Model) GetResolutions() []Resolution {
	return m.resolutions
}

// IsAborted returns true if the user aborted.
func (m Model) IsAborted() bool {
	return m.state == StateAborted
}

// IsDone returns true if the wizard completed.
func (m Model) IsDone() bool {
	return m.state == StateDone
}
