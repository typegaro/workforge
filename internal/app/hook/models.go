package hook

type HookType string

const (
	HookOnLoad   HookType = "on_load"
	HookOnClose  HookType = "on_close"
	HookOnCreate HookType = "on_create"
	HookOnDelete HookType = "on_delete"
)

type HookContext struct {
	ShellCommands []string
	PluginConfigs map[string]interface{}
}

type HookResult struct {
	PluginName string
	Response   string
	Error      error
}
