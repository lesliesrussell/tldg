package cli

import (
	"fmt"
	"os"

	"github.com/lesliesrussell/tldg/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// tldg-5xh

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Inspect and manage configuration",
	}
	cmd.AddCommand(
		&cobra.Command{
			Use:   "path",
			Short: "Report active configuration and data paths",
			Args:  cobra.NoArgs,
			RunE: func(*cobra.Command, []string) error {
				cfgFile, err := config.DefaultConfigFile()
				if err != nil {
					return err
				}
				p, err := config.OSPaths()
				if err != nil {
					return err
				}
				fmt.Printf("config: %s\n", cfgFile)
				fmt.Printf("data:   %s\n", p.Data)
				fmt.Printf("cache:  %s\n", p.Cache)
				return nil
			},
		},
		&cobra.Command{
			Use:   "show",
			Short: "Print the effective configuration",
			Args:  cobra.NoArgs,
			RunE: func(*cobra.Command, []string) error {
				cfg, err := loadConfig()
				if err != nil {
					return coded(ExitBadArgs, err)
				}
				out, err := yaml.Marshal(cfg)
				if err != nil {
					return err
				}
				os.Stdout.Write(out)
				return nil
			},
		},
		&cobra.Command{
			Use:   "validate",
			Short: "Validate the configuration",
			Args:  cobra.NoArgs,
			RunE: func(*cobra.Command, []string) error {
				cfg, err := loadConfig()
				if err != nil {
					return coded(ExitBadArgs, err)
				}
				probs := config.Validate(cfg)
				if len(probs) == 0 {
					fmt.Println("configuration valid")
					return nil
				}
				for _, p := range probs {
					fmt.Fprintf(os.Stderr, "  %s\n", p.String())
				}
				return coded(ExitBadArgs, errSilent)
			},
		},
	)
	return cmd
}
