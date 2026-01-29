package app

import (
	"fmt"

	"workforge/internal/config"
	"workforge/internal/infra/exec"
)

type ConfigService struct{}

func NewConfigService() *ConfigService {
	return &ConfigService{}
}

func (s *ConfigService) LoadConfig(projectPath string, isGWT bool) (config.Config, error) {
	return config.LoadConfig(projectPath, isGWT)
}

func (s *ConfigService) ResolveConfigPath(projectPath string, isGWT bool) string {
	return config.ResolveConfigPath(projectPath, isGWT)
}

func (s *ConfigService) WriteExampleConfig(path *string) error {
	return config.WriteExampleConfig(path)
}

func (s *ConfigService) WriteConfig(path string, cfg config.Config) error {
	return config.WriteFile(path, cfg)
}

func (s *ConfigService) SelectProfile(cfg config.Config, requested *string) (string, error) {
	return config.SelectProfile(cfg, requested)
}

func (s *ConfigService) RunHooks(label string, hooks []string, logFn func(string, ...any)) error {
	for i, cmd := range hooks {
		logFn("running %s hook #%d: %s", label, i+1, cmd)
		if err := exec.RunSyncUserShell(cmd); err != nil {
			return fmt.Errorf("%s hook %d failed: %w", label, i+1, err)
		}
	}
	return nil
}
