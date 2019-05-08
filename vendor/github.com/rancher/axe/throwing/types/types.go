package types

import (
	"bytes"

	"github.com/rancher/axe/throwing/datafeeder"
	"github.com/rivo/tview"
)

var Kubeconfig string

type View struct {
	Actions []Action
	Kind    ResourceKind
	Feeder  datafeeder.DataSource
}

type ResourceView struct {
	tview.Primitive
	Title string
	Kind  string
	Index int
}

type ResourceKind struct {
	Title string
	Kind  string
}

type Refresher func(b *bytes.Buffer) error

type Drawer struct {
	RootPage  string
	ViewMap   map[string]View
	PageNav   map[rune]string
	Shortcuts [][]string
	Footers   []ResourceView
}

type Action struct {
	Name        string
	Description string
	Shortcut    rune
}
