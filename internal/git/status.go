package git

import (
	"fmt"
	"strings"
)

func GetStatus(repoPath string, fetchRemote bool) (RepoStatus, error) {
	var status RepoStatus

	_, err := runGit(repoPath, "symbolic-ref", "--quiet", "HEAD")
	status.DetachedHEAD = err != nil

	porcelain, err := runGit(repoPath, "status", "--porcelain")
	if err != nil {
		return status, fmt.Errorf("git status %s: %w", repoPath, err)
	}

	for _, line := range strings.Split(strings.TrimSpace(porcelain), "\n") {
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "??") {
			status.HasUntracked = true
		} else {
			status.HasUncommitted = true
		}
	}

	_ = fillSyncStatus(&status, repoPath, fetchRemote)

	return status, nil
}
