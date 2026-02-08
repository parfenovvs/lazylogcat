package mainui

import (
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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
		SoftWrap: true,
		Columns: model.Columns{
			Date:    true,
			Time:    true,
			PID:     true,
			TID:     true,
			Level:   true,
			Tag:     true,
			Message: true,
		},
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
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			logcatui.Close(&m.logcatView)
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.windowSize = model.Size{
			Width:  msg.Width - 1,
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

func (m MainModel) View() string {
	style = style.Width(m.windowSize.Width).
		Height(m.windowSize.Height)

	var content string

	switch m.state {
	case logcatView:
		content = m.logcatView.View()
	}

	return style.Render(content)
}
