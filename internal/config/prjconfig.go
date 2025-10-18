package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Projects map[string]Project
type Project struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	GitWorkTree bool   `json:"git_work_tree"`
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

func ListProjects() (Projects, error) {
	workforgePath := os.Getenv("HOME") + "/" + WORK_FORGE_PRJ_CONFIG_DIR
	if _, err := os.Stat(workforgePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("workforge config directory does not exist")
	}
	projects, err := LoadProjects(workforgePath + "/" + WORK_FORGE_PRJ_CONFIG_FILE)
	if err != nil {
		return nil, fmt.Errorf("failed to load existing projects: %w", err)
	}
	return projects, nil
}

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

		if isGWTLeaf(p.Path) {
			out[p.Name] = p
			hitmap[p.Name] = true
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

func isGWTLeaf(path string) bool {
	st, err := os.Stat(filepath.Join(path, ".git"))
	if err != nil {
		return false
	}
	return !st.IsDir()
}
