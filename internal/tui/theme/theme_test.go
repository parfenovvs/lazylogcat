package theme

import (
	"image/color"
	"testing"

	"charm.land/lipgloss/v2"
)

func colorsEqual(a, b color.Color) bool {
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	return ar == br && ag == bg && ab == bb && aa == ba
}

func TestGetLogColor(t *testing.T) {
	tests := []struct {
		name  string
		level string
		want  color.Color
	}{
		{name: "Verbose", level: "V", want: ColorRegular},
		{name: "Debug", level: "D", want: lipgloss.Color(blue)},
		{name: "Info", level: "I", want: lipgloss.Color(green)},
		{name: "Warn", level: "W", want: lipgloss.Color(yellow)},
		{name: "Error", level: "E", want: lipgloss.Color(red)},
		{name: "Fatal", level: "F", want: lipgloss.Color(magenta)},
		{name: "Unknown", level: "X", want: ColorRegular},
		{name: "Empty", level: "", want: ColorRegular},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetLogColor(tt.level)
			if !colorsEqual(got, tt.want) {
				t.Errorf("GetLogColor(%q) = %v, want %v", tt.level, got, tt.want)
			}
		})
	}
}
