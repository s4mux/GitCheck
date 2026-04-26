# gitcheck

See the state of all your local Git repositories in one shot.

```
$ gitcheck

~/projects/api               [uncommitted] [ahead 2]
~/projects/frontend          [untracked]
~/work/client-x              [behind 3]
  └─ vendor/ui-kit           [uncommitted]
~/work/infra                 [no remote]
```

`gitcheck` scans your configured directories, finds every Git repository, and reports what needs attention — uncommitted changes, untracked files, ahead/behind remote, missing remotes, detached HEAD. Clean repos are silent. Submodules are shown indented under their parent.

---

## Install

**Go** (requires Go 1.21+):

```sh
go install github.com/s4mux/gitcheck@latest
```

**Build from source:**

```sh
git clone https://github.com/s4mux/gitcheck
cd gitcheck
go build -o gitcheck .
```

Move the binary somewhere on your `$PATH`, e.g. `/usr/local/bin/`.

---

## Configure

Create a config file at one of these locations (checked in order):

| Path | Notes |
|------|-------|
| `$XDG_CONFIG_HOME/gitcheck/config.toml` | if `$XDG_CONFIG_HOME` is set |
| `~/.config/gitcheck/config.toml` | default XDG location |
| `~/.gitcheck.toml` | simple single-file option |

Or pass `--config <path>` to use any file.

### Minimal config

```toml
[scan]
roots = [
    "~/projects",
    "~/work",
]
```

### Full config with ignore rules

```toml
[scan]
roots = [
    "~/projects",
    "~/work",
    "~/personal",
]

[scan.ignore]
# Skip these exact paths (supports ~)
paths = [
    "~/projects/archived",
    "~/projects/vendor/giant-monorepo",
]

# Skip directories whose name matches these glob patterns
patterns = [
    "*.bak",
    "temp-*",
    "node_modules",
]
```

**`roots`** — required, at least one entry. Each root is recursively scanned for `.git` directories.

**`ignore.paths`** — exact paths to skip entirely. Useful for archived projects or huge repos you don't care about.

**`ignore.patterns`** — glob patterns matched against the directory name (not the full path). Matched directories are skipped and not descended into.

---

## Usage

```
gitcheck [flags]
```

| Flag | Description |
|------|-------------|
| `--config <path>` | Use a specific config file |
| `--verbose` | Show all repos, including clean ones |
| `--fetch` | Fetch from remote before checking ahead/behind (slower, always accurate) |
| `--no-color` | Disable colored output |
| `--json` | Output results as JSON |
| `--version` | Print version and exit |
| `--help` | Show help |

### Examples

**Default** — show only repos that need attention:
```sh
gitcheck
```

**See everything**, including repos that are clean:
```sh
gitcheck --verbose
```

**Accurate remote sync** — fetch before reporting ahead/behind counts:
```sh
gitcheck --fetch
```

**Pipe-friendly output** — color is automatically disabled when stdout is not a TTY:
```sh
gitcheck | grep ahead
```

**JSON output** for scripting:
```sh
gitcheck --json | jq '.[] | select(.status.ahead_by > 0) | .path'
```

---

## Status reference

| Tag | Meaning |
|-----|---------|
| `[uncommitted]` | Staged or unstaged modifications |
| `[untracked]` | Files not tracked by Git |
| `[ahead N]` | N local commits not yet pushed |
| `[behind N]` | N remote commits not yet pulled |
| `[no remote]` | No remote configured |
| `[detached HEAD]` | Not on a named branch (never shown for submodules — always expected there) |

Without `--fetch`, ahead/behind counts reflect the last known remote state (last fetch/pull). Pass `--fetch` for live accuracy.

---

## How scanning works

- Roots are traversed recursively.
- The first `.git` found stops recursion — subdirectories of a repo are not scanned.
- Submodules are discovered via `.gitmodules` and reported indented under their parent repo.
- Unreadable directories produce a warning to stderr; scanning continues.

---

## Requirements

- Go 1.21+ (to build)
- `git` must be in `$PATH` at runtime
