package util

import (
	"bufio"
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/parfenovvs/lazylogcat/internal/model"
)

func TestNewLogcatReader(t *testing.T) {
	r := NewLogcatReader()
	if r == nil {
		t.Fatal("NewLogcatReader() returned nil")
	}
	if r.IsConnected() {
		t.Error("new reader should not be connected")
	}
	if err := r.Err(); err != nil {
		t.Errorf("new reader should have nil error, got %v", err)
	}
}

func TestDrainEmpty(t *testing.T) {
	r := NewLogcatReader()
	lines := r.Drain()
	if lines != nil {
		t.Errorf("Drain() on fresh reader = %v, want nil", lines)
	}
}

func TestDrainReturnsAccumulated(t *testing.T) {
	r := NewLogcatReader()
	r.pending = append(r.pending,
		model.LogLine{Raw: "line1"},
		model.LogLine{Raw: "line2"},
		model.LogLine{Raw: "line3"},
	)

	lines := r.Drain()
	if len(lines) != 3 {
		t.Fatalf("Drain() returned %d lines, want 3", len(lines))
	}
	if lines[0].Raw != "line1" || lines[1].Raw != "line2" || lines[2].Raw != "line3" {
		t.Errorf("Drain() returned unexpected lines: %v", lines)
	}
}

func TestDrainClearsPending(t *testing.T) {
	r := NewLogcatReader()
	r.pending = append(r.pending, model.LogLine{Raw: "line1"})

	_ = r.Drain()
	second := r.Drain()
	if second != nil {
		t.Errorf("second Drain() = %v, want nil", second)
	}
}

func TestDrainPreservesCapacity(t *testing.T) {
	r := NewLogcatReader()
	// Fill with enough lines to grow beyond initial capacity
	for i := range 500 {
		r.pending = append(r.pending, model.LogLine{Raw: strings.Repeat("x", i)})
	}
	oldCap := cap(r.pending)

	_ = r.Drain()
	// New pending slice should reuse the grown capacity hint
	newCap := cap(r.pending)
	if newCap != oldCap {
		t.Errorf("Drain() new cap = %d, want %d (preserved from old slice)", newCap, oldCap)
	}
}

func TestDisconnectIdempotent(t *testing.T) {
	r := NewLogcatReader()
	// Should not panic when called on a reader that was never connected
	r.Disconnect()
	r.Disconnect()
	r.Disconnect()
}

func TestUpdateFilter(t *testing.T) {
	r := NewLogcatReader()
	f := model.Filter{
		Level: model.LvlW,
		Tag:   model.TextFilter{Value: "MyTag"},
	}
	r.UpdateFilter(f)

	r.filterMu.RLock()
	defer r.filterMu.RUnlock()
	if r.filter.Level != model.LvlW {
		t.Errorf("filter.Level = %q, want %q", r.filter.Level, model.LvlW)
	}
	if r.filter.Tag.Value != "MyTag" {
		t.Errorf("filter.Tag.Value = %q, want %q", r.filter.Tag.Value, "MyTag")
	}
}

func TestUpdatePIDSet(t *testing.T) {
	r := NewLogcatReader()
	pids := map[string]struct{}{"1234": {}, "5678": {}}
	r.UpdatePIDSet(pids)

	r.filterMu.RLock()
	defer r.filterMu.RUnlock()
	if len(r.pidSet) != 2 {
		t.Errorf("pidSet has %d entries, want 2", len(r.pidSet))
	}
	if _, ok := r.pidSet["1234"]; !ok {
		t.Error("pidSet missing PID 1234")
	}
}

func TestUpdatePIDSetNil(t *testing.T) {
	r := NewLogcatReader()
	r.UpdatePIDSet(map[string]struct{}{"1": {}})
	r.UpdatePIDSet(nil)

	r.filterMu.RLock()
	defer r.filterMu.RUnlock()
	if r.pidSet != nil {
		t.Errorf("pidSet = %v, want nil", r.pidSet)
	}
}

// readLoopWithScanner is a test helper that runs readLoop with a synthetic scanner
// so we can test the goroutine logic without adb.
func startTestReader(t *testing.T, input string, filter model.Filter, pidSet map[string]struct{}) *LogcatReader {
	t.Helper()
	r := NewLogcatReader()
	r.filter = filter
	r.pidSet = pidSet
	r.connected = true
	r.done = make(chan struct{})

	scanner := bufio.NewScanner(strings.NewReader(input))
	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel
	go r.readLoop(ctx, scanner)
	return r
}

func waitForDone(t *testing.T, r *LogcatReader) {
	t.Helper()
	select {
	case <-r.done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for readLoop to finish")
	}
}

func TestReadLoopBasic(t *testing.T) {
	input := "02-08 12:00:00.000  1000  1001 D MyTag: hello world\n" +
		"02-08 12:00:01.000  1000  1001 I MyTag: second line\n"

	r := startTestReader(t, input, model.Filter{}, nil)
	waitForDone(t, r)

	lines := r.Drain()
	if len(lines) != 2 {
		t.Fatalf("Drain() returned %d lines, want 2", len(lines))
	}
	if lines[0].Tag != "MyTag" || lines[0].Message != "hello world" {
		t.Errorf("line[0] = {Tag:%q, Message:%q}, want {Tag:MyTag, Message:hello world}",
			lines[0].Tag, lines[0].Message)
	}
	if lines[1].Level != "I" {
		t.Errorf("line[1].Level = %q, want I", lines[1].Level)
	}
}

func TestReadLoopFiltersEmptyLines(t *testing.T) {
	input := "\n\n02-08 12:00:00.000  1000  1001 D Tag: msg\n  \n"

	r := startTestReader(t, input, model.Filter{}, nil)
	waitForDone(t, r)

	lines := r.Drain()
	if len(lines) != 1 {
		t.Fatalf("Drain() returned %d lines, want 1", len(lines))
	}
}

func TestReadLoopFiltersLevel(t *testing.T) {
	input := "02-08 12:00:00.000  1000  1001 D Tag: debug msg\n" +
		"02-08 12:00:01.000  1000  1001 W Tag: warning msg\n" +
		"02-08 12:00:02.000  1000  1001 E Tag: error msg\n"

	filter := model.Filter{Level: model.LvlW}
	r := startTestReader(t, input, filter, nil)
	waitForDone(t, r)

	lines := r.Drain()
	if len(lines) != 2 {
		t.Fatalf("Drain() returned %d lines, want 2 (W and E)", len(lines))
	}
	if lines[0].Level != "W" {
		t.Errorf("lines[0].Level = %q, want W", lines[0].Level)
	}
	if lines[1].Level != "E" {
		t.Errorf("lines[1].Level = %q, want E", lines[1].Level)
	}
}

func TestReadLoopFiltersTag(t *testing.T) {
	input := "02-08 12:00:00.000  1000  1001 D MyTag: keep\n" +
		"02-08 12:00:01.000  1000  1001 D OtherTag: drop\n"

	filter := model.Filter{
		Tag: model.TextFilter{Value: "MyTag", Mode: model.FilterModeExact},
	}
	r := startTestReader(t, input, filter, nil)
	waitForDone(t, r)

	lines := r.Drain()
	if len(lines) != 1 {
		t.Fatalf("Drain() returned %d lines, want 1", len(lines))
	}
	if lines[0].Tag != "MyTag" {
		t.Errorf("lines[0].Tag = %q, want MyTag", lines[0].Tag)
	}
}

func TestReadLoopFiltersText(t *testing.T) {
	input := "02-08 12:00:00.000  1000  1001 D Tag: match this needle\n" +
		"02-08 12:00:01.000  1000  1001 D Tag: nothing here\n"

	filter := model.Filter{
		Text: model.TextFilter{Value: "needle", Mode: model.FilterModeContains},
	}
	r := startTestReader(t, input, filter, nil)
	waitForDone(t, r)

	lines := r.Drain()
	if len(lines) != 1 {
		t.Fatalf("Drain() returned %d lines, want 1", len(lines))
	}
	if !strings.Contains(lines[0].Raw, "needle") {
		t.Errorf("expected line containing 'needle', got %q", lines[0].Raw)
	}
}

func TestReadLoopFiltersPID(t *testing.T) {
	input := "02-08 12:00:00.000  1000  1001 D Tag: from app\n" +
		"02-08 12:00:01.000  2000  2001 D Tag: from other\n"

	filter := model.Filter{
		PackageName: model.TextFilter{Value: "com.example"},
	}
	pidSet := map[string]struct{}{"1000": {}}
	r := startTestReader(t, input, filter, pidSet)
	waitForDone(t, r)

	lines := r.Drain()
	if len(lines) != 1 {
		t.Fatalf("Drain() returned %d lines, want 1", len(lines))
	}
	if lines[0].PID != "1000" {
		t.Errorf("lines[0].PID = %q, want 1000", lines[0].PID)
	}
}

func TestReadLoopUnparsedLinesPassWithNoStructuredFilter(t *testing.T) {
	input := "--------- beginning of main\n" +
		"02-08 12:00:00.000  1000  1001 D Tag: parsed\n"

	r := startTestReader(t, input, model.Filter{}, nil)
	waitForDone(t, r)

	lines := r.Drain()
	if len(lines) != 2 {
		t.Fatalf("Drain() returned %d lines, want 2", len(lines))
	}
}

func TestReadLoopUnparsedLinesDroppedWithStructuredFilter(t *testing.T) {
	input := "--------- beginning of main\n" +
		"02-08 12:00:00.000  1000  1001 W Tag: parsed\n"

	filter := model.Filter{Level: model.LvlW}
	r := startTestReader(t, input, filter, nil)
	waitForDone(t, r)

	lines := r.Drain()
	if len(lines) != 1 {
		t.Fatalf("Drain() returned %d lines, want 1", len(lines))
	}
}

func TestReadLoopCancel(t *testing.T) {
	// Scanner from empty input — Scan() returns false immediately
	r := NewLogcatReader()
	r.connected = true
	r.done = make(chan struct{})

	// Create a scanner from an empty reader — it will return false on first Scan()
	scanner := bufio.NewScanner(strings.NewReader(""))
	ctx, cancel := context.WithCancel(context.Background())
	go r.readLoop(ctx, scanner)
	cancel()

	waitForDone(t, r)

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.connected {
		t.Error("reader should be disconnected after readLoop exits")
	}
}

func TestReadLoopSetsDisconnectedOnEOF(t *testing.T) {
	input := "02-08 12:00:00.000  1000  1001 D Tag: one\n"

	r := startTestReader(t, input, model.Filter{}, nil)
	waitForDone(t, r)

	if r.IsConnected() {
		t.Error("reader should not be connected after EOF")
	}
	// io.EOF is not reported as scanner.Err() — scanner returns nil on clean EOF
	if err := r.Err(); err != nil {
		t.Errorf("Err() = %v, want nil (clean EOF)", err)
	}
}

func TestReadLoopConcurrentDrain(t *testing.T) {
	// Generate a large input to stress-test concurrent access
	var b strings.Builder
	for i := range 1000 {
		b.WriteString("02-08 12:00:00.000  1000  1001 D Tag: line ")
		b.WriteString(strings.Repeat("x", i%50))
		b.WriteString("\n")
	}
	input := b.String()

	r := startTestReader(t, input, model.Filter{}, nil)

	// Drain concurrently while readLoop is running
	var totalLines int
	var drainWg sync.WaitGroup
	drainWg.Add(1)
	go func() {
		defer drainWg.Done()
		for {
			lines := r.Drain()
			totalLines += len(lines)

			r.mu.Lock()
			done := !r.connected && len(r.pending) == 0
			r.mu.Unlock()

			if done {
				// One final drain to catch any stragglers
				lines = r.Drain()
				totalLines += len(lines)
				return
			}
			time.Sleep(1 * time.Millisecond)
		}
	}()

	waitForDone(t, r)
	drainWg.Wait()

	if totalLines != 1000 {
		t.Errorf("drained %d total lines, want 1000", totalLines)
	}
}

func TestMatchesFilterEmptyLine(t *testing.T) {
	f := model.Filter{}
	if matchesFilter("", model.LogLine{}, &f, nil) {
		t.Error("empty line should not pass filter")
	}
	if matchesFilter("  \n\r ", model.LogLine{}, &f, nil) {
		t.Error("whitespace-only line should not pass filter")
	}
}

func TestMatchesFilterNoFilter(t *testing.T) {
	f := model.Filter{}
	raw := "02-08 12:00:00.000  1000  1001 D Tag: message"
	line := model.ParseLogLine(raw)
	if !matchesFilter(raw, line, &f, nil) {
		t.Error("line should pass empty filter")
	}
}

func TestMatchesFilterTextContains(t *testing.T) {
	f := model.Filter{
		Text: model.TextFilter{Value: "hello", Mode: model.FilterModeContains},
	}
	tests := []struct {
		name string
		raw  string
		want bool
	}{
		{name: "Match", raw: "02-08 12:00:00.000  1000  1001 D Tag: hello world", want: true},
		{name: "CaseInsensitive", raw: "02-08 12:00:00.000  1000  1001 D Tag: HELLO world", want: true},
		{name: "NoMatch", raw: "02-08 12:00:00.000  1000  1001 D Tag: goodbye world", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line := model.ParseLogLine(tt.raw)
			got := matchesFilter(tt.raw, line, &f, nil)
			if got != tt.want {
				t.Errorf("matchesFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchesFilterLevelThreshold(t *testing.T) {
	tests := []struct {
		name      string
		lineLevel string
		filterLvl model.Level
		want      bool
	}{
		{name: "DebugPassesVerbose", lineLevel: "D", filterLvl: model.LvlV, want: true},
		{name: "DebugFailsWarning", lineLevel: "D", filterLvl: model.LvlW, want: false},
		{name: "ErrorPassesWarning", lineLevel: "E", filterLvl: model.LvlW, want: true},
		{name: "WarningPassesWarning", lineLevel: "W", filterLvl: model.LvlW, want: true},
		{name: "NoFilterLevel", lineLevel: "D", filterLvl: "", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := "02-08 12:00:00.000  1000  1001 " + tt.lineLevel + " Tag: message"
			line := model.ParseLogLine(raw)
			f := model.Filter{Level: tt.filterLvl}
			got := matchesFilter(raw, line, &f, nil)
			if got != tt.want {
				t.Errorf("matchesFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchesFilterTag(t *testing.T) {
	tests := []struct {
		name     string
		tagValue string
		mode     model.TextFilterMode
		lineTag  string
		want     bool
	}{
		{name: "ExactMatch", tagValue: "MyTag", mode: model.FilterModeExact, lineTag: "MyTag", want: true},
		{name: "ExactNoMatch", tagValue: "MyTag", mode: model.FilterModeExact, lineTag: "Other", want: false},
		{name: "ContainsMatch", tagValue: "My", mode: model.FilterModeContains, lineTag: "MyTag", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := "02-08 12:00:00.000  1000  1001 D " + tt.lineTag + ": message"
			line := model.ParseLogLine(raw)
			f := model.Filter{
				Tag: model.TextFilter{Value: tt.tagValue, Mode: tt.mode},
			}
			got := matchesFilter(raw, line, &f, nil)
			if got != tt.want {
				t.Errorf("matchesFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchesFilterPIDSet(t *testing.T) {
	raw := "02-08 12:00:00.000  1000  1001 D Tag: message"
	line := model.ParseLogLine(raw)

	t.Run("InSet", func(t *testing.T) {
		f := model.Filter{PackageName: model.TextFilter{Value: "com.app"}}
		pids := map[string]struct{}{"1000": {}}
		if !matchesFilter(raw, line, &f, pids) {
			t.Error("line with PID in set should pass")
		}
	})

	t.Run("NotInSet", func(t *testing.T) {
		f := model.Filter{PackageName: model.TextFilter{Value: "com.app"}}
		pids := map[string]struct{}{"9999": {}}
		if matchesFilter(raw, line, &f, pids) {
			t.Error("line with PID not in set should not pass")
		}
	})

	t.Run("EmptySet", func(t *testing.T) {
		f := model.Filter{PackageName: model.TextFilter{Value: "com.app"}}
		pids := map[string]struct{}{}
		if matchesFilter(raw, line, &f, pids) {
			t.Error("line should not pass with empty PID set when package filter is active")
		}
	})
}

func TestWaitForDoneNotConnected(t *testing.T) {
	r := NewLogcatReader()
	// WaitForDone should return immediately when never connected
	done := make(chan struct{})
	go func() {
		r.WaitForDone()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("WaitForDone() did not return for disconnected reader")
	}
}

func TestWaitForDoneBlocksUntilEOF(t *testing.T) {
	input := "02-08 12:00:00.000  1000  1001 D Tag: msg\n"
	r := startTestReader(t, input, model.Filter{}, nil)

	done := make(chan struct{})
	go func() {
		r.WaitForDone()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("WaitForDone() did not return after readLoop exited")
	}

	if r.IsConnected() {
		t.Error("reader should be disconnected after WaitForDone returns")
	}
}

func TestPendingCap(t *testing.T) {
	// Generate input larger than maxPendingLines
	var b strings.Builder
	total := maxPendingLines + 500
	for range total {
		b.WriteString("02-08 12:00:00.000  1000  1001 D Tag: msg\n")
	}

	r := startTestReader(t, b.String(), model.Filter{}, nil)
	waitForDone(t, r)

	lines := r.Drain()
	if len(lines) > maxPendingLines {
		t.Errorf("pending exceeded cap: got %d, max %d", len(lines), maxPendingLines)
	}
	if len(lines) != maxPendingLines {
		t.Errorf("expected exactly %d lines (cap reached), got %d", maxPendingLines, len(lines))
	}
}
