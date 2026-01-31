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
- `internal/infra`: OS adapters (exec, git, tmux, fs)
- `internal/util`: pure helpers

Keep this section updated when adding/removing packages.

# MAKEFILE

| Target | Description |
|--------|-------------|
| `make build` | Build binary to `bin/wf` |
| `make build-verbose` | Build with debug/warn visible |
| `make test` | Run tests |
| `make fmt` | Format code |
| `make install` | Install to `/usr/local/bin` |
| `make uninstall` | Remove installed binary |
| `make clean` | Remove build artifacts |

# NIX TOOLING
Use `nix shell nixpkgs#<tool>` for temporary tool installs.

# SERVICE STRUCTURE

Split packages by role: `models.go` (types), `registry.go` (persistence), `service.go` (logic).

**Create service when**: state, DI, multiple ops share deps, need mocking.
**Skip service when**: pure utils, single op, one method.

Inject dependencies via constructor. Registries handle CRUD, services handle logic.
take a look at `internal/app/` for example.

# LOGGING

Use `internal/app/log.LogService` (injected via Orchestrator).

| Level | Default | Verbose |
|-------|---------|---------|
| Error/Info/Success | visible | visible |
| Warn/Debug | hidden | visible |

Hooks always trigger. Build verbose: `make build-verbose`

```go
o.log.Info("context", "msg %s", arg)
o.log.Error("context", err)
```

Do NOT use `internal/infra/log` in app layer.
