package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type Projects map[string]Project

type Project struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	GitWorkTree bool   `json:"git_work_tree"`
}

const (
	configDirName  = "workforge"
	configFileName = "workforge.json"
)

func registryDir() (string, error) {
	if dir, ok := os.LookupEnv("XDG_CONFIG_HOME"); ok && dir != "" {
		return filepath.Join(dir, configDirName), nil
	}
	if dir, err := os.UserConfigDir(); err == nil && dir != "" {
		return filepath.Join(dir, configDirName), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot resolve home directory: %w", err)
	}
	return filepath.Join(home, ".config", configDirName), nil
}

func registryFile() (string, error) {
	dir, err := registryDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

func ensureRegistryFile() (string, error) {
	filePath, err := registryFile()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return "", fmt.Errorf("failed to create workforge config directory: %w", err)
	}
	if _, err := os.Stat(filePath); errors.Is(err, fs.ErrNotExist) {
		if err := os.WriteFile(filePath, []byte("{}"), 0o644); err != nil {
			return "", fmt.Errorf("failed to create workforge config file: %w", err)
		}
	}
	return filePath, nil
}

func SaveProjects(projects Projects) error {
	path, err := ensureRegistryFile()
	if err != nil {
		return err
	}
	if projects == nil {
		projects = make(Projects)
	}
	data, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

func LoadProjects() (Projects, error) {
	path, err := ensureRegistryFile()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}
	if len(data) == 0 {
		return make(Projects), nil
	}
	var projects Projects
	if err := json.Unmarshal(data, &projects); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}
	if projects == nil {
		projects = make(Projects)
	}
	return projects, nil
}

func ListProjects() (Projects, error) {
	projects, err := LoadProjects()
	if err != nil {
		return nil, fmt.Errorf("failed to load existing projects: %w", err)
	}
	return projects, nil
}

// ListProjectsExpanded espande i progetti con GitWorkTree=true in tutte le subdir (solo directory, depth 1).
// Restituisce:
// - Projects "flattened": progetti normali + subdir dei GWT (senza la base GWT).
// - hitmap: true per gli elementi che sono subdir GWT.
func ListProjectsExpanded() (Projects, map[string]bool, error) {
	base, err := ListProjects()
	if err != nil {
		return nil, nil, err
	}
	out := make(Projects)
	hitmap := make(map[string]bool)
	for _, p := range base {
		if !p.GitWorkTree {
			out[p.Name] = p
			hitmap[p.Name] = false
			continue
		}
		entries, err := os.ReadDir(p.Path)
		if err != nil {
			return nil, nil, fmt.Errorf("error reading GWT path %q: %w", p.Path, err)
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			subName := p.Name + "/" + e.Name()
			subPath := filepath.Join(p.Path, e.Name())
			out[subName] = Project{
				Name:        subName,
				Path:        subPath,
				GitWorkTree: false,
			}
			hitmap[subName] = true
		}
	}
	return out, hitmap, nil
}
