# FOLDER 
- `cmd/wf`: binary bootstrap and CLI entrypoint wiring.
- `internal/cli`: cobra command definitions and flag binding.
- `internal/app`: orchestration and core command workflows.
- `internal/app/config`: YAML config parsing and profile selection.
- `internal/app/project`: project registry and project operations.
- `internal/app/plugin`: plugin system (registry, installer, runtime).
- `internal/app/git`: git operations wrapper.
- `internal/infra`: OS/external tool adapters (exec, git, tmux, fs, log).
- `internal/util`: pure helpers (repo name, file copy).

# NIX TOOLING
- If a tool install is needed and approved, use `nix shell nixpkgs#<tool-1> nixpkgs#<tool-2>` for the shell environment.

# SERVICE STRUCTURE

## File Organization Pattern

When a package grows, split by role:

```
internal/app/<domain>/
├── models.go    # Types, constants, error types (no logic)
├── registry.go  # Persistence: load/save from JSON/YAML files
├── service.go   # Business logic: operations on domain objects
├── installer.go # (optional) Setup/teardown operations
```

## When to Create a Service

**DO create a service when:**
- The domain has state or needs dependency injection
- Multiple operations share the same dependencies (paths, other services)
- You need to mock/test the behavior
- The CLI needs to call multiple related operations

**DON'T create a service when:**
- It's pure utility functions with no state (use `util/` or package-level functions)
- It's a single operation that doesn't need shared context
- It would be a service with only one method

## Naming Convention

```go
// Package: internal/app/config
type ConfigService struct { ... }
func NewConfigService() *ConfigService { ... }

// Package: internal/app/project  
type ProjectService struct { ... }
type ProjectRegistryService struct { ... }  // when multiple services in same package

// Package: internal/app/plugin
type PluginService struct { ... }
type PluginRegistryService struct { ... }
type PluginInstallerService struct { ... }
```

## Service Dependencies

Services can depend on other services. Inject via constructor:

```go
type ProjectService struct {
    registry *ProjectRegistryService  // internal dependency
}

type GitService struct {
    projects *ProjectService  // cross-package dependency
}

func NewGitService(projects *ProjectService) *GitService {
    return &GitService{projects: projects}
}
```

## Models File

Keep types separate from logic. `models.go` contains:
- Struct definitions with JSON/YAML tags
- Type aliases (`type Config = map[string]Template`)
- Constants
- Error types with `Error()` method
- No business logic

```go
// models.go
type Project struct {
    Name string `json:"name"`
    Path string `json:"path"`
}

type ProjectNotFoundError struct {
    Name string
}

func (e ProjectNotFoundError) Error() string {
    return fmt.Sprintf("project %q not found", e.Name)
}
```

## Registry vs Service

| Registry | Service |
|----------|---------|
| CRUD on persistent storage | Business logic and workflows |
| Load(), Save(), Add(), Remove() | FindByName(), Validate(), Process() |
| Knows about file paths and formats | Knows about domain rules |
| No domain logic | Uses registry internally |
