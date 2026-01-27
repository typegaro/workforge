package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

const WorkForgeConfigDir = ".config/workforge"
const WorkForgeConfigFile = "workforge.json"

type Projects map[string]Project

type Project struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	GitWorkTree bool   `json:"git_work_tree"`
}

type ProjectEntry struct {
	Project
	IsGWT bool
}

func RegistryPath() string {
	return filepath.Join(os.Getenv("HOME"), WorkForgeConfigDir, WorkForgeConfigFile)
}

func EnsureRegistry() (string, error) {
	path := RegistryPath()
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

func ListProjects() (Projects, error) {
	workforgePath := filepath.Join(os.Getenv("HOME"), WorkForgeConfigDir)
	if _, err := os.Stat(workforgePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("workforge config directory does not exist")
	}
	projects, err := LoadProjects(filepath.Join(workforgePath, WorkForgeConfigFile))
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

		if IsGWTLeaf(p.Path) {
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

func SortedProjectEntries() ([]ProjectEntry, error) {
	projs, hitmap, err := ListProjectsExpanded()
	if err != nil {
		return nil, err
	}
	entries := make([]ProjectEntry, 0, len(projs))
	for name, p := range projs {
		if p.Name == "" {
			p.Name = name
		}
		entries = append(entries, ProjectEntry{Project: p, IsGWT: hitmap[name]})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
	return entries, nil
}

func FindProjectEntry(name string) (ProjectEntry, error) {
	if name == "" {
		return ProjectEntry{}, fmt.Errorf("project name cannot be empty")
	}
	projs, hitmap, err := ListProjectsExpanded()
	if err != nil {
		return ProjectEntry{}, err
	}
	project, ok := projs[name]
	if !ok {
		return ProjectEntry{}, fmt.Errorf("project %q not found", name)
	}
	if project.Name == "" {
		project.Name = name
	}
	return ProjectEntry{Project: project, IsGWT: hitmap[name]}, nil
}

func IsGWTLeaf(path string) bool {
	st, err := os.Stat(filepath.Join(path, ".git"))
	if err != nil {
		return false
	}
	return !st.IsDir()
}
