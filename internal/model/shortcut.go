package model

import "strings"

type Shortcut struct {
	Key  string
	Name string
}

// CtrlXShortcuts lists ctrl+x second-key bindings from [Commands] for help overlays.
func CtrlXShortcuts() []Shortcut {
	var out []Shortcut
	for _, group := range Commands() {
		for _, cmd := range group.Commands {
			if cmd.Shortcut == "" {
				continue
			}
			parts := strings.SplitN(cmd.Shortcut, " ", 2)
			if len(parts) == 2 && parts[0] == "ctrl+x" {
				out = append(out, Shortcut{Key: parts[1], Name: cmd.Name})
			}
		}
	}
	return out
}
