package hook

type HookType string

const (
	HookOnLoad    HookType = "on_load"
	HookOnClose   HookType = "on_close"
	HookOnCreate  HookType = "on_create"
	HookOnDelete  HookType = "on_delete"
	HookOnError   HookType = "on_error"
	HookOnWarning HookType = "on_warning"
	HookOnDebug   HookType = "on_debug"
	HookOnMessage HookType = "on_message"
)

type HookContext struct {
	ShellCommands []string
	PluginConfigs map[string]interface{}
}

type MessagePayload struct {
	Message string `json:"message"`
	Source  string `json:"source,omitempty"`
}

type ErrorPayload struct {
	Error   string `json:"error"`
	Context string `json:"context,omitempty"`
}

type WarningPayload struct {
	Warning string `json:"warning"`
	Context string `json:"context,omitempty"`
}

type DebugPayload struct {
	Message string `json:"message"`
	Context string `json:"context,omitempty"`
}

type HookResult struct {
	PluginName string
	Response   string
	Error      error
}
