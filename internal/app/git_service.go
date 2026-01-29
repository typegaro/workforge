package app

import (
	"workforge/internal/infra/git"
)

type GitService struct {
	projects *ProjectService
}

func NewGitService(projects *ProjectService) *GitService {
	return &GitService{projects: projects}
}

func (s *GitService) Clone(repoURL string, destination *string) error {
	return git.GitClone(repoURL, destination)
}

func (s *GitService) AddWorktree(worktreePath string, branch string, createBranch bool, baseBranch string) error {
	return git.AddWorkTree(worktreePath, branch, createBranch, baseBranch)
}

func (s *GitService) AddWorktreeForProject(projectName string, branch string, createBranch bool, baseBranch string) error {
	path, _, err := s.projects.GetProjectPath(projectName)
	if err != nil {
		return err
	}
	return s.AddWorktree(path, branch, createBranch, baseBranch)
}

func (s *GitService) CurrentBranch() (string, error) {
	return git.GitCurrentBranch()
}

func (s *GitService) CurrentBranchForPath(path string) (string, error) {
	return git.GitCurrentBranchForPath(path)
}

func (s *GitService) WorktreeLeafDirName(name string) string {
	return git.WorktreeLeafDirName(name)
}
