package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Helper functions

// compareConfigs compares two Config structs field by field and reports detailed errors.
func compareConfigs(t *testing.T, got, want Config) {
	t.Helper()

	// Compare Display
	if got.Display.Format != want.Display.Format {
		t.Errorf("Display.Format = %q, want %q", got.Display.Format, want.Display.Format)
	}
	if !slicesEqual(got.Display.Modifiers, want.Display.Modifiers) {
		t.Errorf("Display.Modifiers = %v, want %v", got.Display.Modifiers, want.Display.Modifiers)
	}

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

// slicesEqual compares two string slices for equality.
func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
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
		if got.Display.Format != "time" {
			t.Errorf("Display.Format = %q, want %q", got.Display.Format, "time")
		}
		wantModifiers := []string{"color"}
		if !slicesEqual(got.Display.Modifiers, wantModifiers) {
			t.Errorf("Display.Modifiers = %v, want %v", got.Display.Modifiers, wantModifiers)
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
    "log_format": "json",
    "log_modifiers": ["color", "epoch"]
  },
  "filter": {
    "package_name": "com.example.app",
    "log_tag": "MyTag",
    "log_text": "error"
  }
}`,
			wantConfig: Config{
				Display: Display{
					Format:    "json",
					Modifiers: []string{"color", "epoch"},
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
			name:       "DisplayOnly",
			jsonData:   `{"display": {"log_format": "brief"}}`,
			wantConfig: Config{Display: Display{Format: "brief"}},
		},
		{
			name:       "EmptyObject",
			jsonData:   `{}`,
			wantConfig: Config{},
		},
		{
			name: "EmptyModifiersArray",
			jsonData: `{
  "display": {
    "log_format": "time",
    "log_modifiers": []
  }
}`,
			wantConfig: Config{
				Display: Display{
					Format:    "time",
					Modifiers: []string{},
				},
			},
		},
		{
			name:       "CompactJSON",
			jsonData:   `{"display":{"log_format":"time","log_modifiers":["color"]},"filter":{"package_name":"com.example.app"}}`,
			wantConfig: Config{Display: Display{Format: "time", Modifiers: []string{"color"}}, Filter: Filter{Pkg: TextFilter{Value: "com.example.app"}}},
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
		{name: "WrongTypes", jsonData: `{"display": {"log_format": 123}}`},
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
		data := `{"display": {"log_format": "time", "extra": "ignored"}, "unknown_section": {}}`
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
					Format:    "json",
					Modifiers: []string{"epoch", "uid"},
				},
				Filter: Filter{
					Pkg: TextFilter{Value: "com.example"},
					Tag: TextFilter{Value: "MyTag"},
					Txt: TextFilter{Value: "error"},
				},
			},
			want: Config{
				Display: Display{
					Format:    "json",
					Modifiers: []string{"epoch", "uid"},
				},
				Filter: Filter{
					Pkg: TextFilter{Value: "com.example"},
					Tag: TextFilter{Value: "MyTag"},
					Txt: TextFilter{Value: "error"},
				},
			},
		},
		{
			name: "PartialOverlay_FormatOnly",
			base: DefaultConfig(),
			overlay: Config{
				Display: Display{Format: "brief"},
			},
			want: Config{
				Display: Display{
					Format:    "brief",
					Modifiers: []string{"color"},
				},
			},
		},
		{
			name: "PartialOverlay_FilterOnly",
			base: DefaultConfig(),
			overlay: Config{
				Filter: Filter{
					Pkg: TextFilter{Value: "com.example"},
				},
			},
			want: Config{
				Display: Display{
					Format:    "time",
					Modifiers: []string{"color"},
				},
				Filter: Filter{
					Pkg: TextFilter{Value: "com.example"},
				},
			},
		},
		{
			name: "EmptySliceClearsBase",
			base: Config{
				Display: Display{
					Format:    "time",
					Modifiers: []string{"color"},
				},
			},
			overlay: Config{
				Display: Display{
					Modifiers: []string{},
				},
			},
			want: Config{
				Display: Display{
					Format:    "time",
					Modifiers: []string{},
				},
			},
		},
		{
			name: "NilSlicePreservesBase",
			base: Config{
				Display: Display{
					Format:    "time",
					Modifiers: []string{"color"},
				},
			},
			overlay: Config{
				Display: Display{
					Format: "brief",
					// Modifiers is nil — should not override
				},
			},
			want: Config{
				Display: Display{
					Format:    "brief",
					Modifiers: []string{"color"},
				},
			},
		},
		{
			name: "SliceReplacement",
			base: Config{
				Display: Display{
					Modifiers: []string{"color"},
				},
			},
			overlay: Config{
				Display: Display{
					Modifiers: []string{"epoch", "uid"},
				},
			},
			want: Config{
				Display: Display{
					Modifiers: []string{"epoch", "uid"},
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
				Display: Display{Format: "json"},
				Filter:  Filter{Pkg: TextFilter{Value: "com.example"}},
			},
			want: Config{
				Display: Display{Format: "json"},
				Filter:  Filter{Pkg: TextFilter{Value: "com.example"}},
			},
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

		got, err := Resolve()
		if err != nil {
			t.Errorf("Resolve() error = %v, want nil", err)
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
			"display": {"log_format": "brief"},
			"filter": {"package_name": "com.project"}
		}`)

		got, err := Resolve()
		if err != nil {
			t.Errorf("Resolve() error = %v, want nil", err)
		}

		want := Config{
			Display: Display{
				Format:    "brief",
				Modifiers: []string{"color"}, // From defaults
			},
			Filter: Filter{
				Pkg: TextFilter{Value: "com.project"},
			},
		}
		compareConfigs(t, got, want)
	})

	t.Run("LocalOverridesProject", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Chdir(dir); err != nil {
			t.Fatalf("Failed to chdir: %v", err)
		}
		defer os.Chdir(origDir)

		writeTestFile(t, filepath.Join(dir, ".lazylogcat"), "config.json", `{
			"display": {"log_format": "brief"},
			"filter": {"package_name": "com.project", "log_tag": "ProjectTag"}
		}`)

		writeTestFile(t, filepath.Join(dir, ".lazylogcat"), "config.local.json", `{
			"filter": {"package_name": "com.local"}
		}`)

		got, err := Resolve()
		if err != nil {
			t.Errorf("Resolve() error = %v, want nil", err)
		}

		want := Config{
			Display: Display{
				Format:    "brief",           // From project
				Modifiers: []string{"color"}, // From defaults
			},
			Filter: Filter{
				Pkg: TextFilter{Value: "com.local"},  // Overridden by local
				Tag: TextFilter{Value: "ProjectTag"}, // From project
			},
		}
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

		got, err := Resolve()
		if err == nil {
			t.Error("Resolve() error = nil, want error for invalid file")
		}

		// Local config should still be applied over defaults
		want := Config{
			Display: Display{
				Format:    "time",
				Modifiers: []string{"color"},
			},
			Filter: Filter{
				Pkg: TextFilter{Value: "com.local"},
			},
		}
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
			"display": {"log_format": "brief", "log_modifiers": ["epoch"]},
			"filter": {
				"package_name": "com.project",
				"log_tag": "ProjectTag",
				"log_text": "project_text"
			}
		}`)

		writeTestFile(t, filepath.Join(dir, ".lazylogcat"), "config.local.json", `{
			"display": {"log_format": "json"},
			"filter": {"log_tag": "LocalTag"}
		}`)

		got, err := Resolve()
		if err != nil {
			t.Errorf("Resolve() error = %v, want nil", err)
		}

		want := Config{
			Display: Display{
				Format:    "json",            // Overridden by local
				Modifiers: []string{"epoch"}, // From project (local didn't set modifiers)
			},
			Filter: Filter{
				Pkg: TextFilter{Value: "com.project"},  // From project
				Tag: TextFilter{Value: "LocalTag"},     // Overridden by local
				Txt: TextFilter{Value: "project_text"}, // From project
			},
		}
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
			Format:    "test",
			Modifiers: []string{"mod1"},
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
		`"log_format"`,
		`"log_modifiers"`,
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
	}

	for _, field := range oldFields {
		if strings.Contains(jsonStr, field) {
			t.Errorf("JSON contains old field %s in: %s", field, jsonStr)
		}
	}
}
