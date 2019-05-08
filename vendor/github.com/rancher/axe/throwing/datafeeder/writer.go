package datafeeder

import (
	"bytes"
	"strings"
)

type Row []string

type DataSource interface {
	Data() []Row
	Header() Row
	Refresh() error
}

type dataFeeder struct {
	rows       []Row
	rowLocator map[string]int
	header     Row
	refresher  func(buffer *bytes.Buffer) error
	buffer     *bytes.Buffer
}

func NewDataFeeder(r func(buffer *bytes.Buffer) error) *dataFeeder {
	return &dataFeeder{
		buffer:     new(bytes.Buffer),
		refresher:  r,
		rowLocator: map[string]int{},
	}
}

func (c *dataFeeder) Refresh() error {
	c.buffer.Reset()
	c.header = nil
	c.rows = nil
	return c.refresher(c.buffer)
}

func (c *dataFeeder) Data() []Row {
	content := c.buffer.String()
	body := strings.Split(content, "\n")[1:]
	for _, b := range body {
		c.rows = append(c.rows, Row(strings.Split(b, "\t")))
	}
	return c.rows
}

func (c *dataFeeder) Header() Row {
	content := c.buffer.String()
	header := strings.Split(content, "\n")[0]
	c.header = strings.Split(header, "\t")
	return c.header
}
