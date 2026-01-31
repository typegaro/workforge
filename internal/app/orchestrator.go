package app

import (
	"fmt"
	"os"
	"path/filepath"

	"workforge/internal/app/config"
	appgit "workforge/internal/app/git"
	"workforge/internal/app/project"
	"workforge/internal/infra/exec"
	"workforge/internal/infra/log"
	"workforge/internal/infra/tmux"
	"workforge/internal/util"
)

type Orchestrator struct {
	projects *project.ProjectService
	config   *config.ConfigService
	git      *appgit.Service
}

func NewOrchestrator() *Orchestrator {
	projectService := project.NewService()
	configService := config.NewConfigService()
	gitService := appgit.NewService(projectService)
	return &Orchestrator{
		projects: projectService,
		config:   configService,
		git:      gitService,
	}
}

func (o *Orchestrator) Projects() *project.ProjectService {
	return o.projects
}

func (o *Orchestrator) Config() *config.ConfigService {
	return o.config
}

func (o *Orchestrator) Git() *appgit.Service {
	return o.git
}

func (o *Orchestrator) InitProject(url string, gwt bool) error {
	if url == "" {
		return o.initLocal(gwt)
	}
	return o.initFromURL(url, gwt)
}

func (o *Orchestrator) LoadProject(path string, gwt bool, profile *string) error {
	if err := o.projects.EnterProjectDir(path); err != nil {
		return err
	}
	cfg, err := o.config.LoadConfig(path, gwt)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}
	currentProfile, err := o.config.SelectProfile(cfg, profile)
	if err != nil {
		return err
	}
	o.config.SetLogLevel(cfg[currentProfile].LogLevel)
	onLoad := cfg[currentProfile].Hooks.OnLoad
	log.Debug("using profile: %s", currentProfile)
	if err := o.config.RunHooks("on_load", onLoad, log.Debug); err != nil {
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
	if br, err := o.git.CurrentBranch(); err == nil && br != "" {
		sessionName = fmt.Sprintf("%s/%s", sessionBase, br)
	}
	if err := tmux.NewSession(path, sessionName, tmuxCfg.Attach, tmuxCfg.Windows); err != nil {
		return fmt.Errorf("failed to start tmux session: %w", err)
	}
	return nil
}

func (o *Orchestrator) RunOnDelete(projectPath string, isGWT bool, profile *string) error {
	if err := o.projects.EnterProjectDir(projectPath); err != nil {
		return err
	}
	cfg, err := o.config.LoadConfig(projectPath, isGWT)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}
	currentProfile, err := o.config.SelectProfile(cfg, profile)
	if err != nil {
		return err
	}
	onDelete := cfg[currentProfile].Hooks.OnDelete
	if err := o.config.RunHooks("on_delete", onDelete, log.Info); err != nil {
		return err
	}
	return nil
}

func (o *Orchestrator) RemoveWorktree(name string) (string, error) {
	leafPath, err := o.projects.ResolveWorktreeLeaf(name)
	if err != nil {
		return "", err
	}
	return appgit.RemoveWorktree(leafPath, o.RunOnDelete)
}

func (o *Orchestrator) initFromURL(url string, gwt bool) error {
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

	if err := o.git.Clone(url, &clonePath); err != nil {
		return err
	}

	if gwt {
		branchName, err := o.git.CurrentBranchForPath(clonePath)
		if err != nil {
			return err
		}
		branchDir := o.git.WorktreeLeafDirName(branchName)
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
			if err := o.config.WriteExampleConfig(nil); err != nil {
				return err
			}
		}
	} else {
		if err := o.config.WriteExampleConfig(&clonePath); err != nil {
			return err
		}
	}

	return o.projects.AddProject(repoName, gwt, &projectPath)
}

func (o *Orchestrator) initLocal(gwt bool) error {
	log.Info("Initializing a new Workforge project")
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}
	repoName := filepath.Base(cwd)
	if err := o.config.WriteExampleConfig(nil); err != nil {
		return err
	}
	return o.projects.AddProject(repoName, gwt, nil)
}
