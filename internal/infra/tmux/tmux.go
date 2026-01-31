package tmux

import (
	execinfra "workforge/internal/infra/exec"
)

type WindowCallback func(session string, windowIndex int, command string)

func NewSession(path string, sessionName string, attach bool, windows []string, onWindow WindowCallback) error {
	if err := execinfra.RunSyncCommand("tmux", "new-session", "-s", sessionName, "-d"); err != nil {
		return err
	}
	if err := execinfra.RunSyncCommand("tmux", "send-keys", "-t", sessionName, windows[0], "C-m"); err != nil {
		return err
	}
	if onWindow != nil {
		onWindow(sessionName, 0, windows[0])
	}
	for i, win := range windows[1:] {
		if err := execinfra.RunSyncCommand("tmux", "new-window", "-t", sessionName); err != nil {
			return err
		}
		if err := execinfra.RunSyncCommand("tmux", "send-keys", "-t", sessionName, win, "C-m"); err != nil {
			return err
		}
		if onWindow != nil {
			onWindow(sessionName, i+1, win)
		}
	}
	if attach {
		if err := execinfra.RunSyncCommand("tmux", "attach", "-t", sessionName); err != nil {
			return err
		}
	}
	return nil
}

func KillSession(sessionName string) error {
	return execinfra.RunSyncCommand("tmux", "kill-session", "-t", sessionName)
}

func HasSession(sessionName string) bool {
	err := execinfra.RunSyncCommand("tmux", "has-session", "-t", sessionName)
	return err == nil
}
