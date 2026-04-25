package cmd

import (
	"fmt"
	"os"
	"runtime"
	"sync"

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
)

var rootCmd = &cobra.Command{
	Use:          "gitcheck",
	Short:        "Report status of local Git repositories",
	Version:      "0.1.0",
	SilenceUsage: true,
	RunE:         run,
}

func Execute() {
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
}

func run(cmd *cobra.Command, args []string) error {
	if err := git.CheckAvailable(); err != nil {
		return err
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	s := &scanner.Scanner{
		Ignore: scanner.NewIgnoreRules(cfg.Scan.Ignore.Paths, cfg.Scan.Ignore.Patterns),
	}

	paths, err := s.Scan(cfg.Scan.Roots)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: %v\n", err)
	}

	repos := resolveRepos(paths, fetch)

	if jsonOut {
		return report.RenderJSON(os.Stdout, repos)
	}

	report.Render(os.Stdout, repos, report.Options{
		Verbose:  verbose,
		UseColor: !noColor && report.IsTTY(os.Stdout),
	})
	return nil
}

func resolveRepos(paths []string, fetchRemote bool) []git.Repo {
	results := make([]git.Repo, len(paths))
	var wg sync.WaitGroup
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
		}(i, path)
	}
	wg.Wait()
	return results
}
