package commandui

import (
	"maps"
	"strings"

	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/parfenovvs/lazylogcat/internal/tui/theme"
)

// MultiSelectConfig holds the parameters for creating a MultiSelectModel.
type MultiSelectConfig struct {
	Title   string
	Footer  string
	Columns []table.Column
	Items   []string        // All item names
	Active  map[string]bool // Currently active items
}

// MultiSelectModel is a reusable multi-select table widget with built-in search filtering.
// Items are toggled with Enter or Space. The active set is returned via ActiveItems().
type MultiSelectModel struct {
	title       string
	footer      string
	columns     []table.Column
	allItems    []string
	active      map[string]bool
	table       table.Model
	itemMap     map[int]string
	searchInput textinput.Model
}

// NewMultiSelect creates a new MultiSelectModel from the given config.
func NewMultiSelect(cfg MultiSelectConfig) MultiSelectModel {
	active := make(map[string]bool)
	maps.Copy(active, cfg.Active)
	m := MultiSelectModel{
		title:       cfg.Title,
		footer:      cfg.Footer,
		columns:     cfg.Columns,
		allItems:    cfg.Items,
		active:      active,
		searchInput: newSearchInput(),
	}
	m.buildRows(cfg.Items)
	return m
}

func (m *MultiSelectModel) buildRows(items []string) {
	// Add a marker column as the first column
	cols := make([]table.Column, len(m.columns)+1)
	cols[0] = table.Column{Title: "", Width: 3}
	copy(cols[1:], m.columns)

	// Preserve cursor position by name if possible
	oldCursorName := ""
	if m.itemMap != nil {
		if name, ok := m.itemMap[m.table.Cursor()]; ok {
			oldCursorName = name
		}
	}

	m.itemMap = make(map[int]string)
	var rows []table.Row
	newCursor := 0
	for _, name := range items {
		if name == oldCursorName {
			newCursor = len(rows)
		}
		m.itemMap[len(rows)] = name
		marker := "[ ]"
		if m.active[name] {
			marker = "[✓]"
		}
		rows = append(rows, table.Row{marker, name})
	}

	m.table = newTable(cols, rows, len(rows)+1)
	if len(rows) > 0 {
		if newCursor < len(rows) {
			m.table.SetCursor(newCursor)
		} else {
			m.table.SetCursor(0)
		}
	}
}

// Update handles key messages. Returns the updated model and a tea.Cmd.
func (m MultiSelectModel) Update(msg tea.KeyPressMsg, key string) (MultiSelectModel, tea.Cmd) {
	if key == "enter" || key == "space" {
		if name, ok := m.itemMap[m.table.Cursor()]; ok {
			if m.active[name] {
				delete(m.active, name)
			} else {
				m.active[name] = true
			}
			m.filterRows()
		}
		return m, nil
	}

	// Arrow keys go to table navigation
	if key == "up" || key == "down" || key == "ctrl+k" || key == "ctrl+j" {
		m.table, _ = m.table.Update(msg)
		return m, nil
	}

	// All other keys go to the search input
	prevValue := m.searchInput.Value()
	m.searchInput, _ = m.searchInput.Update(msg)
	if m.searchInput.Value() != prevValue {
		m.filterRows()
	}

	return m, nil
}

// UpdateBlink forwards non-key messages (e.g. cursor blink) to the search input.
func (m MultiSelectModel) UpdateBlink(msg tea.Msg) (MultiSelectModel, tea.Cmd) {
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	return m, cmd
}

// ActiveItems returns a copy of the current active items map.
func (m MultiSelectModel) ActiveItems() map[string]bool {
	result := make(map[string]bool)
	for k, v := range m.active {
		result[k] = v
	}
	return result
}

func (m *MultiSelectModel) filterRows() {
	query := strings.ToLower(strings.TrimSpace(m.searchInput.Value()))

	if query == "" {
		m.buildRows(m.allItems)
		return
	}

	var filtered []string
	for _, name := range m.allItems {
		if strings.Contains(strings.ToLower(name), query) {
			filtered = append(filtered, name)
		}
	}

	m.buildRows(filtered)
}

// View renders the multi-select dialog.
func (m MultiSelectModel) View() string {
	title := dialogTitleWithESC(m.title)
	footer := theme.DialogHelp().Render(m.footer)

	var body string
	if len(m.table.Rows()) == 0 && m.searchInput.Value() != "" {
		body = theme.DialogHelp().Render("\nNo results found")
	} else {
		body = m.table.View()
	}

	search := theme.DialogSearch().Render(m.searchInput.View())
	content := title + "\n\n" + search + "\n" + body + "\n\n" + footer
	return dialogStyle().Render(content)
}
