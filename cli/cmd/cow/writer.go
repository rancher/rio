package cow

import (
	"bytes"
	"strings"
)

type cmdDataFeeder struct {
	rows      []Row
	header    Row
	refresher func(buffer *bytes.Buffer) error
	buffer    *bytes.Buffer
}

func (c *cmdDataFeeder) Refresh() error {
	c.buffer.Reset()
	return c.refresher(c.buffer)
}

func (c *cmdDataFeeder) Data() []Row {
	content := c.buffer.String()
	body := strings.Split(content, "\n")[1:]
	for _, b := range body {
		c.rows = append(c.rows, Row{b})
	}
	return c.rows
}

func (c *cmdDataFeeder) Header() Row {
	content := c.buffer.String()
	header := strings.Split(content, "\n")[0]
	c.header = strings.Split(header, "\t")
	return c.header
}
