package util

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/parfenovvs/lazylogcat/internal/model"
)

const (
	initialLogLines = 1000
	scannerBufSize  = 256 * 1024 // 256 KB scanner buffer
	pendingInitCap  = 256        // initial capacity for the pending slice
)

// LogcatReader encapsulates all logcat reading, parsing, and filtering
// in a dedicated goroutine. It accumulates filtered lines internally
// and exposes Drain() to retrieve them in bulk.
type LogcatReader struct {
	mu        sync.Mutex // protects pending, connected, lastErr
	pending   []model.LogLine
	connected bool
	lastErr   error

	filterMu sync.RWMutex // protects filter, pidSet (read by goroutine, written by caller)
	filter   model.Filter
	pidSet   map[string]struct{}

	cmd    *exec.Cmd
	cancel context.CancelFunc
	done   chan struct{} // closed when the reader goroutine exits
}

// NewLogcatReader creates a reader. It does not start reading —
// call Connect() to begin.
func NewLogcatReader() *LogcatReader {
	return &LogcatReader{
		pending: make([]model.LogLine, 0, pendingInitCap),
	}
}

// Connect starts adb logcat for the given device and spawns the reader
// goroutine. If already connected, the previous connection is closed first.
// Returns an error if the device is not reachable or adb fails to start.
func (r *LogcatReader) Connect(deviceId string, filter model.Filter) error {
	r.Disconnect()

	devices, err := GetConnectedDevices()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToStartLogcat, err)
	}
	deviceFound := false
	for _, d := range devices {
		if d.Id == deviceId {
			deviceFound = true
			break
		}
	}
	if !deviceFound {
		return fmt.Errorf("%w: device %s is not connected", ErrFailedToStartLogcat, deviceId)
	}

	args := []string{"-s", deviceId, "logcat", "-T", strconv.Itoa(initialLogLines), "-v", "threadtime"}
	slog.Debug("LogcatReader: executing adb logcat", "args", args)

	cmd := exec.Command("adb", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToGetStdoutPipe, err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToStartLogcat, err)
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, scannerBufSize), scannerBufSize)

	r.filterMu.Lock()
	r.filter = filter
	r.pidSet = nil
	r.filterMu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())

	r.mu.Lock()
	r.cmd = cmd
	r.cancel = cancel
	r.connected = true
	r.lastErr = nil
	r.pending = make([]model.LogLine, 0, pendingInitCap)
	r.done = make(chan struct{})
	r.mu.Unlock()

	go r.readLoop(ctx, scanner)

	return nil
}

// Disconnect stops the reader goroutine and kills the adb process.
// Safe to call multiple times or when not connected.
func (r *LogcatReader) Disconnect() {
	r.mu.Lock()
	cancel := r.cancel
	cmd := r.cmd
	done := r.done
	wasConnected := r.connected
	r.mu.Unlock()

	if !wasConnected {
		return
	}

	// Signal the goroutine to stop
	if cancel != nil {
		cancel()
	}

	// Kill the adb process so the scanner unblocks
	if cmd != nil && cmd.Process != nil {
		slog.Debug("LogcatReader: killing adb logcat process")
		cmd.Process.Kill()
		cmd.Wait()
	}

	// Wait for the goroutine to finish
	if done != nil {
		<-done
	}

	r.mu.Lock()
	r.connected = false
	r.cmd = nil
	r.cancel = nil
	r.done = nil
	r.pending = make([]model.LogLine, 0, pendingInitCap)
	r.mu.Unlock()
}

// UpdateFilter atomically replaces the filter used by the reader goroutine.
// This does NOT reconnect adb — it only affects in-memory filtering of
// incoming lines. Call filter.Text.Compile() etc. before passing it here.
func (r *LogcatReader) UpdateFilter(filter model.Filter) {
	r.filterMu.Lock()
	defer r.filterMu.Unlock()
	r.filter = filter
}

// UpdatePIDSet atomically replaces the PID set used for package-name filtering.
func (r *LogcatReader) UpdatePIDSet(pidSet map[string]struct{}) {
	r.filterMu.Lock()
	defer r.filterMu.Unlock()
	r.pidSet = pidSet
}

// Drain returns all accumulated lines since the last Drain() call and
// clears the internal pending buffer. Returns nil if no lines are pending.
func (r *LogcatReader) Drain() []model.LogLine {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.pending) == 0 {
		return nil
	}
	lines := r.pending
	r.pending = make([]model.LogLine, 0, cap(lines))
	return lines
}

// IsConnected reports whether the reader goroutine is actively reading.
func (r *LogcatReader) IsConnected() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.connected
}

// Err returns the last error that caused a disconnect (io.EOF, scan error, etc.).
// Returns nil if still connected or if disconnect was voluntary.
func (r *LogcatReader) Err() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.lastErr
}

// readLoop is the core goroutine. It reads from the scanner in a tight loop,
// parses each line, applies the current filter, and appends passing lines to
// the pending buffer.
func (r *LogcatReader) readLoop(ctx context.Context, scanner *bufio.Scanner) {
	defer close(r.done)

	for scanner.Scan() {
		// Check if we've been asked to stop
		select {
		case <-ctx.Done():
			return
		default:
		}

		raw := scanner.Text()

		line := model.ParseLogLine(raw)

		r.filterMu.RLock()
		pass := matchesFilter(raw, line, &r.filter, r.pidSet)
		r.filterMu.RUnlock()

		if !pass {
			continue
		}

		r.mu.Lock()
		r.pending = append(r.pending, line)
		r.mu.Unlock()
	}

	// Scanner stopped — EOF or error
	err := scanner.Err()
	r.mu.Lock()
	r.connected = false
	r.lastErr = err
	r.mu.Unlock()
}

// matchesFilter reports whether a raw log line (and its parsed form)
// passes all active filters. Returns false when the line should be skipped.
func matchesFilter(raw string, line model.LogLine, filter *model.Filter, pidSet map[string]struct{}) bool {
	// Filter empty lines (threadtime format never produces meaningful empty lines)
	if strings.Trim(raw, "\n\r ") == "" {
		return false
	}

	// Filter by text search on the raw line before parsing.
	if !filter.Text.IsEmpty() && !filter.Text.Match(raw) {
		return false
	}

	// Structured filters only apply to successfully parsed lines.
	if line.Parsed() {
		// Filter by PID (package name resolved to PIDs).
		if !filter.PackageName.IsEmpty() {
			if _, ok := pidSet[line.PID]; !ok {
				return false
			}
		}

		// Filter by minimum log level
		if filter.Level != "" && filter.Level != model.LvlV {
			if model.LevelIndex(line.Level) < model.LevelIndex(string(filter.Level)) {
				return false
			}
		}

		// Filter by tag
		if !filter.Tag.IsEmpty() && !filter.Tag.Match(line.Tag) {
			return false
		}
	} else if !filter.PackageName.IsEmpty() || !filter.Tag.IsEmpty() || (filter.Level != "" && filter.Level != model.LvlV) {
		// Skip unparsed lines when any structured filter is active
		return false
	}

	return true
}
