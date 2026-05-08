package util

import (
	"testing"

	"github.com/parfenovvs/lazylogcat/internal/config"
	"github.com/parfenovvs/lazylogcat/internal/model"
)

func TestFilterFromConfig(t *testing.T) {
	tests := []struct {
		name     string
		cfg      config.Config
		wantPkg  model.TextFilter
		wantTag  model.TextFilter
		wantText model.TextFilter
	}{
		{
			name:     "DefaultConfig",
			cfg:      config.DefaultConfig(),
			wantPkg:  model.TextFilter{Mode: model.FilterModeContains},
			wantTag:  model.TextFilter{Mode: model.FilterModeContains},
			wantText: model.TextFilter{Mode: model.FilterModeContains},
		},
		{
			name: "ContainsMode",
			cfg: config.Config{
				Filter: config.Filter{
					Pkg: config.TextFilter{Value: "com.example"},
					Tag: config.TextFilter{Value: "MyTag"},
					Txt: config.TextFilter{Value: "error"},
				},
			},
			wantPkg:  model.TextFilter{Value: "com.example", Mode: model.FilterModeContains},
			wantTag:  model.TextFilter{Value: "MyTag", Mode: model.FilterModeContains},
			wantText: model.TextFilter{Value: "error", Mode: model.FilterModeContains},
		},
		{
			name: "ExactMode",
			cfg: config.Config{
				Filter: config.Filter{
					Pkg: config.TextFilter{Value: "com.example.app", Mode: "exact"},
				},
			},
			wantPkg:  model.TextFilter{Value: "com.example.app", Mode: model.FilterModeExact},
			wantTag:  model.TextFilter{Mode: model.FilterModeContains},
			wantText: model.TextFilter{Mode: model.FilterModeContains},
		},
		{
			name: "RegexMode",
			cfg: config.Config{
				Filter: config.Filter{
					Tag: config.TextFilter{Value: "My.*Tag", Mode: "regex"},
				},
			},
			wantPkg:  model.TextFilter{Mode: model.FilterModeContains},
			wantTag:  model.TextFilter{Value: "My.*Tag", Mode: model.FilterModeRegex},
			wantText: model.TextFilter{Mode: model.FilterModeContains},
		},
		{
			name: "UnknownModeFallsBackToContains",
			cfg: config.Config{
				Filter: config.Filter{
					Pkg: config.TextFilter{Value: "com.example", Mode: "fuzzy"},
				},
			},
			wantPkg:  model.TextFilter{Value: "com.example", Mode: model.FilterModeContains},
			wantTag:  model.TextFilter{Mode: model.FilterModeContains},
			wantText: model.TextFilter{Mode: model.FilterModeContains},
		},
		{
			name: "MixedModes",
			cfg: config.Config{
				Filter: config.Filter{
					Pkg: config.TextFilter{Value: "com.example", Mode: "exact"},
					Tag: config.TextFilter{Value: "err.*", Mode: "regex"},
					Txt: config.TextFilter{Value: "hello"},
				},
			},
			wantPkg:  model.TextFilter{Value: "com.example", Mode: model.FilterModeExact},
			wantTag:  model.TextFilter{Value: "err.*", Mode: model.FilterModeRegex},
			wantText: model.TextFilter{Value: "hello", Mode: model.FilterModeContains},
		},
		{
			name: "ExplicitContainsMode",
			cfg: config.Config{
				Filter: config.Filter{
					Pkg: config.TextFilter{Value: "com.example", Mode: "contains"},
				},
			},
			wantPkg:  model.TextFilter{Value: "com.example", Mode: model.FilterModeContains},
			wantTag:  model.TextFilter{Mode: model.FilterModeContains},
			wantText: model.TextFilter{Mode: model.FilterModeContains},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterFromConfig(&tt.cfg)

			if got.PackageName.Value != tt.wantPkg.Value || got.PackageName.Mode != tt.wantPkg.Mode {
				t.Errorf("PackageName = {Value:%q, Mode:%v}, want {Value:%q, Mode:%v}",
					got.PackageName.Value, got.PackageName.Mode, tt.wantPkg.Value, tt.wantPkg.Mode)
			}
			if got.Tag.Value != tt.wantTag.Value || got.Tag.Mode != tt.wantTag.Mode {
				t.Errorf("Tag = {Value:%q, Mode:%v}, want {Value:%q, Mode:%v}",
					got.Tag.Value, got.Tag.Mode, tt.wantTag.Value, tt.wantTag.Mode)
			}
			if got.Text.Value != tt.wantText.Value || got.Text.Mode != tt.wantText.Mode {
				t.Errorf("Text = {Value:%q, Mode:%v}, want {Value:%q, Mode:%v}",
					got.Text.Value, got.Text.Mode, tt.wantText.Value, tt.wantText.Mode)
			}
			if got.Level != model.LvlV {
				t.Errorf("Level = %q, want %q", got.Level, model.LvlV)
			}
		})
	}
}

func TestFilterFromConfig_RegexCompiled(t *testing.T) {
	cfg := config.Config{
		Filter: config.Filter{
			Tag: config.TextFilter{Value: "My.*Tag", Mode: "regex"},
		},
	}

	got := FilterFromConfig(&cfg)

	// Verify regex was pre-compiled by checking that Match works
	if !got.Tag.Match("MyFooTag") {
		t.Error("Tag.Match(\"MyFooTag\") = false, want true (regex should be pre-compiled)")
	}
	if got.Tag.Match("NoMatch") {
		t.Error("Tag.Match(\"NoMatch\") = true, want false")
	}
}

func boolPtr(v bool) *bool { return &v }

func TestBoolOrDefault(t *testing.T) {
	tests := []struct {
		name string
		ptr  *bool
		def  bool
		want bool
	}{
		{name: "NilDefaultTrue", ptr: nil, def: true, want: true},
		{name: "NilDefaultFalse", ptr: nil, def: false, want: false},
		{name: "TruePtr", ptr: boolPtr(true), def: false, want: true},
		{name: "FalsePtr", ptr: boolPtr(false), def: true, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := boolOrDefault(tt.ptr, tt.def)
			if got != tt.want {
				t.Errorf("boolOrDefault(%v, %v) = %v, want %v", tt.ptr, tt.def, got, tt.want)
			}
		})
	}
}

func TestColorFromConfig(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.Config
		want bool
	}{
		{name: "NilDefaultsTrue", cfg: config.Config{}, want: true},
		{name: "True", cfg: config.Config{Display: config.Display{Color: boolPtr(true)}}, want: true},
		{name: "False", cfg: config.Config{Display: config.Display{Color: boolPtr(false)}}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ColorFromConfig(&tt.cfg)
			if got != tt.want {
				t.Errorf("ColorFromConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWrapFromConfig(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.Config
		want bool
	}{
		{name: "NilDefaultsTrue", cfg: config.Config{}, want: true},
		{name: "True", cfg: config.Config{Display: config.Display{Wrap: boolPtr(true)}}, want: true},
		{name: "False", cfg: config.Config{Display: config.Display{Wrap: boolPtr(false)}}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WrapFromConfig(&tt.cfg)
			if got != tt.want {
				t.Errorf("WrapFromConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func intPtr(v int) *int { return &v }

func TestTagWidthFromConfig(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.Config
		want int
	}{
		{name: "NilDefaultsZero", cfg: config.Config{}, want: 0},
		{name: "ZeroExplicit", cfg: config.Config{Display: config.Display{TagWidth: intPtr(0)}}, want: 0},
		{name: "InRange", cfg: config.Config{Display: config.Display{TagWidth: intPtr(42)}}, want: 42},
		{name: "Max", cfg: config.Config{Display: config.Display{TagWidth: intPtr(MaxTagWidth)}}, want: MaxTagWidth},
		{name: "NegativeClampedToZero", cfg: config.Config{Display: config.Display{TagWidth: intPtr(-1)}}, want: 0},
		{name: "AboveMaxClamped", cfg: config.Config{Display: config.Display{TagWidth: intPtr(1000)}}, want: MaxTagWidth},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TagWidthFromConfig(&tt.cfg)
			if got != tt.want {
				t.Errorf("TagWidthFromConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColumnsFromConfig(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.Config
		want model.Columns
	}{
		{
			name: "NilColumnsHardcodedDefaults",
			cfg:  config.Config{},
			want: model.Columns{Date: false, Time: true, PID: false, TID: false, Level: true, Tag: true, Message: true},
		},
		{
			name: "AllFieldsSet",
			cfg: config.Config{Display: config.Display{Columns: &config.Columns{
				Date:    boolPtr(true),
				Time:    boolPtr(false),
				PID:     boolPtr(true),
				TID:     boolPtr(true),
				Level:   boolPtr(false),
				Tag:     boolPtr(false),
				Message: boolPtr(false),
			}}},
			want: model.Columns{Date: true, Time: false, PID: true, TID: true, Level: false, Tag: false, Message: false},
		},
		{
			name: "PartialMix_NilFallsToDefault",
			cfg: config.Config{Display: config.Display{Columns: &config.Columns{
				Date: boolPtr(true),
				// Time nil → default true
				PID: boolPtr(true),
				// TID nil → default false
				// Level nil → default true
				Tag: boolPtr(false),
				// Message nil → default true
			}}},
			want: model.Columns{Date: true, Time: true, PID: true, TID: false, Level: true, Tag: false, Message: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ColumnsFromConfig(&tt.cfg)
			if got != tt.want {
				t.Errorf("ColumnsFromConfig() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestModeFromConfig(t *testing.T) {
	tests := []struct {
		input string
		want  model.TextFilterMode
	}{
		{input: "", want: model.FilterModeContains},
		{input: "contains", want: model.FilterModeContains},
		{input: "exact", want: model.FilterModeExact},
		{input: "regex", want: model.FilterModeRegex},
		{input: "unknown", want: model.FilterModeContains},
		{input: "EXACT", want: model.FilterModeContains}, // case-sensitive, unknown
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := modeFromConfig(tt.input)
			if got != tt.want {
				t.Errorf("modeFromConfig(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
