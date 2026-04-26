package git

import (
	"fmt"
	"os/exec"
)


type Repo struct {
	Path       string   `json:"path"`
	Status     RepoStatus `json:"status"`
	Submodules []Repo   `json:"submodules,omitempty"`
}

type RepoStatus struct {
	HasUncommitted bool `json:"has_uncommitted"`
	HasUntracked   bool `json:"has_untracked"`
	AheadBy        int  `json:"ahead_by"`
	BehindBy       int  `json:"behind_by"`
	NoRemote       bool `json:"no_remote"`
	DetachedHEAD   bool `json:"detached_head"`
}

func BuildRepo(repoPath string, fetchRemote bool) (Repo, error) {
	status, err := GetStatus(repoPath, fetchRemote)
	if err != nil {
		return Repo{Path: repoPath}, err
	}

	subPaths, err := GetSubmodules(repoPath)
	if err != nil {
		return Repo{Path: repoPath, Status: status}, err
	}

	subs := make([]Repo, 0, len(subPaths))
	for _, sp := range subPaths {
		sub, subErr := BuildRepo(sp, fetchRemote)
		if subErr != nil {
			continue
		}
		sub.Status.DetachedHEAD = false
		subs = append(subs, sub)
	}

	return Repo{Path: repoPath, Status: status, Submodules: subs}, nil
}

func CheckAvailable() error {
	if err := exec.Command("git", "--version").Run(); err != nil {
		return fmt.Errorf("git not found in PATH")
	}
	return nil
}

func runGit(repoPath string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
