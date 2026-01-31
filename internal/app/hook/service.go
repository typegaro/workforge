package hook

import (
	"fmt"
	"strings"

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

func (s *HookService) RunOnShellRunIn(ctx HookContext) ([]HookResult, error) {
	return s.runHook(HookOnShellRunIn, ctx)
}

func (s *HookService) RunOnShellRunOut(ctx HookContext) ([]HookResult, error) {
	return s.runHook(HookOnShellRunOut, ctx)
}

func (s *HookService) RunOnError(payload ErrorPayload) []HookResult {
	return s.runPluginHooksWithPayload(HookOnError, payload)
}

func (s *HookService) RunOnMessage(payload MessagePayload) []HookResult {
	return s.runPluginHooksWithPayload(HookOnMessage, payload)
}

func (s *HookService) RunOnWarning(payload WarningPayload) []HookResult {
	return s.runPluginHooksWithPayload(HookOnWarning, payload)
}

func (s *HookService) RunOnDebug(payload DebugPayload) []HookResult {
	return s.runPluginHooksWithPayload(HookOnDebug, payload)
}

func (s *HookService) RunOnTmuxSessionStart(payload TmuxSessionPayload) []HookResult {
	return s.runPluginHooksWithPayload(HookOnTmuxSessionStart, payload)
}

func (s *HookService) RunOnTmuxWindow(payload TmuxWindowPayload) []HookResult {
	return s.runPluginHooksWithPayload(HookOnTmuxWindow, payload)
}

func (s *HookService) runHook(hookType HookType, ctx HookContext) ([]HookResult, error) {
	if err := s.runShellHooks(hookType, ctx.ShellCommands); err != nil {
		return nil, err
	}

	results := s.runPluginHooks(hookType, ctx.PluginConfigs, ctx.ProjectName)
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

func (s *HookService) runPluginHooks(hookType HookType, pluginConfigs map[string]interface{}, projectName string) []HookResult {
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
		payload := addProjectToPayload(projectName, params)
		resp, err := s.pluginSvc.Call(p.Name, hookName, payload)
		if err != nil {
			results = append(results, HookResult{
				PluginName: p.Name,
				Error:      err,
			})
			continue
		}

		response := cleanResponse(resp)
		if response != "" {
			fmt.Println(response)
		}

		results = append(results, HookResult{
			PluginName: p.Name,
			Response:   response,
		})
	}

	return results
}

func addProjectToPayload(projectName string, payload interface{}) interface{} {
	name := strings.TrimSpace(projectName)
	if payload == nil {
		return map[string]interface{}{"project": name}
	}
	if configMap, ok := payload.(map[string]interface{}); ok {
		out := make(map[string]interface{}, len(configMap)+1)
		for key, value := range configMap {
			out[key] = value
		}
		out["project"] = name
		return out
	}
	return map[string]interface{}{
		"project": name,
		"params":  payload,
	}
}

func (s *HookService) runPluginHooksWithPayload(hookType HookType, payload interface{}) []HookResult {
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

		resp, err := s.pluginSvc.Call(p.Name, hookName, payload)
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

func cleanResponse(resp []byte) string {
	s := strings.TrimSpace(string(resp))
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	return s
}
