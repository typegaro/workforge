package config

import (
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type Config = map[string]Template

type Template struct {
	LogLevel   string `yaml:"log_level,omitempty"`
	Foreground string `yaml:"foreground,omitempty"`
	Hooks      Hooks  `yaml:"hooks,omitempty"`
	Tmux       *Tmux  `yaml:"tmux,omitempty"`
}

type Hooks struct {
	OnCreate []string `yaml:"on_create,omitempty"`
	OnLoad   []string `yaml:"on_load,omitempty"`
	OnClose  []string `yaml:"on_close,omitempty"`
	OnDelete []string `yaml:"on_delete,omitempty"`
}

type Tmux struct {
	Attach      bool     `yaml:"attach"`
	SessionName string   `yaml:"session_name,omitempty"`
	Windows     []string `yaml:"windows,omitempty"`
}

func Parse(r io.Reader) (Config, error) {
	var cfg Config
	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)
	if err := dec.Decode(&cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func LoadFile(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f)
}
