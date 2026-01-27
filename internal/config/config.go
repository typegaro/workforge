package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const ConfigFileName = ".wfconfig.yml"
const DefaultProfile = "default"
const ExampleConfigYAML = `
default:
  log_level: "DEBUG"
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
	if path == nil {
		return os.WriteFile(ConfigFileName, []byte(ExampleConfigYAML), 0o644)
	}
	return os.WriteFile(filepath.Join(*path, ConfigFileName), []byte(ExampleConfigYAML), 0o644)
}

func ResolveConfigPath(projectPath string, isGWT bool) string {
	if isGWT {
		return filepath.Join(projectPath, "..", ConfigFileName)
	}
	return filepath.Join(projectPath, ConfigFileName)
}

func LoadConfig(projectPath string, isGWT bool) (Config, error) {
	cfgPath := ResolveConfigPath(projectPath, isGWT)
	return LoadFile(cfgPath)
}

func SelectProfile(cfg Config, requested *string) (string, error) {
	if requested != nil && *requested != "" {
		if _, ok := cfg[*requested]; !ok {
			return "", fmt.Errorf("profile %q not found", *requested)
		}
		return *requested, nil
	}
	if len(cfg) == 0 {
		return "", fmt.Errorf("no profiles defined in config")
	}
	if len(cfg) == 1 {
		for k := range cfg {
			return k, nil
		}
	}
	if _, ok := cfg[DefaultProfile]; ok {
		return DefaultProfile, nil
	}
	return "", fmt.Errorf("multiple profiles defined; specify --profile")
}
