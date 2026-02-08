package tui

import "github.com/parfenovvs/lazylogcat/internal/model"

type MeasureCmd struct{}

type NavigateToLogcatCmd struct{}

type DeviceSelectedMsg struct {
	Device   model.Device
	Filter   model.Filter
	Color    bool
	SoftWrap bool
}

type ShowDeviceDialogCmd struct{}

type ReconnectLogcatCmd struct{}

type ExitCmd struct{}
