package report

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/s4mux/gitcheck/internal/git"
)

type Options struct {
	Verbose  bool
	UseColor bool
}

func Render(w io.Writer, repos []git.Repo, opts Options) {
	if opts.Verbose {
		for _, r := range repos {
			renderRepo(w, r, opts, 0, "")
		}
		return
	}

	hasAny := false
	for _, r := range repos {
		if needsAttention(r) {
			hasAny = true
			renderRepo(w, r, opts, 0, "")
		}
	}
	if !hasAny {
		fmt.Fprintln(w, "All repositories clean.")
	}
}

func RenderJSON(w io.Writer, repos []git.Repo) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(repos)
}

func renderRepo(w io.Writer, r git.Repo, opts Options, depth int, parentPath string) {
	var dp string
	if depth == 0 {
		dp = homeShortenPath(r.Path)
	} else {
		rel, err := filepath.Rel(parentPath, r.Path)
		if err != nil || strings.HasPrefix(rel, "..") {
			dp = r.Path
		} else {
			dp = rel
		}
	}

	prefix := ""
	if depth > 0 {
		prefix = strings.Repeat("  ", depth) + "└─ "
	}

	tags := formatTags(r.Status, opts.UseColor)
	if opts.Verbose && tags == "" {
		tags = colorize("[ok]", colorGreen, opts.UseColor)
	}

	if tags != "" {
		fmt.Fprintf(w, "%s%-40s %s\n", prefix, dp, tags)
	} else {
		fmt.Fprintf(w, "%s%s\n", prefix, dp)
	}

	for _, sub := range r.Submodules {
		if opts.Verbose || needsAttention(sub) {
			renderRepo(w, sub, opts, depth+1, r.Path)
		}
	}
}

func formatTags(s git.RepoStatus, useColor bool) string {
	var tags []string
	if s.HasUncommitted {
		tags = append(tags, colorize("[uncommitted]", colorYellow, useColor))
	}
	if s.HasUntracked {
		tags = append(tags, colorize("[untracked]", colorYellow, useColor))
	}
	if s.AheadBy > 0 {
		tags = append(tags, colorize(fmt.Sprintf("[ahead %d]", s.AheadBy), colorYellow, useColor))
	}
	if s.BehindBy > 0 {
		tags = append(tags, colorize(fmt.Sprintf("[behind %d]", s.BehindBy), colorRed, useColor))
	}
	if s.NoRemote {
		tags = append(tags, colorize("[no remote]", colorYellow, useColor))
	}
	if s.DetachedHEAD {
		tags = append(tags, colorize("[detached HEAD]", colorRed, useColor))
	}
	return strings.Join(tags, " ")
}

func needsAttention(r git.Repo) bool {
	s := r.Status
	if s.HasUncommitted || s.HasUntracked || s.AheadBy > 0 || s.BehindBy > 0 || s.NoRemote || s.DetachedHEAD {
		return true
	}
	for _, sub := range r.Submodules {
		if needsAttention(sub) {
			return true
		}
	}
	return false
}

func homeShortenPath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	rel, err := filepath.Rel(home, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return path
	}
	return "~/" + rel
}
