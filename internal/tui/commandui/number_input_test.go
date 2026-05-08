package commandui

import (
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestNumberInputCharLimit(t *testing.T) {
	if got := numberInputCharLimit(99); got != 2 {
		t.Errorf("numberInputCharLimit(99) = %d, want 2", got)
	}
	if got := numberInputCharLimit(9); got != 1 {
		t.Errorf("numberInputCharLimit(9) = %d, want 1", got)
	}
	if got := numberInputCharLimit(0); got != defaultNumberInputCharLimit {
		t.Errorf("numberInputCharLimit(0) = %d, want %d", got, defaultNumberInputCharLimit)
	}
}

func TestDigitsOnly(t *testing.T) {
	if got := digitsOnly("a1b2"); got != "12" {
		t.Errorf("digitsOnly(...) = %q, want %q", got, "12")
	}
	if got := digitsOnly(""); got != "" {
		t.Errorf("digitsOnly(\"\") = %q, want empty", got)
	}
}

func TestNumberSelector_InitialZero_EmptyField(t *testing.T) {
	m := newNumberSelect(NumberSelectorConfig{Value: 0, Max: 99})
	if m.input.Value() != "" {
		t.Errorf("initial Value 0 should show empty field, got %q", m.input.Value())
	}
}

func TestNumberSelector_EnterEmpty_SubmitsZero(t *testing.T) {
	m := newNumberSelect(NumberSelectorConfig{
		Value: 99,
		ValidateFn: func(value int) error {
			if value < 0 {
				return fmt.Errorf("negative")
			}
			return nil
		},
	})
	m.input.SetValue("")
	next, _ := m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	if !next.Submitted() {
		t.Fatal("expected submitted")
	}
	if next.Value() != 0 {
		t.Errorf("Value() = %d, want 0", next.Value())
	}
}

func TestNumberSelector_StripsNonDigitsOnUpdate(t *testing.T) {
	m := newNumberSelect(NumberSelectorConfig{Max: 0}) // default digit length cap
	m.input.SetValue("12ab34")
	next := m.applyDigitsOnlyToInput()
	if next.input.Value() != "1234" {
		t.Errorf("after sanitize value = %q, want %q", next.input.Value(), "1234")
	}
}

func TestNumberSelector_MaxTruncatesInitialDigits(t *testing.T) {
	m := newNumberSelect(NumberSelectorConfig{Max: 99, Value: 12345})
	if got := m.input.Value(); got != "12" {
		t.Errorf("initial input = %q, want %q", got, "12")
	}
	if m.charLimit != 2 {
		t.Errorf("charLimit = %d, want 2", m.charLimit)
	}
}

func TestNumberSelector_ValidateReject_NoSubmit(t *testing.T) {
	m := newNumberSelect(NumberSelectorConfig{
		ValidateFn: func(value int) error {
			return fmt.Errorf("bad")
		},
	})
	m.input.SetValue("3")
	next, _ := m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	if next.Submitted() {
		t.Fatal("did not expect submitted")
	}
	if next.errorMsg == "" {
		t.Fatal("expected error message")
	}
}
