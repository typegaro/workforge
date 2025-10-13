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
func AddNewWorkTree(name string, prefix string, branch string, newb bool) error {
	if newb {
		err := RunSyncCommand("git", "worktree", "add", "../"+name, "-b", prefix+"/"+name, branch)
		if err != nil {
			return fmt.Errorf("failed to add the new worktree: %s", err)
		}
	}
	fmt.Println("New worktree added successfully")
	return nil
}
func AddWorkTree(name string) error {
	folder_name := "../" + strings.ReplaceAll(name, "/", "-")
	err := RunSyncCommand("git", "worktree", "add", folder_name, name)
	if err != nil {
		return fmt.Errorf("failed to add the new worktree: %s", err)
	}
	fmt.Println("New worktree added successfully")
	return nil
}
