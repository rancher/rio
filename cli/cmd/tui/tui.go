package tui

import (
	"fmt"
	"time"

	"github.com/rancher/axe/throwing"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/types"
	"k8s.io/apimachinery/pkg/runtime"
	runtime2 "k8s.io/apimachinery/pkg/util/runtime"
)

type Tui struct {
}

func (t *Tui) Run(ctx *clicontext.CLIContext) error {
	signals := map[string]chan struct{}{
		appKind:             make(chan struct{}, 1),
		serviceKind:         make(chan struct{}, 1),
		routeKind:           make(chan struct{}, 1),
		externalServiceKind: make(chan struct{}, 1),
		configKind:          make(chan struct{}, 1),
		podKind:             make(chan struct{}, 1),
		publicdomainKind:    make(chan struct{}, 1),
	}

	h := handler{
		signals: signals,
	}

	rioContext := types.NewContext(ctx.SystemNamespace, ctx.RestConfig)
	go func() {
		rioContext.Rio.Rio().V1().App().AddGenericHandler(ctx.Ctx, "rio-app-tui", h.sync(appKind))
		rioContext.Rio.Rio().V1().Service().AddGenericHandler(ctx.Ctx, "rio-service-tui", h.sync(serviceKind))
		rioContext.Rio.Rio().V1().Router().AddGenericHandler(ctx.Ctx, "rio-router-tui", h.sync(routeKind))
		rioContext.Rio.Rio().V1().PublicDomain().AddGenericHandler(ctx.Ctx, "rio-domain-tui", h.sync(publicdomainKind))
		rioContext.Rio.Rio().V1().ExternalService().AddGenericHandler(ctx.Ctx, "rio-external-tui", h.sync(externalServiceKind))
		runtime2.Must(rioContext.Start(ctx.Ctx))
		<-ctx.Ctx.Done()
	}()

	tui := throwing.NewAppView(ctx.K8s, drawer, tableEventHandler, signals)
	if err := tui.Init(); err != nil {
		return err
	}
	fmt.Println("Initializing...")
	time.Sleep(time.Second * 1)
	return tui.Run()
}

type handler struct {
	signals map[string]chan struct{}
}

func (h handler) sync(kind string) func(string, runtime.Object) (runtime.Object, error) {
	return func(k string, object runtime.Object) (runtime.Object, error) {
		s := h.signals[kind]
		s <- struct{}{}
		if kind == appKind {
			s1 := h.signals[serviceKind]
			s1 <- struct{}{}

			s2 := h.signals[podKind]
			s2 <- struct{}{}
		}
		return object, nil
	}
}
