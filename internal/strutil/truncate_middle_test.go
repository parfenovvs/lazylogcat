package strutil

import (
	"strings"
	"testing"
)

func TestTruncateMiddle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxWidth int
		want     string
	}{
		{name: "ShorterThanMax", input: "hello", maxWidth: 10, want: "hello"},
		{name: "ExactlyMax", input: "hello", maxWidth: 5, want: "hello"},
		{name: "LongerThanMax", input: "abcdefghij", maxWidth: 7, want: "abc…hij"},
		{name: "MaxWidth1", input: "abcdef", maxWidth: 1, want: "…"},
		{name: "MaxWidth2", input: "abcdef", maxWidth: 2, want: "…f"},
		{name: "MaxWidth3", input: "abcdef", maxWidth: 3, want: "a…f"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateMiddle(tt.input, tt.maxWidth)
			if got != tt.want {
				t.Errorf("TruncateMiddle(%q, %d) = %q, want %q", tt.input, tt.maxWidth, got, tt.want)
			}
		})
	}

	t.Run("Unicode", func(t *testing.T) {
		input := "日本語テスト文字列"
		got := TruncateMiddle(input, 5)
		if !strings.Contains(got, "…") {
			t.Errorf("TruncateMiddle(%q, 5) = %q, expected ellipsis in result", input, got)
		}
	})
}

func TestPadANSIWidth(t *testing.T) {
	if got := PadANSIWidth("hi", 5); got != "hi   " {
		t.Errorf("PadANSIWidth(%q, 5) = %q", "hi", got)
	}
	if got := PadANSIWidth("hello", 3); got != "hello" {
		t.Errorf("already wider string should be unchanged, got %q", got)
	}
}
