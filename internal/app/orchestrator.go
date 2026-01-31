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
	"workforge/internal/infra/exec"
	"workforge/internal/infra/tmux"
	"workforge/internal/util"
)

type Orchestrator struct {
	projects *project.ProjectService
	config   *config.ConfigService
	git      *appgit.Service
	hooks    *hook.HookService
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

	return &Orchestrator{
		projects: projectService,
		config:   configService,
		git:      gitService,
		hooks:    hookService,
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
	hookCtx := hook.HookContext{
		ShellCommands: cfg[currentProfile].Hooks.OnLoad,
		PluginConfigs: cfg[currentProfile].Extras,
		ProjectName:   resolvedProjectName,
	}
	if _, err := o.hooks.RunOnLoad(hookCtx); err != nil {
		return err
	}
	defer o.hooks.KillAllPlugins()

	shellRunInCtx := hook.HookContext{
		ShellCommands: cfg[currentProfile].Hooks.OnShellRunIn,
		PluginConfigs: cfg[currentProfile].Extras,
		ProjectName:   resolvedProjectName,
	}
	shellRunOutCtx := hook.HookContext{
		ShellCommands: cfg[currentProfile].Hooks.OnShellRunOut,
		PluginConfigs: cfg[currentProfile].Extras,
		ProjectName:   resolvedProjectName,
	}

	if cfg[currentProfile].Tmux == nil {
		foreground := cfg[currentProfile].Foreground
		if _, err := o.hooks.RunOnShellRunIn(shellRunInCtx); err != nil {
			return err
		}
		execErr := exec.RunSyncUserShell(foreground)
		if _, err := o.hooks.RunOnShellRunOut(shellRunOutCtx); err != nil {
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

	if _, err := o.hooks.RunOnShellRunIn(shellRunInCtx); err != nil {
		return err
	}
	if err := tmux.NewSession(path, sessionName, tmuxCfg.Attach, tmuxCfg.Windows); err != nil {
		return fmt.Errorf("failed to start tmux session: %w", err)
	}
	if _, err := o.hooks.RunOnShellRunOut(shellRunOutCtx); err != nil {
		return err
	}
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
	hookCtx := hook.HookContext{
		ShellCommands: cfg[currentProfile].Hooks.OnDelete,
		PluginConfigs: cfg[currentProfile].Extras,
		ProjectName:   resolvedProjectName,
	}
	if _, err := o.hooks.RunOnDelete(hookCtx); err != nil {
		return err
	}
	defer o.hooks.KillAllPlugins()

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
