package report

import (
	"bytes"
	"strings"
	"testing"

	"github.com/s4mux/gitcheck/internal/git"
)

func opts(verbose bool) Options {
	return Options{Verbose: verbose, UseColor: false}
}

func TestRender_allClean(t *testing.T) {
	repos := []git.Repo{{Path: "/a"}, {Path: "/b"}}
	var buf bytes.Buffer
	Render(&buf, repos, opts(false))
	if !strings.Contains(buf.String(), "All repositories clean.") {
		t.Fatalf("expected clean message, got: %q", buf.String())
	}
}

func TestRender_showsDirtyRepo(t *testing.T) {
	repos := []git.Repo{
		{Path: "/a", Status: git.RepoStatus{HasUncommitted: true}},
		{Path: "/b"},
	}
	var buf bytes.Buffer
	Render(&buf, repos, opts(false))
	out := buf.String()
	if !strings.Contains(out, "[uncommitted]") {
		t.Fatalf("expected uncommitted tag, got: %q", out)
	}
	if strings.Contains(out, "/b") {
		t.Fatalf("clean repo should not appear in default mode, got: %q", out)
	}
}

func TestRender_verbose(t *testing.T) {
	repos := []git.Repo{
		{Path: "/a", Status: git.RepoStatus{HasUncommitted: true}},
		{Path: "/b"},
	}
	var buf bytes.Buffer
	Render(&buf, repos, opts(true))
	out := buf.String()
	if !strings.Contains(out, "[ok]") {
		t.Fatalf("verbose mode should show [ok] for clean repos, got: %q", out)
	}
	if !strings.Contains(out, "[uncommitted]") {
		t.Fatalf("verbose mode should show dirty tags, got: %q", out)
	}
}

func TestRender_submoduleIndented(t *testing.T) {
	repos := []git.Repo{
		{
			Path: "/parent",
			Submodules: []git.Repo{
				{Path: "/parent/sub", Status: git.RepoStatus{HasUntracked: true}},
			},
		},
	}
	var buf bytes.Buffer
	Render(&buf, repos, opts(false))
	out := buf.String()
	if !strings.Contains(out, "└─") {
		t.Fatalf("expected submodule indentation, got: %q", out)
	}
	if !strings.Contains(out, "[untracked]") {
		t.Fatalf("expected untracked tag on submodule, got: %q", out)
	}
}

func TestRenderJSON(t *testing.T) {
	repos := []git.Repo{{Path: "/a", Status: git.RepoStatus{AheadBy: 2}}}
	var buf bytes.Buffer
	if err := RenderJSON(&buf, repos); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, `"ahead_by": 2`) {
		t.Fatalf("expected ahead_by in JSON, got: %q", out)
	}
}

func TestNeedsAttention(t *testing.T) {
	cases := []struct {
		name string
		repo git.Repo
		want bool
	}{
		{"clean", git.Repo{}, false},
		{"uncommitted", git.Repo{Status: git.RepoStatus{HasUncommitted: true}}, true},
		{"untracked", git.Repo{Status: git.RepoStatus{HasUntracked: true}}, true},
		{"ahead", git.Repo{Status: git.RepoStatus{AheadBy: 1}}, true},
		{"behind", git.Repo{Status: git.RepoStatus{BehindBy: 1}}, true},
		{"no remote", git.Repo{Status: git.RepoStatus{NoRemote: true}}, true},
		{"detached", git.Repo{Status: git.RepoStatus{DetachedHEAD: true}}, true},
		{"dirty submodule", git.Repo{Submodules: []git.Repo{{Status: git.RepoStatus{HasUncommitted: true}}}}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := needsAttention(tc.repo); got != tc.want {
				t.Fatalf("needsAttention = %v, want %v", got, tc.want)
			}
		})
	}
}
