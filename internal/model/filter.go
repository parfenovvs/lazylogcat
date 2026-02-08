package model

import "strings"

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
