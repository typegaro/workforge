package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"workforge/terminal"
)

const defaultProfile = "defoult"

// LoadProject loads the Workforge environment for the provided path. It automatically
// discovers the correct configuration file (searching upwards for .wfconfig.yml) and
// detects whether the path belongs to a Git worktree registered in Workforge.
func LoadProject(path string, profile *string) error {
	targetPath := path
	if targetPath == "" {
		targetPath = "."
	}
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to resolve project path: %w", err)
	}

	projects, err := LoadProjects()
	if err != nil {
		return err
	}

	cfgPath, err := findConfigFile(absPath)
	if err != nil {
		return err
	}

	cfg, err := LoadFile(cfgPath)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	prof := defaultProfile
	if profile != nil && *profile != "" {
		prof = *profile
	}
	template, ok := cfg[prof]
	if !ok {
		return fmt.Errorf("profile %q not found in %s", prof, cfgPath)
	}

	if err := os.Chdir(absPath); err != nil {
		return fmt.Errorf("chdir to %q failed: %w", absPath, err)
	}

	isGWT := isGitWorktree(absPath, projects)
	if template.LogLevel == "DEBUG" {
		fmt.Println("using profile:", prof)
		fmt.Printf("configuration file: %s\n", cfgPath)
		if isGWT {
			fmt.Println("detected Git worktree project")
		}
	}

	if err := runHooks(template.Hooks.OnLoad, template.LogLevel); err != nil {
		return err
	}

	if template.Tmux == nil {
		if template.Foreground == "" {
			return nil
		}
		return terminal.RunSyncUserShell(template.Foreground)
	}

	return terminal.TmuxNewSession(absPath, template.Tmux.SessionName, template.Tmux.Attach, template.Tmux.Windows)
}

func runHooks(hooks []string, logLevel string) error {
	for i, cmd := range hooks {
		if logLevel == "DEBUG" {
			fmt.Println("running on_load hook #", i+1, ":", cmd)
		}
		if err := terminal.RunSyncUserShell(cmd); err != nil {
			return fmt.Errorf("hook %d failed: %w", i+1, err)
		}
	}
	return nil
}

func findConfigFile(start string) (string, error) {
	dir := filepath.Clean(start)
	for {
		candidate := filepath.Join(dir, ConfigFileName)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		} else if !errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("error checking %s: %w", candidate, err)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("%s not found starting from %s", ConfigFileName, start)
}

func isGitWorktree(path string, projects Projects) bool {
	cleaned := filepath.Clean(path)
	for _, p := range projects {
		if !p.GitWorkTree {
			continue
		}
		base := filepath.Clean(p.Path)
		if cleaned == base {
			return true
		}
		if strings.HasPrefix(cleaned, base+string(os.PathSeparator)) {
			return true
		}
	}
	return false
}
