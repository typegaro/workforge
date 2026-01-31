package cli

import (
	"fmt"

	"workforge/internal/app"
	"workforge/internal/app/hook"
	"workforge/internal/infra/log"
)

type LogService struct {
	orchestrator *app.Orchestrator
}

func NewLogService(o *app.Orchestrator) *LogService {
	return &LogService{orchestrator: o}
}

func (s *LogService) Error(context string, err error) {
	if err == nil {
		return
	}

	log.Error("%v", err)

	s.orchestrator.Hooks().RunOnError(hook.ErrorPayload{
		Error:   err.Error(),
		Context: context,
	})
	s.orchestrator.Hooks().KillAllPlugins()
}

func (s *LogService) ErrorMsg(context string, msg string, args ...any) {
	formatted := fmt.Sprintf(msg, args...)
	log.Error("%s", formatted)

	s.orchestrator.Hooks().RunOnError(hook.ErrorPayload{
		Error:   formatted,
		Context: context,
	})
	s.orchestrator.Hooks().KillAllPlugins()
}

func (s *LogService) Warn(context string, msg string, args ...any) {
	formatted := fmt.Sprintf(msg, args...)
	log.Warn("%s", formatted)

	s.orchestrator.Hooks().RunOnWarning(hook.WarningPayload{
		Warning: formatted,
		Context: context,
	})
}

func (s *LogService) Debug(context string, msg string, args ...any) {
	formatted := fmt.Sprintf(msg, args...)
	log.Debug("%s", formatted)

	s.orchestrator.Hooks().RunOnDebug(hook.DebugPayload{
		Message: formatted,
		Context: context,
	})
}

func (s *LogService) KillPlugins() {
	s.orchestrator.Hooks().KillAllPlugins()
}
