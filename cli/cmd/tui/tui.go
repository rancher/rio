package tui

import (
	"github.com/rancher/axe/throwing"
	"github.com/rancher/rio/cli/pkg/clicontext"
)

type Tui struct {
}

func (t *Tui) Run(ctx clicontext.CLIContext) error {
	tui := throwing.NewAppView(ctx.K8s, drawer, tableEventHandler, signals)
	if err := app.Init(); err != nil {
		return err
	}
	return app.Run()
}
