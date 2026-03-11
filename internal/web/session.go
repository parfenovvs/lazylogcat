package web

import (
	"log/slog"
	"sync"

	"github.com/parfenovvs/lazylogcat/internal/config"
	"github.com/parfenovvs/lazylogcat/internal/model"
	"github.com/parfenovvs/lazylogcat/internal/util"
)

const defaultBufferCapacity = 10000

// Session holds per-WebSocket connection state: a logcat reader, a ring
// buffer for replay, and the current filter. Each browser tab gets its
// own Session.
type Session struct {
	reader *util.LogcatReader
	buffer *util.RingBuffer
	filter model.Filter
	config config.Config
	mu     sync.Mutex
}

// NewSession creates a session with the given config.
func NewSession(cfg config.Config) *Session {
	return &Session{
		reader: util.NewLogcatReader(),
		buffer: util.NewRingBuffer(defaultBufferCapacity),
		config: cfg,
	}
}

// Connect starts logcat for the given device. If already connected, the
// previous connection is disconnected first.
func (s *Session) Connect(deviceID string, filter model.Filter) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Compile text filter regexes
	filter.PackageName.Compile()
	filter.Tag.Compile()
	filter.Text.Compile()

	s.filter = filter
	s.buffer.Clear()

	if err := s.reader.Connect(deviceID, filter); err != nil {
		return err
	}

	// If package name filter is active, resolve PIDs
	if !filter.PackageName.IsEmpty() {
		processes, err := util.GetProcessList(deviceID)
		if err != nil {
			slog.Warn("Failed to get process list for PID resolution", "error", err)
		} else {
			pidSet := util.ResolvePIDs(processes, &filter.PackageName)
			s.reader.UpdatePIDSet(pidSet)
		}
	}

	return nil
}

// Disconnect stops the logcat reader.
func (s *Session) Disconnect() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reader.Disconnect()
}

// Drain returns all pending lines from the reader and appends them to the
// ring buffer. Returns nil if no lines are pending.
func (s *Session) Drain() []model.LogLine {
	lines := s.reader.Drain()
	for _, l := range lines {
		s.buffer.Append(l)
	}
	return lines
}

// UpdateFilter hot-swaps the filter on the running reader.
func (s *Session) UpdateFilter(filter model.Filter) {
	s.mu.Lock()
	defer s.mu.Unlock()

	filter.PackageName.Compile()
	filter.Tag.Compile()
	filter.Text.Compile()

	s.filter = filter
	s.reader.UpdateFilter(filter)
}

// BufferedLines returns all lines currently in the ring buffer.
func (s *Session) BufferedLines() []model.LogLine {
	return s.buffer.All()
}

// IsConnected reports whether the reader is actively streaming.
func (s *Session) IsConnected() bool {
	return s.reader.IsConnected()
}

// WaitForDone blocks until the reader goroutine exits.
func (s *Session) WaitForDone() {
	s.reader.WaitForDone()
}

// Err returns the last reader error.
func (s *Session) Err() error {
	return s.reader.Err()
}

// Close disconnects the reader and releases resources.
func (s *Session) Close() {
	s.Disconnect()
}
