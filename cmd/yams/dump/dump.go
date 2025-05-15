package dump

import (
	"github.com/nsiow/yams/cmd/yams/cli"
	"github.com/nsiow/yams/internal/smartrw"
)

// Logic for the "dump" subcommand
func Run(opts *cli.Flags) {
	var output string
	var err error

	writer, err := smartrw.NewWriter(opts.Out)
	if err != nil {
		cli.Fail("invalid destination '%s': %v", opts.Out, err)
	}

	switch opts.Target {
	case "org":
		output, err = Org(opts)
	case "config":
		output, err = Config(opts)
	default:
		cli.Fail("unknown dump target '%s'", opts.Target)
	}

	if err != nil {
		cli.Fail("error attempting to dump '%s': %v", opts.Target, err)
	}

	_, err = writer.Write([]byte(output))
	if err != nil {
		cli.Fail("error writing to destination '%s': %v", opts.Out, err)
	}

	err = writer.Close()
	if err != nil {
		cli.Fail("error closing/flushing to destination '%s': %v", opts.Out, err)
	}
}
