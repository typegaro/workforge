package git

import (
	"workforge/internal/infra/exec"
)

type OnDeleteError struct {
	Err error
}

func (e OnDeleteError) Error() string {
	return e.Err.Error()
}

type RemoveWorktreeError struct {
	Err error
}

func (e RemoveWorktreeError) Error() string {
	return e.Err.Error()
}

func RemoveWorktree(leafPath string, projectName string, onDeleteFunc func(string, bool, *string, string) error) (string, error) {
	if err := onDeleteFunc(leafPath, true, nil, projectName); err != nil {
		return "", OnDeleteError{Err: err}
	}
	if err := exec.RunSyncCommand("git", "worktree", "remove", leafPath); err != nil {
		return "", RemoveWorktreeError{Err: err}
	}
	return leafPath, nil
}
