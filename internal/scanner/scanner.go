package scanner

import (
	"fmt"
	"os"
	"path/filepath"
)

type Scanner struct {
	Ignore IgnoreRules
}

func (s *Scanner) Scan(roots []string) ([]string, error) {
	var repos []string
	var errs []error

	for _, root := range roots {
		found, err := s.scanRoot(root)
		if err != nil {
			errs = append(errs, err)
		}
		repos = append(repos, found...)
	}

	if len(errs) > 0 {
		return repos, fmt.Errorf("scan errors: %v", errs)
	}
	return repos, nil
}

func (s *Scanner) scanRoot(root string) ([]string, error) {
	var repos []string

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
			repos = append(repos, path)
			return filepath.SkipDir
		}
		return nil
	})

	return repos, err
}

func isGitRepo(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil
}
