package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Helper functions

// boolPtr returns a pointer to a bool value, useful for Config.Display.Color.
func boolPtr(v bool) *bool {
	return &v
}

// compareBoolPtr compares two *bool values and reports an error if they differ.
func compareBoolPtr(t *testing.T, field string, got, want *bool) {
	t.Helper()
	if got == nil && want == nil {
		return
	}
	if got == nil || want == nil {
		t.Errorf("%s = %v, want %v", field, got, want)
		return
	}
	if *got != *want {
		t.Errorf("%s = %v, want %v", field, *got, *want)
	}
}

// compareColumns compares two *Columns values and reports detailed errors.
func compareColumns(t *testing.T, got, want *Columns) {
	t.Helper()
	if got == nil && want == nil {
		return
	}
	if got == nil || want == nil {
		t.Errorf("Display.Columns = %v, want %v", got, want)
		return
	}
	compareBoolPtr(t, "Display.Columns.Date", got.Date, want.Date)
	compareBoolPtr(t, "Display.Columns.Time", got.Time, want.Time)
	compareBoolPtr(t, "Display.Columns.PID", got.PID, want.PID)
	compareBoolPtr(t, "Display.Columns.TID", got.TID, want.TID)
	compareBoolPtr(t, "Display.Columns.Level", got.Level, want.Level)
	compareBoolPtr(t, "Display.Columns.Tag", got.Tag, want.Tag)
	compareBoolPtr(t, "Display.Columns.Message", got.Message, want.Message)
}

// compareConfigs compares two Config structs field by field and reports detailed errors.
func compareConfigs(t *testing.T, got, want Config) {
	t.Helper()

	// Compare Display
	compareBoolPtr(t, "Display.Color", got.Display.Color, want.Display.Color)
	compareBoolPtr(t, "Display.Wrap", got.Display.Wrap, want.Display.Wrap)
	compareColumns(t, got.Display.Columns, want.Display.Columns)

	// Compare Filter
	if got.Filter.Pkg != want.Filter.Pkg {
		t.Errorf("Filter.Pkg = %+v, want %+v", got.Filter.Pkg, want.Filter.Pkg)
	}
	if got.Filter.Tag != want.Filter.Tag {
		t.Errorf("Filter.Tag = %+v, want %+v", got.Filter.Tag, want.Filter.Tag)
	}
	if got.Filter.Txt != want.Filter.Txt {
		t.Errorf("Filter.Txt = %+v, want %+v", got.Filter.Txt, want.Filter.Txt)
	}
}

// writeTestFile creates a JSON file with given content in a directory.
func writeTestFile(t *testing.T, dir, name, content string) {
	t.Helper()

	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", dir, err)
	}

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file %s: %v", path, err)
	}
}

// Test functions

func TestDefaultConfig(t *testing.T) {
	got := DefaultConfig()

	t.Run("Display", func(t *testing.T) {
		if got.Display.Color == nil {
			t.Fatalf("Display.Color = nil, want non-nil")
		}
		if *got.Display.Color != true {
			t.Errorf("Display.Color = %v, want true", *got.Display.Color)
		}
		if got.Display.Wrap == nil {
			t.Fatalf("Display.Wrap = nil, want non-nil")
		}
		if *got.Display.Wrap != true {
			t.Errorf("Display.Wrap = %v, want true", *got.Display.Wrap)
		}
	})

	t.Run("Columns", func(t *testing.T) {
		cols := got.Display.Columns
		if cols == nil {
			t.Fatalf("Display.Columns = nil, want non-nil")
		}
		checks := []struct {
			name string
			got  *bool
			want bool
		}{
			{"Date", cols.Date, false},
			{"Time", cols.Time, true},
			{"PID", cols.PID, false},
			{"TID", cols.TID, false},
			{"Level", cols.Level, true},
			{"Tag", cols.Tag, true},
			{"Message", cols.Message, true},
		}
		for _, c := range checks {
			if c.got == nil {
				t.Errorf("Columns.%s = nil, want %v", c.name, c.want)
			} else if *c.got != c.want {
				t.Errorf("Columns.%s = %v, want %v", c.name, *c.got, c.want)
			}
		}
	})

	t.Run("Filter", func(t *testing.T) {
		if !got.Filter.Pkg.IsZero() {
			t.Errorf("Filter.Pkg = %+v, want zero value", got.Filter.Pkg)
		}
		if !got.Filter.Tag.IsZero() {
			t.Errorf("Filter.Tag = %+v, want zero value", got.Filter.Tag)
		}
		if !got.Filter.Txt.IsZero() {
			t.Errorf("Filter.Txt = %+v, want zero value", got.Filter.Txt)
		}
	})

	t.Run("Consistency", func(t *testing.T) {
		config1 := DefaultConfig()
		config2 := DefaultConfig()
		compareConfigs(t, config1, config2)
	})
}

func TestTextFilter_IsZero(t *testing.T) {
	tests := []struct {
		name string
		f    TextFilter
		want bool
	}{
		{name: "Empty", f: TextFilter{}, want: true},
		{name: "EmptyValue", f: TextFilter{Value: ""}, want: true},
		{name: "NonEmpty", f: TextFilter{Value: "hello"}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.IsZero(); got != tt.want {
				t.Errorf("IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTextFilter_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    TextFilter
		wantErr bool
	}{
		{name: "PlainString", json: `"hello"`, want: TextFilter{Value: "hello"}},
		{name: "EmptyString", json: `""`, want: TextFilter{Value: ""}},
		{name: "Object", json: `{"value":"hello"}`, want: TextFilter{Value: "hello"}},
		{name: "EmptyObject", json: `{}`, want: TextFilter{}},
		{name: "ObjectEmptyValue", json: `{"value":""}`, want: TextFilter{Value: ""}},
		{name: "ObjectWithModeExact", json: `{"value":"hello","mode":"exact"}`, want: TextFilter{Value: "hello", Mode: "exact"}},
		{name: "ObjectWithModeRegex", json: `{"value":"err.*","mode":"regex"}`, want: TextFilter{Value: "err.*", Mode: "regex"}},
		{name: "ObjectWithModeContains", json: `{"value":"hello","mode":"contains"}`, want: TextFilter{Value: "hello", Mode: "contains"}},
		{name: "ObjectWithModeOmitted", json: `{"value":"hello"}`, want: TextFilter{Value: "hello"}},
		{name: "ObjectExtraFields", json: `{"value":"hello","unknown":"ignored"}`, want: TextFilter{Value: "hello"}},
		{name: "InvalidJSON", json: `{bad`, wantErr: true},
		{name: "Number", json: `123`, wantErr: true},
		{name: "Null", json: `null`, want: TextFilter{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got TextFilter
			err := json.Unmarshal([]byte(tt.json), &got)
			if tt.wantErr {
				if err == nil {
					t.Errorf("UnmarshalJSON() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("UnmarshalJSON() error = %v, want nil", err)
			}
			if got != tt.want {
				t.Errorf("UnmarshalJSON() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestTextFilter_MarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		f    TextFilter
		want string
	}{
		{name: "WithValue", f: TextFilter{Value: "hello"}, want: `"hello"`},
		{name: "Empty", f: TextFilter{}, want: `""`},
		{name: "ModeContains", f: TextFilter{Value: "hello", Mode: "contains"}, want: `"hello"`},
		{name: "ModeEmpty", f: TextFilter{Value: "hello", Mode: ""}, want: `"hello"`},
		{name: "ModeExact", f: TextFilter{Value: "hello", Mode: "exact"}, want: `{"value":"hello","mode":"exact"}`},
		{name: "ModeRegex", f: TextFilter{Value: "err.*", Mode: "regex"}, want: `{"value":"err.*","mode":"regex"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.f)
			if err != nil {
				t.Fatalf("MarshalJSON() error = %v", err)
			}
			if string(data) != tt.want {
				t.Errorf("MarshalJSON() = %s, want %s", string(data), tt.want)
			}
		})
	}
}

func TestLoadFile(t *testing.T) {
	tests := []struct {
		name       string
		jsonData   string
		wantConfig Config
	}{
		{
			name: "FullConfig",
			jsonData: `{
  "display": {
    "color": false,
    "wrap": false,
    "columns": {
      "date": true, "time": true, "pid": true, "tid": true,
      "level": true, "tag": true, "message": true
    }
  },
  "filter": {
    "package_name": "com.example.app",
    "log_tag": "MyTag",
    "log_text": "error"
  }
}`,
			wantConfig: Config{
				Display: Display{
					Color: boolPtr(false),
					Wrap:  boolPtr(false),
					Columns: &Columns{
						Date: boolPtr(true), Time: boolPtr(true), PID: boolPtr(true), TID: boolPtr(true),
						Level: boolPtr(true), Tag: boolPtr(true), Message: boolPtr(true),
					},
				},
				Filter: Filter{
					Pkg: TextFilter{Value: "com.example.app"},
					Tag: TextFilter{Value: "MyTag"},
					Txt: TextFilter{Value: "error"},
				},
			},
		},
		{
			name: "FilterWithObjectForm",
			jsonData: `{
  "filter": {
    "log_tag": {"value": "MyTag"},
    "log_text": {"value": "error"}
  }
}`,
			wantConfig: Config{
				Filter: Filter{
					Tag: TextFilter{Value: "MyTag"},
					Txt: TextFilter{Value: "error"},
				},
			},
		},
		{
			name: "MixedFilterForms",
			jsonData: `{
  "filter": {
    "package_name": "com.example.app",
    "log_tag": {"value": "MyTag"},
    "log_text": "error"
  }
}`,
			wantConfig: Config{
				Filter: Filter{
					Pkg: TextFilter{Value: "com.example.app"},
					Tag: TextFilter{Value: "MyTag"},
					Txt: TextFilter{Value: "error"},
				},
			},
		},
		{
			name: "FilterWithModes",
			jsonData: `{
  "filter": {
    "package_name": {"value": "com.example.app", "mode": "exact"},
    "log_tag": {"value": "My.*Tag", "mode": "regex"},
    "log_text": "error"
  }
}`,
			wantConfig: Config{
				Filter: Filter{
					Pkg: TextFilter{Value: "com.example.app", Mode: "exact"},
					Tag: TextFilter{Value: "My.*Tag", Mode: "regex"},
					Txt: TextFilter{Value: "error"},
				},
			},
		},
		{
			name:       "DisplayColorTrue",
			jsonData:   `{"display": {"color": true}}`,
			wantConfig: Config{Display: Display{Color: boolPtr(true)}},
		},
		{
			name:       "EmptyObject",
			jsonData:   `{}`,
			wantConfig: Config{},
		},
		{
			name:       "DisplayColorFalse",
			jsonData:   `{"display": {"color": false}}`,
			wantConfig: Config{Display: Display{Color: boolPtr(false)}},
		},
		{
			name:       "CompactJSON",
			jsonData:   `{"display":{"color":true},"filter":{"package_name":"com.example.app"}}`,
			wantConfig: Config{Display: Display{Color: boolPtr(true)}, Filter: Filter{Pkg: TextFilter{Value: "com.example.app"}}},
		},
		{
			name:     "WrapTrue",
			jsonData: `{"display": {"wrap": true}}`,
			wantConfig: Config{
				Display: Display{Wrap: boolPtr(true)},
			},
		},
		{
			name:     "WrapFalse",
			jsonData: `{"display": {"wrap": false}}`,
			wantConfig: Config{
				Display: Display{Wrap: boolPtr(false)},
			},
		},
		{
			name: "FullColumns",
			jsonData: `{"display": {"columns": {
				"date": true, "time": false, "pid": true, "tid": true,
				"level": false, "tag": false, "message": true
			}}}`,
			wantConfig: Config{
				Display: Display{
					Columns: &Columns{
						Date: boolPtr(true), Time: boolPtr(false), PID: boolPtr(true), TID: boolPtr(true),
						Level: boolPtr(false), Tag: boolPtr(false), Message: boolPtr(true),
					},
				},
			},
		},
		{
			name:     "PartialColumns",
			jsonData: `{"display": {"columns": {"pid": true}}}`,
			wantConfig: Config{
				Display: Display{
					Columns: &Columns{PID: boolPtr(true)},
				},
			},
		},
		{
			name:     "WrapAndColumns",
			jsonData: `{"display": {"wrap": false, "columns": {"date": true, "tid": true}}}`,
			wantConfig: Config{
				Display: Display{
					Wrap:    boolPtr(false),
					Columns: &Columns{Date: boolPtr(true), TID: boolPtr(true)},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "config.json")
			if err := os.WriteFile(path, []byte(tt.jsonData), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			got, err := loadFile(path)
			if err != nil {
				t.Fatalf("loadFile() error = %v, want nil", err)
			}

			compareConfigs(t, got, tt.wantConfig)
		})
	}
}

func TestLoadFile_ErrorCases(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
	}{
		{name: "EmptyFile", jsonData: ""},
		{name: "InvalidJSON", jsonData: `{this is not valid json}`},
		{name: "IncompleteJSON", jsonData: `{"display": {`},
		{name: "WrongTypes", jsonData: `{"display": {"color": "yes"}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "config.json")
			if err := os.WriteFile(path, []byte(tt.jsonData), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			_, err := loadFile(path)
			if err == nil {
				t.Errorf("loadFile() error = nil, want error")
			}
		})
	}

	t.Run("FileNotFound", func(t *testing.T) {
		_, err := loadFile("/nonexistent/path/config.json")
		if err == nil {
			t.Error("loadFile() with missing file error = nil, want error")
		}
	})

	t.Run("ExtraFieldsAccepted", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.json")
		data := `{"display": {"color": true, "extra": "ignored"}, "unknown_section": {}}`
		if err := os.WriteFile(path, []byte(data), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		_, err := loadFile(path)
		if err != nil {
			t.Errorf("loadFile() with extra fields error = %v, want nil", err)
		}
	})

	t.Run("NullValues", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.json")
		data := `{"display": null, "filter": null}`
		if err := os.WriteFile(path, []byte(data), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		_, err := loadFile(path)
		if err != nil {
			t.Errorf("loadFile() with null values error = %v, want nil", err)
		}
	})
}

func TestMerge(t *testing.T) {
	tests := []struct {
		name    string
		base    Config
		overlay Config
		want    Config
	}{
		{
			name:    "EmptyOverlay",
			base:    DefaultConfig(),
			overlay: Config{},
			want:    DefaultConfig(),
		},
		{
			name: "FullOverlay",
			base: DefaultConfig(),
			overlay: Config{
				Display: Display{
					Color: boolPtr(false),
				},
				Filter: Filter{
					Pkg: TextFilter{Value: "com.example"},
					Tag: TextFilter{Value: "MyTag"},
					Txt: TextFilter{Value: "error"},
				},
			},
			want: func() Config {
				c := DefaultConfig()
				*c.Display.Color = false
				c.Filter.Pkg = TextFilter{Value: "com.example"}
				c.Filter.Tag = TextFilter{Value: "MyTag"}
				c.Filter.Txt = TextFilter{Value: "error"}
				return c
			}(),
		},
		{
			name: "PartialOverlay_ColorOnly",
			base: DefaultConfig(),
			overlay: Config{
				Display: Display{Color: boolPtr(false)},
			},
			want: func() Config {
				c := DefaultConfig()
				*c.Display.Color = false
				return c
			}(),
		},
		{
			name: "PartialOverlay_FilterOnly",
			base: DefaultConfig(),
			overlay: Config{
				Filter: Filter{
					Pkg: TextFilter{Value: "com.example"},
				},
			},
			want: func() Config {
				c := DefaultConfig()
				c.Filter.Pkg = TextFilter{Value: "com.example"}
				return c
			}(),
		},
		{
			name: "NilColorPreservesBase",
			base: Config{
				Display: Display{
					Color: boolPtr(true),
				},
			},
			overlay: Config{
				Display: Display{
					// Color is nil — should not override
				},
			},
			want: Config{
				Display: Display{
					Color: boolPtr(true),
				},
			},
		},
		{
			name: "ColorOverridesBase",
			base: Config{
				Display: Display{
					Color: boolPtr(true),
				},
			},
			overlay: Config{
				Display: Display{
					Color: boolPtr(false),
				},
			},
			want: Config{
				Display: Display{
					Color: boolPtr(false),
				},
			},
		},
		{
			name: "EmptyFilterDoesNotOverride",
			base: Config{
				Filter: Filter{
					Pkg: TextFilter{Value: "com.example"},
					Tag: TextFilter{Value: "MyTag"},
				},
			},
			overlay: Config{
				Filter: Filter{
					Txt: TextFilter{Value: "search"},
				},
			},
			want: Config{
				Filter: Filter{
					Pkg: TextFilter{Value: "com.example"},
					Tag: TextFilter{Value: "MyTag"},
					Txt: TextFilter{Value: "search"},
				},
			},
		},
		{
			name: "EmptyBase",
			base: Config{},
			overlay: Config{
				Display: Display{Color: boolPtr(false)},
				Filter:  Filter{Pkg: TextFilter{Value: "com.example"}},
			},
			want: Config{
				Display: Display{Color: boolPtr(false)},
				Filter:  Filter{Pkg: TextFilter{Value: "com.example"}},
			},
		},
		{
			name: "OverlayWithMode",
			base: Config{
				Filter: Filter{
					Pkg: TextFilter{Value: "com.base"},
				},
			},
			overlay: Config{
				Filter: Filter{
					Pkg: TextFilter{Value: "com.overlay", Mode: "exact"},
				},
			},
			want: Config{
				Filter: Filter{
					Pkg: TextFilter{Value: "com.overlay", Mode: "exact"},
				},
			},
		},
		{
			name: "ModePreservedFromBase",
			base: Config{
				Filter: Filter{
					Pkg: TextFilter{Value: "com.base", Mode: "regex"},
					Tag: TextFilter{Value: "BaseTag", Mode: "exact"},
				},
			},
			overlay: Config{
				Filter: Filter{
					Tag: TextFilter{Value: "OverlayTag"},
				},
			},
			want: Config{
				Filter: Filter{
					Pkg: TextFilter{Value: "com.base", Mode: "regex"},
					Tag: TextFilter{Value: "OverlayTag"},
				},
			},
		},
		{
			name: "WrapOverridesBase",
			base: Config{
				Display: Display{Wrap: boolPtr(true)},
			},
			overlay: Config{
				Display: Display{Wrap: boolPtr(false)},
			},
			want: Config{
				Display: Display{Wrap: boolPtr(false)},
			},
		},
		{
			name: "NilWrapPreservesBase",
			base: Config{
				Display: Display{Wrap: boolPtr(true)},
			},
			overlay: Config{},
			want: Config{
				Display: Display{Wrap: boolPtr(true)},
			},
		},
		{
			name: "PartialColumnsOverlay",
			base: DefaultConfig(),
			overlay: Config{
				Display: Display{
					Columns: &Columns{PID: boolPtr(true), Date: boolPtr(true)},
				},
			},
			want: func() Config {
				c := DefaultConfig()
				*c.Display.Columns.PID = true
				*c.Display.Columns.Date = true
				return c
			}(),
		},
		{
			name: "NilColumnsPreservesBase",
			base: DefaultConfig(),
			overlay: Config{
				Display: Display{Wrap: boolPtr(false)},
			},
			want: func() Config {
				c := DefaultConfig()
				*c.Display.Wrap = false
				return c
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := merge(tt.base, tt.overlay)
			compareConfigs(t, got, tt.want)
		})
	}
}

func TestResolve(t *testing.T) {
	// Save and restore the working directory since Resolve uses relative paths
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	t.Run("NoConfigFiles", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Chdir(dir); err != nil {
			t.Fatalf("Failed to chdir: %v", err)
		}
		defer os.Chdir(origDir)

		got, err := Resolve(nil)
		if err != nil {
			t.Errorf("Resolve(nil) error = %v, want nil", err)
		}

		compareConfigs(t, got, DefaultConfig())
	})

	t.Run("ProjectConfigOnly", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Chdir(dir); err != nil {
			t.Fatalf("Failed to chdir: %v", err)
		}
		defer os.Chdir(origDir)

		writeTestFile(t, filepath.Join(dir, ".lazylogcat"), "config.json", `{
			"display": {"color": false},
			"filter": {"package_name": "com.project"}
		}`)

		got, err := Resolve(nil)
		if err != nil {
			t.Errorf("Resolve(nil) error = %v, want nil", err)
		}

		want := DefaultConfig()
		*want.Display.Color = false
		want.Filter.Pkg = TextFilter{Value: "com.project"}
		compareConfigs(t, got, want)
	})

	t.Run("LocalOverridesProject", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Chdir(dir); err != nil {
			t.Fatalf("Failed to chdir: %v", err)
		}
		defer os.Chdir(origDir)

		writeTestFile(t, filepath.Join(dir, ".lazylogcat"), "config.json", `{
			"display": {"color": false},
			"filter": {"package_name": "com.project", "log_tag": "ProjectTag"}
		}`)

		writeTestFile(t, filepath.Join(dir, ".lazylogcat"), "config.local.json", `{
			"filter": {"package_name": "com.local"}
		}`)

		got, err := Resolve(nil)
		if err != nil {
			t.Errorf("Resolve(nil) error = %v, want nil", err)
		}

		want := DefaultConfig()
		*want.Display.Color = false                       // From project
		want.Filter.Pkg = TextFilter{Value: "com.local"}  // Overridden by local
		want.Filter.Tag = TextFilter{Value: "ProjectTag"} // From project
		compareConfigs(t, got, want)
	})

	t.Run("InvalidFileSkipped", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Chdir(dir); err != nil {
			t.Fatalf("Failed to chdir: %v", err)
		}
		defer os.Chdir(origDir)

		// Invalid project config
		writeTestFile(t, filepath.Join(dir, ".lazylogcat"), "config.json", `{invalid json}`)

		// Valid local config
		writeTestFile(t, filepath.Join(dir, ".lazylogcat"), "config.local.json", `{
			"filter": {"package_name": "com.local"}
		}`)

		got, err := Resolve(nil)
		if err == nil {
			t.Error("Resolve(nil) error = nil, want error for invalid file")
		}

		// Local config should still be applied over defaults
		want := DefaultConfig()
		want.Filter.Pkg = TextFilter{Value: "com.local"}
		compareConfigs(t, got, want)
	})

	t.Run("AllLayersPrecedence", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Chdir(dir); err != nil {
			t.Fatalf("Failed to chdir: %v", err)
		}
		defer os.Chdir(origDir)

		// Simulate global config using the actual XDG config directory path
		// We can't easily mock os.UserConfigDir, so we test project + local layers
		// which are sufficient to verify merge precedence.

		writeTestFile(t, filepath.Join(dir, ".lazylogcat"), "config.json", `{
			"display": {"color": false},
			"filter": {
				"package_name": "com.project",
				"log_tag": "ProjectTag",
				"log_text": "project_text"
			}
		}`)

		writeTestFile(t, filepath.Join(dir, ".lazylogcat"), "config.local.json", `{
			"display": {"color": true},
			"filter": {"log_tag": "LocalTag"}
		}`)

		got, err := Resolve(nil)
		if err != nil {
			t.Errorf("Resolve(nil) error = %v, want nil", err)
		}

		want := DefaultConfig()
		*want.Display.Color = true                          // Overridden by local
		want.Filter.Pkg = TextFilter{Value: "com.project"}  // From project
		want.Filter.Tag = TextFilter{Value: "LocalTag"}     // Overridden by local
		want.Filter.Txt = TextFilter{Value: "project_text"} // From project
		compareConfigs(t, got, want)
	})

	t.Run("FilterModesPersisted", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Chdir(dir); err != nil {
			t.Fatalf("Failed to chdir: %v", err)
		}
		defer os.Chdir(origDir)

		writeTestFile(t, filepath.Join(dir, ".lazylogcat"), "config.json", `{
			"filter": {
				"package_name": {"value": "com.project", "mode": "exact"},
				"log_tag": {"value": "Tag.*", "mode": "regex"}
			}
		}`)

		writeTestFile(t, filepath.Join(dir, ".lazylogcat"), "config.local.json", `{
			"filter": {
				"log_tag": {"value": "LocalTag", "mode": "exact"}
			}
		}`)

		got, err := Resolve(nil)
		if err != nil {
			t.Errorf("Resolve(nil) error = %v, want nil", err)
		}

		want := DefaultConfig()
		want.Filter.Pkg = TextFilter{Value: "com.project", Mode: "exact"}
		want.Filter.Tag = TextFilter{Value: "LocalTag", Mode: "exact"} // Overridden by local
		compareConfigs(t, got, want)
	})

	t.Run("ColumnsPartialOverride", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Chdir(dir); err != nil {
			t.Fatalf("Failed to chdir: %v", err)
		}
		defer os.Chdir(origDir)

		// Project enables PID and Date columns
		writeTestFile(t, filepath.Join(dir, ".lazylogcat"), "config.json", `{
			"display": {"columns": {"pid": true, "date": true}}
		}`)

		// Local disables wrap and re-disables Date
		writeTestFile(t, filepath.Join(dir, ".lazylogcat"), "config.local.json", `{
			"display": {"wrap": false, "columns": {"date": false}}
		}`)

		got, err := Resolve(nil)
		if err != nil {
			t.Errorf("Resolve(nil) error = %v, want nil", err)
		}

		want := DefaultConfig()
		*want.Display.Wrap = false
		*want.Display.Columns.PID = true   // From project
		*want.Display.Columns.Date = false // Overridden by local
		compareConfigs(t, got, want)
	})
}

func TestConfig_String(t *testing.T) {
	cfg := DefaultConfig()
	s := cfg.String()

	if s == "{}" {
		t.Error("String() returned empty object for non-empty config")
	}
	if !strings.Contains(s, `"display"`) {
		t.Error("String() missing display section")
	}
}

func TestConfig_JSONFieldNames(t *testing.T) {
	config := Config{
		Display: Display{
			Color: boolPtr(true),
			Wrap:  boolPtr(true),
			Columns: &Columns{
				Date: boolPtr(true), Time: boolPtr(true), PID: boolPtr(true), TID: boolPtr(true),
				Level: boolPtr(true), Tag: boolPtr(true), Message: boolPtr(true),
			},
		},
		Filter: Filter{
			Pkg: TextFilter{Value: "pkg"},
			Tag: TextFilter{Value: "tag"},
			Txt: TextFilter{Value: "txt"},
		},
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	jsonStr := string(data)

	expectedFields := []string{
		`"display"`,
		`"filter"`,
		`"color"`,
		`"wrap"`,
		`"columns"`,
		`"date"`,
		`"time"`,
		`"pid"`,
		`"tid"`,
		`"level"`,
		`"tag"`,
		`"message"`,
		`"package_name"`,
		`"log_tag"`,
		`"log_text"`,
	}

	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("JSON missing expected field %s in: %s", field, jsonStr)
		}
	}

	// Verify old field names are NOT present
	oldFields := []string{
		`"preferences"`,
		`"session"`,
		`"device_id"`,
		`"log_format"`,
		`"log_modifiers"`,
	}

	for _, field := range oldFields {
		if strings.Contains(jsonStr, field) {
			t.Errorf("JSON contains old field %s in: %s", field, jsonStr)
		}
	}
}
