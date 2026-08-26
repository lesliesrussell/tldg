// Command tldg is a local-first, evidence-grounded repository intelligence CLI.
package main

import (
	"os"

	"github.com/leslierussell/tldg/internal/cli"
)

// tldg-5xh

func main() {
	os.Exit(cli.Execute())
}
