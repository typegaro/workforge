package terminal 

import (
	"fmt"
)

func GitClone(repoURL string, destination *string) error {
	var err error
	if destination != nil{
		fmt.Printf("Clone %s in %s\n", repoURL, *destination)
		err = RunSyncCommand("git", "clone", repoURL, *destination)
	}else {
		fmt.Printf("Clone %s\n", repoURL)
		err = RunSyncCommand("git", "clone", repoURL)
	}
	if err != nil {
		return fmt.Errorf("failed to clone repository: %s", err)
	}
	fmt.Printf("Repository cloned successfully\n")
	return nil
}
