package commandui

import (
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	"github.com/charmbracelet/x/ansi"

	"github.com/parfenovvs/lazylogcat/internal/tui"
	"github.com/parfenovvs/lazylogcat/internal/tui/theme"
)

func newTable(columns []table.Column, rows []table.Row, height int) table.Model {
	km := table.DefaultKeyMap()
	km.GotoTop.SetEnabled(false)
	km.GotoBottom.SetEnabled(false)
	km.HalfPageUp.SetEnabled(false)
	km.HalfPageDown.SetEnabled(false)
	km.PageDown.SetEnabled(false)

	s := table.Styles{
		Header:   theme.TableHeader(),
		Cell:     theme.TableCell(),
		Selected: theme.TableSelected(),
	}

	// Bubbles v2 table uses an inner viewport; width must be set or View() is empty.
	innerW := tui.DialogWidth - 4 // dialog border + horizontal padding
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithWidth(innerW),
		table.WithHeight(height),
		table.WithFocused(true),
		table.WithKeyMap(km),
	)
	t.SetStyles(s)

	return t
}

func truncateMiddle(s string, maxWidth int) string {
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

func newSearchInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "Search"
	ti.CharLimit = 50
	ti.SetWidth(35)
	ti.Prompt = ""
	ti.Focus()
	return ti
}

func resetSearchInput(ti *textinput.Model) {
	ti.SetValue("")
}
