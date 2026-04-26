package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/s4mux/gitcheck/internal/config"
)

// setupConfigEnv points XDG_CONFIG_HOME at a temp dir so DefaultWritePath
// resolves inside it, then returns the expected config path.
func setupConfigEnv(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	return filepath.Join(tmp, "gitcheck", "config.toml")
}

func runCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	rootCmd.SetArgs(nil)
	rootCmd.SetOut(nil)
	return buf.String(), err
}

func TestConfigList_empty(t *testing.T) {
	setupConfigEnv(t)
	out, err := runCmd(t, "config", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "(no roots configured)") {
		t.Fatalf("expected empty message, got: %s", out)
	}
}

func TestConfigAdd_creates(t *testing.T) {
	writePath := setupConfigEnv(t)
	out, err := runCmd(t, "config", "add", "/my/repos")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Added /my/repos") {
		t.Fatalf("unexpected output: %s", out)
	}
	cfg, err := config.Load(writePath)
	if err != nil {
		t.Fatalf("Load after add: %v", err)
	}
	if len(cfg.Scan.Roots) != 1 || cfg.Scan.Roots[0] != "/my/repos" {
		t.Fatalf("unexpected roots: %v", cfg.Scan.Roots)
	}
}

func TestConfigAdd_deduplicates(t *testing.T) {
	setupConfigEnv(t)
	if _, err := runCmd(t, "config", "add", "/my/repos"); err != nil {
		t.Fatalf("first add: %v", err)
	}
	out, err := runCmd(t, "config", "add", "/my/repos")
	if err != nil {
		t.Fatalf("second add: %v", err)
	}
	if !strings.Contains(out, "already in the list") {
		t.Fatalf("expected duplicate message, got: %s", out)
	}
}

func TestConfigList_showsRoots(t *testing.T) {
	setupConfigEnv(t)
	if _, err := runCmd(t, "config", "add", "/foo"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if _, err := runCmd(t, "config", "add", "/bar"); err != nil {
		t.Fatalf("add: %v", err)
	}
	out, err := runCmd(t, "config", "list")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(out, "/foo") || !strings.Contains(out, "/bar") {
		t.Fatalf("expected both roots in output, got: %s", out)
	}
}

func TestConfigRemove_found(t *testing.T) {
	writePath := setupConfigEnv(t)
	if _, err := runCmd(t, "config", "add", "/foo"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if _, err := runCmd(t, "config", "add", "/bar"); err != nil {
		t.Fatalf("add: %v", err)
	}
	out, err := runCmd(t, "config", "remove", "/foo")
	if err != nil {
		t.Fatalf("remove: %v", err)
	}
	if !strings.Contains(out, "Removed /foo") {
		t.Fatalf("unexpected output: %s", out)
	}
	cfg, err := config.Load(writePath)
	if err != nil {
		t.Fatalf("Load after remove: %v", err)
	}
	if len(cfg.Scan.Roots) != 1 || cfg.Scan.Roots[0] != "/bar" {
		t.Fatalf("unexpected roots: %v", cfg.Scan.Roots)
	}
}

func TestConfigRemove_notFound(t *testing.T) {
	setupConfigEnv(t)
	out, err := runCmd(t, "config", "remove", "/nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "not found in config") {
		t.Fatalf("expected not-found message, got: %s", out)
	}
}

func TestConfigRemove_lastRoot(t *testing.T) {
	setupConfigEnv(t)
	if _, err := runCmd(t, "config", "add", "/only"); err != nil {
		t.Fatalf("add: %v", err)
	}
	out, err := runCmd(t, "config", "remove", "/only")
	if err != nil {
		t.Fatalf("remove: %v", err)
	}
	if !strings.Contains(out, "Warning: no roots remain") {
		t.Fatalf("expected warning, got: %s", out)
	}
	if !strings.Contains(out, "Removed /only") {
		t.Fatalf("expected remove confirmation, got: %s", out)
	}
}
