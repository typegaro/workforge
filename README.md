
<p align="center">
  <img src="assets/logo.png" alt="Workforge logo" width="300" />
 </p>

# 🧰 Workforge 
Tired of doing the exact same 4 setup steps every time you open a project? Same. Workforge is a tiny Go CLI that spins up your local dev environments with tmux and Git Worktree support. Hit `wf open`, pick a project, and-poof-windows, commands, and hooks greet you like you meant to be productive.

ALPHA WARNING: This is more MVP than finished software. Interfaces, messages, and behavior can change without notice.

## ✨ What It Does Well

- Keeps a project registry at `~/.config/workforge/workforge.json`.
- Fuzzy-finds projects or Git Worktree leaf directories and loads them fast.
- Per-project YAML config (`.wfconfig.yml`) with `on_load` hooks and either a `foreground` command or a tmux session.
- Git Worktree mode: register a base directory and open its leaves as individual projects.
- Worktree helpers: `wf add` to create/add worktrees, `wf rm` to remove them (with `on_delete` hooks).

## Requirements

- Go (from `go.mod`: Go 1.24+)
- `git` in `PATH`
- `tmux` (only if you want tmux sessions)
- A POSIX shell (`$SHELL`: bash, zsh, fish, sh)

## Install

Build locally (main package is at repo root):

```bash
git clone <repo-url>
cd workforge
go mod tidy
go build -o wf .
```

Install to `GOBIN`/`GOPATH/bin`:

```bash
go install .
```

Note: you choose the binary name via `-o wf`. The root command is `wf`.

## CLI Overview

- `wf init [repo-url]`
  - With URL: clone a repository and register it. Writes an example `.wfconfig.yml` into the cloned project.
  - Without URL: register the current directory and write an example `.wfconfig.yml` here.
  - Flags:
    - `-t, --gwt` - mark as a Git Worktree root. Subdirectories become selectable leaves; config is read from the parent.

- `wf open`
  - Launch a fuzzy finder with all registered entries.
  - Normal entries are plain projects; GWT entries are worktree leaves.
  - On selection: load `.wfconfig.yml`, run `on_load` hooks, then either
    - create/attach a tmux session as configured, or
    - run the `foreground` command.
  - If the config contains multiple profiles, a second fuzzy prompt lets you choose the profile.
  - If there is exactly one profile, it is selected automatically regardless of its name.

- `wf load [dir]` (advanced)
  - Load a path directly. For everyday use, prefer `open`.
  - Flags:
    - `-p, --profile` - select a profile defined in `.wfconfig.yml`.
  - If no profile is specified and the config has exactly one profile, that profile is used automatically.

- `wf add <name> [base-branch]`
  - Create a new worktree from an existing branch, or create a new branch + worktree.
  - Flags:
    - `-b` - create a new branch (prefixed by `--prefix`, default `feature`), optional base branch (default `main`).

- `wf rm <name>`
  - Remove a worktree and run any `on_delete` hooks first.

## Configuration: `.wfconfig.yml`

Place a `.wfconfig.yml` in your project root. In Git Worktree mode, config is read from the parent of the leaf directory.

Important (and yes, intentional): the default profile is spelled `default`. If you don’t specify a profile, this one is used.

Example:

```yaml
default:
  log_level: "DEBUG"          # optional: enables verbose messages
  foreground: "nvim ."        # used when tmux is not configured
  hooks:
    on_load:
      - "echo \"Welcome in your project!\""
  tmux:                        # optional: define a tmux session
    attach: false              # attach right after creating the session
    session_name: "my_project" # if empty, inferred from path/branch
    windows:                   # one command per window (first runs in first window)
      - "nvim ."
      - "nix run nixpkgs#htop"
```

Supported today:

- `foreground` - shell command to run when `tmux` is not provided.
- `hooks.on_load` - commands run before `foreground`/`tmux`.
- `hooks.on_delete` - commands run by `wf rm` before removal.
- `tmux.attach` - whether to attach after creating the session.
- `tmux.session_name` - tmux session name; if empty, inferred (and suffixed with the current branch like `repo/branch`).
- `tmux.windows` - list of commands, one per tmux window.

### Logging and Aesthetics

- Workforge prints colorful, icon-based messages inspired by npm-style output.
- Configure verbosity per profile via `log_level` in `.wfconfig.yml`.
- Accepted values: `DEBUG`, `INFO` (default), `WARN`, `ERROR`, `SILENT`.
- Examples:
  - `DEBUG` shows detailed steps (🐛), window creation, hooks.
  - `INFO` shows key actions (ℹ) and successes (✔).
  - `WARN` shows warnings (⚠), `ERROR` shows errors (✖).

Notes:

- In Worktree mode, config is loaded from `../.wfconfig.yml`.
- Currently executed hooks: `on_load` (open) and `on_delete` (remove).

## Git Worktree Mode

Using `--gwt` with `wf init` registers a worktree root. Workforge then:

- stores the base path in the registry;
- lists only first-level subdirectories in `wf open` (not the base itself);
- treats each subdirectory as an openable project and reads config from the parent.

Great for monorepos or multi-worktree flows.

## Files and Paths

- Project config: `.wfconfig.yml` (or `../.wfconfig.yml` for GWT leaves)
- Project registry: `~/.config/workforge/workforge.json` (auto-created on first use)

## Quick Examples

Normal project:

```bash
# Clone and register
wf init https://github.com/org/repo.git

# Fuzzy-open and load the setup
wf open
```

Worktree root + open a leaf:

```bash
# Clone and register as GWT root
wf init https://github.com/org/mono.git --gwt

# Create a worktree leaf with git, then open it
wf open
```

## Personal Workflow (tmux FTW)

- I open projects with `wf open`; if configured, Workforge creates/uses a tmux session per project.
- Session names include the branch when available, e.g. `repo/branch`, so you instantly know where you are.
- To jump between projects/branches fast: in tmux, press `Ctrl-b` then `s` to list sessions and select the one you want.
  - In practice: `Ctrl-b s` → pick → boom, you’re in the right project/branch.

## Strengths vs Limitations

Strengths

- Fast, minimal friction: fuzzy-find + auto-setup beat muscle memory.
- tmux integration: predictable windows and commands, branch-aware session names.
- Git Worktree aware: treat leaves as first-class projects without extra config duplication.
- Simple YAML: a tiny `.wfconfig.yml` drives your flow.

Limitations (embrace the alpha life)

- ALPHA: command names, help text, and messages may change.
- Only `on_load` and `on_delete` hooks are executed today; others are defined but unused.
- No dedicated `list` command yet (use `wf open` to browse).
- Requires a POSIX shell; your shell comes from `$SHELL`.

## Roadmap

- Richer tmux UX (layouts, panes, smarter attach)
- Registry management commands (list, rename, prune)
- More hooks wired up (`on_create`, `on_close`)
- Clearer help and consistent messaging
- Tests and tighter error handling

Contributions welcome - issues and PRs encouraged. Tasteful GIFs may or may not increase merge speed.
