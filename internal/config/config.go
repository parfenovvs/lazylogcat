package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// Config represents the application's configuration structure.
// It includes display settings and filter parameters.
type Config struct {
	Display Display `json:"display"`
	Filter  Filter  `json:"filter"`
}

// Display holds user preferences for log output appearance.
type Display struct {
	Format    string   `json:"log_format,omitempty"`
	Modifiers []string `json:"log_modifiers,omitempty"`
}

// Filter holds log filtering parameters.
type Filter struct {
	Pkg TextFilter `json:"package_name,omitempty"`
	Tag TextFilter `json:"log_tag,omitempty"`
	Txt TextFilter `json:"log_text,omitempty"`
}

// TextFilter represents a filter field that currently holds a value,
// and is structured to support future extensions (e.g. regex mode).
// It accepts both a plain string and an object form in JSON:
//
//	"log_tag": "MyTag"
//	"log_tag": { "value": "MyTag" }
type TextFilter struct {
	Value string `json:"value"`
}

// IsZero returns true if the TextFilter has no meaningful value set.
func (f TextFilter) IsZero() bool {
	return f.Value == ""
}

// UnmarshalJSON supports both plain string and object forms.
func (f *TextFilter) UnmarshalJSON(data []byte) error {
	// Try plain string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		f.Value = s
		return nil
	}
	// Fall back to object form
	type alias TextFilter
	return json.Unmarshal(data, (*alias)(f))
}

// MarshalJSON outputs the short string form when only Value is set.
func (f TextFilter) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.Value)
}

// DefaultConfig returns the default configuration with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Display: Display{
			Format:    "time",
			Modifiers: []string{"color"},
		},
	}
}

// String returns a JSON string representation of the config for logging.
func (c *Config) String() string {
	data, err := json.Marshal(c)
	if err != nil {
		return "{}"
	}
	return string(data)
}

// Resolve discovers and merges configuration from all layers in order:
//  1. Default configuration
//  2. Global user config: ~/.config/lazylogcat/config.json
//  3. Project config: .lazylogcat/config.json
//  4. Local override: .lazylogcat/config.local.json
//
// Returns the merged config and any non-fatal errors encountered during loading.
// Always returns a usable config, even if some files fail to load.
func Resolve() (Config, error) {
	cfg := DefaultConfig()
	var errs []error

	for _, path := range configPaths() {
		overlay, err := loadFile(path)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				slog.Warn("Failed to load config file, skipping", "path", path, "error", err)
				errs = append(errs, fmt.Errorf("%s: %w", path, err))
			}
			continue
		}
		slog.Debug("Loaded config layer", "path", path)
		cfg = merge(cfg, overlay)
	}

	return cfg, errors.Join(errs...)
}

// configPaths returns the ordered list of config file paths to check.
func configPaths() []string {
	var paths []string

	// Layer 1: Global user config (~/.config/lazylogcat/config.json)
	if dir, err := os.UserConfigDir(); err == nil {
		paths = append(paths, filepath.Join(dir, "lazylogcat", "config.json"))
	}

	// Layer 2: Project config (.lazylogcat/config.json)
	paths = append(paths, filepath.Join(".lazylogcat", "config.json"))

	// Layer 3: Local override (.lazylogcat/config.local.json)
	paths = append(paths, filepath.Join(".lazylogcat", "config.local.json"))

	return paths
}

// loadFile reads and decodes a single JSON config file.
// Returns os.ErrNotExist if the file does not exist.
func loadFile(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to decode config file: %w", err)
	}

	return cfg, nil
}

// merge applies non-zero fields from overlay on top of base.
// Strings: non-empty overlay replaces base.
// Slices: non-nil overlay replaces base entirely ([] explicitly clears).
// TextFilter: non-zero overlay replaces base.
func merge(base, overlay Config) Config {
	result := base

	// Display
	if overlay.Display.Format != "" {
		result.Display.Format = overlay.Display.Format
	}
	if overlay.Display.Modifiers != nil {
		result.Display.Modifiers = overlay.Display.Modifiers
	}

	// Filter
	if !overlay.Filter.Pkg.IsZero() {
		result.Filter.Pkg = overlay.Filter.Pkg
	}
	if !overlay.Filter.Tag.IsZero() {
		result.Filter.Tag = overlay.Filter.Tag
	}
	if !overlay.Filter.Txt.IsZero() {
		result.Filter.Txt = overlay.Filter.Txt
	}

	return result
}
