package logcatui

import (
	"fmt"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"

	"github.com/parfenovvs/lazylogcat/internal/model"
	"github.com/parfenovvs/lazylogcat/internal/tui/theme"
)

func testOutputPrefs() model.OutputPrefs {
	return model.OutputPrefs{
		Color:    false,
		SoftWrap: true,
		Columns: model.Columns{
			Time:    true,
			Level:   true,
			Tag:     true,
			Message: true,
		},
	}
}

func appendLogLines(m *LogcatViewModel, count int) {
	for i := range count {
		raw := fmt.Sprintf("05-01 18:15:%02d.000  1234  5678 I TestTag: message %03d with enough text to wrap in a narrow viewport", i%60, i)
		m.log.Append(model.ParseLogLine(raw))
	}
}

func TestVisualNavigationDoesNotRebuildViewportContent(t *testing.T) {
	m := New(model.Size{Width: 42, Height: 8}, nil, "", model.Filter{}, testOutputPrefs())
	appendLogLines(&m, 25)
	m.Render()
	m.visualMode = true
	m.currentLine = 20
	m.viewport.SetYOffset(m.logLineVisualStart[m.currentLine])

	beforeContent := m.viewport.GetContent()
	beforeStyledView := m.viewportView()

	result := m.handleVisualModeKey("k")
	if result.needsRender {
		t.Fatal("visual navigation requested full render")
	}
	if m.currentLine != 19 {
		t.Fatalf("currentLine = %d, want 19", m.currentLine)
	}
	if got := m.viewport.GetContent(); got != beforeContent {
		t.Fatal("visual navigation rebuilt viewport content")
	}
	if got := m.viewportView(); got == beforeStyledView {
		t.Fatal("visual navigation did not update selected line styling")
	}
}

func TestVisualSelectionStyleUsesWrappedLineMap(t *testing.T) {
	m := New(model.Size{Width: 36, Height: 8}, nil, "", model.Filter{}, testOutputPrefs())
	appendLogLines(&m, 3)
	m.Render()

	if len(m.logLineVisualStart) != 3 {
		t.Fatalf("cached line starts = %d, want 3", len(m.logLineVisualStart))
	}
	if m.logLineVisualStart[1] <= m.logLineVisualStart[0] {
		t.Fatalf("line 1 start = %d, want after line 0 start %d", m.logLineVisualStart[1], m.logLineVisualStart[0])
	}

	m.visualMode = true
	m.currentLine = 1
	start := m.logLineVisualStart[1]
	end := len(m.visualLineToLogLine) - 1
	if len(m.logLineVisualStart) > 2 {
		end = m.logLineVisualStart[2] - 1
	}

	for visualLine := start; visualLine <= end; visualLine++ {
		rendered := m.visualLineStyle(visualLine).Render("selected")
		if !strings.Contains(rendered, "\x1b[") {
			t.Fatalf("visual line %d for selected log line was not styled", visualLine)
		}
	}
}

func TestVisualSelectionForegroundOverridesLogLevelColor(t *testing.T) {
	prefs := testOutputPrefs()
	prefs.Color = true
	m := New(model.Size{Width: 42, Height: 8}, nil, "", model.Filter{}, prefs)
	appendLogLines(&m, 1)
	m.Render()

	if strings.Contains(m.viewport.GetContent(), "\x1b[") {
		t.Fatal("viewport content contains ANSI styling")
	}
	if got := m.viewportView(); !strings.Contains(got, "\x1b[") {
		t.Fatal("viewport view did not apply log level styling")
	}

	m.visualMode = true
	m.currentLine = 0

	got := m.visualLineStyle(0).Render("selected")
	want := lipgloss.NewStyle().
		Background(theme.ColorVisualBG).
		Foreground(theme.ColorVisualFG).
		Width(m.viewport.Width()).
		Render("selected")
	if got != want {
		t.Fatalf("selected visual style did not override log level style\n got: %q\nwant: %q", got, want)
	}
}

func TestBuildShortcutMap(t *testing.T) {
	m := buildShortcutMap()

	expectedKeys := map[string]model.Command{
		"p": model.CommandPackage,
		"t": model.CommandTag,
		"l": model.CommandLevel,
		"c": model.CommandContent,
		"o": model.CommandOutput,
		"r": model.CommandReconnect,
		"d": model.CommandDevices,
	}

	for key, wantCmd := range expectedKeys {
		t.Run("Key_"+key, func(t *testing.T) {
			data, ok := m[key]
			if !ok {
				t.Fatalf("shortcutMap missing key %q", key)
			}
			if data.Command != wantCmd {
				t.Errorf("shortcutMap[%q].Command = %d, want %d", key, data.Command, wantCmd)
			}
		})
	}

	// Verify no unexpected keys (ctrl+c is "ctrl+c", not "ctrl+x c")
	for key := range m {
		if _, ok := expectedKeys[key]; !ok {
			t.Errorf("unexpected key %q in shortcutMap", key)
		}
	}
}
