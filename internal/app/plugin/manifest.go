package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Manifest struct {
	Name       string   `json:"name"`
	ConfigKey  string   `json:"config_key"`
	Hooks      []string `json:"hooks"`
	Entrypoint string   `json:"entrypoint"`
	Runtime    string   `json:"runtime"`
}

func LoadManifest(pluginDir string) (*Manifest, error) {
	path := filepath.Join(pluginDir, "plugin.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}

	if m.Entrypoint == "" {
		m.Entrypoint = "main.py"
	}
	if m.Runtime == "" {
		m.Runtime = "python3"
	}

	return &m, nil
}
