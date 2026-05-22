package scanner

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func mkRepo(t *testing.T, base, name string) string {
	t.Helper()
	dir := filepath.Join(base, name)
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func mkDir(t *testing.T, base, name string) string {
	t.Helper()
	dir := filepath.Join(base, name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func newScanner(paths, patterns []string) *Scanner {
	return &Scanner{Ignore: NewIgnoreRules(paths, patterns)}
}

func TestScan_findsDirectRepo(t *testing.T) {
	root := t.TempDir()
	repo := mkRepo(t, root, "myapp")

	s := newScanner(nil, nil)
	got, err := s.Scan([]string{root}, false)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(got.RepoPaths, repo) {
		t.Fatalf("expected %s in results %v", repo, got.RepoPaths)
	}
}

func TestScan_findsNestedRepo(t *testing.T) {
	root := t.TempDir()
	mkDir(t, root, "work/team")
	repo := mkRepo(t, filepath.Join(root, "work/team"), "api")

	s := newScanner(nil, nil)
	got, err := s.Scan([]string{root}, false)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(got.RepoPaths, repo) {
		t.Fatalf("expected %s in results %v", repo, got.RepoPaths)
	}
}

func TestScan_stopsAtRepoBoundary(t *testing.T) {
	root := t.TempDir()
	outer := mkRepo(t, root, "outer")
	inner := mkRepo(t, outer, "inner") // inside a repo — should not be found by scanner
	_ = inner

	s := newScanner(nil, nil)
	got, err := s.Scan([]string{root}, false)
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(got.RepoPaths, inner) {
		t.Fatalf("inner repo should not be found when scanner stops at outer: %v", got.RepoPaths)
	}
	if !slices.Contains(got.RepoPaths, outer) {
		t.Fatalf("outer repo should be found: %v", got.RepoPaths)
	}
}

func TestScan_ignoreExactPath(t *testing.T) {
	root := t.TempDir()
	keep := mkRepo(t, root, "keep")
	skip := mkRepo(t, root, "skip")

	s := newScanner([]string{skip}, nil)
	got, err := s.Scan([]string{root}, false)
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(got.RepoPaths, skip) {
		t.Fatalf("skipped repo should not be found: %v", got.RepoPaths)
	}
	if !slices.Contains(got.RepoPaths, keep) {
		t.Fatalf("kept repo should be found: %v", got.RepoPaths)
	}
}

func TestScan_ignorePattern(t *testing.T) {
	root := t.TempDir()
	keep := mkRepo(t, root, "myapp")
	skip := mkRepo(t, root, "temp-work")

	s := newScanner(nil, []string{"temp-*"})
	got, err := s.Scan([]string{root}, false)
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(got.RepoPaths, skip) {
		t.Fatalf("pattern-matched repo should not be found: %v", got.RepoPaths)
	}
	if !slices.Contains(got.RepoPaths, keep) {
		t.Fatalf("kept repo should be found: %v", got.RepoPaths)
	}
}

func TestScan_multipleRoots(t *testing.T) {
	r1 := t.TempDir()
	r2 := t.TempDir()
	repo1 := mkRepo(t, r1, "a")
	repo2 := mkRepo(t, r2, "b")

	s := newScanner(nil, nil)
	got, err := s.Scan([]string{r1, r2}, false)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(got.RepoPaths, repo1) || !slices.Contains(got.RepoPaths, repo2) {
		t.Fatalf("expected both repos in results: %v", got.RepoPaths)
	}
}

func TestScan_nonGitDir_found(t *testing.T) {
	root := t.TempDir()
	plain := mkDir(t, root, "projectB")

	s := newScanner(nil, nil)
	got, err := s.Scan([]string{root}, false)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(got.NonGitDirs, plain) {
		t.Fatalf("expected non-git dir %s in NonGitDirs %v", plain, got.NonGitDirs)
	}
	if len(got.RepoPaths) != 0 {
		t.Fatalf("expected no repos, got: %v", got.RepoPaths)
	}
}

func TestScan_nonGitDir_highestLevelOnly(t *testing.T) {
	root := t.TempDir()
	top := mkDir(t, root, "work")
	mkDir(t, root, "work/team") // subdirectory of top — should not appear separately

	s := newScanner(nil, nil)
	got, err := s.Scan([]string{root}, false)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(got.NonGitDirs, top) {
		t.Fatalf("expected top-level non-git dir %s in NonGitDirs %v", top, got.NonGitDirs)
	}
	sub := filepath.Join(root, "work/team")
	if slices.Contains(got.NonGitDirs, sub) {
		t.Fatalf("subdirectory %s should not appear separately in NonGitDirs %v", sub, got.NonGitDirs)
	}
}

func TestScan_nonGitDir_notRoot(t *testing.T) {
	root := t.TempDir() // root itself is not a git repo

	s := newScanner(nil, nil)
	got, err := s.Scan([]string{root}, false)
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(got.NonGitDirs, root) {
		t.Fatalf("scan root itself should not appear in NonGitDirs: %v", got.NonGitDirs)
	}
}

func TestScan_rootIsRepo_noGitMetaInNonGitDirs(t *testing.T) {
	root := mkRepo(t, t.TempDir(), "myrepo")

	s := newScanner(nil, nil)
	got, err := s.Scan([]string{root}, false)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(got.RepoPaths, root) {
		t.Fatalf("expected root repo %s in RepoPaths %v", root, got.RepoPaths)
	}
	if len(got.NonGitDirs) != 0 {
		t.Fatalf("expected no NonGitDirs when root is a repo, got: %v", got.NonGitDirs)
	}
}

func TestScan_gitOnly_excludesNonGit(t *testing.T) {
	root := t.TempDir()
	mkDir(t, root, "projectB")
	repo := mkRepo(t, root, "projectA")

	s := newScanner(nil, nil)
	got, err := s.Scan([]string{root}, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.NonGitDirs) != 0 {
		t.Fatalf("--git-only should yield no NonGitDirs, got: %v", got.NonGitDirs)
	}
	if !slices.Contains(got.RepoPaths, repo) {
		t.Fatalf("expected repo %s in RepoPaths %v", repo, got.RepoPaths)
	}
}
