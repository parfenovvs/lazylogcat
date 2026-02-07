package commandui

import (
	"fmt"
	"maps"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/parfenovvs/lazylogcat/internal/model"
	"github.com/parfenovvs/lazylogcat/internal/tui/theme"
)

// DialogConfig holds the parameters for creating a new command dialog.
type DialogConfig struct {
	Filter         model.Filter
	Format         model.Format
	SoftWrap       bool
	SelectedDevice *model.Device
	DeviceId       string
}

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
	stateTextInput
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

	textInput        textinput.Model
	textInputCommand model.Command
	textInputTitle   string
	textInputError   string

	filter   model.Filter
	deviceId string
}

func NewDialog(cfg DialogConfig) CommandDialogModel {
	resolveValue := func(cmd model.Command) string {
		switch cmd {
		case model.CommandPackage:
			return cfg.Filter.PackageName
		case model.CommandTag:
			return cfg.Filter.Tag
		case model.CommandLevel:
			lvl := string(cfg.Filter.Level)
			if lvl == "" {
				lvl = string(model.LvlV)
			}
			return lvl
		case model.CommandContent:
			return cfg.Filter.Text
		case model.CommandFormat:
			return cfg.Format.Value()
		case model.CommandModifiers:
			mods := cfg.Format.Modifiers()
			switch len(mods) {
			case 0:
				return ""
			case 1:
				return mods[0]
			default:
				return fmt.Sprintf("[%d]", len(mods))
			}
		case model.CommandToggleWrap:
			if cfg.SoftWrap {
				return "on"
			}
			return "off"
		case model.CommandDevices:
			if cfg.SelectedDevice != nil {
				return cfg.SelectedDevice.Name
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

	currentLevel := cfg.Filter.Level
	if currentLevel == "" {
		currentLevel = model.LvlV
	}
	levelTable, levelMap := newLevelTable(currentLevel)
	formatTable, formatMap := newFormatTable(cfg.Format.Value())
	modifiersTable, modifierMap := newModifiersTable(cfg.Format.ActiveModifiers)

	return CommandDialogModel{
		state:           stateCommands,
		table:           t,
		skipRows:        skipRows,
		commandMap:      commandMap,
		levelTable:      levelTable,
		levelMap:        levelMap,
		formatTable:     formatTable,
		formatMap:       formatMap,
		selectedDevice:  cfg.SelectedDevice,
		modifiersTable:  modifiersTable,
		modifierMap:     modifierMap,
		activeModifiers: cfg.Format.ActiveModifiers,
		filter:          cfg.Filter,
		deviceId:        cfg.DeviceId,
	}
}

// NewDialogForCommand creates a command dialog that opens directly in the sub-dialog
// for the given command, skipping the main command list. Returns the model and an optional
// tea.Cmd (needed for text input cursor blink).
func NewDialogForCommand(cfg DialogConfig, cmd model.Command) (CommandDialogModel, tea.Cmd) {
	m := NewDialog(cfg)

	switch cmd {
	case model.CommandLevel:
		m.state = stateLogLevel
	case model.CommandFormat:
		m.state = stateFormat
	case model.CommandModifiers:
		m.tempModifiers = make(map[string]bool)
		maps.Copy(m.tempModifiers, m.activeModifiers)
		m.modifiersTable = m.refreshModifierRows()
		m.state = stateModifiers
	case model.CommandDevices:
		deviceTable, deviceMap, err := loadDevices(cfg.SelectedDevice)
		m.deviceTable = deviceTable
		m.deviceMap = deviceMap
		m.deviceErr = err
		m.state = stateDevices
	case model.CommandPackage:
		m.textInputCommand = cmd
		m.textInputTitle = textInputTitle(cmd)
		m.textInput = newDialogTextInput(textInputPlaceholder(cmd), cfg.Filter.PackageName)
		m.state = stateTextInput
		return m, textinput.Blink
	case model.CommandTag:
		m.textInputCommand = cmd
		m.textInputTitle = textInputTitle(cmd)
		m.textInput = newDialogTextInput(textInputPlaceholder(cmd), cfg.Filter.Tag)
		m.state = stateTextInput
		return m, textinput.Blink
	case model.CommandContent:
		m.textInputCommand = cmd
		m.textInputTitle = textInputTitle(cmd)
		m.textInput = newDialogTextInput(textInputPlaceholder(cmd), cfg.Filter.Text)
		m.state = stateTextInput
		return m, textinput.Blink
	}

	return m, nil
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
		case stateTextInput:
			return m.updateTextInput(msg, key)
		}
	default:
		// Forward non-key messages (e.g. cursor blink) to the text input when active.
		if m.state == stateTextInput {
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
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
			if cmdData.Type == model.CommandTypeNavigation && cmdData.Command == model.CommandPackage {
				m.textInputCommand = cmdData.Command
				m.textInputTitle = textInputTitle(cmdData.Command)
				m.textInput = newDialogTextInput(textInputPlaceholder(cmdData.Command), m.filter.PackageName)
				m.state = stateTextInput
				return m, textinput.Blink
			}
			if cmdData.Type == model.CommandTypeNavigation && cmdData.Command == model.CommandTag {
				m.textInputCommand = cmdData.Command
				m.textInputTitle = textInputTitle(cmdData.Command)
				m.textInput = newDialogTextInput(textInputPlaceholder(cmdData.Command), m.filter.Tag)
				m.state = stateTextInput
				return m, textinput.Blink
			}
			if cmdData.Type == model.CommandTypeNavigation && cmdData.Command == model.CommandContent {
				m.textInputCommand = cmdData.Command
				m.textInputTitle = textInputTitle(cmdData.Command)
				m.textInput = newDialogTextInput(textInputPlaceholder(cmdData.Command), m.filter.Text)
				m.state = stateTextInput
				return m, textinput.Blink
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
	case stateTextInput:
		return m.viewTextInput()
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
