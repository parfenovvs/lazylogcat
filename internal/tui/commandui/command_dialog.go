package commandui

import (
	"maps"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/parfenovvs/lazylogcat/internal/model"
	"github.com/parfenovvs/lazylogcat/internal/tui"
	"github.com/parfenovvs/lazylogcat/internal/tui/theme"
)

// DialogConfig holds the parameters for creating a new command dialog.
type DialogConfig struct {
	Filter         model.Filter
	OutputPrefs    model.OutputPrefs
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
	stateDevices
	stateTextInput
)

var dialogStyle = func() lipgloss.Style {
	return theme.Dialog().
		Width(tui.DialogWidth).
		MaxHeight(tui.DialogMaxHeight)
}

type CommandDialogModel struct {
	state      dialogState
	table      table.Model
	skipRows   map[int]bool
	commandMap map[int]model.CommandData

	// Unfiltered originals for the main command table
	allCommandRows []table.Row
	allSkipRows    map[int]bool
	allCommandMap  map[int]model.CommandData

	// Active subdialog widgets (only one used at a time)
	singleSelect SingleSelectModel
	textInputDlg TextInputModel

	// Context for the active subdialog
	activeCommand model.Command
	currentLevel  model.Level

	// Device-specific state (devices has error/empty states beyond SingleSelect)
	allDevices     []model.Device
	selectedDevice *model.Device
	deviceErr      error

	searchInput textinput.Model

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
		case model.CommandToggleWrap:
			if cfg.OutputPrefs.SoftWrap {
				return "on"
			}
			return "off"
		case model.CommandToggleColor:
			if cfg.OutputPrefs.Color {
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
		{Title: "", Width: tui.DialogWidth - 32},
		{Title: "", Width: 8},
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

	t := newTable(columns, rows, len(rows)+1)
	t.SetCursor(1) // Skip the first group header

	currentLevel := cfg.Filter.Level
	if currentLevel == "" {
		currentLevel = model.LvlV
	}

	// Store unfiltered originals for the main command table
	allSkipRows := make(map[int]bool)
	maps.Copy(allSkipRows, skipRows)
	allCommandMap := make(map[int]model.CommandData)
	maps.Copy(allCommandMap, commandMap)
	allCommandRows := make([]table.Row, len(rows))
	copy(allCommandRows, rows)

	return CommandDialogModel{
		state:          stateCommands,
		table:          t,
		skipRows:       skipRows,
		commandMap:     commandMap,
		allCommandRows: allCommandRows,
		allSkipRows:    allSkipRows,
		allCommandMap:  allCommandMap,
		currentLevel:   currentLevel,
		selectedDevice: cfg.SelectedDevice,
		searchInput:    newSearchInput(),
		filter:         cfg.Filter,
		deviceId:       cfg.DeviceId,
	}
}

// NewDialogForCommand creates a command dialog that opens directly in the sub-dialog
// for the given command, skipping the main command list. Returns the model and an optional
// tea.Cmd (needed for text input cursor blink).
func NewDialogForCommand(cfg DialogConfig, cmd model.Command) (CommandDialogModel, tea.Cmd) {
	m := NewDialog(cfg)
	return m.openSubdialog(cmd)
}

// NewDeviceDialog creates a command dialog that opens directly in the device selection state.
func NewDeviceDialog(selectedDevice *model.Device) CommandDialogModel {
	ss, allDevices, err := loadDevices(selectedDevice)
	return CommandDialogModel{
		state:          stateDevices,
		singleSelect:   ss,
		allDevices:     allDevices,
		deviceErr:      err,
		selectedDevice: selectedDevice,
		searchInput:    newSearchInput(),
	}
}

// openSubdialog transitions the dialog into the appropriate sub-dialog for the given command.
// Returns the updated model and an optional tea.Cmd.
func (m CommandDialogModel) openSubdialog(cmd model.Command) (CommandDialogModel, tea.Cmd) {
	resetSearchInput(&m.searchInput)
	m.activeCommand = cmd

	switch cmd {
	case model.CommandLevel:
		m.singleSelect = newLevelSingleSelect(m.currentLevel)
		m.state = stateLogLevel
	case model.CommandDevices:
		ss, allDevices, err := loadDevices(m.selectedDevice)
		m.singleSelect = ss
		m.allDevices = allDevices
		m.deviceErr = err
		m.state = stateDevices
	case model.CommandPackage:
		m.textInputDlg = newCommandTextInput(cmd, m.filter.PackageName, m.deviceId)
		m.state = stateTextInput
		return m, initTextInputCmd()
	case model.CommandTag:
		m.textInputDlg = newCommandTextInput(cmd, m.filter.Tag, m.deviceId)
		m.state = stateTextInput
		return m, initTextInputCmd()
	case model.CommandContent:
		m.textInputDlg = newCommandTextInput(cmd, m.filter.Text, m.deviceId)
		m.state = stateTextInput
		return m, initTextInputCmd()
	}

	return m, nil
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
		case stateDevices:
			return m.updateDevices(msg, key)
		case stateTextInput:
			return m.updateTextInput(msg, key)
		}
	default:
		// Forward non-key messages (e.g. cursor blink) to the active widget.
		if m.state == stateTextInput {
			var cmd tea.Cmd
			m.textInputDlg, cmd = m.textInputDlg.Update(msg)
			return m, cmd
		}
		// Forward non-key messages to the appropriate widget for cursor blink.
		switch m.state {
		case stateLogLevel, stateDevices:
			var cmd tea.Cmd
			m.singleSelect, cmd = m.singleSelect.UpdateBlink(msg)
			return m, cmd
		default:
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m CommandDialogModel) updateCommands(msg tea.KeyMsg, key string) (CommandDialogModel, tea.Cmd) {
	if key == "enter" {
		if cmdData, ok := m.commandMap[m.table.Cursor()]; ok {
			if cmdData.Type == model.CommandTypeNavigation {
				return m.openSubdialog(cmdData.Command)
			}
			return m, func() tea.Msg { return CommandDialogSelectMsg{Command: cmdData.Command} }
		}
		return m, nil
	}

	// Arrow keys go to table navigation
	if key == "up" || key == "down" {
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

	// All other keys go to the search input
	prevValue := m.searchInput.Value()
	m.searchInput, _ = m.searchInput.Update(msg)
	if m.searchInput.Value() != prevValue {
		m.filterCommandRows()
	}

	return m, nil
}

// filterCommandRows rebuilds the table rows from the unfiltered originals,
// keeping only rows whose command name matches the current search query.
func (m *CommandDialogModel) filterCommandRows() {
	query := strings.ToLower(strings.TrimSpace(m.searchInput.Value()))

	if query == "" {
		// No filter: restore all original rows
		rows := make([]table.Row, len(m.allCommandRows))
		copy(rows, m.allCommandRows)
		m.skipRows = make(map[int]bool)
		maps.Copy(m.skipRows, m.allSkipRows)
		m.commandMap = make(map[int]model.CommandData)
		maps.Copy(m.commandMap, m.allCommandMap)
		m.table.SetRows(rows)
		m.table.SetHeight(len(rows) + 1)
		// Set cursor to first non-skip row
		for i := 0; i < len(rows); i++ {
			if !m.skipRows[i] {
				m.table.SetCursor(i)
				break
			}
		}
		return
	}

	// Build filtered rows: include group headers only if at least one command in the group matches
	var rows []table.Row
	skipRows := make(map[int]bool)
	commandMap := make(map[int]model.CommandData)

	// Walk through the original rows, tracking group boundaries
	i := 0
	for i < len(m.allCommandRows) {
		// Check if this is a group header (skip row that is not a blank separator)
		if m.allSkipRows[i] {
			// Could be a blank separator or a group header
			// Blank separator: all columns empty
			row := m.allCommandRows[i]
			isSeparator := row[0] == "" && row[1] == "" && row[2] == ""

			if isSeparator {
				// Skip blank separator, advance to group header
				i++
				continue
			}

			// This is a group header. Collect commands in this group.
			headerIdx := i
			i++
			var groupCommands []struct {
				origIdx int
				row     table.Row
				cmdData model.CommandData
			}
			for i < len(m.allCommandRows) && !m.allSkipRows[i] {
				if cmdData, ok := m.allCommandMap[i]; ok {
					if strings.Contains(strings.ToLower(cmdData.Name), query) {
						groupCommands = append(groupCommands, struct {
							origIdx int
							row     table.Row
							cmdData model.CommandData
						}{i, m.allCommandRows[i], cmdData})
					}
				}
				i++
			}

			if len(groupCommands) > 0 {
				// Add separator before group (if not the first group in filtered results)
				if len(rows) > 0 {
					skipRows[len(rows)] = true
					rows = append(rows, table.Row{"", "", ""})
				}
				// Add group header
				skipRows[len(rows)] = true
				rows = append(rows, m.allCommandRows[headerIdx])
				// Add matching commands
				for _, gc := range groupCommands {
					commandMap[len(rows)] = gc.cmdData
					rows = append(rows, gc.row)
				}
			}
			continue
		}
		i++
	}

	m.skipRows = skipRows
	m.commandMap = commandMap
	m.table.SetRows(rows)
	m.table.SetHeight(len(rows) + 1)

	// Set cursor to first non-skip row
	cursorSet := false
	for idx := 0; idx < len(rows); idx++ {
		if !skipRows[idx] {
			m.table.SetCursor(idx)
			cursorSet = true
			break
		}
	}
	if !cursorSet {
		m.table.SetCursor(0)
	}
}

func (m CommandDialogModel) View() string {
	switch m.state {
	case stateLogLevel:
		return m.singleSelect.View()
	case stateDevices:
		return m.viewDevices()
	case stateTextInput:
		return m.textInputDlg.View()
	default:
		return m.viewCommands()
	}
}

func (m CommandDialogModel) viewCommands() string {
	title := theme.DialogTitle().Render("Commands")
	footer := theme.DialogHelp().Render("esc to close")

	var body string
	if len(m.table.Rows()) == 0 && m.searchInput.Value() != "" {
		body = theme.DialogHelp().Render("\nNo results found")
	} else {
		body = m.table.View()
	}

	search := theme.DialogSearch().Render(m.searchInput.View())
	content := title + "\n\n" + search + "\n" + body + "\n\n" + footer
	return dialogStyle().Render(content)
}
