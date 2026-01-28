package git

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	execinfra "workforge/internal/infra/exec"
	"workforge/internal/infra/log"
)

func GitClone(repoURL string, destination *string) error {
	var err error
	if destination != nil {
		log.Info("Cloning %s into %s", repoURL, *destination)
		err = execinfra.RunSyncCommand("git", "clone", repoURL, *destination)
	} else {
		log.Info("Cloning %s", repoURL)
		err = execinfra.RunSyncCommand("git", "clone", repoURL)
	}
	if err != nil {
		return fmt.Errorf("failed to clone repository: %s", err)
	}
	log.Success("Repository cloned successfully")
	return nil
}

func AddWorkTree(worktreePath string, branch string, createBranch bool, baseBranch string) error {
	folderName := worktreeFolderName(branch)
	branchRef := strings.TrimSpace(strings.Trim(branch, "/"))
	if branchRef == "" {
		branchRef = branch
	}
	args := []string{"worktree", "add", folderName, branchRef}
	if createBranch {
		exists, err := branchExists(worktreePath, branchRef)
		if err != nil {
			return err
		}
		if !exists {
			if strings.TrimSpace(baseBranch) == "" {
				baseBranch = "main"
			}
			args = []string{"worktree", "add", folderName, "-b", branchRef, baseBranch}
		}
	}
	if err := runGitCommand(worktreePath, args...); err != nil {
		return fmt.Errorf("failed to add the new worktree: %s", err)
	}
	log.Success("New worktree added successfully")
	return nil
}

func WorktreeLeafDirName(name string) string {
	return worktreeLeafName(name)
}

func GitCurrentBranch() (string, error) {
	out, err := execinfra.RunOutput("git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return out, nil
}

func worktreeFolderName(name string) string {
	return filepath.Join("..", worktreeLeafName(name))
}

func worktreeLeafName(name string) string {
	cleaned := strings.TrimSpace(strings.Trim(name, "/"))
	if cleaned == "" {
		cleaned = strings.TrimSpace(name)
	}
	sanitized := strings.ReplaceAll(cleaned, "/", "-")
	sanitized = strings.Join(strings.Fields(sanitized), "-")
	if sanitized == "" {
		sanitized = "worktree"
	}
	return sanitized
}

func runGitCommand(worktreePath string, args ...string) error {
	if strings.TrimSpace(worktreePath) == "" {
		return execinfra.RunSyncCommand("git", args...)
	}
	cmdArgs := append([]string{"-C", worktreePath}, args...)
	return execinfra.RunSyncCommand("git", cmdArgs...)
}

func branchExists(worktreePath string, branch string) (bool, error) {
	branch = strings.TrimSpace(branch)
	if branch == "" {
		return false, fmt.Errorf("branch name cannot be empty")
	}
	args := []string{"show-ref", "--verify", "--quiet", fmt.Sprintf("refs/heads/%s", branch)}
	if strings.TrimSpace(worktreePath) != "" {
		args = append([]string{"-C", worktreePath}, args...)
	}
	_, err := execinfra.RunOutput("git", args...)
	if err == nil {
		return true, nil
	}
	var exitErr interface{ ExitCode() int }
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return false, nil
	}
	return false, fmt.Errorf("failed to check branch %q: %w", branch, err)
}
