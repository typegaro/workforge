package config 

import (
	"os"
)

const ConfigFileName = ".wfconfig.yml"
const ExampleConfigYAML = `# Workforge configuration file (YAML)
# Add younoder own templates below, e.g. Node:
# Profile names 
# defoult:
#   foreground: "code ."
#   background:
#     - "npm run dev"
#   hooks:
#     on_create:
#       - "npm ci"
#     on_close:
#       - "pkill -f node || true"
#   tmux:
#     attach: true
#     windows:
#       - "code ."
#       - "npm run dev"
`

func WriteExampleConfig(path *string) error {
	if path == nil  {
		return os.WriteFile("./"+ConfigFileName, []byte(ExampleConfigYAML), 0o644)
	}else{
		return os.WriteFile(*path+ConfigFileName, []byte(ExampleConfigYAML), 0o644)
	}
}

