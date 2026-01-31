package project

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"workforge/internal/infra/log"
)

type ProjectService struct {
	registry *ProjectRegistryService
}

func NewService() *ProjectService {
	return &ProjectService{registry: NewProjectRegistryService()}
}

func (s *ProjectService) listProjectsExpanded() (Projects, map[string]bool, error) {
	base, err := s.registry.Load()
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

func (s *ProjectService) SortedProjectEntries() ([]ProjectEntry, error) {
	projs, hitmap, err := s.listProjectsExpanded()
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

func (s *ProjectService) FindProjectEntry(name string) (ProjectEntry, error) {
	if name == "" {
		return ProjectEntry{}, fmt.Errorf("project name cannot be empty")
	}
	projs, hitmap, err := s.listProjectsExpanded()
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

func (s *ProjectService) GetProjectPath(name string) (string, bool, error) {
	entry, err := s.FindProjectEntry(name)
	if err != nil {
		return "", false, err
	}
	return entry.Path, entry.IsGWT, nil
}

func (s *ProjectService) AddProject(name string, gwt bool, path *string) error {
	var absPath string
	var err error
	if path != nil {
		absPath, err = s.registry.paths.NormalizePath(*path)
		if err != nil {
			return fmt.Errorf("failed to resolve project path: %w", err)
		}
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		absPath, err = s.registry.paths.NormalizePath(cwd)
		if err != nil {
			return fmt.Errorf("failed to resolve project path: %w", err)
		}
	}
	log.Info("Adding project: %s (path: %s, gwt: %t)", name, absPath, gwt)
	projects, err := s.registry.Load()
	if err != nil {
		projects = make(Projects)
	}
	log.Debug("Loaded existing projects: %+v", projects)
	projects[name] = Project{Name: name, Path: absPath, GitWorkTree: gwt}
	if err := s.registry.Save(projects); err != nil {
		return err
	}
	return nil
}

func (s *ProjectService) AddLeaf(absLeafPath string) error {
	projects, err := s.registry.Load()
	if err != nil {
		projects = make(Projects)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	var baseName string
	for name, p := range projects {
		if p.GitWorkTree && !isGWTLeaf(p.Path) && p.Path == cwd {
			baseName = name
			break
		}
	}

	if baseName == "" {
		parent := filepath.Dir(absLeafPath)
		for name, p := range projects {
			if p.GitWorkTree && !isGWTLeaf(p.Path) && p.Path == parent {
				baseName = name
				break
			}
		}
	}

	leafName := filepath.Base(absLeafPath)
	key := leafName
	if baseName != "" {
		key = baseName + "/" + leafName
	}
	projects[key] = Project{Name: key, Path: absLeafPath, GitWorkTree: true}
	if err := s.registry.Save(projects); err != nil {
		return err
	}
	return nil
}

func (s *ProjectService) AddTag(projectName string, tag string) error {
	projectName = strings.TrimSpace(projectName)
	tag = strings.TrimSpace(tag)
	if projectName == "" {
		return fmt.Errorf("project name cannot be empty")
	}
	if tag == "" {
		return fmt.Errorf("tag cannot be empty")
	}
	projects, err := s.registry.Load()
	if err != nil {
		return err
	}
	project, ok := projects[projectName]
	if !ok {
		return fmt.Errorf("project %q not found", projectName)
	}
	for _, existing := range project.Tags {
		if existing == tag {
			return nil
		}
	}
	project.Tags = append(project.Tags, tag)
	sort.Strings(project.Tags)
	projects[projectName] = project
	return s.registry.Save(projects)
}

func (s *ProjectService) RemoveTag(projectName string, tag string) error {
	projectName = strings.TrimSpace(projectName)
	tag = strings.TrimSpace(tag)
	if projectName == "" {
		return fmt.Errorf("project name cannot be empty")
	}
	if tag == "" {
		return fmt.Errorf("tag cannot be empty")
	}
	projects, err := s.registry.Load()
	if err != nil {
		return err
	}
	project, ok := projects[projectName]
	if !ok {
		return fmt.Errorf("project %q not found", projectName)
	}
	if len(project.Tags) == 0 {
		return nil
	}
	filtered := project.Tags[:0]
	for _, existing := range project.Tags {
		if existing != tag {
			filtered = append(filtered, existing)
		}
	}
	project.Tags = filtered
	projects[projectName] = project
	return s.registry.Save(projects)
}

func (s *ProjectService) EnterProjectDir(projectPath string) error {
	if err := os.Chdir(projectPath); err != nil {
		return fmt.Errorf("chdir to %q failed: %w", projectPath, err)
	}
	return nil
}

func (s *ProjectService) ResolveWorktreeLeaf(name string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting current directory: %w", err)
	}
	cand1 := filepath.Join(cwd, "..", name)
	cand2 := filepath.Join(cwd, "..", strings.ReplaceAll(name, "/", "-"))

	if st, err := os.Stat(cand1); err == nil && st.IsDir() {
		return cand1, nil
	}
	if st, err := os.Stat(cand2); err == nil && st.IsDir() {
		return cand2, nil
	}
	return "", WorktreeNotFoundError{Name: name}
}
