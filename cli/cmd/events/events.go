package events

import (
	"sync"

	"github.com/rancher/rio/cli/pkg/monitor"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/cli/server"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"
)

var (
	writeLock sync.Mutex
)

type Events struct {
	Format string `desc:"'json' or 'yaml' or Custom format: '{{.ID}} {{.Stack.Name}}'" default:"jsoncompact"`
}

func (e *Events) Run(app *cli.Context) error {
	ctx, err := server.NewContext(app)
	if err != nil {
		return err
	}
	defer ctx.Close()

	sc, err := ctx.SpaceClient()
	if err != nil {
		return err
	}

	m := monitor.New(ctx.Client)
	sm := monitor.New(sc)

	eg := errgroup.Group{}
	eg.Go(func() error {
		s := ctx.Client.Types["subscribe"]
		return m.Start(&s)
	})
	eg.Go(func() error {
		s := sc.Types["subscribe"]
		return sm.Start(&s)
	})
	eg.Go(func() error {
		for c := range m.Subscribe().C {
			if err := e.printEvent(c, app); err != nil {
				return err
			}
		}
		return nil
	})
	eg.Go(func() error {
		for c := range sm.Subscribe().C {
			if err := e.printEvent(c, app); err != nil {
				return err
			}
		}
		return nil
	})

	return eg.Wait()
}

func (e *Events) printEvent(event *monitor.Event, app *cli.Context) error {
	writeLock.Lock()
	defer writeLock.Unlock()

	tw := table.NewWriter(nil, app)
	tw.Write(event)
	return tw.Close()
}
