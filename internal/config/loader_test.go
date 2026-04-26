package config

import (
	"os"
	"path/filepath"
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
