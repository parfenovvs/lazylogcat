package commandui

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/parfenovvs/lazylogcat/internal/model"
	"github.com/parfenovvs/lazylogcat/internal/tui/theme"
)

type CommandDialogLevelSelectedMsg struct {
	Level model.Level
}

var levelEntries = []struct {
	level model.Level
	name  string
}{
	{model.LvlV, "Verbose"},
	{model.LvlD, "Debug"},
	{model.LvlI, "Info"},
	{model.LvlW, "Warning"},
	{model.LvlE, "Error"},
	{model.LvlF, "Fatal"},
}

func newLevelTable(currentLevel model.Level) (table.Model, map[int]model.Level) {
	columns := []table.Column{
		{Title: "", Width: 5},
		{Title: "", Width: 12},
		{Title: "", Width: 3},
	}

	levelMap := make(map[int]model.Level)
	var rows []table.Row
	initialCursor := 0
	for i, entry := range levelEntries {
		levelMap[i] = entry.level
		marker := ""
		if entry.level == currentLevel {
			marker = "●"
			initialCursor = i
		}
		rows = append(rows, table.Row{string(entry.level), entry.name, marker})
	}

	t := newTable(columns, rows, len(rows))
	t.SetCursor(initialCursor)

	return t, levelMap
}

func (m CommandDialogModel) updateLogLevel(msg tea.KeyMsg, key string) (CommandDialogModel, tea.Cmd) {
	if key == "enter" {
		if lvl, ok := m.levelMap[m.levelTable.Cursor()]; ok {
			return m, func() tea.Msg { return CommandDialogLevelSelectedMsg{Level: lvl} }
		}
		return m, nil
	}

	m.levelTable, _ = m.levelTable.Update(msg)
	return m, nil
}

func (m CommandDialogModel) viewLogLevel() string {
	title := lipgloss.NewStyle().Bold(true).Render("Log Level")
	footer := lipgloss.NewStyle().Foreground(theme.FGHelp).Render("esc to close")
	content := title + "\n" + m.levelTable.View() + "\n" + footer
	return dialogStyle().Render(content)
}
