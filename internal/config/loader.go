package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
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

func findConfigPath() (string, error) {
	var candidates []string

	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		candidates = append(candidates, filepath.Join(xdg, "gitcheck", "config.toml"))
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}

	candidates = append(candidates,
		filepath.Join(home, ".config", "gitcheck", "config.toml"),
		filepath.Join(home, ".gitcheck.toml"),
	)

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}

	return "", fmt.Errorf("no config file found; checked: %s", strings.Join(candidates, ", "))
}

func expandPaths(paths []string) []string {
	home, _ := os.UserHomeDir()
	result := make([]string, len(paths))
	for i, p := range paths {
		if strings.HasPrefix(p, "~/") {
			p = filepath.Join(home, p[2:])
		}
		result[i] = p
	}
	return result
}
