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
	got, err := s.Scan([]string{root})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(got, repo) {
		t.Fatalf("expected %s in results %v", repo, got)
	}
}

func TestScan_findsNestedRepo(t *testing.T) {
	root := t.TempDir()
	mkDir(t, root, "work/team")
	repo := mkRepo(t, filepath.Join(root, "work/team"), "api")

	s := newScanner(nil, nil)
	got, err := s.Scan([]string{root})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(got, repo) {
		t.Fatalf("expected %s in results %v", repo, got)
	}
}

func TestScan_stopsAtRepoBoundary(t *testing.T) {
	root := t.TempDir()
	outer := mkRepo(t, root, "outer")
	inner := mkRepo(t, outer, "inner") // inside a repo — should not be found by scanner
	_ = inner

	s := newScanner(nil, nil)
	got, err := s.Scan([]string{root})
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(got, inner) {
		t.Fatalf("inner repo should not be found when scanner stops at outer: %v", got)
	}
	if !slices.Contains(got, outer) {
		t.Fatalf("outer repo should be found: %v", got)
	}
}

func TestScan_ignoreExactPath(t *testing.T) {
	root := t.TempDir()
	keep := mkRepo(t, root, "keep")
	skip := mkRepo(t, root, "skip")

	s := newScanner([]string{skip}, nil)
	got, err := s.Scan([]string{root})
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(got, skip) {
		t.Fatalf("skipped repo should not be found: %v", got)
	}
	if !slices.Contains(got, keep) {
		t.Fatalf("kept repo should be found: %v", got)
	}
}

func TestScan_ignorePattern(t *testing.T) {
	root := t.TempDir()
	keep := mkRepo(t, root, "myapp")
	skip := mkRepo(t, root, "temp-work")

	s := newScanner(nil, []string{"temp-*"})
	got, err := s.Scan([]string{root})
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(got, skip) {
		t.Fatalf("pattern-matched repo should not be found: %v", got)
	}
	if !slices.Contains(got, keep) {
		t.Fatalf("kept repo should be found: %v", got)
	}
}

func TestScan_multipleRoots(t *testing.T) {
	r1 := t.TempDir()
	r2 := t.TempDir()
	repo1 := mkRepo(t, r1, "a")
	repo2 := mkRepo(t, r2, "b")

	s := newScanner(nil, nil)
	got, err := s.Scan([]string{r1, r2})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(got, repo1) || !slices.Contains(got, repo2) {
		t.Fatalf("expected both repos in results: %v", got)
	}
}
