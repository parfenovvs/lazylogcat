package util

import (
	"testing"

	"github.com/parfenovvs/lazylogcat/internal/model"
)

func ll(raw string) model.LogLine {
	return model.LogLine{Raw: raw}
}

func TestNewRingBuffer(t *testing.T) {
	t.Run("ZeroCapacity", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		NewRingBuffer(0)
	})

	t.Run("PositiveCapacity", func(t *testing.T) {
		buffer := NewRingBuffer(3)
		if buffer.capacity != 3 {
			t.Errorf("Expected %v, got %v", 3, buffer.capacity)
		}
		if buffer.Size() != 0 {
			t.Errorf("Expected size 0, got %v", buffer.Size())
		}
		if buffer.head != 0 {
			t.Errorf("Expected head 0, got %v", buffer.head)
		}
		if buffer.lines == nil || len(buffer.lines) != 3 {
			t.Errorf("Expected lines slice of length 3, got %v", len(buffer.lines))
		}
	})
}

func TestAppend(t *testing.T) {
	t.Run("WithinCapacity", func(t *testing.T) {
		buffer := NewRingBuffer(3)
		buffer.Append(ll("A"))
		buffer.Append(ll("B"))
		if buffer.Size() != 2 {
			t.Errorf("Expected size 2, got %v", buffer.Size())
		}
		if buffer.head != 2 {
			t.Errorf("Expected head 2, got %v", buffer.head)
		}
		expected := []string{"A", "B", ""}
		for i, v := range expected {
			if buffer.lines[i].Raw != v {
				t.Errorf("At index %d, expected %v, got %v", i, v, buffer.lines[i].Raw)
			}
		}
	})

	t.Run("ExceedCapacity", func(t *testing.T) {
		buffer := NewRingBuffer(2)
		buffer.Append(ll("A"))
		buffer.Append(ll("B"))
		buffer.Append(ll("C"))
		if buffer.Size() != 2 {
			t.Errorf("Expected size 2, got %v", buffer.Size())
		}
		if buffer.head != 1 {
			t.Errorf("Expected head 1, got %v", buffer.head)
		}
		expected := []string{"C", "B"}
		for i, v := range expected {
			if buffer.lines[i].Raw != v {
				t.Errorf("At index %d, expected %v, got %v", i, v, buffer.lines[i].Raw)
			}
		}
	})
}

func TestGetRecent(t *testing.T) {
	t.Run("FromEmpty", func(t *testing.T) {
		buffer := NewRingBuffer(5)
		recent := buffer.Recent(3)
		if len(recent) != 0 {
			t.Errorf("Expected empty slice, got %v", recent)
		}
	})

	t.Run("LessThanCapacity", func(t *testing.T) {
		buffer := NewRingBuffer(5)
		buffer.Append(ll("A"))
		buffer.Append(ll("B"))
		recent := buffer.Recent(3)
		expected := []string{"A", "B"}
		if len(recent) != len(expected) {
			t.Errorf("Expected length %d, got %d", len(expected), len(recent))
		}
		for i, v := range expected {
			if recent[i].Raw != v {
				t.Errorf("At index %d, expected %v, got %v", i, v, recent[i].Raw)
			}
		}
	})

	t.Run("MoreThanCapacity", func(t *testing.T) {
		buffer := NewRingBuffer(5)
		buffer.Append(ll("A"))
		buffer.Append(ll("B"))
		buffer.Append(ll("C"))
		buffer.Append(ll("D"))
		buffer.Append(ll("E"))
		buffer.Append(ll("F"))
		recent := buffer.Recent(3)
		expected := []string{"D", "E", "F"}
		if len(recent) != len(expected) {
			t.Errorf("Expected length %d, got %d", len(expected), len(recent))
		}
		for i, v := range expected {
			if recent[i].Raw != v {
				t.Errorf("At index %d, expected %v, got %v", i, v, recent[i].Raw)
			}
		}
	})

	t.Run("RequestMoreThanSize", func(t *testing.T) {
		buffer := NewRingBuffer(5)
		buffer.Append(ll("A"))
		recent := buffer.Recent(10)
		expected := []string{"A"}
		if len(recent) != len(expected) {
			t.Errorf("Expected length %d, got %d", len(expected), len(recent))
		}
		if recent[0].Raw != "A" {
			t.Errorf("Expected A, got %v", recent[0].Raw)
		}
	})
}

func TestGetAll(t *testing.T) {
	buffer := NewRingBuffer(3)
	buffer.Append(ll("A"))
	buffer.Append(ll("B"))
	all := buffer.All()
	expected := []string{"A", "B"}
	if len(all) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(all))
	}
	for i, v := range expected {
		if all[i].Raw != v {
			t.Errorf("At index %d, expected %v, got %v", i, v, all[i].Raw)
		}
	}
}

func TestSize(t *testing.T) {
	t.Run("EmptyBuffer", func(t *testing.T) {
		buffer := NewRingBuffer(5)
		if buffer.Size() != 0 {
			t.Errorf("Expected size 0, got %v", buffer.Size())
		}
	})

	t.Run("AfterAppend", func(t *testing.T) {
		buffer := NewRingBuffer(5)
		buffer.Append(ll("A"))
		if buffer.Size() != 1 {
			t.Errorf("Expected size 1, got %v", buffer.Size())
		}
		buffer.Append(ll("B"))
		buffer.Append(ll("C"))
		if buffer.Size() != 3 {
			t.Errorf("Expected size 3, got %v", buffer.Size())
		}
	})

	t.Run("AfterExceedingCapacity", func(t *testing.T) {
		buffer := NewRingBuffer(3)
		buffer.Append(ll("A"))
		buffer.Append(ll("B"))
		buffer.Append(ll("C"))
		buffer.Append(ll("D"))
		buffer.Append(ll("E"))
		if buffer.Size() != 3 {
			t.Errorf("Expected size 3 (capacity), got %v", buffer.Size())
		}
	})
}

func TestClear(t *testing.T) {
	t.Run("ClearEmptyBuffer", func(t *testing.T) {
		buffer := NewRingBuffer(3)
		buffer.Clear()
		if buffer.Size() != 0 {
			t.Errorf("Expected size 0, got %v", buffer.Size())
		}
		all := buffer.All()
		if len(all) != 0 {
			t.Errorf("Expected empty slice after clear, got %v", all)
		}
	})

	t.Run("ClearNonEmptyBuffer", func(t *testing.T) {
		buffer := NewRingBuffer(3)
		buffer.Append(ll("A"))
		buffer.Append(ll("B"))
		buffer.Append(ll("C"))
		if buffer.Size() != 3 {
			t.Errorf("Expected size 3 before clear, got %v", buffer.Size())
		}
		buffer.Clear()
		if buffer.Size() != 0 {
			t.Errorf("Expected size 0 after clear, got %v", buffer.Size())
		}
		all := buffer.All()
		if len(all) != 0 {
			t.Errorf("Expected empty slice after clear, got %v", all)
		}
	})

	t.Run("AppendAfterClear", func(t *testing.T) {
		buffer := NewRingBuffer(3)
		buffer.Append(ll("A"))
		buffer.Append(ll("B"))
		buffer.Clear()
		buffer.Append(ll("X"))
		buffer.Append(ll("Y"))
		if buffer.Size() != 2 {
			t.Errorf("Expected size 2 after clear and append, got %v", buffer.Size())
		}
		all := buffer.All()
		expected := []string{"X", "Y"}
		if len(all) != len(expected) {
			t.Errorf("Expected length %d, got %d", len(expected), len(all))
		}
		for i, v := range expected {
			if all[i].Raw != v {
				t.Errorf("At index %d, expected %v, got %v", i, v, all[i].Raw)
			}
		}
	})
}
