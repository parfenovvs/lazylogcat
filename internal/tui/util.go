package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/parfenovvs/lazylogcat/internal/model"
)

// Dialog size constants used by the command dialog overlay.
const (
	DialogWidth     = 48
	DialogMaxHeight = 28
)

// DimView applies a dimming effect to the view content
func DimView(view string) string {
	dimStyle := lipgloss.NewStyle().Faint(true)

	lines := strings.Split(view, "\n")
	dimmedLines := make([]string, len(lines))

	for i, line := range lines {
		dimmedLines[i] = dimStyle.Render(ansi.Strip(line))
	}

	return strings.Join(dimmedLines, "\n")
}

// OverlayDialog overlays a dialog string on top of a base view, using parentSize for layout.
// The vertical position is calculated based on DialogMaxHeight so that all dialogs share the same
// top position regardless of their actual rendered height.
func OverlayDialog(parentSize model.Size, baseView, dialog string) string {
	background := lipgloss.Place(
		parentSize.Width,
		parentSize.Height,
		lipgloss.Left,
		lipgloss.Top,
		baseView,
	)

	bgLines := strings.Split(background, "\n")

	dialogLines := strings.Split(dialog, "\n")
	dialogHeight := len(dialogLines)
	dialogWidth := 0
	for _, line := range dialogLines {
		w := ansi.StringWidth(line)
		if w > dialogWidth {
			dialogWidth = w
		}
	}

	x := (parentSize.Width - dialogWidth) / 2

	verticalSize := dialogHeight
	if DialogMaxHeight > dialogHeight {
		verticalSize = DialogMaxHeight
	}
	y := (parentSize.Height - verticalSize) / 2

	if y < 0 {
		y = 0
	}
	if x < 0 {
		x = 0
	}

	var result strings.Builder
	for i := 0; i < len(bgLines); i++ {
		dialogLineIdx := i - y
		if dialogLineIdx >= 0 && dialogLineIdx < dialogHeight {
			bgLine := bgLines[i]
			dialogLine := dialogLines[dialogLineIdx]
			overlaidLine := overlayLine(bgLine, dialogLine, x)
			result.WriteString(overlaidLine)
		} else {
			result.WriteString(bgLines[i])
		}

		if i < len(bgLines)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

// overlayLine overlays foreground onto background at position x (in visual character positions)
func overlayLine(background, foreground string, x int) string {
	bgWidth := ansi.StringWidth(background)
	fgWidth := ansi.StringWidth(foreground)

	if x >= bgWidth {
		return background
	}

	var result strings.Builder

	if x > 0 {
		prefix := ansi.Truncate(background, x, "")
		result.WriteString(prefix)
	}

	result.WriteString(foreground)

	endPos := x + fgWidth
	if endPos < bgWidth {
		remaining := bgWidth - endPos
		if remaining > 0 {
			result.WriteString(strings.Repeat(" ", remaining))
		}
	}

	return result.String()
}
