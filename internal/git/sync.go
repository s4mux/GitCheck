package git

import (
	"fmt"
	"strings"
)

func fillSyncStatus(status *RepoStatus, repoPath string, fetch bool) error {
	remote, err := runGit(repoPath, "remote")
	if err != nil || strings.TrimSpace(remote) == "" {
		status.NoRemote = true
		return nil
	}

	if fetch {
		if _, err := runGit(repoPath, "fetch", "--quiet"); err != nil {
			return fmt.Errorf("fetch %s: %w", repoPath, err)
		}
	}

	tracking, err := runGit(repoPath, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
	if err != nil {
		return nil // no upstream configured
	}
	tracking = strings.TrimSpace(tracking)

	revList, err := runGit(repoPath, "rev-list", "--left-right", "--count", "HEAD..."+tracking)
	if err != nil {
		return fmt.Errorf("rev-list %s: %w", repoPath, err)
	}

	fmt.Sscanf(strings.TrimSpace(revList), "%d\t%d", &status.AheadBy, &status.BehindBy)
	return nil
}
