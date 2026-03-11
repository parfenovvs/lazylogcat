package web

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/parfenovvs/lazylogcat/internal/config"
	"github.com/parfenovvs/lazylogcat/internal/util"
)

// handleDevices returns the list of connected ADB devices as JSON.
func handleDevices(w http.ResponseWriter, r *http.Request) {
	devices, err := util.GetConnectedDevices()
	if err != nil {
		slog.Error("Failed to get devices", "error", err)
		http.Error(w, `{"error":"failed to get devices"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(devices); err != nil {
		slog.Error("Failed to encode devices", "error", err)
	}
}

// handleConfig returns the resolved configuration as JSON.
func handleConfig(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(cfg); err != nil {
			slog.Error("Failed to encode config", "error", err)
		}
	}
}
