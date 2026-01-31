READ ME BEFORE STARTING DEVELOPMENT

# FOLDERS
- `cmd/wf`: CLI entrypoint
- `internal/cli`: cobra commands
- `internal/app`: orchestration, domain services
- `internal/app/config`: YAML config parsing, profile selection
- `internal/app/project`: project registry and operations
- `internal/app/plugin`: plugin system (registry, installer, runtime)
- `internal/app/git`: git operations wrapper
- `internal/app/log`: logging with hook integration
- `internal/app/hook`: plugin lifecycle hooks
- `internal/app/terminal`: shell command execution with hook triggers
- `internal/infra`: OS adapters (exec, git, tmux, fs)
- `internal/util`: pure helpers

keep business logic in `internal/app`, OS interactions in `internal/infra`, and pure helpers in `internal/util`.

# Makefile
loook at `Makefile` for compile, test, lint, and other commands.

# NIX TOOLING
Use `nix shell nixpkgs#<tool>` for temporary tool installs.

# SERVICE STRUCTURE

Split packages by role: `models.go` (types), `registry.go` (persistence), `service.go` (logic).

**Create service when**: state, DI, multiple ops share deps, need mocking.
**Skip service when**: pure utils, single op, one method.

Inject dependencies via constructor. Registries handle CRUD, services handle logic.
take a look at `internal/app/` for example.

# LOGGING

Plese use `internal/app/log` for logging + hook integration already set up.
Do NOT use `internal/infra/log` in app layer. and avoid direct `fmt` or `log` calls.

# HOOKS
Use `internal/app/hook` to trigger plugin hooks at key points.

# Command EXECUTION
See `internal/app/terminal` for shell command execution with pre/post hooks.

