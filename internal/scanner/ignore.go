package scanner

import "path/filepath"

type IgnoreRules struct {
	paths    map[string]bool
	patterns []string
}

func NewIgnoreRules(paths, patterns []string) IgnoreRules {
	set := make(map[string]bool, len(paths))
	for _, p := range paths {
		set[p] = true
	}
	return IgnoreRules{paths: set, patterns: patterns}
}

func (r IgnoreRules) Match(path string) bool {
	if r.paths[path] {
		return true
	}
	name := filepath.Base(path)
	for _, pattern := range r.patterns {
		if matched, err := filepath.Match(pattern, name); err == nil && matched {
			return true
		}
	}
	return false
}
