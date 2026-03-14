package web

import (
	"encoding/json"

	"github.com/parfenovvs/lazylogcat/internal/model"
)

// --- Server -> Client messages ---

// serverMsg is the envelope for all server-to-client WebSocket messages.
type serverMsg struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

type connectedPayload struct {
	DeviceID string `json:"deviceId"`
}

type disconnectedPayload struct {
	Error string `json:"error,omitempty"`
}

type errorPayload struct {
	Message string `json:"message"`
}

func newLinesMsg(lines []model.LogLine) serverMsg {
	return serverMsg{Type: "lines", Data: lines}
}

func newConnectedMsg(deviceID string) serverMsg {
	return serverMsg{Type: "connected", Data: connectedPayload{DeviceID: deviceID}}
}

func newDisconnectedMsg(errStr string) serverMsg {
	return serverMsg{Type: "disconnected", Data: disconnectedPayload{Error: errStr}}
}

func newDevicesMsg(devices []model.Device) serverMsg {
	return serverMsg{Type: "devices", Data: devices}
}

func newErrorMsg(message string) serverMsg {
	return serverMsg{Type: "error", Data: errorPayload{Message: message}}
}

func newClearLinesMsg() serverMsg {
	return serverMsg{Type: "clearLines"}
}

// --- Client -> Server messages ---

type connectCmd struct {
	DeviceID string       `json:"deviceId"`
	Filter   model.Filter `json:"filter"`
}

type updateFilterCmd struct {
	Filter model.Filter `json:"filter"`
}

// parseClientMsg decodes a raw JSON message into a typed command.
// Returns the type string and the decoded payload (one of connectCmd,
// updateFilterCmd, or nil for parameterless commands).
func parseClientMsg(data []byte) (string, interface{}, error) {
	var env struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return "", nil, err
	}

	switch env.Type {
	case "connect":
		var cmd connectCmd
		if err := json.Unmarshal(data, &cmd); err != nil {
			return env.Type, nil, err
		}
		return env.Type, cmd, nil
	case "updateFilter":
		var cmd updateFilterCmd
		if err := json.Unmarshal(data, &cmd); err != nil {
			return env.Type, nil, err
		}
		return env.Type, cmd, nil
	case "disconnect", "listDevices":
		return env.Type, nil, nil
	default:
		return env.Type, nil, nil
	}
}
