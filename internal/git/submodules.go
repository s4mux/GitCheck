package git

import (
	"os"
	"path/filepath"
	"strings"
)

func GetSubmodules(repoPath string) ([]string, error) {
	if _, err := os.Stat(filepath.Join(repoPath, ".gitmodules")); err != nil {
		return nil, nil
	}

	out, err := runGit(repoPath, "config", "--file", ".gitmodules", "--get-regexp", `submodule\..*\.path`)
	if err != nil {
		return nil, nil
	}

	var paths []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		parts := strings.Fields(line)
		if len(parts) == 2 {
			paths = append(paths, filepath.Join(repoPath, parts[1]))
		}
	}
	return paths, nil
}
