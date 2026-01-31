package plugin

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type PluginInstallerService struct {
	pluginsDir string
	registry   *PluginRegistryService
}

func NewPluginInstallerService(pluginsDir string, registry *PluginRegistryService) *PluginInstallerService {
	return &PluginInstallerService{
		pluginsDir: pluginsDir,
		registry:   registry,
	}
}

func (s *PluginInstallerService) Install(url string) (*PluginEntry, error) {
	if err := os.MkdirAll(s.pluginsDir, 0o755); err != nil {
		return nil, fmt.Errorf("create plugins dir: %w", err)
	}

	name := extractRepoName(url)
	pluginPath := filepath.Join(s.pluginsDir, name)

	if _, err := os.Stat(pluginPath); err == nil {
		return nil, fmt.Errorf("plugin %q already exists at %s", name, pluginPath)
	}

	cmd := exec.Command("git", "clone", url, pluginPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git clone failed: %w", err)
	}

	manifest, err := LoadManifest(pluginPath)
	if err != nil {
		os.RemoveAll(pluginPath)
		return nil, fmt.Errorf("invalid plugin (no plugin.json): %w", err)
	}

	entry := PluginEntry{
		Name:       manifest.Name,
		URL:        url,
		ConfigKey:  manifest.ConfigKey,
		Hooks:      manifest.Hooks,
		Entrypoint: manifest.Entrypoint,
		Runtime:    manifest.Runtime,
	}

	if err := s.registry.Add(entry); err != nil {
		return nil, fmt.Errorf("register plugin: %w", err)
	}

	return &entry, nil
}

func (s *PluginInstallerService) Uninstall(name string) error {
	pluginPath := filepath.Join(s.pluginsDir, name)

	if err := os.RemoveAll(pluginPath); err != nil {
		return fmt.Errorf("remove plugin dir: %w", err)
	}

	return s.registry.Remove(name)
}

func extractRepoName(url string) string {
	url = strings.TrimSuffix(url, ".git")
	parts := strings.Split(url, "/")
	return parts[len(parts)-1]
}
