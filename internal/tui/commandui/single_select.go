package commandui

import (
	"strings"

	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/parfenovvs/lazylogcat/internal/tui/theme"
)

// SingleSelectItem represents one selectable row in a single-select table.
type SingleSelectItem struct {
	Key     string   // Unique identifier used for matching the current selection and search
	Columns []string // Display values for each column (excluding the marker column)
}

// SingleSelectConfig holds the parameters for creating a SingleSelectModel.
type SingleSelectConfig struct {
	Title      string
	Footer     string
	Columns    []table.Column
	Items      []SingleSelectItem
	CurrentKey string // Key of the currently active item (gets "●" marker)
}

// SingleSelectModel is a reusable single-select table widget with built-in search filtering.
type SingleSelectModel struct {
	title       string
	footer      string
	columns     []table.Column
	allItems    []SingleSelectItem
	currentKey  string
	table       table.Model
	itemMap     map[int]SingleSelectItem
	searchInput textinput.Model
	selected    bool
	selectedKey string
}

// NewSingleSelect creates a new SingleSelectModel from the given config.
func NewSingleSelect(cfg SingleSelectConfig) SingleSelectModel {
	m := SingleSelectModel{
		title:       cfg.Title,
		footer:      cfg.Footer,
		columns:     cfg.Columns,
		allItems:    cfg.Items,
		currentKey:  cfg.CurrentKey,
		searchInput: newSearchInput(),
	}
	m.buildRows(cfg.Items, cfg.CurrentKey)
	return m
}

func (m *SingleSelectModel) buildRows(items []SingleSelectItem, currentKey string) {
	// Add a marker column as the first column
	cols := make([]table.Column, len(m.columns)+1)
	cols[0] = table.Column{Title: "", Width: 1}
	copy(cols[1:], m.columns)

	m.itemMap = make(map[int]SingleSelectItem)
	var rows []table.Row
	initialCursor := 0
	for i, item := range items {
		m.itemMap[i] = item
		marker := ""
		if item.Key == currentKey {
			marker = "●"
			initialCursor = i
		}
		row := make(table.Row, len(item.Columns)+1)
		row[0] = marker
		copy(row[1:], item.Columns)
		rows = append(rows, row)
	}

	height := max(len(rows)+1, 2)

	m.table = newTable(cols, rows, height)
	if len(rows) > 0 {
		m.table.SetCursor(initialCursor)
	}
}

// Update handles key messages. Returns the updated model and a tea.Cmd.
// After calling Update, check Selected() to see if an item was chosen.
func (m SingleSelectModel) Update(msg tea.KeyPressMsg, key string) (SingleSelectModel, tea.Cmd) {
	if key == "enter" {
		if item, ok := m.itemMap[m.table.Cursor()]; ok {
			m.selected = true
			m.selectedKey = item.Key
		}
		return m, nil
	}

	// Arrow keys go to table navigation
	if key == "up" || key == "down" {
		if len(m.itemMap) > 0 {
			m.table, _ = m.table.Update(msg)
		}
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
func (m SingleSelectModel) UpdateBlink(msg tea.Msg) (SingleSelectModel, tea.Cmd) {
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	return m, cmd
}

// Selected returns true if an item was selected via Enter.
func (m SingleSelectModel) Selected() bool {
	return m.selected
}

// SelectedKey returns the key of the selected item. Only valid when Selected() is true.
func (m SingleSelectModel) SelectedKey() string {
	return m.selectedKey
}

func (m *SingleSelectModel) filterRows() {
	query := strings.ToLower(strings.TrimSpace(m.searchInput.Value()))

	if query == "" {
		m.buildRows(m.allItems, m.currentKey)
		return
	}

	var filtered []SingleSelectItem
	for _, item := range m.allItems {
		match := false
		if strings.Contains(strings.ToLower(item.Key), query) {
			match = true
		}
		if !match {
			for _, col := range item.Columns {
				if strings.Contains(strings.ToLower(col), query) {
					match = true
					break
				}
			}
		}
		if match {
			filtered = append(filtered, item)
		}
	}

	m.buildRows(filtered, m.currentKey)
}

// ResetSearch clears the search input and restores all items.
func (m *SingleSelectModel) ResetSearch() {
	resetSearchInput(&m.searchInput)
}

// Rebuild recreates the table with new items, keeping the same config.
// Useful for refreshing data (e.g. device list).
func (m *SingleSelectModel) Rebuild(items []SingleSelectItem, currentKey string) {
	m.allItems = items
	m.currentKey = currentKey
	resetSearchInput(&m.searchInput)
	m.buildRows(items, currentKey)
}

// View renders the single-select dialog.
func (m SingleSelectModel) View() string {
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
