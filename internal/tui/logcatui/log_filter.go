package logcatui

import (
	"strings"

	"github.com/parfenovvs/lazylogcat/internal/model"
)

// matchesFilter reports whether a raw log line (and its parsed form)
// passes all active filters. It returns false when the line should be
// skipped.
func matchesFilter(raw string, line model.LogLine, filter model.Filter, pidSet map[string]struct{}) bool {
	// Filter empty lines (threadtime format never produces meaningful empty lines)
	if strings.Trim(raw, "\n\r ") == "" {
		return false
	}

	// Filter by text search (case-insensitive) on the raw line before parsing
	if filter.Text != "" && !strings.Contains(strings.ToLower(raw), strings.ToLower(filter.Text)) {
		return false
	}

	// Structured filters only apply to successfully parsed lines.
	// Unparsed lines (e.g. "--------- beginning of main") are skipped
	// when any structured filter is active.
	if line.Parsed() {
		// Filter by PID (package name resolved to PIDs)
		if len(pidSet) > 0 {
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

		// Filter by tag (case-insensitive contains)
		if filter.Tag != "" && !strings.Contains(strings.ToLower(line.Tag), strings.ToLower(filter.Tag)) {
			return false
		}
	} else if len(pidSet) > 0 || filter.Tag != "" || (filter.Level != "" && filter.Level != model.LvlV) {
		// Skip unparsed lines when any structured filter is active
		return false
	}

	return true
}
