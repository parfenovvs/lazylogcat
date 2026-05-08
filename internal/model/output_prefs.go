package model

// Columns controls which fields of a parsed LogLine are included in ModifiedString output.
type Columns struct {
	Date    bool `json:"date"`
	Time    bool `json:"time"`
	PID     bool `json:"pid"`
	TID     bool `json:"tid"`
	Level   bool `json:"level"`
	Tag     bool `json:"tag"`
	Message bool `json:"message"`
}

// OutputPrefs controls visual output preferences for log rendering.
type OutputPrefs struct {
	Color    bool    `json:"color"`
	SoftWrap bool    `json:"softWrap"`
	Columns  Columns `json:"columns"`
	TagWidth int     `json:"tagWidth"`
}
