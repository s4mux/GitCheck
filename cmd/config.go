package cmd

import (
	"errors"
	"fmt"

	"github.com/s4mux/gitcheck/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configAddCmd)
	configCmd.AddCommand(configRemoveCmd)
}

var configCmd = &cobra.Command{
	Use:          "config",
	Short:        "Manage gitcheck configuration",
	SilenceUsage: true,
}

var configListCmd = &cobra.Command{
	Use:          "list",
	Short:        "List configured scan roots",
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE:         runConfigList,
}

var configAddCmd = &cobra.Command{
	Use:          "add <path>",
	Short:        "Add a scan root path",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE:         runConfigAdd,
}

var configRemoveCmd = &cobra.Command{
	Use:          "remove <path>",
	Short:        "Remove a scan root path",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE:         runConfigRemove,
}

func runConfigList(cmd *cobra.Command, args []string) error {
	cfg, err := loadOrEmpty()
	if err != nil {
		return err
	}
	if len(cfg.Scan.Roots) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "(no roots configured)")
		return nil
	}
	for _, r := range cfg.Scan.Roots {
		fmt.Fprintln(cmd.OutOrStdout(), r)
	}
	return nil
}

func runConfigAdd(cmd *cobra.Command, args []string) error {
	path := args[0]
	cfg, err := loadOrEmpty()
	if err != nil {
		return err
	}
	for _, r := range cfg.Scan.Roots {
		if r == path {
			fmt.Fprintf(cmd.OutOrStdout(), "%s is already in the list\n", path)
			return nil
		}
	}
	cfg.Scan.Roots = append(cfg.Scan.Roots, path)
	writePath, err := config.DefaultWritePath()
	if err != nil {
		return err
	}
	if err := config.Save(writePath, cfg); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Added %s (config: %s)\n", path, writePath)
	return nil
}

func runConfigRemove(cmd *cobra.Command, args []string) error {
	path := args[0]
	cfg, err := loadOrEmpty()
	if err != nil {
		return err
	}
	var updated []string
	found := false
	for _, r := range cfg.Scan.Roots {
		if r == path {
			found = true
		} else {
			updated = append(updated, r)
		}
	}
	if !found {
		fmt.Fprintf(cmd.OutOrStdout(), "%s not found in config\n", path)
		return nil
	}
	cfg.Scan.Roots = updated
	if len(cfg.Scan.Roots) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "Warning: no roots remain — add one with 'gitcheck config add <path>'")
	}
	writePath, err := config.DefaultWritePath()
	if err != nil {
		return err
	}
	if err := config.Save(writePath, cfg); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Removed %s (config: %s)\n", path, writePath)
	return nil
}

// loadOrEmpty loads the current config, or returns an empty Config when no
// config file exists yet. Other errors are returned as-is.
func loadOrEmpty() (config.Config, error) {
	cfg, err := config.Load("")
	if err != nil {
		if errors.Is(err, config.ErrNoConfig) {
			return config.Config{}, nil
		}
		return config.Config{}, err
	}
	return cfg, nil
}
