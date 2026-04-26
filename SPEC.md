# GitCheck – Specification

## Overview

GitCheck is a CLI tool that scans configured root directories for local Git repositories and reports their status: uncommitted changes and synchronization with remote.

---

## Goals

- Give a developer a fast, clear overview of all local repos across multiple working directories
- Require zero manual maintenance of a repo list
- Be portable: single binary, no runtime dependency

## Non-Goals

- Git operations (commit, push, pull) – read-only tool
- GUI or TUI – CLI output only
- Remote authentication management

---

## Configuration

### Location (resolved in order)

1. `--config <path>` CLI flag
2. `$XDG_CONFIG_HOME/gitcheck/config.toml` (Unix) / `%APPDATA%\gitcheck\config.toml` (Windows)
3. `~/.gitcheck.toml`

### Format

```toml
[scan]
roots = [
    "~/projects",
    "~/work"
]

[scan.ignore]
paths = [
    "~/projects/archived",
    "~/projects/vendor/some-huge-monorepo"
]

patterns = [
    "*.bak",
    "temp-*"
]
```

### Rules

- `roots`: required, at least one entry
- `ignore.paths`: optional, exact path matches, resolved relative to home
- `ignore.patterns`: optional, glob patterns matched against directory name
- Paths support `~` expansion

---

## Scanning

### Algorithm

1. For each configured root, recursively traverse directories
2. If a directory contains `.git` → it is a Git repo; stop recursion into its subdirectories
3. If the repo contains `.gitmodules` → read submodules explicitly and treat each as a full repo
4. Ignore rules are applied at every level before descending

### Submodule handling

- Submodules are discovered via `.gitmodules`, not by scanning
- Each submodule undergoes the same status checks as a top-level repo, with one exception: `DetachedHEAD` is suppressed because submodules are always checked out to a specific commit — detached HEAD is their normal state, not a problem
- Nested submodules (submodules within submodules) are supported recursively

---

## Git Status Checks

For each discovered repo the following is determined:

| Check | Description |
|---|---|
| Uncommitted changes | Staged or unstaged modifications |
| Untracked files | Files not known to Git |
| Ahead of remote | Local commits not pushed |
| Behind remote | Remote commits not fetched/merged |
| No remote configured | Repo has no remote |
| Detached HEAD | Not on a named branch |

> **Note:** Remote sync checks (`ahead`/`behind`) require network access to fetch remote state. By default, last known remote state is used. Pass `--fetch` to trigger a live fetch before checking.

---

## Output

### Default mode

Only repos requiring attention are shown. Clean repos produce no output. If all repos are clean, a single summary line is printed:

```
All repositories clean.
```

### Verbose mode (`--verbose`)

All repos are listed regardless of status.

### Format

```
~/projects/myapp          [uncommitted] [ahead 2]
~/projects/libs/core      [behind 3]
~/work/client-x           [untracked]
  └─ vendor/dep-a         [uncommitted]    ← submodule
```

- Submodules are indented under their parent repo
- Status tags are shown inline
- Clean repos in verbose mode show `[ok]`

### Color

Color output is enabled by default when stdout is a TTY. Disabled automatically when piped. Can be forced off with `--no-color`.

---

## CLI Interface

```
gitcheck [flags]

Flags:
  --config <path>     Path to config file
  --verbose           Show all repos, not only those needing attention
  --fetch             Fetch from remote before checking ahead/behind state
  --no-color          Disable colored output
  --json              Output results as JSON
  --help              Show help
  --version           Show version
```

---

## Error Handling

- Unreadable directory: warning printed, scanning continues
- Invalid config: fatal error with clear message
- Repo with no remote: reported as status `[no remote]`, not an error
- Git not found in PATH: fatal error

---

## Future Scope

- Individual repo entries in config (in addition to root paths)
