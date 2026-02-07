package commandui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/parfenovvs/lazylogcat/internal/model"
	"github.com/parfenovvs/lazylogcat/internal/tui/theme"
)

type CommandDialogCloseMsg struct{}

type CommandDialogSelectMsg struct {
	Command model.Command
}

type dialogState int

const (
	stateCommands dialogState = iota
	stateLogLevel
	stateFormat
	stateDevices
	stateModifiers
)

var dialogStyle = func() lipgloss.Style {
	return theme.ActivePanel().
		Padding(1, 2)
}

type CommandDialogModel struct {
	state      dialogState
	table      table.Model
	skipRows   map[int]bool
	commandMap map[int]model.CommandData

	levelTable table.Model
	levelMap   map[int]model.Level

	formatTable table.Model
	formatMap   map[int]string

	deviceTable    table.Model
	deviceMap      map[int]model.Device
	selectedDevice *model.Device
	deviceErr      error

	modifiersTable  table.Model
	modifierMap     map[int]string
	tempModifiers   map[string]bool
	activeModifiers map[string]bool
}

func NewDialog(filter model.Filter, format model.Format, softWrap bool, selectedDevice *model.Device) CommandDialogModel {
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
				return fmt.Sprintf("[%d]", len(mods))
			}
		case model.CommandToggleWrap:
			if softWrap {
				return "on"
			}
			return "off"
		case model.CommandDevices:
			if selectedDevice != nil {
				return selectedDevice.Name
			}
			return ""
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
	formatTable, formatMap := newFormatTable(format.Value())
	modifiersTable, modifierMap := newModifiersTable(format.ActiveModifiers)

	return CommandDialogModel{
		state:           stateCommands,
		table:           t,
		skipRows:        skipRows,
		commandMap:      commandMap,
		levelTable:      levelTable,
		levelMap:        levelMap,
		formatTable:     formatTable,
		formatMap:       formatMap,
		selectedDevice:  selectedDevice,
		modifiersTable:  modifiersTable,
		modifierMap:     modifierMap,
		activeModifiers: format.ActiveModifiers,
	}
}

// NewDeviceDialog creates a command dialog that opens directly in the device selection state.
func NewDeviceDialog(selectedDevice *model.Device) CommandDialogModel {
	deviceTable, deviceMap, err := loadDevices(selectedDevice)
	return CommandDialogModel{
		state:          stateDevices,
		deviceTable:    deviceTable,
		deviceMap:      deviceMap,
		deviceErr:      err,
		selectedDevice: selectedDevice,
	}
}

func (m CommandDialogModel) Update(msg tea.Msg) (CommandDialogModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		if key == "ctrl+p" || key == "esc" {
			if m.state == stateModifiers {
				return m, func() tea.Msg {
					return CommandDialogModifiersSelectedMsg{Modifiers: m.tempModifiers}
				}
			}
			return m, func() tea.Msg { return CommandDialogCloseMsg{} }
		}

		switch m.state {
		case stateCommands:
			return m.updateCommands(msg, key)
		case stateLogLevel:
			return m.updateLogLevel(msg, key)
		case stateFormat:
			return m.updateFormat(msg, key)
		case stateDevices:
			return m.updateDevices(msg, key)
		case stateModifiers:
			return m.updateModifiers(msg, key)
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
			if cmdData.Type == model.CommandTypeNavigation && cmdData.Command == model.CommandFormat {
				m.state = stateFormat
				return m, nil
			}
			if cmdData.Type == model.CommandTypeNavigation && cmdData.Command == model.CommandModifiers {
				m.tempModifiers = make(map[string]bool)
				for k, v := range m.activeModifiers {
					m.tempModifiers[k] = v
				}
				m.modifiersTable = m.refreshModifierRows()
				m.state = stateModifiers
				return m, nil
			}
			if cmdData.Type == model.CommandTypeNavigation && cmdData.Command == model.CommandDevices {
				deviceTable, deviceMap, err := loadDevices(m.selectedDevice)
				m.deviceTable = deviceTable
				m.deviceMap = deviceMap
				m.deviceErr = err
				m.state = stateDevices
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

func (m CommandDialogModel) View() string {
	switch m.state {
	case stateLogLevel:
		return m.viewLogLevel()
	case stateFormat:
		return m.viewFormat()
	case stateDevices:
		return m.viewDevices()
	case stateModifiers:
		return m.viewModifiers()
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
