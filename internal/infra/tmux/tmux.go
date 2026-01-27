package tmux

import (
	execinfra "workforge/internal/infra/exec"
	"workforge/internal/infra/log"
)

func NewSession(path string, sessionName string, attach bool, windows []string) error {
	if err := execinfra.RunSyncCommand("tmux", "new-session", "-s", sessionName, "-d"); err != nil {
		return err
	}
	if err := execinfra.RunSyncCommand("tmux", "send-keys", "-t", sessionName, windows[0], "C-m"); err != nil {
		return err
	}
	for _, win := range windows[1:] {
		log.Step("Creating window: %s", win)
		if err := execinfra.RunSyncCommand("tmux", "new-window", "-t", sessionName); err != nil {
			return err
		}
		if err := execinfra.RunSyncCommand("tmux", "send-keys", "-t", sessionName, win, "C-m"); err != nil {
			return err
		}
	}
	if attach {
		if err := execinfra.RunSyncCommand("tmux", "attach", "-t", sessionName); err != nil {
			return err
		}
	}
	return nil
}
