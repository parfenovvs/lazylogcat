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
	Color    *bool    `json:"color,omitempty"`
	Wrap     *bool    `json:"wrap,omitempty"`
	TagWidth *int     `json:"tag_width,omitempty"`
	Columns  *Columns `json:"columns,omitempty"`
}

// Columns controls which fields of a parsed log line are visible.
type Columns struct {
	Date    *bool `json:"date,omitempty"`
	Time    *bool `json:"time,omitempty"`
	PID     *bool `json:"pid,omitempty"`
	TID     *bool `json:"tid,omitempty"`
	Level   *bool `json:"level,omitempty"`
	Tag     *bool `json:"tag,omitempty"`
	Message *bool `json:"message,omitempty"`
}

// Filter holds log filtering parameters.
type Filter struct {
	Pkg TextFilter `json:"package_name,omitempty"`
	Tag TextFilter `json:"log_tag,omitempty"`
	Txt TextFilter `json:"log_text,omitempty"`
}

// TextFilter represents a filter field with an optional matching mode.
// It accepts both a plain string and an object form in JSON:
//
//	"log_tag": "MyTag"
//	"log_tag": { "value": "MyTag" }
//	"log_tag": { "value": "MyTag", "mode": "exact" }
//
// Valid mode values: "contains" (default), "exact", "regex".
// When mode is omitted or empty, "contains" is assumed.
type TextFilter struct {
	Value string `json:"value"`
	Mode  string `json:"mode,omitempty"`
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

// MarshalJSON outputs the short string form when Mode is default (empty or
// "contains"), and the full object form when a non-default mode is set.
func (f TextFilter) MarshalJSON() ([]byte, error) {
	if f.Mode == "" || f.Mode == "contains" {
		return json.Marshal(f.Value)
	}
	type alias TextFilter
	return json.Marshal(alias(f))
}

// DefaultConfig returns the default configuration with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Display: Display{
			Color: ptrBool(true),
			Wrap:  ptrBool(true),
			Columns: &Columns{
				Date:    ptrBool(false),
				Time:    ptrBool(true),
				PID:     ptrBool(false),
				TID:     ptrBool(false),
				Level:   ptrBool(true),
				Tag:     ptrBool(true),
				Message: ptrBool(true),
			},
		},
	}
}

// ptrBool returns a pointer to a new bool value.
func ptrBool(v bool) *bool {
	return &v
}

// String returns a JSON string representation of the config for logging.
func (c *Config) String() string {
	data, err := json.Marshal(c)
	if err != nil {
		return "{}"
	}
	return string(data)
}

// ResolveOptions overrides default paths for the project and local config layers.
// Zero value means use the standard paths (.lazylogcat/config.json and
// .lazylogcat/config.local.json). The global user config path is not overridden.
type ResolveOptions struct {
	ProjectConfigPath string
	LocalConfigPath   string
}

// Resolve discovers and merges configuration from all layers in order:
//  1. Default configuration
//  2. Global user config: ~/.config/lazylogcat/config.json
//  3. Project config: .lazylogcat/config.json (or ProjectConfigPath when set)
//  4. Local override: .lazylogcat/config.local.json (or LocalConfigPath when set)
//
// opts may be nil. Returns the merged config and any non-fatal errors encountered
// during loading. Always returns a usable config, even if some files fail to load.
func Resolve(opts *ResolveOptions) (Config, error) {
	cfg := DefaultConfig()
	var errs []error

	for _, path := range configPaths(opts) {
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
func configPaths(opts *ResolveOptions) []string {
	var paths []string

	// Layer 1: Global user config (~/.config/lazylogcat/config.json)
	if dir, err := os.UserConfigDir(); err == nil {
		paths = append(paths, filepath.Join(dir, "lazylogcat", "config.json"))
	}

	// Layer 2: Project config (.lazylogcat/config.json)
	proj := filepath.Join(".lazylogcat", "config.json")
	if opts != nil && opts.ProjectConfigPath != "" {
		proj = opts.ProjectConfigPath
	}
	paths = append(paths, proj)

	// Layer 3: Local override (.lazylogcat/config.local.json)
	local := filepath.Join(".lazylogcat", "config.local.json")
	if opts != nil && opts.LocalConfigPath != "" {
		local = opts.LocalConfigPath
	}
	paths = append(paths, local)

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
// *bool, *int: non-nil overlay replaces base.
func merge(base, overlay Config) Config {
	result := base

	// Display
	if overlay.Display.Color != nil {
		result.Display.Color = overlay.Display.Color
	}
	if overlay.Display.Wrap != nil {
		result.Display.Wrap = overlay.Display.Wrap
	}
	if overlay.Display.TagWidth != nil {
		result.Display.TagWidth = overlay.Display.TagWidth
	}
	if overlay.Display.Columns != nil {
		if result.Display.Columns != nil {
			copied := *result.Display.Columns
			result.Display.Columns = &copied
		}
		mergeColumns(result.Display.Columns, overlay.Display.Columns)
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

// mergeColumns applies non-nil fields from overlay on top of base.
func mergeColumns(base, overlay *Columns) {
	if base == nil || overlay == nil {
		return
	}
	if overlay.Date != nil {
		base.Date = overlay.Date
	}
	if overlay.Time != nil {
		base.Time = overlay.Time
	}
	if overlay.PID != nil {
		base.PID = overlay.PID
	}
	if overlay.TID != nil {
		base.TID = overlay.TID
	}
	if overlay.Level != nil {
		base.Level = overlay.Level
	}
	if overlay.Tag != nil {
		base.Tag = overlay.Tag
	}
	if overlay.Message != nil {
		base.Message = overlay.Message
	}
}
