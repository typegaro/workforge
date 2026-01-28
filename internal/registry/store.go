package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

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

type FilterOptions struct {
	OnlyGWT      bool
	OnlyProjects bool
	Tags         []string
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

func ListProjects() (Projects, error) {
	resolver := fs.NewPathResolver()
	workforgePath, err := resolver.WorkforgeConfigDir()
	if err != nil {
		return nil, err
	}
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

func FilterProjectEntries(entries []ProjectEntry, opts FilterOptions) ([]ProjectEntry, error) {
	if opts.OnlyGWT && opts.OnlyProjects {
		return nil, fmt.Errorf("cannot combine --gwt with --projects")
	}
	filtered := make([]ProjectEntry, 0, len(entries))
	requiredTags := normalizeTags(opts.Tags)
	for _, entry := range entries {
		if opts.OnlyGWT && !entry.IsGWT {
			continue
		}
		if opts.OnlyProjects && entry.IsGWT {
			continue
		}
		if len(requiredTags) > 0 {
			entryTags := normalizeTags(entry.Tags)
			if !matchesTags(entryTags, requiredTags) {
				continue
			}
		}
		filtered = append(filtered, entry)
	}
	return filtered, nil
}

func AddProjectTags(name string, tags []string) error {
	return updateProjectTags(name, tags, nil)
}

func RemoveProjectTags(name string, tags []string) error {
	return updateProjectTags(name, nil, tags)
}

func updateProjectTags(name string, addTags []string, removeTags []string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}
	regPath, err := EnsureRegistry()
	if err != nil {
		return err
	}
	projects, err := LoadProjects(regPath)
	if err != nil {
		return err
	}
	project, ok := projects[name]
	if !ok {
		return fmt.Errorf("project %q not found", name)
	}
	current := normalizeTags(project.Tags)
	added := normalizeTags(addTags)
	removed := normalizeTags(removeTags)

	updated := applyTagUpdates(current, added, removed)
	project.Tags = updated
	projects[name] = project
	return SaveProjects(regPath, projects)
}

func applyTagUpdates(current []string, addTags []string, removeTags []string) []string {
	tagSet := make(map[string]struct{}, len(current))
	for _, tag := range current {
		tagSet[tag] = struct{}{}
	}
	for _, tag := range addTags {
		tagSet[tag] = struct{}{}
	}
	for _, tag := range removeTags {
		delete(tagSet, tag)
	}
	if len(tagSet) == 0 {
		return nil
	}
	updated := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		updated = append(updated, tag)
	}
	sort.Strings(updated)
	return updated
}

func normalizeTags(tags []string) []string {
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(strings.ToLower(tag))
		if tag == "" {
			continue
		}
		out = append(out, tag)
	}
	return out
}

func matchesTags(entryTags []string, required []string) bool {
	if len(required) == 0 {
		return true
	}
	tagSet := make(map[string]struct{}, len(entryTags))
	for _, tag := range entryTags {
		tagSet[tag] = struct{}{}
	}
	for _, tag := range required {
		if _, ok := tagSet[tag]; !ok {
			return false
		}
	}
	return true
}

func IsGWTLeaf(path string) bool {
	st, err := os.Stat(filepath.Join(path, ".git"))
	if err != nil {
		return false
	}
	return !st.IsDir()
}
