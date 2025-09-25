# 🧰 Workforge (Go)

**Workforge** is a **Go** CLI tool to quickly and efficiently manage a workflow based on **git worktree** and **tmux**.
It rewrites the original Rust idea in idiomatic Go, keeping the same UX for commands (`wf new|open|list|remove`).


## ✨ Features

* **Fast creation** of branch + worktree + (optional) tmux session

  ```bash
  wf new feature/login
  ```
* **Instant opening** of an existing branch with automatic resume/attach of the tmux session

  ```bash
  wf open feature/login
  ```
* **List active worktrees**

  ```bash
  wf list
  ```
* **Safe removal** (tmux + worktree + cleanup hooks)

  ```bash
  wf remove feature/login
  ```


## ⚙️ Project configuration

Each repo can define a `.wfconfig.yml` (or `.wfconfig.yaml`) file with custom rules.

Minimal and correct example:

```yaml
default:
  foreground: "vim ."
  background:
    - "uvicorn main:app --reload"
  hooks:
    on_create:
      - "poetry install"
      - "pre-commit install"
    on_open:
      - "docker compose up -d"
    on_close:
      - "docker compose down"

dev:
  tmux:
    attach: true
    panes:
      - "vim ."

dev-test:
  tmux:
    attach: false
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
```

### Supported hooks

* `on_create`: commands executed when creating the branch/worktree
* `on_open`: commands executed when opening a session
* `on_close`: cleanup when removing the branch/worktree
* `on_delete`: final operations after removal

### Tmux session templates

The `tmux` section allows defining `attach` (true/false) and a list of `panes` with commands to run in each pane.

> If the config file does not exist, Workforge falls back to sensible defaults.

---

## 🛠️ Tech stack

* **Go** for ease of distribution and toolchain
* [`spf13/cobra`](https://github.com/spf13/cobra) → CLI parsing
* [`gopkg.in/yaml.v3`](https://pkg.go.dev/gopkg.in/yaml.v3) → config management
* Standard API [`os/exec`](https://pkg.go.dev/os/exec) → external processes (`git`, `tmux`)

---

## 📦 Installation & build

```bash
git clone <repo>
cd workforge-go
go mod tidy
go build -o wf ./cmd/wf
```

To install in your PATH:

```bash
go install ./cmd/wf
```

## 🚀 Example workflow

```bash
# Create a new branch (worktree in .worktrees/feature/login)
wf new feature/login --template dev

# Open an existing branch (attach to tmux)
wf open feature/login

# See all active worktrees
wf list

# Clean up when done
wf remove feature/login
```

## 🔮 Roadmap

* [ ] Support for **branch_rules** (different configs for `feature/*`, `bugfix/*`, etc.)
* [ ] **Interactive TUI** to navigate between worktrees and sessions (fzf-like)
* [ ] Multi-project support (registry of repos with dedicated configs)
* [ ] Community-installable extensions/plugins

