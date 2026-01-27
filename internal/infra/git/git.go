package git

import (
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

func AddNewWorkTree(name string, prefix string, baseBranch string) error {
	folderName := worktreeFolderName(prefix + "-" + name)
	branchName := worktreeBranchName(name, prefix)
	if branchName == "" {
		branchName = worktreeLeafName(name)
	}

	args := []string{"worktree", "add", folderName, "-b", branchName, baseBranch}
	if err := execinfra.RunSyncCommand("git", args...); err != nil {
		return fmt.Errorf("failed to add the new worktree: %s", err)
	}
	log.Success("New worktree added successfully")
	return nil
}

func AddWorkTree(name string) error {
	folderName := worktreeFolderName(name)
	branchRef := strings.TrimSpace(strings.Trim(name, "/"))
	if branchRef == "" {
		branchRef = name
	}
	if err := execinfra.RunSyncCommand("git", "worktree", "add", folderName, branchRef); err != nil {
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

func worktreeBranchName(name string, prefix string) string {
	parts := make([]string, 0, 2)
	cleanedPrefix := strings.Trim(prefix, "/")
	cleanedName := strings.Trim(name, "/")

	if cleanedPrefix != "" {
		parts = append(parts, cleanedPrefix)
	}
	if cleanedName != "" {
		parts = append(parts, cleanedName)
	}

	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "/")
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
