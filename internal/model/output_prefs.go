package model

// Columns controls which fields of a parsed LogLine are included in ModifiedString output.
type Columns struct {
	Date    bool
	Time    bool
	PID     bool
	TID     bool
	Level   bool
	Tag     bool
	Message bool
}

// OutputPrefs controls visual output preferences for log rendering.
type OutputPrefs struct {
	Color    bool
	SoftWrap bool
	Columns  Columns
}
