package ctr

import (
	"github.com/urfave/cli"
)

func NewCtrCommand() cli.Command {
	return cli.Command{
		Name:            "ctr",
		Usage:           "Run ctr to troubleshoot containerd backend (required root)",
		Category:        "DEBUGGING",
		SkipFlagParsing: true,
		SkipArgReorder:  true,
		Action:          ctr,
	}
}
