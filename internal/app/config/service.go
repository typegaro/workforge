package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
	"workforge/internal/infra/log"
)

type ConfigService struct{}

func NewConfigService() *ConfigService {
	return &ConfigService{}
}

func (s *ConfigService) LoadConfig(projectPath string, isGWT bool) (Config, error) {
	cfgPath := s.ResolveConfigPath(projectPath, isGWT)
	f, err := os.Open(cfgPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return s.parseConfig(f)
}

func (s *ConfigService) parseConfig(r io.Reader) (Config, error) {
	var cfg Config
	dec := yaml.NewDecoder(r)
	if err := dec.Decode(&cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (s *ConfigService) ResolveConfigPath(projectPath string, isGWT bool) string {
	if isGWT {
		return filepath.Join(projectPath, "..", ConfigFileName)
	}
	return filepath.Join(projectPath, ConfigFileName)
}

func (s *ConfigService) WriteExampleConfig(path *string) error {
	if path == nil {
		return os.WriteFile(ConfigFileName, []byte(ExampleConfigYAML), 0o644)
	}
	return os.WriteFile(filepath.Join(*path, ConfigFileName), []byte(ExampleConfigYAML), 0o644)
}

func (s *ConfigService) WriteConfig(path string, cfg Config) error {
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

func (s *ConfigService) SelectProfile(cfg Config, requested *string) (string, error) {
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

func (s *ConfigService) SetLogLevel(level string) {
	log.SetLogLevelFromString(level)
}
