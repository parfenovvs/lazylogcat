package commandui

import (
	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"

	"github.com/parfenovvs/lazylogcat/internal/model"
	"github.com/parfenovvs/lazylogcat/internal/tui"
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

func newLevelSingleSelect(currentLevel model.Level) SingleSelectModel {
	var items []SingleSelectItem
	for _, entry := range levelEntries {
		items = append(items, SingleSelectItem{
			Key:     string(entry.level),
			Columns: []string{string(entry.level), entry.name},
		})
	}
	return NewSingleSelect(SingleSelectConfig{
		Title:  "Log Level",
		Footer: "",
		Columns: []table.Column{
			{Title: "", Width: 1},
			{Title: "", Width: tui.DialogWidth - 10},
		},
		Items:      items,
		CurrentKey: string(currentLevel),
	})
}

func (m CommandDialogModel) updateLogLevel(msg tea.KeyPressMsg, key string) (CommandDialogModel, tea.Cmd) {
	var cmd tea.Cmd
	m.singleSelect, cmd = m.singleSelect.Update(msg, key)
	if m.singleSelect.Selected() {
		lvl := model.Level(m.singleSelect.SelectedKey())
		return m, func() tea.Msg { return CommandDialogLevelSelectedMsg{Level: lvl} }
	}
	return m, cmd
}
