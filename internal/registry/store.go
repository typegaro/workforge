package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"workforge/internal/infra/fs"
)

const WorkForgeConfigFile = "workforge.json"

type Projects map[string]Project

type Project struct {
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	GitWorkTree bool     `json:"git_work_tree"`
	Tags        []string `json:"tags,omitempty"`
}

type ProjectEntry struct {
	Project
	IsGWT bool
}

func RegistryPath() (string, error) {
	resolver := fs.NewPathResolver()
	return resolver.RegistryPath()
}

func EnsureRegistry() (string, error) {
	path, err := RegistryPath()
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", fmt.Errorf("failed to create workforge config directory: %w", err)
		}
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
			return "", fmt.Errorf("failed to create workforge config file: %w", err)
		}
	}
	return path, nil
}

func SaveProjects(filename string, projects Projects) error {
	data, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}
	return os.WriteFile(filename, data, 0644)
}

func LoadProjects(filename string) (Projects, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}
	var projects Projects
	if err := json.Unmarshal(data, &projects); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}
	return projects, nil
}

func IsGWTLeaf(path string) bool {
	st, err := os.Stat(filepath.Join(path, ".git"))
	if err != nil {
		return false
	}
	return !st.IsDir()
}
