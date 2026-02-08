package model

import "strings"

// LogLine represents a parsed logcat log line in the threadtime format.
//
// Threadtime format:
//
//	02-08 12:12:09.629  3950  4005 D BusinessScope: Enqueuing the block
//	<date> <time> <pid> <tid> <level> <tag>: <message>
//
// If the line cannot be parsed (e.g. separator lines like "--------- beginning of main"),
// only Raw is populated and the parsed fields remain empty.
type LogLine struct {
	Date    string
	Time    string
	PID     string
	TID     string
	Level   string
	Tag     string
	Message string
	Raw     string
}

// ParseLogLine parses a raw logcat line in the threadtime format into a LogLine.
// If parsing fails, the returned LogLine has only Raw set.
//
// The tag and colon separator appear in two forms:
//
//	"Tag: message"   — colon attached to the tag (most common)
//	"Tag : message"  — colon separated from the tag by a space
func ParseLogLine(raw string) LogLine {
	ll := LogLine{Raw: raw}

	// Threadtime format has at least 6 whitespace-delimited fields:
	//   date time pid tid level tag[:]  [:]  message...
	parts := strings.Fields(raw)
	if len(parts) < 6 {
		return ll
	}

	// Validate level field: must be a single character from the known set
	level := parts[4]
	if len(level) != 1 || !strings.ContainsRune("VDIWEF", rune(level[0])) {
		return ll
	}

	// Determine how the tag and colon are arranged.
	//   Case 1: "Tag:" — parts[5] ends with ":"
	//   Case 2: "Tag" ":" — parts[5] is the tag, parts[6] is ":"
	var tag string
	var colonTokens int // how many tokens the tag+colon span (1 or 2)
	tagField := parts[5]
	if strings.HasSuffix(tagField, ":") {
		tag = strings.TrimSuffix(tagField, ":")
		colonTokens = 6
	} else if len(parts) > 6 && parts[6] == ":" {
		tag = tagField
		colonTokens = 7
	} else {
		return ll
	}

	// Message is everything after the colon token in the original raw string.
	colonEnd := findTokenEnd(raw, parts, colonTokens)
	var message string
	if colonEnd < len(raw) {
		// Skip the single space after the colon if present
		rest := raw[colonEnd:]
		if len(rest) > 0 && rest[0] == ' ' {
			rest = rest[1:]
		}
		message = rest
	}

	ll.Date = parts[0]
	ll.Time = parts[1]
	ll.PID = parts[2]
	ll.TID = parts[3]
	ll.Level = level
	ll.Tag = tag
	ll.Message = message

	return ll
}

// findTokenEnd returns the index in raw right after the n-th whitespace-delimited token.
func findTokenEnd(raw string, parts []string, n int) int {
	pos := 0
	for i := 0; i < n; i++ {
		// Skip whitespace
		for pos < len(raw) && raw[pos] == ' ' {
			pos++
		}
		// Skip the token
		pos += len(parts[i])
	}
	return pos
}

// String returns the original raw line. This guarantees the output is identical
// to what adb logcat produced, preserving all original spacing and formatting.
func (l LogLine) String() string {
	return l.Raw
}

// Parsed returns true if the line was successfully parsed into structured fields.
func (l LogLine) Parsed() bool {
	return l.Level != ""
}
