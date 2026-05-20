package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ScanResult struct {
	RepoPaths  []string
	NonGitDirs []string
}

type Scanner struct {
	Ignore IgnoreRules
}

func (s *Scanner) Scan(roots []string, gitOnly bool) (ScanResult, error) {
	var result ScanResult
	var errs []error

	for _, root := range roots {
		r, err := s.scanRoot(root, gitOnly)
		if err != nil {
			errs = append(errs, err)
		}
		result.RepoPaths = append(result.RepoPaths, r.RepoPaths...)
		result.NonGitDirs = append(result.NonGitDirs, r.NonGitDirs...)
	}

	if len(errs) > 0 {
		return result, fmt.Errorf("scan errors: %v", errs)
	}
	return result, nil
}

func (s *Scanner) scanRoot(root string, gitOnly bool) (ScanResult, error) {
	var result ScanResult

	// Pass 1: collect all git repos (existing behaviour).
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: cannot read %s: %v\n", path, err)
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		if path != root && s.Ignore.Match(path) {
			return filepath.SkipDir
		}
		if isGitRepo(path) {
			result.RepoPaths = append(result.RepoPaths, path)
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil || gitOnly {
		return result, err
	}

	// Pass 2: find highest-level directories that contain no git repos.
	// A directory qualifies if it is not a git repo, not the scan root, passes
	// ignore rules, and is not an ancestor of any repo found in pass 1.
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.IsDir() || path == root {
			return nil
		}
		if s.Ignore.Match(path) {
			return filepath.SkipDir
		}
		if isGitRepo(path) {
			return filepath.SkipDir
		}
		if isAncestorOfAnyRepo(path, result.RepoPaths) {
			return nil // descend to find repos inside
		}
		result.NonGitDirs = append(result.NonGitDirs, path)
		return filepath.SkipDir
	})

	return result, err
}

// isAncestorOfAnyRepo reports whether path is a proper ancestor directory of
// any repo in repoPaths (i.e. a repo lives somewhere inside path).
func isAncestorOfAnyRepo(path string, repoPaths []string) bool {
	prefix := path + string(filepath.Separator)
	for _, repo := range repoPaths {
		if strings.HasPrefix(repo, prefix) {
			return true
		}
	}
	return false
}

func isGitRepo(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil
}
