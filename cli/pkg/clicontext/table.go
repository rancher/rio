package clicontext

import (
	"io"
)

func (c *CLIContext) IDs() bool {
	return c.CLI.Bool("ids")
}

func (c *CLIContext) AllNamespaceSet() bool {
	return c.AllNamespace
}

func (c *CLIContext) Quiet() bool {
	return c.CLI.Bool("quiet")
}

func (c *CLIContext) Format() string {
	return c.CLI.String("format")
}

func (c *CLIContext) Writer() io.Writer {
	return c.Config.Writer
}

func (c *CLIContext) WithWriter(writer io.Writer) {
	c.Config.Writer = writer
}

func (c *CLIContext) GetSetNamespace() string {
	if c.CLI.GlobalBool("all-namespaces") {
		return ""
	}
	ns := c.CLI.GlobalString("namespace")
	if ns == "" {
		return c.GetDefaultNamespace()
	}
	return ns
}

func (c *CLIContext) GetDefaultNamespace() string {
	return c.DefaultNamespace
}

func (c *CLIContext) GetSystemNamespace() string {
	return c.SystemNamespace
}
