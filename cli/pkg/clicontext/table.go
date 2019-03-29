package clicontext

import (
	"io"
	"os"
)

func (c *CLIContext) IDs() bool {
	return c.CLI.Bool("ids")
}

func (c *CLIContext) Quiet() bool {
	return c.CLI.Bool("quiet")
}

func (c *CLIContext) Format() string {
	return c.CLI.String("format")
}

func (c *CLIContext) Writer() io.Writer {
	return os.Stdout
}

func (c *CLIContext) GetDefaultStackName() string {
	return c.DefaultStackName
}
