package config

import (
    "fmt"
    "os"
    "path/filepath"
)

const ConfigFileName = ".wfconfig.yml"
const ExampleConfigYAML = `
default:
  log_level: "DEBUG"
  foreground: "nvim ."
  hooks:
    on_load:
      - "echo \"Welcome in your project!\""
  tmux:
    attach: false 
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

func AddWorkforgePrj(name string, gwt bool) error {
	workforgePath := os.Getenv("HOME") + "/" + WORK_FORGE_PRJ_CONFIG_DIR 
	absPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	fmt.Println("Adding project:", name, "path:", absPath, "gwt:", gwt)
	if _, err := os.Stat(workforgePath+"/"+WORK_FORGE_PRJ_CONFIG_FILE); os.IsNotExist(err) {
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
		fmt.Println("Workforge config:", workforgePath+"/"+WORK_FORGE_PRJ_CONFIG_FILE)
		fmt.Println("Loaded existing projects", projects)
		if err != nil {
			return fmt.Errorf("failed to load existing projects: %w", err)
		}else{
			projects[name] = Project{Name: name, Path: absPath, GitWorkTree: gwt}
			SaveProjects(workforgePath+"/"+WORK_FORGE_PRJ_CONFIG_FILE, projects)
		}
	}
	return nil
}

func AddWorkforgeLeaf(absLeafPath string) error {
    workforgePath := os.Getenv("HOME") + "/" + WORK_FORGE_PRJ_CONFIG_DIR

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
        projects = make(Projects)
    }

    cwd, err := os.Getwd()
    if err != nil {
        return fmt.Errorf("failed to get current directory: %w", err)
    }

    var baseName string
    var basePath string
    for name, p := range projects {
        if p.GitWorkTree && !isGWTLeaf(p.Path) && p.Path == cwd {
            baseName = name
            basePath = p.Path
            break
        }
    }

    if baseName == "" {
        parent := filepath.Dir(absLeafPath)
        for name, p := range projects {
            if p.GitWorkTree && !isGWTLeaf(p.Path) && p.Path == parent {
                baseName = name
                basePath = p.Path
                break
            }
        }
        _ = basePath 
    }

    leafName := filepath.Base(absLeafPath)
    key := leafName
    if baseName != "" {
        key = baseName + "/" + leafName
    }

    projects[key] = Project{Name: key, Path: absLeafPath, GitWorkTree: true}
    if err := SaveProjects(workforgePath+"/"+WORK_FORGE_PRJ_CONFIG_FILE, projects); err != nil {
        return err
    }
    return nil
}
