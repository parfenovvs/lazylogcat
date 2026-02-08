package model

import "testing"

func TestLevelIndex(t *testing.T) {
	tests := []struct {
		name  string
		level string
		want  int
	}{
		{name: "Verbose", level: "V", want: 0},
		{name: "Debug", level: "D", want: 1},
		{name: "Info", level: "I", want: 2},
		{name: "Warn", level: "W", want: 3},
		{name: "Error", level: "E", want: 4},
		{name: "Fatal", level: "F", want: 5},
		{name: "Unknown", level: "X", want: -1},
		{name: "Empty", level: "", want: -1},
		{name: "Lowercase", level: "v", want: -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LevelIndex(tt.level)
			if got != tt.want {
				t.Errorf("LevelIndex(%q) = %d, want %d", tt.level, got, tt.want)
			}
		})
	}
}
