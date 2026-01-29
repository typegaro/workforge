package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
	"workforge/internal/infra/exec"
	"workforge/internal/infra/log"
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

type Config = map[string]Template

type Template struct {
	LogLevel   string `yaml:"log_level,omitempty"`
	Foreground string `yaml:"foreground,omitempty"`
	Hooks      Hooks  `yaml:"hooks,omitempty"`
	Tmux       *Tmux  `yaml:"tmux,omitempty"`
}

type Hooks struct {
	OnCreate []string `yaml:"on_create,omitempty"`
	OnLoad   []string `yaml:"on_load,omitempty"`
	OnClose  []string `yaml:"on_close,omitempty"`
	OnDelete []string `yaml:"on_delete,omitempty"`
}

type Tmux struct {
	Attach      bool     `yaml:"attach"`
	SessionName string   `yaml:"session_name,omitempty"`
	Windows     []string `yaml:"windows,omitempty"`
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) LoadConfig(projectPath string, isGWT bool) (Config, error) {
	cfgPath := s.ResolveConfigPath(projectPath, isGWT)
	f, err := os.Open(cfgPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return s.parseConfig(f)
}

func (s *Service) parseConfig(r io.Reader) (Config, error) {
	var cfg Config
	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)
	if err := dec.Decode(&cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (s *Service) ResolveConfigPath(projectPath string, isGWT bool) string {
	if isGWT {
		return filepath.Join(projectPath, "..", ConfigFileName)
	}
	return filepath.Join(projectPath, ConfigFileName)
}

func (s *Service) WriteExampleConfig(path *string) error {
	if path == nil {
		return os.WriteFile(ConfigFileName, []byte(ExampleConfigYAML), 0o644)
	}
	return os.WriteFile(filepath.Join(*path, ConfigFileName), []byte(ExampleConfigYAML), 0o644)
}

func (s *Service) WriteConfig(path string, cfg Config) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	if err := enc.Encode(cfg); err != nil {
		return err
	}
	return enc.Close()
}

func (s *Service) SelectProfile(cfg Config, requested *string) (string, error) {
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

func (s *Service) RunHooks(label string, hooks []string, logFn func(string, ...any)) error {
	for i, cmd := range hooks {
		logFn("running %s hook #%d: %s", label, i+1, cmd)
		if err := exec.RunSyncUserShell(cmd); err != nil {
			return fmt.Errorf("%s hook %d failed: %w", label, i+1, err)
		}
	}
	return nil
}

func (s *Service) SetLogLevel(level string) {
	log.SetLogLevelFromString(level)
}
