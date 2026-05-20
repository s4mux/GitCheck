package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/s4mux/gitcheck/internal/config"
	"github.com/s4mux/gitcheck/internal/git"
	"github.com/s4mux/gitcheck/internal/report"
	"github.com/s4mux/gitcheck/internal/scanner"
	"github.com/spf13/cobra"
)

var (
	configPath string
	verbose    bool
	fetch      bool
	noColor    bool
	jsonOut    bool
	gitOnly    bool
)

var rootCmd = &cobra.Command{
	Use:          "gitcheck",
	Short:        "Report status of local Git repositories",
	SilenceUsage: true,
	RunE:         run,
}

func Execute(version string) {
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVar(&configPath, "config", "", "path to config file")
	rootCmd.Flags().BoolVar(&verbose, "verbose", false, "show all repos, not only those needing attention")
	rootCmd.Flags().BoolVar(&fetch, "fetch", false, "fetch from remote before checking ahead/behind state")
	rootCmd.Flags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.Flags().BoolVar(&jsonOut, "json", false, "output results as JSON")
	rootCmd.Flags().BoolVar(&gitOnly, "git-only", false, "only show git repositories, skip non-git directories")
}

func run(cmd *cobra.Command, args []string) error {
	if err := git.CheckAvailable(); err != nil {
		return err
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		var prompt string
		switch {
		case errors.Is(err, config.ErrNoConfig) && configPath == "":
			prompt = "No config file found. Which folder should gitcheck scan?"
		case errors.Is(err, config.ErrEmptyRoots) && configPath == "":
			prompt = "No scan roots configured. Which folder should gitcheck scan?"
		default:
			return err
		}
		cfg, err = runFirstRunSetup(cmd, prompt)
		if err != nil {
			return err
		}
	}

	if !jsonOut {
		if resolvedPath, err := config.ResolvePath(configPath); err == nil {
			fmt.Fprintf(os.Stderr, "config: %s\n", resolvedPath)
		}
	}

	s := &scanner.Scanner{
		Ignore: scanner.NewIgnoreRules(cfg.Scan.Ignore.Paths, cfg.Scan.Ignore.Patterns),
	}

	prog := newProgress(os.Stdout, jsonOut)
	prog.StartScanning()

	result, err := s.Scan(cfg.Scan.Roots, gitOnly)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: %v\n", err)
	}

	prog.FoundRepos(len(result.RepoPaths))
	repos := resolveRepos(result.RepoPaths, fetch, prog)
	prog.Done()

	if jsonOut {
		return report.RenderJSON(os.Stdout, repos, result.NonGitDirs)
	}

	report.Render(os.Stdout, repos, report.Options{
		Verbose:    verbose,
		UseColor:   !noColor && report.IsTTY(os.Stdout),
		NonGitDirs: result.NonGitDirs,
	})
	return nil
}

func runFirstRunSetup(cmd *cobra.Command, prompt string) (config.Config, error) {
	if !isStdinTTY() {
		return config.Config{}, fmt.Errorf(
			"no scan roots configured and stdin is not a terminal; " +
				"create a config file or run gitcheck interactively")
	}
	writePath, err := config.DefaultWritePath()
	if err != nil {
		return config.Config{}, err
	}
	return doFirstRunSetup(cmd.InOrStdin(), cmd.OutOrStdout(), writePath, prompt)
}

func doFirstRunSetup(r io.Reader, w io.Writer, writePath string, prompt string) (config.Config, error) {
	fmt.Fprintln(w, prompt)
	fmt.Fprint(w, "> ")
	reader := bufio.NewReader(r)
	line, err := reader.ReadString('\n')
	if err != nil && (err != io.EOF || strings.TrimSpace(line) == "") {
		return config.Config{}, fmt.Errorf("first-run setup: reading input: %w", err)
	}
	root := strings.TrimSpace(line)
	if root == "" {
		return config.Config{}, fmt.Errorf("first-run setup: no path provided")
	}
	cfg := config.Config{Scan: config.ScanConfig{Roots: []string{root}}}
	if err := config.Save(writePath, cfg); err != nil {
		return config.Config{}, fmt.Errorf("first-run setup: %w", err)
	}
	fmt.Fprintf(w, "Config saved to %s\n", writePath)
	return config.Load(writePath)
}

func isStdinTTY() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func resolveRepos(paths []string, fetchRemote bool, prog *Progress) []git.Repo {
	results := make([]git.Repo, len(paths))
	var wg sync.WaitGroup
	var completed int32
	sem := make(chan struct{}, runtime.NumCPU())

	for i, path := range paths {
		wg.Add(1)
		go func(i int, path string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			repo, err := git.BuildRepo(path, fetchRemote)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: %s: %v\n", path, err)
			}
			results[i] = repo
			n := int(atomic.AddInt32(&completed, 1))
			prog.RepoCompleted(n, len(paths), path)
		}(i, path)
	}
	wg.Wait()
	return results
}
