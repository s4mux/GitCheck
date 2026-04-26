package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/s4mux/gitcheck/internal/config"
)

func TestDoFirstRunSetup(t *testing.T) {
	tmp := t.TempDir()
	writePath := filepath.Join(tmp, "config.toml")

	input := strings.NewReader("~/projects\n")
	var out bytes.Buffer

	cfg, err := doFirstRunSetup(input, &out, writePath)
	if err != nil {
		t.Fatalf("doFirstRunSetup: %v", err)
	}
	if len(cfg.Scan.Roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(cfg.Scan.Roots))
	}
	if !strings.Contains(out.String(), "Config saved to") {
		t.Fatalf("expected confirmation message, got: %s", out.String())
	}
}

func TestDoFirstRunSetup_emptyInput(t *testing.T) {
	tmp := t.TempDir()
	writePath := filepath.Join(tmp, "config.toml")

	input := strings.NewReader("\n")
	var out bytes.Buffer

	_, err := doFirstRunSetup(input, &out, writePath)
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestDoFirstRunSetup_eofNoInput(t *testing.T) {
	tmp := t.TempDir()
	writePath := filepath.Join(tmp, "config.toml")

	input := strings.NewReader("")
	var out bytes.Buffer

	_, err := doFirstRunSetup(input, &out, writePath)
	if err == nil {
		t.Fatal("expected error for EOF with no path")
	}
}

func TestDoFirstRunSetup_writesConfig(t *testing.T) {
	tmp := t.TempDir()
	writePath := filepath.Join(tmp, "config.toml")

	input := strings.NewReader("/my/repos\n")
	var out bytes.Buffer

	_, err := doFirstRunSetup(input, &out, writePath)
	if err != nil {
		t.Fatalf("doFirstRunSetup: %v", err)
	}

	// verify the file was written with the correct content
	cfg, err := config.Load(writePath)
	if err != nil {
		t.Fatalf("Load after setup: %v", err)
	}
	if len(cfg.Scan.Roots) != 1 || cfg.Scan.Roots[0] != "/my/repos" {
		t.Fatalf("unexpected roots: %v", cfg.Scan.Roots)
	}
}
