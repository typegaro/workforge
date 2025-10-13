package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const ConfigFileName = ".wfconfig.yml"
const ExampleConfigYAML = `# Workforge configuration file (YAML)
# Add your own templates below, e.g. Node:
defoult:
  log_level: "DEBUG"
  foreground: "nvim ."
  hooks:
    on_load:
      - "echo \"Welcome in your project!\""
  tmux:
    attach: true
    session_name: "test_prj"
    windows:
      - "nvim ."
      - "nix run nixpkgs#htop"
`

func WriteExampleConfig(path *string) error {
	targetDir := "."
	if path != nil && *path != "" {
		targetDir = *path
	}
	if !filepath.IsAbs(targetDir) {
		abs, err := filepath.Abs(targetDir)
		if err == nil {
			targetDir = abs
		}
	}
	return os.WriteFile(filepath.Join(targetDir, ConfigFileName), []byte(ExampleConfigYAML), 0o644)
}

// gwt: git work tree
func AddWorkforgePrj(name string, path *string, gwt bool) error {
	projects, err := LoadProjects()
	if err != nil {
		return fmt.Errorf("failed to load existing projects: %w", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	projectPath := cwd
	if path != nil && *path != "" {
		if filepath.IsAbs(*path) {
			projectPath = filepath.Clean(*path)
		} else {
			projectPath = filepath.Join(cwd, *path)
		}
	}
	projects[name] = Project{Name: name, Path: projectPath, GitWorkTree: gwt}
	if err := SaveProjects(projects); err != nil {
		return err
	}
	return nil
}

// AddWorkforgeLeaf adds a newly created Git worktree leaf into the Workforge registry.
// The provided path should be the absolute filesystem path of the worktree directory.
// The entry is stored with GitWorkTree=false. If the current working directory matches
// a registered GitWorkTree root, the leaf will be keyed as "<baseName>/<leafName>"; otherwise
// it will use just "<leafName>" as the key.
func AddWorkforgeLeaf(absLeafPath string) error {
	projects, err := LoadProjects()
	if err != nil {
		return fmt.Errorf("failed to load existing projects: %w", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	var baseName string
	for name, p := range projects {
		if p.GitWorkTree && filepath.Clean(p.Path) == filepath.Clean(cwd) {
			baseName = name
			break
		}
	}

	leafName := filepath.Base(absLeafPath)
	key := leafName
	if baseName != "" {
		key = baseName + "/" + leafName
	}

	projects[key] = Project{Name: key, Path: filepath.Clean(absLeafPath), GitWorkTree: false}
	if err := SaveProjects(projects); err != nil {
		return err
	}
	return nil
}
