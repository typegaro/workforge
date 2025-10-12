package config

import (
	"fmt"
	"os"
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

