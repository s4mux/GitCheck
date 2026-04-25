package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "Test")
	return dir
}

func commit(t *testing.T, dir, msg string) {
	t.Helper()
	cmd := exec.Command("git", "commit", "--allow-empty", "-m", msg)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("commit: %v\n%s", err, out)
	}
}

func TestGetStatus_clean(t *testing.T) {
	dir := initRepo(t)
	commit(t, dir, "init")

	s, err := GetStatus(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if s.HasUncommitted || s.HasUntracked || s.DetachedHEAD {
		t.Fatalf("expected clean status, got %+v", s)
	}
}

func TestGetStatus_uncommitted(t *testing.T) {
	dir := initRepo(t)
	commit(t, dir, "init")
	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hi"), 0644); err != nil {
		t.Fatal(err)
	}
	exec.Command("git", "-C", dir, "add", "file.txt").Run()

	s, err := GetStatus(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if !s.HasUncommitted {
		t.Fatalf("expected HasUncommitted, got %+v", s)
	}
}

func TestGetStatus_untracked(t *testing.T) {
	dir := initRepo(t)
	commit(t, dir, "init")
	if err := os.WriteFile(filepath.Join(dir, "new.txt"), []byte("hi"), 0644); err != nil {
		t.Fatal(err)
	}

	s, err := GetStatus(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if !s.HasUntracked {
		t.Fatalf("expected HasUntracked, got %+v", s)
	}
}

func TestGetStatus_noRemote(t *testing.T) {
	dir := initRepo(t)
	commit(t, dir, "init")

	s, err := GetStatus(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if !s.NoRemote {
		t.Fatalf("expected NoRemote, got %+v", s)
	}
}

func TestGetStatus_detachedHEAD(t *testing.T) {
	dir := initRepo(t)
	commit(t, dir, "init")

	// Get the commit hash and checkout detached
	out, err := exec.Command("git", "-C", dir, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	hash := string(out[:len(out)-1])
	if err := exec.Command("git", "-C", dir, "checkout", "--detach", hash).Run(); err != nil {
		t.Fatal(err)
	}

	s, err := GetStatus(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if !s.DetachedHEAD {
		t.Fatalf("expected DetachedHEAD, got %+v", s)
	}
}

func TestGetStatus_aheadBehind(t *testing.T) {
	// Create a "remote" repo
	remote := initRepo(t)
	commit(t, remote, "base")

	// Clone it
	localDir := t.TempDir()
	if out, err := exec.Command("git", "clone", remote, localDir).CombinedOutput(); err != nil {
		t.Fatalf("clone: %v\n%s", err, out)
	}
	exec.Command("git", "-C", localDir, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", localDir, "config", "user.name", "Test").Run()

	// Local commit (ahead by 1)
	commit(t, localDir, "local change")

	s, err := GetStatus(localDir, false)
	if err != nil {
		t.Fatal(err)
	}
	if s.AheadBy != 1 {
		t.Fatalf("expected AheadBy=1, got %+v", s)
	}
}

func TestGetSubmodules_none(t *testing.T) {
	dir := initRepo(t)
	paths, err := GetSubmodules(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 0 {
		t.Fatalf("expected no submodules, got %v", paths)
	}
}

func TestBuildRepo_basic(t *testing.T) {
	dir := initRepo(t)
	commit(t, dir, "init")

	repo, err := BuildRepo(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if repo.Path != dir {
		t.Fatalf("expected path %s, got %s", dir, repo.Path)
	}
}
