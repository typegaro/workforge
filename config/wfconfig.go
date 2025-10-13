package config

import (
    "fmt"
    "os"
    "path/filepath"
    "workforge/terminal"
)

const DefaultProfile = "defoult"

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
        fmt.Println("error loading config:", err)
        return nil
    }

    currentProfile := DefaultProfile
    if profile != nil && *profile != "" {
        currentProfile = *profile
    }

    logLevel := cfg[currentProfile].LogLevel
    onLoad := cfg[currentProfile].Hooks.OnLoad
    if logLevel == "DEBUG" {
        fmt.Println("using profile:", currentProfile)
    }
    for i, cmd := range onLoad {
        if logLevel == "DEBUG" {
            fmt.Println("running on_load hook #", i+1, ":", cmd)
        }
        if err := terminal.RunSyncUserShell(cmd); err != nil {
            return fmt.Errorf("hook %d failed: %w", i+1, err)
        }
    }

    if cfg[currentProfile].Tmux == nil {
        foreground := cfg[currentProfile].Foreground
        return terminal.RunSyncUserShell(foreground)
    }

    tmux := cfg[currentProfile].Tmux
    // Build session name as <base>/<branch>.
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
