package cli

import (
	"fmt"

	"github.com/lesliesrussell/tldg/internal/version"
	"github.com/spf13/cobra"
)

// tldg-5xh

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the tldg version",
		Args:  cobra.NoArgs,
		RunE: func(*cobra.Command, []string) error {
			fmt.Printf("tldg %s\n", version.String())
			return nil
		},
	}
}
