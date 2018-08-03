package util

import "github.com/urfave/cli"

func DefaultAction(obj interface{}) func(ctx *cli.Context) error {
	fn := obj.(func(*cli.Context) error)
	return func(ctx *cli.Context) error {
		if ctx.Bool("help") {
			cli.ShowAppHelp(ctx)
			return nil
		}
		return fn(ctx)
	}
}
