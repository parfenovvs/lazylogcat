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

// CommandDialogCloseMsg is emitted when the dialog should close.
type CommandDialogCloseMsg struct{}

var dialogStyle = func() lipgloss.Style {
	return theme.ActivePanel().
		Padding(1, 2)
}

// CommandDialogModel is a Bubble Tea component for the command palette dialog.
type CommandDialogModel struct {
	table    table.Model
	skipRows map[int]bool
}

// NewDialog creates a command dialog that resolves display values from domain inputs.
func NewDialog(filter model.Filter, format model.Format, softWrap bool) CommandDialogModel {
	resolveValue := func(cmd Command) string {
		switch cmd {
		case CommandPackage:
			return filter.PackageName
		case CommandTag:
			return filter.Tag
		case CommandLevel:
			lvl := string(filter.Level)
			if lvl == "" {
				lvl = "V"
			}
			return lvl
		case CommandContent:
			return filter.Text
		case CommandFormat:
			return format.Value()
		case CommandModifiers:
			mods := format.Modifiers()
			switch len(mods) {
			case 0:
				return ""
			case 1:
				return mods[0]
			default:
				return fmt.Sprintf("[%d] mods", len(mods))
			}
		case CommandToggleWrap:
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
	var rows []table.Row
	for i, group := range Commands() {
		if i > 0 {
			skipRows[len(rows)] = true
			rows = append(rows, table.Row{"", "", ""})
		}
		skipRows[len(rows)] = true
		groupName := lipgloss.NewStyle().Bold(true).Render(group.Name)
		rows = append(rows, table.Row{groupName, "", ""})
		for _, cmd := range group.Commands {
			value := truncateMiddle(resolveValue(cmd.Command), 10)
			rows = append(rows, table.Row{cmd.Name, value, cmd.Shortcut})
		}
	}

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
		table.WithHeight(len(rows)),
		table.WithFocused(true),
		table.WithKeyMap(km),
	)
	t.SetStyles(s)
	t.SetCursor(1) // Skip the first group header

	return CommandDialogModel{
		table:    t,
		skipRows: skipRows,
	}
}

// Update handles key messages while the dialog is active.
func (m CommandDialogModel) Update(msg tea.Msg) (CommandDialogModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		if key == "ctrl+p" || key == "esc" {
			return m, func() tea.Msg { return CommandDialogCloseMsg{} }
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
	}

	return m, nil
}

// View renders the styled dialog box.
func (m CommandDialogModel) View() string {
	title := lipgloss.NewStyle().Bold(true).Render("Command List")
	footer := lipgloss.NewStyle().Foreground(theme.FGHelp).Render("esc to close")
	content := title + "\n" + m.table.View() + "\n" + footer
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
