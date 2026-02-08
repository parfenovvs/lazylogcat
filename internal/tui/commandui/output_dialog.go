package commandui

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/parfenovvs/lazylogcat/internal/model"
	"github.com/parfenovvs/lazylogcat/internal/tui"
)

type OutputPrefsChangedMsg struct {
	OutputPrefs model.OutputPrefs
}

// outputMultiSelectItems defines the items shown in the Output multiselect dialog.
// The order matches the display order; keys are used for the active map.
var outputMultiSelectItems = []string{
	"Color",
	"Wrap",
	"Date",
	"Time",
	"PID",
	"TID",
	"Level",
	"Tag",
	"Message",
}

func newOutputMultiSelect(prefs model.OutputPrefs) MultiSelectModel {
	active := make(map[string]bool)
	if prefs.Color {
		active["Color"] = true
	}
	if prefs.SoftWrap {
		active["Wrap"] = true
	}
	if prefs.Columns.Date {
		active["Date"] = true
	}
	if prefs.Columns.Time {
		active["Time"] = true
	}
	if prefs.Columns.PID {
		active["PID"] = true
	}
	if prefs.Columns.TID {
		active["TID"] = true
	}
	if prefs.Columns.Level {
		active["Level"] = true
	}
	if prefs.Columns.Tag {
		active["Tag"] = true
	}
	if prefs.Columns.Message {
		active["Message"] = true
	}
	return NewMultiSelect(MultiSelectConfig{
		Title:   "Output",
		Footer:  "space/enter to toggle · esc to apply",
		Columns: []table.Column{{Title: "", Width: tui.DialogWidth - 9}},
		Items:   outputMultiSelectItems,
		Active:  active,
	})
}

func outputPrefsFromActive(active map[string]bool) model.OutputPrefs {
	return model.OutputPrefs{
		Color:    active["Color"],
		SoftWrap: active["Wrap"],
		Columns: model.Columns{
			Date:    active["Date"],
			Time:    active["Time"],
			PID:     active["PID"],
			TID:     active["TID"],
			Level:   active["Level"],
			Tag:     active["Tag"],
			Message: active["Message"],
		},
	}
}

func (m CommandDialogModel) updateMultiSelect(msg tea.KeyMsg, key string) (CommandDialogModel, tea.Cmd) {
	_ = key
	m.multiSelect, _ = m.multiSelect.Update(msg, key)
	return m, nil
}
