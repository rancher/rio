package events

import (
	"sync"

	"github.com/rancher/rio/cli/pkg/clicontext"
)

var (
	writeLock sync.Mutex
)

type Events struct {
	Format string `desc:"'json' or 'yaml' or Custom format: '{{.ID}} {{.Stack.Name}}'" default:"jsoncompact"`
}

func (e *Events) Run(ctx *clicontext.CLIContext) error {
	// todo
	return nil
}

//func (e *Events) Run(ctx *clicontext.CLIContext) error {
//	sc, err := ctx.ClusterClient()
//	if err != nil {
//		return err
//	}
//
//	wc, err := ctx.ProjectClient()
//	if err != nil {
//		return err
//	}
//
//	m := monitor.New(wc)
//	sm := monitor.New(sc)
//
//	parentCtx, cancel := context.WithCancel(context.Background())
//	eg, childCtx := errgroup.WithContext(parentCtx)
//	eg.Go(func() error {
//		defer cancel()
//		s := wc.Types["subscribe"]
//		return m.Start(childCtx, &s)
//	})
//	eg.Go(func() error {
//		defer cancel()
//		s := sc.Types["subscribe"]
//		return sm.Start(childCtx, &s)
//	})
//	eg.Go(func() error {
//		defer cancel()
//		for c := range m.Subscribe().C {
//			if err := e.printEvent(c, ctx); err != nil {
//				return err
//			}
//		}
//		return nil
//	})
//	eg.Go(func() error {
//		defer cancel()
//		for c := range sm.Subscribe().C {
//			if err := e.printEvent(c, ctx); err != nil {
//				return err
//			}
//		}
//		return nil
//	})
//
//	return eg.Wait()
//}
//
//func (e *Events) printEvent(event *monitor.Event, ctx *clicontext.CLIContext) error {
//	writeLock.Lock()
//	defer writeLock.Unlock()
//
//	tw := table.NewWriter(nil, ctx)
//	tw.Write(event)
//	return tw.Close()
//}
