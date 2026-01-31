package config

const ConfigFileName = ".wfconfig.yml"
const DefaultProfile = "default"
const ExampleConfigYAML = `
default:
  log_level: "DEBUG"
  hooks:
    on_load:
      - "echo \"Welcome in your project!\""
    on_shell_run_in:
      - "echo \"Starting shell session...\""
    on_shell_run_out:
      - "echo \"Shell session ended.\""
  tmux:
    attach: false 
    session_name: "test_prj"
    windows:
      - "nvim ."
      - "nix run nixpkgs#htop"
`

type Config = map[string]Template

type Template struct {
	LogLevel   string                 `yaml:"log_level,omitempty"`
	Foreground string                 `yaml:"foreground,omitempty"`
	Hooks      Hooks                  `yaml:"hooks,omitempty"`
	Tmux       *Tmux                  `yaml:"tmux,omitempty"`
	Extras     map[string]interface{} `yaml:",inline"`
}

type Hooks struct {
	OnCreate      []string `yaml:"on_create,omitempty"`
	OnLoad        []string `yaml:"on_load,omitempty"`
	OnClose       []string `yaml:"on_close,omitempty"`
	OnDelete      []string `yaml:"on_delete,omitempty"`
	OnShellRunIn  []string `yaml:"on_shell_run_in,omitempty"`
	OnShellRunOut []string `yaml:"on_shell_run_out,omitempty"`
}

type Tmux struct {
	Attach      bool     `yaml:"attach"`
	SessionName string   `yaml:"session_name,omitempty"`
	Windows     []string `yaml:"windows,omitempty"`
}
