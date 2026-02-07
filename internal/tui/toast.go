package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/parfenovvs/lazylogcat/internal/tui/theme"
)

const toastDuration = 2 * time.Second

// ToastExpiredMsg is sent when a toast's display timer expires.
// The ID field ensures stale timers don't dismiss newer toasts.
type ToastExpiredMsg struct {
	ID int
}

// ToastModel is a reusable component for showing temporary messages.
// Embed it in any view model that needs transient user notifications.
type ToastModel struct {
	message string
	visible bool
	id      int
}

// Show displays a toast message and returns a tea.Cmd that will dismiss it
// after the standard duration. Each call increments the internal ID so that
// only the most recent timer can dismiss the toast.
func (t *ToastModel) Show(message string) tea.Cmd {
	t.id++
	t.message = message
	t.visible = true
	id := t.id
	return tea.Tick(toastDuration, func(_ time.Time) tea.Msg {
		return ToastExpiredMsg{ID: id}
	})
}

// Update handles ToastExpiredMsg. Call this from the parent model's Update.
func (t *ToastModel) Update(msg tea.Msg) {
	if msg, ok := msg.(ToastExpiredMsg); ok {
		if msg.ID == t.id {
			t.visible = false
		}
	}
}

// View returns the styled toast string. If the toast is not visible, returns
// an empty string. The caller is responsible for positioning this in the view.
func (t ToastModel) View() string {
	if !t.visible {
		return ""
	}
	style := lipgloss.NewStyle().
		Foreground(theme.FGHelp)
	return style.Render(t.message)
}

// IsVisible reports whether the toast is currently displayed.
func (t ToastModel) IsVisible() bool {
	return t.visible
}
