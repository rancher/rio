package ctr

import (
	"github.com/urfave/cli"
)

func NewCtrCommand() cli.Command {
	return cli.Command{
		Name:            "ctr",
		Usage:           "ctr backdoor",
		Hidden:          true,
		SkipFlagParsing: true,
		SkipArgReorder:  true,
		Action:          ctr,
	}
}
