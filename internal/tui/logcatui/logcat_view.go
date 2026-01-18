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

	helpTextNormal = "ctrl+f filters • ctrl+r reconnect • ctrl+d devices • W toggle wrap • L toggle level • G jump to recent • C clear • v visual"
	helpTextVisual = "j/↓ down • k/↑ up • V select multiple • y copy • esc exit visual"
)

type LogcatViewModel struct {
	parentSize    model.Size
	viewport      viewport.Model
	device        model.Device
	filter        model.Filter
	format        model.Format
	log           *util.RingBuffer
	pendingLogs   []string
	visualMode    bool
	currentLine   int
	startSelected int
	softWrap      bool
	err           error
}

type logcatMsg struct {
	Line string
}

type logcatErrorMsg struct {
	Err error
}

type logcatConnectedMsg struct{}

type batchTickMsg struct{}

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
	if !m.format.Long && strings.Trim(line, "\n\r ") == "" {
		return nil
	}

	// Filter by text search
	if m.filter.Text != "" && !strings.Contains(line, m.filter.Text) {
		return nil
	}

	return logcatMsg{Line: line}
}

func tickForBatch() tea.Cmd {
	return tea.Tick(batchTimeout, func(t time.Time) tea.Msg {
		return batchTickMsg{}
	})
}

func New(parentSize model.Size, device model.Device, filter model.Filter, format model.Format) LogcatViewModel {
	m := LogcatViewModel{
		parentSize:    parentSize,
		device:        device,
		log:           util.NewRingBuffer(maxLogLines),
		softWrap:      true,
		startSelected: -1,
		filter:        filter,
		format:        format,
	}

	headerHeight := lipgloss.Height(m.headerView())
	footerHeight := lipgloss.Height(m.footerView())
	vp := viewport.New(parentSize.Width, parentSize.Height-footerHeight-headerHeight-1)
	m.viewport = vp

	return m
}

func (m LogcatViewModel) Update(msg tea.Msg) (LogcatViewModel, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
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
		m.Render()
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+r":
			m.visualMode = false
			return m, func() tea.Msg {
				return tui.ReconnectLogcatCmd{}
			}

		case "ctrl+d":
			return m, func() tea.Msg {
				return tui.NavigateToDevicesCmd{}
			}

		case "ctrl+f":
			return m, func() tea.Msg {
				return tui.NavigateToFilterCmd{}
			}

		case "W":
			if !m.visualMode {
				m.softWrap = !m.softWrap
				return m, func() tea.Msg {
					return tui.ReconnectLogcatCmd{}
				}
			}
			return m, nil

		case "G":
			m.viewport.GotoBottom()
			return m, nil

		case "L":
			if !m.visualMode {
				m.filter.Level = m.filter.Level.Next()
				return m, func() tea.Msg {
					return tui.ReconnectLogcatCmd{}
				}
			}

		case "C":
			if !m.visualMode {
				m.log.Clear()
				m.Render()
				return m, nil
			}

		case "v":
			m.visualMode = !m.visualMode
			if m.visualMode {
				m.viewport.GotoBottom()
				m.currentLine = m.log.Size() - 1
				m.Render()
				return m, nil
			}
			return m, func() tea.Msg {
				return tui.ReconnectLogcatCmd{}
			}

		case "V":
			if m.visualMode {
				if m.startSelected >= 0 {
					m.startSelected = -1
				} else {
					m.startSelected = m.currentLine
				}
				m.Render()
				return m, nil
			}

		case "esc":
			if m.visualMode {
				if m.startSelected >= 0 {
					m.startSelected = -1
					m.Render()
					return m, nil
				}
				m.visualMode = false
				m.startSelected = -1
				return m, func() tea.Msg {
					return tui.ReconnectLogcatCmd{}
				}
			}

		case "y":
			if m.visualMode && m.currentLine >= 0 && m.currentLine < m.log.Size() {
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
					m.Render()
				} else {
					logs := m.log.Recent(m.log.Size() - m.currentLine)
					lineText := strings.TrimSpace(logs[0])
					err = util.CopyToClipboard(lineText)
				}
				if err != nil {
					slog.Error("Failed to copy to clipboard", "error", err)
				}
			}
			return m, nil

		case "j", "down":
			if m.visualMode && m.currentLine < m.log.Size()-1 {
				m.currentLine++
				m.Render()
				m.ensureLineVisible()
			}

		case "k", "up":
			if m.visualMode && m.currentLine > 0 {
				m.currentLine--
				m.Render()
				m.ensureLineVisible()
			}
		}

	case tui.ReconnectLogcatCmd:
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

	case batchTickMsg:
		if m.visualMode {
			return m, nil
		}
		if len(m.pendingLogs) > 0 {
			wasAtBottom := m.viewport.AtBottom()
			for _, line := range m.pendingLogs {
				m.log.Append(line + "\n")
			}
			m.pendingLogs = nil
			m.Render()
			if wasAtBottom {
				m.viewport.GotoBottom()
			}
		}
		return m, tickForBatch()

	case logcatErrorMsg:
		m.err = msg.Err
		util.CloseLogcat()
		return m, nil
	}

	if !m.visualMode {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
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
		if m.format.Color {
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

func (m LogcatViewModel) View() string {
	if m.err != nil {
		slog.Error("Logcat view error", "error", m.err)
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	return fmt.Sprintf(
		"%s\n%s\n%s",
		m.headerView(),
		m.viewport.View(),
		m.footerView(),
	)
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
	var mods []string
	var modsStr string

	if m.format.Color {
		mods = append(mods, "color")
	}
	if m.format.Descriptive {
		mods = append(mods, "descriptive")
	}
	if m.format.Epoch {
		mods = append(mods, "epoch")
	}
	if m.format.Monotonic {
		mods = append(mods, "monotonic")
	}
	if m.format.Printable {
		mods = append(mods, "printable")
	}
	if m.format.Uid {
		mods = append(mods, "uid")
	}
	if m.format.Usec {
		mods = append(mods, "usec")
	}
	if m.format.UTC {
		mods = append(mods, "UTC")
	}
	if m.format.Year {
		mods = append(mods, "year")
	}
	if m.format.Zone {
		mods = append(mods, "zone")
	}

	if len(mods) > 0 {
		modsStr = fmt.Sprintf(" | %s", strings.Join(mods, ","))
	}

	filtersStr := ""
	if len(filters) > 0 {
		filtersStr = fmt.Sprintf("\n%s", strings.Join(filters, " | "))
	}

	deviceName := lipgloss.NewStyle().Bold(true).Render(m.device.Name)
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
