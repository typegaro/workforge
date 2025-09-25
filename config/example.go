package config 

import (
	"os"
)

const ConfigFileName = ".wfconfig.yml"
const ExampleConfigYAML = `# Workforge configuration file (YAML)
# -----------------------------------
# Per-project templates used by: wf new|open|list|remove
# Each top-level key is a template name: e.g. "default", "dev", "dev-test".
# A template can define:
#   - foreground: one interactive cmd in the main pane (editor, REPL, etc.)
#   - background: long-running cmds (servers, watchers)
#   - tmux: session layout + attach behavior
#   - hooks: lifecycle cmds (on_create, on_open, on_close, on_delete)
#
# Notes:
# - All commands run from the worktree root.
# - Executed via the system shell (e.g., /bin/sh -c).
# - If this file is missing, Workforge falls back to sensible defaults.

default:
  # Interactive command in the primary pane (omit to skip).
  foreground: "vim ."  # e.g., "code .", "nvim .", "vim ."

  # Long-running side processes to start with the session (omit for none).
  background:
    - "uvicorn main:app --reload"
    # - "npm run dev"
    # - "docker compose up"

  # Lifecycle hooks
  hooks:
    on_create:
      - "poetry install"
      - "pre-commit install"
      # - "npm ci"
      # - "pip install -r requirements.txt"

    on_open:
      - "docker compose up -d"
      # - "asdf install"
      # - "direnv allow"

    on_close:
      - "docker compose down"
      # - "pkill -f uvicorn || true"

    # Runs after the worktree/session has been removed (optional).
    on_delete:
      - "true"  # keep or replace with your own steps

  # tmux session layout (omit to let Workforge decide).
  tmux:
    attach: true

    # Commands for panes (left-to-right / top-to-bottom).
    panes:
      - "vim ."
      # - "pytest -q"
      # - "htop"

# Simple dev template: attach immediately, single pane.
dev:
  tmux:
    attach: true
    panes:
      - "vim ."
  # background: []  # rely on hooks/on_open or start manually

# Heavier template for integration tests/services.
dev-test:
  tmux:
    attach: false  # create/resume but do not attach
    panes:
      - "vim ."
      - "docker compose up"
  background:
    - "pytest tests/"
  hooks:
    on_create:
      - "docker compose build"
    on_close:
      - "docker compose down"

# Add your own templates below, e.g. Node:
# node:
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
#     panes:
#       - "code ."
#       - "npm run dev"
`

func WriteExampleConfig(path *string) error {
	if path == nil  {*path = "./"}
	return os.WriteFile(*path+ConfigFileName, []byte(ExampleConfigYAML), 0o644)
}

