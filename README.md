# 🧰 Workforge (Go)

Workforge is a lightweight Go CLI that helps you open and manage local development projects. It integrates with tmux and supports Git Worktree layouts, so you can jump into a project (or a worktree) and have your environment come up exactly how you want it.

Status: early alpha. Interfaces and messages may change.

## ✨ Features

- Project registry stored under `~/.config/workforge/workforge.json`.
- Fuzzy finder to quickly open a project or a Git Worktree subdirectory.
- Per-project YAML config (`.wfconfig.yml`) to define on-load hooks and either a foreground command or a tmux session.
- Optional Git Worktree mode: register a base directory as a worktree root and open any subdirectory as a project.

## Prerequisites

- Go (per `go.mod`, Go 1.24+)
- `git` available in `PATH`
- `tmux` (only if you use tmux sessions)
- A POSIX shell (bash, zsh, fish, sh). The tool uses your `SHELL` env var.

## Install

Build from source (main package is at the repository root):

```bash
git clone <repo-url>
cd workforge
go mod tidy
go build -o wf .
```

Optionally install into `GOBIN`/`GOPATH/bin`:

```bash
go install .
```

Note: The compiled binary name is up to you (`-o wf`). The internal Cobra command currently shows as `mio-cli` in help output; functionality is unaffected.

## CLI Overview

- `init <repo-url> [path]`
  - Clone a repository and register it in Workforge. Also writes an example `.wfconfig.yml` in the cloned project.
  - Flags:
    - `-t, --gwt` — mark the project as a Git Worktree root. When set, Workforge treats subdirectories as openable “leaf” projects and looks for config one level up.

- `open`
  - Launch an interactive fuzzy finder listing all registered projects.
  - Entries are labeled as `[Repo]` (normal project) or `[GWT]` (a Git Worktree subdirectory).
  - Selecting a project loads its `.wfconfig.yml`, runs any `on_load` hooks, and then either:
    - starts a tmux session as configured; or
    - runs the configured `foreground` command directly.

- `load [dir]` (advanced)
  - Load a project directly by path. Prefer `open` for everyday use.

## Configuration: `.wfconfig.yml`

Place a `.wfconfig.yml` in your project root. Workforge looks for this file in the current directory when loading a normal project, or in the parent directory when opening a Git Worktree subdirectory.

Important: the current implementation expects a profile named `defoult` (note the spelling). That profile is used at load time.

Example configuration:

```yaml
defoult:
  log_level: "DEBUG"          # optional: currently used for verbose messages
  foreground: "nvim ."        # used if no tmux session is defined
  hooks:
    on_load:
      - "echo \"Welcome to your project!\""
  tmux:                        # optional: define a tmux session instead of foreground
    attach: true               # attach to the session after creation
    session_name: "my_project"
    windows:                   # commands to run, one per tmux window (first command runs in the first window)
      - "nvim ."
      - "htop"
```

Supported fields (current implementation):

- `foreground` — shell command to run if `tmux` is not provided.
- `hooks.on_load` — list of shell commands executed before `foreground` or `tmux`.
- `tmux.attach` — whether to attach after creating the session.
- `tmux.session_name` — tmux session name to create/use.
- `tmux.windows` — list of commands; the first runs in the first window, the rest each open in a new window.

Notes:

- Additional hook names (`on_create`, `on_close`, `on_delete`) exist in types but are not yet executed by the current code.
- When opening a Git Worktree subdirectory, Workforge loads config from `../.wfconfig.yml`.

## Git Worktree Mode

Register a repository as a Git Worktree root using `--gwt` during `init`. Workforge will:

- store the base path in the registry;
- list only first-level subdirectories in the fuzzy finder (not the base directory itself);
- treat each subdirectory as a selectable project, loading config from the parent.

This is convenient if you manage multiple worktrees under a common directory.

## Files and Paths

- Project config: `.wfconfig.yml` in the project root (or parent, for Git Worktree leaves).
- Registry: `~/.config/workforge/workforge.json` (auto-created on first use).

## Nix (optional)

A basic `flake.nix` is provided to enter a Go development shell:

```bash
nix develop
```

## Examples

Initialize and open a normal project:

```bash
# Clone and register
wf init https://github.com/org/repo.git

# Fuzzy-open and start the configured environment
wf open
```

Initialize a Git Worktree root and open a subdirectory:

```bash
# Clone and register as GWT root
wf init https://github.com/org/mono.git --gwt

# Create a worktree subdir yourself (outside of Workforge)
# then open it via fuzzy finder
wf open
```

## Known Limitations

- Early alpha: command names, help text, and messages are still evolving. Some messages are in Italian.
- Only `hooks.on_load` is executed today; other hooks are defined but unused.
- `list`/`remove` commands mentioned in earlier documentation are not implemented in this Go variant.
- Help output shows the command as `mio-cli`; build output name is whatever you pass to `-o`.

## License

TBD by repository owner.
