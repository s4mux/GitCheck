package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/s4mux/gitcheck/internal/report"
)

const barWidth = 30

type Progress struct {
	mu    sync.Mutex
	w     io.Writer // nil = disabled (non-TTY or --json)
	lines int
}

func newProgress(w io.Writer, jsonOut bool) *Progress {
	if jsonOut || !report.IsTTY(w) {
		return &Progress{}
	}
	return &Progress{w: w}
}

func (p *Progress) StartScanning() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.w == nil {
		return
	}
	fmt.Fprintf(p.w, "Scanning for repositories...\n")
	p.lines = 1
}

func (p *Progress) FoundRepos(n int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.w == nil {
		return
	}
	if n == 0 {
		p.erase()
		return
	}
	p.redraw([]string{
		renderBar(0, n),
		"",
	})
}

func (p *Progress) RepoCompleted(done, total int, path string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.w == nil {
		return
	}
	p.redraw([]string{
		renderBar(done, total),
		shortenPath(path),
	})
}

func (p *Progress) Done() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.w == nil {
		return
	}
	p.erase()
}

// redraw replaces the currently displayed lines with new content.
// Must be called with p.mu held.
func (p *Progress) redraw(lines []string) {
	if p.lines > 0 {
		fmt.Fprintf(p.w, "\033[%dA", p.lines)
	}
	for _, line := range lines {
		fmt.Fprintf(p.w, "\r\033[2K%s\n", line)
	}
	p.lines = len(lines)
}

// erase removes all displayed progress lines.
// Must be called with p.mu held.
func (p *Progress) erase() {
	if p.lines == 0 {
		return
	}
	fmt.Fprintf(p.w, "\033[%dA\033[J", p.lines)
	p.lines = 0
}

func renderBar(done, total int) string {
	filled := 0
	if total > 0 {
		filled = done * barWidth / total
		if filled > barWidth {
			filled = barWidth
		}
	}
	bar := strings.Repeat("=", filled)
	if done < total && filled < barWidth {
		bar += ">"
		bar += strings.Repeat(" ", barWidth-filled-1)
	} else {
		bar += strings.Repeat(" ", barWidth-filled)
	}
	return fmt.Sprintf("[%s] %d/%d", bar, done, total)
}

func shortenPath(path string) string {
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
