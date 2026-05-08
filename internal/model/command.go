package model

type Command int

const (
	// Filters
	CommandPackage Command = iota
	CommandTag
	CommandLevel
	CommandContent

	// Style
	CommandOutput
	CommandTagWidth

	// Connection
	CommandReconnect
	CommandDevices

	// Other
	CommandExportBuffer
	CommandStartRecording
	CommandStopRecording
	CommandExit
)

type CommandType int

const (
	CommandTypeNavigation CommandType = iota
	CommandTypeAction
)

type CommandData struct {
	Command  Command
	Type     CommandType
	Name     string
	Value    string // Optional
	Shortcut string // Optional
}

type CommandGroup struct {
	Name     string
	Commands []CommandData
}

func Commands() []CommandGroup {
	return []CommandGroup{
		{
			Name: "Filters",
			Commands: []CommandData{
				{Command: CommandPackage, Type: CommandTypeNavigation, Name: "Package", Shortcut: "ctrl+x p"},
				{Command: CommandTag, Type: CommandTypeNavigation, Name: "Tag", Shortcut: "ctrl+x t"},
				{Command: CommandLevel, Type: CommandTypeNavigation, Name: "Log level", Shortcut: "ctrl+x l"},
				{Command: CommandContent, Type: CommandTypeNavigation, Name: "Content", Shortcut: "ctrl+x c"},
			},
		},
		{
			Name: "Style",
			Commands: []CommandData{
				{Command: CommandOutput, Type: CommandTypeNavigation, Name: "Output", Shortcut: "ctrl+x o"},
				{Command: CommandTagWidth, Type: CommandTypeNavigation, Name: "Tag width"},
			},
		},
		{
			Name: "Connection",
			Commands: []CommandData{
				{Command: CommandReconnect, Type: CommandTypeAction, Name: "Reconnect", Shortcut: "ctrl+x r"},
				{Command: CommandDevices, Type: CommandTypeNavigation, Name: "Devices", Shortcut: "ctrl+x d"},
			},
		},
		{
			Name: "Other",
			Commands: []CommandData{
				{Command: CommandExportBuffer, Type: CommandTypeAction, Name: "Export buffer"},
				{Command: CommandStartRecording, Type: CommandTypeAction, Name: "Start recording"},
				{Command: CommandStopRecording, Type: CommandTypeAction, Name: "Stop recording"},
				{Command: CommandExit, Type: CommandTypeAction, Name: "Exit", Shortcut: "ctrl+c"},
			},
		},
	}
}
