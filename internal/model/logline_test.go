package model

import "testing"

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
