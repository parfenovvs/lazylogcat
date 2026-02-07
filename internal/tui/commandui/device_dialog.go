package commandui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/parfenovvs/lazylogcat/internal/model"
	"github.com/parfenovvs/lazylogcat/internal/tui/theme"
	"github.com/parfenovvs/lazylogcat/internal/util"
)

type CommandDialogDeviceSelectedMsg struct {
	Device model.Device
}

func loadDevices(selectedDevice *model.Device) (table.Model, map[int]model.Device, error) {
	devices, err := util.GetConnectedDevices()
	if err != nil {
		return table.Model{}, nil, fmt.Errorf("failed to get devices: %w", err)
	}

	t, dm := newDeviceTable(devices, selectedDevice)
	return t, dm, nil
}

func newDeviceTable(devices []model.Device, selectedDevice *model.Device) (table.Model, map[int]model.Device) {
	columns := []table.Column{
		{Title: "", Width: 20},
		{Title: "", Width: 18},
		{Title: "", Width: 3},
	}

	deviceMap := make(map[int]model.Device)
	var rows []table.Row
	initialCursor := 0
	for i, device := range devices {
		deviceMap[i] = device
		marker := ""
		if selectedDevice != nil && selectedDevice.Id == device.Id {
			marker = "●"
			initialCursor = i
		}

		name := device.Name
		maxNameLen := 18
		if len(name) > maxNameLen {
			name = name[:maxNameLen-3] + "..."
		}

		rows = append(rows, table.Row{name, device.Id, marker})
	}

	height := len(rows) + 1
	if height < 2 {
		height = 2
	}

	t := newTable(columns, rows, height)
	if len(rows) > 0 {
		t.SetCursor(initialCursor)
	}

	return t, deviceMap
}

func (m CommandDialogModel) updateDevices(msg tea.KeyMsg, key string) (CommandDialogModel, tea.Cmd) {
	if key == "r" {
		deviceTable, deviceMap, err := loadDevices(m.selectedDevice)
		m.deviceTable = deviceTable
		m.deviceMap = deviceMap
		m.deviceErr = err
		return m, nil
	}

	if key == "enter" {
		if device, ok := m.deviceMap[m.deviceTable.Cursor()]; ok {
			return m, func() tea.Msg { return CommandDialogDeviceSelectedMsg{Device: device} }
		}
		return m, nil
	}

	if len(m.deviceMap) > 0 {
		m.deviceTable, _ = m.deviceTable.Update(msg)
	}
	return m, nil
}

func (m CommandDialogModel) viewDevices() string {
	title := lipgloss.NewStyle().Bold(true).Render("Select Device")

	var body string
	if m.deviceErr != nil {
		errorMsg := lipgloss.NewStyle().
			Foreground(theme.FGHelp).
			Render(fmt.Sprintf("Error: %s", m.deviceErr.Error()))
		hint := lipgloss.NewStyle().
			Foreground(theme.FGHelp).
			Render("Press 'r' to retry")
		body = "\n" + errorMsg + "\n\n" + hint
	} else if len(m.deviceMap) == 0 {
		emptyMsg := lipgloss.NewStyle().
			Foreground(theme.FGHelp).
			Render("No devices connected.")
		hint := lipgloss.NewStyle().
			Foreground(theme.FGHelp).
			Render("Press 'r' to refresh")
		body = "\n" + emptyMsg + "\n\n" + hint
	} else {
		body = m.deviceTable.View()
	}

	footer := lipgloss.NewStyle().Foreground(theme.FGHelp).Render("esc to close • r refresh")
	content := title + "\n" + body + "\n" + footer
	return dialogStyle().Render(content)
}
