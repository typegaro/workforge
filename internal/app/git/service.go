package git

import (
	"workforge/internal/app/project"
	"workforge/internal/infra/git"
)

type Service struct {
	projects *project.Service
}

func NewService(projects *project.Service) *Service {
	return &Service{projects: projects}
}

func (s *Service) Clone(repoURL string, destination *string) error {
	return git.GitClone(repoURL, destination)
}

func (s *Service) AddWorktree(worktreePath string, branch string, createBranch bool, baseBranch string) error {
	return git.AddWorkTree(worktreePath, branch, createBranch, baseBranch)
}

func (s *Service) AddWorktreeForProject(projectName string, branch string, createBranch bool, baseBranch string) error {
	path, _, err := s.projects.GetProjectPath(projectName)
	if err != nil {
		return err
	}
	return s.AddWorktree(path, branch, createBranch, baseBranch)
}

func (s *Service) CurrentBranch() (string, error) {
	return git.GitCurrentBranch()
}

func (s *Service) CurrentBranchForPath(path string) (string, error) {
	return git.GitCurrentBranchForPath(path)
}

func (s *Service) WorktreeLeafDirName(name string) string {
	return git.WorktreeLeafDirName(name)
}
