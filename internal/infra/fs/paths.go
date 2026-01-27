package fs

import (
	"fmt"
	"os"
	"path/filepath"
)

type PathResolver struct {
	userConfigDir func() (string, error)
	userHomeDir   func() (string, error)
}

func NewPathResolver() *PathResolver {
	return &PathResolver{
		userConfigDir: os.UserConfigDir,
		userHomeDir:   os.UserHomeDir,
	}
}

func (r *PathResolver) WorkforgeConfigDir() (string, error) {
	if r == nil {
		return "", fmt.Errorf("path resolver is nil")
	}
	if cfgDir, err := r.userConfigDir(); err == nil && cfgDir != "" {
		return filepath.Join(cfgDir, "workforge"), nil
	}
	homeDir, err := r.userHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home directory: %w", err)
	}
	return filepath.Join(homeDir, ".config", "workforge"), nil
}

func (r *PathResolver) RegistryPath() (string, error) {
	configDir, err := r.WorkforgeConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "workforge.json"), nil
}

func (r *PathResolver) NormalizePath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path is empty")
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve absolute path: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(absPath)
	if err == nil {
		return resolved, nil
	}
	if os.IsNotExist(err) {
		return absPath, nil
	}
	return "", fmt.Errorf("resolve symlinks: %w", err)
}
