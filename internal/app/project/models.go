package project

import "fmt"

const WorkForgeConfigFile = "workforge.json"

type Projects map[string]Project

type Project struct {
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	GitWorkTree bool     `json:"git_work_tree"`
	Tags        []string `json:"tags,omitempty"`
}

type ProjectEntry struct {
	Project
	IsGWT bool
}

type WorktreeNotFoundError struct {
	Name string
}

func (e WorktreeNotFoundError) Error() string {
	return fmt.Sprintf("worktree %q not found", e.Name)
}
