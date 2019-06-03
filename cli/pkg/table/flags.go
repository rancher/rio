package table

import "github.com/urfave/cli"

var stackLsFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "quiet,q",
		Usage: "Only display Names",
	},
	cli.StringFlag{
		Name:   "format",
		EnvVar: "FORMAT",
		Usage:  "'json' or 'yaml' or Custom format: '{{.Name}} {{.Obj.Name}}'",
	},
}

func WriterFlags() []cli.Flag {
	return stackLsFlags
}
