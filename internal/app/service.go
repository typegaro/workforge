package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"workforge/internal/config"
	"workforge/internal/infra/exec"
	"workforge/internal/infra/fs"
	"workforge/internal/infra/git"
	"workforge/internal/infra/log"
	"workforge/internal/infra/tmux"
	"workforge/internal/registry"
	"workforge/internal/util"
)

type WorktreeNotFoundError struct {
	Name string
}

func (e WorktreeNotFoundError) Error() string {
	return fmt.Sprintf("worktree %q not found", e.Name)
}

type OnDeleteError struct {
	Err error
}

func (e OnDeleteError) Error() string {
	return e.Err.Error()
}

type RemoveWorktreeError struct {
	Err error
}

func (e RemoveWorktreeError) Error() string {
	return e.Err.Error()
}

type Service struct {
	paths *fs.PathResolver
}

func NewService() *Service {
	return &Service{paths: fs.NewPathResolver()}
}

func (s *Service) InitProject(url string, gwt bool) error {
	if url == "" {
		return s.initLocal(gwt)
	}
	return s.initFromURL(url, gwt)
}

func (s *Service) LoadProject(path string, gwt bool, profile *string) error {
	if err := enterProjectDir(path); err != nil {
		return err
	}
	cfg, err := config.LoadConfig(path, gwt)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}
	currentProfile, err := config.SelectProfile(cfg, profile)
	if err != nil {
		return err
	}
	log.SetLogLevelFromString(cfg[currentProfile].LogLevel)
	onLoad := cfg[currentProfile].Hooks.OnLoad
	log.Debug("using profile: %s", currentProfile)
	for i, cmd := range onLoad {
		log.Debug("running on_load hook #%d: %s", i+1, cmd)
		if err := exec.RunSyncUserShell(cmd); err != nil {
			return fmt.Errorf("hook %d failed: %w", i+1, err)
		}
	}

	if cfg[currentProfile].Tmux == nil {
		foreground := cfg[currentProfile].Foreground
		return exec.RunSyncUserShell(foreground)
	}

	tmuxCfg := cfg[currentProfile].Tmux
	sessionBase := tmuxCfg.SessionName
	if sessionBase == "" {
		if gwt {
			sessionBase = filepath.Base(filepath.Dir(path))
		} else {
			sessionBase = filepath.Base(path)
		}
	}
	sessionName := sessionBase
	if br, err := git.GitCurrentBranch(); err == nil && br != "" {
		sessionName = fmt.Sprintf("%s/%s", sessionBase, br)
	}
	if err := tmux.NewSession(path, sessionName, tmuxCfg.Attach, tmuxCfg.Windows); err != nil {
		return fmt.Errorf("failed to start tmux session: %w", err)
	}
	return nil
}

func (s *Service) RunOnDelete(projectPath string, isGWT bool, profile *string) error {
	if err := enterProjectDir(projectPath); err != nil {
		return err
	}
	cfg, err := config.LoadConfig(projectPath, isGWT)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}
	currentProfile, err := config.SelectProfile(cfg, profile)
	if err != nil {
		return err
	}
	onDelete := cfg[currentProfile].Hooks.OnDelete
	for i, cmd := range onDelete {
		log.Info("running on_delete hook #%d: %s", i+1, cmd)
		if err := exec.RunSyncUserShell(cmd); err != nil {
			return fmt.Errorf("on_delete hook %d failed: %w", i+1, err)
		}
	}
	return nil
}

func (s *Service) SortedProjectEntries() ([]registry.ProjectEntry, error) {
	return registry.SortedProjectEntries()
}

func (s *Service) FindProjectEntry(name string) (registry.ProjectEntry, error) {
	return registry.FindProjectEntry(name)
}

func (s *Service) AddWorkTree(worktreePath string, branch string, createBranch bool, baseBranch string) error {
	return git.AddWorkTree(worktreePath, branch, createBranch, baseBranch)
}

func (s *Service) RemoveWorktree(name string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting current directory: %w", err)
	}
	cand1 := filepath.Join(cwd, "..", name)
	cand2 := filepath.Join(cwd, "..", strings.ReplaceAll(name, "/", "-"))

	leafPath := ""
	if st, err := os.Stat(cand1); err == nil && st.IsDir() {
		leafPath = cand1
	} else if st, err := os.Stat(cand2); err == nil && st.IsDir() {
		leafPath = cand2
	} else {
		return "", WorktreeNotFoundError{Name: name}
	}

	if err := s.RunOnDelete(leafPath, true, nil); err != nil {
		return "", OnDeleteError{Err: err}
	}
	if err := exec.RunSyncCommand("git", "worktree", "remove", leafPath); err != nil {
		return "", RemoveWorktreeError{Err: err}
	}
	return leafPath, nil
}

func (s *Service) AddProject(name string, gwt bool, path *string) error {
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

func (s *Service) AddLeaf(absLeafPath string) error {
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

func (s *Service) initFromURL(url string, gwt bool) error {
	var entries []os.DirEntry
	repoName := util.RepoUrlToName(url)
	path := repoName
	entries, err := os.ReadDir("./")
	if err != nil {
		return fmt.Errorf("directory error: %w", err)
	}
	if len(entries) > 0 {
		if !gwt {
			for _, entry := range entries {
				if entry.Name() == config.ConfigFileName {
					log.Warn("This is a Workforge directory")
					log.Warn("You can't clone a new repo here")
					return nil
				}
			}
		} else {
			log.Warn("Directory not empty, aborting")
			return nil
		}
	}

	if err := git.GitClone(url, &path); err != nil {
		return err
	}

	if gwt {
		configFilePath := filepath.Join(repoName, config.ConfigFileName)
		if _, err := os.Stat(configFilePath); err == nil {
			log.Info("Copying Workforge config from the cloned repo")
			if err := util.CopyFile(configFilePath, config.ConfigFileName); err != nil {
				return err
			}
		} else {
			if err := config.WriteExampleConfig(&path); err != nil {
				return err
			}
		}
	} else {
		if err := config.WriteExampleConfig(&path); err != nil {
			return err
		}
	}

	return s.AddProject(repoName, gwt, &path)
}

func (s *Service) initLocal(gwt bool) error {
	log.Info("Initializing a new Workforge project")
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}
	repoName := filepath.Base(cwd)
	if err := config.WriteExampleConfig(nil); err != nil {
		return err
	}
	return s.AddProject(repoName, gwt, nil)
}

func enterProjectDir(projectPath string) error {
	if err := os.Chdir(projectPath); err != nil {
		return fmt.Errorf("chdir to %q failed: %w", projectPath, err)
	}
	return nil
}
