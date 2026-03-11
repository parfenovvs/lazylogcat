package model

import (
	"encoding/json"
	"regexp"
	"strings"
)

type Level string

const (
	LvlV Level = "V"
	LvlD Level = "D"
	LvlI Level = "I"
	LvlW Level = "W"
	LvlE Level = "E"
	LvlF Level = "F"
)

const lvls = "VDIWEF"

// TextFilterMode determines how a text filter value is matched.
type TextFilterMode int

const (
	FilterModeContains TextFilterMode = iota
	FilterModeExact
	FilterModeRegex
)

// Next cycles to the next filter mode: contains -> exact -> regex -> contains.
func (m TextFilterMode) Next() TextFilterMode {
	return (m + 1) % 3
}

// String returns the human-readable name of the filter mode.
func (m TextFilterMode) String() string {
	switch m {
	case FilterModeExact:
		return "exact"
	case FilterModeRegex:
		return "regex"
	default:
		return "contains"
	}
}

// TextFilter holds a text filter value together with its matching mode.
type TextFilter struct {
	Value      string         `json:"value"`
	Mode       TextFilterMode `json:"mode"`
	compiled   *regexp.Regexp
	compileErr error
}

// MarshalJSON outputs the short form (just the value string) when Mode is
// the default (FilterModeContains). When a non-default mode is set, the
// full object form is emitted.
func (tf TextFilter) MarshalJSON() ([]byte, error) {
	if tf.Mode == FilterModeContains {
		return json.Marshal(tf.Value)
	}
	return json.Marshal(struct {
		Value string         `json:"value"`
		Mode  TextFilterMode `json:"mode"`
	}{
		Value: tf.Value,
		Mode:  tf.Mode,
	})
}

// UnmarshalJSON supports both a plain string and an object form.
// Plain string: "hello" -> TextFilter{Value: "hello", Mode: FilterModeContains}
// Object form: {"value": "hello", "mode": 1} -> TextFilter{Value: "hello", Mode: FilterModeExact}
func (tf *TextFilter) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		tf.Value = s
		tf.Mode = FilterModeContains
		return nil
	}
	type alias TextFilter
	return json.Unmarshal(data, (*alias)(tf))
}

// IsEmpty returns true if the filter has no meaningful value set.
func (tf TextFilter) IsEmpty() bool {
	return tf.Value == ""
}

// Compile pre-compiles the regex pattern when Mode is FilterModeRegex.
// Call this once after creating or changing the filter to avoid recompiling
// on every Match() call. For non-regex modes this is a no-op.
func (tf *TextFilter) Compile() {
	if tf.Mode != FilterModeRegex {
		tf.compiled = nil
		tf.compileErr = nil
		return
	}
	tf.compiled, tf.compileErr = regexp.Compile("(?i)" + tf.Value)
}

// FindAllMatchIndexes returns all non-overlapping [start, end) byte-index
// pairs of matches within haystack. Returns nil when the filter is empty,
// in exact mode, or has no matches. The indices refer to the original
// (not lowered) haystack bytes.
func (tf *TextFilter) FindAllMatchIndexes(haystack string) [][]int {
	if tf.Value == "" || tf.Mode == FilterModeExact {
		return nil
	}

	switch tf.Mode {
	case FilterModeRegex:
		re := tf.compiled
		if re == nil {
			var err error
			re, err = regexp.Compile("(?i)" + tf.Value)
			if err != nil {
				return nil
			}
		} else if tf.compileErr != nil {
			return nil
		}
		return re.FindAllStringIndex(haystack, -1)

	default: // FilterModeContains
		needle := strings.ToLower(tf.Value)
		lower := strings.ToLower(haystack)
		nLen := len(needle)
		var matches [][]int
		start := 0
		for {
			idx := strings.Index(lower[start:], needle)
			if idx < 0 {
				break
			}
			absStart := start + idx
			matches = append(matches, []int{absStart, absStart + nLen})
			start = absStart + nLen
		}
		return matches
	}
}

// Match reports whether haystack matches this filter's value according
// to the current Mode.
//
//   - FilterModeContains: case-insensitive substring match
//   - FilterModeExact:    case-insensitive exact match (strings.EqualFold)
//   - FilterModeRegex:    regexp match with auto (?i); invalid pattern → false
//
// If the filter is empty, Match returns true (no filtering).
func (tf *TextFilter) Match(haystack string) bool {
	if tf.Value == "" {
		return true
	}

	switch tf.Mode {
	case FilterModeExact:
		return strings.EqualFold(haystack, tf.Value)
	case FilterModeRegex:
		re := tf.compiled
		if re == nil {
			// Not pre-compiled; compile on the fly.
			var err error
			re, err = regexp.Compile("(?i)" + tf.Value)
			if err != nil {
				return false // invalid regex → nothing matches
			}
		} else if tf.compileErr != nil {
			return false // pre-compilation failed → nothing matches
		}
		return re.MatchString(haystack)
	default: // FilterModeContains
		return strings.Contains(strings.ToLower(haystack), strings.ToLower(tf.Value))
	}
}

type Filter struct {
	PackageName TextFilter `json:"packageName"`
	Level       Level      `json:"level"`
	Tag         TextFilter `json:"tag"`
	Text        TextFilter `json:"text"`
}

func (f *Filter) IsEmpty() bool {
	return f.PackageName.IsEmpty() &&
		(f.Level == "" || f.Level == LvlV) &&
		f.Tag.IsEmpty() &&
		f.Text.IsEmpty()
}

func (l Level) Next() Level {
	if l == LvlF {
		return LvlV
	}

	i := strings.Index(lvls, string(l))
	if i == -1 {
		return LvlV
	}
	return Level(lvls[i+1])
}

// LevelIndex returns the numeric severity index of a log level string.
// V=0, D=1, I=2, W=3, E=4, F=5. Returns -1 if the level is unknown or empty.
func LevelIndex(l string) int {
	if len(l) != 1 {
		return -1
	}
	return strings.Index(lvls, l)
}
