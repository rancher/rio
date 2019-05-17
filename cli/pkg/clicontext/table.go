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

func (c *CLIContext) GetSetNamespace() string {
	return c.CLI.GlobalString("namespace")
}

func (c *CLIContext) GetDefaultNamespace() string {
	return "default"
}

func (c *CLIContext) GetSystemNamespace() string {
	return c.SystemNamespace
}
