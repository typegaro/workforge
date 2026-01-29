
<p align="center">
  <img src="assets/logo.png" alt="Workforge logo" width="300" />
 </p>

# 🧰 Workforge 

A CLI tool that automates your dev environment setup with tmux and Git Worktree support.

**ALPHA:** Interfaces may change without notice.

## What It Does

- Manages a project registry (`~/.config/workforge/workforge.json`)
- Launches projects with automated setup via `.wfconfig.yml`
- Supports Git Worktrees as first-class projects
- Creates/attaches tmux sessions with branch-aware naming

## Install

```bash
git clone <repo-url>
cd workforge
go build -o wf ./cmd/wf
# Optional: move to PATH
```

**Requirements:** Go 1.24+, git, tmux (optional), POSIX shell

## Quick Start

```bash
# Register a project
wf init https://github.com/org/repo.git

# List projects
wf list

# Open a project (runs setup from .wfconfig.yml)
wf open <project-name>
```

## Commands

| Command | Description |
|---------|-------------|
| `wf init [url]` | Clone and register a repo, or register current directory |
| `wf list` | List registered projects |
| `wf open <name>` | Open project (runs hooks, starts tmux/foreground command) |
| `wf add <branch>` | Create/add a git worktree |
| `wf rm <name>` | Remove a worktree (runs on_delete hooks) |

**Flags:**
- `--gwt` on `init`: Register as Git Worktree root
- `--profile` on `open`: Select config profile
- `-c, --create-branch` on `add`: Create branch if missing

## Configuration (`.wfconfig.yml`)

Place in project root:

```yaml
default:
  log_level: "INFO"           # DEBUG | INFO | WARN | ERROR | SILENT
  foreground: "nvim ."        # Command when not using tmux
  hooks:
    on_load: ["echo 'Hello'"] # Run before starting
    on_delete: []             # Run before worktree removal
  tmux:
    attach: false             # Auto-attach after creation
    session_name: "project"   # Default: inferred from path/branch
    windows:
      - "nvim ."
      - "git status"
```

**Worktree mode:** Config is read from `../.wfconfig.yml`

## Git Worktree Workflow

```bash
# Register base repo as worktree root
wf init https://github.com/org/repo.git --gwt

# Add worktrees
wf add feature-x --create-branch --base main

# List shows: repo/feature-x
wf list

# Open worktree (each gets its own tmux session: repo/feature-x)
wf open repo/feature-x
```

## Why This Exists

Stop manually:
1. `cd`-ing to projects
2. Starting tmux sessions
3. Opening your editor
4. Running setup commands

Just `wf open project` and everything's ready. With worktrees, each branch gets its own tmux session named `repo/branch` for easy switching (`Ctrl-b s`).

## Limitations

- ALPHA quality - expect changes
- Only `on_load` and `on_delete` hooks work currently
- No interactive project selector (use with fzf/rofi)
- Requires POSIX shell
