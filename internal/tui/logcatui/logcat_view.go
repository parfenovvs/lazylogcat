package logcatui

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/parfenovvs/lazylogcat/internal/model"
	"github.com/parfenovvs/lazylogcat/internal/tui"
	"github.com/parfenovvs/lazylogcat/internal/tui/commandui"
	"github.com/parfenovvs/lazylogcat/internal/tui/theme"
	"github.com/parfenovvs/lazylogcat/internal/util"
)

const maxLogLines = 10000
const batchTimeout = 50 * time.Millisecond

var (
	titleStyle = func() lipgloss.Style {
		return theme.Panel().
			Padding(0, 1)
	}()

	helpTextNormal = "ctrl+p commands • G jump to recent • C clear • v visual"
	helpTextVisual = "j/↓ down • k/↑ up • shift+V select multiple • y copy • esc to normal"
)

// shortcutMap maps the second key of a ctrl+x shortcut to its CommandData.
// Built once from model.Commands() at package init.
var shortcutMap = buildShortcutMap()

func buildShortcutMap() map[string]model.CommandData {
	m := make(map[string]model.CommandData)
	for _, group := range model.Commands() {
		for _, cmd := range group.Commands {
			if cmd.Shortcut == "" {
				continue
			}
			// Shortcuts are in the format "ctrl+x <key>"
			parts := strings.SplitN(cmd.Shortcut, " ", 2)
			if len(parts) == 2 && parts[0] == "ctrl+x" {
				m[parts[1]] = cmd
			}
		}
	}
	return m
}

type LogcatViewModel struct {
	parentSize        model.Size
	viewport          viewport.Model
	device            *model.Device
	filter            model.Filter
	format            model.Format
	log               *util.RingBuffer
	pendingLogs       []string
	visualMode        bool
	currentLine       int
	startSelected     int
	softWrap          bool
	err               error
	awaitingShortcut  bool
	showCommandDialog bool
	commandDialog     commandui.CommandDialogModel
	deviceRequired    bool
}

type logcatMsg struct {
	Line string
}

type logcatEmptyMsg struct{}

type logcatErrorMsg struct {
	Err error
}

type logcatConnectedMsg struct{}

type batchTickMsg struct{}

// updateResult is returned by key handlers to indicate what action to take
type updateResult struct {
	cmd         tea.Cmd
	needsRender bool
}

func readNext(m LogcatViewModel) tea.Msg {
	line, err := util.ReadNextLogLine()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil // End of stream
		}
		slog.Warn("Error reading logcat line", "error", err)
		return nil
	}

	// Filter empty lines (allowed in long format)
	if !m.format.IsFormatValue("long") && strings.Trim(line, "\n\r ") == "" {
		return logcatEmptyMsg{}
	}

	// Filter by text search (case-insensitive)
	if m.filter.Text != "" && !strings.Contains(strings.ToLower(line), strings.ToLower(m.filter.Text)) {
		return logcatEmptyMsg{}
	}

	return logcatMsg{Line: line}
}

func tickForBatch() tea.Cmd {
	return tea.Tick(batchTimeout, func(t time.Time) tea.Msg {
		return batchTickMsg{}
	})
}

func New(parentSize model.Size, device *model.Device, filter model.Filter, format model.Format, softWrap bool) LogcatViewModel {
	m := LogcatViewModel{
		parentSize:     parentSize,
		device:         device,
		deviceRequired: device == nil,
		log:            util.NewRingBuffer(maxLogLines),
		softWrap:       softWrap,
		startSelected:  -1,
		filter:         filter,
		format:         format,
	}

	headerHeight := lipgloss.Height(m.headerView())
	footerHeight := lipgloss.Height(m.footerView())
	vp := viewport.New(parentSize.Width, parentSize.Height-footerHeight-headerHeight-1)
	m.viewport = vp

	return m
}

func (m LogcatViewModel) Update(msg tea.Msg) (LogcatViewModel, tea.Cmd) {
	var (
		cmd         tea.Cmd
		cmds        []tea.Cmd
		needsRender bool
		gotoBottom  bool
	)

	switch msg := msg.(type) {
	case model.Size:
		m.parentSize = msg
		return m, func() tea.Msg {
			return tui.MeasureCmd{}
		}

	case tui.MeasureCmd:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		m.viewport.Width = m.parentSize.Width
		m.viewport.Height = m.parentSize.Height - footerHeight - headerHeight - 1
		needsRender = true

	case commandui.CommandDialogCloseMsg:
		if m.deviceRequired {
			return m, nil // Cannot close dialog until a device is selected
		}
		m.showCommandDialog = false
		return m, nil

	case commandui.CommandDialogLevelSelectedMsg:
		m.showCommandDialog = false
		m.filter.Level = msg.Level
		return m, func() tea.Msg { return tui.ReconnectLogcatCmd{} }

	case commandui.CommandDialogFormatSelectedMsg:
		m.showCommandDialog = false
		m.format.SelectedFormat = msg.Format
		return m, func() tea.Msg { return tui.ReconnectLogcatCmd{} }

	case commandui.CommandDialogModifiersSelectedMsg:
		m.showCommandDialog = false
		m.format.ActiveModifiers = msg.Modifiers
		return m, func() tea.Msg { return tui.ReconnectLogcatCmd{} }

	case commandui.CommandDialogDeviceSelectedMsg:
		m.showCommandDialog = false
		m.deviceRequired = false
		return m, func() tea.Msg {
			return tui.DeviceSelectedMsg{Device: msg.Device, Filter: m.filter, Format: m.format, SoftWrap: m.softWrap}
		}

	case commandui.CommandDialogTextInputAppliedMsg:
		m.showCommandDialog = false
		switch msg.Command {
		case model.CommandPackage:
			m.filter.PackageName = msg.Value
			return m, func() tea.Msg { return tui.ReconnectLogcatCmd{} }
		case model.CommandTag:
			m.filter.Tag = msg.Value
			return m, func() tea.Msg { return tui.ReconnectLogcatCmd{} }
		case model.CommandContent:
			m.filter.Text = msg.Value
			return m, func() tea.Msg { return tui.ReconnectLogcatCmd{} }
		}
		return m, nil

	case commandui.CommandDialogSelectMsg:
		m.showCommandDialog = false
		switch msg.Command {
		case model.CommandToggleWrap:
			m.softWrap = !m.softWrap
			return m, func() tea.Msg { return tui.ReconnectLogcatCmd{} }
		case model.CommandReconnect:
			m.visualMode = false
			return m, func() tea.Msg { return tui.ReconnectLogcatCmd{} }
		case model.CommandExit:
			return m, func() tea.Msg { return tui.ExitCmd{} }
		}
		return m, nil

	case tea.KeyMsg:
		result := m.handleKeyMsg(msg)
		cmd = result.cmd
		needsRender = result.needsRender
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case tui.ShowDeviceDialogCmd:
		m.showCommandDialog = true
		m.commandDialog = commandui.NewDeviceDialog(m.device)
		return m, nil

	case tui.ReconnectLogcatCmd:
		if m.device == nil {
			return m, nil // No device selected, nothing to connect
		}
		util.CloseLogcat()
		m.log = util.NewRingBuffer(maxLogLines)
		m.pendingLogs = nil
		return m, tea.Batch(
			func() tea.Msg {
				err := util.ConnectLogcat(m.device.Id, m.filter, m.format)
				if err != nil {
					return logcatErrorMsg{Err: err}
				}
				return logcatConnectedMsg{}
			},
			func() tea.Msg {
				return tui.MeasureCmd{}
			},
		)

	case logcatConnectedMsg:
		return m, tea.Batch(
			func() tea.Msg { return readNext(m) },
			tickForBatch(),
		)

	case logcatMsg:
		if m.visualMode {
			return m, nil
		}
		m.pendingLogs = append(m.pendingLogs, msg.Line)
		return m, func() tea.Msg {
			return readNext(m)
		}

	case logcatEmptyMsg:
		if m.visualMode {
			return m, nil
		}
		return m, func() tea.Msg {
			return readNext(m)
		}

	case batchTickMsg:
		if !m.visualMode && len(m.pendingLogs) > 0 {
			wasAtBottom := m.viewport.AtBottom()
			for _, line := range m.pendingLogs {
				m.log.Append(line + "\n")
			}
			m.pendingLogs = nil
			needsRender = true
			if wasAtBottom {
				gotoBottom = true
			}
		}
		cmds = append(cmds, tickForBatch())

	case logcatErrorMsg:
		m.err = msg.Err
		util.CloseLogcat()
		return m, nil
	}

	if needsRender {
		m.Render()
	}
	if gotoBottom {
		m.viewport.GotoBottom()
	}

	// Viewport update for non-visual mode scrolling (skip internal tick messages)
	if !m.visualMode && !m.showCommandDialog && !m.awaitingShortcut {
		if _, isBatchTick := msg.(batchTickMsg); !isBatchTick {
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *LogcatViewModel) Render() {
	var b strings.Builder
	logs := m.log.All()
	for i, msg := range logs {
		if m.visualMode {
			selected := false
			if m.startSelected >= 0 {
				min := min(m.currentLine, m.startSelected)
				max := max(m.currentLine, m.startSelected)
				selected = i >= min && i <= max
			} else if i == m.currentLine {
				selected = true
			}
			line := strings.TrimSuffix(msg, "\n")
			if selected {
				styled := lipgloss.NewStyle().
					Bold(true).
					Background(theme.BGCursor).
					Foreground(theme.FGSelected).
					Width(m.viewport.Width).
					Render(line)
				b.WriteString(styled)
				b.WriteString("\n")
				continue
			}
		}
		if m.format.IsModifierActive("color") {
			line := strings.TrimSuffix(msg, "\n")
			style := lipgloss.NewStyle().
				Foreground(theme.GetLogColor(util.GetLogLevel(line, m.format)))
			if m.softWrap {
				style = style.Width(m.viewport.Width)
			}
			styled := style.Render(line)
			b.WriteString(styled)
			b.WriteString("\n")
		} else {
			b.WriteString(msg)
		}
	}
	wrapped := b.String()
	if m.softWrap {
		wrapped = lipgloss.NewStyle().Width(m.viewport.Width).Render(wrapped)
	}
	m.viewport.SetContent(wrapped)
}

// handleKeyMsg routes key messages to appropriate handlers based on mode
func (m *LogcatViewModel) handleKeyMsg(msg tea.KeyMsg) updateResult {
	key := msg.String()

	// When command dialog is open, delegate all keys to the dialog
	if m.showCommandDialog {
		var cmd tea.Cmd
		m.commandDialog, cmd = m.commandDialog.Update(msg)
		return updateResult{cmd: cmd}
	}

	// When awaiting the second key of a ctrl+x shortcut
	if m.awaitingShortcut {
		m.awaitingShortcut = false
		return m.handleShortcutKey(key)
	}

	// Try global keys first (work in both modes)
	if result, handled := m.handleGlobalKey(key); handled {
		return result
	}

	// Mode-specific handling
	if m.visualMode {
		return m.handleVisualModeKey(key)
	}
	return m.handleNormalModeKey(key)
}

// handleGlobalKey handles keys that work in both normal and visual modes
func (m *LogcatViewModel) handleGlobalKey(key string) (updateResult, bool) {
	switch key {
	case "G":
		m.viewport.GotoBottom()
		return updateResult{}, true

	case "v":
		m.visualMode = !m.visualMode
		if m.visualMode {
			m.viewport.GotoBottom()
			m.currentLine = m.log.Size() - 1
			return updateResult{needsRender: true}, true
		}
		return updateResult{
			cmd: func() tea.Msg { return tui.ReconnectLogcatCmd{} },
		}, true
	}

	return updateResult{}, false
}

// handleNormalModeKey handles keys specific to normal (non-visual) mode
func (m *LogcatViewModel) handleNormalModeKey(key string) updateResult {
	switch key {
	case "C":
		m.log.Clear()
		return updateResult{needsRender: true}

	case "ctrl+p":
		m.showCommandDialog = true
		deviceId := ""
		if m.device != nil {
			deviceId = m.device.Id
		}
		m.commandDialog = commandui.NewDialog(commandui.DialogConfig{
			Filter:         m.filter,
			Format:         m.format,
			SoftWrap:       m.softWrap,
			SelectedDevice: m.device,
			DeviceId:       deviceId,
		})
		return updateResult{needsRender: true}

	case "ctrl+x":
		m.awaitingShortcut = true
		return updateResult{}
	}

	return updateResult{}
}

// handleShortcutKey handles the second key of a ctrl+x shortcut sequence.
// It looks up the key in the shortcut map and either performs an action directly
// or opens the appropriate sub-dialog.
func (m *LogcatViewModel) handleShortcutKey(key string) updateResult {
	cmdData, ok := shortcutMap[key]
	if !ok {
		return updateResult{}
	}

	// Action commands execute immediately without opening a dialog
	if cmdData.Type == model.CommandTypeAction {
		switch cmdData.Command {
		case model.CommandToggleWrap:
			m.softWrap = !m.softWrap
			return updateResult{cmd: func() tea.Msg { return tui.ReconnectLogcatCmd{} }}
		case model.CommandReconnect:
			m.visualMode = false
			return updateResult{cmd: func() tea.Msg { return tui.ReconnectLogcatCmd{} }}
		case model.CommandExit:
			return updateResult{cmd: func() tea.Msg { return tui.ExitCmd{} }}
		}
		return updateResult{}
	}

	// Navigation commands open the sub-dialog directly
	cfg := commandui.DialogConfig{
		Filter:         m.filter,
		Format:         m.format,
		SoftWrap:       m.softWrap,
		SelectedDevice: m.device,
	}
	if m.device != nil {
		cfg.DeviceId = m.device.Id
	}

	var cmd tea.Cmd
	m.commandDialog, cmd = commandui.NewDialogForCommand(cfg, cmdData.Command)
	m.showCommandDialog = true
	return updateResult{cmd: cmd, needsRender: true}
}

// handleVisualModeKey handles keys specific to visual mode
func (m *LogcatViewModel) handleVisualModeKey(key string) updateResult {
	switch key {
	case "V":
		if m.startSelected >= 0 {
			m.startSelected = -1
		} else {
			m.startSelected = m.currentLine
		}
		return updateResult{needsRender: true}

	case "esc":
		if m.startSelected >= 0 {
			m.startSelected = -1
			return updateResult{needsRender: true}
		}
		m.visualMode = false
		m.startSelected = -1
		return updateResult{
			cmd: func() tea.Msg { return tui.ReconnectLogcatCmd{} },
		}

	case "y":
		if m.currentLine >= 0 && m.currentLine < m.log.Size() {
			var err error
			if m.startSelected >= 0 {
				start := min(m.currentLine, m.startSelected)
				end := max(m.currentLine, m.startSelected)
				var lines []string
				logs := m.log.Recent(m.log.Size() - start)
				for i := 0; i <= end-start; i++ {
					lines = append(lines, strings.TrimSpace(logs[i]))
				}
				err = util.CopyToClipboard(lines...)
				m.startSelected = -1
				return updateResult{needsRender: true}
			} else {
				logs := m.log.Recent(m.log.Size() - m.currentLine)
				lineText := strings.TrimSpace(logs[0])
				err = util.CopyToClipboard(lineText)
			}
			if err != nil {
				slog.Error("Failed to copy to clipboard", "error", err)
			}
		}
		return updateResult{}

	case "j", "down":
		if m.currentLine < m.log.Size()-1 {
			m.currentLine++
			m.ensureLineVisible()
			return updateResult{needsRender: true}
		}

	case "k", "up":
		if m.currentLine > 0 {
			m.currentLine--
			m.ensureLineVisible()
			return updateResult{needsRender: true}
		}
	}

	return updateResult{}
}

func (m *LogcatViewModel) ensureLineVisible() {
	if !m.visualMode || m.currentLine < 0 || m.currentLine >= m.log.Size() {
		return
	}

	min := min(m.currentLine, m.startSelected)
	max := max(m.currentLine, m.startSelected)

	logs := m.log.All()
	linesUpToCurrent := 0
	for i := 0; i <= m.currentLine; i++ {
		line := strings.TrimSuffix(logs[i], "\n")
		if m.softWrap || (m.startSelected != -1 && i >= min && i <= max) {
			linesUpToCurrent += lipgloss.Height(lipgloss.NewStyle().Width(m.viewport.Width).Render(line))
		} else {
			linesUpToCurrent++
		}
	}

	if linesUpToCurrent < m.viewport.YOffset+2 {
		m.viewport.HalfPageUp()
	} else if linesUpToCurrent > m.viewport.YOffset+m.viewport.Height-1 {
		m.viewport.HalfPageDown()
	}
}

func (m LogcatViewModel) renderBaseView() string {
	return fmt.Sprintf(
		"%s\n%s\n%s",
		m.headerView(),
		m.viewport.View(),
		m.footerView(),
	)
}

func (m LogcatViewModel) View() string {
	if m.err != nil {
		slog.Error("Logcat view error", "error", m.err)
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	baseView := m.renderBaseView()

	if m.awaitingShortcut {
		return tui.DimView(baseView)
	}

	if m.showCommandDialog {
		dimmedBaseView := tui.DimView(baseView)
		dialogContent := m.commandDialog.View()
		return tui.OverlayDialog(m.parentSize, dimmedBaseView, dialogContent)
	}

	return baseView
}

func (m LogcatViewModel) headerView() string {
	var filters []string
	if !m.filter.IsEmpty() {
		if m.filter.PackageName != "" {
			filters = append(filters, fmt.Sprintf("pkg:%s", m.filter.PackageName))
		}
		if m.filter.Level != "" && m.filter.Level != model.LvlV {
			filters = append(filters, fmt.Sprintf("level:%s", m.filter.Level))
		}
		if m.filter.Tag != "" {
			filters = append(filters, fmt.Sprintf("tag:%s", m.filter.Tag))
		}
		if m.filter.Text != "" {
			filters = append(filters, fmt.Sprintf("text:%s", m.filter.Text))
		}
	}

	format := m.format.Value()
	mods := m.format.Modifiers()
	var modsStr string
	if len(mods) > 0 {
		modsStr = fmt.Sprintf(" | %s", strings.Join(mods, ","))
	}

	filtersStr := ""
	if len(filters) > 0 {
		filtersStr = fmt.Sprintf("\n%s", strings.Join(filters, " | "))
	}

	var name string
	if m.device != nil {
		name = m.device.Name
	} else {
		name = "No device"
	}
	deviceName := lipgloss.NewStyle().Bold(true).Render(name)
	headerText := titleStyle.Render(fmt.Sprintf("%s | %s%s%s", deviceName, format, modsStr, filtersStr))
	width := lipgloss.Width(headerText)

	if width > m.viewport.Width {
		headerText = titleStyle.Render(fmt.Sprintf("%s | ...", deviceName))
	}

	return headerText
}

func (m LogcatViewModel) footerView() string {
	var helpText string
	if m.visualMode {
		helpText = helpTextVisual
	} else {
		helpText = helpTextNormal
	}

	help := lipgloss.NewStyle().
		Foreground(theme.FGHelp).
		Width(m.viewport.Width).
		AlignHorizontal(lipgloss.Center).
		Padding(0, 2).
		Render(helpText)

	return help
}

func Close(m *LogcatViewModel) {
	util.CloseLogcat()
	m.log = util.NewRingBuffer(maxLogLines)
}
