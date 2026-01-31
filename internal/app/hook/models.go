package hook

// HookType identifies the type of hook event
type HookType string

const (
	HookOnLoad             HookType = "on_load"
	HookOnClose            HookType = "on_close"
	HookOnCreate           HookType = "on_create"
	HookOnDelete           HookType = "on_delete"
	HookOnShellRunIn       HookType = "on_shell_run_in"
	HookOnShellRunOut      HookType = "on_shell_run_out"
	HookOnPluginWakeup     HookType = "on_plugin_wakeup"
	HookOnError            HookType = "on_error"
	HookOnWarning          HookType = "on_warning"
	HookOnDebug            HookType = "on_debug"
	HookOnMessage          HookType = "on_message"
	HookOnTmuxSessionStart HookType = "on_tmux_session_start"
	HookOnTmuxWindow       HookType = "on_tmux_window"
)

const (
	FieldError   = "error"
	FieldWarning = "warning"
	FieldMessage = "message"
	FieldContext = "context"
	FieldSource  = "source"
	FieldSession = "session"
	FieldWindow  = "window"
	FieldCommand = "command"
)

type HookPayload struct {
	Project string         `json:"project"`
	Type    HookType       `json:"hook_type"`
	Data    map[string]any `json:"data,omitempty"`
	Config  map[string]any `json:"config,omitempty"`
}

func NewPayload(project string, hookType HookType) *HookPayload {
	return &HookPayload{
		Project: project,
		Type:    hookType,
		Data:    make(map[string]any),
	}
}

func (p *HookPayload) WithError(err error) *HookPayload {
	if err != nil {
		p.Data[FieldError] = err.Error()
	}
	return p
}

func (p *HookPayload) WithErrorMsg(msg string) *HookPayload {
	p.Data[FieldError] = msg
	return p
}

func (p *HookPayload) WithWarning(msg string) *HookPayload {
	p.Data[FieldWarning] = msg
	return p
}

func (p *HookPayload) WithMessage(msg string) *HookPayload {
	p.Data[FieldMessage] = msg
	return p
}

func (p *HookPayload) WithContext(ctx string) *HookPayload {
	p.Data[FieldContext] = ctx
	return p
}

func (p *HookPayload) WithSource(src string) *HookPayload {
	p.Data[FieldSource] = src
	return p
}

func (p *HookPayload) WithSession(session string) *HookPayload {
	p.Data[FieldSession] = session
	return p
}

func (p *HookPayload) WithWindow(idx int) *HookPayload {
	p.Data[FieldWindow] = idx
	return p
}

func (p *HookPayload) WithCommand(cmd string) *HookPayload {
	p.Data[FieldCommand] = cmd
	return p
}

func (p *HookPayload) WithField(key string, value any) *HookPayload {
	p.Data[key] = value
	return p
}

func (p *HookPayload) WithConfig(cfg map[string]any) *HookPayload {
	p.Config = cfg
	return p
}

type HookResult struct {
	PluginName string
	Response   string
	Error      error
}
