package plugin

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type PluginEntry struct {
	Name       string   `json:"name"`
	URL        string   `json:"url"`
	ConfigKey  string   `json:"config_key"`
	Hooks      []string `json:"hooks"`
	Entrypoint string   `json:"entrypoint"`
	Runtime    string   `json:"runtime"`
}

type Registry struct {
	Plugins []PluginEntry `json:"plugins"`
}

type PluginRegistryService struct {
	path string
}

func NewPluginRegistryService(path string) *PluginRegistryService {
	return &PluginRegistryService{path: path}
}

func DefaultRegistryPath() string {
	configDir, _ := os.UserConfigDir()
	return filepath.Join(configDir, "workforge", "plugins.json")
}

func (r *PluginRegistryService) Load() (*Registry, error) {
	data, err := os.ReadFile(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Registry{Plugins: []PluginEntry{}}, nil
		}
		return nil, err
	}

	var reg Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil, err
	}
	return &reg, nil
}

func (r *PluginRegistryService) Save(reg *Registry) error {
	if err := os.MkdirAll(filepath.Dir(r.path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.path, data, 0o644)
}

func (r *PluginRegistryService) Add(entry PluginEntry) error {
	reg, err := r.Load()
	if err != nil {
		return err
	}

	for i, p := range reg.Plugins {
		if p.Name == entry.Name {
			reg.Plugins[i] = entry
			return r.Save(reg)
		}
	}

	reg.Plugins = append(reg.Plugins, entry)
	return r.Save(reg)
}

func (r *PluginRegistryService) Remove(name string) error {
	reg, err := r.Load()
	if err != nil {
		return err
	}

	for i, p := range reg.Plugins {
		if p.Name == name {
			reg.Plugins = append(reg.Plugins[:i], reg.Plugins[i+1:]...)
			return r.Save(reg)
		}
	}
	return nil
}

func (r *PluginRegistryService) Find(name string) (*PluginEntry, bool) {
	reg, err := r.Load()
	if err != nil {
		return nil, false
	}

	for _, p := range reg.Plugins {
		if p.Name == name {
			return &p, true
		}
	}
	return nil, false
}

func (r *PluginRegistryService) List() ([]PluginEntry, error) {
	reg, err := r.Load()
	if err != nil {
		return nil, err
	}
	return reg.Plugins, nil
}
