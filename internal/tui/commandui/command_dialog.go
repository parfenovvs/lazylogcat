package commandui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/parfenovvs/lazylogcat/internal/model"
	"github.com/parfenovvs/lazylogcat/internal/tui/theme"
)

type CommandDialogCloseMsg struct{}

type CommandDialogSelectMsg struct {
	Command model.Command
}

type CommandDialogLevelSelectedMsg struct {
	Level model.Level
}

type dialogState int

const (
	stateCommands dialogState = iota
	stateLogLevel
)

var dialogStyle = func() lipgloss.Style {
	return theme.ActivePanel().
		Padding(1, 2)
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

type CommandDialogModel struct {
	state      dialogState
	table      table.Model
	skipRows   map[int]bool
	commandMap map[int]model.CommandData

	levelTable table.Model
	levelMap   map[int]model.Level
}

func NewDialog(filter model.Filter, format model.Format, softWrap bool) CommandDialogModel {
	resolveValue := func(cmd model.Command) string {
		switch cmd {
		case model.CommandPackage:
			return filter.PackageName
		case model.CommandTag:
			return filter.Tag
		case model.CommandLevel:
			lvl := string(filter.Level)
			if lvl == "" {
				lvl = string(model.LvlV)
			}
			return lvl
		case model.CommandContent:
			return filter.Text
		case model.CommandFormat:
			return format.Value()
		case model.CommandModifiers:
			mods := format.Modifiers()
			switch len(mods) {
			case 0:
				return ""
			case 1:
				return mods[0]
			default:
				return fmt.Sprintf("[%d] mods", len(mods))
			}
		case model.CommandToggleWrap:
			if softWrap {
				return "on"
			}
			return "off"
		default:
			return ""
		}
	}

	columns := []table.Column{
		{Title: "", Width: 16},
		{Title: "", Width: 10},
		{Title: "", Width: 10},
	}

	skipRows := make(map[int]bool)
	commandMap := make(map[int]model.CommandData)
	var rows []table.Row
	for i, group := range model.Commands() {
		if i > 0 {
			skipRows[len(rows)] = true
			rows = append(rows, table.Row{"", "", ""})
		}
		skipRows[len(rows)] = true
		groupName := lipgloss.NewStyle().Bold(true).Render(group.Name)
		rows = append(rows, table.Row{groupName, "", ""})
		for _, cmd := range group.Commands {
			commandMap[len(rows)] = cmd
			value := truncateMiddle(resolveValue(cmd.Command), 10)
			rows = append(rows, table.Row{cmd.Name, value, cmd.Shortcut})
		}
	}

	t := newTable(columns, rows, len(rows))
	t.SetCursor(1) // Skip the first group header

	currentLevel := filter.Level
	if currentLevel == "" {
		currentLevel = model.LvlV
	}
	levelTable, levelMap := newLevelTable(currentLevel)

	return CommandDialogModel{
		state:      stateCommands,
		table:      t,
		skipRows:   skipRows,
		commandMap: commandMap,
		levelTable: levelTable,
		levelMap:   levelMap,
	}
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

func newTable(columns []table.Column, rows []table.Row, height int) table.Model {
	km := table.DefaultKeyMap()
	km.GotoTop.SetEnabled(false)
	km.GotoBottom.SetEnabled(false)
	km.HalfPageUp.SetEnabled(false)
	km.HalfPageDown.SetEnabled(false)
	km.PageDown.SetEnabled(false)

	s := table.Styles{
		Header:   lipgloss.NewStyle(),
		Cell:     lipgloss.NewStyle().Padding(0, 1),
		Selected: lipgloss.NewStyle().Bold(true).Foreground(theme.FGSelected).Background(theme.BGCursor),
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(height),
		table.WithFocused(true),
		table.WithKeyMap(km),
	)
	t.SetStyles(s)

	return t
}

func (m CommandDialogModel) Update(msg tea.Msg) (CommandDialogModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		if key == "ctrl+p" || key == "esc" {
			return m, func() tea.Msg { return CommandDialogCloseMsg{} }
		}

		switch m.state {
		case stateCommands:
			return m.updateCommands(msg, key)
		case stateLogLevel:
			return m.updateLogLevel(msg, key)
		}
	}

	return m, nil
}

func (m CommandDialogModel) updateCommands(msg tea.KeyMsg, key string) (CommandDialogModel, tea.Cmd) {
	if key == "enter" {
		if cmdData, ok := m.commandMap[m.table.Cursor()]; ok {
			if cmdData.Type == model.CommandTypeNavigation && cmdData.Command == model.CommandLevel {
				m.state = stateLogLevel
				return m, nil
			}
			return m, func() tea.Msg { return CommandDialogSelectMsg{Command: cmdData.Command} }
		}
		return m, nil
	}

	prevCursor := m.table.Cursor()
	m.table, _ = m.table.Update(msg)
	newCursor := m.table.Cursor()

	if m.skipRows[newCursor] && newCursor != prevCursor {
		dir := 1
		if newCursor < prevCursor {
			dir = -1
		}
		rowCount := len(m.table.Rows())
		target := newCursor + dir
		for target >= 0 && target < rowCount && m.skipRows[target] {
			target += dir
		}
		if target >= 0 && target < rowCount {
			m.table.SetCursor(target)
		} else {
			m.table.SetCursor(prevCursor)
		}
	}

	return m, nil
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

func (m CommandDialogModel) View() string {
	switch m.state {
	case stateLogLevel:
		return m.viewLogLevel()
	default:
		return m.viewCommands()
	}
}

func (m CommandDialogModel) viewCommands() string {
	title := lipgloss.NewStyle().Bold(true).Render("Command List")
	footer := lipgloss.NewStyle().Foreground(theme.FGHelp).Render("esc to close")
	content := title + "\n" + m.table.View() + "\n" + footer
	return dialogStyle().Render(content)
}

func (m CommandDialogModel) viewLogLevel() string {
	title := lipgloss.NewStyle().Bold(true).Render("Log Level")
	footer := lipgloss.NewStyle().Foreground(theme.FGHelp).Render("esc to close")
	content := title + "\n" + m.levelTable.View() + "\n" + footer
	return dialogStyle().Render(content)
}

func truncateMiddle(s string, maxWidth int) string {
	w := ansi.StringWidth(s)
	if w <= maxWidth {
		return s
	}
	left := (maxWidth - 1) / 2
	right := maxWidth - 1 - left

	runes := []rune(s)
	var suffix string
	suffixW := 0
	for i := len(runes) - 1; i >= 0 && suffixW < right; i-- {
		suffixW++
		suffix = string(runes[i]) + suffix
	}

	return ansi.Truncate(s, left, "") + "…" + suffix
}
