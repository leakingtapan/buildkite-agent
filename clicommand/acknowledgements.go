package clicommand

import (
	"compress/gzip"
	"context"
	"embed"
	"fmt"
	"io"

	"github.com/urfave/cli"
)

const acknowledgementsHelpDescription = `Usage:

    buildkite-agent acknowledgements

Description:

Prints the licenses and notices of open source software incorporated into
this software.

Example:

    $ buildkite-agent acknowledgements`

//go:embed *.md.gz
var files embed.FS

type AcknowledgementsConfig struct{}

var AcknowledgementsCommand = cli.Command{
	Name:        "acknowledgements",
	Usage:       "Prints the licenses and notices of open source software incorporated into this software.",
	Description: acknowledgementsHelpDescription,
	Action: func(c *cli.Context) error {
		ctx := context.Background()
		_, _, _, _, done := setupLoggerAndConfig[AcknowledgementsConfig](ctx, c)
		defer done()

		// The main acknowledgements file should be generated by
		// scripts/generate-acknowledgements.sh.
		f, err := files.Open("ACKNOWLEDGEMENTS.md.gz")
		if err != nil {
			f, err = files.Open("dummy.md.gz")
			if err != nil {
				return fmt.Errorf("Couldn't open any embedded acknowledgements files: %w", err)
			}
		}
		r, err := gzip.NewReader(f)
		if err != nil {
			return fmt.Errorf("Couldn't create a gzip reader: %w", err)
		}
		if _, err := io.Copy(c.App.Writer, r); err != nil {
			return fmt.Errorf("Couldn't copy acknowledgments to output: %w", err)
		}
		return nil
	},
}
