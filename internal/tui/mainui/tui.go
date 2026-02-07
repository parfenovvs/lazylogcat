package mainui

import (
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/parfenovvs/lazylogcat/internal/config"
	"github.com/parfenovvs/lazylogcat/internal/model"
	"github.com/parfenovvs/lazylogcat/internal/tui"
	"github.com/parfenovvs/lazylogcat/internal/tui/filterui"
	"github.com/parfenovvs/lazylogcat/internal/tui/logcatui"
	"github.com/parfenovvs/lazylogcat/internal/util"
)

var style = lipgloss.NewStyle()

type sessionState int

const (
	logcatView sessionState = iota
	filterView
)

type MainModel struct {
	windowSize model.Size

	state sessionState

	currentDevice  *model.Device
	deviceRequired bool
	filter         model.Filter
	format         model.Format

	logcatView logcatui.LogcatViewModel
	filterView filterui.FilterViewModel
}

func InitMainModel(c config.Config) MainModel {
	var m MainModel

	devices, err := util.GetConnectedDevices()
	if err != nil {
		slog.Error("Failed to get connected devices", "error", err)
	}

	// Device resolution:
	// 1. If config has a device_id, try to find it among connected devices
	if err == nil && len(devices) > 0 && c.Session.DeviceId != "" {
		for _, d := range devices {
			if d.Id == c.Session.DeviceId {
				m.currentDevice = &d
				break
			}
		}
	}

	// 2. If config device not found, but devices exist, use first available
	if m.currentDevice == nil && err == nil && len(devices) > 0 {
		if c.Session.DeviceId != "" {
			// Config specified a device that's not connected - show device dialog
			m.deviceRequired = true
		}
		if !m.deviceRequired {
			m.currentDevice = &devices[0]
		}
	}

	// 3. No devices at all - show device dialog
	if m.currentDevice == nil && !m.deviceRequired {
		m.deviceRequired = true
	}

	// Clear package filter if no device selected
	if m.currentDevice == nil {
		c.Session.Pkg = ""
	}
	if c.Session.Pkg != "" {
		_, err := util.GetPidByPackageName(m.currentDevice.Id, c.Session.Pkg)
		if err != nil {
			c.Session.Pkg = ""
		}
	}

	m.filter = util.FilterFromConfig(&c)
	m.format = util.FormatFromConfig(&c)

	// Always start in logcat view
	m.state = logcatView
	m.logcatView = logcatui.New(m.windowSize, m.currentDevice, m.filter, m.format)

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

	case tui.DeviceSelectedMsg:
		m.currentDevice = &msg.Device
		m.deviceRequired = false
		return m, func() tea.Msg {
			return tui.NavigateToLogcatCmd{}
		}

	case tui.NavigateToFilterCmd:
		m.state = filterView
		m.filterView = filterui.New(
			m.windowSize,
			m.currentDevice.Id,
			m.filter,
			m.format,
		)
		return m, nil

	case tui.UpdateFilterCmd:
		m.filter = msg.Filter
		m.format = msg.Format
		return m, func() tea.Msg {
			return tui.NavigateToLogcatCmd{}
		}

	case tui.NavigateToLogcatCmd:
		m.state = logcatView
		logcatui.Close(&m.logcatView)
		m.logcatView = logcatui.New(m.windowSize, m.currentDevice, m.filter, m.format)
		return m, func() tea.Msg {
			return tui.ReconnectLogcatCmd{}
		}
	}

	switch m.state {
	case logcatView:
		newLogcatViewing, newCmd := m.logcatView.Update(msg)
		m.logcatView = newLogcatViewing
		cmd = newCmd

	case filterView:
		newFilterView, newCmd := m.filterView.Update(msg)
		m.filterView = newFilterView
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
	case filterView:
		content = m.filterView.View()
	}

	return style.Render(content)
}
