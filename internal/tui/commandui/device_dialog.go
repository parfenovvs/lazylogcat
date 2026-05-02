package commandui

import (
	"fmt"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"

	"github.com/parfenovvs/lazylogcat/internal/model"
	"github.com/parfenovvs/lazylogcat/internal/tui"
	"github.com/parfenovvs/lazylogcat/internal/tui/commonui"
	"github.com/parfenovvs/lazylogcat/internal/tui/theme"
	"github.com/parfenovvs/lazylogcat/internal/util"
)

type CommandDialogDeviceSelectedMsg struct {
	Device model.Device
}

func deviceSingleSelectItems(devices []model.Device) []SingleSelectItem {
	var items []SingleSelectItem
	for _, device := range devices {
		name := device.Name
		maxNameLen := 18
		if len(name) > maxNameLen {
			name = name[:maxNameLen-3] + "..."
		}
		items = append(items, SingleSelectItem{
			Key:     device.Id,
			Columns: []string{name, device.Id},
		})
	}
	return items
}

func newDeviceSingleSelect(devices []model.Device, selectedDevice *model.Device) SingleSelectModel {
	currentKey := ""
	if selectedDevice != nil {
		currentKey = selectedDevice.Id
	}
	return NewSingleSelect(SingleSelectConfig{
		Title:  "Select Device",
		Footer: "Refresh: r",
		Columns: []table.Column{
			{Title: "", Width: 14},
			{Title: "", Width: tui.DialogWidth - 23},
		},
		Items:      deviceSingleSelectItems(devices),
		CurrentKey: currentKey,
	})
}

func loadDevices(selectedDevice *model.Device) (SingleSelectModel, []model.Device, error) {
	devices, err := util.GetConnectedDevices()
	if err != nil {
		return SingleSelectModel{}, nil, fmt.Errorf("failed to get devices: %w", err)
	}
	ss := newDeviceSingleSelect(devices, selectedDevice)
	return ss, devices, nil
}

func (m CommandDialogModel) updateDevices(msg tea.KeyPressMsg, key string) (CommandDialogModel, tea.Cmd) {
	if key == "r" {
		ss, allDevices, err := loadDevices(m.selectedDevice)
		m.singleSelect = ss
		m.allDevices = allDevices
		m.deviceErr = err
		return m, nil
	}

	var cmd tea.Cmd
	m.singleSelect, cmd = m.singleSelect.Update(msg, key)
	if m.singleSelect.Selected() {
		deviceId := m.singleSelect.SelectedKey()
		// Find the device by ID from allDevices
		for _, device := range m.allDevices {
			if device.Id == deviceId {
				d := device
				return m, func() tea.Msg { return CommandDialogDeviceSelectedMsg{Device: d} }
			}
		}
	}
	return m, cmd
}

func (m CommandDialogModel) viewDevices() string {
	// Devices has special error/empty states, so we render manually instead of using singleSelect.View()
	title := commonui.DialogTitleWithESC("Select Device")

	var body string
	if m.deviceErr != nil {
		errorMsg := theme.DialogHelp().
			Render(fmt.Sprintf("Error: %s", m.deviceErr.Error()))
		body = "\n" + errorMsg + "\n"
	} else if len(m.allDevices) == 0 {
		emptyMsg := theme.DialogHelp().
			Render("No devices connected.")
		body = "\n" + emptyMsg + "\n"
	} else {
		// Delegate to singleSelect.View() when we have devices
		return m.singleSelect.View()
	}

	footer := theme.DialogHelp().Render("Refresh: r")
	content := title + "\n\n" + body + "\n\n" + footer
	return commonui.DialogFrameStyle().Render(content)
}
