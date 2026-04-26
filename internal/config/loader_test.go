package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeConfig(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoad_valid(t *testing.T) {
	dir := t.TempDir()
	path := writeConfig(t, dir, `
[scan]
roots = ["/tmp/projects"]
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Scan.Roots) != 1 || cfg.Scan.Roots[0] != "/tmp/projects" {
		t.Fatalf("unexpected roots: %v", cfg.Scan.Roots)
	}
}

func TestLoad_missingRoots(t *testing.T) {
	dir := t.TempDir()
	path := writeConfig(t, dir, `
[scan]
roots = []
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for empty roots")
	}
}

func TestLoad_tildeExpansion(t *testing.T) {
	dir := t.TempDir()
	path := writeConfig(t, dir, `
[scan]
roots = ["~/projects"]

[scan.ignore]
paths = ["~/projects/archived"]
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, "projects")
	if cfg.Scan.Roots[0] != want {
		t.Fatalf("expected %s, got %s", want, cfg.Scan.Roots[0])
	}
	wantIgnore := filepath.Join(home, "projects/archived")
	if cfg.Scan.Ignore.Paths[0] != wantIgnore {
		t.Fatalf("expected %s, got %s", wantIgnore, cfg.Scan.Ignore.Paths[0])
	}
}

func TestLoad_tildeBackslashExpansion(t *testing.T) {
	dir := t.TempDir()
	path := writeConfig(t, dir, `
[scan]
roots = ["~\\projects"]
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, "projects")
	if cfg.Scan.Roots[0] != want {
		t.Fatalf("expected %s, got %s", want, cfg.Scan.Roots[0])
	}
}

func TestLoad_invalidTOML(t *testing.T) {
	dir := t.TempDir()
	path := writeConfig(t, dir, `not valid toml [[[`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid TOML")
	}
}

func TestLoad_noConfigFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.toml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestErrNoConfig_sentinel(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp) // Windows
	_, err := Load("")
	if err == nil {
		t.Fatal("expected error when no config exists")
	}
	if !errors.Is(err, ErrNoConfig) {
		t.Fatalf("expected ErrNoConfig, got: %v", err)
	}
}

func TestSave_roundtrip(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "sub", "config.toml")
	want := Config{
		Scan: ScanConfig{
			Roots: []string{"/foo", "/bar"},
			Ignore: IgnoreConfig{
				Paths:    []string{"/foo/skip"},
				Patterns: []string{"*.bak"},
			},
		},
	}
	if err := Save(path, want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load after Save: %v", err)
	}
	if len(got.Scan.Roots) != 2 || got.Scan.Roots[0] != "/foo" || got.Scan.Roots[1] != "/bar" {
		t.Fatalf("unexpected roots: %v", got.Scan.Roots)
	}
	if len(got.Scan.Ignore.Paths) != 1 || got.Scan.Ignore.Paths[0] != "/foo/skip" {
		t.Fatalf("unexpected ignore paths: %v", got.Scan.Ignore.Paths)
	}
	if len(got.Scan.Ignore.Patterns) != 1 || got.Scan.Ignore.Patterns[0] != "*.bak" {
		t.Fatalf("unexpected ignore patterns: %v", got.Scan.Ignore.Patterns)
	}
}

func TestDefaultWritePath(t *testing.T) {
	path, err := DefaultWritePath()
	if err != nil {
		t.Fatalf("DefaultWritePath: %v", err)
	}
	if !filepath.IsAbs(path) {
		t.Fatalf("expected absolute path, got: %s", path)
	}
	if !strings.HasSuffix(path, filepath.Join("gitcheck", "config.toml")) {
		t.Fatalf("unexpected path suffix: %s", path)
	}
}
