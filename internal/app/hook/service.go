package hook

import (
	"fmt"
	"strings"

	"workforge/internal/app/plugin"
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

func (s *HookService) Run(payload *HookPayload) []HookResult {
	return s.runPluginHooks(payload)
}

func (s *HookService) runPluginHooks(payload *HookPayload) []HookResult {
	plugins, err := s.pluginRegistry.List()
	if err != nil {
		return []HookResult{{Error: fmt.Errorf("load plugin registry: %w", err)}}
	}

	var results []HookResult
	hookName := string(payload.Type)

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

		wirePayload := s.buildWirePayload(payload, p.ConfigKey)

		resp, err := s.pluginSvc.Call(p.Name, hookName, wirePayload)
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

func (s *HookService) buildWirePayload(payload *HookPayload, configKey string) map[string]any {
	wire := map[string]any{
		"project":   strings.TrimSpace(payload.Project),
		"hook_type": string(payload.Type),
	}

	if len(payload.Data) > 0 {
		wire["data"] = payload.Data
	}

	if payload.Config != nil {
		if cfg, ok := payload.Config[configKey]; ok {
			wire["config"] = cfg
		}
	}

	return wire
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
