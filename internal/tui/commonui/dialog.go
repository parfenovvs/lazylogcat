package commonui

import (
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/parfenovvs/lazylogcat/internal/tui"
	"github.com/parfenovvs/lazylogcat/internal/tui/theme"
)

// DialogTitleWithESC renders a title line with "ESC" right-aligned.
// Used by all dialog View() methods to show the dismiss hint in the title bar.
func DialogTitleWithESC(title string) string {
	titleStr := theme.DialogTitle().Render(title)
	escStr := theme.DialogHelp().Render("ESC")
	// innerWidth = DialogWidth - border(2) - padding(2)
	innerWidth := tui.DialogWidth - 4
	titleWidth := ansi.StringWidth(titleStr)
	rightWidth := innerWidth - titleWidth
	rightPart := lipgloss.NewStyle().
		Width(rightWidth).
		AlignHorizontal(lipgloss.Right).
		Render(escStr)
	return titleStr + rightPart
}

// DialogFrameStyle returns the standard overlay dialog chrome (border, padding, default size).
func DialogFrameStyle() lipgloss.Style {
	return theme.Dialog().
		Width(tui.DialogWidth).
		MaxHeight(tui.DialogMaxHeight)
}
