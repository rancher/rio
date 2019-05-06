package table

import "github.com/urfave/cli"

var stackLsFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "quiet,q",
		Usage: "Only display IDs",
	},
	cli.StringFlag{
		Name:  "format",
		Usage: "'json' or 'yaml' or Custom format: '{{.ID}} {{.Stack.ServiceName}}'",
	},
}

func WriterFlags() []cli.Flag {
	return stackLsFlags
}
