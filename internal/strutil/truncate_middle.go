package strutil

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// TruncateMiddle shortens s to about maxWidth terminal columns using an ellipsis in the middle.
func TruncateMiddle(s string, maxWidth int) string {
	w := ansi.StringWidth(s)
	if w <= maxWidth {
		return s
	}
	left := (maxWidth - 1) / 2
	right := maxWidth - 1 - left

	runes := []rune(s)
	var suffix string
	suffixW := 0
	for i := len(runes) - 1; i >= 0 && suffixW < right; i-- {
		suffixW++
		suffix = string(runes[i]) + suffix
	}

	return ansi.Truncate(s, left, "") + "…" + suffix
}

// PadANSIWidth pads s with ASCII spaces so ansi.StringWidth(s) is at least targetWidth.
// If s is already wider, it returns s unchanged.
func PadANSIWidth(s string, targetWidth int) string {
	if targetWidth <= 0 {
		return s
	}
	sw := ansi.StringWidth(s)
	if sw >= targetWidth {
		return s
	}
	return s + strings.Repeat(" ", targetWidth-sw)
}
