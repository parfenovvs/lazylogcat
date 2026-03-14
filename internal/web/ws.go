package web

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	"github.com/parfenovvs/lazylogcat/internal/config"
	"github.com/parfenovvs/lazylogcat/internal/model"
	"github.com/parfenovvs/lazylogcat/internal/util"
)

const (
	drainInterval = 50 * time.Millisecond
)

// handleWebSocket upgrades the HTTP connection to a WebSocket and manages
// the per-connection session lifecycle.
func handleWebSocket(cfg config.Config, demo bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			// Allow any origin during development; tighten for production.
			InsecureSkipVerify: true,
		})
		if err != nil {
			slog.Error("WebSocket accept failed", "error", err)
			return
		}
		defer conn.CloseNow()

		ctx := r.Context()
		var session *Session
		if demo {
			session = NewDemoSession(cfg)
		} else {
			session = NewSession(cfg)
		}
		defer session.Close()

		slog.Debug("WebSocket connection established", "demo", demo)

		// Start read loop in a goroutine — it dispatches client commands.
		readCtx, readCancel := context.WithCancel(ctx)
		defer readCancel()

		go wsReadLoop(readCtx, conn, session, demo)

		// Write loop: drain session every drainInterval and send batches.
		wsWriteLoop(ctx, conn, session)
	}
}

// wsReadLoop reads client messages and dispatches them to the session.
func wsReadLoop(ctx context.Context, conn *websocket.Conn, session *Session, demo bool) {
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			// Connection closed or context cancelled
			slog.Debug("WebSocket read error", "error", err)
			return
		}

		msgType, payload, err := parseClientMsg(data)
		if err != nil {
			slog.Warn("Malformed WebSocket message", "error", err, "data", string(data))
			continue
		}

		slog.Debug("WebSocket received", "type", msgType)

		switch msgType {
		case "connect":
			cmd := payload.(connectCmd)
			if err := session.Connect(cmd.DeviceID, cmd.Filter); err != nil {
				slog.Error("Session connect failed", "error", err)
				writeJSON(ctx, conn, newErrorMsg("failed to connect: "+err.Error()))
				continue
			}
			writeJSON(ctx, conn, newConnectedMsg(cmd.DeviceID))

			// Start a goroutine to detect when the reader stops (EOF/error)
			go func(deviceID string) {
				session.WaitForDone()
				errStr := ""
				if session.Err() != nil {
					errStr = session.Err().Error()
				}
				writeJSON(ctx, conn, newDisconnectedMsg(errStr))
			}(cmd.DeviceID)

		case "disconnect":
			session.Disconnect()

		case "updateFilter":
			cmd := payload.(updateFilterCmd)
			session.UpdateFilter(cmd.Filter)
			session.ClearBuffer()
			writeJSON(ctx, conn, newClearLinesMsg())

		case "listDevices":
			if demo {
				writeJSON(ctx, conn, newDevicesMsg([]model.Device{demoDevice}))
				continue
			}
			devices, err := util.GetConnectedDevices()
			if err != nil {
				slog.Error("Failed to get devices", "error", err)
				writeJSON(ctx, conn, newErrorMsg("failed to get devices: "+err.Error()))
				continue
			}
			writeJSON(ctx, conn, newDevicesMsg(devices))

		default:
			slog.Warn("Unknown WebSocket message type", "type", msgType)
		}
	}
}

// wsWriteLoop drains the session at a fixed interval and sends line batches
// over the WebSocket.
func wsWriteLoop(ctx context.Context, conn *websocket.Conn, session *Session) {
	ticker := time.NewTicker(drainInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			lines := session.Drain()
			if len(lines) == 0 {
				continue
			}
			msg := newLinesMsg(lines)
			if err := writeJSON(ctx, conn, msg); err != nil {
				slog.Debug("WebSocket write error", "error", err)
				return
			}
		}
	}
}

// writeJSON encodes and sends a JSON message over the WebSocket.
func writeJSON(ctx context.Context, conn *websocket.Conn, v interface{}) error {
	// Use a short write deadline to avoid blocking forever
	writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return wsjson.Write(writeCtx, conn, v)
}
