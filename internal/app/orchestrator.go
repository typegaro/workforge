package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"workforge/internal/app/config"
	appgit "workforge/internal/app/git"
	"workforge/internal/app/hook"
	applog "workforge/internal/app/log"
	"workforge/internal/app/plugin"
	"workforge/internal/app/project"
	"workforge/internal/app/terminal"
	"workforge/internal/infra/tmux"
	"workforge/internal/util"
)

type Orchestrator struct {
	projects *project.ProjectService
	config   *config.ConfigService
	git      *appgit.Service
	hooks    *hook.HookService
	terminal *terminal.TerminalService
	log      *applog.LogService
}

func NewOrchestrator() *Orchestrator {
	projectService := project.NewService()
	configService := config.NewConfigService()
	gitService := appgit.NewService(projectService)

	pluginsDir := plugin.DefaultPluginsDir()
	pluginRegistry := plugin.NewPluginRegistryService(plugin.DefaultRegistryPath())
	pluginSvc := plugin.NewPluginService(pluginsDir, plugin.DefaultSocketsDir())
	hookService := hook.NewHookService(pluginSvc, pluginRegistry)
	logService := applog.NewLogService(hookService)
	terminalService := terminal.NewTerminalService(hookService, logService)

	return &Orchestrator{
		projects: projectService,
		config:   configService,
		git:      gitService,
		hooks:    hookService,
		terminal: terminalService,
		log:      logService,
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

func (o *Orchestrator) Hooks() *hook.HookService {
	return o.hooks
}

func (o *Orchestrator) Log() *applog.LogService {
	return o.log
}

func (o *Orchestrator) InitProject(url string, gwt bool) error {
	if url == "" {
		return o.initLocal(gwt)
	}
	return o.initFromURL(url, gwt)
}

func (o *Orchestrator) LoadProject(path string, gwt bool, profile *string, projectName string) error {
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
	o.log.Debug("load", "using profile: %s", currentProfile)

	resolvedProjectName := resolveProjectName(path, projectName)
	extras := cfg[currentProfile].Extras

	if err := o.terminal.RunCommands(hook.HookOnLoad, cfg[currentProfile].Hooks.OnLoad, resolvedProjectName, extras); err != nil {
		return err
	}

	if cfg[currentProfile].Tmux == nil {
		if err := o.terminal.RunCommands(hook.HookOnShellRunIn, cfg[currentProfile].Hooks.OnShellRunIn, resolvedProjectName, extras); err != nil {
			return err
		}
		execErr := o.terminal.RunForeground(cfg[currentProfile].Foreground)
		if err := o.terminal.RunCommands(hook.HookOnShellRunOut, cfg[currentProfile].Hooks.OnShellRunOut, resolvedProjectName, extras); err != nil {
			if execErr == nil {
				return err
			}
		}
		return execErr
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

	if err := o.terminal.RunCommands(hook.HookOnShellRunIn, cfg[currentProfile].Hooks.OnShellRunIn, resolvedProjectName, extras); err != nil {
		return err
	}

	onWindowCreated := func(session string, windowIndex int, command string) {
		payload := hook.NewPayload(resolvedProjectName, hook.HookOnTmuxWindow).
			WithSession(session).
			WithWindow(windowIndex).
			WithCommand(command)
		o.hooks.Run(payload)
	}

	if err := tmux.NewSession(path, sessionName, tmuxCfg.Attach, tmuxCfg.Windows, onWindowCreated); err != nil {
		return fmt.Errorf("failed to start tmux session: %w", err)
	}

	sessionPayload := hook.NewPayload(resolvedProjectName, hook.HookOnTmuxSessionStart).
		WithSession(sessionName)
	o.hooks.Run(sessionPayload)

	if err := o.terminal.RunCommands(hook.HookOnShellRunOut, cfg[currentProfile].Hooks.OnShellRunOut, resolvedProjectName, extras); err != nil {
		return err
	}
	return nil
}

func (o *Orchestrator) CloseProject(name string, profile *string) error {
	entry, err := o.projects.FindProjectEntry(name)
	if err != nil {
		return err
	}

	sessionName := name
	if !tmux.HasSession(sessionName) {
		return fmt.Errorf("no tmux session found for %q", name)
	}

	cfg, err := o.config.LoadConfig(entry.Path, entry.IsGWT)
	if err != nil {
		o.log.Warn("close", "could not load config: %v", err)
	}

	if cfg != nil {
		currentProfile, err := o.config.SelectProfile(cfg, profile)
		if err == nil {
			if err := o.terminal.RunCommands(hook.HookOnClose, cfg[currentProfile].Hooks.OnClose, entry.Name, cfg[currentProfile].Extras); err != nil {
				o.log.Warn("close", "on_close hook failed: %v", err)
			}
		}
	}

	if err := tmux.KillSession(sessionName); err != nil {
		return fmt.Errorf("failed to kill tmux session: %w", err)
	}

	o.log.Success("close", "closed project %s", name)
	return nil
}

func (o *Orchestrator) RunOnDelete(projectPath string, isGWT bool, profile *string, projectName string) error {
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

	resolvedProjectName := resolveProjectName(projectPath, projectName)
	if err := o.terminal.RunCommands(hook.HookOnDelete, cfg[currentProfile].Hooks.OnDelete, resolvedProjectName, cfg[currentProfile].Extras); err != nil {
		return err
	}

	return nil
}

func (o *Orchestrator) RemoveWorktree(name string) (string, error) {
	leafPath, err := o.projects.ResolveWorktreeLeaf(name)
	if err != nil {
		return "", err
	}
	return appgit.RemoveWorktree(leafPath, name, o.RunOnDelete)
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
					o.log.Warn("init", "This is a Workforge directory")
					o.log.Warn("init", "You can't clone a new repo here")
					return nil
				}
			}
		} else {
			o.log.Warn("init", "Directory not empty, aborting")
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
			o.log.Info("init", "Renaming cloned repo to %s", branchDir)
			if err := os.Rename(clonePath, branchDir); err != nil {
				return fmt.Errorf("failed to rename cloned repo: %w", err)
			}
			clonePath = branchDir
		}
		configFilePath := filepath.Join(clonePath, config.ConfigFileName)
		if _, err := os.Stat(configFilePath); err == nil {
			o.log.Info("init", "Copying Workforge config from the cloned repo")
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
	o.log.Info("init", "Initializing a new Workforge project")
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

func resolveProjectName(path string, projectName string) string {
	name := strings.TrimSpace(projectName)
	if name != "" {
		return name
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return filepath.Base(path)
	}
	return filepath.Base(absPath)
}
