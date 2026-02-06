package filterui

import (
	"fmt"
	"log/slog"
	"maps"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/parfenovvs/lazylogcat/internal/model"
	"github.com/parfenovvs/lazylogcat/internal/tui"
	"github.com/parfenovvs/lazylogcat/internal/tui/theme"
	"github.com/parfenovvs/lazylogcat/internal/util"
)

type FilterViewModel struct {
	viewportSize model.Size
	packageInput textinput.Model
	tagInput     textinput.Model
	textInput    textinput.Model

	deviceId   string
	filter     model.Filter
	format     model.Format
	tempFilter model.Filter
	tempFormat model.Format

	activePanel    int
	formatCursor   int
	modifierCursor int
	validationErr  string
}

func New(viewportSize model.Size, deviceId string, filter model.Filter, format model.Format) FilterViewModel {
	inputWidth := (viewportSize.Width-12)/3 - 4 // Account for panel padding

	pi := textinput.New()
	pi.Placeholder = "Package name..."
	pi.CharLimit = 100
	pi.Width = inputWidth
	pi.SetValue(filter.PackageName)
	pi.Blur()

	tagInput := textinput.New()
	tagInput.Placeholder = "Tag value..."
	tagInput.CharLimit = 100
	tagInput.Width = inputWidth
	tagInput.SetValue(filter.Tag)
	tagInput.Blur()

	txtInput := textinput.New()
	txtInput.Placeholder = "Search text..."
	txtInput.CharLimit = 100
	txtInput.Width = inputWidth
	txtInput.SetValue(filter.Text)
	txtInput.Blur()

	return FilterViewModel{
		viewportSize:   viewportSize,
		deviceId:       deviceId,
		filter:         filter,
		format:         format,
		tempFilter:     filter,
		tempFormat:     cloneFormat(format),
		packageInput:   pi,
		tagInput:       tagInput,
		textInput:      txtInput,
		activePanel:    0,
		formatCursor:   format.FormatIndex(),
		modifierCursor: 0,
	}
}

func cloneFormat(f model.Format) model.Format {
	return model.Format{
		SelectedFormat:  f.SelectedFormat,
		ActiveModifiers: maps.Clone(f.ActiveModifiers),
	}
}

func formatsEqual(a, b model.Format) bool {
	if a.SelectedFormat != b.SelectedFormat {
		return false
	}
	if len(a.ActiveModifiers) != len(b.ActiveModifiers) {
		return false
	}
	for k, v := range a.ActiveModifiers {
		if b.ActiveModifiers[k] != v {
			return false
		}
	}
	return true
}

func (m *FilterViewModel) Exit(apply bool) (bool, error) {
	if apply {
		newPackage := strings.TrimSpace(m.packageInput.Value())

		if newPackage != m.filter.PackageName {
			if newPackage != "" {
				_, err := util.GetPidByPackageName(m.deviceId, newPackage)
				if err != nil {
					m.validationErr = "Package not found."
					return false, err
				}
			}

			m.tempFilter.PackageName = newPackage
		}

		m.tempFilter.Tag = strings.TrimSpace(m.tagInput.Value())
		m.tempFilter.Text = strings.TrimSpace(m.textInput.Value())

		filterChanged := m.filter != m.tempFilter
		formatChanged := !formatsEqual(m.format, m.tempFormat)

		m.filter = m.tempFilter
		m.format = m.tempFormat

		if m.packageInput.Focused() {
			m.packageInput.Blur()
		}
		if m.tagInput.Focused() {
			m.tagInput.Blur()
		}
		if m.textInput.Focused() {
			m.textInput.Blur()
		}

		return filterChanged || formatChanged, nil
	}

	m.validationErr = ""
	m.packageInput.SetValue(m.filter.PackageName)
	m.tagInput.SetValue(m.filter.Tag)
	m.textInput.SetValue(m.filter.Text)
	if m.packageInput.Focused() {
		m.packageInput.Blur()
	}
	if m.tagInput.Focused() {
		m.tagInput.Blur()
	}
	if m.textInput.Focused() {
		m.textInput.Blur()
	}
	return false, nil
}

func (m FilterViewModel) Update(msg tea.Msg) (FilterViewModel, tea.Cmd) {
	switch msg := msg.(type) {
	case model.Size:
		m.viewportSize = msg
		inputWidth := (msg.Width-12)/3 - 4
		m.packageInput.Width = inputWidth
		m.tagInput.Width = inputWidth
		m.textInput.Width = inputWidth

	case tea.KeyMsg:
		if m.activePanel == 0 || m.activePanel == 1 {
			switch msg.String() {
			case "j", "down":
				switch m.activePanel {
				case 0:
					if m.formatCursor < len(model.AllFormats)-1 {
						m.formatCursor++
					}
				case 1:
					if m.modifierCursor < len(model.AllModifiers)-1 {
						m.modifierCursor++
					}
				}
			case "k", "up":
				switch m.activePanel {
				case 0:
					if m.formatCursor > 0 {
						m.formatCursor--
					}
				case 1:
					if m.modifierCursor > 0 {
						m.modifierCursor--
					}
				}
			case " ", "enter":
				switch m.activePanel {
				case 0:
					m.tempFormat.SetFormatByIndex(m.formatCursor)
				case 1:
					m.tempFormat.ToggleModifierByIndex(m.modifierCursor)
				}
			}
		}

		switch msg.String() {
		case "ctrl+s":
			changed, err := m.Exit(true)
			if err != nil {
				slog.Error("Error saving filter", "err", err)
				return m, nil
			}
			return m, func() tea.Msg {
				if changed {
					return tui.UpdateFilterCmd{
						Filter: m.filter,
						Format: m.format,
					}
				}
				return tui.NavigateToLogcatCmd{}
			}

		case "esc":
			m.Exit(false)
			return m, func() tea.Msg {
				return tui.NavigateToLogcatCmd{}
			}
		case "tab", "shift+tab":
			switch m.activePanel {
			case 2:
				m.packageInput.Blur()
			case 3:
				m.tagInput.Blur()
			case 4:
				m.textInput.Blur()
			}
			if msg.String() == "shift+tab" {
				m.activePanel = (m.activePanel - 1 + 5) % 5
			} else {
				m.activePanel = (m.activePanel + 1) % 5
			}
			if m.activePanel == 2 {
				m.packageInput.Focus()
				return m, textinput.Blink
			}
			if m.activePanel == 3 {
				m.tagInput.Focus()
				return m, textinput.Blink
			}
			if m.activePanel == 4 {
				m.textInput.Focus()
				return m, textinput.Blink
			}
		default:
			if m.activePanel == 2 {
				var cmd tea.Cmd
				m.packageInput, cmd = m.packageInput.Update(msg)
				if m.validationErr != "" {
					m.validationErr = ""
				}
				return m, cmd
			}
			if m.activePanel == 3 {
				var cmd tea.Cmd
				m.tagInput, cmd = m.tagInput.Update(msg)
				return m, cmd
			}
			if m.activePanel == 4 {
				var cmd tea.Cmd
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
		}
	}
	return m, nil
}

func (m FilterViewModel) View() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Bold(true).
		Render("Filter & Format Settings")
	b.WriteString(title + "\n\n")

	modifierPanel := m.renderModifierPanel()
	formatPanel := m.renderFormatPanel()
	panels := lipgloss.JoinHorizontal(lipgloss.Top, formatPanel, "  ", modifierPanel)
	b.WriteString(panels + "\n\n")

	// Calculate equal width for three filter panels
	filterPanelWidth := (m.viewportSize.Width - 12) / 3 // 12 for spacing (2 gaps * 2 spaces + 4*2 padding)

	packagePanel := m.renderPackagePanel(filterPanelWidth)
	tagPanel := m.renderTagPanel(filterPanelWidth)
	textPanel := m.renderTextPanel(filterPanelWidth)

	filterPanels := lipgloss.JoinHorizontal(lipgloss.Top, packagePanel, "  ", tagPanel, "  ", textPanel)
	b.WriteString(filterPanels + "\n\n")

	help := lipgloss.NewStyle().
		Foreground(theme.FGHelp).
		AlignHorizontal(lipgloss.Center).
		Render("tab switch panels • ↑/k up • ↓/j down • space/enter select • esc back • ctrl+s apply")
	b.WriteString(help)

	return lipgloss.Place(
		m.viewportSize.Width,
		m.viewportSize.Height,
		lipgloss.Center,
		lipgloss.Center,
		b.String(),
	)
}

func (m FilterViewModel) renderFormatPanel() string {
	var b strings.Builder

	panelTitleStyle := lipgloss.NewStyle().
		Bold(true)

	if m.activePanel == 0 {
		panelTitleStyle = panelTitleStyle.Foreground(theme.FGActiveTitle)
	}

	b.WriteString(panelTitleStyle.Render("Format") + "\n\n")

	for i, name := range model.AllFormats {
		selected := m.tempFormat.IsFormatValue(name)
		cursor := m.activePanel == 0 && m.formatCursor == i

		line := m.renderRadioButton(name, selected, cursor)
		b.WriteString(line)
		if i < len(model.AllFormats)-1 {
			b.WriteString("\n")
		}
	}

	hCompensator := model.MaxModifiers - len(model.AllFormats)
	for range hCompensator {
		b.WriteString("\n")
	}

	panelStyle := theme.Panel().
		Width(m.viewportSize.Width/2 - 4)

	if m.activePanel == 0 {
		panelStyle = theme.ActivePanel().
			Width(m.viewportSize.Width/2 - 4)
	}

	return panelStyle.Render(b.String())
}

func (m FilterViewModel) renderModifierPanel() string {
	var b strings.Builder

	panelTitleStyle := lipgloss.NewStyle().
		Bold(true)

	if m.activePanel == 1 {
		panelTitleStyle = panelTitleStyle.Foreground(theme.FGActiveTitle)
	}

	b.WriteString(panelTitleStyle.Render("Modifiers") + "\n\n")

	for i, name := range model.AllModifiers {
		selected := m.tempFormat.IsModifierActive(name)
		cursor := m.activePanel == 1 && m.modifierCursor == i

		line := m.renderCheckbox(name, selected, cursor)
		b.WriteString(line)
		if i < len(model.AllModifiers)-1 {
			b.WriteString("\n")
		}
	}

	panelStyle := theme.Panel().
		Width(m.viewportSize.Width/2 - 4)

	if m.activePanel == 1 {
		panelStyle = theme.ActivePanel().
			Width(m.viewportSize.Width/2 - 4)
	}

	result := b.String()
	return panelStyle.Render(result)
}

func (m FilterViewModel) renderPackagePanel(width int) string {
	var b strings.Builder

	panelTitleStyle := lipgloss.NewStyle().
		Bold(true)

	if m.activePanel == 2 {
		panelTitleStyle = panelTitleStyle.Foreground(theme.FGActiveTitle)
	}

	b.WriteString(panelTitleStyle.Render("Package") + "\n\n")

	b.WriteString(m.packageInput.View())

	if m.validationErr != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(theme.FGError)
		b.WriteString("\n\n" + errorStyle.Render(m.validationErr))
	}

	panelStyle := theme.Panel().
		Width(width)

	if m.activePanel == 2 {
		panelStyle = theme.ActivePanel().
			Width(width)
	}

	return panelStyle.Render(b.String())
}

func (m FilterViewModel) renderTagPanel(width int) string {
	var b strings.Builder

	panelTitleStyle := lipgloss.NewStyle().
		Bold(true)

	if m.activePanel == 3 {
		panelTitleStyle = panelTitleStyle.Foreground(theme.FGActiveTitle)
	}

	b.WriteString(panelTitleStyle.Render("Tag") + "\n\n")

	b.WriteString(m.tagInput.View())

	panelStyle := theme.Panel().
		Width(width)

	if m.activePanel == 3 {
		panelStyle = theme.ActivePanel().
			Width(width)
	}

	return panelStyle.Render(b.String())
}

func (m FilterViewModel) renderTextPanel(width int) string {
	var b strings.Builder

	panelTitleStyle := lipgloss.NewStyle().
		Bold(true)

	if m.activePanel == 4 {
		panelTitleStyle = panelTitleStyle.Foreground(theme.FGActiveTitle)
	}

	b.WriteString(panelTitleStyle.Render("Text") + "\n\n")

	b.WriteString(m.textInput.View())

	panelStyle := theme.Panel().
		Width(width)

	if m.activePanel == 4 {
		panelStyle = theme.ActivePanel().
			Width(width)
	}

	return panelStyle.Render(b.String())
}

func (m FilterViewModel) renderRadioButton(label string, selected bool, cursor bool) string {
	indicator := "( )"
	if selected {
		indicator = "(●)"
	}

	line := fmt.Sprintf("%s %s", indicator, label)

	if cursor {
		// Highlight current cursor position
		return lipgloss.NewStyle().
			Background(theme.BGCursor).
			Foreground(theme.FGSelected).
			Width(25).
			Render(line)
	}

	return line
}

func (m FilterViewModel) renderCheckbox(label string, checked bool, cursor bool) string {
	indicator := "[ ]"
	if checked {
		indicator = "[✓]"
	}

	line := fmt.Sprintf("%s %s", indicator, label)

	if cursor {
		return lipgloss.NewStyle().
			Background(theme.BGCursor).
			Foreground(theme.FGSelected).
			Width(30).
			Render(line)
	}

	return line
}
