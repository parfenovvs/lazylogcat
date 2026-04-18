package mainui

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/parfenovvs/lazylogcat/internal/config"
	"github.com/parfenovvs/lazylogcat/internal/model"
	"github.com/parfenovvs/lazylogcat/internal/tui"
	"github.com/parfenovvs/lazylogcat/internal/tui/logcatui"
	"github.com/parfenovvs/lazylogcat/internal/util"
)

var style = lipgloss.NewStyle()

type sessionState int

const (
	logcatView sessionState = iota
)

type MainModel struct {
	windowSize model.Size

	state sessionState

	currentDevice  *model.Device
	deviceRequired bool
	filter         model.Filter
	outputPrefs    model.OutputPrefs

	logcatView logcatui.LogcatViewModel
}

func (m MainModel) deviceId() string {
	if m.currentDevice != nil {
		return m.currentDevice.Id
	}
	return ""
}

func InitMainModel(c config.Config) MainModel {
	var m MainModel

	devices, err := util.GetConnectedDevices()
	if err != nil {
		slog.Error("Failed to get connected devices", "error", err)
	}

	// Device resolution: use first available device, or show device dialog
	if err == nil && len(devices) > 0 {
		m.currentDevice = &devices[0]
	} else {
		m.deviceRequired = true
	}

	m.filter = util.FilterFromConfig(&c)
	m.outputPrefs = model.OutputPrefs{
		Color:    util.ColorFromConfig(&c),
		SoftWrap: util.WrapFromConfig(&c),
		Columns:  util.ColumnsFromConfig(&c),
	}

	// Always start in logcat view
	m.state = logcatView
	m.logcatView = logcatui.New(m.windowSize, m.currentDevice, m.deviceId(), m.filter, m.outputPrefs)

	return m
}

func (m MainModel) Init() tea.Cmd {
	if m.deviceRequired {
		return func() tea.Msg {
			return tui.ShowDeviceDialogCmd{}
		}
	}
	return func() tea.Msg {
		return tui.ReconnectLogcatCmd{}
	}
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			logcatui.Close(&m.logcatView)
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.windowSize = model.Size{
			Width:  msg.Width,
			Height: msg.Height,
		}
		return m, func() tea.Msg {
			return m.windowSize
		}

	case tui.ExitCmd:
		logcatui.Close(&m.logcatView)
		return m, tea.Quit

	case tui.DeviceSelectedMsg:
		m.currentDevice = &msg.Device
		m.deviceRequired = false
		m.filter = msg.Filter
		m.outputPrefs = msg.OutputPrefs
		return m, func() tea.Msg {
			return tui.NavigateToLogcatCmd{}
		}

	case tui.NavigateToLogcatCmd:
		m.state = logcatView
		logcatui.Close(&m.logcatView)
		m.logcatView = logcatui.New(m.windowSize, m.currentDevice, m.deviceId(), m.filter, m.outputPrefs)
		return m, func() tea.Msg {
			return tui.ReconnectLogcatCmd{}
		}
	}

	switch m.state {
	case logcatView:
		newLogcatViewing, newCmd := m.logcatView.Update(msg)
		m.logcatView = newLogcatViewing
		cmd = newCmd
	}

	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m MainModel) View() tea.View {
	style = style.Width(m.windowSize.Width).
		Height(m.windowSize.Height)

	var content string

	switch m.state {
	case logcatView:
		content = m.logcatView.View()
	}

	v := tea.NewView(style.Render(content))
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}
