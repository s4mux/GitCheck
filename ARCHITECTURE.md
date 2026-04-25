# GitCheck – Architecture

## Stack

| Concern | Choice | Rationale |
|---|---|---|
| Language | Go | Single binary, no runtime dependency, good concurrency primitives |
| CLI framework | `cobra` | Standard in Go ecosystem, clean flag/command handling |
| Config parsing | `BurntSushi/toml` | Minimal, correct, no magic |
| Git operations | `os/exec` → `git` subprocess | Delegates to the actual Git binary; no reimplementation risk |

---

## Project Structure

```
gitcheck/
├── main.go
├── cmd/
│   └── root.go              # cobra root command, flag wiring
├── internal/
│   ├── config/
│   │   ├── config.go        # Config struct, defaults
│   │   └── loader.go        # File resolution, TOML parsing
│   ├── scanner/
│   │   ├── scanner.go       # Entry point: walk roots, apply ignore rules
│   │   └── ignore.go        # Path and pattern matching
│   ├── git/
│   │   ├── repo.go          # Repo and RepoStatus types
│   │   ├── status.go        # Uncommitted changes, untracked files
│   │   ├── sync.go          # Ahead/behind remote
│   │   └── submodules.go    # .gitmodules parsing
│   └── report/
│       ├── report.go        # Rendering logic
│       └── color.go         # TTY detection, color helpers
└── go.mod
```

---

## Module Responsibilities

### `config`

Loads and validates configuration. Resolves `~` in paths. Determines config file location via XDG convention. Exposes a single `Config` struct; nothing else in the codebase reads files or env vars for configuration.

```go
type Config struct {
    Scan ScanConfig
}

type ScanConfig struct {
    Roots  []string
    Ignore IgnoreConfig
}

type IgnoreConfig struct {
    Paths    []string
    Patterns []string
}
```

### `scanner`

Traverses the filesystem. Knows nothing about Git internals – only finds directories that look like repos (contains `.git`). Returns a flat list of discovered repo paths. Applies ignore rules before descending. Stops recursion when a repo is found.

```go
type Scanner struct {
    Ignore IgnoreRules
}

func (s *Scanner) Scan(roots []string) ([]string, error)
```

The scanner returns paths only. Status resolution is not its concern.

### `git`

All Git interactions live here. Spawns `git` subprocesses. Parses output. Returns typed results.

```go
type Repo struct {
    Path      string
    Status    RepoStatus
    Submodules []Repo      // recursive; empty if none
}

type RepoStatus struct {
    HasUncommitted bool
    HasUntracked   bool
    AheadBy        int
    BehindBy       int
    NoRemote       bool
    DetachedHEAD   bool
}
```

Key functions:

```go
func GetStatus(repoPath string) (RepoStatus, error)
func GetSubmodules(repoPath string) ([]string, error)  // returns paths
func BuildRepo(repoPath string, fetchRemote bool) (Repo, error)
```

`BuildRepo` composes status + submodule resolution into a complete `Repo`. Submodule paths are resolved recursively via the same function.

### `report`

Formats and prints. Knows about `Repo` and `RepoStatus`, nothing else. Accepts an `io.Writer` to stay testable.

```go
type Options struct {
    Verbose  bool
    UseColor bool
}

func Render(w io.Writer, repos []Repo, opts Options)
```

---

## Data Flow

```
main.go
  └─ cmd/root.go
        ├─ config.Load()          → Config
        ├─ scanner.Scan(roots)    → []string (repo paths)
        ├─ git.BuildRepo(path)    → []Repo   (parallel)
        └─ report.Render(repos)
```

---

## Concurrency

Repo status resolution is I/O bound (subprocess + disk). Repos are independent. Resolution runs in parallel via a worker pool.

```go
// Conceptual – implementation detail of cmd layer
results := make([]Repo, len(paths))
var wg sync.WaitGroup
sem := make(chan struct{}, runtime.NumCPU())

for i, path := range paths {
    wg.Add(1)
    go func(i int, path string) {
        defer wg.Done()
        sem <- struct{}{}
        defer func() { <-sem }()
        results[i], _ = git.BuildRepo(path, cfg.FetchRemote)
    }(i, path)
}
wg.Wait()
```

Pool size defaults to `runtime.NumCPU()`. Can be made configurable later if needed.

---

## Dependency Direction

```
cmd  →  config
cmd  →  scanner
cmd  →  git
cmd  →  report

scanner  →  (nothing internal)
git      →  (nothing internal)
report   →  git (types only)
```

No circular dependencies. `internal/` packages do not import each other except `report` consuming `git` types. If that coupling becomes a problem, shared types move to an `internal/domain` package.

---

## Error Strategy

- Errors from individual repos are collected and reported at the end, not fatal
- Config errors are fatal (exit 1) with a clear message
- `git` not in PATH is detected at startup, fatal
- All errors are written to stderr; report goes to stdout

---

## Testing Approach

- `config`: unit tests with temp TOML files
- `scanner`: unit tests with temp directory trees
- `git`: integration tests against real git repos created in `t.TempDir()`
- `report`: unit tests with captured `io.Writer` output

No mocking of the `git` binary – integration tests on real repos are more reliable and less brittle.

---

## Future Considerations

- `--json` output: `report.RenderJSON(w io.Writer, repos []Repo)` alongside `Render`; no structural change needed
- Individual repo entries in config: scanner accepts both roots and explicit paths; additive change
- `--watch` mode: periodic re-scan; wrapper around existing flow
