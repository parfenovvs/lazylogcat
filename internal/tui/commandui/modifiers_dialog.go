package commandui

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/parfenovvs/lazylogcat/internal/model"
	"github.com/parfenovvs/lazylogcat/internal/tui/theme"
)

type CommandDialogModifiersSelectedMsg struct {
	Modifiers map[string]bool
}

func newModifiersTable(activeModifiers map[string]bool) (table.Model, map[int]string) {
	columns := []table.Column{
		{Title: "", Width: 14},
		{Title: "", Width: 3},
	}

	modifierMap := make(map[int]string)
	var rows []table.Row
	for i, name := range model.AllModifiers {
		modifierMap[i] = name
		marker := ""
		if activeModifiers[name] {
			marker = "✓"
		}
		rows = append(rows, table.Row{name, marker})
	}

	t := newTable(columns, rows, len(rows))

	return t, modifierMap
}

func (m CommandDialogModel) refreshModifierRows() table.Model {
	var rows []table.Row
	for _, name := range model.AllModifiers {
		marker := ""
		if m.tempModifiers[name] {
			marker = "✓"
		}
		rows = append(rows, table.Row{name, marker})
	}
	m.modifiersTable.SetRows(rows)
	return m.modifiersTable
}

func (m CommandDialogModel) updateModifiers(msg tea.KeyMsg, key string) (CommandDialogModel, tea.Cmd) {
	if key == "enter" || key == " " {
		if name, ok := m.modifierMap[m.modifiersTable.Cursor()]; ok {
			if m.tempModifiers[name] {
				delete(m.tempModifiers, name)
			} else {
				m.tempModifiers[name] = true
			}
			m.modifiersTable = m.refreshModifierRows()
		}
		return m, nil
	}

	m.modifiersTable, _ = m.modifiersTable.Update(msg)
	return m, nil
}

func (m CommandDialogModel) viewModifiers() string {
	title := lipgloss.NewStyle().Bold(true).Render("Modifiers")
	footer := lipgloss.NewStyle().Foreground(theme.FGHelp).Render("space/enter toggle, esc to apply")
	content := title + "\n" + m.modifiersTable.View() + "\n" + footer
	return dialogStyle().Render(content)
}
