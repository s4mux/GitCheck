# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

`gitcheck` is a Go CLI tool that scans configured root directories for local Git repositories and reports their status (uncommitted changes, ahead/behind remote). Read-only — no Git writes. See `SPEC.md` for full feature spec.

## Commands

```bash
go build ./...          # build
go test ./...           # all tests
go test ./internal/...  # internal packages only
go test -run TestName ./internal/git/  # single test
go vet ./...
```

Binary name: `gitcheck`. Entry point: `main.go` → `cmd/root.go`.

Flags: `--config`, `--verbose`, `--fetch` (live remote fetch; default uses last known state), `--no-color`, `--json`.

## Architecture

Four internal packages with strict one-way dependency flow. See `ARCHITECTURE.md` for full detail.

```
cmd  →  config, scanner, git, report
report  →  git (types only)
scanner, git, config  →  (nothing internal)
```

**`internal/config`** — loads `config.toml` (XDG or `~/.gitcheck.toml`), resolves `~`, exposes a single `Config` struct. The only place that reads files/env for configuration.

**`internal/scanner`** — walks roots, applies ignore rules, returns `[]string` of repo paths. Stops recursion at `.git`. Knows nothing about Git internals.

**`internal/git`** — all Git interactions via `os/exec`. Key types: `Repo` (with recursive `Submodules []Repo`) and `RepoStatus`. Entry point is `BuildRepo(path, fetchRemote)`.

**`internal/report`** — formats and prints. Accepts `io.Writer`. No Git calls.

Concurrency: `cmd` resolves repo statuses in parallel via a worker pool (`runtime.NumCPU()` workers).

## Key Principles (from PRINCIPLES.md)

- **No cross-package leakage**: Git logic stays in `git`, ignore logic stays in `scanner/ignore.go`, formatting stays in `report`. If a package must change when its own responsibility hasn't changed, coupling is too high.
- **Explicit over implicit**: dependencies passed as parameters, no global state, errors returned explicitly (never `panic` except true programmer errors).
- **No preemptive abstraction**: interfaces appear when testability or substitutability requires them, not preventively.
- **Error strategy**: per-repo errors are collected and reported at the end (non-fatal); config errors and missing `git` binary are fatal (exit 1).
- **Testing**: `git` package tested against real temp repos created with `t.TempDir()` — no mocking of the `git` binary.

## Stack

- Language: Go
- CLI: `cobra`
- Config: `BurntSushi/toml`
- Git: `os/exec` → system `git` binary
