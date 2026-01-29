package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"workforge/internal/infra/fs"
	"workforge/internal/infra/log"
	"workforge/internal/registry"
)

type ProjectService struct {
	paths *fs.PathResolver
}

func NewProjectService() *ProjectService {
	return &ProjectService{paths: fs.NewPathResolver()}
}

func (s *ProjectService) EnsureRegistry() (string, error) {
	return registry.EnsureRegistry()
}

func (s *ProjectService) LoadProjects() (registry.Projects, error) {
	regPath, err := s.EnsureRegistry()
	if err != nil {
		return nil, err
	}
	projects, err := registry.LoadProjects(regPath)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func (s *ProjectService) SaveProjects(projects registry.Projects) error {
	regPath, err := s.EnsureRegistry()
	if err != nil {
		return err
	}
	return registry.SaveProjects(regPath, projects)
}

func (s *ProjectService) SortedProjectEntries() ([]registry.ProjectEntry, error) {
	return registry.SortedProjectEntries()
}

func (s *ProjectService) FindProjectEntry(name string) (registry.ProjectEntry, error) {
	return registry.FindProjectEntry(name)
}

func (s *ProjectService) GetProjectPath(name string) (string, bool, error) {
	entry, err := s.FindProjectEntry(name)
	if err != nil {
		return "", false, err
	}
	return entry.Path, entry.IsGWT, nil
}

func (s *ProjectService) AddProject(name string, gwt bool, path *string) error {
	regPath, err := registry.EnsureRegistry()
	if err != nil {
		return err
	}
	var absPath string
	if path != nil {
		absPath, err = s.paths.NormalizePath(*path)
		if err != nil {
			return fmt.Errorf("failed to resolve project path: %w", err)
		}
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		absPath, err = s.paths.NormalizePath(cwd)
		if err != nil {
			return fmt.Errorf("failed to resolve project path: %w", err)
		}
	}
	log.Info("Adding project: %s (path: %s, gwt: %t)", name, absPath, gwt)
	projects, err := registry.LoadProjects(regPath)
	log.Debug("Workforge config: %s", regPath)
	if err != nil {
		projects = make(registry.Projects)
	}
	log.Debug("Loaded existing projects: %+v", projects)
	projects[name] = registry.Project{Name: name, Path: absPath, GitWorkTree: gwt}
	if err := registry.SaveProjects(regPath, projects); err != nil {
		return err
	}
	return nil
}

func (s *ProjectService) AddLeaf(absLeafPath string) error {
	regPath, err := registry.EnsureRegistry()
	if err != nil {
		return err
	}
	projects, err := registry.LoadProjects(regPath)
	if err != nil {
		projects = make(registry.Projects)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	var baseName string
	for name, p := range projects {
		if p.GitWorkTree && !registry.IsGWTLeaf(p.Path) && p.Path == cwd {
			baseName = name
			break
		}
	}

	if baseName == "" {
		parent := filepath.Dir(absLeafPath)
		for name, p := range projects {
			if p.GitWorkTree && !registry.IsGWTLeaf(p.Path) && p.Path == parent {
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
	projects[key] = registry.Project{Name: key, Path: absLeafPath, GitWorkTree: true}
	if err := registry.SaveProjects(regPath, projects); err != nil {
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
	projects, err := s.LoadProjects()
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
	return s.SaveProjects(projects)
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
	projects, err := s.LoadProjects()
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
	return s.SaveProjects(projects)
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
