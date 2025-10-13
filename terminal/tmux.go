package terminal

import (
	"fmt"
	"path/filepath"
)

func TmuxNewSession(path string, sessionName string, attach bool, windows []string) error {
	if sessionName == "" {
		sessionName = filepath.Base(path)
		if sessionName == "" {
			sessionName = "workforge"
		}
	}

	if err := RunSyncCommand("tmux", "new-session", "-s", sessionName, "-d"); err != nil {
		return err
	}

	if len(windows) > 0 {
		if err := RunSyncCommand("tmux", "send-keys", "-t", sessionName, windows[0], "C-m"); err != nil {
			return err
		}
		for _, win := range windows[1:] {
			fmt.Printf("Creating window: %s\n", win)

			if err := RunSyncCommand("tmux", "new-window", "-t", sessionName); err != nil {
				return err
			}
			if err := RunSyncCommand("tmux", "send-keys", "-t", sessionName, win, "C-m"); err != nil {
				return err
			}
		}
	}

	if attach {
		if err := RunSyncCommand("tmux", "attach", "-t", sessionName); err != nil {
			return err
		}
	}
	return nil
}
