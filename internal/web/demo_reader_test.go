package web

import (
	"testing"
	"time"

	"github.com/parfenovvs/lazylogcat/internal/model"
)

func TestNewDemoReader(t *testing.T) {
	r := newDemoReader()
	if r == nil {
		t.Fatal("newDemoReader() returned nil")
	}
	if r.IsConnected() {
		t.Error("new demo reader should not be connected")
	}
	if err := r.Err(); err != nil {
		t.Errorf("new demo reader should have nil error, got %v", err)
	}
}

func TestDemoReaderConnectAndDrain(t *testing.T) {
	r := newDemoReader()
	if err := r.Connect("demo-device", model.Filter{}); err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	defer r.Disconnect()

	if !r.IsConnected() {
		t.Error("reader should be connected after Connect()")
	}

	// Wait for some lines to be generated
	var lines []model.LogLine
	deadline := time.After(2 * time.Second)
	for len(lines) == 0 {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for demo lines")
		default:
			lines = r.Drain()
			if len(lines) == 0 {
				time.Sleep(50 * time.Millisecond)
			}
		}
	}

	if len(lines) == 0 {
		t.Fatal("expected at least one line from demo reader")
	}

	// Verify lines are well-formed (parsed by ParseLogLine)
	for i, l := range lines {
		if !l.Parsed() {
			t.Errorf("line[%d] not parsed: %q", i, l.Raw)
		}
		if l.Tag == "" {
			t.Errorf("line[%d] has empty tag", i)
		}
		if l.PID == "" {
			t.Errorf("line[%d] has empty PID", i)
		}
	}
}

func TestDemoReaderDisconnect(t *testing.T) {
	r := newDemoReader()
	if err := r.Connect("demo-device", model.Filter{}); err != nil {
		t.Fatalf("Connect() error: %v", err)
	}

	r.Disconnect()

	if r.IsConnected() {
		t.Error("reader should not be connected after Disconnect()")
	}

	// Drain should return nil after disconnect
	lines := r.Drain()
	if lines != nil {
		t.Errorf("Drain() after disconnect = %v, want nil", lines)
	}
}

func TestDemoReaderDisconnectIdempotent(t *testing.T) {
	r := newDemoReader()
	// Should not panic
	r.Disconnect()
	r.Disconnect()

	if err := r.Connect("demo-device", model.Filter{}); err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	r.Disconnect()
	r.Disconnect()
}

func TestDemoReaderFilterLevel(t *testing.T) {
	r := newDemoReader()
	filter := model.Filter{Level: model.LvlE}
	filter.Text.Compile()
	filter.Tag.Compile()
	filter.PackageName.Compile()

	if err := r.Connect("demo-device", filter); err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	defer r.Disconnect()

	// Collect lines for a short period
	var allLines []model.LogLine
	deadline := time.After(3 * time.Second)
	for {
		select {
		case <-deadline:
			goto check
		default:
			lines := r.Drain()
			allLines = append(allLines, lines...)
			if len(allLines) >= 5 {
				goto check
			}
			time.Sleep(50 * time.Millisecond)
		}
	}

check:
	if len(allLines) == 0 {
		t.Fatal("expected some lines with level filter E")
	}

	for i, l := range allLines {
		lvl := model.LevelIndex(l.Level)
		minLvl := model.LevelIndex("E")
		if lvl < minLvl {
			t.Errorf("line[%d] level %s is below E threshold", i, l.Level)
		}
	}
}

func TestDemoReaderUpdateFilter(t *testing.T) {
	r := newDemoReader()
	if err := r.Connect("demo-device", model.Filter{}); err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	defer r.Disconnect()

	// Update to a very restrictive filter
	f := model.Filter{Level: model.LvlF}
	f.Text.Compile()
	f.Tag.Compile()
	f.PackageName.Compile()
	r.UpdateFilter(f)

	// The update should not panic or error
	if !r.IsConnected() {
		t.Error("reader should still be connected after UpdateFilter")
	}
}

func TestDemoReaderWaitForDoneNotConnected(t *testing.T) {
	r := newDemoReader()
	done := make(chan struct{})
	go func() {
		r.WaitForDone()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("WaitForDone() should return immediately when not connected")
	}
}

func TestDemoDevice(t *testing.T) {
	if demoDevice.Id == "" {
		t.Error("demoDevice.Id is empty")
	}
	if demoDevice.Name == "" {
		t.Error("demoDevice.Name is empty")
	}
}

func TestPickLevel(t *testing.T) {
	// Ensure pickLevel returns valid levels and doesn't panic
	levels := map[string]bool{"V": true, "D": true, "I": true, "W": true, "E": true, "F": true}
	for range 100 {
		l := pickLevel()
		if !levels[l] {
			t.Errorf("pickLevel() returned invalid level %q", l)
		}
	}
}

func TestFillTemplate(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "NoVerbs", input: "simple message"},
		{name: "IntVerb", input: "count: %d items"},
		{name: "StringVerb", input: "package: %s"},
		{name: "PercentLiteral", input: "battery: %d%%"},
		{name: "Mixed", input: "proc %d:%s running"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fillTemplate(tt.input)
			if result == "" {
				t.Error("fillTemplate returned empty string")
			}
		})
	}
}
