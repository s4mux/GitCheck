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

func findConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}

	candidates := []string{
		filepath.Join(xdg.ConfigHome, "gitcheck", "config.toml"),
		filepath.Join(home, ".gitcheck.toml"),
	}

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
		if len(p) >= 2 && p[0] == '~' && (p[1] == '/' || p[1] == '\\') {
			p = filepath.Join(home, p[2:])
		}
		result[i] = p
	}
	return result
}
