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

	// Compare Prefs
	if got.Prefs.Format != want.Prefs.Format {
		t.Errorf("Prefs.Format = %q, want %q", got.Prefs.Format, want.Prefs.Format)
	}
	if !slicesEqual(got.Prefs.Modifiers, want.Prefs.Modifiers) {
		t.Errorf("Prefs.Modifiers = %v, want %v", got.Prefs.Modifiers, want.Prefs.Modifiers)
	}

	// Compare Session
	if got.Session.DeviceId != want.Session.DeviceId {
		t.Errorf("Session.DeviceID = %q, want %q", got.Session.DeviceId, want.Session.DeviceId)
	}
	if got.Session.Pkg != want.Session.Pkg {
		t.Errorf("Session.Pkg = %q, want %q", got.Session.Pkg, want.Session.Pkg)
	}
	if got.Session.Tag != want.Session.Tag {
		t.Errorf("Session.Tag = %q, want %q", got.Session.Tag, want.Session.Tag)
	}
	if got.Session.Txt != want.Session.Txt {
		t.Errorf("Session.Txt = %q, want %q", got.Session.Txt, want.Session.Txt)
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

// createTestFile creates a file with given content in a temp directory.
func createTestFile(t *testing.T, dir, content string) *os.File {
	t.Helper()

	filePath := filepath.Join(dir, "test_config.json")
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}

	return file
}

// Test functions

func TestDefaultConfig(t *testing.T) {
	got := DefaultConfig()

	// Test Prefs defaults
	t.Run("Prefs", func(t *testing.T) {
		if got.Prefs.Format != "time" {
			t.Errorf("Prefs.Format = %q, want %q", got.Prefs.Format, "time")
		}
		wantModifiers := []string{"color"}
		if !slicesEqual(got.Prefs.Modifiers, wantModifiers) {
			t.Errorf("Prefs.Modifiers = %v, want %v", got.Prefs.Modifiers, wantModifiers)
		}
	})

	// Test Session defaults
	t.Run("Session", func(t *testing.T) {
		if got.Session.DeviceId != "" {
			t.Errorf("Session.DeviceID = %q, want empty string", got.Session.DeviceId)
		}
		if got.Session.Pkg != "" {
			t.Errorf("Session.Pkg = %q, want empty string", got.Session.Pkg)
		}
		if got.Session.Tag != "" {
			t.Errorf("Session.Tag = %q, want empty string", got.Session.Tag)
		}
		if got.Session.Txt != "" {
			t.Errorf("Session.Txt = %q, want empty string", got.Session.Txt)
		}
	})

	// Test multiple calls return equal configs
	t.Run("Consistency", func(t *testing.T) {
		config1 := DefaultConfig()
		config2 := DefaultConfig()
		compareConfigs(t, config1, config2)
	})
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name       string
		jsonData   string
		wantConfig Config
	}{
		{
			name: "ValidDefaultConfig",
			jsonData: `{
  "preferences": {
    "log_format": "time",
    "log_modifiers": ["color"]
  },
  "session": {
    "device_id": "",
    "package_name": "",
    "log_tag": "",
    "log_text": ""
  }
}`,
			wantConfig: DefaultConfig(),
		},
		{
			name: "ValidCustomConfig",
			jsonData: `{
  "preferences": {
    "log_format": "json",
    "log_modifiers": ["color", "timestamp"]
  },
  "session": {
    "device_id": "emulator-5554",
    "package_name": "com.example.app",
    "log_tag": "MyTag",
    "log_text": "search"
  }
}`,
			wantConfig: Config{
				Prefs: Prefs{
					Format:    "json",
					Modifiers: []string{"color", "timestamp"},
				},
				Session: Session{
					DeviceId: "emulator-5554",
					Pkg:      "com.example.app",
					Tag:      "MyTag",
					Txt:      "search",
				},
			},
		},
		{
			name:       "CompactJSON",
			jsonData:   `{"preferences":{"log_format":"time","log_modifiers":["color"]},"session":{"device_id":"","package_name":"","log_tag":"","log_text":""}}`,
			wantConfig: DefaultConfig(),
		},
		{
			name: "EmptyModifiersArray",
			jsonData: `{
  "preferences": {
    "log_format": "time",
    "log_modifiers": []
  },
  "session": {
    "device_id": "",
    "package_name": "",
    "log_tag": "",
    "log_text": ""
  }
}`,
			wantConfig: Config{
				Prefs: Prefs{
					Format:    "time",
					Modifiers: []string{},
				},
				Session: Session{
					DeviceId: "",
					Pkg:      "",
					Tag:      "",
					Txt:      "",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			file := createTestFile(t, tempDir, tt.jsonData)
			defer file.Close()

			got, err := Load(file)
			if err != nil {
				t.Fatalf("Load() error = %v, want nil", err)
			}

			compareConfigs(t, got, tt.wantConfig)
		})
	}
}

func TestLoadConfig_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		jsonData    string
		wantError   bool
		skipExtraOK bool // ExtraFields case doesn't error
	}{
		{
			name:      "EmptyFile",
			jsonData:  "",
			wantError: true,
		},
		{
			name:      "InvalidJSON",
			jsonData:  `{this is not valid json}`,
			wantError: true,
		},
		{
			name:      "IncompleteJSON",
			jsonData:  `{"preferences": {`,
			wantError: true,
		},
		{
			name: "WrongTypes",
			jsonData: `{
  "preferences": {
    "log_format": 123,
    "log_modifiers": "not an array"
  },
  "session": {
    "device_id": 456,
    "package_name": true,
    "log_tag": {},
    "log_text": null
  }
}`,
			wantError: true,
		},
		{
			name:      "MissingFields",
			jsonData:  `{"preferences": {}}`,
			wantError: false, // JSON decoder fills with zero values
		},
		{
			name: "NullValues",
			jsonData: `{
  "preferences": null,
  "session": null
}`,
			wantError: false, // JSON decoder fills with zero values
		},
		{
			name: "ExtraFields",
			jsonData: `{
  "preferences": {
    "log_format": "time",
    "log_modifiers": ["color"],
    "extra_field": "should be ignored"
  },
  "session": {
    "device_id": "",
    "package_name": "",
    "log_tag": "",
    "log_text": "",
    "another_extra": 123
  }
}`,
			wantError:   false,
			skipExtraOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			file := createTestFile(t, tempDir, tt.jsonData)
			defer file.Close()

			got, err := Load(file)

			// Check error expectation
			if tt.wantError && err == nil {
				t.Errorf("Load() error = nil, want error for %s", tt.name)
			}
			if !tt.wantError && !tt.skipExtraOK && err != nil {
				t.Errorf("Load() error = %v, want nil", err)
			}

			// Even on error, should return default config
			if err != nil {
				want := DefaultConfig()
				compareConfigs(t, got, want)
			}
		})
	}

	t.Run("ClosedFile", func(t *testing.T) {
		tempDir := t.TempDir()
		file := createTestFile(t, tempDir, `{}`)
		file.Close() // Close before reading

		_, err := Load(file)
		if err == nil {
			t.Error("Load() with closed file error = nil, want error")
		}
	})
}

func TestSaveConfig(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name:   "DefaultConfig",
			config: DefaultConfig(),
		},
		{
			name: "CustomConfig",
			config: Config{
				Prefs: Prefs{
					Format:    "brief",
					Modifiers: []string{"color", "epoch"},
				},
				Session: Session{
					DeviceId: "emulator-5554",
					Pkg:      "com.example.app",
					Tag:      "MyTag",
					Txt:      "error",
				},
			},
		},
		{
			name: "EmptySession",
			config: Config{
				Prefs: Prefs{
					Format:    "time",
					Modifiers: []string{"color"},
				},
				Session: Session{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			path := filepath.Join(tempDir, "export.json")

			err := Save(tt.config, path)
			if err != nil {
				t.Fatalf("Save() error = %v, want nil", err)
			}

			// Read back and verify
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("Failed to read saved file: %v", err)
			}

			var got Config
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("Failed to unmarshal saved file: %v", err)
			}

			compareConfigs(t, got, tt.config)
		})
	}
}

func TestSaveConfig_ErrorCases(t *testing.T) {
	t.Run("InvalidPath", func(t *testing.T) {
		err := Save(DefaultConfig(), "/nonexistent/dir/config.json")
		if err == nil {
			t.Error("Save() with invalid path error = nil, want error")
		}
	})
}

func TestSaveConfig_PrettyJSON(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "export.json")

	err := Save(DefaultConfig(), path)
	if err != nil {
		t.Fatalf("Save() error = %v, want nil", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "\n") {
		t.Error("Save() output is not indented, expected pretty-printed JSON")
	}
	if !strings.Contains(content, "  ") {
		t.Error("Save() output missing indentation")
	}
}

func TestConfig_JSONFieldNames(t *testing.T) {
	// Verify JSON tag mappings are correct
	t.Run("PrefsJSONTags", func(t *testing.T) {
		config := Config{
			Prefs: Prefs{
				Format:    "test",
				Modifiers: []string{"mod1"},
			},
			Session: Session{
				DeviceId: "device",
				Pkg:      "pkg",
				Tag:      "tag",
				Txt:      "txt",
			},
		}

		data, err := json.Marshal(config)
		if err != nil {
			t.Fatalf("json.Marshal() error = %v", err)
		}

		jsonStr := string(data)

		// Verify JSON field names
		expectedFields := []string{
			`"log_format"`,
			`"log_modifiers"`,
			`"device_id"`,
			`"package_name"`,
			`"log_tag"`,
			`"log_text"`,
		}

		for _, field := range expectedFields {
			if !strings.Contains(jsonStr, field) {
				t.Errorf("JSON missing expected field %s", field)
			}
		}
	})
}
