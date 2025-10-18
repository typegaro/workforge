package config

import (
    "fmt"
    "os"
    "path/filepath"
    "workforge/terminal"
)

const DefaultProfile = "default"

func ResolveConfigPath(projectPath string, isGWT bool) string {
    if isGWT {
        return filepath.Join(projectPath, "..", ConfigFileName)
    }
    return filepath.Join(projectPath, ConfigFileName)
}

func EnterProjectDir(projectPath string) error {
    if err := os.Chdir(projectPath); err != nil {
        return fmt.Errorf("chdir to %q failed: %w", projectPath, err)
    }
    return nil
}

func LoadConfig(projectPath string, isGWT bool) (Config, error) {
    cfgPath := ResolveConfigPath(projectPath, isGWT)
    return LoadFile(cfgPath)
}

func LoadProject(path string, gwt bool, profile *string) error {
    if err := EnterProjectDir(path); err != nil {
        return err
    }

    cfg, err := LoadConfig(path, gwt)
    if err != nil {
        terminal.Error("error loading config: %v", err)
        return nil
    }

    currentProfile := DefaultProfile
    if profile != nil && *profile != "" {
        currentProfile = *profile
    } else if len(cfg) == 1 {
        for k := range cfg { // pick the only defined profile
            currentProfile = k
        }
    }

    logLevel := cfg[currentProfile].LogLevel
    // configure global logger once profile is known
    terminal.SetLogLevelFromString(logLevel)
    onLoad := cfg[currentProfile].Hooks.OnLoad
    terminal.Debug("using profile: %s", currentProfile)
    for i, cmd := range onLoad {
        terminal.Debug("running on_load hook #%d: %s", i+1, cmd)
        if err := terminal.RunSyncUserShell(cmd); err != nil {
            return fmt.Errorf("hook %d failed: %w", i+1, err)
        }
    }

    if cfg[currentProfile].Tmux == nil {
        foreground := cfg[currentProfile].Foreground
        return terminal.RunSyncUserShell(foreground)
    }

    tmux := cfg[currentProfile].Tmux
    sessionBase := tmux.SessionName
    if sessionBase == "" {
        if gwt {
            sessionBase = filepath.Base(filepath.Dir(path))
        } else {
            sessionBase = filepath.Base(path)
        }
    }
    sessionName := sessionBase
    if br, err := terminal.GitCurrentBranch(); err == nil && br != "" {
        sessionName = fmt.Sprintf("%s/%s", sessionBase, br)
    }
    if err := terminal.TmuxNewSession(path, sessionName, tmux.Attach, tmux.Windows); err != nil {
        return fmt.Errorf("failed to start tmux session: %w", err)
    }
    return nil
}

func RunOnDelete(projectPath string, isGWT bool, profile *string) error {
    if err := EnterProjectDir(projectPath); err != nil {
        return err
    }
    cfg, err := LoadConfig(projectPath, isGWT)
    if err != nil {
        terminal.Error("error loading config: %v", err)
        return nil
    }
    currentProfile := DefaultProfile
    if profile != nil && *profile != "" {
        currentProfile = *profile
    } else if len(cfg) == 1 {
        for k := range cfg { // pick the only defined profile
            currentProfile = k
        }
    }
    onDelete := cfg[currentProfile].Hooks.OnDelete
    for i, cmd := range onDelete {
        terminal.Info("running on_delete hook #%d: %s", i+1, cmd)
        if err := terminal.RunSyncUserShell(cmd); err != nil {
            return fmt.Errorf("on_delete hook %d failed: %w", i+1, err)
        }
    }
    return nil
}
