package hook

type HookType string

const (
	HookOnLoad         HookType = "on_load"
	HookOnClose        HookType = "on_close"
	HookOnCreate       HookType = "on_create"
	HookOnDelete       HookType = "on_delete"
	HookOnShellRunIn   HookType = "on_shell_run_in"
	HookOnShellRunOut  HookType = "on_shell_run_out"
	HookOnPluginWakeup HookType = "on_plugin_wakeup"
	HookOnError        HookType = "on_error"
	HookOnWarning      HookType = "on_warning"
	HookOnDebug        HookType = "on_debug"
	HookOnMessage      HookType = "on_message"
)

type HookContext struct {
	ShellCommands []string
	PluginConfigs map[string]interface{}
	ProjectName   string
}

type MessagePayload struct {
	Message string `json:"message"`
	Source  string `json:"source,omitempty"`
	Project string `json:"project"`
}

type ErrorPayload struct {
	Error   string `json:"error"`
	Context string `json:"context,omitempty"`
	Project string `json:"project"`
}

type WarningPayload struct {
	Warning string `json:"warning"`
	Context string `json:"context,omitempty"`
	Project string `json:"project"`
}

type DebugPayload struct {
	Message string `json:"message"`
	Context string `json:"context,omitempty"`
	Project string `json:"project"`
}

type HookResult struct {
	PluginName string
	Response   string
	Error      error
}
