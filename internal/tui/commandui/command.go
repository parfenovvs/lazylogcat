package commandui

type Command int

const (
	// Filters
	CommandPackage Command = iota
	CommandTag
	CommandLevel
	CommandContent

	// Output format
	CommandFormat
	CommandModifiers
	CommandToggleWrap

	// Connection
	CommandReconnect
	CommandDevices

	// Other
	CommandExportConfig
	CommandExit
)

type CommandType int

const (
	CommandTypeNavigation CommandType = iota
	CommandTypeToggle
	CommandTypeInput
	CommandTypeAction
	CommandTypeSelection
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
				{Command: CommandPackage, Type: CommandTypeInput, Name: "Package", Shortcut: "ctrl+x p"},
				{Command: CommandTag, Type: CommandTypeInput, Name: "Tag", Shortcut: "ctrl+x t"},
				{Command: CommandLevel, Type: CommandTypeNavigation, Name: "Log level", Shortcut: "ctrl+x l"},
				{Command: CommandContent, Type: CommandTypeInput, Name: "Content", Shortcut: "ctrl+x c"},
			},
		},
		{
			Name: "Output",
			Commands: []CommandData{
				{Command: CommandFormat, Type: CommandTypeNavigation, Name: "Format", Shortcut: "ctrl+x f"},
				{Command: CommandModifiers, Type: CommandTypeSelection, Name: "Modifiers", Shortcut: "ctrl+x m"},
				{Command: CommandToggleWrap, Type: CommandTypeToggle, Name: "Toggle wrap", Shortcut: "ctrl+x w"},
			},
		},
		{
			Name: "Connection",
			Commands: []CommandData{
				{Command: CommandReconnect, Type: CommandTypeAction, Name: "Reconnect", Shortcut: "ctrl+x r"},
				{Command: CommandDevices, Type: CommandTypeAction, Name: "Devices", Shortcut: "ctrl+x d"},
			},
		},
		{
			Name: "Other",
			Commands: []CommandData{
				{Command: CommandExportConfig, Type: CommandTypeAction, Name: "Export config", Shortcut: "ctrl+x e"},
				{Command: CommandExit, Type: CommandTypeAction, Name: "Exit", Shortcut: "ctrl+c"},
			},
		},
	}
}
