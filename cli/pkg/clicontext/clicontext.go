package clicontext

import (
	"context"

	"github.com/urfave/cli"
)

const dataKey = "config"

type CLIContext struct {
	*Config
	Ctx context.Context
	CLI *cli.Context
}

func Lookup(data map[string]interface{}) *CLIContext {
	return data[dataKey].(*CLIContext)
}

func (c *CLIContext) Store(data map[string]interface{}) {
	data[dataKey] = c
}

func Wrap(f func(*CLIContext) error) func(context2 *cli.Context) error {
	return func(app *cli.Context) error {
		cc := Lookup(app.App.Metadata)
		cc.CLI = app
		return f(cc)
	}
}

func DefaultAction(action interface{}) interface{} {
	if fn, ok := action.(func(ctx *CLIContext) error); ok {
		return Wrap(func(ctx *CLIContext) error {
			if ctx.CLI.Bool("help") {
				cli.ShowAppHelp(ctx.CLI)
				return nil
			}
			return fn(ctx)
		})
	}
	return action
}
