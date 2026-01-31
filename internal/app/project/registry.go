package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"workforge/internal/infra/fs"
)

type ProjectRegistryService struct {
	paths *fs.PathResolver
}

func NewProjectRegistryService() *ProjectRegistryService {
	return &ProjectRegistryService{paths: fs.NewPathResolver()}
}

func (s *ProjectRegistryService) registryPath() (string, error) {
	return s.paths.RegistryPath()
}

func (s *ProjectRegistryService) ensureRegistry() (string, error) {
	path, err := s.registryPath()
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

func (s *ProjectRegistryService) Load() (Projects, error) {
	regPath, err := s.ensureRegistry()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(regPath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}
	var projects Projects
	if err := json.Unmarshal(data, &projects); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}
	return projects, nil
}

func (s *ProjectRegistryService) Save(projects Projects) error {
	regPath, err := s.ensureRegistry()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}
	return os.WriteFile(regPath, data, 0644)
}
