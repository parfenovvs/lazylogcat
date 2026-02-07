package tui

import "github.com/parfenovvs/lazylogcat/internal/model"

type MeasureCmd struct{}

type NavigateToLogcatCmd struct{}

type NavigateToFilterCmd struct{}

type DeviceSelectedMsg struct {
	Device   model.Device
	Filter   model.Filter
	Format   model.Format
	SoftWrap bool
}

type ShowDeviceDialogCmd struct{}

type UpdateFilterCmd struct {
	Filter model.Filter
	Format model.Format
}

type ReconnectLogcatCmd struct{}

type ExitCmd struct{}
