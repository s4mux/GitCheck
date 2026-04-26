package config

import "errors"

// ErrNoConfig is returned by Load when no config file exists at any candidate
// path and no explicit path was provided.
var ErrNoConfig = errors.New("no config file found")

// ErrEmptyRoots is returned by Load when a config file exists but scan.roots
// is empty. Callers can use this to trigger an interactive add-root flow.
var ErrEmptyRoots = errors.New("config: scan.roots must have at least one entry")

type Config struct {
	Scan ScanConfig `toml:"scan"`
}

type ScanConfig struct {
	Roots  []string     `toml:"roots"`
	Ignore IgnoreConfig `toml:"ignore"`
}

type IgnoreConfig struct {
	Paths    []string `toml:"paths"`
	Patterns []string `toml:"patterns"`
}
