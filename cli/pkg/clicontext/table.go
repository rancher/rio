package clicontext

import (
	"io"
)

func (c *CLIContext) IDs() bool {
	return c.CLI.Bool("ids")
}

func (c *CLIContext) AllNamespaceSet() bool {
	return c.CLI.GlobalBool("--all-namespaces")
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
	if c.CLI.GlobalBool("A") {
		return ""
	}
	return c.CLI.GlobalString("namespace")
}

func (c *CLIContext) GetDefaultNamespace() string {
	return "default"
}

func (c *CLIContext) GetSystemNamespace() string {
	return c.SystemNamespace
}
