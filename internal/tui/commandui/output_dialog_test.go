package commandui

import (
	"testing"

	"github.com/parfenovvs/lazylogcat/internal/model"
)

func TestOutputPrefsFromActive(t *testing.T) {
	tests := []struct {
		name   string
		active map[string]bool
		want   model.OutputPrefs
	}{
		{
			name:   "EmptyMap_AllFalse",
			active: map[string]bool{},
			want:   model.OutputPrefs{},
		},
		{
			name: "AllKeysTrue",
			active: map[string]bool{
				"Color": true, "Wrap": true,
				"Date": true, "Time": true, "PID": true, "TID": true,
				"Level": true, "Tag": true, "Message": true,
			},
			want: model.OutputPrefs{
				Color:    true,
				SoftWrap: true,
				Columns: model.Columns{
					Date: true, Time: true, PID: true, TID: true,
					Level: true, Tag: true, Message: true,
				},
			},
		},
		{
			name:   "OnlyColor",
			active: map[string]bool{"Color": true},
			want:   model.OutputPrefs{Color: true},
		},
		{
			name:   "OnlyWrap",
			active: map[string]bool{"Wrap": true},
			want:   model.OutputPrefs{SoftWrap: true},
		},
		{
			name:   "OnlyColumns",
			active: map[string]bool{"Date": true, "Level": true, "Message": true},
			want: model.OutputPrefs{
				Columns: model.Columns{Date: true, Level: true, Message: true},
			},
		},
		{
			name:   "PartialMix",
			active: map[string]bool{"Color": true, "Time": true, "Tag": true},
			want: model.OutputPrefs{
				Color:   true,
				Columns: model.Columns{Time: true, Tag: true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := outputPrefsFromActive(tt.active)
			if got != tt.want {
				t.Errorf("outputPrefsFromActive(%v) = %+v, want %+v", tt.active, got, tt.want)
			}
		})
	}
}

func TestOutputPrefsMergePreservesTagWidth(t *testing.T) {
	// Esc from Output multiselect merges columns from the widget but must keep TagWidth
	// from the dialog model (see command_dialog Esc handling).
	prev := model.OutputPrefs{
		TagWidth: 17,
		Color:    true,
		Columns:  model.Columns{Time: true, Tag: true},
	}
	active := map[string]bool{"Wrap": true, "Level": true}
	got := outputPrefsFromActive(active)
	got.TagWidth = prev.TagWidth
	want := model.OutputPrefs{
		TagWidth: 17,
		SoftWrap: true,
		Columns:  model.Columns{Level: true},
	}
	if got != want {
		t.Errorf("merged prefs = %+v, want %+v", got, want)
	}
}
