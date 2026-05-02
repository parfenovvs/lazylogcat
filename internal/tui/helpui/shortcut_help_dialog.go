package helpui

import (
	"slices"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/parfenovvs/lazylogcat/internal/model"
	"github.com/parfenovvs/lazylogcat/internal/tui/commonui"
	"github.com/parfenovvs/lazylogcat/internal/tui/theme"
)

// ShortcutHelpDialogModel shows ctrl+x shortcut hints while awaiting the second key.
type ShortcutHelpDialogModel struct {
	shortcuts []model.Shortcut
}

func NewShortcutHelpDialogModel(shortcuts []model.Shortcut) ShortcutHelpDialogModel {
	cp := slices.Clone(shortcuts)
	slices.SortFunc(cp, func(a, b model.Shortcut) int {
		return strings.Compare(a.Key, b.Key)
	})
	return ShortcutHelpDialogModel{shortcuts: cp}
}

func (m ShortcutHelpDialogModel) View() string {
	title := commonui.DialogTitleWithESC("")
	var body string
	if len(m.shortcuts) == 0 {
		body = theme.DialogHelp().Render("\nNo shortcuts")
	} else {
		var b strings.Builder
		for _, s := range m.shortcuts {
			b.WriteString("  ")
			b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(theme.ColorBrightMagenta).Render(s.Key))
			b.WriteString("  →  ")
			b.WriteString(s.Name)
			b.WriteString("\n")
		}
		body = b.String()
	}
	content := title + "\n\n" + body
	return commonui.DialogFrameStyle().Render(content)
}
