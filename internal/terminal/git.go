package terminal

import (
	"fmt"
	"strings"
)

func GitClone(repoURL string, destination *string) error {
	var err error
	if destination != nil {
		fmt.Printf("Clone %s in %s\n", repoURL, *destination)
		err = RunSyncCommand("git", "clone", repoURL, *destination)
	} else {
		fmt.Printf("Clone %s\n", repoURL)
		err = RunSyncCommand("git", "clone", repoURL)
	}
	if err != nil {
		return fmt.Errorf("failed to clone repository: %s", err)
	}
	fmt.Printf("Repository cloned successfully\n")
	return nil
}

func AddNewWorktree(name string, prefix string, branch string, newb bool) error {
	if newb {
		err := RunSyncCommand("git", "worktree", "add", "../"+name, "-b", prefix+"/"+name, branch)
		if err != nil {
			return fmt.Errorf("failed to add the new worktree: %s", err)
		}
	}
	fmt.Println("New worktree added successfully")
	return nil
}

func AddWorktree(name string) error {
	folderName := "../" + strings.ReplaceAll(name, "/", "-")
	err := RunSyncCommand("git", "worktree", "add", folderName, name)
	if err != nil {
		return fmt.Errorf("failed to add the new worktree: %s", err)
	}
	fmt.Println("New worktree added successfully")
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
