package hook

import (
	"fmt"

	"workforge/internal/app/plugin"
	"workforge/internal/infra/exec"
	"workforge/internal/infra/log"
)

type HookService struct {
	pluginSvc      *plugin.PluginService
	pluginRegistry *plugin.PluginRegistryService
}

func NewHookService(pluginSvc *plugin.PluginService, pluginRegistry *plugin.PluginRegistryService) *HookService {
	return &HookService{
		pluginSvc:      pluginSvc,
		pluginRegistry: pluginRegistry,
	}
}

func (s *HookService) RunOnLoad(ctx HookContext) ([]HookResult, error) {
	return s.runHook(HookOnLoad, ctx)
}

func (s *HookService) RunOnClose(ctx HookContext) ([]HookResult, error) {
	return s.runHook(HookOnClose, ctx)
}

func (s *HookService) RunOnCreate(ctx HookContext) ([]HookResult, error) {
	return s.runHook(HookOnCreate, ctx)
}

func (s *HookService) RunOnDelete(ctx HookContext) ([]HookResult, error) {
	return s.runHook(HookOnDelete, ctx)
}

func (s *HookService) runHook(hookType HookType, ctx HookContext) ([]HookResult, error) {
	if err := s.runShellHooks(hookType, ctx.ShellCommands); err != nil {
		return nil, err
	}

	results := s.runPluginHooks(hookType, ctx.PluginConfigs)
	return results, nil
}

func (s *HookService) runShellHooks(hookType HookType, commands []string) error {
	for i, cmd := range commands {
		log.Debug("running %s hook #%d: %s", hookType, i+1, cmd)
		if err := exec.RunSyncUserShell(cmd); err != nil {
			return fmt.Errorf("%s hook %d failed: %w", hookType, i+1, err)
		}
	}
	return nil
}

func (s *HookService) runPluginHooks(hookType HookType, pluginConfigs map[string]interface{}) []HookResult {
	plugins, err := s.pluginRegistry.List()
	if err != nil {
		return []HookResult{{Error: fmt.Errorf("load plugin registry: %w", err)}}
	}

	var results []HookResult
	hookName := string(hookType)

	for _, p := range plugins {
		if !hasHook(p.Hooks, hookName) {
			continue
		}

		if err := s.pluginSvc.Wakeup(p.Name); err != nil {
			results = append(results, HookResult{
				PluginName: p.Name,
				Error:      err,
			})
			continue
		}

		params := pluginConfigs[p.ConfigKey]
		resp, err := s.pluginSvc.Call(p.Name, hookName, params)
		if err != nil {
			results = append(results, HookResult{
				PluginName: p.Name,
				Error:      err,
			})
			continue
		}

		results = append(results, HookResult{
			PluginName: p.Name,
			Response:   string(resp),
		})
	}

	return results
}

func (s *HookService) KillAllPlugins() {
	s.pluginSvc.KillAll()
}

func hasHook(hooks []string, target string) bool {
	for _, h := range hooks {
		if h == target {
			return true
		}
	}
	return false
}
