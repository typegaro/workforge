package config

import (
	"fmt"
	"workforge/terminal"
)

func LoadProject(path *string,profile *string) error {
	cfg , err := LoadFile(*path + ".wfconfig.yml")
	if err != nil {
		fmt.Println("error loading config:", err)
		return nil
	}else{
		currentProfile := "defoult"
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
			terminal.RunSyncUserShell(foreground)
		}else{
			tmux := cfg[currentProfile].Tmux
			err := terminal.TmuxNewSession(tmux.SessionName, tmux.Attach, tmux.Windows)
			if err != nil {
				return fmt.Errorf("failed to start tmux session: %w", err)
			}
		}
	}
	return nil
}
