package app

import (
	"fmt"
	"os"
	"path/filepath"

	"workforge/internal/config"
	"workforge/internal/infra/exec"
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
	projects *ProjectService
	config   *ConfigService
	git      *GitService
}

func NewService() *Service {
	projects := NewProjectService()
	configService := NewConfigService()
	gitService := NewGitService(projects)
	return &Service{
		projects: projects,
		config:   configService,
		git:      gitService,
	}
}

func (s *Service) InitProject(url string, gwt bool) error {
	if url == "" {
		return s.initLocal(gwt)
	}
	return s.initFromURL(url, gwt)
}

func (s *Service) LoadProject(path string, gwt bool, profile *string) error {
	if err := s.projects.EnterProjectDir(path); err != nil {
		return err
	}
	cfg, err := s.config.LoadConfig(path, gwt)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}
	currentProfile, err := s.config.SelectProfile(cfg, profile)
	if err != nil {
		return err
	}
	log.SetLogLevelFromString(cfg[currentProfile].LogLevel)
	onLoad := cfg[currentProfile].Hooks.OnLoad
	log.Debug("using profile: %s", currentProfile)
	if err := s.config.RunHooks("on_load", onLoad, log.Debug); err != nil {
		return err
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
	if br, err := s.git.CurrentBranch(); err == nil && br != "" {
		sessionName = fmt.Sprintf("%s/%s", sessionBase, br)
	}
	if err := tmux.NewSession(path, sessionName, tmuxCfg.Attach, tmuxCfg.Windows); err != nil {
		return fmt.Errorf("failed to start tmux session: %w", err)
	}
	return nil
}

func (s *Service) RunOnDelete(projectPath string, isGWT bool, profile *string) error {
	if err := s.projects.EnterProjectDir(projectPath); err != nil {
		return err
	}
	cfg, err := s.config.LoadConfig(projectPath, isGWT)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}
	currentProfile, err := s.config.SelectProfile(cfg, profile)
	if err != nil {
		return err
	}
	onDelete := cfg[currentProfile].Hooks.OnDelete
	if err := s.config.RunHooks("on_delete", onDelete, log.Info); err != nil {
		return err
	}
	return nil
}

func (s *Service) SortedProjectEntries() ([]registry.ProjectEntry, error) {
	return s.projects.SortedProjectEntries()
}

func (s *Service) FindProjectEntry(name string) (registry.ProjectEntry, error) {
	return s.projects.FindProjectEntry(name)
}

func (s *Service) AddWorkTree(worktreePath string, branch string, createBranch bool, baseBranch string) error {
	return s.git.AddWorktree(worktreePath, branch, createBranch, baseBranch)
}

func (s *Service) RemoveWorktree(name string) (string, error) {
	leafPath, err := s.projects.ResolveWorktreeLeaf(name)
	if err != nil {
		return "", err
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
	return s.projects.AddProject(name, gwt, path)
}

func (s *Service) AddLeaf(absLeafPath string) error {
	return s.projects.AddLeaf(absLeafPath)
}

func (s *Service) initFromURL(url string, gwt bool) error {
	var entries []os.DirEntry
	repoName := util.RepoUrlToName(url)
	clonePath := repoName
	projectPath := repoName
	if gwt {
		projectPath = "."
	}
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

	if err := s.git.Clone(url, &clonePath); err != nil {
		return err
	}

	if gwt {
		branchName, err := s.git.CurrentBranchForPath(clonePath)
		if err != nil {
			return err
		}
		branchDir := s.git.WorktreeLeafDirName(branchName)
		if branchDir != "" && branchDir != clonePath {
			if _, err := os.Stat(branchDir); err == nil {
				return fmt.Errorf("destination %q already exists", branchDir)
			} else if !os.IsNotExist(err) {
				return fmt.Errorf("failed to check destination %q: %w", branchDir, err)
			}
			log.Info("Renaming cloned repo to %s", branchDir)
			if err := os.Rename(clonePath, branchDir); err != nil {
				return fmt.Errorf("failed to rename cloned repo: %w", err)
			}
			clonePath = branchDir
		}
		configFilePath := filepath.Join(clonePath, config.ConfigFileName)
		if _, err := os.Stat(configFilePath); err == nil {
			log.Info("Copying Workforge config from the cloned repo")
			if err := util.CopyFile(configFilePath, config.ConfigFileName); err != nil {
				return err
			}
		} else {
			if err := s.config.WriteExampleConfig(nil); err != nil {
				return err
			}
		}
	} else {
		if err := s.config.WriteExampleConfig(&clonePath); err != nil {
			return err
		}
	}

	return s.projects.AddProject(repoName, gwt, &projectPath)
}

func (s *Service) initLocal(gwt bool) error {
	log.Info("Initializing a new Workforge project")
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}
	repoName := filepath.Base(cwd)
	if err := s.config.WriteExampleConfig(nil); err != nil {
		return err
	}
	return s.projects.AddProject(repoName, gwt, nil)
}
