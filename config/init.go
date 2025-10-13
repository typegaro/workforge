package config

import (
    "fmt"
    "os"
    "path/filepath"
)

const ConfigFileName = ".wfconfig.yml"
const ExampleConfigYAML = `# Workforge configuration file (YAML)
# Add younoder own templates below, e.g. Node:
# Profile names 
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
	if path == nil  {
		return os.WriteFile("./"+ConfigFileName, []byte(ExampleConfigYAML), 0o644)
	}else{
		return os.WriteFile(*path+"/"+ConfigFileName, []byte(ExampleConfigYAML), 0o644)
	}
}


const WORK_FORGE_PRJ_CONFIG_DIR= ".config/workforge"
const WORK_FORGE_PRJ_CONFIG_FILE = "workforge.json" 

//gwt: git work tree
func AddWorkforgePrj(name string ,path *string, gwt bool) error {
	workforgePath := os.Getenv("HOME") + "/" + WORK_FORGE_PRJ_CONFIG_DIR 
	absPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	if path != nil {
		absPath = absPath + "/" + *path
	}
	if _, err := os.Stat(workforgePath); os.IsNotExist(err) {
		if err := os.MkdirAll(workforgePath, 0o755); err != nil {
			return fmt.Errorf("failed to create workforge config directory: %w", err)
		}
		if err := os.WriteFile(workforgePath+"/"+WORK_FORGE_PRJ_CONFIG_FILE, []byte(""), 0o644); err != nil {
			return fmt.Errorf("failed to create workforge config file: %w", err)
		}
		projects := Projects{name: {Name: name, Path: absPath, GitWorkTree: gwt}}
		SaveProjects(workforgePath+"/"+WORK_FORGE_PRJ_CONFIG_FILE, projects)
	} else {
		projects, err := LoadProjects(workforgePath + "/" + WORK_FORGE_PRJ_CONFIG_FILE)
		if err != nil {
			return fmt.Errorf("failed to load existing projects: %w", err)
		}else{
			projects[name] = Project{Name: name, Path: absPath, GitWorkTree: gwt}
			SaveProjects(workforgePath+"/"+WORK_FORGE_PRJ_CONFIG_FILE, projects)
		}
	}
	return nil
}

// AddWorkforgeLeaf adds a newly created Git worktree leaf into the Workforge registry.
// The provided path should be the absolute filesystem path of the worktree directory.
// The entry is stored with GitWorkTree=false. If the current working directory matches
// a registered GitWorkTree root, the leaf will be keyed as "<baseName>/<leafName>"; otherwise
// it will use just "<leafName>" as the key.
func AddWorkforgeLeaf(absLeafPath string) error {
    workforgePath := os.Getenv("HOME") + "/" + WORK_FORGE_PRJ_CONFIG_DIR

    // Ensure registry exists
    if _, err := os.Stat(workforgePath); os.IsNotExist(err) {
        if err := os.MkdirAll(workforgePath, 0o755); err != nil {
            return fmt.Errorf("failed to create workforge config directory: %w", err)
        }
        if err := os.WriteFile(workforgePath+"/"+WORK_FORGE_PRJ_CONFIG_FILE, []byte(""), 0o644); err != nil {
            return fmt.Errorf("failed to create workforge config file: %w", err)
        }
    }

    projects, err := LoadProjects(workforgePath + "/" + WORK_FORGE_PRJ_CONFIG_FILE)
    if err != nil {
        return fmt.Errorf("failed to load existing projects: %w", err)
    }

    // Try to find the base GWT project based on current working directory
    cwd, err := os.Getwd()
    if err != nil {
        return fmt.Errorf("failed to get current directory: %w", err)
    }

    var baseName string
    for name, p := range projects {
        if p.GitWorkTree && p.Path == cwd {
            baseName = name
            break
        }
    }

    leafName := filepath.Base(absLeafPath)
    key := leafName
    if baseName != "" {
        key = baseName + "/" + leafName
    }

    projects[key] = Project{Name: key, Path: absLeafPath, GitWorkTree: false}
    if err := SaveProjects(workforgePath+"/"+WORK_FORGE_PRJ_CONFIG_FILE, projects); err != nil {
        return err
    }
    return nil
}
