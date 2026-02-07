package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
)

// Config represents the application's configuration structure.
// It includes user preferences and session details.
// Prefs holds user preferences for log display.
// Session holds the last used session parameters.
type Config struct {
	Prefs   Prefs   `json:"preferences"`
	Session Session `json:"session"`
}

type Prefs struct {
	Format    string   `json:"log_format,omitempty"`
	Modifiers []string `json:"log_modifiers,omitempty"`
}

type Session struct {
	DeviceId string `json:"device_id,omitempty"`
	Pkg      string `json:"package_name,omitempty"`
	Tag      string `json:"log_tag,omitempty"`
	Txt      string `json:"log_text,omitempty"`
}

func DefaultConfig() Config {
	return Config{
		Prefs: Prefs{
			Format:    "time",
			Modifiers: []string{"color"},
		},
		Session: Session{
			DeviceId: "",
			Pkg:      "",
			Tag:      "",
			Txt:      "",
		},
	}
}

func (c *Config) String() string {
	data, err := json.Marshal(c)
	if err != nil {
		return "{}"
	}
	return string(data)
}

// Save writes the configuration to a JSON file at the given path.
func Save(cfg Config, path string) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return nil
}

// Load reads the configuration from the provided file.
// If the file is invalid or cannot be read, it returns a default configuration and an error.
func Load(file *os.File) (Config, error) {
	config := DefaultConfig()

	decoder := json.NewDecoder(file)
	err := decoder.Decode(&config)

	if err != nil {
		slog.Error("Failed to decode config file", "error", err)
		return DefaultConfig(), fmt.Errorf("failed to decode config file: %w", err)
	}

	return config, nil
}
