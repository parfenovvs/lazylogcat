package commandui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/parfenovvs/lazylogcat/internal/model"
	"github.com/parfenovvs/lazylogcat/internal/tui/theme"
	"github.com/parfenovvs/lazylogcat/internal/util"
)

// CommandDialogTextInputAppliedMsg is sent when the user confirms the text input value with enter.
type CommandDialogTextInputAppliedMsg struct {
	Command model.Command
	Value   string
}

func newDialogTextInput(placeholder string, currentValue string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 100
	ti.Width = 30
	ti.SetValue(currentValue)
	ti.Focus()
	return ti
}

func textInputTitle(cmd model.Command) string {
	switch cmd {
	case model.CommandPackage:
		return "Package"
	case model.CommandTag:
		return "Tag"
	case model.CommandContent:
		return "Content"
	default:
		return ""
	}
}

func textInputPlaceholder(cmd model.Command) string {
	switch cmd {
	case model.CommandPackage:
		return "Package name..."
	case model.CommandTag:
		return "Tag value..."
	case model.CommandContent:
		return "Search text..."
	default:
		return ""
	}
}

func (m CommandDialogModel) updateTextInput(msg tea.KeyMsg, key string) (CommandDialogModel, tea.Cmd) {
	if key == "enter" {
		value := strings.TrimSpace(m.textInput.Value())

		// Validate package name if non-empty
		if m.textInputCommand == model.CommandPackage && value != "" {
			_, err := util.GetPidByPackageName(m.deviceId, value)
			if err != nil {
				m.textInputError = "Package not found on device"
				return m, nil
			}
		}

		cmd := m.textInputCommand
		return m, func() tea.Msg {
			return CommandDialogTextInputAppliedMsg{Command: cmd, Value: value}
		}
	}

	// Clear error when user changes the input
	prevValue := m.textInput.Value()
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	if m.textInput.Value() != prevValue {
		m.textInputError = ""
	}
	return m, cmd
}

func (m CommandDialogModel) viewTextInput() string {
	title := lipgloss.NewStyle().Bold(true).Render(m.textInputTitle)
	footer := lipgloss.NewStyle().Foreground(theme.FGHelp).Render("enter to apply, esc to cancel")

	var errorLine string
	if m.textInputError != "" {
		errorLine = "\n" + lipgloss.NewStyle().Foreground(theme.FGError).Render(m.textInputError)
	}

	content := title + "\n\n" + m.textInput.View() + errorLine + "\n\n" + footer
	return dialogStyle().Render(content)
}
