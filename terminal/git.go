package terminal

import (
    "fmt"
    "strings"
)

func GitClone(repoURL string, destination *string) error {
	var err error
    if destination != nil{
        Info("Cloning %s into %s", repoURL, *destination)
        err = RunSyncCommand("git", "clone", repoURL, *destination)
    }else {
        Info("Cloning %s", repoURL)
        err = RunSyncCommand("git", "clone", repoURL)
    }
	if err != nil {
		return fmt.Errorf("failed to clone repository: %s", err)
	}
    Success("Repository cloned successfully")
    return nil
}
func AddNewWorkTree(name string,prefix string, branch string, newb bool) error {
	if newb {
		err := RunSyncCommand("git", "worktree", "add", "../"+name, "-b", prefix+"/"+name, branch)
		if err != nil {
			return fmt.Errorf("failed to add the new worktree: %s", err)
		}
	}
    Success("New worktree added successfully")
    return nil
}
func AddWorkTree(name string) error {
	folder_name := "../" + strings.ReplaceAll(name, "/", "-")
	err := RunSyncCommand("git", "worktree", "add", folder_name, name)
	if err != nil {
		return fmt.Errorf("failed to add the new worktree: %s", err)
	}
    Success("New worktree added successfully")
    return nil
}

// GitCurrentBranch returns the current git branch name for the working directory.
func GitCurrentBranch() (string, error) {
    out, err := RunOutput("git", "rev-parse", "--abbrev-ref", "HEAD")
    if err != nil {
        return "", fmt.Errorf("failed to get current branch: %w", err)
    }
    return out, nil
}
