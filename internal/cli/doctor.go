package cli

import (
	"context"
	"os"

	"github.com/leslierussell/tldg/internal/app"
	"github.com/spf13/cobra"
)

// tldg-5xh

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose git, model, index, and configuration",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return coded(ExitBadArgs, err)
			}
			report := app.RunDoctor(context.Background(), cfg, app.DoctorOptions{Offline: gf.offline})

			if gf.jsonOut {
				if err := app.RenderDoctorJSON(os.Stdout, report); err != nil {
					return err
				}
			} else {
				if err := app.RenderDoctorText(os.Stdout, report); err != nil {
					return err
				}
			}
			if report.HasFailure() {
				return coded(ExitGeneralFailure, errSilent)
			}
			return nil
		},
	}
}
