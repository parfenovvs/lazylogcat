package model

import (
	"slices"
	"strings"
)

var AllFormats = []string{
	"brief",
	"long",
	"process",
	"raw",
	"tag",
	"thread",
	"threadtime",
	"time",
}

var AllModifiers = []string{
	"color",
	"descriptive",
	"epoch",
	"monotonic",
	"printable",
	"uid",
	"usec",
	"UTC",
	"year",
	"zone",
}

var MaxModifiers = len(AllModifiers)

type Level string

const (
	LvlV Level = "V"
	LvlD Level = "D"
	LvlI Level = "I"
	LvlW Level = "W"
	LvlE Level = "E"
	LvlF Level = "F"
)

const lvls = "VDIWEF"

type Filter struct {
	PackageName string
	Level       Level
	Tag         string
	Text        string
}

type Format struct {
	SelectedFormat  string
	ActiveModifiers map[string]bool
}

func NewFormat(format string, modifiers ...string) Format {
	f := Format{
		SelectedFormat:  format,
		ActiveModifiers: make(map[string]bool),
	}
	for _, m := range modifiers {
		f.ActiveModifiers[m] = true
	}
	return f
}

func (f *Filter) IsEmpty() bool {
	return f.PackageName == "" &&
		(f.Level == "" || f.Level == LvlV) &&
		f.Tag == "" &&
		f.Text == ""
}

func (l Level) Next() Level {
	if l == LvlF {
		return LvlV
	}

	i := strings.Index(lvls, string(l))
	if i == -1 {
		return LvlV
	}
	return Level(lvls[i+1])
}

func (f *Format) Value() string {
	if f.SelectedFormat == "" {
		return "time"
	}
	return f.SelectedFormat
}

func (f *Format) Modifiers() []string {
	var mods []string
	for _, name := range AllModifiers {
		if f.ActiveModifiers[name] {
			mods = append(mods, name)
		}
	}
	return mods
}

func (f *Format) IsEmpty() bool {
	return f.SelectedFormat == "" && len(f.ActiveModifiers) == 0
}

func (f *Format) FormatIndex() int {
	i := slices.Index(AllFormats, f.SelectedFormat)
	if i < 0 {
		return 0
	}
	return i
}

func (f *Format) SetFormatByIndex(i int) {
	if i >= 0 && i < len(AllFormats) {
		f.SelectedFormat = AllFormats[i]
	}
}

func (f *Format) ToggleModifierByIndex(i int) {
	if i >= 0 && i < len(AllModifiers) {
		name := AllModifiers[i]
		if f.ActiveModifiers == nil {
			f.ActiveModifiers = make(map[string]bool)
		}
		f.ActiveModifiers[name] = !f.ActiveModifiers[name]
		if !f.ActiveModifiers[name] {
			delete(f.ActiveModifiers, name)
		}
	}
}

func (f *Format) IsModifierActive(name string) bool {
	return f.ActiveModifiers[name]
}

func (f *Format) IsFormatValue(name string) bool {
	return f.Value() == name
}
