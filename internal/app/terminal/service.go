package terminal

import (
	"fmt"

	"workforge/internal/app/hook"
	applog "workforge/internal/app/log"
	"workforge/internal/infra/exec"
)

type TerminalService struct {
	hooks *hook.HookService
	log   *applog.LogService
}

func NewTerminalService(hooks *hook.HookService, log *applog.LogService) *TerminalService {
	return &TerminalService{hooks: hooks, log: log}
}

func (s *TerminalService) RunCommands(hookType hook.HookType, commands []string, project string, pluginConfigs map[string]any) error {
	for i, cmd := range commands {
		s.log.Debug("terminal", "running %s command #%d: %s", hookType, i+1, cmd)
		if err := exec.RunSyncUserShell(cmd); err != nil {
			return s.log.Error("terminal", fmt.Errorf("%s command %d failed: %w", hookType, i+1, err))
		}
	}

	payload := hook.NewPayload(project, hookType).WithConfig(pluginConfigs)
	s.hooks.Run(payload)

	return nil
}

func (s *TerminalService) RunForeground(cmd string) error {
	return exec.RunSyncUserShell(cmd)
}
