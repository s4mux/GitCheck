package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/adrg/xdg"
)

func Load(path string) (Config, error) {
	if path == "" {
		var err error
		path, err = findConfigPath()
		if err != nil {
			return Config{}, err
		}
	}

	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return Config{}, fmt.Errorf("load config %s: %w", path, err)
	}

	if len(cfg.Scan.Roots) == 0 {
		return Config{}, fmt.Errorf("config: scan.roots must have at least one entry")
	}

	cfg.Scan.Roots = expandPaths(cfg.Scan.Roots)
	cfg.Scan.Ignore.Paths = expandPaths(cfg.Scan.Ignore.Paths)

	return cfg, nil
}

// configHome returns the XDG config home directory, reading the environment
// at call time so that tests can override XDG_CONFIG_HOME via t.Setenv.
func configHome() string {
	if v := os.Getenv("XDG_CONFIG_HOME"); v != "" {
		return v
	}
	return xdg.ConfigHome
}

func findConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}

	candidates := []string{
		filepath.Join(configHome(), "gitcheck", "config.toml"),
		filepath.Join(home, ".gitcheck.toml"),
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}

	return "", fmt.Errorf("%w; checked: %s", ErrNoConfig, strings.Join(candidates, ", "))
}

// DefaultWritePath returns the canonical path where Save writes a new config
// file — always the XDG candidate, regardless of whether it exists.
func DefaultWritePath() (string, error) {
	return filepath.Join(configHome(), "gitcheck", "config.toml"), nil
}

// Save writes cfg to path as TOML, creating parent directories as needed.
// The write is atomic: a temp file is written first, then renamed into place.
func Save(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("save config: create directory: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".gitcheck-config-*.toml")
	if err != nil {
		return fmt.Errorf("save config: create temp file: %w", err)
	}
	tmpName := tmp.Name()
	encErr := toml.NewEncoder(tmp).Encode(cfg)
	closeErr := tmp.Close()
	if encErr != nil {
		os.Remove(tmpName)
		return fmt.Errorf("save config: encode: %w", encErr)
	}
	if closeErr != nil {
		os.Remove(tmpName)
		return fmt.Errorf("save config: close: %w", closeErr)
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("save config: rename: %w", err)
	}
	return nil
}

func expandPaths(paths []string) []string {
	home, _ := os.UserHomeDir()
	result := make([]string, len(paths))
	for i, p := range paths {
		if len(p) >= 2 && p[0] == '~' && (p[1] == '/' || p[1] == '\\') {
			p = filepath.Join(home, p[2:])
		}
		result[i] = p
	}
	return result
}
