package commandui

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/parfenovvs/lazylogcat/internal/model"
	"github.com/parfenovvs/lazylogcat/internal/tui/theme"
)

// TextInputConfig holds the parameters for creating a TextInputModel.
type TextInputConfig struct {
	Title       string
	Footer      string
	Placeholder string
	Value       string
	Mode        model.TextFilterMode     // Initial filter mode (contains/exact/regex)
	ModeEnabled bool                     // When true, TAB cycles through filter modes
	ValidateFn  func(value string) error // Optional validation run on Enter; nil means no validation
}

// TextInputModel is a reusable text input dialog widget with optional validation.
type TextInputModel struct {
	title       string
	footer      string
	input       textinput.Model
	mode        model.TextFilterMode
	modeEnabled bool
	validateFn  func(value string) error
	errorMsg    string
	submitted   bool
}

// NewTextInput creates a new TextInputModel from the given config.
func NewTextInput(cfg TextInputConfig) TextInputModel {
	ti := textinput.New()
	ti.Placeholder = cfg.Placeholder
	ti.CharLimit = 100
	ti.SetWidth(30)
	ti.SetValue(cfg.Value)
	ti.Focus()

	footer := cfg.Footer

	return TextInputModel{
		title:       cfg.Title,
		footer:      footer,
		input:       ti,
		mode:        cfg.Mode,
		modeEnabled: cfg.ModeEnabled,
		validateFn:  cfg.ValidateFn,
	}
}

// Update handles messages. Returns the updated model and a tea.Cmd.
// After calling Update, check Submitted() to see if the user pressed Enter successfully.
func (m TextInputModel) Update(msg tea.Msg) (TextInputModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		key := msg.String()
		if key == "enter" {
			value := strings.TrimSpace(m.input.Value())
			if m.validateFn != nil && value != "" {
				if err := m.validateFn(value); err != nil {
					m.errorMsg = err.Error()
					return m, nil
				}
			}
			m.submitted = true
			return m, nil
		}

		if key == "tab" && m.modeEnabled {
			m.mode = m.mode.Next()
			return m, nil
		}

		// Forward other keys to the text input, clear error on change
		prevValue := m.input.Value()
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		if m.input.Value() != prevValue {
			m.errorMsg = ""
		}
		return m, cmd

	default:
		// Forward non-key messages (e.g. cursor blink)
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
}

// Submitted returns true if the user pressed Enter and validation passed.
func (m TextInputModel) Submitted() bool {
	return m.submitted
}

// Value returns the current trimmed text input value.
func (m TextInputModel) Value() string {
	return strings.TrimSpace(m.input.Value())
}

// Mode returns the currently selected filter mode.
func (m TextInputModel) Mode() model.TextFilterMode {
	return m.mode
}

// View renders the text input dialog.
func (m TextInputModel) View() string {
	title := dialogTitleWithESC(m.title)

	var footer string
	if m.modeEnabled {
		footer = theme.DialogHelp().Render(m.renderModeHint())
	} else {
		footer = theme.DialogHelp().Render(m.footer)
	}

	var errorLine string
	if m.errorMsg != "" {
		errorLine = "\n" + theme.DialogError().Render(m.errorMsg)
	}

	input := theme.DialogSearch().Render(m.input.View())
	content := title + "\n\n" + input + errorLine + "\n\n" + footer
	return dialogStyle().Render(content)
}

// renderModeHint builds the "TAB contains | exact | regex" footer string
// with the active mode highlighted.
func (m TextInputModel) renderModeHint() string {
	modes := []struct {
		mode model.TextFilterMode
		name string
	}{
		{model.FilterModeContains, "contains"},
		{model.FilterModeExact, "exact"},
		{model.FilterModeRegex, "regex"},
	}

	active := lipgloss.NewStyle().Bold(true).Foreground(theme.ColorPrimary)
	muted := lipgloss.NewStyle().Foreground(theme.ColorMuted)

	parts := make([]string, len(modes))
	for i, md := range modes {
		if md.mode == m.mode {
			parts[i] = active.Render(md.name)
		} else {
			parts[i] = muted.Render(md.name)
		}
	}

	return "TAB " + strings.Join(parts, muted.Render(" | "))
}
