# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.1] - 2026-05-02

### Added

- Interactive TTY progress display: shows "Scanning for repositories..." during the directory walk, then a progress bar and the path of the most recently analyzed repo during parallel status resolution; the display erases itself before the final report is printed
- Progress is automatically suppressed when stdout is not a TTY (pipes, `--json`, redirects) — no ANSI codes leak into captured output

## [0.1.0] - 2026-04-26

### Added

- Recursive scanning of configured root directories for Git repositories
- Status checks per repository: uncommitted changes, untracked files, ahead/behind remote, no remote configured, detached HEAD
- Submodule support: submodules are discovered via `.gitmodules`, resolved recursively, and shown indented under their parent repo; detached HEAD is suppressed for submodules (always expected, never a problem)
- Terminal output with color (auto-disabled when stdout is not a TTY) and `--verbose` mode to show clean repos
- JSON output via `--json` for scripting and piping
- `--fetch` flag to trigger a live remote fetch before reporting ahead/behind counts
- `--no-color` flag to force-disable color
- `--version` flag; version is injected at build time via GoReleaser
- First-run interactive setup: when no config is found, `gitcheck` prompts for a scan root and saves the config automatically; returns a clear error when stdin is not a terminal
- `config` subcommand with `list`, `add`, and `remove` to manage scan roots without editing TOML by hand
- XDG-compliant config file location: `$XDG_CONFIG_HOME/gitcheck/config.toml` (Linux/macOS) / `%APPDATA%\gitcheck\config.toml` (Windows), with `~/.gitcheck.toml` as fallback
- Ignore rules: exact paths (`ignore.paths`) and glob patterns matched against directory names (`ignore.patterns`)
- Active config file path printed to stderr on each run (suppressed with `--json`)
- Parallel repository resolution via a worker pool sized to `runtime.NumCPU()`
- Pre-built binaries for Linux, macOS, and Windows via GoReleaser and GitHub Actions

[Unreleased]: https://github.com/s4mux/gitcheck/compare/v0.1.1...HEAD
[0.1.1]: https://github.com/s4mux/gitcheck/releases/tag/v0.1.1
[0.1.0]: https://github.com/s4mux/gitcheck/releases/tag/v0.1.0
