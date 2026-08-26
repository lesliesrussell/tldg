package cli

import (
	"context"
	"os"

	"github.com/leslierussell/tldg/internal/app"
	"github.com/leslierussell/tldg/internal/render"
	"github.com/spf13/cobra"
)

// tldg-5xh

func newSummaryCmd() *cobra.Command {
	var depth string
	cmd := &cobra.Command{
		Use:   "summary [target]",
		Short: "Produce an evidence-backed repository overview",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := guardDeferred(); err != nil {
				return coded(ExitBadArgs, err)
			}
			cfg, err := loadConfig()
			if err != nil {
				return coded(ExitBadArgs, err)
			}
			tgt := "."
			if len(args) == 1 {
				tgt = args[0]
			}
			res, err := app.RunSummary(context.Background(), cfg, app.SummaryOptions{
				Target:        tgt,
				Offline:       gf.offline,
				NoIndex:       gf.noIndex,
				Depth:         depth,
				ModelOverride: gf.model,
			})
			if err != nil {
				return classifyErr(err)
			}
			return output(res)
		},
	}
	cmd.Flags().StringVar(&depth, "depth", "standard", "brief|standard|architecture|exhaustive")
	cmd.Flags().String("branch", "", "analyze a specific branch")
	cmd.Flags().String("ref", "", "analyze a specific commit or tag")
	cmd.Flags().String("include", "", "areas to include, e.g. docs,code,history")
	cmd.Flags().String("exclude", "", "areas to exclude")
	return cmd
}

// output renders a Result honoring --json/--markdown/text (spec §8).
func output(res *render.Result) error {
	switch {
	case gf.jsonOut:
		return render.JSON(os.Stdout, res)
	case gf.markdown:
		return render.Markdown(os.Stdout, res)
	default:
		return render.Text(os.Stdout, res)
	}
}
