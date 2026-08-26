// Package cli defines the tldg command tree (spec §7) using spf13/cobra.
package cli

import (
	"fmt"

	"github.com/leslierussell/tldg/internal/config"
	"github.com/spf13/cobra"
)

// tldg-5xh

// globalFlags holds parsed global flag values (spec §7.2).
type globalFlags struct {
	configPath string
	profile    string
	model      string
	offline    bool
	web        bool
	sources    string
	refresh    bool
	noIndex    bool
	jsonOut    bool
	markdown   bool
	quiet      bool
	verbose    bool
	yes        bool
	policy     string
}

var gf globalFlags

// loadConfig loads and applies profile override from global flags.
func loadConfig() (*config.Config, error) {
	cfg, err := config.Load(gf.configPath)
	if err != nil {
		return nil, err
	}
	if gf.profile != "" {
		cfg.Profile = gf.profile
	}
	return cfg, nil
}

// guardDeferred rejects flags for subsystems not yet implemented (spec §7.2).
func guardDeferred() error {
	switch {
	case gf.web:
		return fmt.Errorf("--web: external research arrives in milestone 5")
	case gf.sources != "":
		return fmt.Errorf("--sources: source selection arrives in milestone 5")
	case gf.policy != "":
		return fmt.Errorf("--policy: policy profiles arrive in milestone 5")
	case gf.refresh:
		return fmt.Errorf("--refresh: external caching arrives in milestone 5")
	}
	return nil
}

// NewRootCmd builds the root command with global flags and subcommands.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "tldg",
		Short:         "tldg — too long didn't git: local-first repository intelligence",
		Long:          "tldg explains what repositories actually do, answers questions from code and history, and cites its evidence. Local-first and offline-capable.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	pf := root.PersistentFlags()
	pf.StringVar(&gf.configPath, "config", "", "use a non-default configuration file")
	pf.StringVar(&gf.profile, "profile", "", "select a named configuration profile")
	pf.StringVar(&gf.model, "model", "", "override the active LLM configuration")
	pf.BoolVar(&gf.offline, "offline", false, "forbid all network activity")
	pf.BoolVar(&gf.web, "web", false, "permit configured external research (milestone 5)")
	pf.StringVar(&gf.sources, "sources", "", "limit enabled evidence sources (milestone 5)")
	pf.BoolVar(&gf.refresh, "refresh", false, "bypass caches and retrieve fresh data (milestone 5)")
	pf.BoolVar(&gf.noIndex, "no-index", false, "analyze without persisting an index")
	pf.BoolVar(&gf.jsonOut, "json", false, "emit stable machine-readable JSON")
	pf.BoolVar(&gf.markdown, "markdown", false, "emit Markdown rather than terminal text")
	pf.BoolVar(&gf.quiet, "quiet", false, "suppress progress output")
	pf.BoolVar(&gf.verbose, "verbose", false, "show retrieval and diagnostic output")
	pf.BoolVar(&gf.yes, "yes", false, "accept non-sensitive prompts automatically")
	pf.StringVar(&gf.policy, "policy", "", "select a privacy/execution policy profile (milestone 5)")

	root.AddCommand(
		newDoctorCmd(),
		newSummaryCmd(),
		newConfigCmd(),
		newVersionCmd(),
	)
	return root
}
