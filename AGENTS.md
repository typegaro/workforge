# FOLDER 
- `cmd/wf`: binary bootstrap and CLI entrypoint wiring.
- `internal/cli`: cobra command definitions and flag binding.
- `internal/app`: orchestration and core command workflows.
- `internal/config`: YAML config parsing and profile selection.
- `internal/registry`: project registry read/write.
- `internal/infra`: OS/external tool adapters (exec, git, tmux, fs, log).
- `internal/util`: pure helpers (repo name, file copy).

# NIX TOOLING
- If a tool install is needed and approved, use `nix shell nixpkgs#<tool-1> nixpkgs#<tool-2>` for the shell environment.
