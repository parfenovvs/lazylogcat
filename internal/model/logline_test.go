package model

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseLogLine(t *testing.T) {
	tests := []struct {
		name   string
		raw    string
		want   LogLine
		parsed bool
	}{
		{
			name: "StandardThreadtime",
			raw:  "02-08 12:12:09.629  3950  4005 D BusinessScope: Enqueuing the block",
			want: LogLine{
				Date:    "02-08",
				Time:    "12:12:09.629",
				PID:     "3950",
				TID:     "4005",
				Level:   "D",
				Tag:     "BusinessScope",
				Message: "Enqueuing the block",
				Raw:     "02-08 12:12:09.629  3950  4005 D BusinessScope: Enqueuing the block",
			},
			parsed: true,
		},
		{
			name: "MessageWithColons",
			raw:  "01-15 08:30:00.123  1234  5678 I MyApp: key:value pair: something:else",
			want: LogLine{
				Date:    "01-15",
				Time:    "08:30:00.123",
				PID:     "1234",
				TID:     "5678",
				Level:   "I",
				Tag:     "MyApp",
				Message: "key:value pair: something:else",
				Raw:     "01-15 08:30:00.123  1234  5678 I MyApp: key:value pair: something:else",
			},
			parsed: true,
		},
		{
			name: "TagWithDots",
			raw:  "02-08 12:12:09.630  3950  4069 D com.example.app: Start running",
			want: LogLine{
				Date:    "02-08",
				Time:    "12:12:09.630",
				PID:     "3950",
				TID:     "4069",
				Level:   "D",
				Tag:     "com.example.app",
				Message: "Start running",
				Raw:     "02-08 12:12:09.630  3950  4069 D com.example.app: Start running",
			},
			parsed: true,
		},
		{
			name: "WarningLevel",
			raw:  "03-01 14:00:00.000   100   200 W SomeTag: warning message",
			want: LogLine{
				Date:    "03-01",
				Time:    "14:00:00.000",
				PID:     "100",
				TID:     "200",
				Level:   "W",
				Tag:     "SomeTag",
				Message: "warning message",
				Raw:     "03-01 14:00:00.000   100   200 W SomeTag: warning message",
			},
			parsed: true,
		},
		{
			name: "ErrorLevel",
			raw:  "12-31 23:59:59.999 99999 99999 E CrashHandler: fatal error occurred",
			want: LogLine{
				Date:    "12-31",
				Time:    "23:59:59.999",
				PID:     "99999",
				TID:     "99999",
				Level:   "E",
				Tag:     "CrashHandler",
				Message: "fatal error occurred",
				Raw:     "12-31 23:59:59.999 99999 99999 E CrashHandler: fatal error occurred",
			},
			parsed: true,
		},
		{
			name: "FatalLevel",
			raw:  "01-01 00:00:00.000     1     1 F Kernel: panic",
			want: LogLine{
				Date:    "01-01",
				Time:    "00:00:00.000",
				PID:     "1",
				TID:     "1",
				Level:   "F",
				Tag:     "Kernel",
				Message: "panic",
				Raw:     "01-01 00:00:00.000     1     1 F Kernel: panic",
			},
			parsed: true,
		},
		{
			name: "EmptyMessage",
			raw:  "02-08 12:12:09.629  3950  4005 D BusinessScope:",
			want: LogLine{
				Date:    "02-08",
				Time:    "12:12:09.629",
				PID:     "3950",
				TID:     "4005",
				Level:   "D",
				Tag:     "BusinessScope",
				Message: "",
				Raw:     "02-08 12:12:09.629  3950  4005 D BusinessScope:",
			},
			parsed: true,
		},
		{
			name: "MessageWithExtraSpaces",
			raw:  "02-08 12:12:09.629  3950  4005 D Tag:  multiple  spaces  here",
			want: LogLine{
				Date:    "02-08",
				Time:    "12:12:09.629",
				PID:     "3950",
				TID:     "4005",
				Level:   "D",
				Tag:     "Tag",
				Message: " multiple  spaces  here",
				Raw:     "02-08 12:12:09.629  3950  4005 D Tag:  multiple  spaces  here",
			},
			parsed: true,
		},
		{
			name:   "SeparatorLine",
			raw:    "--------- beginning of main",
			want:   LogLine{Raw: "--------- beginning of main"},
			parsed: false,
		},
		{
			name:   "EmptyLine",
			raw:    "",
			want:   LogLine{Raw: ""},
			parsed: false,
		},
		{
			name:   "TooFewFields",
			raw:    "02-08 12:12:09.629 3950",
			want:   LogLine{Raw: "02-08 12:12:09.629 3950"},
			parsed: false,
		},
		{
			name:   "InvalidLevel",
			raw:    "02-08 12:12:09.629  3950  4005 X BusinessScope: message",
			want:   LogLine{Raw: "02-08 12:12:09.629  3950  4005 X BusinessScope: message"},
			parsed: false,
		},
		{
			name:   "LevelTooLong",
			raw:    "02-08 12:12:09.629  3950  4005 DEBUG BusinessScope: message",
			want:   LogLine{Raw: "02-08 12:12:09.629  3950  4005 DEBUG BusinessScope: message"},
			parsed: false,
		},
		{
			name: "TagWithSpaceBeforeColon",
			raw:  "02-08 12:12:09.629  3950  4005 D BusinessScope : Enqueuing the block",
			want: LogLine{
				Date:    "02-08",
				Time:    "12:12:09.629",
				PID:     "3950",
				TID:     "4005",
				Level:   "D",
				Tag:     "BusinessScope",
				Message: "Enqueuing the block",
				Raw:     "02-08 12:12:09.629  3950  4005 D BusinessScope : Enqueuing the block",
			},
			parsed: true,
		},
		{
			name: "TagWithSpaceBeforeColonEmptyMessage",
			raw:  "02-08 12:12:09.629  3950  4005 D BusinessScope :",
			want: LogLine{
				Date:    "02-08",
				Time:    "12:12:09.629",
				PID:     "3950",
				TID:     "4005",
				Level:   "D",
				Tag:     "BusinessScope",
				Message: "",
				Raw:     "02-08 12:12:09.629  3950  4005 D BusinessScope :",
			},
			parsed: true,
		},
		{
			name:   "TagWithoutColon",
			raw:    "02-08 12:12:09.629  3950  4005 D BusinessScope message here",
			want:   LogLine{Raw: "02-08 12:12:09.629  3950  4005 D BusinessScope message here"},
			parsed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseLogLine(tt.raw)
			if got.Date != tt.want.Date {
				t.Errorf("Date = %q, want %q", got.Date, tt.want.Date)
			}
			if got.Time != tt.want.Time {
				t.Errorf("Time = %q, want %q", got.Time, tt.want.Time)
			}
			if got.PID != tt.want.PID {
				t.Errorf("PID = %q, want %q", got.PID, tt.want.PID)
			}
			if got.TID != tt.want.TID {
				t.Errorf("TID = %q, want %q", got.TID, tt.want.TID)
			}
			if got.Level != tt.want.Level {
				t.Errorf("Level = %q, want %q", got.Level, tt.want.Level)
			}
			if got.Tag != tt.want.Tag {
				t.Errorf("Tag = %q, want %q", got.Tag, tt.want.Tag)
			}
			if got.Message != tt.want.Message {
				t.Errorf("Message = %q, want %q", got.Message, tt.want.Message)
			}
			if got.Raw != tt.want.Raw {
				t.Errorf("Raw = %q, want %q", got.Raw, tt.want.Raw)
			}
			if got.Parsed() != tt.parsed {
				t.Errorf("Parsed() = %v, want %v", got.Parsed(), tt.parsed)
			}
		})
	}
}

func TestLogLineString(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{
			name: "ParsedLine",
			raw:  "02-08 12:12:09.629  3950  4005 D BusinessScope: Enqueuing the block",
		},
		{
			name: "UnparsedLine",
			raw:  "--------- beginning of main",
		},
		{
			name: "EmptyLine",
			raw:  "",
		},
		{
			name: "LineWithExtraSpaces",
			raw:  "02-08 12:12:09.629  3950  4005 D Tag:  multiple  spaces  here",
		},
		{
			name: "TagWithSpaceBeforeColon",
			raw:  "02-08 12:12:09.629  3950  4005 D BusinessScope : Enqueuing the block",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseLogLine(tt.raw).String()
			if got != tt.raw {
				t.Errorf("String() = %q, want %q", got, tt.raw)
			}
		})
	}
}

func TestPrefixWidth(t *testing.T) {
	allCols := Columns{Date: true, Time: true, PID: true, TID: true, Level: true, Tag: true, Message: true}

	parsed := ParseLogLine("02-08 12:12:09.629  3950  4005 D BusinessScope: Enqueuing the block")
	// Prefix with all columns: "02-08 12:12:09.629  3950  4005 D BusinessScope: "
	// Date(5) + sp + Time(12) + sp + PID(5) + sp + TID(5) + sp + Level(1) + sp + Tag:(14) = 5+1+12+1+5+1+5+1+1+1+14 = 47, +1 trailing space = 48
	expectedAllPrefix := fmt.Sprintf("%s %s %5s %5s %s %s:",
		parsed.Date, parsed.Time, parsed.PID, parsed.TID, parsed.Level, parsed.Tag)
	expectedAllWidth := len(expectedAllPrefix) + 1 // +1 for trailing space before message

	tests := []struct {
		name string
		line LogLine
		cols Columns
		want int
	}{
		{
			name: "AllColumns",
			line: parsed,
			cols: allCols,
			want: expectedAllWidth,
		},
		{
			name: "NoPrefixColumns",
			line: parsed,
			cols: Columns{Message: true},
			want: 0,
		},
		{
			name: "OnlyTagAndMessage",
			line: parsed,
			cols: Columns{Tag: true, Message: true},
			want: len(parsed.Tag+":") + 1,
		},
		{
			name: "OnlyDateAndTime",
			line: parsed,
			cols: Columns{Date: true, Time: true},
			want: len(parsed.Date + " " + parsed.Time),
		},
		{
			name: "MessageDisabled",
			line: parsed,
			cols: Columns{Date: true, Time: true, Level: true, Tag: true},
			want: len(strings.Join([]string{parsed.Date, parsed.Time, parsed.Level, parsed.Tag + ":"}, " ")),
		},
		{
			name: "UnparsedLine",
			line: ParseLogLine("--------- beginning of main"),
			cols: allCols,
			want: 0,
		},
		{
			name: "EmptyLine",
			line: ParseLogLine(""),
			cols: allCols,
			want: 0,
		},
		{
			name: "ShortTag",
			line: ParseLogLine("02-08 12:12:09.629  3950  4005 D X: msg"),
			cols: allCols,
			want: len("02-08 12:12:09.629  3950  4005 D X:") + 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.line.PrefixWidth(OutputPrefs{Columns: tt.cols})
			if got != tt.want {
				t.Errorf("PrefixWidth() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestPrefixWidth_MatchesModifiedString(t *testing.T) {
	// Verify that PrefixWidth equals the position of the message start
	// in the ModifiedString output for various column configurations.
	line := ParseLogLine("02-08 12:12:09.629  3950  4005 D BusinessScope: Enqueuing the block")

	colSets := []struct {
		name string
		cols Columns
	}{
		{"AllColumns", Columns{Date: true, Time: true, PID: true, TID: true, Level: true, Tag: true, Message: true}},
		{"TagAndMessage", Columns{Tag: true, Message: true}},
		{"TimeAndMessage", Columns{Time: true, Message: true}},
		{"PIDTIDMessage", Columns{PID: true, TID: true, Message: true}},
		{"LevelTagMessage", Columns{Level: true, Tag: true, Message: true}},
	}

	for _, tc := range colSets {
		t.Run(tc.name, func(t *testing.T) {
			prefs := OutputPrefs{Columns: tc.cols}
			pw := line.PrefixWidth(prefs)
			full := line.ModifiedString(prefs)
			if pw == 0 {
				return // No prefix columns, nothing to check
			}
			// The message should start at position pw in the full string
			if pw > len(full) {
				t.Fatalf("PrefixWidth(%d) > len(ModifiedString)(%d)", pw, len(full))
			}
			msgPart := full[pw:]
			if msgPart != line.Message {
				t.Errorf("full[PrefixWidth:] = %q, want %q (full=%q, pw=%d)",
					msgPart, line.Message, full, pw)
			}
		})
	}
}

func TestModifiedString_TagWidth(t *testing.T) {
	line := ParseLogLine("02-08 12:12:09.629  3950  4005 D Short: hi")
	longTag := ParseLogLine("02-08 12:12:09.629  3950  4005 D VeryLongTagHere: there")

	tests := []struct {
		name  string
		line  LogLine
		prefs OutputPrefs
		want  string
	}{
		{
			name: "AutoWidth_Unchanged",
			line: line,
			prefs: OutputPrefs{
				Columns:  Columns{Tag: true, Message: true},
				TagWidth: 0,
			},
			want: "Short: hi",
		},
		{
			name: "PadShortTag",
			line: line,
			prefs: OutputPrefs{
				Columns:  Columns{Tag: true, Message: true},
				TagWidth: 10,
			},
			want: "Short     : hi",
		},
		{
			name: "TruncateLongTag",
			line: longTag,
			prefs: OutputPrefs{
				Columns:  Columns{Tag: true, Message: true},
				TagWidth: 8,
			},
			want: "Ver…Here: there",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.line.ModifiedString(tt.prefs)
			if got != tt.want {
				t.Errorf("ModifiedString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestModifiedString_TagWidth_Runes(t *testing.T) {
	raw := "02-08 12:12:09.629  3950  4005 D 日本語: ok"
	line := ParseLogLine(raw)
	prefs := OutputPrefs{
		Columns:  Columns{Tag: true, Message: true},
		TagWidth: 2,
	}
	got := line.ModifiedString(prefs)
	want := "日…: ok"
	if got != want {
		t.Errorf("ModifiedString() = %q, want %q", got, want)
	}
}

func TestModifiedString_TagWidth_NarrowTruncate(t *testing.T) {
	long := ParseLogLine("02-08 12:12:09.629  3950  4005 D Abcde: msg")
	t.Run("Width1_firstRune", func(t *testing.T) {
		got := long.ModifiedString(OutputPrefs{
			Columns:  Columns{Tag: true, Message: true},
			TagWidth: 1,
		})
		want := "A: msg"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
	t.Run("Width2_firstRuneAndEllipsis", func(t *testing.T) {
		got := long.ModifiedString(OutputPrefs{
			Columns:  Columns{Tag: true, Message: true},
			TagWidth: 2,
		})
		want := "A…: msg"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestPrefixWidth_MatchesModifiedString_TagWidth(t *testing.T) {
	line := ParseLogLine("02-08 12:12:09.629  3950  4005 D BusinessScope: padded message text")
	prefs := OutputPrefs{
		Columns:  Columns{Tag: true, Message: true},
		TagWidth: 6,
	}
	pw := line.PrefixWidth(prefs)
	full := line.ModifiedString(prefs)
	if pw > len(full) {
		t.Fatalf("PrefixWidth(%d) > len(ModifiedString)(%d)", pw, len(full))
	}
	if got := full[pw:]; got != line.Message {
		t.Errorf("full[PrefixWidth:] = %q, want %q (full=%q)", got, line.Message, full)
	}
	if want := "Bu…ope: padded message text"; full != want {
		t.Errorf("ModifiedString() = %q, want %q", full, want)
	}
}
