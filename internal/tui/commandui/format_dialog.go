package commandui

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/parfenovvs/lazylogcat/internal/model"
	"github.com/parfenovvs/lazylogcat/internal/tui/theme"
)

type CommandDialogFormatSelectedMsg struct {
	Format string
}

func newFormatTable(currentFormat string) (table.Model, map[int]string) {
	columns := []table.Column{
		{Title: "", Width: 12},
		{Title: "", Width: 3},
	}

	formatMap := make(map[int]string)
	var rows []table.Row
	initialCursor := 0
	for i, name := range model.AllFormats {
		formatMap[i] = name
		marker := ""
		if name == currentFormat {
			marker = "●"
			initialCursor = i
		}
		rows = append(rows, table.Row{name, marker})
	}

	t := newTable(columns, rows, len(rows))
	t.SetCursor(initialCursor)

	return t, formatMap
}

func (m CommandDialogModel) updateFormat(msg tea.KeyMsg, key string) (CommandDialogModel, tea.Cmd) {
	if key == "enter" {
		if fmt, ok := m.formatMap[m.formatTable.Cursor()]; ok {
			return m, func() tea.Msg { return CommandDialogFormatSelectedMsg{Format: fmt} }
		}
		return m, nil
	}

	m.formatTable, _ = m.formatTable.Update(msg)
	return m, nil
}

func (m CommandDialogModel) viewFormat() string {
	title := lipgloss.NewStyle().Bold(true).Render("Format")
	footer := lipgloss.NewStyle().Foreground(theme.FGHelp).Render("esc to close")
	content := title + "\n" + m.formatTable.View() + "\n" + footer
	return dialogStyle().Render(content)
}
