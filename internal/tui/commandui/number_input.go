package commandui

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/parfenovvs/lazylogcat/internal/tui/commonui"
	"github.com/parfenovvs/lazylogcat/internal/tui/theme"
)

// defaultNumberInputCharLimit is used when NumberSelectorConfig.Max is unset (<= 0).
const defaultNumberInputCharLimit = 12

type NumberSelectorConfig struct {
	Title  string
	Footer string
	Value  int
	// Max optionally caps the allowed numeric range for sizing the input; when > 0,
	// CharLimit is len(strconv.Itoa(Max)). When <= 0, defaultNumberInputCharLimit is used.
	Max        int
	ValidateFn func(value int) error
}

type NumberSelectorModel struct {
	title      string
	footer     string
	input      textinput.Model
	value      int
	charLimit  int
	validateFn func(value int) error
	errorMsg   string
	submitted  bool
}

func numberInputCharLimit(max int) int {
	if max <= 0 {
		return defaultNumberInputCharLimit
	}
	return len(strconv.Itoa(max))
}

func digitsOnly(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func clampDigitLen(s string, maxLen int) string {
	if maxLen <= 0 || len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

func newNumberSelect(cfg NumberSelectorConfig) NumberSelectorModel {
	limit := numberInputCharLimit(cfg.Max)
	ti := textinput.New()
	ti.CharLimit = limit
	ti.SetWidth(30)
	initial := ""
	if cfg.Value != 0 {
		initial = clampDigitLen(digitsOnly(fmt.Sprint(cfg.Value)), limit)
	}
	ti.SetValue(initial)
	ti.Focus()
	return NumberSelectorModel{
		title:      cfg.Title,
		footer:     cfg.Footer,
		input:      ti,
		value:      cfg.Value,
		charLimit:  limit,
		validateFn: cfg.ValidateFn,
		errorMsg:   "",
	}
}

func (m NumberSelectorModel) applyDigitsOnlyToInput() NumberSelectorModel {
	v := clampDigitLen(digitsOnly(m.input.Value()), m.charLimit)
	if v != m.input.Value() {
		m.input.SetValue(v)
	}
	return m
}

func (m NumberSelectorModel) Update(msg tea.Msg) (NumberSelectorModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		key := msg.String()
		if key == "enter" {
			value := strings.TrimSpace(digitsOnly(m.input.Value()))
			if value == "" {
				m.submitted = true
				return m, nil
			}
			intValue, err := strconv.Atoi(value)
			if err != nil {
				m.errorMsg = "Please enter a valid number"
				return m, nil
			}
			if m.validateFn != nil {
				if err := m.validateFn(intValue); err != nil {
					m.errorMsg = err.Error()
					return m, nil
				}
			}
			m.submitted = true
			return m, nil
		}

		if len(key) == 1 {
			r := rune(key[0])
			if !unicode.IsDigit(r) && (unicode.IsPrint(r) || r == '\t') {
				return m, nil
			}
		}

		// Forward other keys to the text input, clear error on change
		prevValue := m.input.Value()
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		m = m.applyDigitsOnlyToInput()
		if m.input.Value() != prevValue {
			m.errorMsg = ""
		}
		return m, cmd

	default:
		// Forward non-key messages (e.g. cursor blink)
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		m = m.applyDigitsOnlyToInput()
		return m, cmd
	}
}

// View renders the number selector dialog.
func (m NumberSelectorModel) View() string {
	title := commonui.DialogTitleWithESC(m.title)
	footer := theme.DialogHelp().Render(m.footer)

	var errorLine string
	if m.errorMsg != "" {
		errorLine = "\n" + theme.DialogError().Render(m.errorMsg)
	}

	input := theme.DialogSearch().Render(m.input.View())
	content := title + "\n\n" + input + errorLine + "\n\n" + footer
	return commonui.DialogFrameStyle().Render(content)
}

func (m NumberSelectorModel) Submitted() bool {
	return m.submitted
}

func (m NumberSelectorModel) Value() int {
	value, _ := strconv.Atoi(strings.TrimSpace(digitsOnly(m.input.Value())))
	return value
}
