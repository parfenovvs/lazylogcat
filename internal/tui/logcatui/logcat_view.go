package logcatui

import (
	"fmt"
	"image/color"
	"log/slog"
	"os"
	"strings"
	"time"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/parfenovvs/lazylogcat/internal/model"
	"github.com/parfenovvs/lazylogcat/internal/tui"
	"github.com/parfenovvs/lazylogcat/internal/tui/commandui"
	"github.com/parfenovvs/lazylogcat/internal/tui/theme"
	"github.com/parfenovvs/lazylogcat/internal/util"
)

const maxLogLines = 10000
const batchTimeout = 16 * time.Millisecond

var (
	titleStyle = func() lipgloss.Style {
		return theme.Panel().
			Padding(0, 1)
	}()

	helpTextNormal = "Commands: ctrl+p • Jump to recent: shift+g • Clear: shift+c • Visual: v"
	helpTextVisual = "Multiline: shift+v • Copy: y • Open in editor: ctrl+e • Exit visual: ESC"
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
	parentSize          model.Size
	viewport            viewport.Model
	device              *model.Device
	deviceId            string
	filter              model.Filter
	outputPrefs         model.OutputPrefs
	log                 *util.RingBuffer
	reader              *util.LogcatReader
	visualMode          bool
	currentLine         int
	startSelected       int
	visualLineToLogLine []int
	logLineVisualStart  []int
	logLineColor        []color.Color
	awaitingShortcut    bool
	connGen             uint64 // incremented on each voluntary reconnect; used to discard stale messages
	toast               tui.ToastModel
	showCommandDialog   bool
	commandDialog       commandui.CommandDialogModel
	deviceRequired      bool
	recordingStartTime  *time.Time
}

type logcatErrorMsg struct {
	Err error
}

type logcatConnectedMsg struct{}

type logcatDisconnectedMsg struct{ gen uint64 }

type batchTickMsg struct{}

type pidRefreshTickMsg struct{}

type reconnectTickMsg struct{ gen uint64 }

type pidRefreshMsg struct {
	pidSet map[string]struct{}
}

// updateResult is returned by key handlers to indicate what action to take
type updateResult struct {
	cmd         tea.Cmd
	needsRender bool
	consumed    bool // when true, the key is fully handled and must not be forwarded to the viewport
}

// watchReaderDone returns a cmd that blocks until the reader goroutine exits,
// then delivers a logcatDisconnectedMsg tagged with the connection generation
// so the handler can discard stale notifications from previous connections.
func watchReaderDone(reader *util.LogcatReader, gen uint64) tea.Cmd {
	return func() tea.Msg {
		reader.WaitForDone()
		return logcatDisconnectedMsg{gen: gen}
	}
}

func tickForBatch() tea.Cmd {
	return tea.Tick(batchTimeout, func(t time.Time) tea.Msg {
		return batchTickMsg{}
	})
}

const pidRefreshInterval = 1 * time.Second
const reconnectInterval = 2 * time.Second

func reconnectTick(gen uint64) tea.Cmd {
	return tea.Tick(reconnectInterval, func(t time.Time) tea.Msg {
		return reconnectTickMsg{gen: gen}
	})
}

func pidRefreshTick() tea.Cmd {
	return tea.Tick(pidRefreshInterval, func(t time.Time) tea.Msg {
		return pidRefreshTickMsg{}
	})
}

func refreshPIDs(deviceId string, filter *model.TextFilter) tea.Cmd {
	return func() tea.Msg {
		processes, err := util.GetProcessList(deviceId)
		if err != nil {
			slog.Warn("Failed to get process list", "error", err)
			return pidRefreshMsg{}
		}
		return pidRefreshMsg{pidSet: util.ResolvePIDs(processes, filter)}
	}
}

func New(parentSize model.Size, device *model.Device, deviceId string, filter model.Filter, outputPrefs model.OutputPrefs) LogcatViewModel {
	m := LogcatViewModel{
		parentSize:     parentSize,
		device:         device,
		deviceId:       deviceId,
		deviceRequired: device == nil,
		log:            util.NewRingBuffer(maxLogLines),
		reader:         util.NewLogcatReader(),
		startSelected:  -1,
		filter:         filter,
		outputPrefs:    outputPrefs,
	}

	headerHeight := lipgloss.Height(m.headerView())
	footerHeight := lipgloss.Height(m.footerView())
	vp := viewport.New(
		viewport.WithWidth(parentSize.Width),
		viewport.WithHeight(parentSize.Height-footerHeight-headerHeight),
	)
	m.viewport = vp

	return m
}

func (m LogcatViewModel) Update(msg tea.Msg) (LogcatViewModel, tea.Cmd) {
	var (
		cmd           tea.Cmd
		cmds          []tea.Cmd
		needsRender   bool
		gotoBottom    bool
		keyConsumed   bool
		yOffsetAdjust int // visual lines evicted from the top; used to stabilise scroll position
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
		m.viewport.SetWidth(m.parentSize.Width)
		m.viewport.SetHeight(m.parentSize.Height - footerHeight - headerHeight - 1)
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

	case commandui.CommandDialogDeviceSelectedMsg:
		m.showCommandDialog = false
		m.deviceRequired = false
		return m, func() tea.Msg {
			return tui.DeviceSelectedMsg{Device: msg.Device, Filter: m.filter, OutputPrefs: m.outputPrefs}
		}

	case commandui.CommandDialogTextInputAppliedMsg:
		m.showCommandDialog = false
		msg.Filter.Compile()
		switch msg.Command {
		case model.CommandPackage:
			m.filter.PackageName = msg.Filter
			if msg.Filter.IsEmpty() {
				m.reader.UpdatePIDSet(nil)
			}
			return m, func() tea.Msg { return tui.ReconnectLogcatCmd{} }
		case model.CommandTag:
			m.filter.Tag = msg.Filter
			return m, func() tea.Msg { return tui.ReconnectLogcatCmd{} }
		case model.CommandContent:
			m.filter.Text = msg.Filter
			return m, func() tea.Msg { return tui.ReconnectLogcatCmd{} }
		}
		return m, nil

	case commandui.CommandDialogSelectMsg:
		m.showCommandDialog = false
		switch msg.Command {
		case model.CommandReconnect:
			m.exitVisualMode()
			return m, func() tea.Msg { return tui.ReconnectLogcatCmd{} }
		case model.CommandExportBuffer:
			return m, m.exportBuffer()
		case model.CommandStartRecording:
			return m, m.startRecording()
		case model.CommandStopRecording:
			return m, m.stopRecording()
		case model.CommandExit:
			return m, func() tea.Msg { return tui.ExitCmd{} }
		}
		return m, nil

	case commandui.OutputPrefsChangedMsg:
		m.showCommandDialog = false
		oldWrap := m.outputPrefs.SoftWrap
		m.outputPrefs = msg.OutputPrefs
		if m.outputPrefs.SoftWrap != oldWrap {
			return m, func() tea.Msg { return tui.ReconnectLogcatCmd{} }
		}
		m.Render()
		return m, nil

	case tui.ToastExpiredMsg:
		m.toast.Update(msg)

	case tui.EditorFinishedMsg:
		if msg.Err != nil {
			slog.Warn("Editor exited with error", "error", msg.Err)
			toastCmd := m.toast.Show("Editor error: "+msg.Err.Error(), tui.ToastError)
			m.Render()
			return m, toastCmd
		}
		m.Render()
		return m, nil

	case tea.KeyPressMsg:
		result := m.handleKeyMsg(msg)
		cmd = result.cmd
		needsRender = result.needsRender
		keyConsumed = result.consumed
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
		m.connGen++ // invalidate any in-flight messages from the previous connection
		m.reader.Disconnect()
		m.log = util.NewRingBuffer(maxLogLines)
		return m, tea.Batch(
			func() tea.Msg {
				err := m.reader.Connect(m.device.Id, m.filter)
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
		m.log = util.NewRingBuffer(maxLogLines)
		cmds := []tea.Cmd{
			watchReaderDone(m.reader, m.connGen),
			tickForBatch(),
			func() tea.Msg { return tui.MeasureCmd{} },
		}
		if !m.filter.PackageName.IsEmpty() {
			// Start PID refresh cycle and do an immediate resolution
			cmds = append(cmds, pidRefreshTick(), refreshPIDs(m.deviceId, &m.filter.PackageName))
		}
		return m, tea.Batch(cmds...)

	case pidRefreshTickMsg:
		if !m.filter.PackageName.IsEmpty() {
			return m, tea.Batch(
				refreshPIDs(m.deviceId, &m.filter.PackageName),
				pidRefreshTick(),
			)
		}
		return m, nil

	case logcatDisconnectedMsg:
		// Discard stale notifications from a previous connection generation.
		if msg.gen != m.connGen {
			return m, nil
		}
		// The reader goroutine has exited (EOF or adb crash).
		if !m.reader.IsConnected() && m.device != nil {
			toastCmd := m.toast.Show("Reconnecting...", tui.ToastInfo)
			return m, tea.Batch(toastCmd, reconnectTick(m.connGen))
		}
		return m, nil

	case reconnectTickMsg:
		// Discard stale ticks from a previous connection generation.
		if msg.gen != m.connGen {
			return m, nil
		}
		if m.device == nil || m.reader.IsConnected() {
			return m, nil
		}
		gen := m.connGen // capture for the closure
		return m, func() tea.Msg {
			err := m.reader.Connect(m.device.Id, m.filter)
			if err != nil {
				slog.Warn("Reconnect attempt failed", "error", err)
				return logcatDisconnectedMsg{gen: gen}
			}
			return logcatConnectedMsg{}
		}

	case pidRefreshMsg:
		m.reader.UpdatePIDSet(msg.pidSet)
		return m, nil

	case batchTickMsg:
		if !m.visualMode {
			if lines := m.reader.Drain(); len(lines) > 0 {
				wasAtBottom := m.viewport.AtBottom()

				// When the user has scrolled up and the ring buffer is at
				// capacity, each Append evicts the oldest line from the top
				// of the content. Count the visual lines that will be lost
				// so we can adjust YOffset after Render to keep the
				// viewport pinned to the same content.
				if !wasAtBottom && m.log.Size() >= maxLogLines {
					evictCount := min(len(lines), maxLogLines)
					oldest := m.log.All()
					cols := m.outputPrefs.Columns
					for i := 0; i < evictCount && i < len(oldest); i++ {
						if m.outputPrefs.SoftWrap {
							rendered := softWrapIndent(
								oldest[i].ModifiedString(cols),
								oldest[i].PrefixWidth(cols),
								m.viewport.Width(),
							)
							yOffsetAdjust += lipgloss.Height(rendered)
						} else {
							yOffsetAdjust++
						}
					}
				}

				for _, line := range lines {
					m.log.Append(line)
				}
				needsRender = true
				if wasAtBottom {
					gotoBottom = true
				}
			}
		}
		cmds = append(cmds, tickForBatch())

	case logcatErrorMsg:
		slog.Warn("Logcat connect error", "error", msg.Err)
		m.reader.Disconnect()
		toastCmd := m.toast.Show(msg.Err.Error(), tui.ToastError)
		return m, tea.Batch(toastCmd, reconnectTick(m.connGen))
	}

	if needsRender {
		m.Render()
	}
	if gotoBottom {
		m.viewport.GotoBottom()
	} else if yOffsetAdjust > 0 {
		// Compensate for lines evicted from the top of the ring buffer
		// so that the viewport stays pinned to the same content.
		newOffset := m.viewport.YOffset() - yOffsetAdjust
		if newOffset < 0 {
			newOffset = 0
		}
		m.viewport.SetYOffset(newOffset)
	}

	// Viewport update for non-visual mode scrolling (skip internal tick messages)
	if !m.visualMode && !m.showCommandDialog && !m.awaitingShortcut && !keyConsumed {
		if _, isBatchTick := msg.(batchTickMsg); !isBatchTick {
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// softWrapIndent wraps line so that the first terminal line occupies up to
// viewportWidth characters and every continuation line is indented by
// prefixWidth spaces (aligning with the start of the message column).
// When prefixWidth is 0 or leaves no room for the message, it falls back to
// a plain width-constrained render via lipgloss.
func softWrapIndent(line string, prefixWidth, viewportWidth int) string {
	msgWidth := viewportWidth - prefixWidth
	if prefixWidth <= 0 || msgWidth < 4 {
		// Fallback: no useful indent possible
		return lipgloss.NewStyle().Width(viewportWidth).Render(line)
	}

	// Split the assembled line into prefix and message portions.
	// PrefixWidth includes the trailing space, so line[:prefixWidth] is the
	// prefix with its separator and line[prefixWidth:] is the message text.
	var prefix, message string
	if prefixWidth < len(line) {
		prefix = line[:prefixWidth]
		message = line[prefixWidth:]
	} else {
		// Line is shorter than or equal to the prefix (no message content)
		return lipgloss.NewStyle().Width(viewportWidth).Render(line)
	}

	// Wrap the message part at the reduced width
	wrapped := ansi.Wrap(message, msgWidth, " ")
	msgLines := strings.Split(wrapped, "\n")

	indent := strings.Repeat(" ", prefixWidth)
	var sb strings.Builder
	for j, ml := range msgLines {
		if j == 0 {
			sb.WriteString(prefix)
		} else {
			sb.WriteString("\n")
			sb.WriteString(indent)
		}
		sb.WriteString(ml)
	}
	return sb.String()
}

func (m *LogcatViewModel) Render() {
	var b strings.Builder
	logs := m.log.All()
	cols := m.outputPrefs.Columns
	m.visualLineToLogLine = m.visualLineToLogLine[:0]
	m.logLineVisualStart = m.logLineVisualStart[:0]
	m.logLineColor = m.logLineColor[:0]
	for i, logLine := range logs {
		m.logLineVisualStart = append(m.logLineVisualStart, len(m.visualLineToLogLine))
		m.logLineColor = append(m.logLineColor, theme.GetLogColor(logLine.Level))
		line := logLine.ModifiedString(cols)
		var content string
		if m.outputPrefs.SoftWrap {
			content = softWrapIndent(line, logLine.PrefixWidth(cols), m.viewport.Width())
		} else {
			content = line
		}

		b.WriteString(content)
		b.WriteString("\n")
		for range strings.Count(content, "\n") + 1 {
			m.visualLineToLogLine = append(m.visualLineToLogLine, i)
		}
	}
	m.viewport.SetContent(b.String())
}

// handleKeyMsg routes key messages to appropriate handlers based on mode
func (m *LogcatViewModel) handleKeyMsg(msg tea.KeyPressMsg) updateResult {
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
		result := m.handleShortcutKey(key)
		result.consumed = true
		return result
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
		if m.visualMode {
			m.exitVisualMode()
			// Reader goroutine kept running during visual mode;
			// next batchTickMsg will drain accumulated lines.
			return updateResult{needsRender: true}, true
		}
		m.enterVisualMode()
		return updateResult{needsRender: true}, true
	}

	return updateResult{}, false
}

func (m *LogcatViewModel) enterVisualMode() {
	m.visualMode = true
	m.startSelected = -1
	m.viewport.GotoBottom()
	m.currentLine = m.log.Size() - 1
}

func (m *LogcatViewModel) exitVisualMode() {
	m.visualMode = false
	m.currentLine = -1
	m.startSelected = -1
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
			OutputPrefs:    m.outputPrefs,
			SelectedDevice: m.device,
			DeviceId:       deviceId,
			Recording:      m.recordingStartTime != nil,
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
		cmd := m.toast.Show("Unknown shortcut: ctrl+x "+key, tui.ToastWarning)
		return updateResult{cmd: cmd}
	}

	// Action commands execute immediately without opening a dialog
	if cmdData.Type == model.CommandTypeAction {
		switch cmdData.Command {
		case model.CommandReconnect:
			m.exitVisualMode()
			return updateResult{cmd: func() tea.Msg { return tui.ReconnectLogcatCmd{} }}
		case model.CommandExit:
			return updateResult{cmd: func() tea.Msg { return tui.ExitCmd{} }}
		}
		return updateResult{}
	}

	// Navigation commands open the sub-dialog directly
	cfg := commandui.DialogConfig{
		Filter:         m.filter,
		OutputPrefs:    m.outputPrefs,
		SelectedDevice: m.device,
		Recording:      m.recordingStartTime != nil,
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
		return updateResult{}

	case "esc":
		m.exitVisualMode()
		// Reader goroutine kept running during visual mode;
		// next batchTickMsg will drain accumulated lines.
		return updateResult{}

	case "y":
		if m.currentLine >= 0 && m.currentLine < m.log.Size() {
			var err error
			if m.startSelected >= 0 {
				start := min(m.currentLine, m.startSelected)
				end := max(m.currentLine, m.startSelected)
				var lines []string
				logs := m.log.Recent(m.log.Size() - start)
				for i := 0; i <= end-start; i++ {
					lines = append(lines, strings.TrimSpace(logs[i].ModifiedString(m.outputPrefs.Columns)))
				}
				err = util.CopyToClipboard(lines...)
				m.startSelected = -1
				return updateResult{}
			} else {
				logs := m.log.Recent(m.log.Size() - m.currentLine)
				lineText := strings.TrimSpace(logs[0].ModifiedString(m.outputPrefs.Columns))
				err = util.CopyToClipboard(lineText)
			}
			if err != nil {
				slog.Error("Failed to copy to clipboard", "error", err)
			}
		}
		return updateResult{}

	case "ctrl+e":
		lines := m.selectedLines()
		if len(lines) == 0 {
			return updateResult{}
		}
		cmd, err := util.OpenInEditor(lines...)
		if err != nil {
			slog.Warn("Failed to open editor", "error", err)
			toastCmd := m.toast.Show(err.Error(), tui.ToastError)
			return updateResult{cmd: toastCmd}
		}
		execCmd := tea.ExecProcess(cmd, func(err error) tea.Msg {
			return tui.EditorFinishedMsg{Err: err}
		})
		return updateResult{cmd: execCmd}

	case "j", "down":
		if m.currentLine < m.log.Size()-1 {
			m.currentLine++
			m.ensureLineVisible()
			return updateResult{}
		}

	case "k", "up":
		if m.currentLine > 0 {
			m.currentLine--
			m.ensureLineVisible()
			return updateResult{}
		}
	}

	return updateResult{}
}

// selectedLines returns the text of the currently selected log lines in visual mode.
// If a multi-line selection is active (startSelected >= 0), it returns all lines in the range.
// Otherwise, it returns the single line at currentLine.
func (m *LogcatViewModel) selectedLines() []string {
	if m.currentLine < 0 || m.currentLine >= m.log.Size() {
		return nil
	}

	cols := m.outputPrefs.Columns
	if m.startSelected >= 0 {
		start := min(m.currentLine, m.startSelected)
		end := max(m.currentLine, m.startSelected)
		logs := m.log.Recent(m.log.Size() - start)
		lines := make([]string, 0, end-start+1)
		for i := 0; i <= end-start; i++ {
			lines = append(lines, strings.TrimSpace(logs[i].ModifiedString(cols)))
		}
		return lines
	}

	logs := m.log.Recent(m.log.Size() - m.currentLine)
	return []string{strings.TrimSpace(logs[0].ModifiedString(cols))}
}

func (m *LogcatViewModel) ensureLineVisible() {
	if !m.visualMode || m.currentLine < 0 || m.currentLine >= m.log.Size() {
		return
	}
	if m.currentLine >= len(m.logLineVisualStart) {
		return
	}

	lineTop := m.logLineVisualStart[m.currentLine]
	lineBottom := lineTop
	if next := m.currentLine + 1; next < len(m.logLineVisualStart) {
		lineBottom = m.logLineVisualStart[next] - 1
	} else if len(m.visualLineToLogLine) > 0 {
		lineBottom = len(m.visualLineToLogLine) - 1
	}

	viewportTop := m.viewport.YOffset()
	viewportBottom := viewportTop + m.viewport.Height() - 1
	if lineTop < viewportTop {
		m.viewport.SetYOffset(lineTop)
	} else if lineBottom > viewportBottom {
		m.viewport.SetYOffset(lineBottom - m.viewport.Height() + 1)
	}
}

func (m LogcatViewModel) visualLineStyle(lineIndex int) lipgloss.Style {
	if lineIndex < 0 || lineIndex >= len(m.visualLineToLogLine) {
		return lipgloss.NewStyle()
	}
	logLine := m.visualLineToLogLine[lineIndex]
	if m.visualMode && m.logLineSelected(logLine) {
		return lipgloss.NewStyle().
			Background(theme.ColorVisualBG).
			Foreground(theme.ColorVisualFG).
			Width(m.viewport.Width())
	}
	if !m.outputPrefs.Color {
		return lipgloss.NewStyle()
	}
	if logLine < 0 || logLine >= len(m.logLineColor) {
		return lipgloss.NewStyle()
	}
	levelColor := m.logLineColor[logLine]
	if theme.IsDefaultForeground(levelColor) {
		return lipgloss.NewStyle()
	}
	return lipgloss.NewStyle().Foreground(levelColor)
}

func (m LogcatViewModel) logLineSelected(line int) bool {
	if m.startSelected >= 0 {
		return line >= min(m.currentLine, m.startSelected) &&
			line <= max(m.currentLine, m.startSelected)
	}
	return line == m.currentLine
}

func (m LogcatViewModel) viewportView() string {
	viewport := m.viewport
	viewport.StyleLineFunc = m.visualLineStyle
	return viewport.View()
}

func (m LogcatViewModel) renderBaseView() string {
	return fmt.Sprintf(
		"%s\n%s\n%s",
		m.headerView(),
		m.viewportView(),
		m.footerView(),
	)
}

func (m LogcatViewModel) View() string {
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

func filterModeSep(mode model.TextFilterMode) string {
	switch mode {
	case model.FilterModeExact:
		return "=:"
	case model.FilterModeRegex:
		return "~:"
	default:
		return ":"
	}
}

func (m LogcatViewModel) headerView() string {
	// Build device label
	var name string
	if m.device != nil {
		name = m.device.Name
	} else {
		name = "No device"
	}
	if m.recordingStartTime != nil {
		recDot := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("●")
		name = recDot + " " + name
	}
	deviceName := lipgloss.NewStyle().Bold(true).Render(name)

	// Build filter parts: package, tag, content (text), log level
	badge := theme.FilterBadge()
	var filters []string
	if !m.filter.PackageName.IsEmpty() {
		filters = append(filters, badge.Render(fmt.Sprintf("pkg%s%s", filterModeSep(m.filter.PackageName.Mode), m.filter.PackageName.Value)))
	}
	if m.filter.Level != "" && m.filter.Level != model.LvlV {
		filters = append(filters, badge.Render(fmt.Sprintf("lvl:%s", string(m.filter.Level))))
	}
	if !m.filter.Tag.IsEmpty() {
		filters = append(filters, badge.Render(fmt.Sprintf("tag%s%s", filterModeSep(m.filter.Tag.Mode), m.filter.Tag.Value)))
	}
	if !m.filter.Text.IsEmpty() {
		filters = append(filters, badge.Render(fmt.Sprintf("content%s%s", filterModeSep(m.filter.Text.Mode), m.filter.Text.Value)))
	}

	// Compose single-line header content
	headerContent := deviceName
	if len(filters) > 0 {
		headerContent = fmt.Sprintf("%s: %s", deviceName, strings.Join(filters, " "))
	}

	// border (2) + padding (2) = 4 chars horizontal overhead
	innerWidth := m.viewport.Width() - 4
	style := titleStyle.Width(m.viewport.Width())

	// Toast has higher priority: reserve space for it first, then truncate header content
	toastStr := m.toast.View()
	if toastStr != "" {
		toastWidth := ansi.StringWidth(toastStr)
		gap := 2 // spacing between header content and toast
		contentMaxWidth := innerWidth - toastWidth - gap

		if contentMaxWidth > 3 {
			if ansi.StringWidth(headerContent) > contentMaxWidth {
				headerContent = ansi.Truncate(headerContent, contentMaxWidth-3, "...")
			}
		} else {
			headerContent = ""
		}

		// Compose line: left-aligned header content + right-aligned toast
		headerWidth := ansi.StringWidth(headerContent)
		rightWidth := innerWidth - headerWidth
		rightPart := lipgloss.NewStyle().
			Width(rightWidth).
			AlignHorizontal(lipgloss.Right).
			Render(toastStr)
		return style.Render(headerContent + rightPart)
	}

	// No toast: truncate header content if needed
	if ansi.StringWidth(headerContent) > innerWidth {
		if innerWidth > 3 {
			headerContent = ansi.Truncate(headerContent, innerWidth-3, "...")
		} else {
			headerContent = ansi.Truncate(headerContent, innerWidth, "")
		}
	}

	return style.Render(headerContent)
}

func (m LogcatViewModel) footerView() string {
	var helpText string
	if m.visualMode {
		helpText = helpTextVisual
	} else {
		helpText = helpTextNormal
	}

	help := lipgloss.NewStyle().
		Foreground(theme.ColorMuted).
		Width(m.viewport.Width()).
		AlignHorizontal(lipgloss.Center).
		Padding(0, 2).
		Render(helpText)

	return help
}

// exportBuffer writes all buffered log lines to a timestamped file in the
// working directory. Returns a tea.Cmd that shows a toast on success or error.
func (m *LogcatViewModel) exportBuffer() tea.Cmd {
	lines := m.log.All()
	if len(lines) == 0 {
		return m.toast.Show("Buffer is empty, nothing to export", tui.ToastWarning)
	}

	var sb strings.Builder
	for _, l := range lines {
		sb.WriteString(strings.TrimSpace(l.ModifiedString(m.outputPrefs.Columns)))
		sb.WriteByte('\n')
	}

	filename := fmt.Sprintf("logcat_%s.log", time.Now().Format("20060102_150405"))
	if err := os.WriteFile(filename, []byte(sb.String()), 0644); err != nil {
		slog.Error("Failed to export buffer", "error", err)
		return m.toast.Show("Export failed: "+err.Error(), tui.ToastError)
	}

	slog.Debug("Buffer exported", "file", filename, "lines", len(lines))
	return m.toast.Show(fmt.Sprintf("Exported %d lines to %s", len(lines), filename), tui.ToastInfo)
}

func (m *LogcatViewModel) startRecording() tea.Cmd {
	now := time.Now()
	m.recordingStartTime = &now
	slog.Debug("Recording started", "time", now)
	return m.toast.Show("Recording started", tui.ToastInfo)
}

func (m *LogcatViewModel) stopRecording() tea.Cmd {
	startTime := m.recordingStartTime
	m.recordingStartTime = nil

	lines := m.log.All()
	var recorded []model.LogLine
	for _, l := range lines {
		if t, ok := parseLogLineTime(l); !ok || !t.Before(*startTime) {
			recorded = append(recorded, l)
		}
	}

	if len(recorded) == 0 {
		return m.toast.Show("No lines recorded", tui.ToastWarning)
	}

	var sb strings.Builder
	for _, l := range recorded {
		sb.WriteString(strings.TrimSpace(l.ModifiedString(m.outputPrefs.Columns)))
		sb.WriteByte('\n')
	}

	filename := fmt.Sprintf("logcat_rec_%s.log", time.Now().Format("20060102_150405"))
	if err := os.WriteFile(filename, []byte(sb.String()), 0644); err != nil {
		slog.Error("Failed to save recording", "error", err)
		return m.toast.Show("Export failed: "+err.Error(), tui.ToastError)
	}

	slog.Debug("Recording saved", "file", filename, "lines", len(recorded))
	return m.toast.Show(fmt.Sprintf("Recorded %d lines to %s", len(recorded), filename), tui.ToastInfo)
}

// parseLogLineTime parses the Date and Time fields of a log line into a time.Time.
// Returns (time, true) on success, or (zero, false) if the line is unparsed or malformed.
func parseLogLineTime(l model.LogLine) (time.Time, bool) {
	if !l.Parsed() {
		return time.Time{}, false
	}
	year := time.Now().Year()
	str := fmt.Sprintf("%d-%s %s", year, l.Date, l.Time)
	t, err := time.Parse("2006-01-02 15:04:05.000", str)
	return t, err == nil
}

func Close(m *LogcatViewModel) {
	m.reader.Disconnect()
	m.log = util.NewRingBuffer(maxLogLines)
}
